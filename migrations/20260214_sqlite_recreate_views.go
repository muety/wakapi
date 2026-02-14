package migrations

import (
	"github.com/muety/wakapi/config"
	"gorm.io/gorm"
)

func init() {
	const preName = "202601214-sqlite_recreate_views-pre"
	const postName = "202601214-sqlite_recreate_views-post"

	// This pair of pre- and post migrations is a workaround for https://github.com/go-gorm/sqlite/issues/225.
	// In case of SQLite, we can't drop and re-create a table (with altered schema), if a view still referenced it.
	// This, we simply drop all existing views before auto-migration runs and recreate them afterwards,
	// unless they either already exist of were "dequeued" from the backup as part of a custom migration.

	registerPreMigration(migrationFunc{
		name:       preName,
		background: false,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if !cfg.Db.IsSQLite() {
				return nil
			}
			return backupView(db)
		},
	})

	registerPostMigration(migrationFunc{
		name:       postName,
		background: false,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if !cfg.Db.IsSQLite() {
				return nil
			}
			return restoreView(db)
		},
	})
}
