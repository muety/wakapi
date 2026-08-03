package migrations

import (
	"fmt"
	"strings"

	"github.com/muety/wakapi/config"
	"gorm.io/gorm"
)

// Converts the time-column from TEXT (ISO 8601) to INTEGER (Unix epoch ms) on SQLite, enabling correct timezone-agnostic comparison. Also drops the now unnecessary `time_real` generated column if present.
// See https://github.com/muety/wakapi/issues/882, https://github.com/muety/wakapi/issues/960.

func init() {
	const name = "20260722-sqlite_convert_time_to_integer"
	const postName = "20260722-sqlite_recreate_view_post"

	registerPreMigration(migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			if !cfg.Db.IsSQLite() {
				return nil
			}

			colType1, err1 := getColumnTypeSqlite(db, "heartbeats", "time")
			colType2, err2 := getColumnTypeSqlite(db, "durations", "time")
			if colType1 == "integer" && colType2 == "integer" {
				return nil
			}

			if err1 == nil && err2 != nil {
				if err := backupSqliteDb(cfg.Db.Name); err != nil {
					return fmt.Errorf("failed to backup database before migration: %w", err)
				}
			}

			if err := convertTimeColumn(db, "heartbeats"); err != nil {
				return fmt.Errorf("failed to convert heartbeats: %w", err)
			}

			if err := convertTimeColumn(db, "durations"); err != nil {
				return fmt.Errorf("failed to convert durations: %w", err)
			}

			// drop view temporarily (it will be recreated referencing the new 'time' column)
			if err := db.Exec("DROP VIEW IF EXISTS user_heartbeats_range").Error; err != nil {
				return fmt.Errorf("failed to drop view user_heartbeats_range: %w", err)
			}

			dequeueBackedUpView(db, "user_heartbeats_range") // dequeue the backed up view so the old time_real DDL is not restored in post-migration

			setHasRun(name, db)
			return nil
		},
	})

	registerPostMigration(migrationFunc{
		name: postName,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(postName, db) {
				return nil
			}

			if !cfg.Db.IsSQLite() {
				return nil
			}

			// Recreate view with correct definition (without time_real)
			const viewQ = "CREATE VIEW IF NOT EXISTS user_heartbeats_range AS select u.id as user_id, min(h.time) as first, max(h.time) as last " +
				"from users u left join heartbeats h on u.id = h.user_id " +
				"group by u.id"

			if err := db.Exec(viewQ).Error; err != nil {
				return fmt.Errorf("failed to recreate view user_heartbeats_range: %w", err)
			}

			setHasRun(postName, db)
			return nil
		},
	})
}

type colInfo struct {
	Name    string
	Type    string
	NotNull int
	Default *string
	PK      int
}

func convertTimeColumn(db *gorm.DB, tblName string) error {
	var tableExists int
	if err := db.Raw("select count(*) from sqlite_master where type='table' and name=?", tblName).Scan(&tableExists).Error; err != nil {
		return err
	}
	if tableExists == 0 {
		return nil
	}

	var info []colInfo
	if err := db.Raw(fmt.Sprintf("pragma table_info(%s)", tblName)).Scan(&info).Error; err != nil {
		return err
	}
	if len(info) == 0 {
		return nil
	}

	// check if time column type is already INTEGER
	colType, err := getColumnTypeSqlite(db, tblName, "time")
	if err != nil {
		return err
	}
	if colType == "integer" {
		return nil
	}

	// build CREATE TABLE for the new table without time_real and with time as INTEGER
	var colDefs []string
	var allCols []string
	var selectParts []string
	for _, c := range info {
		if c.Name == "time_real" {
			continue // drop the old generated column
		}

		colDef := fmt.Sprintf(`"%s"`, c.Name)
		if c.Name == "time" {
			colDef += " integer"
		} else if c.Type != "" {
			colDef += " " + c.Type
		}
		if c.NotNull > 0 {
			colDef += " not null"
		}
		if c.Default != nil && *c.Default != "" {
			colDef += " default " + *c.Default
		}
		if c.PK > 0 {
			colDef += " primary key"
		}
		colDefs = append(colDefs, colDef)

		allCols = append(allCols, `"`+c.Name+`"`)
		if c.Name == "time" {
			selectParts = append(selectParts, `cast(round((julianday("time") - 2440587.5) * 86400000) as integer) as "time"`)
		} else {
			selectParts = append(selectParts, `"`+c.Name+`"`)
		}
	}

	createDDL := fmt.Sprintf("CREATE TABLE \"%s_new\" (%s)", tblName, strings.Join(colDefs, ", "))

	// drop existing indexes on the old table (they'd be orphaned after table drop)
	var indexes []struct{ Name string }
	if err := db.Raw("select name from sqlite_master where type='index' and tbl_name=? and name not like 'sqlite_%'", tblName).Scan(&indexes).Error; err != nil {
		return err
	}
	for _, idx := range indexes {
		if err := db.Exec(fmt.Sprintf("drop index if exists \"%s\"", idx.Name)).Error; err != nil {
			return err
		}
	}

	// create new table
	if err := db.Exec(createDDL).Error; err != nil {
		return fmt.Errorf("create table failed: %w", err)
	}

	// copy data with conversion
	colsStr := strings.Join(allCols, ", ")
	selectStr := strings.Join(selectParts, ", ")
	copySQL := fmt.Sprintf("insert into \"%s_new\" (%s) select %s from \"%s\"", tblName, colsStr, selectStr, tblName)
	if err := db.Exec(copySQL).Error; err != nil {
		return fmt.Errorf("data copy failed: %w", err)
	}

	// drop old table, rename new table
	if err := db.Exec(fmt.Sprintf("drop table \"%s\"", tblName)).Error; err != nil {
		return err
	}
	if err := db.Exec(fmt.Sprintf("alter table \"%s_new\" rename to \"%s\"", tblName, tblName)).Error; err != nil {
		return err
	}

	return nil
}
