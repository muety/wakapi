package view

type ImprintViewModel struct {
	SharedViewModel
	HtmlText string
}

func (s *ImprintViewModel) WithSuccess(m string) *ImprintViewModel {
	s.SetSuccess(m)
	return s
}

func (s *ImprintViewModel) WithError(m string) *ImprintViewModel {
	s.SetError(m)
	return s
}

func (s *ImprintViewModel) WithHtmlText(t string) *ImprintViewModel {
	s.HtmlText = t
	return s
}
