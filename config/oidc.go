package config

import (
	"context"
	"fmt"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type OidcProvider struct {
	Name     string
	OAuth2   *oauth2.Config
	Verifier *oidc.IDTokenVerifier
}

type IdTokenPayload struct {
	Issuer            string `json:"iss"`
	Subject           string `json:"sub"`
	Expiry            int64  `json:"exp"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
	ProviderName      string `json:"-"`
}

func (token *IdTokenPayload) Exp() time.Time {
	return time.Unix(token.Expiry, 0)
}

func (token *IdTokenPayload) IsValid() bool {
	return token.Exp().After(time.Now())
}

var oidcProviders = make(map[string]*OidcProvider)

func RegisterOidcProvider(providerCfg *oidcProviderConfig) {
	cfg := Get()

	provider, err := oidc.NewProvider(context.Background(), providerCfg.Endpoint)
	if err != nil {
		Log().Fatal(fmt.Sprintf("failed to initialize oidc provider at %s", providerCfg.Endpoint), err)
		return
	}

	oauth2Conf := oauth2.Config{
		ClientID:     providerCfg.ClientID,
		ClientSecret: providerCfg.ClientSecret,
		RedirectURL:  fmt.Sprintf("%s/oidc/%s/callback", cfg.Server.GetPublicUrl(), providerCfg.Name),
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	oidcProviders[providerCfg.Name] = &OidcProvider{
		Name:     providerCfg.Name,
		OAuth2:   &oauth2Conf,
		Verifier: provider.Verifier(&oidc.Config{ClientID: providerCfg.ClientID}),
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
