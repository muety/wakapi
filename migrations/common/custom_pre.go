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
				migrator := db.Migrator()
				oldTableName, newTableName := "custom_rules", "language_mappings"
				oldIndexName, newIndexName := "idx_customrule_user", "idx_language_mapping_user"

				if migrator.HasTable(oldTableName) {
					log.Printf("renaming '%s' table to '%s'\n", oldTableName, newTableName)
					if err := migrator.RenameTable(oldTableName, &models.LanguageMapping{}); err != nil {
						return err
					}

					log.Printf("renaming '%s' index to '%s'\n", oldIndexName, newIndexName)
					return migrator.RenameIndex(&models.LanguageMapping{}, oldIndexName, newIndexName)
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
