package models

type Config struct {
	Port            int
	Addr            string
	DbHost          string
	DbUser          string
	DbPassword      string
	DbName          string
	DbDialect       string
	CustomLanguages map[string]string
}
