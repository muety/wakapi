package models

import (
	"encoding/json"
	"github.com/gorilla/securecookie"
	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
	migrate "github.com/rubenv/sql-migrate"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

var cfg *Config

type Config struct {
	Env        string
	Version    string
	Port       int
	Addr       string
	BasePath   string
	DbHost     string
	DbPort     uint
	DbUser     string
	DbPassword string
	DbName     string
	DbDialect  string
	DbMaxConn  uint
	CleanUp    bool
	// this is actually a pepper (https://en.wikipedia.org/wiki/Pepper_(cryptography))
	PasswordSalt         string
	SecureCookieHashKey  string
	SecureCookieBlockKey string
	InsecureCookies      bool
	CustomLanguages      map[string]string
	LanguageColors       map[string]string
	SecureCookie         *securecookie.SecureCookie
}

func (c *Config) IsDev() bool {
	return IsDev(c.Env)
}

func (c *Config) GetMigrationFunc(dbDialect string) MigrationFunc {
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
			db.AutoMigrate(&Alias{})
			db.AutoMigrate(&Summary{})
			db.AutoMigrate(&SummaryItem{})
			db.AutoMigrate(&User{})
			db.AutoMigrate(&Heartbeat{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
			db.AutoMigrate(&SummaryItem{}).AddForeignKey("summary_id", "summaries(id)", "CASCADE", "CASCADE")
			db.AutoMigrate(&KeyStringValue{})
			return nil
		}
	}
}

func (c *Config) GetFixturesFunc(dbDialect string) MigrationFunc {
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

func SetConfig(config *Config) {
	cfg = config
}

func GetConfig() *Config {
	return cfg
}

func LookupFatal(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("missing env variable '%s'", key)
	}
	return v
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

func readConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	version := readVersion()

	env := LookupFatal("ENV")
	dbType := LookupFatal("WAKAPI_DB_TYPE")
	dbUser := LookupFatal("WAKAPI_DB_USER")
	dbPassword := LookupFatal("WAKAPI_DB_PASSWORD")
	dbHost := LookupFatal("WAKAPI_DB_HOST")
	dbName := LookupFatal("WAKAPI_DB_NAME")
	dbPortStr := LookupFatal("WAKAPI_DB_PORT")
	passwordSalt := LookupFatal("WAKAPI_PASSWORD_SALT")
	dbPort, err := strconv.Atoi(dbPortStr)

	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Fatalf("Fail to read file: %v", err)
	}

	if dbType == "" {
		dbType = "mysql"
	}

	dbMaxConn := cfg.Section("database").Key("max_connections").MustUint(1)
	addr := cfg.Section("server").Key("listen").MustString("127.0.0.1")
	insecureCookies := IsDev(env) || cfg.Section("server").Key("insecure_cookies").MustBool(false)
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		port = cfg.Section("server").Key("port").MustInt()
	}

	basePath := cfg.Section("server").Key("base_path").MustString("/")
	if strings.HasSuffix(basePath, "/") {
		basePath = basePath[:len(basePath)-1]
	}

	cleanUp := cfg.Section("app").Key("cleanup").MustBool(false)

	// Read custom languages
	customLangs := make(map[string]string)
	languageKeys := cfg.Section("languages").Keys()
	for _, k := range languageKeys {
		customLangs[k.Name()] = k.MustString("unknown")
	}

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

	// TODO: Read keys from env, so that users are not logged out every time the server is restarted
	secureCookie := securecookie.New(
		securecookie.GenerateRandomKey(64),
		securecookie.GenerateRandomKey(32),
	)

	return &Config{
		Env:             env,
		Version:         version,
		Port:            port,
		Addr:            addr,
		BasePath:        basePath,
		DbHost:          dbHost,
		DbPort:          uint(dbPort),
		DbUser:          dbUser,
		DbPassword:      dbPassword,
		DbName:          dbName,
		DbDialect:       dbType,
		DbMaxConn:       dbMaxConn,
		CleanUp:         cleanUp,
		InsecureCookies: insecureCookies,
		SecureCookie:    secureCookie,
		PasswordSalt:    passwordSalt,
		CustomLanguages: customLangs,
		LanguageColors:  colors,
	}
}
