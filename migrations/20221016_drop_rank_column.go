package migrations

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
	"log/slog"
)

func init() {
	const name = "20221016-drop_rank_column"
	f := migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			migrator := db.Migrator()

			if migrator.HasTable(&models.LeaderboardItem{}) && migrator.HasColumn(&models.LeaderboardItem{}, "rank") {
				slog.Info("running migration", "name", name)

				if err := migrator.DropColumn(&models.LeaderboardItem{}, "rank"); err != nil {
					slog.Warn("failed to drop column", "column", "rank", "error", err)
				}
			}

			setHasRun(name, db)
			return nil
		},
	}

	registerPostMigration(f)
}
