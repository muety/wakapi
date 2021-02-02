package migrations

import (
	"github.com/muety/wakapi/config"
	"gorm.io/gorm"
)

func init() {
	f := migrationFunc{
		name: "000-apply_fixtures",
		f: func(db *gorm.DB, cfg *config.Config) error {
			return cfg.GetFixturesFunc(cfg.Db.Dialect)(db)
		},
	}

	registerPostMigration(f)
}
