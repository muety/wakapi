package repositories

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	conf "github.com/muety/wakapi/config"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const chunkSize = 1024 // 4096 worked fine for mysql, but not for sqlite

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
	if dialector := r.GetDialector(); dialector == "mysql" {
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

func (r *BaseRepository) RunInTx(f func(tx *gorm.DB) error) error {
	return r.db.Transaction(f)
}

func (r *BaseRepository) VacuumOrOptimize() {
	// sqlite and postgres require manual vacuuming regularly to reclaim free storage from deleted records
	// see https://www.postgresql.org/docs/current/sql-vacuum.html and https://www.sqlite.org/lang_vacuum.html
	// mysql (with innodb storage engine) runs a vacuuming-like operation automatically in the background (https://mariadb.com/docs/server/server-usage/storage-engines/innodb/innodb-purge)
	// instead, mysql optionally provides table optimization, that is, a sort of "defragmentation" (https://dev.mysql.com/doc/refman/8.4/en/optimize-table.html)
	// also see https://github.com/muety/wakapi/issues/785
	t0 := time.Now()

	if strings.HasPrefix(r.db.Dialector.Name(), "sqlite") || r.db.Dialector.Name() == "postgres" {
		if err := r.db.Exec("vacuum").Error; err != nil {
			conf.Log().Error("vacuuming failed", "error", err.Error())
			return
		}
		conf.Log().Info("vacuuming done", "time_elapsed", time.Since(t0))
		return
	}

	if r.db.Dialector.Name() == "mysql" {
		tables, err := r.db.Migrator().GetTables()
		if err != nil {
			conf.Log().Error("failed to retrieve mysql table names", "error", err.Error())
			return
		}

		for table := range tables {
			conf.Log().Info("optimizing table", "table", table)
			if err := r.db.Exec("optimize table ?", table).Error; err != nil {
				conf.Log().Error("failed to optimize table", "table", table)
				continue
			}
		}

		conf.Log().Info("table optimizing done", "time_elapsed", time.Since(t0))
		return
	}

	conf.Log().Info("skipping vacuuming or optimization, because running on neither sqlite, nor postgres, nor mysql")
}

func InsertBatchChunked[T any](data []T, model T, db *gorm.DB) error {
	// insert in chunks, because otherwise sqlite (later also mysql) will complain about too many placeholders in prepared query, see https://github.com/muety/wakapi/issues/840
	return db.Transaction(func(tx *gorm.DB) error {
		chunks := slice.Chunk[T](data, chunkSize)
		for _, chunk := range chunks {
			if err := insertBatch[T](chunk, model, tx); err != nil {
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

func streamRowsBatched[T any](rows *sql.Rows, channel chan []*T, db *gorm.DB, batchSize int, onErr func(error)) {
	defer close(channel)
	defer rows.Close()

	buffer := make([]*T, 0, batchSize)

	for rows.Next() {
		var item T
		if err := db.ScanRows(rows, &item); err != nil {
			onErr(err)
			continue
		}

		buffer = append(buffer, &item)

		if len(buffer) == batchSize {
			channel <- buffer
			buffer = make([]*T, 0, batchSize)
		}
	}

	if len(buffer) > 0 {
		channel <- buffer
	}
}

func filteredQuery(q *gorm.DB, filterMap map[string][]string) *gorm.DB {
	for col, vals := range filterMap {
		q = q.Where(col+" in ?", slice.Map[string, string](vals, func(i int, val string) string {
			// query for "unknown" projects, languages, etc.
			if val == "-" {
				return ""
			}
			return val
		}))
	}
	return q
}
