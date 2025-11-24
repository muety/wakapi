package migrations

import (
	"log/slog"
	"sort"
	"strings"

	"gorm.io/gorm"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
)

type GormMigrationFunc func(db *gorm.DB) error

type migrationFunc struct {
	f    func(db *gorm.DB, cfg *config.Config) error
	name string
	// be careful with background migrations, because they must not lock or have side effects on each other, and they must be safe to fail silently
	background bool
}

type migrationFuncs []migrationFunc

var (
	preMigrations  migrationFuncs
	postMigrations migrationFuncs
)

func GetMigrationFunc(cfg *config.Config) GormMigrationFunc {
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
			if err := db.AutoMigrate(&models.Duration{}); err != nil && !cfg.Db.AutoMigrateFailSilently {
				return err
			}
			if err := db.AutoMigrate(&models.ApiKey{}); err != nil && !cfg.Db.AutoMigrateFailSilently {
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
		config.Log().Fatal("migration failed", "error", err)
	}
}

func RunPreMigrations(db *gorm.DB, cfg *config.Config) {
	sort.Sort(preMigrations)

	for _, m := range preMigrations {
		slog.Info("potentially running migration", "name", m.name)
		if err := m.f(db, cfg); err != nil {
			config.Log().Fatal("migration failed", "name", m.name, "error", err)
		}
	}
}

func RunPostMigrations(db *gorm.DB, cfg *config.Config) {
	sort.Sort(postMigrations)

	for _, m := range postMigrations {
		slog.Info("potentially running migration", "name", m.name)

		run := func(m migrationFunc) {
			if err := m.f(db, cfg); err != nil {
				config.Log().Fatal("migration failed", "name", m.name, "error", err)
			}
		}

		if m.background {
			go run(m)
		} else {
			run(m)
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
