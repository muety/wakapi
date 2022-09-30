package migrations

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

func init() {
	const name = "20220930-drop_heartbeats_entity_idx"
	const idxName = "idx_entity"

	f := migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if !db.Migrator().HasTable(&models.Heartbeat{}) || !db.Migrator().HasIndex(&models.Heartbeat{}, idxName) {
				return nil
			}
			return db.Migrator().DropIndex(&models.Heartbeat{}, idxName)
		},
	}

	registerPreMigration(f)
}
