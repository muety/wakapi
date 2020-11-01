package common

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
	"log"
)

var customPreMigrations []migrationFunc

func init() {
	customPreMigrations = []migrationFunc{
		{
			f: func(db *gorm.DB, cfg *config.Config) error {
				if db.Migrator().HasTable("custom_rules") {
					return db.Migrator().RenameTable("custom_rules", &models.LanguageMapping{})
				}
				return nil
			},
			name: "rename language mappings table",
		},
	}
}

func RunCustomPreMigrations(db *gorm.DB, cfg *config.Config) {
	for _, m := range customPreMigrations {
		log.Printf("running migration '%s'\n", m.name)
		if err := m.f(db, cfg); err != nil {
			log.Fatalf("migration '%s' failed – %v\n", m.name, err)
		}
	}
}
