package routes

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/oauth2-proxy/mockoidc"
)

// WakapiMockOIDCUser implements mockoidc.User interface with custom claims.
// Kept in a _test.go file so the mock OIDC server (and its transitive
// dependencies) is only linked into test binaries, never into the
// production build.
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
