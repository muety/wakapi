package utils

import (
	"log"
	"os"
)

func LookupFatal(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("missing env variable '%s'", key)
	}
	return v
}
