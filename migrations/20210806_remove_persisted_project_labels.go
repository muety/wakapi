package migrations

import (
	"fmt"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
	"log/slog"
)

func init() {
	const name = "20210806-remove_persisted_project_labels"
	f := migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			rawDb, err := db.DB()
			if err != nil {
				slog.Error("failed to retrieve raw sql db instance")
				return err
			}
			if _, err := rawDb.Exec(fmt.Sprintf("delete from summary_items where type = %d", models.SummaryLabel)); err != nil {
				slog.Error("failed to delete project label summary items")
				return err
			}
			slog.Info("successfully deleted project label summary items")

			setHasRun(name, db)
			return nil
		},
	}

	registerPostMigration(f)
}
