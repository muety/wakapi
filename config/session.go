package config

import (
	"github.com/gorilla/sessions"
)

// sessions are only used for displaying flash messages

var sessionStore *sessions.CookieStore

func NewSessionStore() *sessions.CookieStore {
	return sessions.NewCookieStore(
		Get().Security.SessionKey,
		Get().Security.SessionKey,
	)
}

func GetSessionStore() *sessions.CookieStore {
	if sessionStore == nil {
		sessionStore = NewSessionStore()
	}
	return sessionStore
}
