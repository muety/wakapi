package migrations

import (
	"log/slog"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

func init() {
	const name = "20260214-nullify-empty-email"
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

			if !db.Migrator().HasColumn(&models.User{}, "Email") {
				slog.Warn("skipping migration because email column does not exist", "name", name)
				return nil
			}

			slog.Info("running migration", "name", name)

			if err := db.Model(&models.User{}).Where("email = ?", "").Update("email", nil).Error; err != nil {
				slog.Error("failed to update empty email fields", "error", err)
				return err
			}

			// drop old index so it can be recreated as unique
			if db.Migrator().HasIndex(&models.User{}, "idx_user_email") {
				if err := db.Migrator().DropIndex(&models.User{}, "idx_user_email"); err != nil {
					slog.Error("failed to drop index idx_user_email", "error", err)
					return err
				}
			}

			setHasRun(name, db)
			return nil
		},
	}

	registerPreMigration(f)
}
