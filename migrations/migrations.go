package migrations

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
	"log"
	"log/slog"
	"sort"
	"strings"
)

type gormMigrationFunc func(db *gorm.DB) error

type migrationFunc struct {
	f    func(db *gorm.DB, cfg *config.Config) error
	name string
}

type migrationFuncs []migrationFunc

var (
	preMigrations  migrationFuncs
	postMigrations migrationFuncs
)

func GetMigrationFunc(cfg *config.Config) gormMigrationFunc {
	switch cfg.Db.Dialect {
	default:
		return func(db *gorm.DB) error {
			if err := db.AutoMigrate(&models.User{}); err != nil && !cfg.Db.AutoMigrateFailSilently {
				return err
			}
			if err := db.AutoMigrate(&models.KeyStringValue{}); err != nil && !cfg.Db.AutoMigrateFailSilently {
				return err
			}
			if err := db.AutoMigrate(&models.Alias{}); err != nil && !cfg.Db.AutoMigrateFailSilently {
				return err
			}
			if err := db.AutoMigrate(&models.Heartbeat{}); err != nil && !cfg.Db.AutoMigrateFailSilently {
				return err
			}
			if err := db.AutoMigrate(&models.Summary{}); err != nil && !cfg.Db.AutoMigrateFailSilently {
				return err
			}
			if err := db.AutoMigrate(&models.SummaryItem{}); err != nil && !cfg.Db.AutoMigrateFailSilently {
				return err
			}
			if err := db.AutoMigrate(&models.LanguageMapping{}); err != nil && !cfg.Db.AutoMigrateFailSilently {
				return err
			}
			if err := db.AutoMigrate(&models.ProjectLabel{}); err != nil && !cfg.Db.AutoMigrateFailSilently {
				return err
			}
			if err := db.AutoMigrate(&models.Diagnostics{}); err != nil && !cfg.Db.AutoMigrateFailSilently {
				return err
			}
			if err := db.AutoMigrate(&models.LeaderboardItem{}); err != nil && !cfg.Db.AutoMigrateFailSilently {
				return err
			}
			return nil
		}
	}
}

func registerPreMigration(f migrationFunc) {
	preMigrations = append(preMigrations, f)
}

func registerPostMigration(f migrationFunc) {
	postMigrations = append(postMigrations, f)
}

func Run(db *gorm.DB, cfg *config.Config) {
	RunPreMigrations(db, cfg)
	RunSchemaMigrations(db, cfg)
	RunPostMigrations(db, cfg)
}

func RunSchemaMigrations(db *gorm.DB, cfg *config.Config) {
	if err := GetMigrationFunc(cfg)(db); err != nil {
		log.Fatal(err.Error())
	}
}

func RunPreMigrations(db *gorm.DB, cfg *config.Config) {
	sort.Sort(preMigrations)

	for _, m := range preMigrations {
		slog.Info("potentially running migration", "name", m.name)
		if err := m.f(db, cfg); err != nil {
			log.Fatalf("migration '%s' failed - %v", m.name, err)
		}
	}
}

func RunPostMigrations(db *gorm.DB, cfg *config.Config) {
	sort.Sort(postMigrations)

	for _, m := range postMigrations {
		slog.Info("potentially running migration", "name", m.name)
		if err := m.f(db, cfg); err != nil {
			log.Fatalf("migration '%s' failed - %v", m.name, err)
		}
	}
}

func (m migrationFuncs) Len() int {
	return len(m)
}

func (m migrationFuncs) Less(i, j int) bool {
	return strings.Compare(m[i].name, m[j].name) < 0
}

func (m migrationFuncs) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
