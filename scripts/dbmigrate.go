package main

/*
--------------------------------------------------------------------
A script to migrate Wakapi data from / to SQLite, MySQL or Postgres.
--------------------------------------------------------------------

Usage:
------
1. Set up an empty MySQL or Postgres database (see docker_[mysql|postgres].sh for example)
2. Create a migration config file (e.g. config.yml) as shown below
3. go run dbmigrate.go -config config.yml

Example: config.yml
-------------------
with_key_values: true
with_users: true
with_leaderboard: false
with_language_mappings: true
with_aliases: true
with_summaries: false
with_durations: false
with_heartbeats: true
with_project_labels: true
users:
	# optional, if not set, all users will be migrated
	- user1
	- user2

source:
  name: ../wakapi_db.db
  dialect: sqlite  # or mysql, postgres

target:
  host: localhost
  port: 3306
  user: user
  password: pw
  name: wakapi_db
  dialect: mysql    # or postgres, sqlite
  compress: false   # mysql only, enable if your target (or source) database is on a remote server

Troubleshooting:
----------------
- Check https://wiki.postgresql.org/wiki/Fixing_Sequences in case of errors with Postgres
- Check https://github.com/muety/wakapi/pull/181#issue-621585477 on further details about Postgres migration

To Do:
------
This script could be sped up dramatically by using streaming and batch-inserting for all entities (not only heartbeats and durations).
Alternatively, it would probably also help to not use a separate transaction for every individual insert, however, this would require otherwise unnecessary changes in the code base.
*/

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/jinzhu/configor"
	wakapiConfig "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
	"github.com/schollz/progressbar/v3"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type config struct {
	WithKeyValues        bool     `yaml:"with_key_values" default:"true"`
	WithUsers            bool     `yaml:"with_users" default:"true"`
	WithLeaderboard      bool     `yaml:"with_leaderboard" default:"false"`
	WithLanguageMappings bool     `yaml:"with_language_mappings" default:"true"`
	WithAliases          bool     `yaml:"with_aliases" default:"true"`
	WithSummaries        bool     `yaml:"with_summaries" default:"false"`
	WithDurations        bool     `yaml:"with_durations" default:"false"`
	WithHeartbeats       bool     `yaml:"with_heartbeats" default:"true"`
	WithProjectLabels    bool     `yaml:"with_project_labels" default:"true"`
	Users                []string `yaml:"users"`
	Source               dbConfig
	Target               dbConfig
}

type dbConfig struct {
	Host     string
	Port     uint
	User     string
	Password string
	Name     string
	Dialect  string `default:"mysql"`
	Compress bool   `default:"false"`
}

const InsertBatchSize = 1_024

var cfg *config
var dbSource, dbTarget *gorm.DB
var cFlag *string

func init() {
	cfg = &config{}

	wakapiConfig.Set(wakapiConfig.Empty())
	wakapiConfig.Get().Db.Dialect = cfg.Source.Dialect // only required because of the "postgresTimezoneHack" in shared.go

	if f := flag.Lookup("config"); f == nil {
		cFlag = flag.String("config", "sqlite2mysql.yml", "config file location")
	} else {
		ff := f.Value.(flag.Getter).Get().(string)
		cFlag = &ff
	}
	flag.Parse()

	if err := configor.New(&configor.Config{}).Load(cfg, mustConfigPath()); err != nil {
		log.Fatalln("failed to read config", err)
	}

	log.Printf("attempting to open %s source database\n", cfg.Source.Dialect)
	if db, err := getDb(&cfg.Source); err != nil {
		log.Fatalln(err)
	} else {
		dbSource = db
	}

	log.Printf("attempting to open %s target database\n", cfg.Target.Dialect)
	if db, err := getDb(&cfg.Target); err != nil {
		log.Fatalln(err)
	} else {
		dbTarget = db
	}
}

func destroy() {
	if db, _ := dbSource.DB(); db != nil {
		db.Close()
	}
	if db, _ := dbTarget.DB(); db != nil {
		db.Close()
	}
}

