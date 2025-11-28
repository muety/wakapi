package config

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/duke-git/lancet/v2/strutil"

	"log/slog"

	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/securecookie"
	"github.com/jinzhu/configor"
	"github.com/muety/wakapi/data"
	"github.com/muety/wakapi/utils"
	"github.com/robfig/cron/v3"
)

const (
	DefaultConfigPath = "config.yml"

	SQLDialectMysql    = "mysql"
	SQLDialectPostgres = "postgres"
	SQLDialectSqlite   = "sqlite3"
	SQLDialectMssql    = "mssql"

	KeyLatestTotalTime              = "latest_total_time"
	KeyLatestTotalUsers             = "latest_total_users"
	KeyLastImport                   = "last_import"            // import attempt
	KeyLastImportSuccess            = "last_successful_import" // last actual successful import
	KeySubscriptionNotificationSent = "sub_reminder"
	KeyNewsbox                      = "newsbox"
	KeyInviteCode                   = "invite"
	KeySharedData                   = "shared_data"

	CookieKeySession               = "wakapi_session"
	CookieKeyAuth                  = "wakapi_auth"
	SessionValueOidcState          = "oidc_state"
	SessionValueOidcIdTokenPayload = "oidc_id_token"

	SimpleDateFormat     = "2006-01-02"
	SimpleDateTimeFormat = "2006-01-02 15:04:05"

	ErrUnauthorized        = "401 unauthorized"
	ErrBadRequest          = "400 bad request"
	ErrNotFound            = "404 not found"
	ErrInternalServerError = "500 internal server error"
)

const (
	WakatimeApiUrl               = "https://api.wakatime.com/api/v1"
	WakatimeApiUserUrl           = "/users/current"
	WakatimeApiAllTimeUrl        = "/users/current/all_time_since_today"
	WakatimeApiHeartbeatsUrl     = "/users/current/heartbeats"
	WakatimeApiHeartbeatsBulkUrl = "/users/current/heartbeats.bulk"
	WakatimeApiUserAgentsUrl     = "/users/current/user_agents"
	WakatimeApiMachineNamesUrl   = "/users/current/machine_names"
	WakatimeApiDataDumpUrl       = "/users/current/data_dumps"
)

const (
	MailProviderSmtp = "smtp"
)

var emailProviders = []string{
	MailProviderSmtp,
}

// first wakatime commit was on this day ;-) so no real heartbeats should exist before
// https://github.com/wakatime/legacy-python-cli/commit/3da94756aa1903c1cca5035803e3f704e818c086
const heartbeatsMinDate = "2013-07-06"
const colorsFile = "data/colors.json"

var leaderboardScopes = []string{"24_hours", "week", "month", "year", "7_days", "14_days", "30_days", "6_months", "12_months", "all_time"}

var appStartTime = time.Now()

var cfg *Config
var env string

