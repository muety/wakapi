package models

import "github.com/gorilla/securecookie"

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
