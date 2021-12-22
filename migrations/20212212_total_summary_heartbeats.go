package migrations

import (
	"database/sql"
	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

func init() {
	const name = "20212212-total_summary_heartbeats"
	f := migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			logbuch.Info("this may take a while!")

			// this turns out to actually be way faster than using joins and instead has the benefit of being cross-dialect compatible

			var summaries []*models.Summary
			if err := db.Model(&models.Summary{}).
				Select("id, from_time, to_time, user_id").
				Scan(&summaries).Error; err != nil {
				return err
			}

			tx := db.Begin()
			for _, s := range summaries {
				query := "UPDATE summaries SET num_heartbeats = (SELECT count(id) AS num_heartbeats FROM heartbeats WHERE user_id = @user AND time BETWEEN @from AND @to) WHERE id = @id"
				tx.Exec(query, sql.Named("from", s.FromTime), sql.Named("to", s.ToTime), sql.Named("id", s.ID), sql.Named("user", s.UserID))
			}
			if err := tx.Commit().Error; err != nil {
				tx.Rollback()
				logbuch.Error("failed to retroactively determine total summary heartbeats")
				return err
			}

			setHasRun(name, db)
			return nil
		},
	}

	registerPostMigration(f)
}
