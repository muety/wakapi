package repositories

import (
	"database/sql"
	"errors"
	"gorm.io/gorm"
)

type BaseRepository struct {
	db *gorm.DB
}

func NewBaseRepository(db *gorm.DB) BaseRepository {
	return BaseRepository{db: db}
}

func (r *BaseRepository) GetDialector() string {
	return r.db.Dialector.Name()
}

func (r *BaseRepository) GetTableDDLMysql(tableName string) (result string, err error) {
	if dialector := r.GetDialector(); dialector == "sqlite" || dialector == "sqlite3" {
		err = r.db.Raw("show create table ?", tableName).Scan(&result).Error
	} else {
		err = errors.New("not a mysql database")
	}
	return result, err
}

func (r *BaseRepository) GetTableDDLSqlite(tableName string) (result string, err error) {
	if dialector := r.GetDialector(); dialector == "sqlite" || dialector == "sqlite3" {
		err = r.db.Table("sqlite_master").
			Select("sql").
			Where("type = ?", "table").
			Where("name = ?", tableName).
			Take(&result).Error
	} else {
		err = errors.New("not an sqlite database")
	}
	return result, err
}

func streamRows[T any](rows *sql.Rows, channel chan *T, db *gorm.DB, onErr func(error)) {
	defer close(channel)
	defer rows.Close()
	for rows.Next() {
		var item T
		if err := db.ScanRows(rows, &item); err != nil {
			onErr(err)
			continue
		}
		channel <- &item
	}
}
