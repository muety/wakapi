package view

type SummaryViewModel struct {
	Success string
	Error   string
}

func (s *SummaryViewModel) WithSuccess(m string) *SummaryViewModel {
	s.Success = m
	return s
}

func (s *SummaryViewModel) WithError(m string) *SummaryViewModel {
	s.Error = m
	return s
}
