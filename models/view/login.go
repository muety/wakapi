package view

type LoginViewModel struct {
	SharedViewModel
	AllowSignup      bool
	DisableLocalAuth bool
	DisableWebAuthn  bool
	CaptchaId        string
	InviteCode       string
	OidcProviders    []LoginViewModelOidcProvider
}

type LoginViewModelOidcProvider struct {
	Name        string
	DisplayName string
}

type SetPasswordViewModel struct {
	LoginViewModel
	Token string
}

func (s *LoginViewModel) OidcProviderIcon(provider string) string {
	return GetOidcProviderIcon(provider)
}

func (s *LoginViewModel) WithSuccess(m string) *LoginViewModel {
	s.SetSuccess(m)
	return s
}

func (s *LoginViewModel) WithError(m string) *LoginViewModel {
	s.SetError(m)
	return s
}
