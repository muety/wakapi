package models

import "time"

type Config struct {
	Port                int
	DbHost              string
	DbUser              string
	DbPassword          string
	DbName              string
	DbDialect           string
	AggregationInterval time.Duration
}
