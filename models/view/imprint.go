package view

type ImprintViewModel struct {
	HtmlText string
	Success  string
	Error    string
}

func (s *ImprintViewModel) WithSuccess(m string) *ImprintViewModel {
	s.Success = m
	return s
}

func (s *ImprintViewModel) WithError(m string) *ImprintViewModel {
	s.Error = m
	return s
}

func (s *ImprintViewModel) WithHtmlText(t string) *ImprintViewModel {
	s.HtmlText = t
	return s
}