type appConfig struct {
	LeaderboardEnabled        bool                         `yaml:"leaderboard_enabled" default:"true" env:"WAKAPI_LEADERBOARD_ENABLED"`
	LeaderboardScope          string                       `yaml:"leaderboard_scope" default:"7_days" env:"WAKAPI_LEADERBOARD_SCOPE"`
	LeaderboardGenerationTime string                       `yaml:"leaderboard_generation_time" default:"0 0 6 * * *,0 0 18 * * *" env:"WAKAPI_LEADERBOARD_GENERATION_TIME"`
	LeaderboardRequireAuth    bool                         `yaml:"leaderboard_require_auth" default:"false" env:"WAKAPI_LEADERBOARD_REQUIRE_AUTH"`
	AggregationTime           string                       `yaml:"aggregation_time" default:"0 15 2 * * *" env:"WAKAPI_AGGREGATION_TIME"`
	ReportTimeWeekly          string                       `yaml:"report_time_weekly" default:"0 0 18 * * 5" env:"WAKAPI_REPORT_TIME_WEEKLY"`
	DataCleanupTime           string                       `yaml:"data_cleanup_time" default:"0 0 6 * * 0" env:"WAKAPI_DATA_CLEANUP_TIME"`
	OptimizeDatabaseTime      string                       `yaml:"optimize_database_time" default:"0 0 8 1 * *" env:"WAKAPI_OPTIMIZE_DATABASE_TIME"`
	ImportEnabled             bool                         `yaml:"import_enabled" default:"true" env:"WAKAPI_IMPORT_ENABLED"`
	ImportBackoffMin          int                          `yaml:"import_backoff_min" default:"5" env:"WAKAPI_IMPORT_BACKOFF_MIN"`
	ImportMaxRate             int                          `yaml:"import_max_rate" default:"24" env:"WAKAPI_IMPORT_MAX_RATE"` // at max one successful import every x hours
	ImportBatchSize           int                          `yaml:"import_batch_size" default:"50" env:"WAKAPI_IMPORT_BATCH_SIZE"`
	InactiveDays              int                          `yaml:"inactive_days" default:"7" env:"WAKAPI_INACTIVE_DAYS"`
	HeartbeatMaxAge           string                       `yaml:"heartbeat_max_age" default:"168h" env:"WAKAPI_HEARTBEAT_MAX_AGE"`
	CountCacheTTLMin          int                          `yaml:"count_cache_ttl_min" default:"30" env:"WAKAPI_COUNT_CACHE_TTL_MIN"`
	DataRetentionMonths       int                          `yaml:"data_retention_months" default:"-1" env:"WAKAPI_DATA_RETENTION_MONTHS"`
	DataCleanupDryRun         bool                         `yaml:"data_cleanup_dry_run" default:"false" env:"WAKAPI_DATA_CLEANUP_DRY_RUN"` // for debugging only
	MaxInactiveMonths         int                          `yaml:"max_inactive_months" default:"-1" env:"WAKAPI_MAX_INACTIVE_MONTHS"`
	WarmCaches                bool                         `yaml:"warm_caches" default:"true" env:"WAKAPI_WARM_CACHES"`
	AvatarURLTemplate         string                       `yaml:"avatar_url_template" default:"api/avatar/{username_hash}.svg" env:"WAKAPI_AVATAR_URL_TEMPLATE"`
	SupportContact            string                       `yaml:"support_contact" default:"hostmaster@wakapi.dev" env:"WAKAPI_SUPPORT_CONTACT"`
	DateFormat                string                       `yaml:"date_format" default:"Mon, 02 Jan 2006" env:"WAKAPI_DATE_FORMAT"`
	DateTimeFormat            string                       `yaml:"datetime_format" default:"Mon, 02 Jan 2006 15:04" env:"WAKAPI_DATETIME_FORMAT"`
	CustomLanguages           map[string]string            `yaml:"custom_languages"`
	CanonicalLanguageNames    map[string]string            `yaml:"canonical_language_names"` // lower case, compacted representation -> canonical name
	Colors                    map[string]map[string]string `yaml:"-"`
}

type securityConfig struct {
	AllowSignup      bool `yaml:"allow_signup" default:"true" env:"WAKAPI_ALLOW_SIGNUP"`
	OidcAllowSignup  bool `yaml:"oidc_allow_signup" default:"true" env:"WAKAPI_OIDC_ALLOW_SIGNUP"`
	SignupCaptcha    bool `yaml:"signup_captcha" default:"false" env:"WAKAPI_SIGNUP_CAPTCHA"`
	InviteCodes      bool `yaml:"invite_codes" default:"true" env:"WAKAPI_INVITE_CODES"`
	ExposeMetrics    bool `yaml:"expose_metrics" default:"false" env:"WAKAPI_EXPOSE_METRICS"`
	EnableProxy      bool `yaml:"enable_proxy" default:"false" env:"WAKAPI_ENABLE_PROXY"` // only intended for production instance at wakapi.dev
	DisableFrontpage bool `yaml:"disable_frontpage" default:"false" env:"WAKAPI_DISABLE_FRONTPAGE"`
	// this is actually a pepper (https://en.wikipedia.org/wiki/Pepper_(cryptography))
	PasswordSalt                 string                     `yaml:"password_salt" default:"" env:"WAKAPI_PASSWORD_SALT"`
	InsecureCookies              bool                       `yaml:"insecure_cookies" default:"false" env:"WAKAPI_INSECURE_COOKIES"`
	CookieMaxAgeSec              int                        `yaml:"cookie_max_age" default:"172800" env:"WAKAPI_COOKIE_MAX_AGE"`
	TrustedHeaderAuth            bool                       `yaml:"trusted_header_auth" default:"false" env:"WAKAPI_TRUSTED_HEADER_AUTH"`
	TrustedHeaderAuthKey         string                     `yaml:"trusted_header_auth_key" default:"Remote-User" env:"WAKAPI_TRUSTED_HEADER_AUTH_KEY"`
	TrustedHeaderAuthAllowSignup bool                       `yaml:"trusted_header_auth_allow_signup" default:"false" env:"WAKAPI_TRUSTED_HEADER_AUTH_ALLOW_SIGNUP"`
	TrustReverseProxyIps         string                     `yaml:"trust_reverse_proxy_ips" default:"" env:"WAKAPI_TRUST_REVERSE_PROXY_IPS"` // comma-separated list of trusted reverse proxy ips
	SignupMaxRate                string                     `yaml:"signup_max_rate" default:"5/1h" env:"WAKAPI_SIGNUP_MAX_RATE"`
	LoginMaxRate                 string                     `yaml:"login_max_rate" default:"10/1m" env:"WAKAPI_LOGIN_MAX_RATE"`
	PasswordResetMaxRate         string                     `yaml:"password_reset_max_rate" default:"5/1h" env:"WAKAPI_PASSWORD_RESET_MAX_RATE"`
	SecureCookie                 *securecookie.SecureCookie `yaml:"-"`
	SessionKey                   []byte                     `yaml:"-"`
	OidcProviders                []oidcProviderConfig       `yaml:"oidc"`
	trustReverseProxyIpsParsed   []net.IPNet
}

