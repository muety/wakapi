package common

import (
	"github.com/jinzhu/gorm"
	"github.com/muety/wakapi/config"
	"log"
)

func ApplyFixtures(db *gorm.DB) {
	cfg := config.Get()

	if err := cfg.GetFixturesFunc(cfg.Db.Dialect)(db); err != nil {
		log.Fatal(err)
	}
}
