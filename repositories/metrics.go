package repositories

import (
	"github.com/muety/wakapi/config"
	"gorm.io/gorm"
)

type MetricsRepository struct {
	BaseRepository
	config *config.Config
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
	return &MetricsRepository{BaseRepository: NewBaseRepository(db), config: config.Get()}
}

func (srv *MetricsRepository) GetDatabaseSize() (size int64, err error) {
	cfg := srv.config.Db

	query := srv.db.Raw("SELECT 0")
	if cfg.IsMySQL() {
		query = srv.db.Raw(sizeTplMysql, cfg.Name)
	} else if cfg.IsPostgres() {
		query = srv.db.Raw(sizeTplPostgres, cfg.Name)
	} else if cfg.IsSQLite() {
		query = srv.db.Raw(sizeTplSqlite)
	}

	err = query.Scan(&size).Error
	return size, err
}
