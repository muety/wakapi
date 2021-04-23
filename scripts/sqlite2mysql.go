package main

/*
A script to migrate Wakapi data from SQLite to MySQL or Postgres.

Usage:
---
1. Set up an empty MySQL or Postgres database (see docker_[mysql|postgres].sh for example)
2. Create a migration config file (e.g. config.yml) as shown below
3. go run sqlite2mysql.go -config config.yml

Example: config.yml
---
source:
  name: ../wakapi_db.db

# MySQL / Postgres
target:
  host:
  port:
  user:
  password:
  name:
  dialect:

Troubleshooting:
---
- Check https://wiki.postgresql.org/wiki/Fixing_Sequences in case of errors with Postgres
- Check https://github.com/muety/wakapi/pull/181#issue-621585477 on further details about Postgres migration
*/

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jinzhu/configor"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type config struct {
	Source dbConfig // sqlite
	Target dbConfig // mysql / postgres
}

type dbConfig struct {
	Host     string
	Port     uint
	User     string
	Password string
	Name     string
	Dialect  string `default:"mysql"`
}

const InsertBatchSize = 100

var cfg *config
var dbSource, dbTarget *gorm.DB
var cFlag *string

func init() {
	cfg = &config{}

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

	log.Println("attempting to open sqlite database as source")
	if db, err := gorm.Open(sqlite.Open(cfg.Source.Name), &gorm.Config{}); err != nil {
		log.Fatalln(err)
	} else {
		dbSource = db
	}

	if cfg.Target.Dialect == "postgres" {
		log.Println("attempting to open postgresql database as target")
		if db, err := gorm.Open(postgres.Open(fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=disable timezone=Europe/Berlin",
			cfg.Target.User,
			cfg.Target.Password,
			cfg.Target.Host,
			cfg.Target.Port,
			cfg.Target.Name,
		)), &gorm.Config{}); err != nil {
			log.Fatalln(err)
		} else {
			dbTarget = db
		}
	} else {
		log.Println("attempting to open mysql database as target")
		if db, err := gorm.Open(mysql.New(mysql.Config{
			DriverName: "mysql",
			DSN: fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=%s&sql_mode=ANSI_QUOTES",
				cfg.Target.User,
				cfg.Target.Password,
				cfg.Target.Host,
				cfg.Target.Port,
				cfg.Target.Name,
				"utf8mb4",
				"Local",
			),
		}), &gorm.Config{}); err != nil {
			log.Fatalln(err)
		} else {
			dbTarget = db
		}
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

	languageMappingSource := repositories.NewLanguageMappingRepository(dbSource)
	languageMappingTarget := repositories.NewLanguageMappingRepository(dbTarget)

	aliasSource := repositories.NewAliasRepository(dbSource)
	aliasTarget := repositories.NewAliasRepository(dbTarget)

	summarySource := repositories.NewSummaryRepository(dbSource)
	summaryTarget := repositories.NewSummaryRepository(dbTarget)

	heartbeatSource := repositories.NewHeartbeatRepository(dbSource)
	heartbeatTarget := repositories.NewHeartbeatRepository(dbTarget)

	// TODO: things could be optimized through batch-inserts / inserts within a single transaction

	log.Println("Migrating key-value pairs ...")
	if data, err := keyValueSource.GetAll(); err == nil {
		for _, e := range data {
			if err := keyValueTarget.PutString(e); err != nil {
				log.Fatalln(err)
			}
		}
	} else {
		log.Fatalln(err)
	}

	log.Println("Migrating users ...")
	if data, err := userSource.GetAll(); err == nil {
		for _, e := range data {
			if _, _, err := userTarget.InsertOrGet(e); err != nil {
				log.Fatalln(err)
			}
		}
	} else {
		log.Fatalln(err)
	}

	log.Println("Migrating language mappings ...")
	if data, err := languageMappingSource.GetAll(); err == nil {
		for _, e := range data {
			e.ID = 0
			if _, err := languageMappingTarget.Insert(e); err != nil {
				log.Fatalln(err)
			}
		}
	} else {
		log.Fatalln(err)
	}

	log.Println("Migrating aliases ...")
	if data, err := aliasSource.GetAll(); err == nil {
		for _, e := range data {
			e.ID = 0
			if _, err := aliasTarget.Insert(e); err != nil {
				log.Fatalln(err)
			}
		}
	} else {
		log.Fatalln(err)
	}

	log.Println("Migrating summaries ...")
	if data, err := summarySource.GetAll(); err == nil {
		for _, e := range data {
			e.ID = 0
			if err := summaryTarget.Insert(e); err != nil {
				log.Fatalln(err)
			}
		}
	} else {
		log.Fatalln(err)
	}

	// TODO: copy in mini-batches instead of loading all heartbeats into memory (potentially millions)

	log.Println("Migrating heartbeats ...")

	if data, err := heartbeatSource.GetAll(); err == nil {
		log.Printf("Got %d heartbeats loaded into memory. Batch-inserting them now ...\n", len(data))

		var slice = make([]*models.Heartbeat, len(data))
		for i, heartbeat := range data {
			heartbeat = heartbeat.Hashed()
			slice[i] = heartbeat
		}

		left, right, size := 0, InsertBatchSize, len(slice)
		for right < size {
			log.Printf("Inserting batch from %d", left)
			if err := heartbeatTarget.InsertBatch(slice[left:right]); err != nil {
				log.Fatalln(err)
			}
			left += InsertBatchSize
			right += InsertBatchSize
		}
		if err := heartbeatTarget.InsertBatch(slice[left:]); err != nil {
			log.Fatalln(err)
		}
	} else {
		log.Fatalln(err)
	}
}

func createSchema() error {
	if err := dbTarget.AutoMigrate(&models.User{}); err != nil {
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
	return nil
}

func mustConfigPath() string {
	if _, err := os.Stat(*cFlag); err != nil {
		log.Fatalln("failed to find config file at", *cFlag)
	}
	return *cFlag
}
