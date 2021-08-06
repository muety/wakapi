package migrations

import (
	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

func init() {
	const name = "20210806-remove_persisted_project_labels"
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

			rawDb, err := db.DB()
			if err != nil {
				logbuch.Error("failed to retrieve raw sql db instance")
				return err
			}
			if _, err := rawDb.Exec("delete from summary_items where type = ?", models.SummaryLabel); err != nil {
				logbuch.Error("failed to delete project label summary items")
				return err
			}
			logbuch.Info("successfully deleted project label summary items")

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
