package view

type Newsbox struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type HomeViewModel struct {
	Success    string
	Error      string
	TotalHours int
	TotalUsers int
	Newsbox    *Newsbox
}

func (s *HomeViewModel) WithSuccess(m string) *HomeViewModel {
	s.Success = m
	return s
}

func (s *HomeViewModel) WithError(m string) *HomeViewModel {
	s.Error = m
	return s
}
