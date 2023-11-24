package utils

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCommon_ParseUserAgent(t *testing.T) {
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
		{
			"wakatime/v1.18.11 (linux-5.13.8-200.fc34.x86_64-x86_64) go1.16.7 emacs-wakatime/1.0.2",
			"Linux",
			"emacs",
			nil,
		},
		{
			"wakatime/unset (linux-5.11.0-44-generic-x86_64) go1.16.13 emacs-wakatime/1.0.2",
			"Linux",
			"emacs",
			nil,
		},
		{
			"wakatime/ (Linux-6.0.42-1-lts-foobar-x86_64) KTextEditor/5.111.0 kate-wakatime/1.3.10",
			"Linux",
			"kate",
			nil,
		},
		{
			"Chrome/111.0.0.0 chrome-wakatime/3.0.6",
			"",
			"chrome",
			nil,
		},
		{
			"Chrome/114.0.0.0 linux_x86-64 chrome-wakatime/3.0.17",
			"Linux",
			"chrome",
			nil,
		},
		{
			"Chrome/115.0.0.0 mac_arm64 chrome-wakatime/3.0.19",
			"Mac",
			"chrome",
			nil,
		},
		{
			"Chrome/117.0.0.0 win_x86-64 chrome-wakatime/3.0.19",
			"Windows",
			"chrome",
			nil,
		},
		{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36 Edg/116.0.1938.62 win_x86-64 edge-wakatime/3.0.18",
			"Windows",
			"Edge",
			nil,
		},
	}

	for _, test := range tests {
		println(test.in)
		os, editor, err := ParseUserAgent(test.in)
		assert.True(t, checkErr(err, test.outError))
		assert.Equal(t, test.outOs, os)
		assert.Equal(t, test.outEditor, editor)
	}
}

func checkErr(expected, actual error) bool {
	return (expected == nil && actual == nil) || (expected != nil && actual != nil)
}
