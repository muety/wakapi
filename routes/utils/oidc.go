package utils

import (
	"context"
	"errors"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/duke-git/lancet/v2/random"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"golang.org/x/oauth2"
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

func DecodeOidcIdToken(token string, provider *conf.OidcProvider, ctx context.Context) (*conf.IdTokenPayload, error) {
	idToken, err := provider.Verifier.Verify(ctx, token)
	if err != nil {
		return nil, err
	}

	var payload conf.IdTokenPayload
	if err := idToken.Claims(&payload); err != nil {
		return nil, err
	}
	payload.ProviderName = provider.Name

	var allClaims map[string]interface{}
	if err := idToken.Claims(&allClaims); err == nil {
		payload.AllClaims = allClaims
	}

	payload.UsernameClaim = provider.UsernameClaim
	return &payload, nil
}

func RefreshOidcIdToken(ctx context.Context, provider *conf.OidcProvider, refreshToken string) (*oauth2.Token, error) {
	ts := provider.OAuth2.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken})
	return ts.Token()
}

func ExtractOidcAuth(w http.ResponseWriter, r *http.Request) (*conf.IdTokenPayload, error) {
	providerCookie, err := r.Cookie(models.OidcProviderCookieKey)
	if err != nil {
		return nil, errors.New("missing authentication")
	}
	idTokenCookie, err := r.Cookie(models.OidcIdTokenCookieKey)
	if err != nil {
		return nil, errors.New("missing authentication")
	}
	refreshToken, _ := r.Cookie(models.OidcRefreshTokenCookieKey)

	provider, err := conf.GetOidcProvider(providerCookie.Value)
	if err != nil {
		return nil, errors.New("invalid OIDC provider")
	}

	oidcContext := conf.GetOidcContext(r.Context())
	idTokenPayload, err := DecodeOidcIdToken(idTokenCookie.Value, provider, oidcContext)
	if err == nil {
		return idTokenPayload, nil
	}

	_, ok := errors.AsType[*oidc.TokenExpiredError](err)
	if !ok || refreshToken == nil {
		return nil, err
	}

	authToken, err := RefreshOidcIdToken(oidcContext, provider, refreshToken.Value)
	if err != nil {
		return nil, err
	}
	rawIdToken, ok := authToken.Extra("id_token").(string)
	if !ok {
		return nil, err
	}
	idTokenPayload, err = DecodeOidcIdToken(rawIdToken, provider, oidcContext)
	if err != nil {
		return nil, err
	}

	http.SetCookie(w, conf.Get().CreateCookie(models.OidcIdTokenCookieKey, rawIdToken))
	http.SetCookie(w, conf.Get().CreateCookie(models.OidcRefreshTokenCookieKey, authToken.RefreshToken))

	return idTokenPayload, nil
}
