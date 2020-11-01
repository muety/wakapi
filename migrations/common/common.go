package common

import (
	"github.com/muety/wakapi/config"
	"gorm.io/gorm"
)

type migrationFunc struct {
	f    func(db *gorm.DB, cfg *config.Config) error
	name string
}
