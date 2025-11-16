package view

type SetupViewModel struct {
	SharedLoggedInViewModel
}

func (s *SetupViewModel) WithSuccess(m string) *SetupViewModel {
	s.SetSuccess(m)
	return s
}

func (s *SetupViewModel) WithError(m string) *SetupViewModel {
	s.SetError(m)
	return s
}
