package models

import (
	"fmt"
	"github.com/gofrs/uuid/v5"
	"strings"
	"time"
)

const HtmlType = "text/html; charset=UTF-8"
const PlainType = "text/html; charset=UTF-8"

type Mail struct {
	From      MailAddress
	To        MailAddresses
	Subject   string
	Body      string
	Type      string
	Date      time.Time
	MessageID string
}

func (m *Mail) WithText(text string) *Mail {
	m.Body = text
	m.Type = PlainType
	return m
}

func (m *Mail) WithHTML(html string) *Mail {
	m.Body = html
	m.Type = HtmlType
	return m
}

func (m *Mail) Sanitized() *Mail {
	if m.Type == "" {
		m.Type = PlainType
	}
	if m.Date.IsZero() {
		m.Date = time.Now()
	}
	if m.MessageID == "" {
		m.MessageID = fmt.Sprintf("<%s@%s>", uuid.Must(uuid.NewV4()).String(), m.From.Domain())
	}
	return m
}

func (m *Mail) String() string {
	return fmt.Sprintf("To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"Message-ID: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: %s\r\n"+
		"Content-Transfer-Encoding: 8bit\r\n"+
		"Date: %s\r\n"+
		"\r\n"+
		"%s\r\n",
		strings.Join(m.To.RawStrings(), ", "),
		m.From.String(),
		m.Subject,
		m.MessageID,
		m.Type,
		m.Date.Format(time.RFC1123Z),
		m.Body,
	)
}

func (m *Mail) Reader() *strings.Reader {
	return strings.NewReader(m.String())
}
