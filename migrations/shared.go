package migrations

import (
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"gorm.io/gorm"
	"log/slog"
)

func hasRun(name string, db *gorm.DB) bool {
	condition := utils.QuoteSql(db, "%s = ?", "key")

	lookupResult := db.Where(condition, name).First(&models.KeyStringValue{})
	if lookupResult.Error == nil && lookupResult.RowsAffected > 0 {
		slog.Info("no need to migrate", "name", name)
		return true
	}
	return false
}

func setHasRun(name string, db *gorm.DB) {
	if err := db.Create(&models.KeyStringValue{
		Key:   name,
		Value: "done",
	}).Error; err != nil {
		slog.Error("failed to mark migration as run", "name", name, "error", err)
	}
}
