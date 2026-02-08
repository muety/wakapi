package config

func WithOidcProvider(c *Config, name, clientId, clientSecret, Endpoint string, usernameClaim string) *Config {
	return WithOidcProviderAndScopes(c, name, clientId, clientSecret, Endpoint, usernameClaim, nil)
}

func WithOidcProviderAndScopes(c *Config, name, clientId, clientSecret, Endpoint string, usernameClaim string, scopes []string) *Config {
	providerConf := oidcProviderConfig{
		Name:          name,
		ClientID:      clientId,
		ClientSecret:  clientSecret,
		Endpoint:      Endpoint,
		UsernameClaim: usernameClaim,
		Scopes:        scopes,
	}

	c.Security.OidcProviders = append(c.Security.OidcProviders, providerConf)
	RegisterOidcProvider(&providerConf) // config must be Set() for this to work
	return c
}
