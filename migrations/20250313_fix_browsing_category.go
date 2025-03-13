package migrations

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

func init() {
	const name = "20250313-fix_browsing_category"
	f := migrationFunc{
		name:       name,
		background: true,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			db.
				Model(&models.Heartbeat{}).
				Where("category = ?", "").
				Where(db.
					Where("type = ?", "domain").
					Or("type = ?", "url")).
				Update("category", "browsing")

			setHasRun(name, db)
			return nil
		},
	}

	registerPostMigration(f)
}
