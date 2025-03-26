package config

import (
	"fmt"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"net/url"
)

/*
A quick note to myself including some clarifications about time zones.

- There are basically four time zones (at least in case of MySQL): (1) User, (2) Wakapi (host system), (3) MySQL server, (4) MySQL session
- From my understanding, MySQL server tz is only a fallback and can be ignored as long as a connection tz is specified
- All times are currently stored inside TIMESTAMP columns (alternatives would be DATETIME and BIGINT (plain Unix timestamps))
- TIMESTAMP columns, to my understanding, do not keep any time zone information, but only the very time they store
- Setting a `loc` parameter specifies location parsed time.Time objects are interpreted in, however, does not affect the session time zone setting (https://github.com/go-sql-driver/mysql#loc)
- I.e., when not setting `time_zone` in addition, the session time zone will default to the SYSTEM time zone (UTC in case of Docker)
- Session time zone will result in conversions of inserted times from that time zone to UTC (but, again, we don't use that)
- TIMESTAMP only stores a plain time value without tz information and then converts it only for retrieval to whatever tz is set for the session
- E.g., when inserting '2021-04-27 08:26:07' with session tz set to Europe/Berlin and then viewing the database table with UTC tz will return '2021-04-27 06:26:07' instead
- Loc is a parameter of the driver and used in conjunction with parseTime. It results in time.Time objects being converted to that respective zone before storage and converted back to it during retrieval
- E.g., when inserting '2021-04-27 08:26:07 +01:00' with loc set to UTC results in the actual database column to show '2021-04-27 07:26:07'
- A loc value of "Local" is effective the same as setting loc to the value of the TZ environment variable (or system default otherwise)
- Note that if loc is different from "Local", when using models.CustomTime, values are first received in loc zone and then converted to whatever Wakapi's current zone is (e.g. specified through TZ env.), see https://github.com/muety/wakapi/blob/bc2096f4117275d110a84f5b367aa8fdb4bd87ba/models/shared.go#L95.
- Currently, we don't set a session tz (only loc), so the database server will assume it receives dates in its system time. However, as no tz is set when retrieving the values either, they are also going to be returned just as is and as long as `loc=Local` is set properly, they are parsed in Go code with the correct time zone
- In other words, using loc causes the Go driver to take care of zones, while using a session timezone causes MySQL to take care of zones (?)
- As long as the Wakapi server always runs in the same time zone, it will always parse these dates the same way (i.e. as time.Local, Europe/Berlin in case of Wakapi.dev)
- Using TIMESTAMP columns would only become problematic when either data needs to be migrated to a Wakapi instance in a different tz or if two consumers in different tzs were reading and writing to the same table
- It is important to have same `time_zone` and `loc` parameters set when sending and receiving, no matter what it is (writing / reading in 'UTC' will yield same results as writing / reading in 'Europe/Berlin')
- "The session time zone setting affects display and storage of time values that are zone-sensitive. This includes the values displayed by functions such as NOW() or CURTIME(), and values stored in and retrieved from TIMESTAMP columns. Values for TIMESTAMP columns are converted from the session time zone to UTC for storage, and from UTC to the session time zone for retrieval." (https://dev.mysql.com/doc/refman/8.0/en/time-zone-support.html)
- Wakapi always uses time.Local for everything, i.e. all times in the database have to be interpreted with that tz
- New heartbeats are sent with Python-like Unix timestamps, i.e. are absolute points in time as therefore not subject to any kind of tz issues
- E.g. with Wakapi running in Europe/Berlin, 1619379014.7335322 (2021-04-25T19:30:14.733Z (UTC)) will be inserted as 2021-04-25T21:30:14.733+0200 (CEST), but obviously represents the exact same point in time no matter where it originated from
- The reason why we need to explicitly care about tzs in the first place is the fact that user's can request their data within intervals and the results should correspond to their tz
	- Users from California wouldn't have to care about their heartbeats being stored in German time zone
	- However, they DO care when requesting their summaries
	- A request with `?from=2021-04-25` from California (PST / UTC-7) would ideally have to be translated into a database query like `from >= 2021-04-25T00:00:00+0900)`, assuming that Wakapi runs at CEST (UTC+2)
	- This translation comes from either the user explicitly requesting with a specified tz (i.e. sending `from` as ISO8601 / RFC3999) or them having specified a tz in their profile
	- Implicit intervals are tricky, too, as they are generated on the server, but still have to respect the user's tz, as `today` is different for a user in Cali and one in Karlsruhe

For Postgres, the story is even a different one.
- There are `timestamp` and `timestamptz` columns, whereby the former seems to behave very strangely.
- Apparently, when storing dates, they'll just chop off zone information entirely, while for retrieval, all dates are interpreted as UTC
- If Wakapi runs in CEST, '2025-03-25T14:00:00 +02:00' will end up as '2025-03-25T14:00:00' in the database and become '2025-03-25T16:00:00 +02:00' when retrieved back (at least in case of our annoying models.CustomTime, because of https://github.com/muety/wakapi/blob/bc2096f4117275d110a84f5b367aa8fdb4bd87ba/models/shared.go#L95)
- For `timestamptz`, columns don't actually store tz information either (https://github.com/jackc/pgx/issues/520#issuecomment-479692198), but at least allow for correct retrieval. The driver will return dates already in the application's TZ
- Here (https://github.com/go-gorm/gorm/issues/4834) is an interesting discussion on the issue
- According to https://www.cybertec-postgresql.com/en/time-zone-management-in-postgresql/, good practice is to always use timestamptz, leaving conversions to the database itself
*/

func (c *dbConfig) GetDialector() gorm.Dialector {
	switch c.Dialect {
	case SQLDialectMysql:
		return mysql.New(mysql.Config{
			DriverName: c.Dialect,
			DSN:        mysqlConnectionString(c),
		})
	case SQLDialectPostgres:
		return postgres.New(postgres.Config{
			DSN: postgresConnectionString(c),
		})
	case SQLDialectSqlite:
		return sqlite.Open(sqliteConnectionString(c))
	case SQLDialectMssql:
		return sqlserver.Open(mssqlConnectionString(c))
	}

	return nil
}

func mysqlConnectionString(config *dbConfig) string {
	if len(config.DSN) > 0 {
		return config.DSN
	}

	host := fmt.Sprintf("tcp(%s:%d)", config.Host, config.Port)

	if config.Socket != "" {
		host = fmt.Sprintf("unix(%s)", config.Socket)
	}

	return fmt.Sprintf("%s:%s@%s/%s?charset=%s&parseTime=true&loc=%s&sql_mode=ANSI_QUOTES",
		config.User,
		config.Password,
		host,
		config.Name,
		config.Charset,
		"Local",
	)
}

func postgresConnectionString(config *dbConfig) string {
	if len(config.DSN) > 0 {
		return config.DSN
	}

	sslmode := "disable"
	if config.Ssl {
		sslmode = "require"
	}

	// note: passing a `timezone` param here doesn't seem to have any effect, neither with `timestamp`, not for `timestamptz` columns
	// possibly related to https://github.com/go-gorm/postgres/issues/199 ?

	return fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Name,
		config.Password,
		sslmode,
	)
}

func sqliteConnectionString(config *dbConfig) string {
	return config.Name
}

func mssqlConnectionString(config *dbConfig) string {
	query := url.Values{}
	query.Add("database", config.Name)

	if config.Ssl {
		query.Add("encrypt", "true")
	}

	u := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(config.User, config.Password),
		Host:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		RawQuery: query.Encode(),
	}

	return u.String()
}