func main() {
	defer destroy()
	if err := createSchema(); err != nil {
		log.Fatalln(err)
	}

	keyValueSource := repositories.NewKeyValueRepository(dbSource)
	keyValueTarget := repositories.NewKeyValueRepository(dbTarget)

	userSource := repositories.NewUserRepository(dbSource)
	userTarget := repositories.NewUserRepository(dbTarget)

	leaderboardSource := repositories.NewLeaderboardRepository(dbSource)
	leaderboardTarget := repositories.NewLeaderboardRepository(dbTarget)

	languageMappingSource := repositories.NewLanguageMappingRepository(dbSource)
	languageMappingTarget := repositories.NewLanguageMappingRepository(dbTarget)

	aliasSource := repositories.NewAliasRepository(dbSource)
	aliasTarget := repositories.NewAliasRepository(dbTarget)

	summarySource := repositories.NewSummaryRepository(dbSource)
	summaryTarget := repositories.NewSummaryRepository(dbTarget)

	durationsSource := repositories.NewDurationRepository(dbSource)
	durationsTarget := repositories.NewDurationRepository(dbTarget)

	heartbeatSource := repositories.NewHeartbeatRepository(dbSource)
	heartbeatTarget := repositories.NewHeartbeatRepository(dbTarget)

	projectLabelsSource := repositories.NewProjectLabelRepository(dbSource)
	projectLabelsTarget := repositories.NewProjectLabelRepository(dbTarget)

	var bar *progressbar.ProgressBar

	getUsers := userSource.GetAll
	if len(cfg.Users) > 0 {
		getUsers = func() ([]*models.User, error) {
			return userSource.GetMany(cfg.Users)
		}
	}
	users, err := getUsers()
	if err != nil {
		log.Fatalln(err)
	}

	if cfg.WithKeyValues {
		log.Println("Migrating key-value pairs ...")
		if data, err := keyValueSource.GetAll(); err == nil {
			bar = progressbar.Default(int64(len(data)))
			for _, e := range data {
				if err := keyValueTarget.PutString(e); err != nil {
					log.Printf("warning: failed to insert key-value pair %s (%s)\n", e.Key, err)
					continue
				}
				bar.Add(1)
			}
		} else {
			log.Fatalln(err)
		}
	}

	if cfg.WithUsers {
		log.Println("Migrating users ...")
		bar = progressbar.Default(int64(len(users)))
		for _, e := range users {
			if _, _, err := userTarget.InsertOrGet(e); err != nil {
				log.Printf("warning: failed to insert user %s (%s)\n", e.ID, err)
				continue
			}
			bar.Add(1)
		}
	}

	if cfg.WithLanguageMappings {
		log.Println("Migrating language mappings ...")
		bar = progressbar.Default(int64(len(users)))
		for _, user := range users {
			if data, err := languageMappingSource.GetByUser(user.ID); err == nil {
				for _, e := range data {
					id := e.ID
					e.ID = 0
					if _, err := languageMappingTarget.Insert(e); err != nil {
						log.Printf("warning: failed to insert language mapping %d (%s)\n", id, err)
						continue
					}
				}
			} else {
				log.Fatalln(err)
			}
			bar.Add(1)
		}
	}

	if cfg.WithProjectLabels {
		log.Println("Migrating project labels ...")
		bar = progressbar.Default(int64(len(users)))
		for _, user := range users {
			if data, err := projectLabelsSource.GetByUser(user.ID); err == nil {
				for _, e := range data {
					id := e.ID
					e.ID = 0
					if _, err := projectLabelsTarget.Insert(e); err != nil {
						log.Printf("warning: failed to insert project label %d (%s)\n", id, err)
						continue
					}
				}
			} else {
				log.Fatalln(err)
			}
			bar.Add(1)
		}
	}

	if cfg.WithAliases {
		log.Println("Migrating aliases ...")
		bar = progressbar.Default(int64(len(users)))
		for _, user := range users {
			if data, err := aliasSource.GetByUser(user.ID); err == nil {
				for _, e := range data {
					id := e.ID
					e.ID = 0
					if _, err := aliasTarget.Insert(e); err != nil {
						log.Printf("warning: failed to insert alias %d (%s)\n", id, err)
						continue
					}
				}
			} else {
				log.Fatalln(err)
			}
			bar.Add(1)
		}
	}

	if cfg.WithLeaderboard {
		log.Println("Migrating leaderboard ...")
		if data, err := leaderboardSource.GetAll(); err == nil {
			if err := leaderboardTarget.InsertBatch(data); err != nil {
				log.Printf("warning: failed to migrate leaderboards (%s)\n", err)
			}
			bar.Add(len(data))
		} else {
			log.Fatalln(err)
		}
	}

	if cfg.WithSummaries {
		// TODO: stream and batch-insert
		log.Println("Migrating summaries ...")
		bar = progressbar.Default(int64(len(users)))
		for _, user := range users {
			if data, err := summarySource.GetByUserWithin(user, time.Time{}, time.Now()); err == nil {
				for _, e := range data {
					id := e.ID
					e.ID = 0
					if err := summaryTarget.Insert(e); err != nil {
						log.Printf("warning: failed to insert summary %d (%s)\n", id, err)
						continue
					}
				}
			} else {
				log.Fatalln(err)
			}
			bar.Add(1)
		}
	}

	if cfg.WithDurations {
		// TODO: stream and batch-insert
		log.Println("Migrating durations ...")
		bar = progressbar.Default(int64(len(users)))
		for _, user := range users {
			if data, err := durationsSource.StreamByUserBatched(user, InsertBatchSize); err == nil {
				for durations := range data {
					if err := durationsTarget.InsertBatch(durations); err != nil {
						log.Printf("warning: failed to insert batch of durations (%s)\n", err)
						continue
					}
				}
			} else {
				log.Fatalln(err)
			}
			bar.Add(1)
		}
	}

	if cfg.WithHeartbeats {
		log.Println("Migrating heartbeats ...")
		bar = progressbar.Default(int64(len(users)))
		for _, user := range users {
			if data, err := heartbeatSource.StreamWithinBatched(time.Time{}, time.Now(), user, InsertBatchSize); err == nil {
				for heartbeats := range data {
					fixHeartbeatsBatched(heartbeats)
					if err := heartbeatTarget.InsertBatch(heartbeats); err != nil {
						log.Printf("warning: failed to insert batch of heartbeats for user (%s)\n", user.ID, err)
						continue
					}
				}
			} else {
				log.Fatalln(err)
			}

			bar.Add(1)
		}
	}
}

