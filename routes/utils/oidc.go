package utils

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/duke-git/lancet/v2/random"
	conf "github.com/muety/wakapi/config"
)

func SetOidcState(state string, r *http.Request, w http.ResponseWriter) {
	session, _ := conf.GetSessionStore().Get(r, conf.CookieKeySession)
	session.Values[conf.SessionValueOidcState] = state
	session.Save(r, w)
}

func SetNewOidcState(r *http.Request, w http.ResponseWriter) string {
	state := random.RandString(16)
	SetOidcState(state, r, w)
	return state
}

func GetOidcState(r *http.Request) string {
	session, _ := conf.GetSessionStore().Get(r, conf.CookieKeySession)
	state, ok := session.Values[conf.SessionValueOidcState]
	if !ok {
		return ""
	}
	return state.(string)
}

func ClearOidcState(r *http.Request, w http.ResponseWriter) {
	session, _ := conf.GetSessionStore().Get(r, conf.CookieKeySession)
	delete(session.Values, conf.SessionValueOidcState)
	session.Save(r, w)
}

func SetOidcIdTokenPayload(payload *conf.IdTokenPayload, r *http.Request, w http.ResponseWriter) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		conf.Log().Request(r).Error("failed marshal oidc id token", "error", err.Error())
		return
	}

	session, _ := conf.GetSessionStore().Get(r, conf.CookieKeySession)
	session.Values[conf.SessionValueOidcIdTokenPayload] = string(encoded)
	session.Save(r, w)
}

func GetOidcIdTokenPayload(r *http.Request) *conf.IdTokenPayload {
	session, _ := conf.GetSessionStore().Get(r, conf.CookieKeySession)

	encoded, ok := session.Values[conf.SessionValueOidcIdTokenPayload]
	if !ok {
		return nil
	}

	var payload conf.IdTokenPayload
	if err := json.Unmarshal([]byte(encoded.(string)), &payload); err != nil {
		conf.Log().Request(r).Error("failed unmarshal oidc id token", "error", err.Error())
		return nil
	}

	return &payload
}

func DecodeOidcIdToken(token string, provider *conf.OidcProvider, ctx context.Context) (*conf.IdTokenPayload, error) {
	idToken, err := provider.Verifier.Verify(ctx, token)
	if err != nil {
		return nil, err
	}

	var payload conf.IdTokenPayload
	if err := idToken.Claims(&payload); err != nil || !payload.IsValid() {
		return nil, err
	}
	payload.ProviderName = provider.Name

	return &payload, nil
}
