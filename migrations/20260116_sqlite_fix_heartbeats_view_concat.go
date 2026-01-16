package migrations

import (
	"github.com/muety/wakapi/config"
	"gorm.io/gorm"
)

func init() {
	const name = "20260116-sqlite_fix_heartbeats_view_concat"

	f := migrationFunc{
		name:       name,
		background: false,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			if !cfg.Db.IsSQLite() {
				return nil
			}

			if err := db.Transaction(func(tx *gorm.DB) error {
				var viewExists int
				if err := tx.Raw("select count(*) from sqlite_master where type = 'view' and name = 'user_heartbeats_range';").Scan(&viewExists).Error; err != nil {
					return err
				}

				if viewExists == 1 {
					if err := tx.Migrator().DropView("user_heartbeats_range"); err != nil {
						return err
					}
				}

				const viewDdl = "select u.id as user_id, datetime(min(h.time_real)) || '+00:00' as first, datetime(max(h.time_real)) || '+00:00' as last " +
					"from users u left join heartbeats h on u.id = h.user_id " +
					"group by u.id"

				if err := tx.Migrator().CreateView("user_heartbeats_range", gorm.ViewOption{
					Query:   db.Raw(viewDdl),
					Replace: false, // SQLite does not support CREATE OR REPLACE
				}); err != nil {
					return err
				}

				return nil
			}); err != nil {
				return err
			}

			setHasRun(name, db)
			return nil
		},
	}

	registerPostMigration(f)
}
