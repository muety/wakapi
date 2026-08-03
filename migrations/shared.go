package migrations

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"gorm.io/gorm"
)

const viewBackupTable = "wakapi_tmp_view_backups"

func hasRun(name string, db *gorm.DB) bool {
	condition := utils.QuoteSql(db, "%s = ?", "key")

	lookupResult := db.Where(condition, name).First(&models.KeyStringValue{})
	if lookupResult.Error == nil && lookupResult.RowsAffected > 0 {
		slog.Info("no need to migrate", "name", name)
		return true
	}
	return false
}

func setHasRun(name string, db *gorm.DB) {
	if err := db.Create(&models.KeyStringValue{
		Key:   name,
		Value: "done",
	}).Error; err != nil {
		slog.Error("failed to mark migration as run", "name", name, "error", err)
	}
}

// save view ddl to temporary table to be restored later
func backupView(db *gorm.DB) error {
	if db.Dialector.Name() != "sqlite" {
		return nil
	}

	if err := db.Exec("CREATE TABLE IF NOT EXISTS " + viewBackupTable + " (name TEXT PRIMARY KEY, sql TEXT)").Error; err != nil {
		return err
	}

	if err := db.Exec("INSERT OR IGNORE INTO " + viewBackupTable + " (name, sql) SELECT name, sql FROM sqlite_master WHERE type = 'view' AND sql IS NOT NULL").Error; err != nil {
		return err
	}

	rows, err := db.Raw("SELECT name FROM sqlite_master WHERE type = 'view' AND sql IS NOT NULL ORDER BY rowid").Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	var viewsToDrop []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		viewsToDrop = append(viewsToDrop, name)
	}
	rows.Close()

	for _, name := range viewsToDrop {
		slog.Info("dropping view temporarily", "view", name)
		if err := db.Migrator().DropView(name); err != nil {
			slog.Error("failed to drop view", "view", name, "error", err)
			return err
		}
	}

	return nil
}

// restore view from backed up ddl
func restoreView(db *gorm.DB) error {
	if db.Dialector.Name() != "sqlite" {
		return nil
	}

	if !db.Migrator().HasTable(viewBackupTable) {
		return nil
	}

	rows, err := db.Raw("SELECT name, sql FROM " + viewBackupTable).Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	type viewBackup struct {
		Name string
		SQL  string
	}
	var backups []viewBackup

	for rows.Next() {
		var b viewBackup
		if err := rows.Scan(&b.Name, &b.SQL); err != nil {
			return err
		}
		backups = append(backups, b)
	}
	rows.Close()

	for _, b := range backups {
		var count int
		if err := db.Raw("SELECT count(*) FROM sqlite_master WHERE type = 'view' AND name = ?", b.Name).Scan(&count).Error; err != nil {
			return err
		}

		if count > 0 {
			slog.Info("view already exists, skipping restoration", "view", b.Name)
			continue
		}

		slog.Info("restoring view", "view", b.Name)
		if err := db.Exec(b.SQL).Error; err != nil {
			slog.Error("failed to restore view", "view", b.Name, "error", err)
			return err
		}
	}

	if err := db.Exec("DROP TABLE " + viewBackupTable).Error; err != nil {
		slog.Error("failed to drop backup table", "table", viewBackupTable, "error", err)
		return err
	}

	return nil
}

// remove backed up view from temporary table so it won't be restored later
func dequeueBackedUpView(db *gorm.DB, name string) error {
	if db.Dialector.Name() != "sqlite" {
		return nil
	}
	return db.Exec("DELETE FROM "+viewBackupTable+" WHERE name = ?", name).Error
}

func getColumnTypeSqlite(db *gorm.DB, tblName, colName string) (string, error) {
	var info []colInfo
	if err := db.Raw(fmt.Sprintf("pragma table_info(%s)", tblName)).Scan(&info).Error; err != nil {
		return "", err
	}
	if len(info) == 0 {
		return "", errors.New("no columns found")
	}

	for _, c := range info {
		if c.Name == colName {
			return strings.ToLower(c.Type), nil
		}
	}

	return "", errors.New("column not found")
}

func backupSqliteDb(dbPath string) error {
	info, err := os.Stat(dbPath)
	if err != nil {
		return fmt.Errorf("failed to stat database file: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("database path is a directory: %s", dbPath)
	}

	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	backupPath := absPath + ".bak." + timestamp + ".gz"

	srcFile, err := os.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(backupPath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer dstFile.Close()

	gzWriter := gzip.NewWriter(dstFile)
	defer gzWriter.Close()

	if _, err := io.Copy(gzWriter, srcFile); err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}

	return nil
}
