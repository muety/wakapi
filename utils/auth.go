package utils

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"regexp"
	"strings"
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
	if len(authHeader) != 2 || authHeader[0] != "Basic" {
		return key, errors.New("failed to extract API key")
	}

	keyBytes, err := base64.StdEncoding.DecodeString(authHeader[1])
	return string(keyBytes), err
}

func ExtractCookieAuth(r *http.Request, config *config.Config) (login *models.Login, err error) {
	cookie, err := r.Cookie(models.AuthCookieKey)
	if err != nil {
		return nil, errors.New("missing authentication")
	}

	if err := config.Security.SecureCookie.Decode(models.AuthCookieKey, cookie.Value, &login); err != nil {
		return nil, errors.New("invalid parameters")
	}

	return login, nil
}

func IsMd5(hash string) bool {
	return md5Regex.Match([]byte(hash))
}

func CompareBcrypt(wanted, actual, pepper string) bool {
	plainPassword := []byte(strings.TrimSpace(actual) + pepper)
	err := bcrypt.CompareHashAndPassword([]byte(wanted), plainPassword)
	return err == nil
}

// deprecated, only here for backwards compatibility
func CompareMd5(wanted, actual, pepper string) bool {
	return HashMd5(actual, pepper) == wanted
}

func HashBcrypt(plain, pepper string) (string, error) {
	plainPepperedPassword := []byte(strings.TrimSpace(plain) + pepper)
	bytes, err := bcrypt.GenerateFromPassword(plainPepperedPassword, bcrypt.DefaultCost)
	if err == nil {
		return string(bytes), nil
	}
	return "", err
}

func HashMd5(plain, pepper string) string {
	plainPepperedPassword := []byte(strings.TrimSpace(plain) + pepper)
	hash := md5.Sum(plainPepperedPassword)
	hashStr := hex.EncodeToString(hash[:])
	return hashStr
}
