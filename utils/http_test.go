package utils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var userAgents = []struct {
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
	{
		"wakatime/v1.86.5 (linux-6.6.4-200.fc39.x86_64-unknown) go1.21.3 neovim/900 vim-wakatime/11.1.1",
		"Linux",
		"neovim",
		nil,
	},
	{
		"wakatime/v1.102.1 (windows-10.0.27723.1000-x86_64) go1.22.5 Skype/unknown windows-wakatime/0.5.0", // desktop-wakatime
		"Windows",
		"Skype",
		nil,
	},
	{
		"wakatime/v1.102.1 (windows-10.0.27718.1000-x86_64) go1.22.5 Notepad++/unknown windows-wakatime/0.5.0", // desktop-wakatime
		"Windows",
		"Notepad++",
		nil,
	},
	{
		"wakatime/v1.105.0 (linux-6.11.8-zen1-2-zen-unknown) go1.23.3 cursor/1.93.1 vscode-wakatime/24.8.0", // https://github.com/muety/wakapi/issues/712
		"Linux",
		"cursor",
		nil,
	},
	{
		// https://github.com/muety/wakapi/issues/817
		// https://github.com/muety/wakapi/issues/718 (previously)
		"wakatime/v1.106.1 (linux-5.15.167.4-microsoft-standard-WSL2-unknown) go1.23.3 cursor/1.93.1 vscode-wakatime/24.9.2",
		"WSL",
		"cursor",
		nil,
	},
	{
		"HBuilder X/4.56 (Windows_NT 10.0.26100)", // https://github.com/muety/wakapi/issues/765
		"Windows",
		"HBuilder X",
		nil,
	},
	{
		"wakatime/1.139.1 (linux-6.18.8-unknown) go1.25.5 helix/25.07.1 (74075bb5) wakatime-ls/0.2.2 helix-wakatime/0.2.2", // https://github.com/muety/wakapi/issues/914
		"Linux",
		"helix",
		nil,
	},
	{
		"wakatime/v1.105.0 (linux-6.11.9-zen1-1-zen-unknown) go1.23.3 vscode/1.95.3 vscode-wakatime/24.8.0",
		"Linux",
		"vscode",
		nil,
	},
	{
		"wakatime/v2.7.0 (linux-6.19.12-200.fc43.x86_64-unknown) go1.25.9 Claude/2.1.118",
		"Linux",
		"Claude",
		nil,
	},
	{
		"wakatime/v1.107.0 (linux-6.11.8) go1.23.3 Claude/2.1.118 jetbrains/PyCharm/2023.1",
		"Linux",
		"Claude",
		nil,
	},
	{
		"wakatime/v1.115.2 (windows-10.0.22631.5335-x86_64) go1.24.2 Claude/unknown windows-wakatime/2.1.6",
		"Windows",
		"Claude",
		nil,
	},
	{
		"wakatime/v1.130.1 (linux-6.6.87.2-microsoft-standard-WSL2-x86_64) go1.24.4 claude-code-wakatime/2.1.0",
		"WSL",
		"Claude",
		nil,
	},
	{
		"wakatime/v1.131.0 (darwin-25.0.0-arm64) go1.24.4 Claude/0.11.3-0.11.3 macos-wakatime/5.27.2",
		"Macos",
		"Claude",
		nil,
	},
	{
		"wakatime/v1.139.1 (darwin-25.2.0-arm64) go1.25.5 claude-code",
		"Macos",
		"", // not a properly formatted user-agent string in our understanding
		nil,
	},
	{
		"wakatime/v1.123.0 (darwin-23.4.0-arm64) go1.24.4 windsurf/1.99.3 vscode-wakatime/25.1.1",
		"Macos",
		"windsurf",
		nil,
	},
	{
		"wakatime/v1.124.1 (windows-10.0.26100.4652-x86_64) go1.24.4 kiro/1.94.0 vscode-wakatime/25.2.0",
		"Windows",
		"kiro",
		nil,
	},
}

func BenchmarkCommon_ParseUserAgent(b *testing.B) {
	for range b.N {
		for _, test := range userAgents {
			_, _, _ = ParseUserAgent(test.in)
		}
	}
}

func TestCommon_ParseUserAgent(t *testing.T) {
	for _, test := range userAgents {
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
