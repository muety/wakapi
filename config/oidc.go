package config

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

type OidcProvider struct {
	Name          string
	DisplayName   string
	UsernameClaim string
	OAuth2        *oauth2.Config
	Verifier      *oidc.IDTokenVerifier
}

type IdTokenPayload struct {
	Issuer            string                 `json:"iss"`
	Audience          jwt.ClaimStrings       `json:"aud"`
	Subject           string                 `json:"sub"`
	Expiry            int64                  `json:"exp"`
	Name              string                 `json:"name"`
	Nickname          string                 `json:"nickname"`
	PreferredUsername string                 `json:"preferred_username"`
	Email             string                 `json:"email"`
	EmailVerified     bool                   `json:"email_verified"`
	ProviderName      string                 `json:"provider_name"` // custom field, not part of actual id token response
	AllClaims         map[string]interface{} `json:"-"`
	UsernameClaim     string                 `json:"-"`
}

// Implement jwt.Claims methods

func (token *IdTokenPayload) GetExpirationTime() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{
		Time: time.Unix(token.Expiry, 0),
	}, nil
}

func (token *IdTokenPayload) GetIssuedAt() (*jwt.NumericDate, error) {
	if token.AllClaims == nil {
		return nil, fmt.Errorf("no token claims found")
	}
	iatVal, ok := token.AllClaims["iat"]
	if !ok {
		return nil, fmt.Errorf("no issued_at claim found")
	}
	iatInt, ok := iatVal.(int64)
	if !ok {
		return nil, fmt.Errorf("issued_at claim is not a number")
	}
	return &jwt.NumericDate{
		Time: time.Unix(iatInt, 0),
	}, nil
}

func (token *IdTokenPayload) GetNotBefore() (*jwt.NumericDate, error) {
	if token.AllClaims == nil {
		return nil, fmt.Errorf("no token claims found")
	}
	iatVal, ok := token.AllClaims["nbf"]
	if !ok {
		return nil, fmt.Errorf("no not_before claim found")
	}
	iatInt, ok := iatVal.(int64)
	if !ok {
		return nil, fmt.Errorf("not_before claim is not a number")
	}
	return &jwt.NumericDate{
		Time: time.Unix(iatInt, 0),
	}, nil
}

func (token *IdTokenPayload) GetIssuer() (string, error) {
	if token.Issuer == "" {
		return "", fmt.Errorf("no issuer specified")
	}
	return token.Issuer, nil
}

func (token *IdTokenPayload) GetSubject() (string, error) {
	if token.Subject == "" {
		return "", fmt.Errorf("no subject specified")
	}
	return token.Subject, nil
}

func (token *IdTokenPayload) GetAudience() (jwt.ClaimStrings, error) {
	if len(token.Audience) == 0 {
		return nil, fmt.Errorf("no audience specified")
	}
	return token.Audience, nil
}

func (token *IdTokenPayload) Username() string {
	// Check custom claim first if configured
	if token.UsernameClaim != "" {
		if val := token.getClaimValue(token.UsernameClaim); val != "" {
			return val
		}
	}
	// Fall back to default behavior
	// https://openid.net/specs/openid-connect-core-1_0.html#StandardClaims
	if s := strings.TrimSpace(token.PreferredUsername); s != "" {
		return s
	}
	if s := strings.TrimSpace(token.Nickname); s != "" {
		return s
	}
	if s := strings.TrimSpace(token.Subject); s != "" {
		return s
	}
	return ""
}

func (token *IdTokenPayload) getClaimValue(claimName string) string {
	if token.AllClaims != nil {
		if val, ok := token.AllClaims[claimName]; ok {
			if strVal, ok := val.(string); ok {
				return strings.TrimSpace(strVal)
			}
		}
	}
	return ""
}

var oidcProviders = make(map[string]*OidcProvider)

func GetOidcContext(ctx context.Context) context.Context {
	tp := http.DefaultTransport.(*http.Transport).Clone()
	tp.DisableCompression = true
	tp.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: cfg.Security.OidcInsecure,
	}
	return oidc.ClientContext(ctx, &http.Client{
		Transport: tp,
	})
}

func RegisterOidcProvider(providerCfg *oidcProviderConfig) {
	cfg := Get()

	provider, err := oidc.NewProvider(GetOidcContext(context.Background()), providerCfg.Endpoint)
	if err != nil {
		Log().Fatal(fmt.Sprintf("failed to initialize oidc provider at %s", providerCfg.Endpoint), "error", err)
		return
	}

	scopes := []string{oidc.ScopeOpenID, "profile", "email", "offline_access"}
	for _, s := range providerCfg.Scopes {
		if s != oidc.ScopeOpenID && s != "profile" && s != "email" {
			scopes = append(scopes, s)
		}
	}

	oauth2Conf := oauth2.Config{
		ClientID:     providerCfg.ClientID,
		ClientSecret: providerCfg.ClientSecret,
		RedirectURL:  fmt.Sprintf("%s/oidc/%s/callback", cfg.Server.GetPublicUrl(), providerCfg.Name),
		Endpoint:     provider.Endpoint(),
		Scopes:       scopes,
	}

	oidcProviders[providerCfg.Name] = &OidcProvider{
		Name:          providerCfg.Name,
		DisplayName:   providerCfg.String(),
		UsernameClaim: providerCfg.UsernameClaim,
		OAuth2:        &oauth2Conf,
		Verifier:      provider.Verifier(&oidc.Config{ClientID: providerCfg.ClientID}),
	}
}

func GetOidcProvider(name string) (*OidcProvider, error) {
	provider, ok := oidcProviders[name]
	if !ok {
		return nil, fmt.Errorf("oidc provider not found: %s", name)
	}
	return provider, nil
}

func MustGetOidcProvider(name string) *OidcProvider {
	provider, _ := GetOidcProvider(name)
	return provider
}
