package view

type LoginViewModel struct {
	Success    string
	Error      string
	TotalUsers int
}

func (s *LoginViewModel) WithSuccess(m string) *LoginViewModel {
	s.Success = m
	return s
}

func (s *LoginViewModel) WithError(m string) *LoginViewModel {
	s.Error = m
	return s
}
