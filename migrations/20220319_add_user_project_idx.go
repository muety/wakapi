package migrations

import (
	"fmt"
	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

func init() {
	const name = "20220319-add_user_project_idx"
	f := migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			idxName := "idx_user_project"

			if !db.Migrator().HasIndex(&models.Heartbeat{}, idxName) {
				logbuch.Info("running migration '%s'", name)
				if err := db.Exec(fmt.Sprintf("create index %s on heartbeats (user_id, project)", idxName)).Error; err != nil {
					logbuch.Warn("failed to create %s", idxName)
				}
			}

			setHasRun(name, db)
			return nil
		},
	}

	registerPostMigration(f)
}
