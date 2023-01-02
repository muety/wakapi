package view

type LoginViewModel struct {
	Messages
	TotalUsers  int
	AllowSignup bool
}

type SetPasswordViewModel struct {
	LoginViewModel
	Token string
}

func (s *LoginViewModel) WithSuccess(m string) *LoginViewModel {
	s.SetSuccess(m)
	return s
}

func (s *LoginViewModel) WithError(m string) *LoginViewModel {
	s.SetError(m)
	return s
}
