package common

import (
	"github.com/jinzhu/gorm"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"log"
)

func MigrateLanguages(db *gorm.DB) {
	cfg := config.Get()

	for k, v := range cfg.App.CustomLanguages {
		result := db.Model(models.Heartbeat{}).
			Where("language = ?", "").
			Where("entity LIKE ?", "%."+k).
			Updates(models.Heartbeat{Language: v})
		if result.Error != nil {
			log.Fatal(result.Error)
		}
		if result.RowsAffected > 0 {
			log.Printf("Migrated %+v rows for custom language %+s.\n", result.RowsAffected, k)
		}
	}
}
