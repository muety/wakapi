package migrations

import (
	"fmt"

	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"gorm.io/gorm"
)

func hasRun(name string, db *gorm.DB) bool {
	condition := fmt.Sprintf("%s = ?", utils.QuoteDbIdentifier(db, "key"))

	lookupResult := db.Where(condition, name).First(&models.KeyStringValue{})
	if lookupResult.Error == nil && lookupResult.RowsAffected > 0 {
		logbuch.Info("no need to migrate '%s'", name)
		return true
	}
	return false
}

func setHasRun(name string, db *gorm.DB) {
	if err := db.Create(&models.KeyStringValue{
		Key:   name,
		Value: "done",
	}).Error; err != nil {
		logbuch.Error("failed to mark migration %s as run - %v", name, err)
	}
}
