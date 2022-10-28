package migrations

import (
	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

// due to an error in the model definition, idx_time_user used to only cover 'user_id', but not time column
// if that's the case in the current state of the database, drop the index and let it be recreated by auto migration afterwards
func init() {
	const name = "20221028-fix_heartbeats_time_user_idx"
	f := migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			migrator := db.Migrator()

			indexes, err := migrator.GetIndexes(&models.Heartbeat{})
			if err != nil {
				return err
			}

			for _, idx := range indexes {
				if idx.Table() == "heartbeats" && idx.Name() == "idx_time_user" {
					if len(idx.Columns()) == 1 {
						if err := migrator.DropIndex(&models.Heartbeat{}, "idx_time_user"); err != nil {
							return err
						}
						logbuch.Info("index 'idx_time_user' needs to be recreated, this may take a while")
						return nil
					}
				}
			}

			return nil
		},
	}

	registerPreMigration(f)
}
