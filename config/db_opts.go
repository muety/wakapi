package config

import (
	"gorm.io/gorm"
)

type WakapiDBOpts struct {
	dbConfig *dbConfig
}

func GetWakapiDBOpts(dbConfig *dbConfig) *WakapiDBOpts {
	return &WakapiDBOpts{dbConfig: dbConfig}
}

func (opts WakapiDBOpts) Apply(config *gorm.Config) error {
	return nil
}

func (opts WakapiDBOpts) AfterInitialize(db *gorm.DB) error {
	// initial session variables
	if opts.dbConfig.Type == "cockroach" {
		// https://www.cockroachlabs.com/docs/stable/experimental-features.html#alter-column-types
		if err := db.Exec("SET enable_experimental_alter_column_type_general = true;").Error; err != nil {
			return err
		}
	}

	if opts.dbConfig.IsSQLite() {
		if err := db.Exec("PRAGMA foreign_keys = ON;").Error; err != nil {
			return err
		}
	}

	return nil
}