func getDb(cfg *dbConfig) (*gorm.DB, error) {
	gormLogger := logger.New(
		log.New(os.Stdout, "", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Minute,
			Colorful:      false,
			LogLevel:      logger.Silent,
		},
	)

	if cfg.Dialect == "sqlite" {
		return gorm.Open(sqlite.Open(cfg.Name), &gorm.Config{
			Logger: gormLogger,
		})
	}
	if cfg.Dialect == "mysql" {
		return gorm.Open(mysql.New(mysql.Config{
			DriverName: "mysql",
			DSN: fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=%s&compress=%v&sql_mode=ANSI_QUOTES",
				cfg.User,
				cfg.Password,
				cfg.Host,
				cfg.Port,
				cfg.Name,
				"utf8mb4",
				"Local",
				cfg.Compress,
			),
		}), &gorm.Config{
			Logger: gormLogger,
		})
	}
	if cfg.Dialect == "postgres" {
		return gorm.Open(postgres.Open(fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=disable timezone=Europe/Berlin",
			cfg.User,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Name,
		)), &gorm.Config{
			Logger: gormLogger,
		})
	}

	return nil, fmt.Errorf("unsupported dialect %s", cfg.Dialect)
}

func createSchema() error {
	if err := dbTarget.AutoMigrate(&models.User{}); err != nil {
		return err
	}
	if err := dbTarget.AutoMigrate(&models.Credential{}); err != nil {
		return err
	}
	if err := dbTarget.AutoMigrate(&models.KeyStringValue{}); err != nil {
		return err
	}
	if err := dbTarget.AutoMigrate(&models.Alias{}); err != nil {
		return err
	}
	if err := dbTarget.AutoMigrate(&models.Heartbeat{}); err != nil {
		return err
	}
	if err := dbTarget.AutoMigrate(&models.Summary{}); err != nil {
		return err
	}
	if err := dbTarget.AutoMigrate(&models.SummaryItem{}); err != nil {
		return err
	}
	if err := dbTarget.AutoMigrate(&models.LanguageMapping{}); err != nil {
		return err
	}
	if err := dbTarget.AutoMigrate(&models.Diagnostics{}); err != nil {
		return err
	}
	if err := dbTarget.AutoMigrate(&models.LeaderboardItem{}); err != nil {
		return err
	}
	if err := dbTarget.AutoMigrate(&models.ProjectLabel{}); err != nil {
		return err
	}
	if err := dbTarget.AutoMigrate(&models.Duration{}); err != nil {
		return err
	}
	return nil
}

func fixHeartbeatsBatched(heartbeats []*models.Heartbeat) {
	// normally, there shouldn't be any heartbeats without a hash,
	// however, we observed this case in production, so make sure to fix them to prevent data loss
	for _, h := range heartbeats {
		if h.Hash == "" {
			h.Hashed()
		}
	}
}

func mustConfigPath() string {
	if _, err := os.Stat(*cFlag); err != nil {
		log.Fatalln("failed to find config file at", *cFlag)
	}
	return *cFlag
}
