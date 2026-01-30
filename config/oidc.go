package config

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
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
	Subject           string                 `json:"sub"`
	Expiry            int64                  `json:"exp"`
	Name              string                 `json:"name"`
	Nickname          string                 `json:"nickname"`
	PreferredUsername string                 `json:"preferred_username"`
	Email             string                 `json:"email"`
	EmailVerified     bool                   `json:"email_verified"`
	ProviderName      string                 `json:"provider_name"`  // custom field, not part of actual id token response
	CustomClaims      map[string]interface{} `json:"-"`
	UsernameClaim     string                 `json:"-"`
}

func (token *IdTokenPayload) Exp() time.Time {
	return time.Unix(token.Expiry, 0)
}

func (token *IdTokenPayload) IsValid() bool {
	return token.Exp().After(time.Now())
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
	switch claimName {
	case "preferred_username":
		return strings.TrimSpace(token.PreferredUsername)
	case "nickname":
		return strings.TrimSpace(token.Nickname)
	case "sub":
		return strings.TrimSpace(token.Subject)
	case "email":
		return strings.TrimSpace(token.Email)
	case "name":
		return strings.TrimSpace(token.Name)
	}
	if token.CustomClaims != nil {
		if val, ok := token.CustomClaims[claimName]; ok {
			if strVal, ok := val.(string); ok {
				return strings.TrimSpace(strVal)
			}
		}
	}
	return ""
}

var oidcProviders = make(map[string]*OidcProvider)

func RegisterOidcProvider(providerCfg *oidcProviderConfig) {
	cfg := Get()

	provider, err := oidc.NewProvider(context.Background(), providerCfg.Endpoint)
	if err != nil {
		Log().Fatal(fmt.Sprintf("failed to initialize oidc provider at %s", providerCfg.Endpoint), "error", err)
		return
	}

	scopes := []string{oidc.ScopeOpenID, "profile", "email"}
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
