package models

import (
	"fmt"
	"strings"
)

type Mail struct {
	From    MailAddress
	To      MailAddresses
	Subject string
	Body    string
	Type    string
}

func (m *Mail) WithText(text string) *Mail {
	m.Body = text
	m.Type = "text/plain; charset=UTF-8"
	return m
}

func (m *Mail) WithHTML(html string) *Mail {
	m.Body = html
	m.Type = "text/html; charset=UTF-8"
	return m
}

func (m *Mail) String() string {
	return fmt.Sprintf("To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: %s\r\n"+
		"\r\n"+
		"%s\r\n",
		strings.Join(m.To.RawStrings(), ", "),
		m.From.String(),
		m.Subject,
		m.Type,
		m.Body,
	)
}

func (m *Mail) Reader() *strings.Reader {
	return strings.NewReader(m.String())
}
