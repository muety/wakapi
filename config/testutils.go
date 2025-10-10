package config

func WithOidcProvider(c *Config, name, clientId, clientSecret, Endpoint string) *Config {
	providerConf := oidcProviderConfig{
		Name:         name,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Endpoint:     Endpoint,
	}
	c.Security.OidcProviders = append(c.Security.OidcProviders, providerConf)
	RegisterOidcProvider(&providerConf) // config must be Set() for this to work
	return c
}
