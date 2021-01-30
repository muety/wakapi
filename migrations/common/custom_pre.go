package common

import (
	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
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
					logbuch.Info("renaming '%s' table to '%s'", oldTableName, newTableName)
					if err := migrator.RenameTable(oldTableName, &models.LanguageMapping{}); err != nil {
						return err
					}

					logbuch.Info("renaming '%s' index to '%s'", oldIndexName, newIndexName)
					return migrator.RenameIndex(&models.LanguageMapping{}, oldIndexName, newIndexName)
				}
				return nil
			},
			name: "rename language mappings table",
		},
		{
			f: func(db *gorm.DB, cfg *config.Config) error {
				// drop all already existing foreign key constraints
				// afterwards let them be re-created by auto migrate with the newly introduced cascade settings,

				migrator := db.Migrator()
				const lookupKey = "20201106-migration_cascade_constraints"

				if cfg.Db.Dialect == config.SQLDialectSqlite {
					// https://stackoverflow.com/a/1884893/3112139
					// unfortunately, we can't migrate existing sqlite databases to the newly introduced cascade settings
					// things like deleting all summaries won't work in those cases unless an entirely new db is created
					logbuch.Info("not attempting to drop and regenerate constraints on sqlite")
					return nil
				}

				if !migrator.HasTable(&models.KeyStringValue{}) {
					logbuch.Info("key-value table not yet existing")
					return nil
				}

				condition := "key = ?"
				if cfg.Db.Dialect == config.SQLDialectMysql {
					condition = "`key` = ?"
				}
				lookupResult := db.Where(condition, lookupKey).First(&models.KeyStringValue{})
				if lookupResult.Error == nil && lookupResult.RowsAffected > 0 {
					logbuch.Info("no need to migrate cascade constraints")
					return nil
				}

				// SELECT * FROM INFORMATION_SCHEMA.table_constraints;
				constraints := map[string]interface{}{
					"fk_summaries_editors":           &models.SummaryItem{},
					"fk_summaries_languages":         &models.SummaryItem{},
					"fk_summaries_machines":          &models.SummaryItem{},
					"fk_summaries_operating_systems": &models.SummaryItem{},
					"fk_summaries_projects":          &models.SummaryItem{},
					"fk_summary_items_summary":       &models.SummaryItem{},
					"fk_summaries_user":              &models.Summary{},
					"fk_language_mappings_user":      &models.LanguageMapping{},
					"fk_heartbeats_user":             &models.Heartbeat{},
					"fk_aliases_user":                &models.Alias{},
				}

				for name, table := range constraints {
					if migrator.HasConstraint(table, name) {
						logbuch.Info("dropping constraint '%s'", name)
						if err := migrator.DropConstraint(table, name); err != nil {
							return err
						}
					}
				}

				if err := db.Create(&models.KeyStringValue{
					Key:   lookupKey,
					Value: "done",
				}).Error; err != nil {
					return err
				}

				return nil
			},
			name: "add cascade constraints",
		},
	}
}

func RunCustomPreMigrations(db *gorm.DB, cfg *config.Config) {
	for _, m := range customPreMigrations {
		logbuch.Info("potentially running migration '%s'", m.name)
		if err := m.f(db, cfg); err != nil {
			logbuch.Fatal("migration '%s' failed – %v", m.name, err)
		}
	}
}
