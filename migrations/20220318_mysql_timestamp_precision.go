package migrations

import (
	"github.com/muety/wakapi/config"
	"gorm.io/gorm"
	"log/slog"
)

func init() {
	const name = "20220318-mysql_timestamp_precision"
	f := migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			if cfg.Db.IsMySQL() {
				slog.Info("altering heartbeats table, this may take a while (up to hours)")

				db.Exec("SET foreign_key_checks=0;")
				db.Exec("SET unique_checks=0;")
				if err := db.Exec("ALTER TABLE heartbeats MODIFY COLUMN `time` TIMESTAMP(3) NOT NULL").Error; err != nil {
					return err
				}
				if err := db.Exec("ALTER TABLE heartbeats MODIFY COLUMN `created_at` TIMESTAMP(3) NOT NULL").Error; err != nil {
					return err
				}
				db.Exec("SET foreign_key_checks=1;")
				db.Exec("SET unique_checks=1;")

				slog.Info("migrated timestamp columns to millisecond precision")
			}

			setHasRun(name, db)
			return nil
		},
	}

	registerPostMigration(f)
}
