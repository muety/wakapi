package migrations

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
	"strings"
)

// fix for https://github.com/muety/wakapi/issues/416

func init() {
	const name = "20221002-fix_summary_id_types"

	f := migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if cfg.Db.Dialect != config.SQLDialectMysql {
				return nil
			}

			if !db.Migrator().HasTable(&models.Summary{}) || !db.Migrator().HasTable(&models.SummaryItem{}) {
				return nil
			}

			var currentType string
			if err := db.
				Table("information_schema.columns").
				Select("data_type").
				Where("table_name = ?", "summary_items").
				Where("column_name = ?", "summary_id").
				Limit(1).
				Row().Scan(&currentType); err != nil {
				return err
			}

			if strings.ToLower(currentType) != "int" {
				if db.Migrator().HasConstraint(&models.SummaryItem{}, "fk_summaries_editors") {
					if err := db.Migrator().DropConstraint(&models.SummaryItem{}, "fk_summaries_editors"); err != nil {
						return err
					}
				}
				if db.Migrator().HasConstraint(&models.SummaryItem{}, "fk_summaries_languages") {
					if err := db.Migrator().DropConstraint(&models.SummaryItem{}, "fk_summaries_languages"); err != nil {
						return err
					}
				}
				if db.Migrator().HasConstraint(&models.SummaryItem{}, "fk_summaries_machines") {
					if err := db.Migrator().DropConstraint(&models.SummaryItem{}, "fk_summaries_machines"); err != nil {
						return err
					}
				}
				if db.Migrator().HasConstraint(&models.SummaryItem{}, "fk_summaries_operating_systems") {
					if err := db.Migrator().DropConstraint(&models.SummaryItem{}, "fk_summaries_operating_systems"); err != nil {
						return err
					}
				}
				if db.Migrator().HasConstraint(&models.SummaryItem{}, "fk_summaries_projects") {
					if err := db.Migrator().DropConstraint(&models.SummaryItem{}, "fk_summaries_projects"); err != nil {
						return err
					}
				}
				// https://github.com/muety/wakapi/issues/416#issuecomment-1271674792
				if db.Migrator().HasConstraint(&models.SummaryItem{}, "fk_summary_items_summary") {
					if err := db.Migrator().DropConstraint(&models.SummaryItem{}, "fk_summary_items_summary"); err != nil {
						return err
					}
				}
				if db.Migrator().HasConstraint(&models.SummaryItem{}, "fk_summaries_labels") {
					if err := db.Migrator().DropConstraint(&models.SummaryItem{}, "fk_summaries_labels"); err != nil {
						return err
					}
				}

				if err := db.Migrator().AlterColumn(&models.Summary{}, "id"); err != nil {
					return err
				}
				if err := db.Migrator().AlterColumn(&models.SummaryItem{}, "summary_id"); err != nil {
					return err
				}
			}

			return nil
		},
	}

	registerPreMigration(f)
}
