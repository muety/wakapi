package common

import (
	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/config"
	"gorm.io/gorm"
)

var customPostMigrations []migrationFunc

func init() {
	customPostMigrations = []migrationFunc{
		{
			f: func(db *gorm.DB, cfg *config.Config) error {
				return cfg.GetFixturesFunc(cfg.Db.Dialect)(db)
			},
			name: "apply fixtures",
		},
		// TODO: add function to modify aggregated summaries according to configured custom language mappings
	}
}

func RunCustomPostMigrations(db *gorm.DB, cfg *config.Config) {
	for _, m := range customPostMigrations {
		logbuch.Info("potentially running migration '%s'", m.name)
		if err := m.f(db, cfg); err != nil {
			logbuch.Fatal("migration '%s' failed – %v", m.name, err)
		}
	}
}
