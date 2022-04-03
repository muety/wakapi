package migrations

import (
	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

// migration to fix https://github.com/muety/wakapi/issues/346
// caused by https://github.com/muety/wakapi/blob/2.3.2/migrations/20220319_add_user_project_idx.go in combination with
// the wrongly defined index at https://github.com/muety/wakapi/blob/5aae18e2415d9e620f383f98cd8cbdf39cd99f27/models/heartbeat.go#L18
// and https://github.com/go-gorm/sqlite/issues/87
// -> drop index and let it be auto-created again with properly formatted ddl

func init() {
	const name = "20220403-drop_user_project_idx"
	const idxName = "idx_user_project"

	f := migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if !db.Migrator().HasTable(&models.KeyStringValue{}) || hasRun(name, db) {
				return nil
			}

			if cfg.Db.IsSQLite() && db.Migrator().HasIndex(&models.Heartbeat{}, idxName) {
				logbuch.Info("running migration '%s'", name)
				if err := db.Migrator().DropIndex(&models.Heartbeat{}, idxName); err != nil {
					logbuch.Warn("failed to drop %s", idxName)
				}
			}

			setHasRun(name, db)
			return nil
		},
	}

	registerPreMigration(f)
}
