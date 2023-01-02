package config

import "github.com/gorilla/sessions"

// sessions are only used for displaying flash messages

var sessionStore *sessions.CookieStore

func GetSessionStore() *sessions.CookieStore {
	if sessionStore == nil {
		sessionStore = sessions.NewCookieStore(Get().Security.SessionKey)
	}
	return sessionStore
}
