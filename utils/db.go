package utils

import (
	"github.com/emvi/logbuch"
	"gorm.io/gorm"
)

func IsCleanDB(db *gorm.DB) bool {
	if db.Dialector.Name() == "sqlite" {
		var count int64
		if err := db.Raw("SELECT count(*) from sqlite_master WHERE type = 'table'").Scan(&count).Error; err != nil {
			logbuch.Error("failed to check if database is clean - '%v'", err)
			return false
		}
		return count == 0
	}
	logbuch.Warn("IsCleanDB is not yet implemented for dialect '%s'", db.Dialector.Name())
	return false
}

func HasConstraints(db *gorm.DB) bool {
	if db.Dialector.Name() == "sqlite" {
		var count int64
		if err := db.Raw("SELECT count(*) from sqlite_master WHERE sql LIKE '%CONSTRAINT%'").Scan(&count).Error; err != nil {
			logbuch.Error("failed to check if database has constraints - '%v'", err)
			return false
		}
		return count != 0
	}
	logbuch.Warn("HasForeignKeyConstraints is not yet implemented for dialect '%s'", db.Dialector.Name())
	return false
}
