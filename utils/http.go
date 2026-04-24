package utils

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/duke-git/lancet/v2/condition"
	set "github.com/duke-git/lancet/v2/datastructure/set"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/mileusna/useragent"

	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	cacheMaxAgePattern = `max-age=(\d+)`
)

var cacheMaxAgeRe *regexp.Regexp

// https://github.com/muety/wakapi/issues/914
var editorMiddlewares = set.New[string]("wakatime-ls", "wakatime-cli")

// https://github.com/wakatime/wakatime-cli/blob/2d84fc82f57b9a8bc113edbd24f6b41762b96816/pkg/ai/ai.go#L45
var aiTools = set.New[string]("claude", "chatgpt", "copilot", "codex", "cursor", "windsurf", "cline", "roo-code", "gemini", "pi", "goose")
var knownOs = set.New[string]("linux", "windows", "macos", "darwin", "win", "mac", "wsl")

func init() {
	cacheMaxAgeRe = regexp.MustCompile(cacheMaxAgePattern)
}

type PageParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

func (p *PageParams) Limit() int {
	if p.PageSize < 0 {
		return 0
	}
	return p.PageSize
}

func (p *PageParams) Offset() int {
	if p.PageSize <= 0 {
		return 0
	}
	return (p.Page - 1) * p.PageSize
}

// IsNoCache checks whether returning a cached resource no older than cacheTtl is allowed given the incoming request
func IsNoCache(r *http.Request, cacheTtl time.Duration) bool {
	cacheControl := r.Header.Get("cache-control")
	if strings.Contains(cacheControl, "no-cache") {
		return true
	}
	if match := cacheMaxAgeRe.FindStringSubmatch(cacheControl); match != nil && len(match) > 1 {
		if maxAge, _ := strconv.Atoi(match[1]); time.Duration(maxAge)*time.Second <= cacheTtl {
			return true
		}
	}
	return false
}

func ParsePageParams(r *http.Request) *PageParams {
	pageParams := &PageParams{}
	page := r.URL.Query().Get("page")
	pageSize := r.URL.Query().Get("page_size")
	if p, err := strconv.Atoi(page); err == nil {
		pageParams.Page = p
	}
	if p, err := strconv.Atoi(pageSize); err == nil && pageParams.Page > 0 {
		pageParams.PageSize = p
	}
	return pageParams
}

func ParsePageParamsWithDefault(r *http.Request, page, size int) *PageParams {
	pageParams := ParsePageParams(r)
	if pageParams.Page == 0 {
		pageParams.Page = page
	}
	if pageParams.PageSize == 0 {
		pageParams.PageSize = size
	}
	return pageParams
}

// ParseUserAgent extracts the operating system and editor from a User-Agent string.
func ParseUserAgent(ua string) (string, string, error) { // os, editor, err
	ua = strings.TrimSpace(ua)
	if ua == "" {
		return "", "", errors.New("empty user agent")
	}

	// Try to parse WakaTime and browser extension user agents
	if parts := strings.Fields(ua); len(parts) >= 2 {
		first := strings.ToLower(parts[0])

		if strings.HasPrefix(first, "wakatime/") ||
			strings.HasPrefix(first, "chrome/") ||
			strings.HasPrefix(first, "firefox/") ||
			strings.HasPrefix(first, "edge/") {

			editor := extractEditor(ua, parts)
			if editor == "KTextEditor" { // special treatment for neovim
				editor = "kate"
			}
			if editor == "claude-code" { // special treatment for Claude Code
				editor = "Claude"
			}

			os := extractOS(parts)
			if os == "" && strings.HasPrefix(first, "wakatime/") {
				return "", "", errors.New("failed to parse os from wakatime user agent")
			}
			if os == "win" {
				os = "windows"
			} else if os == "darwin" {
				os = "macos"
			}
			// special treatment for wsl (see https://github.com/muety/wakapi/issues/817)
			osAllCaps := false
			if strings.Contains(ua, "-WSL2-") {
				os = "wsl"
				osAllCaps = true
			}
			os = condition.Ternary[bool, string](osAllCaps, strings.ToUpper(os), strutil.Capitalize(os))

			return os, editor, nil
		}
	}

	// Try parse browser user agent as a fallback
	if parsedUa := useragent.Parse(ua); len(parsedUa.Name) > 0 {
		if len(parsedUa.OS) > 0 {
			return strutil.Capitalize(parsedUa.OS), parsedUa.Name, nil
		} else if strings.Contains(strings.ToLower(ua), "windows") {
			return "Windows", parsedUa.Name, nil // special treatment for https://github.com/muety/wakapi/issues/765
		}
	}
	return "", "", errors.New("failed to parse user agent string")
}

