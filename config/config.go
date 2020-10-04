package config

import (
	"encoding/json"
	"flag"
	"github.com/gorilla/securecookie"
	"github.com/jinzhu/configor"
	"github.com/jinzhu/gorm"
	"github.com/muety/wakapi/models"
	migrate "github.com/rubenv/sql-migrate"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const (
	defaultConfigPath          = "config.yml"
	defaultConfigPathLegacy    = "config.ini"
	defaultEnvConfigPathLegacy = ".env"
)

var cfg *Config

type appConfig struct {
	CleanUp         bool              `default:"false" env:"WAKAPI_CLEANUP"`
	CustomLanguages map[string]string `yaml:"custom_languages"`
	LanguageColors  map[string]string `yaml:"-"`
}

type securityConfig struct {
	// this is actually a pepper (https://en.wikipedia.org/wiki/Pepper_(cryptography))
	PasswordSalt    string                     `yaml:"password_salt" default:"" env:"WAKAPI_PASSWORD_SALT"`
	InsecureCookies bool                       `yaml:"insecure_cookies" default:"false" env:"WAKAPI_INSECURE_COOKIES"`
	SecureCookie    *securecookie.SecureCookie `yaml:"-"`
}

type dbConfig struct {
	Host     string `env:"WAKAPI_DB_HOST"`
	Port     uint   `env:"WAKAPI_DB_PORT"`
	User     string `env:"WAKAPI_DB_USER"`
	Password string `env:"WAKAPI_DB_PASSWORD"`
	Name     string `default:"wakapi_db.db" env:"WAKAPI_DB_NAME"`
	Dialect  string `default:"sqlite3" env:"WAKAPI_DB_TYPE"`
	MaxConn  uint   `yaml:"max_conn" default:"2" env:"WAKAPI_DB_MAX_CONNECTIONS"`
}

type serverConfig struct {
	Port     int    `default:"3000" env:"WAKAPI_PORT"`
	Addr     string `default:"127.0.0.1" env:"WAKAPI_LISTEN_IPV4"`
	BasePath string `yaml:"base_path" default:"/" env:"WAKAPI_BASE_PATH"`
}

type Config struct {
	Env      string `default:"dev" env:"ENVIRONMENT"`
	Version  string `yaml:"-"`
	App      appConfig
	Security securityConfig
	Db       dbConfig
	Server   serverConfig
}

func (c *Config) IsDev() bool {
	return IsDev(c.Env)
}

func (c *Config) GetMigrationFunc(dbDialect string) models.MigrationFunc {
	switch dbDialect {
	case "sqlite3":
		return func(db *gorm.DB) error {
			migrations := &migrate.FileMigrationSource{
				Dir: "migrations/sqlite3",
			}

			migrate.SetIgnoreUnknown(true)
			n, err := migrate.Exec(db.DB(), "sqlite3", migrations, migrate.Up)
			if err != nil {
				return err
			}

			log.Printf("applied %d migrations\n", n)
			return nil
		}
	default:
		return func(db *gorm.DB) error {
			db.AutoMigrate(&models.Alias{})
			db.AutoMigrate(&models.Summary{})
			db.AutoMigrate(&models.SummaryItem{})
			db.AutoMigrate(&models.User{})
			db.AutoMigrate(&models.Heartbeat{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
			db.AutoMigrate(&models.SummaryItem{}).AddForeignKey("summary_id", "summaries(id)", "CASCADE", "CASCADE")
			db.AutoMigrate(&models.KeyStringValue{})
			return nil
		}
	}
}

func (c *Config) GetFixturesFunc(dbDialect string) models.MigrationFunc {
	return func(db *gorm.DB) error {
		migrations := &migrate.FileMigrationSource{
			Dir: "migrations/common/fixtures",
		}

		migrate.SetIgnoreUnknown(true)
		n, err := migrate.Exec(db.DB(), dbDialect, migrations, migrate.Up)
		if err != nil {
			return err
		}

		log.Printf("applied %d fixtures\n", n)
		return nil
	}
}

func IsDev(env string) bool {
	return env == "dev" || env == "development"
}

func readVersion() string {
	file, err := os.Open("version.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	return string(bytes)
}

func readLanguageColors() map[string]string {
	// Read language colors
	// Source: https://raw.githubusercontent.com/ozh/github-colors/master/colors.json
	var colors = make(map[string]string)
	var rawColors map[string]struct {
		Color string `json:"color"`
		Url   string `json:"url"`
	}

	data, err := ioutil.ReadFile("data/colors.json")
	if err != nil {
		log.Fatal(err)
	}

	if err := json.Unmarshal(data, &rawColors); err != nil {
		log.Fatal(err)
	}

	for k, v := range rawColors {
		colors[strings.ToLower(k)] = v.Color
	}

	return colors
}

func mustReadConfigLocation() string {
	var cFlag = flag.String("c", defaultConfigPath, "config file location")

	flag.Parse()

	if _, err := os.Stat(*cFlag); err != nil {
		log.Fatalf("failed to find config file at '%s'\n", *cFlag)
	}

	return *cFlag
}

func Set(config *Config) {
	cfg = config
}

func Get() *Config {
	return cfg
}

func Load() *Config {
	config := &Config{}

	maybeMigrateLegacyConfig()

	if err := configor.New(&configor.Config{}).Load(config, mustReadConfigLocation()); err != nil {
		log.Fatalf("failed to read config: %v\n", err)
	}

	config.Version = readVersion()
	config.App.LanguageColors = readLanguageColors()
	// TODO: Read keys from env, so that users are not logged out every time the server is restarted
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

	Set(config)
	return Get()
}
