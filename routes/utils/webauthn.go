package utils

import (
	"net/http"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/muety/wakapi/config"
)

func SetWebAuthnSession(session *webauthn.SessionData, r *http.Request, w http.ResponseWriter) error {
	sess, _ := config.GetSessionStore().Get(r, config.CookieKeySession)
	sess.Values[config.SessionValueWebAuthn] = session
	return sess.Save(r, w)
}

func GetWebAuthnSession(r *http.Request) (*webauthn.SessionData, error) {
	sess, _ := config.GetSessionStore().Get(r, config.CookieKeySession)
	val, ok := sess.Values[config.SessionValueWebAuthn]
	if !ok {
		return &webauthn.SessionData{}, nil
	}
	session, ok := val.(*webauthn.SessionData)
	if !ok {
		config.Log().Error("webauthn session data has invalid type")
		return &webauthn.SessionData{}, nil
	}
	return session, nil
}
