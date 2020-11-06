package config

import (
	"github.com/joho/godotenv"
	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

func maybeMigrateLegacyConfig() {
	if yes, err := shouldMigrateLegacyConfig(); err != nil {
		log.Fatalf("failed to determine whether to migrate legacy config: %v\n", err)
	} else if yes {
		log.Printf("migrating legacy config (%s, %s) to new format (%s); see https://github.com/muety/wakapi/issues/54\n", defaultConfigPathLegacy, defaultEnvConfigPathLegacy, defaultConfigPath)
		if err := migrateLegacyConfig(); err != nil {
			log.Fatalf("failed to migrate legacy config: %v\n", err)
		}
		log.Printf("config migration successful; please delete %s and %s now\n", defaultConfigPathLegacy, defaultEnvConfigPathLegacy)
	}
}

func shouldMigrateLegacyConfig() (bool, error) {
	if _, err := os.Stat(defaultConfigPath); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return true, err
	}
	return true, nil
}

func migrateLegacyConfig() error {
	// step 1: read envVars file parameters
	envFile, err := os.Open(defaultEnvConfigPathLegacy)
	if err != nil {
		return err
	}
	envVars, err := godotenv.Parse(envFile)
	if err != nil {
		return err
	}

	env := envVars["ENV"]
	dbType := envVars["WAKAPI_DB_TYPE"]
	dbUser := envVars["WAKAPI_DB_USER"]
	dbPassword := envVars["WAKAPI_DB_PASSWORD"]
	dbHost := envVars["WAKAPI_DB_HOST"]
	dbName := envVars["WAKAPI_DB_NAME"]
	dbPortStr := envVars["WAKAPI_DB_PORT"]
	passwordSalt := envVars["WAKAPI_PASSWORD_SALT"]
	dbPort, _ := strconv.Atoi(dbPortStr)

	// step 2: read ini file
	cfg, err := ini.Load(defaultConfigPathLegacy)
	if err != nil {
		return err
	}

	if dbType == "" {
		dbType = "sqlite3"
	}

	dbMaxConn := cfg.Section("database").Key("max_connections").MustUint(2)
	addr := cfg.Section("server").Key("listen").MustString("127.0.0.1")
	insecureCookies := cfg.Section("server").Key("insecure_cookies").MustBool(false)
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		port = cfg.Section("server").Key("port").MustInt()
	}

	basePathEnv, basePathEnvExists := os.LookupEnv("WAKAPI_BASE_PATH")
	basePath := cfg.Section("server").Key("base_path").MustString("/")
	if basePathEnvExists {
		basePath = basePathEnv
	}

	// Read custom languages
	customLangs := make(map[string]string)
	languageKeys := cfg.Section("languages").Keys()
	for _, k := range languageKeys {
		customLangs[k.Name()] = k.MustString("unknown")
	}

	// step 3: instantiate config
	config := &Config{
		Env: env,
		App: appConfig{
			CustomLanguages: customLangs,
		},
		Security: securityConfig{
			PasswordSalt:    passwordSalt,
			InsecureCookies: insecureCookies,
		},
		Db: dbConfig{
			Host:     dbHost,
			Port:     uint(dbPort),
			User:     dbUser,
			Password: dbPassword,
			Name:     dbName,
			Dialect:  dbType,
			MaxConn:  dbMaxConn,
		},
		Server: serverConfig{
			Port:       port,
			ListenIpV4: addr,
			BasePath:   basePath,
		},
	}

	// step 4: serialize to yaml
	yamlConfig, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	// step 5: write file
	if err := ioutil.WriteFile(defaultConfigPath, yamlConfig, 0600); err != nil {
		return err
	}

	return nil
}
