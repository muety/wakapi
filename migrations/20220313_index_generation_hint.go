package migrations

import (
	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/config"
	"gorm.io/gorm"
)

func init() {
	const name = "20220313-index_generation_hint"
	f := migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}
			logbuch.Info("please note: the following migrations might take a few minutes, as column types are changed and new indexes are created, have some patience")
			setHasRun(name, db)
			return nil
		},
	}

	registerPreMigration(f)
}
