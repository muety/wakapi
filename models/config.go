package models

type Config struct {
	Port            int
	Addr            string
	DbHost          string
	DbPort          uint
	DbUser          string
	DbPassword      string
	DbName          string
	DbDialect       string
	DbMaxConn       uint
	CustomLanguages map[string]string
}
