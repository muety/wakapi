package models

import (
	"testing"
	"time"

	"github.com/muety/wakapi/config"
	"github.com/stretchr/testify/assert"
)

func TestCustomTime_Value_SQLite(t *testing.T) {
	config.Set(config.Empty())
	config.Get().Db.Dialect = config.SQLDialectSqlite

	now := time.Date(2026, 8, 27, 14, 30, 0, 123000000, time.UTC)
	ct := CustomTime(now)

	val, err := ct.Value()
	assert.NoError(t, err)
	assert.IsType(t, int64(0), val)
	assert.Equal(t, now.UnixMilli(), val)
}

func TestCustomTime_Value_NonSQLite(t *testing.T) {
	for _, dialect := range []string{config.SQLDialectPostgres, config.SQLDialectMysql} {
		config.Set(config.Empty())
		config.Get().Db.Dialect = dialect

		now := time.Date(2026, 8, 27, 14, 30, 0, 123000000, time.UTC)
		ct := CustomTime(now)

		val, err := ct.Value()
		assert.NoError(t, err)
		assert.IsType(t, time.Time{}, val)
		assert.Equal(t, now.Round(time.Millisecond), val)
	}
}

func TestCustomTime_Scan(t *testing.T) {
	config.Set(config.Empty())
	config.Get().Db.Dialect = config.SQLDialectSqlite

	refTime := time.Date(2026, 8, 27, 14, 30, 0, 123000000, time.Local)
	refMillis := refTime.UnixMilli()

	// int64
	var ct1 CustomTime
	err := ct1.Scan(refMillis)
	assert.NoError(t, err)
	assert.Equal(t, refTime.Round(time.Millisecond).UnixMilli(), ct1.T().UnixMilli())

	// float64
	var ct2 CustomTime
	err = ct2.Scan(float64(refMillis))
	assert.NoError(t, err)
	assert.Equal(t, refTime.Round(time.Millisecond).UnixMilli(), ct2.T().UnixMilli())

	// string RFC3339Nano
	var ct3 CustomTime
	err = ct3.Scan(refTime.Format(time.RFC3339Nano))
	assert.NoError(t, err)
	assert.Equal(t, refTime.Round(time.Millisecond).UnixMilli(), ct3.T().UnixMilli())

	// time.Time (Local)
	var ct4 CustomTime
	err = ct4.Scan(refTime)
	assert.NoError(t, err)
	assert.Equal(t, refTime.Round(time.Millisecond).UnixMilli(), ct4.T().UnixMilli())

	// time.Time (UTC - PostgreSQL behavior for timestamp without timezone -> "postgres hack")
	utcTime := time.Date(2026, 8, 27, 14, 30, 0, 123000000, time.UTC)
	var ct5 CustomTime
	err = ct5.Scan(utcTime)
	assert.NoError(t, err)
	assert.Equal(t, 14, ct5.T().Hour())
	assert.Equal(t, 30, ct5.T().Minute())
	assert.Equal(t, time.Local, ct5.T().Location())
}

func TestResolveDbDialect(t *testing.T) {
	assert.Equal(t, "sqlite3", config.ResolveDbDialect("sqlite"))
	assert.Equal(t, "sqlite3", config.ResolveDbDialect("sqlite3"))
	assert.Equal(t, "postgres", config.ResolveDbDialect("cockroach"))
	assert.Equal(t, "postgres", config.ResolveDbDialect("postgres"))
	assert.Equal(t, "mysql", config.ResolveDbDialect("mariadb"))
	assert.Equal(t, "mysql", config.ResolveDbDialect("mysql"))
}
