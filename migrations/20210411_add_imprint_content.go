package migrations

import (
	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func init() {
	const name = "20210411-add_imprint_content"
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

			imprintKv := &models.KeyStringValue{Key: "imprint", Value: "no content here"}
			if err := db.
				Clauses(clause.OnConflict{UpdateAll: false, DoNothing: true}).
				Where(condition, imprintKv.Key).
				Assign(imprintKv).
				Create(imprintKv).Error; err != nil {
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
