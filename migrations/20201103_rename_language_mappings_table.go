package migrations

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
	"log/slog"
)

func init() {
	f := migrationFunc{
		name: "20201103-rename_language_mappings_table",
		f: func(db *gorm.DB, cfg *config.Config) error {
			migrator := db.Migrator()
			oldTableName, newTableName := "custom_rules", "language_mappings"
			oldIndexName, newIndexName := "idx_customrule_user", "idx_language_mapping_user"

			if migrator.HasTable(oldTableName) {
				slog.Info("renaming '%s' table to '%s'", oldTableName, newTableName)
				if err := migrator.RenameTable(oldTableName, &models.LanguageMapping{}); err != nil {
					return err
				}

				slog.Info("renaming '%s' index to '%s'", oldIndexName, newIndexName)
				return migrator.RenameIndex(&models.LanguageMapping{}, oldIndexName, newIndexName)
			}
			return nil
		},
	}

	registerPreMigration(f)
}
