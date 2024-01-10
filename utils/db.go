package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/emvi/logbuch"
	"gorm.io/gorm"
)

func IsCleanDB(db *gorm.DB) bool {
	if db.Dialector.Name() == "sqlite" {
		var count int64
		if err := db.Raw("SELECT count(*) from sqlite_master WHERE type = 'table'").Scan(&count).Error; err != nil {
			logbuch.Error("failed to check if database is clean - '%v'", err)
			return false
		}
		return count == 0
	}
	logbuch.Warn("IsCleanDB is not yet implemented for dialect '%s'", db.Dialector.Name())
	return false
}

func HasConstraints(db *gorm.DB) bool {
	if db.Dialector.Name() == "sqlite" {
		var count int64
		if err := db.Raw("SELECT count(*) from sqlite_master WHERE sql LIKE '%CONSTRAINT%'").Scan(&count).Error; err != nil {
			logbuch.Error("failed to check if database has constraints - '%v'", err)
			return false
		}
		return count != 0
	}
	logbuch.Warn("HasForeignKeyConstraints is not yet implemented for dialect '%s'", db.Dialector.Name())
	return false
}

func WhereNullable(query *gorm.DB, col string, val any) *gorm.DB {
	if val == nil || reflect.ValueOf(val).IsNil() {
		return query.Where(fmt.Sprintf("%s is null", col))
	}
	return query.Where(fmt.Sprintf("%s = ?", col), val)
}

func WithPaging(query *gorm.DB, limit, skip int) *gorm.DB {
	if limit >= 0 {
		query = query.Limit(limit)
	}
	if skip >= 0 {
		query = query.Offset(skip)
	}
	return query
}

type stringWriter struct {
	*strings.Builder
}

func (s stringWriter) WriteByte(c byte) error {
	return s.Builder.WriteByte(c)
}

func (s stringWriter) WriteString(str string) (int, error) {
	return s.Builder.WriteString(str)
}

// QuoteDbIdentifier quotes a column name used in a query.
func QuoteDbIdentifier(query *gorm.DB, identifier string) string {

	builder := stringWriter{Builder: &strings.Builder{}}

	query.Dialector.QuoteTo(builder, identifier)

	return builder.Builder.String()
}
