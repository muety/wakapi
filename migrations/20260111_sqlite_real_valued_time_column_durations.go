package migrations

import (
	"fmt"
	"strings"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

// https://github.com/muety/wakapi/issues/882

func init() {
	const name = "20260111-sqlite_real_valued_time_column_durations"

	f := migrationFunc{
		name:       name,
		background: false,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			if !cfg.Db.IsSQLite() {
				return nil
			}

			if err := db.Transaction(func(tx *gorm.DB) error {
				// drop the indexes to allow them be recreated by next auto-migrate
				indexes, err := tx.Migrator().GetIndexes(&models.Duration{})
				if err != nil {
					return err
				}
				for _, index := range indexes {
					if err := tx.Migrator().DropIndex(&models.Duration{}, index.Name()); err != nil {
						return err
					}
				}

				// create identical copy of the durations table
				var createDdl string
				if err := tx.Raw("SELECT sql FROM sqlite_master WHERE type = 'table' AND name = 'durations'").Scan(&createDdl).Error; err != nil {
					return err
				}
				createDdl = strings.ToLower(createDdl)
				createDdl = strings.Replace(createDdl, "create table \"durations\"", "create table \"durations_new\"", 1)
				createDdl = strings.Replace(createDdl, "create table durations", "create table \"durations_new\"", 1)
				createDdl = strings.Replace(createDdl, "create table `durations`", "create table \"durations_new\"", 1)
				if idx := strings.LastIndex(createDdl, ")"); idx != -1 { // inject new column definition
					createDdl = createDdl[:idx] + ", time_real real as (julianday(time)) stored" + createDdl[idx:]
				} else {
					return fmt.Errorf("could not modify ddl for durations table")
				}
				if err := tx.Exec(createDdl).Error; err != nil {
					return err
				}

				// copy data dynamically
				var columns []string
				type colInfo struct{ Name string }
				var info []colInfo
				if err := tx.Raw("pragma table_info(durations)").Scan(&info).Error; err != nil {
					return err
				}
				for _, c := range info {
					columns = append(columns, c.Name)
				}
				quotedCols := "\"" + strings.Join(columns, "\", \"") + "\""
				if err := tx.Exec(fmt.Sprintf("insert into durations_new (%s) select %s from durations", quotedCols, quotedCols)).Error; err != nil {
					return err
				}

				// swap tables
				if err := tx.Exec("drop table durations").Error; err != nil {
					return err
				}
				if err := tx.Exec("alter table durations_new rename to durations").Error; err != nil {
					return err
				}

				// add new indexes for real-valued time column
				if err := tx.Exec("create index idx_time_real_duration on durations(time_real)").Error; err != nil {
					return err
				}
				if err := tx.Exec("create index idx_time_real_user_duration on durations(user_id, time_real)").Error; err != nil {
					return err
				}

				// auto-migrate to recreate all other indexes and constraints
				if err := tx.Migrator().AutoMigrate(&models.Duration{}); err != nil {
					return err
				}

				return nil
			}); err != nil {
				return err
			}

			setHasRun(name, db)
			return nil
		},
	}

	registerPostMigration(f)
}