type dbConfig struct {
	Host                    string `env:"WAKAPI_DB_HOST"`
	Socket                  string `env:"WAKAPI_DB_SOCKET"`
	Port                    uint   `env:"WAKAPI_DB_PORT"`
	User                    string `env:"WAKAPI_DB_USER"`
	Password                string `env:"WAKAPI_DB_PASSWORD"`
	Name                    string `default:"wakapi_db.db" env:"WAKAPI_DB_NAME"`
	Dialect                 string `yaml:"-"`
	Charset                 string `default:"utf8mb4" env:"WAKAPI_DB_CHARSET"`
	Type                    string `yaml:"dialect" default:"sqlite3" env:"WAKAPI_DB_TYPE"`
	DSN                     string `yaml:"DSN" default:"" env:"WAKAPI_DB_DSN"`
	MaxConn                 uint   `yaml:"max_conn" default:"10" env:"WAKAPI_DB_MAX_CONNECTIONS"`
	Ssl                     bool   `default:"false" env:"WAKAPI_DB_SSL"`
	Compress                bool   `yaml:"compress" default:"false" env:"WAKAPI_DB_COMPRESS"`
	MysqlOptimize           bool   `default:"false" env:"WAKAPI_MYSQL_OPTIMIZE"` // apparently not recommended, because usually has very little effect but takes forever and partially locks the table
	AutoMigrateFailSilently bool   `yaml:"automigrate_fail_silently" default:"false" env:"WAKAPI_DB_AUTOMIGRATE_FAIL_SILENTLY"`
}

type serverConfig struct {
	Port             int    `default:"3000" env:"WAKAPI_PORT"`
	ListenIpV4       string `yaml:"listen_ipv4" default:"127.0.0.1" env:"WAKAPI_LISTEN_IPV4"`
	ListenIpV6       string `yaml:"listen_ipv6" default:"::1" env:"WAKAPI_LISTEN_IPV6"`
	ListenSocket     string `yaml:"listen_socket" default:"" env:"WAKAPI_LISTEN_SOCKET"`
	ListenSocketMode uint32 `yaml:"listen_socket_mode" default:"0666" env:"WAKAPI_LISTEN_SOCKET_MODE"`
	TimeoutSec       int    `yaml:"timeout_sec" default:"30" env:"WAKAPI_TIMEOUT_SEC"`
	BasePath         string `yaml:"base_path" default:"/" env:"WAKAPI_BASE_PATH"`
	PublicUrl        string `yaml:"public_url" default:"http://localhost:3000" env:"WAKAPI_PUBLIC_URL"`
	TlsCertPath      string `yaml:"tls_cert_path" default:"" env:"WAKAPI_TLS_CERT_PATH"`
	TlsKeyPath       string `yaml:"tls_key_path" default:"" env:"WAKAPI_TLS_KEY_PATH"`
}

