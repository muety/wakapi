package utils

import (
	"errors"
	"testing"
)

func TestParseUserAgent(t *testing.T) {
	tests := []struct {
		in        string
		outOs     string
		outEditor string
		outError  error
	}{
		{
			"wakatime/13.0.7 (Linux-4.15.0-96-generic-x86_64-with-glibc2.4) Python3.8.0.final.0 GoLand/2019.3.4 GoLand-wakatime/11.0.1",
			"Linux",
			"GoLand",
			nil,
		},
		{
			"wakatime/13.0.4 (Linux-5.4.64-x86_64-with-glibc2.2.5) Python3.7.6.final.0 emacs-wakatime/1.0.2",
			"Linux",
			"emacs",
			nil,
		},
		{
			"",
			"",
			"",
			errors.New(""),
		},
		{
			"wakatime/13.0.7 Python3.8.0.final.0 GoLand/2019.3.4 GoLand-wakatime/11.0.1",
			"",
			"",
			errors.New(""),
		},
	}

	for i, test := range tests {
		if os, editor, err := ParseUserAgent(test.in); os != test.outOs || editor != test.outEditor || !checkErr(test.outError, err) {
			t.Errorf("[%d] Unexpected result of parsing '%s'; got '%v', '%v', '%v'", i, test.in, os, editor, err)
		}
	}
}

func checkErr(expected, actual error) bool {
	return (expected == nil && actual == nil) || (expected != nil && actual != nil)
}
