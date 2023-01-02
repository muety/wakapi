package view

type Newsbox struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type HomeViewModel struct {
	Messages
	TotalHours int
	TotalUsers int
	Newsbox    *Newsbox
}

func (s *HomeViewModel) WithSuccess(m string) *HomeViewModel {
	s.SetSuccess(m)
	return s
}

func (s *HomeViewModel) WithError(m string) *HomeViewModel {
	s.SetError(m)
	return s
}
