package utils

import (
	"encoding/base64"
	"errors"
	"net/http"
	"regexp"
	"runtime"
	"strings"

	"github.com/alexedwards/argon2id"
	"github.com/duke-git/lancet/v2/mathutil"
	"golang.org/x/crypto/bcrypt"
)

var md5Regex = regexp.MustCompile(`^[a-f0-9]{32}$`)

func ExtractBasicAuth(r *http.Request) (username, password string, err error) {
	authHeader := strings.Split(r.Header.Get("Authorization"), " ")
	if len(authHeader) != 2 || authHeader[0] != "Basic" {
		return username, password, errors.New("failed to extract API key")
	}

	hash, err := base64.StdEncoding.DecodeString(authHeader[1])
	userKey := strings.TrimSpace(string(hash))
	if err != nil {
		return username, password, err
	}

	re := regexp.MustCompile(`^(.+):(.+)$`)
	groups := re.FindAllStringSubmatch(userKey, -1)
	if len(groups) == 0 || len(groups[0]) != 3 {
		return username, password, errors.New("failed to parse user agent string")
	}
	username, password = groups[0][1], groups[0][2]
	return username, password, err
}

func ExtractBearerAuth(r *http.Request) (key string, err error) {
	authHeader := strings.Split(r.Header.Get("Authorization"), " ")
	if len(authHeader) != 2 || (authHeader[0] != "Basic" && authHeader[0] != "Bearer") {
		return key, errors.New("failed to extract API key")
	}

	keyBytes, err := base64.StdEncoding.DecodeString(authHeader[1])
	return string(keyBytes), err
}

// password hashing

func ComparePassword(hashed, plain, pepper string) bool {
	if hashed[0:10] == "$argon2id$" {
		return CompareArgon2Id(hashed, plain, pepper)
	}
	return CompareBcrypt(hashed, plain, pepper)
}

func HashPassword(plain, pepper string) (string, error) {
	return HashArgon2Id(plain, pepper)
}

func CompareBcrypt(hashed, plain, pepper string) bool {
	plainPepperedPassword := []byte(strings.TrimSpace(plain) + pepper)
	err := bcrypt.CompareHashAndPassword([]byte(hashed), plainPepperedPassword)
	return err == nil
}

func HashBcrypt(plain, pepper string) (string, error) {
	plainPepperedPassword := []byte(strings.TrimSpace(plain) + pepper)
	bytes, err := bcrypt.GenerateFromPassword(plainPepperedPassword, bcrypt.DefaultCost)
	if err == nil {
		return string(bytes), nil
	}
	return "", err
}

func CompareArgon2Id(hashed, plain, pepper string) bool {
	plainPepperedPassword := strings.TrimSpace(plain) + pepper
	match, err := argon2id.ComparePasswordAndHash(plainPepperedPassword, hashed)
	return err == nil && match
}

func HashArgon2Id(plain, pepper string) (string, error) {
	plainPepperedPassword := strings.TrimSpace(plain) + pepper
	params := *argon2id.DefaultParams
	if params.Parallelism == 0 { // https://github.com/muety/wakapi/issues/866
		params.Parallelism = uint8(mathutil.Min[int](runtime.NumCPU(), 255))
	}
	hash, err := argon2id.CreateHash(plainPepperedPassword, &params)
	if err == nil {
		return hash, nil
	}
	return "", err
}
