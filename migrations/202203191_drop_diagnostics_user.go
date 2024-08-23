package migrations

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
	"log/slog"
)

func init() {
	const name = "202203191-drop_diagnostics_user"
	f := migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			migrator := db.Migrator()

			if migrator.HasColumn(&models.Diagnostics{}, "user_id") {
				slog.Info("running migration", "name", name)

				if err := migrator.DropConstraint(&models.Diagnostics{}, "fk_diagnostics_user"); err != nil {
					slog.Warn("failed to drop constraint", "constraint", "fk_diagnostics_user", "error", err)
				}

				if err := migrator.DropColumn(&models.Diagnostics{}, "user_id"); err != nil {
					slog.Warn("failed to drop column", "table", "diagnostics", "column", "user_id", "error", err)
				}
			}

			setHasRun(name, db)
			return nil
		},
	}

	registerPostMigration(f)
}
