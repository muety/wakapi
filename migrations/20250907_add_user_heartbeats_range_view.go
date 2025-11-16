package migrations

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/utils"
	"gorm.io/gorm"
)

func init() {
	const name = "20250907-add_user_heartbeats_range_view"
	f := migrationFunc{
		name:       name,
		background: false,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			const q = "select u.id as user_id, min(h.time) as first, max(h.time) as last " +
				"from users u left join heartbeats h on u.id = h.user_id " +
				"group by u.id"

			if err := db.Transaction(func(tx *gorm.DB) error {
				// https://stackoverflow.com/a/1236008/3112139
				if cfg.Db.IsSQLite() {
					if err := tx.Migrator().DropView("user_heartbeats_range"); err != nil {
						return err
					}
				}

				if err := tx.Migrator().CreateView("user_heartbeats_range", gorm.ViewOption{
					Query:   db.Raw(q),
					Replace: !cfg.Db.IsSQLite(),
				}); err != nil {
					return err
				}
				if err := tx.Exec("delete from key_string_values where "+utils.QuoteSql(db, "%s like ?", "key"), "first_heartbeat_%").Error; err != nil {
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
