package migrations

import (
	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

func hasRun(name string, db *gorm.DB) bool {
	condition := "key = ?"
	if config.Get().Db.Dialect == config.SQLDialectMysql {
		condition = "`key` = ?"
	}
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
