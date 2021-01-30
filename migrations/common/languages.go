package common

import (
	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

func MigrateLanguages(db *gorm.DB) {
	cfg := config.Get()

	for k, v := range cfg.App.CustomLanguages {
		result := db.Model(models.Heartbeat{}).
			Where("language = ?", "").
			Where("entity LIKE ?", "%."+k).
			Updates(models.Heartbeat{Language: v})
		if result.Error != nil {
			logbuch.Fatal(result.Error.Error())
		}
		if result.RowsAffected > 0 {
			logbuch.Info("migrated %+v rows for custom language %+s", result.RowsAffected, k)
		}
	}
}
