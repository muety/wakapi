package main

/*
Usage:
---
1. Set up a MySQL instance (see docker_mysql.sh for example)
2. Create config file (e.g. migrate.yml) as shown below
3. go run sqlite2mysql.go -config migrate.yml

Example: migrate.yml
---
# SQLite
source:
  # Example: ../wakapi_db.db (relative to script path)
  name:

# MySQL
target:
  host:
  port:
  user:
  password:
  name:
*/

import (
	"flag"
	"fmt"
	"github.com/jinzhu/configor"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"os"
)

type config struct {
	Source dbConfig // sqlite
	Target dbConfig // mysql
}

type dbConfig struct {
	Host     string
	Port     uint
	User     string
	Password string
	Name     string
}

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

	log.Println("attempting to open sqlite database as Source")
	if db, err := gorm.Open(sqlite.Open(cfg.Source.Name), &gorm.Config{}); err != nil {
		log.Fatalln(err)
	} else {
		dbSource = db
	}

	log.Println("attempting to open mysql database as Target")
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
		if err := heartbeatTarget.InsertBatch(data); err != nil {
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
