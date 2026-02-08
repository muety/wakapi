package routes

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/oauth2-proxy/mockoidc"
)

// https://github.com/go-chi/chi/issues/76#issuecomment-370145140
func WithUrlParam(r *http.Request, key, value string) *http.Request {
	r.URL.RawPath = strings.Replace(r.URL.RawPath, "{"+key+"}", value, 1)
	r.URL.Path = strings.Replace(r.URL.Path, "{"+key+"}", value, 1)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	return r
}

// WakapiMockOIDCUser implements mockoidc.User interface with custom claims
type WakapiMockOIDCUser struct {
	mockoidc.MockUser
	CustomClaimName  string
	CustomClaimValue string
}

func (u *WakapiMockOIDCUser) Claims(scope []string, claims *mockoidc.IDTokenClaims) (jwt.Claims, error) {
	// see https://pkg.go.dev/github.com/oauth2-proxy/mockoidc#MockUser
	mapClaims := jwt.MapClaims{
		"sub":                u.Subject,
		"email":              u.Email,
		"email_verified":     u.EmailVerified,
		"preferred_username": u.PreferredUsername,
		"name":               u.Subject,
	}

	// copy registered claims to dynamic claims map
	if claims != nil && claims.RegisteredClaims != nil {
		if claims.Issuer != "" {
			mapClaims["iss"] = claims.Issuer
		}
		if claims.Subject != "" {
			mapClaims["sub"] = claims.Subject
		}
		if len(claims.Audience) > 0 {
			mapClaims["aud"] = claims.Audience
		}
		if !claims.ExpiresAt.IsZero() {
			mapClaims["exp"] = claims.ExpiresAt.Unix()
		}
		if !claims.IssuedAt.IsZero() {
			mapClaims["iat"] = claims.IssuedAt.Unix()
		}
		if claims.ID != "" {
			mapClaims["jti"] = claims.ID
		}
		if claims.Nonce != "" {
			mapClaims["nonce"] = claims.Nonce
		}
	}

	mapClaims[u.CustomClaimName] = u.CustomClaimValue
	return mapClaims, nil
}
