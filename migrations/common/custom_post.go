package common

import (
	"github.com/muety/wakapi/config"
	"gorm.io/gorm"
	"log"
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
		log.Printf("potentially running migration '%s'\n", m.name)
		if err := m.f(db, cfg); err != nil {
			log.Fatalf("migration '%s' failed – %v\n", m.name, err)
		}
	}
}
