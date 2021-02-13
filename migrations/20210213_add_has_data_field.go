package migrations

import (
	"github.com/muety/wakapi/config"
	"gorm.io/gorm"
)

func init() {
	f := migrationFunc{
		name: "20210213_add_has_data_field",
		f: func(db *gorm.DB, cfg *config.Config) error {
			if err := db.Exec("UPDATE users SET has_data = TRUE WHERE 1").Error; err != nil {
				return err
			}
			return nil
		},
	}

	registerPostMigration(f)
}
