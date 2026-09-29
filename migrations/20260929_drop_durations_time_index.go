package migrations

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

// Index is effectively superseded by idx_time_duration_users, see https://github.com/muety/wakapi/issues/976

func init() {
	const name = "20260929-drop_durations_time_index"
	f := migrationFunc{
		name:       name,
		background: true,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			if db.Migrator().HasIndex(&models.Duration{}, "idx_time_duration") {
				if err := db.Migrator().DropIndex(&models.Duration{}, "idx_time_duration"); err != nil {
					return err
				}
			}

			setHasRun(name, db)
			return nil
		},
	}

	registerPostMigration(f)
}
