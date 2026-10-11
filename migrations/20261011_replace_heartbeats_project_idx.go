package migrations

import (
	"log/slog"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

// Adds a new index to facilitate lookups by user, project and time, see https://github.com/muety/wakapi/issues/990.
// Drops redundant indexes or such that are now superseded by the new composite one.
// The new index is picked up from the model definition by auto migration, while the old indexes must be explicitly dropped here (AutoMigrate won't remove or change any existing indexes).

func init() {
	const name = "20261011-replace_heartbeats_project_idx"
	idxNames := []string{"idx_project", "idx_user_project", "idx_time_user"}

	f := migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if !db.Migrator().HasTable(&models.Heartbeat{}) {
				return nil
			}

			for _, idxName := range idxNames {
				if db.Migrator().HasIndex(&models.Heartbeat{}, idxName) {
					slog.Info("dropping index as part of migration", "index", idxName, "migration", name)
					if err := db.Migrator().DropIndex(&models.Heartbeat{}, idxName); err != nil {
						return err
					}
				}
			}

			slog.Info("index 'idx_user_project_time' needs to be created, this may take a while")

			return nil
		},
	}

	registerPreMigration(f)
}
