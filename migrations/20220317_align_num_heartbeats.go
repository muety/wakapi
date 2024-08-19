package migrations

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
	"log/slog"
)

func init() {
	const name = "20220317-align_num_heartbeats"
	f := migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			slog.Info("this may take a while!")

			// find all summaries whose num_heartbeats is zero even though they have items
			var faultyIds []uint

			if err := db.Model(&models.Summary{}).
				Distinct("summaries.id").
				Joins("INNER JOIN summary_items ON summaries.num_heartbeats = 0 AND summaries.id = summary_items.summary_id").
				Scan(&faultyIds).Error; err != nil {
				return err
			}

			// update their heartbeats counter
			result := db.
				Table("summaries").
				Where("summaries.id IN ?", faultyIds).
				Update(
					"num_heartbeats",
					db.
						Model(&models.Heartbeat{}).
						Select("COUNT(*)").
						Where("user_id = ?", gorm.Expr("summaries.user_id")).
						Where("time BETWEEN ? AND ?", gorm.Expr("summaries.from_time"), gorm.Expr("summaries.to_time")),
				)

			if err := result.Error; err != nil {
				return err
			}

			slog.Info("corrected heartbeats counter of %d summaries", result.RowsAffected)

			setHasRun(name, db)
			return nil
		},
	}

	registerPostMigration(f)
}
