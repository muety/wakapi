package config

import (
	"testing"

	"github.com/oauth2-proxy/mockoidc"
	"github.com/stretchr/testify/assert"
)

func TestIdTokenPayload_Username(t *testing.T) {
	testCases := []struct {
		name     string
		token    IdTokenPayload
		expected string
	}{
		{
			name: "custom claim takes priority",
			token: IdTokenPayload{
				UsernameClaim:     "custom_claim",
				AllClaims:         map[string]interface{}{"custom_claim": "custom_user"},
				PreferredUsername: "preferred_user",
				Nickname:          "nickname_user",
				Subject:           "subject_user",
			},
			expected: "custom_user",
		},
		{
			name: "preferred_username fallback",
			token: IdTokenPayload{
				PreferredUsername: "preferred_user",
				Nickname:          "nickname_user",
				Subject:           "subject_user",
			},
			expected: "preferred_user",
		},
		{
			name: "nickname fallback",
			token: IdTokenPayload{
				PreferredUsername: "",
				Nickname:          "nickname_user",
				Subject:           "subject_user",
			},
			expected: "nickname_user",
		},
		{
			name: "subject fallback",
			token: IdTokenPayload{
				PreferredUsername: "",
				Nickname:          "",
				Subject:           "subject_user",
			},
			expected: "subject_user",
		},
		{
			name: "empty when nothing available",
			token: IdTokenPayload{
				PreferredUsername: "",
				Nickname:          "",
				Subject:           "",
			},
			expected: "",
		},
		{
			name: "custom claim not found falls back to preferred_username",
			token: IdTokenPayload{
				UsernameClaim:     "missing_claim",
				AllClaims:         map[string]interface{}{"other_claim": "other_value"},
				PreferredUsername: "preferred_user",
			},
			expected: "preferred_user",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.token.Username()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func resetOidcProviders() {
	oidcProviders = make(map[string]*OidcProvider)
}

func TestRegisterOidcProviderNormalizesName(t *testing.T) {
	resetOidcProviders()

	mock, err := mockoidc.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Shutdown()

	Set(Empty())
	WithOidcProvider(Get(), "Authentik", mock.ClientID, mock.ClientSecret, mock.Addr()+"/oidc", "")

	provider, err := GetOidcProvider("Authentik")
	assert.NoError(t, err)
	assert.NotNil(t, provider)

	assert.Equal(t, "authentik", provider.Name)
	assert.Equal(t, "Authentik", provider.DisplayName)
	assert.Contains(t, provider.OAuth2.RedirectURL, "/oidc/authentik/callback")

	for _, name := range []string{"authentik", "AUTHENTIK", "aUtHeNtIk", "Authentik"} {
		p, err := GetOidcProvider(name)
		assert.NoError(t, err, name)
		assert.Same(t, provider, p, name)
	}
}

func TestGetOidcProviderNotFound(t *testing.T) {
	resetOidcProviders()

	_, err := GetOidcProvider("nonexistent")
	assert.Error(t, err)

	_, err = GetOidcProvider("Nonexistent")
	assert.Error(t, err)
}

func TestRegisterOidcProviderCaseCollisionOverwrites(t *testing.T) {
	resetOidcProviders()

	mock1, err := mockoidc.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mock1.Shutdown()

	mock2, err := mockoidc.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mock2.Shutdown()

	Set(Empty())
	cfg := Get()
	WithOidcProvider(cfg, "authentik", mock1.ClientID, mock1.ClientSecret, mock1.Addr()+"/oidc", "")
	WithOidcProvider(cfg, "Authentik", mock2.ClientID, mock2.ClientSecret, mock2.Addr()+"/oidc", "")

	assert.Len(t, oidcProviders, 1)

	provider, err := GetOidcProvider("authentik")
	assert.NoError(t, err)
	assert.Equal(t, mock2.ClientID, provider.OAuth2.ClientID)
}

func TestInitOpenIDConnectNormalizesProviderNames(t *testing.T) {
	resetOidcProviders()

	mock, err := mockoidc.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Shutdown()

	Set(Empty())
	cfg := Get()
	cfg.Security.OidcProviders = []oidcProviderConfig{
		{Name: "Authentik", ClientID: mock.ClientID, ClientSecret: mock.ClientSecret, Endpoint: mock.Addr() + "/oidc"},
	}

	initOpenIDConnect(cfg)

	assert.Equal(t, "authentik", cfg.Security.OidcProviders[0].Name)
	assert.Equal(t, []string{"authentik"}, cfg.Security.ListOidcProviders())
}
