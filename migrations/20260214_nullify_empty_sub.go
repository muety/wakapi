package migrations

import (
	"log/slog"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

func init() {
	const name = "20260214-nullify-empty-sub"
	f := migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			if !db.Migrator().HasTable(&models.User{}) {
				slog.Warn("skipping migration because users table does not exist", "name", name)
				return nil
			}
			if !db.Migrator().HasColumn(&models.User{}, "Sub") {
				slog.Warn("skipping migration because sub column does not exist", "name", name)
				return nil
			}

			slog.Info("running migration", "name", name)

			if err := db.Model(&models.User{}).Where("sub = ?", "").Update("sub", nil).Error; err != nil {
				slog.Error("failed to update empty sub fields", "error", err)
				return err
			}

			setHasRun(name, db)
			return nil
		},
	}

	registerPreMigration(f)
}
