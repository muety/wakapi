package migrations

import (
	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

func init() {
	const name = "20210213-add_has_data_field"
	f := migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			condition := "key = ?"
			if cfg.Db.Dialect == config.SQLDialectMysql {
				condition = "`key` = ?"
			}
			lookupResult := db.Where(condition, name).First(&models.KeyStringValue{})
			if lookupResult.Error == nil && lookupResult.RowsAffected > 0 {
				logbuch.Info("no need to migrate '%s'", name)
				return nil
			}

			if err := db.Exec("UPDATE users SET has_data = TRUE WHERE TRUE").Error; err != nil {
				return err
			}

			if err := db.Create(&models.KeyStringValue{
				Key:   name,
				Value: "done",
			}).Error; err != nil {
				return err
			}

			return nil
		},
	}

	registerPostMigration(f)
}
