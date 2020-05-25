package utils

import (
	"encoding/base64"
	"errors"
	"github.com/muety/wakapi/models"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"regexp"
	"strings"
)

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

func ExtractCookieAuth(r *http.Request, config *models.Config) (login *models.Login, err error) {
	cookie, err := r.Cookie(models.AuthCookieKey)
	if err != nil {
		return nil, errors.New("missing authentication")
	}

	if err := config.SecureCookie.Decode(models.AuthCookieKey, cookie.Value, &login); err != nil {
		return nil, errors.New("invalid parameters")
	}

	return login, nil
}

func CheckPassword(user *models.User, password, salt string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password+salt))
	return err == nil
}

// inplace
func HashPassword(u *models.User, salt string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(u.Password+salt), bcrypt.DefaultCost)
	if err == nil {
		u.Password = string(bytes)
	}
	return err
}
