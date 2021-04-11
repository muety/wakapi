package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/emvi/logbuch"
	"github.com/getsentry/sentry-go"
	"github.com/gorilla/securecookie"
	"github.com/jinzhu/configor"
	"github.com/muety/wakapi/data"
	"github.com/muety/wakapi/models"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

	ErrUnauthorized        = "401 unauthorized"
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

const (
	MailProviderSmtp      = "smtp"
	MailProviderMailWhale = "mailwhale"
)

var emailProviders = []string{
	MailProviderSmtp,
	MailProviderMailWhale,
}

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
	Charset  string `default:"utf8mb4" env:"WAKAPI_DB_CHARSET"`
	Type     string `yaml:"dialect" default:"sqlite3" env:"WAKAPI_DB_TYPE"`
	MaxConn  uint   `yaml:"max_conn" default:"2" env:"WAKAPI_DB_MAX_CONNECTIONS"`
	Ssl      bool   `default:"false" env:"WAKAPI_DB_SSL"`
}

type serverConfig struct {
	Port        int    `default:"3000" env:"WAKAPI_PORT"`
	ListenIpV4  string `yaml:"listen_ipv4" default:"127.0.0.1" env:"WAKAPI_LISTEN_IPV4"`
	ListenIpV6  string `yaml:"listen_ipv6" default:"::1" env:"WAKAPI_LISTEN_IPV6"`
	BasePath    string `yaml:"base_path" default:"/" env:"WAKAPI_BASE_PATH"`
	PublicUrl   string `yaml:"public_url" default:"http://localhost:3000" env:"WAKAPI_PUBLIC_URL"`
	TlsCertPath string `yaml:"tls_cert_path" default:"" env:"WAKAPI_TLS_CERT_PATH"`
	TlsKeyPath  string `yaml:"tls_key_path" default:"" env:"WAKAPI_TLS_KEY_PATH"`
}

type sentryConfig struct {
	Dsn                  string  `env:"WAKAPI_SENTRY_DSN"`
	EnableTracing        bool    `yaml:"enable_tracing" env:"WAKAPI_SENTRY_TRACING"`
	SampleRate           float32 `yaml:"sample_rate" default:"0.75" env:"WAKAPI_SENTRY_SAMPLE_RATE"`
	SampleRateHeartbeats float32 `yaml:"sample_rate_heartbeats" default:"0.1" env:"WAKAPI_SENTRY_SAMPLE_RATE_HEARTBEATS"`
}

type mailConfig struct {
	Enabled   bool                 `env:"WAKAPI_MAIL_ENABLED" default:"true"`
	Provider  string               `env:"WAKAPI_MAIL_PROVIDER" default:"smtp"`
	MailWhale *MailwhaleMailConfig `yaml:"mailwhale"`
	Smtp      *SMTPMailConfig      `yaml:"smtp"`
}

type MailwhaleMailConfig struct {
	Url          string `env:"WAKAPI_MAIL_MAILWHALE_URL"`
	ClientId     string `yaml:"client_id" env:"WAKAPI_MAIL_MAILWHALE_CLIENT_ID"`
	ClientSecret string `yaml:"client_secret" env:"WAKAPI_MAIL_MAILWHALE_CLIENT_SECRET"`
}

type SMTPMailConfig struct {
	Host     string `env:"WAKAPI_MAIL_SMTP_HOST"`
	Port     uint   `env:"WAKAPI_MAIL_SMTP_PORT"`
	Username string `env:"WAKAPI_MAIL_SMTP_USER"`
	Password string `env:"WAKAPI_MAIL_SMTP_PASS"`
	TLS      bool   `env:"WAKAPI_MAIL_SMTP_TLS"`
	Sender   string `env:"WAKAPI_MAIL_SMTP_SENDER"`
}

type Config struct {
	Env      string `default:"dev" env:"ENVIRONMENT"`
	Version  string `yaml:"-"`
	App      appConfig
	Security securityConfig
	Db       dbConfig
	Server   serverConfig
	Sentry   sentryConfig
	Mail     mailConfig
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

func (c *serverConfig) GetPublicUrl() string {
	return strings.TrimSuffix(c.PublicUrl, "/")
}

func (c *SMTPMailConfig) ConnStr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func IsDev(env string) bool {
	return env == "dev" || env == "development"
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
	if err := json.Unmarshal(data.ColorsFile, &colors); err != nil {
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

func initSentry(config sentryConfig, debug bool) {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:   config.Dsn,
		Debug: debug,
		TracesSampler: sentry.TracesSamplerFunc(func(ctx sentry.SamplingContext) sentry.Sampled {
			if !config.EnableTracing {
				return sentry.SampledFalse
			}

			hub := sentry.GetHubFromContext(ctx.Span.Context())
			txName := hub.Scope().Transaction()

			if strings.HasPrefix(txName, "GET /assets") || strings.HasPrefix(txName, "GET /api/health") {
				return sentry.SampledFalse
			}
			if txName == "POST /api/heartbeat" {
				return sentry.UniformTracesSampler(config.SampleRateHeartbeats).Sample(ctx)
			}
			return sentry.UniformTracesSampler(config.SampleRate).Sample(ctx)
		}),
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			type principalGetter interface {
				GetPrincipal() *models.User
			}
			if hint.Context != nil {
				if req, ok := hint.Context.Value(sentry.RequestContextKey).(*http.Request); ok {
					if p := req.Context().Value("principal"); p != nil {
						event.User.ID = p.(principalGetter).GetPrincipal().ID
					}
				}
			}
			return event
		},
	}); err != nil {
		logbuch.Fatal("failed to initialized sentry – %v", err)
	}
}

func findString(needle string, haystack []string, defaultVal string) string {
	for _, s := range haystack {
		if s == needle {
			return s
		}
	}
	return defaultVal
}

func Set(config *Config) {
	cfg = config
}

func Get() *Config {
	return cfg
}

func Load(version string) *Config {
	config := &Config{}

	flag.Parse()

	if err := configor.New(&configor.Config{}).Load(config, mustReadConfigLocation()); err != nil {
		logbuch.Fatal("failed to read config: %v", err)
	}

	config.Version = strings.TrimSpace(version)
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

	if config.Sentry.Dsn != "" {
		logbuch.Info("enabling sentry integration")
		initSentry(config.Sentry, config.IsDev())
	}

	if config.Mail.Provider != "" && findString(config.Mail.Provider, emailProviders, "") == "" {
		logbuch.Fatal("unknown mail provider '%s'", config.Mail.Provider)
	}

	Set(config)
	return Get()
}
