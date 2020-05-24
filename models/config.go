package models

import (
	"encoding/json"
	"github.com/gorilla/securecookie"
	"github.com/joho/godotenv"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

var cfg *Config

type Config struct {
	Env                  string
	Port                 int
	Addr                 string
	BasePath             string
	DbHost               string
	DbPort               uint
	DbUser               string
	DbPassword           string
	DbName               string
	DbDialect            string
	DbMaxConn            uint
	CleanUp              bool
	DefaultUserName      string
	DefaultUserPassword  string
	SecureCookieHashKey  string
	SecureCookieBlockKey string
	CustomLanguages      map[string]string
	LanguageColors       map[string]string
	SecureCookie         *securecookie.SecureCookie
}

func (c *Config) IsDev() bool {
	return c.Env == "dev"
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

func readConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	env := LookupFatal("ENV")
	dbType := LookupFatal("WAKAPI_DB_TYPE")
	dbUser := LookupFatal("WAKAPI_DB_USER")
	dbPassword := LookupFatal("WAKAPI_DB_PASSWORD")
	dbHost := LookupFatal("WAKAPI_DB_HOST")
	dbName := LookupFatal("WAKAPI_DB_NAME")
	dbPortStr := LookupFatal("WAKAPI_DB_PORT")
	defaultUserName := LookupFatal("WAKAPI_DEFAULT_USER_NAME")
	defaultUserPassword := LookupFatal("WAKAPI_DEFAULT_USER_PASSWORD")
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
		Env:                 env,
		Port:                port,
		Addr:                addr,
		BasePath:            basePath,
		DbHost:              dbHost,
		DbPort:              uint(dbPort),
		DbUser:              dbUser,
		DbPassword:          dbPassword,
		DbName:              dbName,
		DbDialect:           dbType,
		DbMaxConn:           dbMaxConn,
		CleanUp:             cleanUp,
		SecureCookie:        secureCookie,
		DefaultUserName:     defaultUserName,
		DefaultUserPassword: defaultUserPassword,
		CustomLanguages:     customLangs,
		LanguageColors:      colors,
	}
}
