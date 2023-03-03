package config

import (
	"github.com/emvi/logbuch"
	"github.com/gorilla/securecookie"
	"io"
	"os"
	"path/filepath"
)

func getTemporarySecureKeys() (hashKey, blockKey []byte) {
	keyFile := filepath.Join(os.TempDir(), ".wakapi-dev-keys")

	// key file already exists
	if _, err := os.Stat(keyFile); err == nil {
		file, err := os.Open(keyFile)
		if err != nil {
			logbuch.Fatal("failed to open dev keys file, %v", err)
		}
		defer file.Close()

		combinedKey, err := io.ReadAll(file)
		if err != nil {
			logbuch.Fatal("failed to read key from file")
		}
		return combinedKey[:32], combinedKey[32:64]
	}

	// otherwise, generate random keys and save them
	file, err := os.OpenFile(keyFile, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		logbuch.Fatal("failed to open dev keys file, %v", err)
	}
	defer file.Close()

	combinedKey := securecookie.GenerateRandomKey(64)
	if _, err := file.Write(combinedKey); err != nil {
		logbuch.Fatal("failed to write key to file")
	}
	return combinedKey[:32], combinedKey[32:64]
}
