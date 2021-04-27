package config

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

/*
A quick note to myself including some clarifications about time zones.

- There are basically four time zones (at least in case of MySQL): (1) User, (2) Wakapi (host system), (3) MySQL server, (4) MySQL session
- From my understanding, MySQL server tz is only a fallback and can be ignored as long as a connection tz is specified
- All times are currently stored inside TIMESTAMP columns (alternatives would be DATETIME and BIGINT (plain Unix timestamps))
- TIMESTAMP columns, to my understanding, do not keep any time zone information, but only the very time they store
- Setting a `loc` parameter specifies what location parsed time.Time objects will be in, however, does not affect the session time zone setting (https://github.com/go-sql-driver/mysql#loc)
- I.e., when not setting `time_zone` in addition, the session time zone will probably default to the server time zone (UTC in case of Docker)
- Session time zone will result in conversions of inserted times from that time zone to UTC
- From my understanding, TIMESTAMP only stores a plain time value without tz information and then converts it only for retrieval to whatever tz is set for the session
- E.g., when inserting '2021-04-27 08:26:07' with session tz set to Europe/Berlin and then viewing the database table with UTC tz will return '2021-04-27 06:26:07' instead
- Currently, no session tz is set (only loc), so the database server will assume it receives UTC. However, as no tz is set when retrieving the values either, they are also going to be returned just as is and as long as `loc=Local` is set properly, they are parsed in Go code with the correct time zone
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
	}
	return nil
}

func mysqlConnectionString(config *dbConfig) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=%s&sql_mode=ANSI_QUOTES",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Name,
		config.Charset,
		"Local",
	)
}

func postgresConnectionString(config *dbConfig) string {
	sslmode := "disable"
	if config.Ssl {
		sslmode = "require"
	}

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
