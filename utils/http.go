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

// Canonical list of AI parsers and harness tokens from wakatime-cli.
// In wakatime-cli, AI tools identify themselves either via their parser name (e.g. "Claude", "Cursor", "Windsurf", "Copilot")
// or via dedicated harness tokens (e.g. "claude-code", "codex-cli", "opencode-cli", "github-copilot-cli", "antigravity-cli").
// See https://github.com/wakatime/wakatime-cli/blob/cb6c885aa57ec70f55acbb581c25fd3d367db853/pkg/ai/ai.go#L51-L138
// and https://github.com/wakatime/wakatime-cli/blob/cb6c885aa57ec70f55acbb581c25fd3d367db853/pkg/ai/ai.go#L883-L989
var aiTools = set.New[string](
	"claude", "claude-code",
	"chatgpt",
	"copilot", "github-copilot-cli",
	"codex", "codex-cli",
	"continue",
	"cody",
	"roo-code", "roocode",
	"opencode", "opencode-cli",
	"cursor", "cursor-agent", "cursor-agent-cli",
	"windsurf",
	"qoder",
	"kiro",
	"cline", "cline-cli",
	"qwen-code", "qwen-code-cli",
	"pi",
	"goose",
	"amp",
	"grok-build",
	"codebuff", "codebuff-cli",
	"codewhale", "codewhale-cli",
	"crush", "crush-cli",
	"devin",
	"droid", "droid-cli",
	"forge", "forge-cli",
	"hermes",
	"ibm-bob", "ibm-bob-cli",
	"kilo-code", "kilo-code-cli",
	"kimi-code", "kimi-code-cli",
	"lingtai-tui",
	"mistral-vibe", "mistral-vibe-cli",
	"mux",
	"omp", "omp-cli",
	"openclaw", "openclaw-cli",
	"open-design",
	"quickdesk",
	"warp",
	"zcode",
	"zed",
	"zerostack",
)

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

// ParsedUserAgent holds the result of parsing a User-Agent string.
type ParsedUserAgent struct {
	OS      string
	Editor  string
	AiModel string
}

// ParseUserAgent extracts the operating system, editor and – if present – the
// AI model from a User-Agent string.
func ParseUserAgent(ua string) (ParsedUserAgent, error) {
	ua = strings.TrimSpace(ua)
	if ua == "" {
		return ParsedUserAgent{}, errors.New("empty user agent")
	}

	// Try to parse WakaTime and browser extension user agents
	if parts := strings.Fields(ua); len(parts) >= 2 {
		first := strings.ToLower(parts[0])

		if strings.HasPrefix(first, "wakatime/") ||
			strings.HasPrefix(first, "chrome/") ||
			strings.HasPrefix(first, "firefox/") ||
			strings.HasPrefix(first, "edge/") {

			aiModel := extractAiModel(ua, parts)
			editor := extractEditor(ua, parts, aiModel)
			if editor == "KTextEditor" { // special treatment for neovim
				editor = "kate"
			}
			if editor == "claude-code" { // special treatment for Claude Code
				editor = "Claude"
			}

			os := extractOS(parts)
			if os == "" && strings.HasPrefix(first, "wakatime/") {
				return ParsedUserAgent{}, errors.New("failed to parse os from wakatime user agent")
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

			return ParsedUserAgent{
				OS:      os,
				Editor:  editor,
				AiModel: aiModel,
			}, nil
		}
	}

	// Try parse browser user agent as a fallback
	if parsedUa := useragent.Parse(ua); len(parsedUa.Name) > 0 {
		if len(parsedUa.OS) > 0 {
			return ParsedUserAgent{
				OS:     strutil.Capitalize(parsedUa.OS),
				Editor: parsedUa.Name,
			}, nil
		} else if strings.Contains(strings.ToLower(ua), "windows") {
			return ParsedUserAgent{
				OS:     "Windows",
				Editor: parsedUa.Name, // special treatment for https://github.com/muety/wakapi/issues/765
			}, nil
		}
	}
	return ParsedUserAgent{}, errors.New("failed to parse user agent string")
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

// isAiHarness checks whether a token represents an AI harness or parser.
// wakatime-cli AI harnesses conventionally use a "-cli" / "-tui" suffix,
// an "antigravity-" / "codex-" prefix, or match a known AI parser name in aiTools.
func isAiHarness(lowerName string) bool {
	if strings.HasSuffix(lowerName, "-cli") ||
		strings.HasSuffix(lowerName, "-tui") ||
		strings.HasPrefix(lowerName, "antigravity-") ||
		strings.HasPrefix(lowerName, "codex-") {
		return true
	}
	return aiTools.Contain(lowerName)
}

func extractEditor(ua string, parts []string, aiModel string) string {
	var primaryEditor string
	var wakatimePluginEditor string
	var aiHarness string

	aiModelLower := strings.ToLower(aiModel)

	// Scan parts for editors, AI harnesses, and plugins
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

		// skip the AI model token so it is not picked as the editor
		if aiModelLower != "" && nameLower == aiModelLower {
			continue
		}

		// track known AI harness tokens (e.g., "claude-code", "Claude", "codex-cli") so that the harness (not the model!) is reported as the editor
		if isAiHarness(nameLower) {
			if aiHarness == "" {
				aiHarness = name
			}
			continue
		}

		if primaryEditor == "" {
			primaryEditor = name
		}
	}

	// prefer AI harness, then primary editor, then wakatime plugin editor
	return condition.Ternary[bool, string](
		aiHarness != "", aiHarness,
		condition.Ternary[bool, string](primaryEditor != "", primaryEditor, wakatimePluginEditor),
	)
}

// extractAiModel returns the AI model name (e.g. "opus", "gpt", "gemini", "composer", "swe", "M")
// from a user agent string, or an empty string if the user agent does not contain an AI model token.
//
// AI session heartbeats that include a model format the user agent by prepending the model (e.g. "opus/4.1-medium", "gpt/5.5-high", "gemini/3-flash-preview") in front of the AI harness or IDE editor token
// See https://github.com/wakatime/wakatime-cli/blob/cb6c885aa57ec70f55acbb581c25fd3d367db853/pkg/ai/ai.go#L958-L989.
//
// - If multiple candidate tokens exist and the first token is not an AI harness itself, the first token is the prepended AI model token.
// - If the first token is an AI harness (e.g. "Claude/2.1.118 PyCharm/2023.1"), it represents the harness itself without a separate model token.
// - Single-token or non-AI user agents return empty.
func extractAiModel(ua string, parts []string) string {
	candidates := extractCandidates(parts)
	if len(candidates) < 2 {
		return ""
	}

	c0Name := strings.Split(candidates[0], "/")[0]
	c0Lower := strings.ToLower(c0Name)

	if isAiHarness(c0Lower) {
		return "" // first token is the AI harness itself (e.g. Claude/2.1.118 in PyCharm)
	}

	return c0Name
}

// extractCandidates filters user agent parts for tokens with a version slash,
// excluding wakatime core, middlewares, runtimes, and -wakatime plugins.
func extractCandidates(parts []string) []string {
	var candidates []string
	for i := 1; i < len(parts); i++ {
		p := parts[i]
		if !strings.Contains(p, "/") {
			continue
		}

		name := strings.Split(p, "/")[0]
		nameLower := strings.ToLower(name)

		if nameLower == "wakatime" || editorMiddlewares.Contain(nameLower) {
			continue
		}
		if isRuntime(nameLower) {
			continue
		}
		if strings.HasSuffix(nameLower, "-wakatime") {
			continue
		}

		candidates = append(candidates, p)
	}
	return candidates
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
