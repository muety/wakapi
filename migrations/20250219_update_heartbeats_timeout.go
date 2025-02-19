package migrations

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

func init() {
	const name = "20250219-update_heartbeats_timeout"
	f := migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			minTimeout := models.MinHeartbeatsTimeout.Seconds()
			maxTimeout := models.MaxHeartbeatsTimeout.Seconds()
			defaultTimeout := models.DefaultHeartbeatsTimeout.Seconds()
			defaultTimeoutLegacy := models.DefaultHeartbeatsTimeoutLegacy.Seconds()

			db.
				Model(&models.User{}).
				Where("heartbeats_timeout_sec < ?", minTimeout).
				Update("heartbeats_timeout_sec", minTimeout)

			db.
				Model(&models.User{}).
				Where("heartbeats_timeout_sec > ?", maxTimeout).
				Update("heartbeats_timeout_sec", maxTimeout)

			db.
				Model(&models.User{}).
				Where("heartbeats_timeout_sec = ?", defaultTimeoutLegacy).
				Update("heartbeats_timeout_sec", defaultTimeout)

			setHasRun(name, db)
			return nil
		},
	}

	registerPostMigration(f)
}