type subscriptionsConfig struct {
	Enabled              bool   `yaml:"enabled" default:"false" env:"WAKAPI_SUBSCRIPTIONS_ENABLED"`
	ExpiryNotifications  bool   `yaml:"expiry_notifications" default:"true" env:"WAKAPI_SUBSCRIPTIONS_EXPIRY_NOTIFICATIONS"`
	StripeApiKey         string `yaml:"stripe_api_key" env:"WAKAPI_SUBSCRIPTIONS_STRIPE_API_KEY"`
	StripeSecretKey      string `yaml:"stripe_secret_key" env:"WAKAPI_SUBSCRIPTIONS_STRIPE_SECRET_KEY"`
	StripeEndpointSecret string `yaml:"stripe_endpoint_secret" env:"WAKAPI_SUBSCRIPTIONS_STRIPE_ENDPOINT_SECRET"`
	StandardPriceId      string `yaml:"standard_price_id" env:"WAKAPI_SUBSCRIPTIONS_STANDARD_PRICE_ID"`
	StandardPrice        string `yaml:"-"`
}

type sentryConfig struct {
	Dsn                  string  `env:"WAKAPI_SENTRY_DSN"`
	Environment          string  `env:"WAKAPI_SENTRY_ENVIRONMENT"`
	EnableTracing        bool    `yaml:"enable_tracing" env:"WAKAPI_SENTRY_TRACING"`
	SampleRate           float32 `yaml:"sample_rate" default:"0.75" env:"WAKAPI_SENTRY_SAMPLE_RATE"`
	SampleRateHeartbeats float32 `yaml:"sample_rate_heartbeats" default:"0.1" env:"WAKAPI_SENTRY_SAMPLE_RATE_HEARTBEATS"`
}

type mailConfig struct {
	Enabled            bool           `env:"WAKAPI_MAIL_ENABLED" default:"false"`
	Provider           string         `env:"WAKAPI_MAIL_PROVIDER" default:"smtp"`
	Smtp               SMTPMailConfig `yaml:"smtp"`
	Sender             string         `env:"WAKAPI_MAIL_SENDER" yaml:"sender"`
	SkipVerifyMXRecord bool           `yaml:"skip_verify_mx_record" env:"WAKAPI_MAIL_SKIP_VERIFY_MX_RECORD" default:"false"`
}

type SMTPMailConfig struct {
	Host       string `env:"WAKAPI_MAIL_SMTP_HOST"`
	Port       uint   `env:"WAKAPI_MAIL_SMTP_PORT"`
	Username   string `env:"WAKAPI_MAIL_SMTP_USER"`
	Password   string `env:"WAKAPI_MAIL_SMTP_PASS"`
	TLS        bool   `env:"WAKAPI_MAIL_SMTP_TLS"`
	SkipVerify bool   `env:"WAKAPI_MAIL_SMTP_SKIP_VERIFY"`
}

type oidcProviderConfig struct {
	// for environment variables format, see renameEnvVars() down below
	Name         string `yaml:"name"`
	DisplayName  string `yaml:"display_name"` // optional
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	Endpoint     string `yaml:"endpoint"` // base url from which auto-discovery (.well-known/openid-configuration) can be found
}

type Config struct {
	Env            string `default:"dev" env:"ENVIRONMENT"`
	Version        string `yaml:"-"`
	QuickStart     bool   `yaml:"quick_start" env:"WAKAPI_QUICK_START"`
	SkipMigrations bool   `yaml:"skip_migrations" env:"WAKAPI_SKIP_MIGRATIONS"`
	InstanceId     string `yaml:"-"` // only temporary, changes between runs
	EnablePprof    bool   `yaml:"enable_pprof" env:"WAKAPI_ENABLE_PPROF"`
	App            appConfig
	Security       securityConfig
	Db             dbConfig
	Server         serverConfig
	Subscriptions  subscriptionsConfig
	Sentry         sentryConfig
	Mail           mailConfig
}

func (c *oidcProviderConfig) String() string {
	if c.DisplayName != "" {
		return c.DisplayName
	}
	return strutil.Capitalize(c.Name)
}

func (c *oidcProviderConfig) Validate() error {
	var namePattern = regexp.MustCompile("^[a-zA-Z0-9-]+$")
	var endpointPattern = regexp.MustCompile("^https?://")

	if !namePattern.MatchString(c.Name) {
		return fmt.Errorf("invalid provider name '%s', must only contain alphanumeric characters or '-'", c.Name)
	}
	if c.ClientID == "" {
		return fmt.Errorf("provider '%s' is missing client id", c.Name)
	}
	if c.ClientSecret == "" {
		return fmt.Errorf("provider '%s' is missing client secret", c.Name)
	}
	if !endpointPattern.MatchString(c.Endpoint) {
		return fmt.Errorf("provider '%s' is missing endpoint", c.Name)
	}
	return nil
}

