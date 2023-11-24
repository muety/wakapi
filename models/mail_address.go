package models

import (
	"regexp"
	"strings"
)

const (
	MailPattern      = "[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\\.[a-zA-Z0-9-.]+"
	EmailAddrPattern = ".*\\s<(" + MailPattern + ")>|(" + MailPattern + ")"
)

var (
	mailRegex      *regexp.Regexp
	emailAddrRegex *regexp.Regexp
)

func init() {
	mailRegex = regexp.MustCompile(MailPattern)
	emailAddrRegex = regexp.MustCompile(EmailAddrPattern)
}

type MailAddress string

type MailAddresses []MailAddress

func (m MailAddress) String() string {
	return string(m)
}

func (m MailAddress) Raw() string {
	match := emailAddrRegex.FindStringSubmatch(string(m))
	if len(match) == 3 {
		if match[2] != "" {
			return match[2]
		}
		return match[1]
	}
	return ""
}

func (m MailAddress) Domain() string {
	split := strings.Split(m.Raw(), "@")
	if len(split) != 2 {
		return ""
	}
	return split[1]
}

func (m MailAddress) Valid() bool {
	return emailAddrRegex.MatchString(string(m))
}

func (m MailAddresses) Strings() []string {
	out := make([]string, len(m))
	for i, s := range m {
		out[i] = s.String()
	}
	return out
}

func (m MailAddresses) RawStrings() []string {
	out := make([]string, len(m))
	for i, s := range m {
		out[i] = s.Raw()
	}
	return out
}

func (m MailAddresses) AllValid() bool {
	for _, a := range m {
		if !a.Valid() {
			return false
		}
	}
	return true
}
