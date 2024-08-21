package models

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/duke-git/lancet/v2/strutil"
	"github.com/mileusna/useragent"
	uuid "github.com/satori/go.uuid"
)

type PluginUserAgent struct {
	ID                 string     `gorm:"type:uuid;primary_key;column:id" json:"id"`
	UserID             string     `gorm:"type:uuid;" json:"-"`
	CliVersion         string     `json:"cli_version"`
	Editor             string     `json:"editor"`
	GoVersion          string     `json:"go_version"`
	IsBrowserExtension bool       `json:"is_browser_extension"`
	IsDesktopApp       bool       `json:"is_desktop_app"`
	Plugin             *string    `json:"plugin"`
	OS                 string     `json:"os"`
	Value              string     `json:"value"`
	Version            *string    `json:"version"`
	CreatedAt          CustomTime `json:"created_at" gorm:"default:CURRENT_TIMESTAMP" swaggertype:"string" format:"date" example:"2006-01-02 15:04:05.000"`
	LastSeenAt         CustomTime `json:"last_seen_at" gorm:"default:CURRENT_TIMESTAMP" swaggertype:"string" format:"date" example:"2006-01-02 15:04:05.000"`
}

// duplicate of ParseUserAgent
func parseUserAgent(ua string) (string, string, error) { // os, editor, err
	// try parse wakatime client user agents
	re := regexp.MustCompile(`(?iU)^(?:(?:wakatime|chrome|firefox|edge)\/(?:v?[\d+.]+|unset)?\s)?(?:\(?(\w+)[-_].*\)?.+\s)?([^\/\s]+)-wakatime\/.+$`)
	if groups := re.FindAllStringSubmatch(ua, -1); len(groups) > 0 && len(groups[0]) == 3 {
		if groups[0][1] == "win" {
			groups[0][1] = "windows"
		}
		if groups[0][2] == "vim" && strings.Contains(ua, "neovim/") {
			groups[0][2] = "neovim"
		}
		return strutil.Capitalize(groups[0][1]), groups[0][2], nil
	}

	if parsed := useragent.Parse(ua); len(parsed.Name) > 0 && len(parsed.OS) > 0 {
		return strutil.Capitalize(parsed.OS), parsed.Name, nil
	}
	return "", "", errors.New("failed to parse user agent string")
}

func IsBrowserUserAgent(userAgent string) bool {
	userAgent = strings.ToLower(userAgent)

	browserPatterns := []string{
		"chrome",  // Chrome browser
		"firefox", // Firefox browser
		"safari",  // Safari browser (Note: Safari also includes 'Version')
		"msie",    // Internet Explorer
		"trident", // Internet Explorer (Trident is used in newer versions)
		"edge",    // Edge browser
	}

	for _, pattern := range browserPatterns {
		if strings.Contains(userAgent, pattern) {
			return true
		}
	}

	browserSpecificPatterns := []string{
		"seamonkey", // SeaMonkey browser
		"opera",     // Opera browser
		"vivaldi",   // Vivaldi browser
		"camino",    // Camino browser
		"konqueror", // Konqueror browser
	}

	for _, pattern := range browserSpecificPatterns {
		if strings.Contains(userAgent, pattern) {
			return true
		}
	}

	return false
}

func ExtractGoVersion(input string) (string, error) {
	// Define a regular expression to match the Go version pattern
	re := regexp.MustCompile(`go([0-9]+\.[0-9]+\.[0-9]+)`)

	// Find the first match of the Go version
	match := re.FindStringSubmatch(input)

	// If no match found, return an error
	if len(match) < 2 {
		return "", fmt.Errorf("go version not found in string")
	}

	// Return the Go version (match[1] contains the version without the "go" prefix)
	return match[1], nil
}

func ExtractWakatimeCliVersion(userAgent string) (string, error) {
	re := regexp.MustCompile(`wakatime/v[0-9]+\.[0-9]+\.[0-9]+`)
	match := re.FindString(userAgent)

	if match == "" {
		return "", fmt.Errorf("wakatime version not found in user-agent string")
	}
	parts := strings.Split(match, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("wakatime version not found in user-agent string")
	}
	return (parts[1]), nil
}

// ExtractPluginInfo extracts the plugin type and version from a user agent string.
func ExtractPluginInfo(userAgent string) (string, string) {
	// Regular expression to match plugin name and version
	// This pattern captures plugin name with optional suffix and version
	re := regexp.MustCompile(`(\w+(-\w+)?-wakatime)/([\d.]+)`)

	matches := re.FindStringSubmatch(userAgent)
	if len(matches) < 4 {
		return "", ""
	}

	pluginType := matches[1] // Capture group 1: plugin name (e.g., vscode-wakatime)
	version := matches[3]    // Capture group 3: version (e.g., 24.5.0)

	return pluginType, version
}

func formatEditor(editor string) string {
	if strings.ToLower(editor) == "vscode" {
		return "VS Code"
	}
	// TODO: add support for others
	return editor
}

func NewPluginUserAgent(agent, user_id string) (*PluginUserAgent, error) {
	parsed := useragent.Parse(agent)
	os, editor, err := parseUserAgent(agent)
	if err != nil {
		return nil, err
	}
	ua := PluginUserAgent{
		UserID:             user_id,
		ID:                 uuid.NewV4().String(),
		Value:              agent,
		OS:                 os,
		Editor:             formatEditor(editor),
		IsBrowserExtension: IsBrowserUserAgent(agent),
		IsDesktopApp:       parsed.Desktop,
		Plugin:             nil,
		Version:            nil,
	}
	goVersion, err := ExtractGoVersion(agent)
	if err == nil {
		ua.GoVersion = goVersion
	}

	cliVersion, err := ExtractGoVersion(agent)
	if err == nil {
		ua.CliVersion = cliVersion
	}

	wakatimeCliVersion, err := ExtractWakatimeCliVersion(agent)
	if err == nil {
		ua.CliVersion = wakatimeCliVersion
	}

	pluginType, version := ExtractPluginInfo(agent)
	if pluginType != "" {
		ua.Version = &version
		ua.Plugin = &pluginType
	}

	return &ua, nil
}
