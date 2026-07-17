package config

import (
	"crypto/hkdf"
	"crypto/sha256"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

const SessionKeyDeriveInfo = "COOKIE-SESSION;SIGNED:HMAC-SHA256;ENCRYPTED:AES-256-CTR"
const AuthKeyDeriveInfo = "AUTH-KEY;SIGNED:HMAC-SHA256;ENCRYPTED:AES-256-CTR"

func deriveKey(info string, keyLen int) ([]byte, error) {
	return hkdf.Key(sha256.New, Get().Security.CookieKeyBytes, nil, info, keyLen)
}

type cookieConfig struct {
	sessionStore *sessions.CookieStore
	authCookie   *securecookie.SecureCookie
}

var cookieConfigInstance = cookieConfig{
	sessionStore: nil,
	authCookie:   nil,
}

func (config *cookieConfig) newSessionStore() {
	sessionKeys, err := deriveKey(SessionKeyDeriveInfo, 64+32) // 64 bytes for authentication key, 32 bytes for encryption key
	if err != nil {
		Log().Fatal("error while deriving session keys", "error", err)
	}

	store := sessions.NewCookieStore(sessionKeys[:64], sessionKeys[64:])
	store.Options.SameSite = http.SameSiteLaxMode
	store.Options.HttpOnly = true

	if Get().Security.InsecureCookies {
		store.Options.Secure = false
	}

	config.sessionStore = store
}

func (config *cookieConfig) newAuthCookie() {
	authKeys, err := deriveKey(AuthKeyDeriveInfo, 64+32) // 64 bytes for authentication key, 32 bytes for encryption key
	if err != nil {
		Log().Fatal("error while deriving auth keys", "error", err)
	}

	config.authCookie = securecookie.New(authKeys[:64], authKeys[64:])
	config.authCookie.SetSerializer(securecookie.JSONEncoder{})
}

func (config *cookieConfig) init() {
	config.newSessionStore()
	config.newAuthCookie()
}

func InitializeCookies() {
	cookieConfigInstance.init()
}

func GetSessionStore() *sessions.CookieStore {
	if cookieConfigInstance.sessionStore == nil {
		cookieConfigInstance.newSessionStore()
	}
	return cookieConfigInstance.sessionStore
}

func GetAuthCookie() *securecookie.SecureCookie {
	if cookieConfigInstance.authCookie == nil {
		cookieConfigInstance.newAuthCookie()
	}
	return cookieConfigInstance.authCookie
}
