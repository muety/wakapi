package migrations

import (
	"github.com/alitto/pond"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"gorm.io/gorm"
)

func init() {
	const name = "20231023-fill_last_branch"
	f := migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			var heartbeats []*models.Heartbeat
			if err := db.Where(&models.Heartbeat{Branch: "<<LAST_BRANCH>>"}).Find(&heartbeats).Error; err != nil {
				return err
			}

			wp := pond.New(utils.AllCPUs(), 0)

			// this is the most inefficient way to perform the update, but i couldn't find a way to do this is a single query
			for _, h := range heartbeats {
				h := h
				wp.Submit(func() {
					var latest models.Heartbeat
					if err := db.
						Where(&models.Heartbeat{UserID: h.UserID, Project: h.Project}).
						Not("branch", "<<LAST_BRANCH>>").
						Where("time < ?", h.Time).
						Order("time desc").
						First(&latest).Error; err != nil {
						return
					}
					db.
						Model(&models.Heartbeat{}).
						Where("id", h.ID).
						Update("branch", latest.Branch)
				})
			}

			wp.StopAndWait()

			setHasRun(name, db)
			return nil
		},
	}

	registerPostMigration(f)
}
