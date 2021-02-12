package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/emvi/logbuch"
	"github.com/gorilla/securecookie"
	"github.com/jinzhu/configor"
	"github.com/markbates/pkger"
	"github.com/muety/wakapi/models"
	migrate "github.com/rubenv/sql-migrate"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

const (
	defaultConfigPath = "config.yml"

	SQLDialectMysql    = "mysql"
	SQLDialectPostgres = "postgres"
	SQLDialectSqlite   = "sqlite3"

	KeyLatestTotalTime  = "latest_total_time"
	KeyLatestTotalUsers = "latest_total_users"
	KeyLastImportImport = "last_import"

	SimpleDateFormat     = "2006-01-02"
	SimpleDateTimeFormat = "2006-01-02 15:04:05"

	ErrInternalServerError = "500 internal server error"
)

const (
	WakatimeApiUrl               = "https://wakatime.com/api/v1"
	WakatimeApiUserUrl           = "/users/current"
	WakatimeApiAllTimeUrl        = "/users/current/all_time_since_today"
	WakatimeApiHeartbeatsUrl     = "/users/current/heartbeats"
	WakatimeApiHeartbeatsBulkUrl = "/users/current/heartbeats.bulk"
	WakatimeApiUserAgentsUrl     = "/users/current/user_agents"
	WakatimeApiMachineNamesUrl   = "/users/current/machine_names"
)

var cfg *Config
var cFlag = flag.String("config", defaultConfigPath, "config file location")

type appConfig struct {
	AggregationTime  string                       `yaml:"aggregation_time" default:"02:15" env:"WAKAPI_AGGREGATION_TIME"`
	ImportBackoffMin int                          `yaml:"import_backoff_min" default:"5" env:"WAKAPI_IMPORT_BACKOFF_MIN"`
	ImportBatchSize  int                          `yaml:"import_batch_size" default:"100" env:"WAKAPI_IMPORT_BATCH_SIZE"`
	InactiveDays     int                          `yaml:"inactive_days" default:"7" env:"WAKAPI_INACTIVE_DAYS"`
	CustomLanguages  map[string]string            `yaml:"custom_languages"`
	Colors           map[string]map[string]string `yaml:"-"`
}

type securityConfig struct {
	AllowSignup   bool `yaml:"allow_signup" default:"true" env:"WAKAPI_ALLOW_SIGNUP"`
	ExposeMetrics bool `yaml:"expose_metrics" default:"false" env:"WAKAPI_EXPOSE_METRICS"`
	// this is actually a pepper (https://en.wikipedia.org/wiki/Pepper_(cryptography))
	PasswordSalt    string                     `yaml:"password_salt" default:"" env:"WAKAPI_PASSWORD_SALT"`
	InsecureCookies bool                       `yaml:"insecure_cookies" default:"false" env:"WAKAPI_INSECURE_COOKIES"`
	CookieMaxAgeSec int                        `yaml:"cookie_max_age" default:"172800" env:"WAKAPI_COOKIE_MAX_AGE"`
	SecureCookie    *securecookie.SecureCookie `yaml:"-"`
}

type dbConfig struct {
	Host     string `env:"WAKAPI_DB_HOST"`
	Port     uint   `env:"WAKAPI_DB_PORT"`
	User     string `env:"WAKAPI_DB_USER"`
	Password string `env:"WAKAPI_DB_PASSWORD"`
	Name     string `default:"wakapi_db.db" env:"WAKAPI_DB_NAME"`
	Dialect  string `yaml:"-"`
	Type     string `yaml:"dialect" default:"sqlite3" env:"WAKAPI_DB_TYPE"`
	MaxConn  uint   `yaml:"max_conn" default:"2" env:"WAKAPI_DB_MAX_CONNECTIONS"`
	Ssl      bool   `default:"false" env:"WAKAPI_DB_SSL"`
}

type serverConfig struct {
	Port        int    `default:"3000" env:"WAKAPI_PORT"`
	ListenIpV4  string `yaml:"listen_ipv4" default:"127.0.0.1" env:"WAKAPI_LISTEN_IPV4"`
	ListenIpV6  string `yaml:"listen_ipv6" default:"::1" env:"WAKAPI_LISTEN_IPV6"`
	BasePath    string `yaml:"base_path" default:"/" env:"WAKAPI_BASE_PATH"`
	TlsCertPath string `yaml:"tls_cert_path" default:"" env:"WAKAPI_TLS_CERT_PATH"`
	TlsKeyPath  string `yaml:"tls_key_path" default:"" env:"WAKAPI_TLS_KEY_PATH"`
}

type Config struct {
	Env      string `default:"dev" env:"ENVIRONMENT"`
	Version  string `yaml:"-"`
	App      appConfig
	Security securityConfig
	Db       dbConfig
	Server   serverConfig
}

func (c *Config) CreateCookie(name, value, path string) *http.Cookie {
	return c.createCookie(name, value, path, c.Security.CookieMaxAgeSec)
}

func (c *Config) GetClearCookie(name, path string) *http.Cookie {
	return c.createCookie(name, "", path, -1)
}

