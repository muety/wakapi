package migrations

import (
	"github.com/muety/wakapi/config"
	"gorm.io/gorm"
)

func init() {
	const name = "20210213-add_has_data_field"
	f := migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			statement := "UPDATE users SET has_data = TRUE"

			if cfg.Db.IsMssql() {
				statement = "UPDATE users SET has_data = 1"
			}

			if err := db.Exec(statement).Error; err != nil {
				return err
			}

			setHasRun(name, db)
			return nil
		},
	}

	registerPostMigration(f)
}