func (c *Config) AppStartTimestamp() string {
	return fmt.Sprintf("%d", appStartTime.Unix())
}

func (c *Config) CreateCookie(name, value string) *http.Cookie {
	return c.createCookie(name, value, c.Server.BasePath, c.Security.CookieMaxAgeSec)
}

func (c *Config) GetClearCookie(name string) *http.Cookie {
	return c.createCookie(name, "", c.Server.BasePath, -1)
}

func (c *Config) createCookie(name, value, path string, maxAge int) *http.Cookie {
	if path == "" {
		path = "/"
	}

	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		MaxAge:   maxAge,
		Secure:   !c.Security.InsecureCookies,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

func (c *Config) IsDev() bool {
	return IsDev(c.Env)
}

func (c *Config) UseTLS() bool {
	return c.Server.TlsCertPath != "" && c.Server.TlsKeyPath != ""
}

func (c *appConfig) GetCustomLanguages() map[string]string {
	return utils.CloneStringMap(c.CustomLanguages, false)
}

func (c *appConfig) GetCanonicalLanguageNames() map[string]string {
	if c.CanonicalLanguageNames == nil {
		return make(map[string]string)
	}
	return utils.CloneStringMap(c.CanonicalLanguageNames, false)
}

func (c *appConfig) GetLanguageColors() map[string]string {
	return utils.CloneStringMap(c.Colors["languages"], true)
}

func (c *appConfig) GetEditorColors() map[string]string {
	return utils.CloneStringMap(c.Colors["editors"], true)
}

func (c *appConfig) GetOSColors() map[string]string {
	return utils.CloneStringMap(c.Colors["operating_systems"], true)
}

func (c *appConfig) GetAggregationTimeCron() string {
	if strings.Contains(c.AggregationTime, ":") {
		// old gocron format, e.g. "15:04"
		timeParts := strings.Split(c.AggregationTime, ":")
		h, err := strconv.Atoi(timeParts[0])
		if err != nil {
			Log().Fatal(err.Error())
		}

		m, err := strconv.Atoi(timeParts[1])
		if err != nil {
			Log().Fatal(err.Error())
		}

		return fmt.Sprintf("0 %d %d * * *", m, h)
	}

	return utils.CronPadToSecondly(c.AggregationTime)
}

func (c *appConfig) GetWeeklyReportCron() string {
	if strings.Contains(c.ReportTimeWeekly, ",") {
		// old gocron format, e.g. "fri,18:00"
		split := strings.Split(c.ReportTimeWeekly, ",")
		weekday := utils.ParseWeekday(split[0])
		timeParts := strings.Split(split[1], ":")

		h, err := strconv.Atoi(timeParts[0])
		if err != nil {
			Log().Fatal(err.Error())
		}

		m, err := strconv.Atoi(timeParts[1])
		if err != nil {
			Log().Fatal(err.Error())
		}

		return fmt.Sprintf("0 %d %d * * %d", m, h, weekday)
	}

	return utils.CronPadToSecondly(c.ReportTimeWeekly)
}

func (c *appConfig) GetLeaderboardGenerationTimeCron() []string {
	crons := []string{}

	var parse func(string) string

	if strings.Contains(c.LeaderboardGenerationTime, ":") {
		// old gocron format, e.g. "15:04"
		parse = func(s string) string {
			timeParts := strings.Split(s, ":")
			h, err := strconv.Atoi(timeParts[0])
			if err != nil {
				Log().Fatal(err.Error())
			}

			m, err := strconv.Atoi(timeParts[1])
			if err != nil {
				Log().Fatal(err.Error())
			}

			return fmt.Sprintf("0 %d %d * * *", m, h)
		}
	} else {
		parse = func(s string) string {
			return utils.CronPadToSecondly(s)
		}
	}

	for _, s := range utils.SplitMulti(c.LeaderboardGenerationTime, ",", ";") {
		crons = append(crons, parse(strings.TrimSpace(s)))
	}

	return crons
}

func (c *appConfig) HeartbeatsMaxAge() time.Duration {
	d, _ := time.ParseDuration(c.HeartbeatMaxAge)
	return d
}

func (c *securityConfig) ParseTrustReverseProxyIPs() {
	c.trustReverseProxyIpsParsed = make([]net.IPNet, 0)

	for _, ip := range strings.Split(c.TrustReverseProxyIps, ",") {
		// the config value is empty by default
		if ip == "" {
			continue
		}

		// try parse as address range
		_, parsedIpNet, err := net.ParseCIDR(ip)
		if err == nil {
			c.trustReverseProxyIpsParsed = append(c.trustReverseProxyIpsParsed, *parsedIpNet)
			continue
		}

		// try parse as single ip
		parsedIp := net.ParseIP(strings.TrimSpace(ip))
		if parsedIp != nil {
			ipBits := net.IPv4len * 8
			if parsedIp.To4() == nil {
				ipBits = net.IPv6len * 8
			}
			ipNet := net.IPNet{IP: parsedIp, Mask: net.CIDRMask(ipBits, ipBits)}
			c.trustReverseProxyIpsParsed = append(c.trustReverseProxyIpsParsed, ipNet)
			continue
		}

		slog.Warn("failed to parse reverse proxy ip ranges")
	}
}

func (c *securityConfig) TrustReverseProxyIPs() []net.IPNet {
	return c.trustReverseProxyIpsParsed
}

func (c *securityConfig) GetSignupMaxRate() (int, time.Duration) {
	return c.parseRate(c.SignupMaxRate)
}

func (c *securityConfig) GetLoginMaxRate() (int, time.Duration) {
	return c.parseRate(c.LoginMaxRate)
}

func (c *securityConfig) GetPasswordResetMaxRate() (int, time.Duration) {
	return c.parseRate(c.PasswordResetMaxRate)
}

func (c *securityConfig) GetOidcProvider(name string) (*OidcProvider, error) {
	return GetOidcProvider(name)
}

func (c *securityConfig) ListOidcProviders() []string {
	return slice.Map[oidcProviderConfig, string](c.OidcProviders, func(i int, provider oidcProviderConfig) string {
		return provider.Name
	})
}

func (c *securityConfig) parseRate(rate string) (int, time.Duration) {
	pattern := regexp.MustCompile("(\\d+)/(\\d+)([smh])")
	matches := pattern.FindStringSubmatch(rate)
	if len(matches) != 4 {
		Log().Fatal("failed to parse rate pattern", "rate", rate)
	}

	limit, _ := strconv.Atoi(matches[1])
	window, _ := strconv.Atoi(matches[2])

	var windowScale time.Duration
	switch matches[3] {
	case "s":
		windowScale = time.Second
	case "m":
		windowScale = time.Minute
	case "h":
		windowScale = time.Hour
	}

	return limit, time.Duration(window) * windowScale
}

func (c *dbConfig) IsSQLite() bool {
	return c.Dialect == "sqlite3"
}

func (c *dbConfig) IsMySQL() bool {
	return c.Dialect == "mysql"
}

func (c *dbConfig) IsPostgres() bool {
	return c.Dialect == "postgres"
}

func (c *dbConfig) IsMssql() bool {
	return c.Dialect == SQLDialectMssql
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
	// - https://raw.githubusercontent.com/ozh/github-colors/master/colors.json
	// - https://wakatime.com/colors/operating_systems
	// - https://wakatime.com/colors/editors
	// Extracted from Wakatime website with XPath (see below) and did a bit of regex magic after.
	// - $x('//span[@class="editor-icon tip"]/@data-original-title').map(e => e.nodeValue)
	// - $x('//span[@class="editor-icon tip"]/div[1]/text()').map(e => e.nodeValue)

	raw := data.ColorsFile
	if IsDev(env) {
		if _, err := os.Stat(colorsFile); err == nil {
			raw, _ = os.ReadFile(colorsFile)
		} else {
			Log().Warn("attempted to read colors from local fs in dev mode, but failed", "file", colorsFile)
		}
	}

	var colors = make(map[string]map[string]string)
	if err := json.Unmarshal(raw, &colors); err != nil {
		Log().Fatal(err.Error())
	}

	return colors
}

func resolveDbDialect(dbType string) string {
	if dbType == "cockroach" {
		return "postgres"
	}
	if dbType == "sqlite" {
		return "sqlite3"
	}
	if dbType == "mariadb" {
		return "mysql"
	}
	return dbType
}

func Set(config *Config) {
	cfg = config
}

func Get() *Config {
	return cfg
}

func Load(configFlag string, version string) *Config {
	renameEnvVars()

	config := &Config{}
	if err := configor.New(&configor.Config{ENVPrefix: "WAKAPI"}).Load(config, configFlag); err != nil {
		Log().Fatal("failed to read config", err)
	}

	env = config.Env

	InitLogger(config.IsDev())

	config.Version = strings.TrimSpace(version)
	tagVersionMatch, _ := regexp.MatchString(`\d+\.\d+\.\d+`, config.Version)
	if tagVersionMatch {
		config.Version = "v" + config.Version
	}

	config.InstanceId = uuid.Must(uuid.NewV4()).String()
	config.App.Colors = readColors()
	config.Db.Dialect = resolveDbDialect(config.Db.Type)
	if config.Db.Type == "cockroach" {
		slog.Warn("cockroach is not officially supported, it is strongly recommended to migrate to postgres instead")
	}
	if config.Db.IsMssql() {
		slog.Error("mssql is not supported anymore, sorry")
		os.Exit(1)
	}

	hashKey := securecookie.GenerateRandomKey(64)
	blockKey := securecookie.GenerateRandomKey(32)
	sessionKey := securecookie.GenerateRandomKey(32)

	if IsDev(env) {
		slog.Warn("⚠️ using temporary keys to sign and encrypt cookies in dev mode, make sure to set env to production for real-world use")
		hashKey, blockKey = getTemporarySecureKeys()
		blockKey = hashKey
	}
	if config.Security.InsecureCookies {
		slog.Warn("⚠️ it is strongly advised NOT to use insecure cookies, are you sure about this setting?")
	}

	config.Security.SecureCookie = securecookie.New(hashKey, blockKey)
	config.Security.SessionKey = sessionKey
	config.Security.ParseTrustReverseProxyIPs()

	config.Server.BasePath = strings.TrimSuffix(config.Server.BasePath, "/")

	for k, v := range config.App.CustomLanguages {
		if v == "" {
			config.App.CustomLanguages[k] = "unknown"
		}
	}

	if config.Sentry.Dsn != "" {
		if config.Sentry.Environment == "" {
			config.Sentry.Environment = config.Env
		}
		slog.Info("enabling sentry integration", "environment", config.Sentry.Environment)
		initSentry(config.Sentry, config.IsDev(), config.Version)
	}

	if config.App.DataRetentionMonths <= 0 {
		slog.Info("disabling data retention policy, keeping data forever")
	} else {
		dataRetentionWarning := fmt.Sprintf("⚠️ data retention policy will cause user data older than %d months to be deleted", config.App.DataRetentionMonths)
		if config.Subscriptions.Enabled {
			dataRetentionWarning += " (except for users with active subscriptions)"
		}
		slog.Warn(dataRetentionWarning)
	}

	// some validation checks
	if config.Server.ListenIpV4 == "-" && config.Server.ListenIpV6 == "-" && config.Server.ListenSocket == "" {
		Log().Fatal("either of listen_ipv4 or listen_ipv6 or listen_socket must be set")
	}
	if config.Db.MaxConn < 2 && !config.Db.IsSQLite() {
		Log().Warn("you should use a pool of at least 2 database connections")
	}
	if config.Db.MaxConn > 1 && config.Db.IsSQLite() {
		Log().Warn("with sqlite, only a single connection is supported") // otherwise 'PRAGMA foreign_keys=ON' would somehow have to be set for every connection in the pool
		config.Db.MaxConn = 1
	}
	if config.Mail.Provider != "" && utils.FindString(config.Mail.Provider, emailProviders, "") == "" {
		Log().Fatal("unknown mail provider", "provider", config.Mail.Provider)
	}
	if config.Mail.Enabled && config.Mail.Sender == "" {
		Log().Fatal("mail sender is required")
	}
	if _, err := time.ParseDuration(config.App.HeartbeatMaxAge); err != nil {
		Log().Fatal("invalid duration set for heartbeat_max_age")
	}
	if config.Security.TrustedHeaderAuth && len(config.Security.trustReverseProxyIpsParsed) == 0 {
		config.Security.TrustedHeaderAuth = false
	}
	if d, err := time.Parse(config.App.DateFormat, config.App.DateFormat); err != nil || !d.Equal(time.Date(2006, time.January, 2, 0, 0, 0, 0, d.Location())) {
		Log().Fatal("invalid date format", "format", config.App.DateFormat)
	}
	if d, err := time.Parse(config.App.DateTimeFormat, config.App.DateTimeFormat); err != nil || !d.Equal(time.Date(2006, time.January, 2, 15, 4, 0, 0, d.Location())) {
		Log().Fatal("invalid datetime format", "format", config.App.DateTimeFormat)
	}
	for _, provider := range config.Security.OidcProviders {
		if err := provider.Validate(); err != nil {
			Log().Fatal("invalid oidc provider config", "provider", provider.Name, "error", err)
		}
	}

	cronParser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

	if _, err := cronParser.Parse(config.App.GetWeeklyReportCron()); err != nil {
		Log().Fatal("invalid cron expression for report_time_weekly")
	}
	if _, err := cronParser.Parse(config.App.GetAggregationTimeCron()); err != nil {
		Log().Fatal("invalid cron expression for aggregation_time")
	}
	for _, c := range config.App.GetLeaderboardGenerationTimeCron() {
		if _, err := cronParser.Parse(c); err != nil {
			Log().Fatal("invalid cron expression for leaderboard_generation_time")
		}
	}

	// see models/interval.go
	if !slice.Contain[string](leaderboardScopes, config.App.LeaderboardScope) {
		Log().Fatal("leaderboard scope is not a valid constant")
	}

	// deprecation notices
	if strings.Contains(config.App.AggregationTime, ":") {
		slog.Warn("you're using deprecated syntax for 'aggregation_time', please change it to a valid cron expression")
	}
	if strings.Contains(config.App.ReportTimeWeekly, ":") {
		slog.Warn("you're using deprecated syntax for 'report_time_weekly', please change it to a valid cron expression")
	}
	if strings.Contains(config.App.LeaderboardGenerationTime, ":") {
		slog.Warn("you're using deprecated syntax for 'leaderboard_generation_time', please change it to a semicolon-separated list if valid cron expressions")
	}

	Set(config)

	// post config-load tasks
	initOpenIDConnect(config)

	return Get()
}

func Empty() *Config {
	return &Config{
		App:           appConfig{},
		Security:      securityConfig{},
		Db:            dbConfig{},
		Server:        serverConfig{},
		Subscriptions: subscriptionsConfig{},
		Sentry:        sentryConfig{},
		Mail:          mailConfig{},
	}
}

func BeginningOfWakatime() time.Time {
	t, _ := time.Parse(SimpleDateFormat, heartbeatsMinDate)
	return t
}

func initOpenIDConnect(config *Config) {
	// openid connect
	for _, c := range config.Security.OidcProviders {
		RegisterOidcProvider(&c)
		slog.Info("registered openid connect provider", "provider", c.Name)
	}
}

func renameEnvVars() {
	// Hacky way to get configor to read a slice of structs from environment variables using custom keys.
	// Specifically, for the OpenID Connect providers config, configor would expect variables in this format:
	// > WAKAPI_SECURITY_OIDCPROVIDERS_0_CLIENTID=<client id here>
	// What we want instead (for consistency and beauty), rather is:
	// > WAKAPI_OIDC_0_CLIENT_ID=<client id here>
	// Since configor cannot parse slices with custom keys (see https://github.com/jinzhu/configor/issues/93)
	// and neither allows to specify prefixes via tags (only the entire variable name as "env:"), we simply rename variables from the "Wakapi-style" format to what configor expects.
	// In the long run, we might want to migrate to a different config parser (e.g. https://github.com/knadh/koanf), since configor seems to be dead.
	// Also see https://github.com/muety/wakapi/issues/856.
	var envOidcPrefix = regexp.MustCompile("WAKAPI_OIDC_PROVIDERS_(\\d+)_([A-Z_]+)")

	for _, e := range os.Environ() {
		parts := strings.Split(e, "=")
		k, v := parts[0], parts[1]

		// oidc providers config
		if matches := envOidcPrefix.FindStringSubmatch(k); matches != nil {
			index, _ := strconv.Atoi(matches[1]) // regex already made sure this is a proper integer
			subkey := matches[2]

			if err := os.Setenv(fmt.Sprintf("WAKAPI_SECURITY_OIDCPROVIDERS_%d_%s", index, strings.ReplaceAll(subkey, "_", "")), v); err != nil {
				slog.Error("failed to rename env. variable", "key", k, "value", v, "error", err)
				os.Exit(1)
			}
			os.Unsetenv(k)
		}
	}
}
