package config

import (
	"net/http"

	"github.com/gorilla/sessions"
)

// sessions are only used for displaying flash messages

var sessionStore *sessions.CookieStore

func NewSessionStore() *sessions.CookieStore {
	store := sessions.NewCookieStore(
		Get().Security.SessionKey,
		Get().Security.SessionKey,
	)

	if Get().Security.InsecureCookies {
		store.Options.SameSite = http.SameSiteLaxMode
		store.Options.Secure = false
	}

	return store
}

func GetSessionStore() *sessions.CookieStore {
	if sessionStore == nil {
		sessionStore = NewSessionStore()
	}
	return sessionStore
}

func ResetSessionStore() {
	sessionStore = nil
}
