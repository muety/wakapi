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
	const name = "20260111-sqlite_real_valued_time_column_heartbeats"

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
				// drop view to recreate it later, then referencing the new column
				var viewExists int
				if err := tx.Raw("select count(*) from sqlite_master where type = 'view' and name = 'user_heartbeats_range';").Scan(&viewExists).
					Error; err != nil {
					return err
				}
				if viewExists == 1 {
					if err := tx.Migrator().DropView("user_heartbeats_range"); err != nil {
						return err
					}
				}

				// drop the indexes to allow them be recreated by next auto-migrate
				indexes, err := tx.Migrator().GetIndexes(&models.Heartbeat{})
				if err != nil {
					return err
				}
				for _, index := range indexes {
					if err := tx.Migrator().DropIndex(&models.Heartbeat{}, index.Name()); err != nil {
						return err
					}
				}

				// create identical copy of the heartbeats table
				var createDdl string
				if err := tx.Raw("SELECT sql FROM sqlite_master WHERE type = 'table' AND name = 'heartbeats'").Scan(&createDdl).Error; err != nil {
					return err
				}
				createDdl = strings.ToLower(createDdl)
				createDdl = strings.Replace(createDdl, "create table \"heartbeats\"", "create table \"heartbeats_new\"", 1)
				createDdl = strings.Replace(createDdl, "create table heartbeats", "create table \"heartbeats_new\"", 1)
				createDdl = strings.Replace(createDdl, "create table `heartbeats`", "create table \"heartbeats_new\"", 1)
				if idx := strings.Index(createDdl, "constraint"); idx != -1 { // inject new column definition
					createDdl = createDdl[:idx] + " time_real real as (julianday(time)) stored, " + createDdl[idx:]
				} else {
					return fmt.Errorf("could not modify ddl for heartbeats table")
				}
				if err := tx.Exec(createDdl).Error; err != nil {
					return err
				}

				// copy data dynamically
				var columns []string
				type colInfo struct{ Name string }
				var info []colInfo
				if err := tx.Raw("pragma table_info(heartbeats)").Scan(&info).Error; err != nil {
					return err
				}
				for _, c := range info {
					columns = append(columns, c.Name)
				}
				quotedCols := "\"" + strings.Join(columns, "\", \"") + "\""
				if err := tx.Exec(fmt.Sprintf("insert into heartbeats_new (%s) select %s from heartbeats", quotedCols, quotedCols)).Error; err != nil {
					return err
				}

				// swap tables
				if err := tx.Exec("drop table heartbeats").Error; err != nil {
					return err
				}
				if err := tx.Exec("alter table heartbeats_new rename to heartbeats").Error; err != nil {
					return err
				}

				// recreate view, involving new column now
				const viewDdl = "select u.id as user_id, concat(datetime(min(h.time_real)), '+00:00') as first, concat(datetime(max(h.time_real)), '+00:00') as last " +
					"from users u left join heartbeats h on u.id = h.user_id " +
					"group by u.id"
				if err := tx.Migrator().CreateView("user_heartbeats_range", gorm.ViewOption{
					Query:   db.Raw(viewDdl),
					Replace: !cfg.Db.IsSQLite(),
				}); err != nil {
					return err
				}

				// add new indexes for real-valued time column
				if err := tx.Exec("create index idx_time_real on heartbeats(time_real)").Error; err != nil {
					return err
				}
				if err := tx.Exec("create index idx_time_real_user on heartbeats(user_id, time_real)").Error; err != nil {
					return err
				}

				// auto-migrate to recreate all other indexes and constraints
				if err := tx.Migrator().AutoMigrate(&models.Heartbeat{}); err != nil {
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
