package view

type HomeViewModel struct {
	Success string
	Error   string
}

func (s *HomeViewModel) WithSuccess(m string) *HomeViewModel {
	s.Success = m
	return s
}

func (s *HomeViewModel) WithError(m string) *HomeViewModel {
	s.Error = m
	return s
}
