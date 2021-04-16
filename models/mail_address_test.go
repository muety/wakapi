package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMailAddress_SingleRaw(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{
			"john.doe@example.org",
			"john.doe@example.org",
		},
		{
			"John Doe <john.doe@example.org>",
			"john.doe@example.org",
		},
		{
			"invalid",
			"",
		},
	}

	for _, test := range tests {
		out := MailAddress(test.in).Raw()
		assert.Equal(t, test.out, out)
	}
}

func TestMailAddress_AllRaw(t *testing.T) {
	tests := []struct {
		in  []string
		out []string
	}{
		{
			[]string{"john.doe@example.org", "foo@bar.com"},
			[]string{"john.doe@example.org", "foo@bar.com"},
		},
		{
			[]string{"John Doe <john.doe@example.org>", "foo@bar.com"},
			[]string{"john.doe@example.org", "foo@bar.com"},
		},
		{
			[]string{"john.doe@example.org", "invalid"},
			[]string{"john.doe@example.org", ""},
		},
	}

	for _, test := range tests {
		out := castAddresses(test.in).RawStrings()
		assert.EqualValues(t, test.out, out)
	}
}

func TestMailAddress_AllValid(t *testing.T) {
	tests := []struct {
		in  []string
		out bool
	}{
		{
			[]string{"john.doe@example.org", "foo@bar.com"},
			true,
		},
		{
			[]string{"John Doe <john.doe@example.org>", "ínvalid"},
			false,
		},
		{
			[]string{"", "invalid"},
			false,
		},
	}

	for _, test := range tests {
		out := castAddresses(test.in).AllValid()
		assert.EqualValues(t, test.out, out)
	}
}

func castAddresses(addresses []string) (m MailAddresses) {
	for _, a := range addresses {
		m = append(m, MailAddress(a))
	}
	return m
}
