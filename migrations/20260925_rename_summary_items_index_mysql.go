package migrations

import (
	"log/slog"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MySQL auto-creates indexes for foreign key columns, e.g. summary_items.summary_id referencing summary.id once for every entity type (projects, languages, etc.)
// This resulted in an index named "fk_summaries_projects" to be created (or, potentially non-deterministically also "fk_summaries_languages", etc.).
// Neither Postgres nor SQLite have this auto-index-creation mechanism though, which is why we later explicitly added an according index via a GORM tag.
// As explained at https://stackoverflow.com/a/26536118/3112139, MySQL will reuse this already existing index when and not create its own.
// To avoid having two indexes on the same column with different names ("fk_..." and "idx_..."), we rename the existing one (if existing) to what is declared by the GORM tag.
// See https://github.com/muety/wakapi/issues/974 for details.

func init() {
	const name = "20260925-rename_summary_items_index_mysql"
	f := migrationFunc{
		name:       name,
		background: false,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			if !cfg.Db.IsMySQL() || !db.Migrator().HasTable(&models.SummaryItem{}) {
				return nil
			}

			// this is only supported from mysql 5.7.9 and mariadb 10.5.2, see https://github.com/muety/wakapi/issues/981
			// however, we require mysql 8 and mariadb 11.4 in the project requirements anyway, so we can just enforce index renaming here
			db.Dialector.(*mysql.Dialector).Config.DontSupportRenameIndex = false

			indexCandidates := []string{
				"fk_summaries_projects",
				"fk_summaries_languages",
				"fk_summaries_editors",
				"fk_summaries_operating_systems",
				"fk_summaries_machines",
				"fk_summaries_labels",
				"fk_summaries_branches",
				"fk_summaries_entities",
				"fk_summaries_categories",
			}
			for _, indexName := range indexCandidates {
				if db.Migrator().HasIndex(&models.SummaryItem{}, indexName) {
					// rename old, auto-created index to new naming scheme
					slog.Info("renaming summary_items foreign key column index as part of migration", "index_old", indexName, "index_new", "idx_summary_item_summary", "migration", name)
					if err := db.Migrator().RenameIndex(&models.SummaryItem{}, indexName, "idx_summary_item_summary"); err != nil {
						return err
					}
					break
				}
			}
			for _, indexName := range indexCandidates {
				if db.Migrator().HasIndex(&models.SummaryItem{}, indexName) {
					// in addition, just in case there were multiple old indexes on the same column, additionally drop the redundant ones
					slog.Info("dropping summary_items index as part of migration", "index", indexName, "migration", name)
					if err := db.Migrator().DropIndex(&models.SummaryItem{}, indexName); err != nil {
						return err
					}
				}
			}

			setHasRun(name, db)
			return nil
		},
	}

	registerPreMigration(f)
}
