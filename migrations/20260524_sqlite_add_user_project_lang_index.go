package migrations

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

func init() {
	const name = "20260524-sqlite_add_user_project_lang_index"

	f := migrationFunc{
		name:       name,
		background: false,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			if cfg.Db.IsSQLite() {
				if db.Migrator().HasColumn(&models.Heartbeat{}, "time_real") {
					if err := db.Exec("create index if not exists idx_time_real_user_project_lang on heartbeats(user_id, time_real, project, language)").Error; err != nil {
						return err
					}
				}
			}

			setHasRun(name, db)
			return nil
		},
	}

	registerPostMigration(f)
}
