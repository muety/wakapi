package migrations

import (
	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

func init() {
	const name = "20201106-migration_cascade_constraints"

	f := migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			// drop all already existing foreign key constraints
			// afterwards let them be re-created by auto migrate with the newly introduced cascade settings,

			migrator := db.Migrator()

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
			lookupResult := db.Where(condition, name).First(&models.KeyStringValue{})
			if lookupResult.Error == nil && lookupResult.RowsAffected > 0 {
				logbuch.Info("no need to migrate '%s'", name)
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
				Key:   name,
				Value: "done",
			}).Error; err != nil {
				return err
			}

			return nil
		},
	}

	registerPreMigration(f)
}
