package repositories

import (
	"github.com/muety/wakapi/config"
	"gorm.io/gorm"
)

type MetricsRepository struct {
	Config *config.Config
	DB     *gorm.DB
}

const sizeTplMysql = `
SELECT SUM(data_length + index_length)
FROM information_schema.tables
WHERE table_schema = ?
GROUP BY table_schema`

const sizeTplPostgres = `SELECT pg_database_size(?);`

const sizeTplSqlite = `
SELECT page_count * page_size as size
FROM pragma_page_count(), pragma_page_size();`

func NewMetricsRepository(db *gorm.DB) *MetricsRepository {
	return &MetricsRepository{Config: config.Get(), DB: db}
}

func (srv *MetricsRepository) GetDatabaseSize() (size int64, err error) {
	cfg := srv.Config.Db

	query := srv.DB.Raw("SELECT 0")
	if cfg.IsMySQL() {
		query = srv.DB.Raw(sizeTplMysql, cfg.Name)
	} else if cfg.IsPostgres() {
		query = srv.DB.Raw(sizeTplPostgres, cfg.Name)
	} else if cfg.IsSQLite() {
		query = srv.DB.Raw(sizeTplSqlite)
	}

	err = query.Scan(&size).Error
	return size, err
}