func RaiseForStatus(res *http.Response, err error) (*http.Response, error) {
	if err != nil {
		return res, err
	}
	if res.StatusCode >= 400 {
		message := "<body omitted or empty>"
		contentType := res.Header.Get("content-type")
		if strings.HasPrefix(contentType, "text/") || strings.HasPrefix(contentType, "application/json") {
			body, _ := io.ReadAll(res.Body)
			res.Body.Close()
			res.Body = io.NopCloser(bytes.NewBuffer(body))
			message = string(body)
		}
		return res, fmt.Errorf("got response status %d for '%s %s' - %s", res.StatusCode, res.Request.Method, res.Request.URL.String(), message)
	}
	return res, nil
}

// extractOS identifies the operating system from user agents parts, looking for "(OS-...)" or standalone "os_arch" patterns.
func extractOS(parts []string) string {
	if len(parts) < 2 {
		return ""
	}

	osPart := parts[1]

	if strings.HasPrefix(osPart, "(") { // handle OS inside parentheses: "(Linux-4.15.0...)"
		osPart = strings.TrimPrefix(osPart, "(")
		osPart = strings.Split(osPart, ")")[0]
		osPart = strings.Split(osPart, " ")[0]
		os := strings.Split(osPart, "-")[0]
		return strings.Split(os, "_")[0]
	}

	if strings.Contains(osPart, "-") || strings.Contains(osPart, "_") { // handle standalone OS like "linux_x86-64"
		candidate := strings.Split(osPart, "-")[0]
		candidate = strings.Split(candidate, "_")[0]
		if knownOs.Contain(strings.ToLower(candidate)) {
			return candidate
		}
	}

	return ""
}

func extractEditor(ua string, parts []string) string {
	// 1. Prioritize known AI tools (e.g. "Claude/2.1.118")
	for tool := range aiTools {
		if idx := strings.Index(strings.ToLower(ua), tool+"/"); idx != -1 {
			return ua[idx : idx+len(tool)]
		}
	}

	var primaryEditor string
	var wakatimePluginEditor string

	// 2. Try find primary editor from parts
	for i := 1; i < len(parts); i++ {
		p := parts[i]
		if !strings.Contains(p, "/") {
			continue // valid editors usually have a version slash (Editor/1.0)
		}

		name := strings.Split(p, "/")[0]
		nameLower := strings.ToLower(name)

		if nameLower == "wakatime" || editorMiddlewares.Contain(nameLower) {
			continue // skip wakatime core components and known middlewares
		}

		if isRuntime(nameLower) {
			continue // skip programming language runtimes (e.g., "Python3.8.0", "go1.21.3")
		}

		// track plugins ending with "-wakatime" (e.g., "vscode-wakatime")
		if strings.HasSuffix(nameLower, "-wakatime") {
			candidate := strings.TrimSuffix(name, "-wakatime")
			if !knownOs.Contain(strings.ToLower(candidate)) { // make sure to not mistakenly pick up "windows-wakatime" or "linux-wakatime"
				wakatimePluginEditor = candidate
			}
			continue
		}

		if primaryEditor == "" {
			primaryEditor = name
		}
	}

	return condition.Ternary(primaryEditor != "", primaryEditor, wakatimePluginEditor)
}

// isRuntime heuristically checks if a string is a language runtime rather than an editor.
func isRuntime(lowerName string) bool {
	if lowerName == "python" || lowerName == "go" {
		return true
	}
	if strings.HasPrefix(lowerName, "python") && len(lowerName) > 6 && isDigit(lowerName[6]) { // e.g. python3.8.0
		return true
	}
	if strings.HasPrefix(lowerName, "go") && len(lowerName) > 2 && isDigit(lowerName[2]) { // e.g. go1.21.3
		return true
	}
	return false
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}
