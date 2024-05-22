package utils

import (
	"encoding/base64"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type AuthJwtClaim struct {
	UID string `json:"uid"`
}

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

func ExtractUserIDFromAuthToken(r *http.Request, authSecret string) (key string, err error) {
	token := r.Header.Get("Token")

	if token == "" {
		return key, errors.New("failed to extract API Token from header")
	}

	claims, err := GetTokenClaims(token, authSecret)
	if err != nil {
		return key, errors.New("Invalid JWT token")
	}

	return claims.UID, err
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
	hash, err := argon2id.CreateHash(plainPepperedPassword, argon2id.DefaultParams)
	if err == nil {
		return hash, nil
	}
	return "", err
}

func authenticateToken(tokenString string, jwtSecret string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return claims, errors.Wrap(err, "invalid id token")
	}

	exp, ok := claims["exp"]
	if !ok {
		return claims, errors.Wrap(err, "invalid exp claim")
	}

	expClaim := int64(exp.(float64))
	if time.Now().After(time.Unix(expClaim, 0)) {
		return claims, errors.Wrap(err, "token expired")
	}

	return claims, nil
}

func GetTokenClaims(token string, jwtSecret string) (*AuthJwtClaim, error) {
	var uid string

	if token != "" {
		claims, err := authenticateToken(token, jwtSecret)
		if err != nil {
			return nil, err
		}

		uid = claims["uid"].(string)

		return &AuthJwtClaim{UID: uid}, nil
	}

	return nil, nil
}
