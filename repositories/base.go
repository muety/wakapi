package repositories

import (
	"database/sql"
	"errors"
	"github.com/duke-git/lancet/v2/slice"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
)

const chunkSize = 4096

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

func InsertBatchChunked[T any](data []T, model T, db *gorm.DB) error {
	// insert in chunks, because otherwise mysql will complain about too many placeholders in prepared query
	return db.Transaction(func(tx *gorm.DB) error {
		chunks := slice.Chunk[T](data, chunkSize)
		for _, chunk := range chunks {
			if err := insertBatch[T](chunk, model, db); err != nil {
				return err
			}
		}
		return nil
	})
}

func insertBatch[T any](data []T, model T, db *gorm.DB) error {
	// sqlserver on conflict has bug https://github.com/go-gorm/sqlserver/issues/100
	// As a workaround, insert one by one, and ignore duplicate key error
	if db.Dialector.Name() == (sqlserver.Dialector{}).Name() {
		for _, h := range data {
			err := db.Create(h).Error
			if err != nil {
				if strings.Contains(err.Error(), "Cannot insert duplicate key row in object") {
					// ignored
				} else {
					return err
				}
			}
		}
		return nil
	}

	if err := db.
		Clauses(clause.OnConflict{
			DoNothing: true,
		}).
		Model(model).
		Create(&data).Error; err != nil {
		return err
	}
	return nil
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