func (c *Config) createCookie(name, value, path string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		MaxAge:   maxAge,
		Secure:   !c.Security.InsecureCookies,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}

func (c *Config) IsDev() bool {
	return IsDev(c.Env)
}

func (c *Config) UseTLS() bool {
	return c.Server.TlsCertPath != "" && c.Server.TlsKeyPath != ""
}

func (c *Config) GetMigrationFunc(dbDialect string) models.MigrationFunc {
	switch dbDialect {
	default:
		return func(db *gorm.DB) error {
			db.AutoMigrate(&models.User{})
			db.AutoMigrate(&models.KeyStringValue{})
			db.AutoMigrate(&models.Alias{})
			db.AutoMigrate(&models.Heartbeat{})
			db.AutoMigrate(&models.Summary{})
			db.AutoMigrate(&models.SummaryItem{})
			db.AutoMigrate(&models.LanguageMapping{})
			return nil
		}
	}
}

func (c *Config) GetFixturesFunc(dbDialect string) models.MigrationFunc {
	return func(db *gorm.DB) error {
		migrations := &migrate.HttpFileSystemMigrationSource{
			FileSystem: pkger.Dir("/migrations"),
		}

		migrate.SetIgnoreUnknown(true)
		sqlDb, _ := db.DB()
		n, err := migrate.Exec(sqlDb, dbDialect, migrations, migrate.Up)
		if err != nil {
			return err
		}

		logbuch.Info("applied %d fixtures", n)
		return nil
	}
}

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
	//location, _ := time.LoadLocation("Local")
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=true&loc=%s&sql_mode=ANSI_QUOTES",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Name,
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

func (c *appConfig) GetCustomLanguages() map[string]string {
	return cloneStringMap(c.CustomLanguages, false)
}

func (c *appConfig) GetLanguageColors() map[string]string {
	return cloneStringMap(c.Colors["languages"], true)
}

func (c *appConfig) GetEditorColors() map[string]string {
	return cloneStringMap(c.Colors["editors"], true)
}

func (c *appConfig) GetOSColors() map[string]string {
	return cloneStringMap(c.Colors["operating_systems"], true)
}

func IsDev(env string) bool {
	return env == "dev" || env == "development"
}

func readVersion() string {
	file, err := pkger.Open("/version.txt")
	if err != nil {
		logbuch.Fatal(err.Error())
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		logbuch.Fatal(err.Error())
	}

	return strings.TrimSpace(string(bytes))
}

func readColors() map[string]map[string]string {
	// Read language colors
	// Source:
	// – https://raw.githubusercontent.com/ozh/github-colors/master/colors.json
	// – https://wakatime.com/colors/operating_systems
	// - https://wakatime.com/colors/editors
	// Extracted from Wakatime website with XPath (see below) and did a bit of regex magic after.
	// – $x('//span[@class="editor-icon tip"]/@data-original-title').map(e => e.nodeValue)
	// – $x('//span[@class="editor-icon tip"]/div[1]/text()').map(e => e.nodeValue)
	var colors = make(map[string]map[string]string)

	file, err := pkger.Open("/data/colors.json")
	if err != nil {
		logbuch.Fatal(err.Error())
	}
	defer file.Close()
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		logbuch.Fatal(err.Error())
	}

	if err := json.Unmarshal(bytes, &colors); err != nil {
		logbuch.Fatal(err.Error())
	}

	return colors
}

func mustReadConfigLocation() string {
	if _, err := os.Stat(*cFlag); err != nil {
		logbuch.Fatal("failed to find config file at '%s'", *cFlag)
	}
	return *cFlag
}

func resolveDbDialect(dbType string) string {
	if dbType == "cockroach" {
		return "postgres"
	}
	return dbType
}

func Set(config *Config) {
	cfg = config
}

func Get() *Config {
	return cfg
}

func Load() *Config {
	config := &Config{}

	flag.Parse()

	if err := configor.New(&configor.Config{}).Load(config, mustReadConfigLocation()); err != nil {
		logbuch.Fatal("failed to read config: %v", err)
	}

	config.Version = readVersion()
	config.App.Colors = readColors()
	config.Db.Dialect = resolveDbDialect(config.Db.Type)
	config.Security.SecureCookie = securecookie.New(
		securecookie.GenerateRandomKey(64),
		securecookie.GenerateRandomKey(32),
	)

	if strings.HasSuffix(config.Server.BasePath, "/") {
		config.Server.BasePath = config.Server.BasePath[:len(config.Server.BasePath)-1]
	}

	for k, v := range config.App.CustomLanguages {
		if v == "" {
			config.App.CustomLanguages[k] = "unknown"
		}
	}

	if config.Server.ListenIpV4 == "" && config.Server.ListenIpV6 == "" {
		logbuch.Fatal("either of listen_ipv4 or listen_ipv6 must be set")
	}

	if config.Db.MaxConn <= 0 {
		logbuch.Fatal("you must allow at least one database connection")
	}

	Set(config)
	return Get()
}
