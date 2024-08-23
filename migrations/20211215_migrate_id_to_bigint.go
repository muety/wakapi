package migrations

import (
	"github.com/muety/wakapi/config"
	"gorm.io/gorm"
	"log/slog"
)

func init() {
	const name = "20211215-migrate_id_to_bigint-add_has_data_field"
	f := migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			slog.Info("this may take a while!")

			if cfg.Db.IsMySQL() {
				tx := db.Begin()
				if err := tx.Exec("ALTER TABLE heartbeats MODIFY COLUMN id BIGINT UNSIGNED AUTO_INCREMENT").Error; err != nil {
					return err
				}
				if err := tx.Exec("ALTER TABLE summary_items MODIFY COLUMN id BIGINT UNSIGNED AUTO_INCREMENT").Error; err != nil {
					return err
				}
				tx.Commit()
			} else if cfg.Db.IsPostgres() {
				// postgres does not have unsigned data types
				// https://www.postgresql.org/docs/10/datatype-numeric.html
				tx := db.Begin()
				if err := tx.Exec("ALTER TABLE heartbeats ALTER COLUMN id TYPE BIGINT").Error; err != nil {
					return err
				}
				if err := tx.Exec("ALTER TABLE summary_items ALTER COLUMN id TYPE BIGINT").Error; err != nil {
					return err
				}
				tx.Commit()
			} else {
				// sqlite doesn't allow for changing column type easily
				// https://stackoverflow.com/a/2083562/3112139
				slog.Warn("unable to migrate id columns to bigint", "dialect", cfg.Db.Dialect)
			}

			setHasRun(name, db)
			return nil
		},
	}

	registerPostMigration(f)
}
