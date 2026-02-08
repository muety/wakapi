package config

func WithOidcProvider(c *Config, name, clientId, clientSecret, Endpoint string, usernameClaim ...string) *Config {
	providerConf := oidcProviderConfig{
		Name:         name,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Endpoint:     Endpoint,
	}
	if len(usernameClaim) > 0 {
		providerConf.UsernameClaim = usernameClaim[0]
	}
	c.Security.OidcProviders = append(c.Security.OidcProviders, providerConf)
	RegisterOidcProvider(&providerConf) // config must be Set() for this to work
	return c
}
