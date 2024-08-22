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

// Checks if user agent belongs to a browser
func IsBrowserUserAgent(userAgent string) bool {
	browserPatterns := []string{
		"chrome", "firefox", "safari", "msie", "trident", "edge",
		"seamonkey", "opera", "vivaldi", "camino", "konqueror",
	}
	userAgent = strings.ToLower(userAgent)
	for _, pattern := range browserPatterns {
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

func ParseWakatimeUserAgent(agent, user_id string) (*PluginUserAgent, error) {
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
		Editor:             strutil.Capitalize(editor),
		IsBrowserExtension: IsBrowserUserAgent(agent),
		IsDesktopApp:       parsed.Desktop,
		Plugin:             nil,
		Version:            nil,
	}

	if goVersion, err := ExtractGoVersion(agent); err == nil {
		ua.GoVersion = goVersion
	}

	if wakatimeCliVersion, err := ExtractWakatimeCliVersion(agent); err == nil {
		ua.CliVersion = wakatimeCliVersion
	}

	if pluginType, version := ExtractPluginInfo(agent); pluginType != "" {
		ua.Plugin = &pluginType
		ua.Version = &version
	}

	return &ua, nil
}

type NameVersion struct {
	Name    string
	Version string
}

// Creates a NameVersion object from a string formatted as "name/version"
func MakeNameVersion(input string) (*NameVersion, error) {
	parts := strings.Split(input, "/")
	if len(parts) != 2 {
		return nil, errors.New("invalid input string")
	}

	textPart := parts[0]
	version := parts[1]

	re := regexp.MustCompile(`[^vV]+`)
	numericPart := strings.Join(re.FindAllString(version, -1), "")

	return &NameVersion{Name: textPart, Version: numericPart}, nil
}

// parses the given section based on the index and 'mutates' pluginUserAgent accordingly
func parseSection(index int, portion string, pluginUserAgent *PluginUserAgent) {
	switch index {
	case 0: // wakatime cli
		if nameVersion, err := MakeNameVersion(portion); err == nil {
			pluginUserAgent.CliVersion = nameVersion.Version
		}
	case 1: // os
		parts := strings.Split(portion, "-")
		if len(parts) >= 2 {
			pluginUserAgent.OS = strutil.Capitalize(strings.TrimLeft(parts[0], "("))
		}
	case 2: // go version
		pluginUserAgent.GoVersion = strings.TrimLeft(portion, "go")
	case 3: // Editor
		if editor, err := MakeNameVersion(portion); err == nil {
			pluginUserAgent.Editor = strutil.Capitalize(editor.Name)
		}
	case 4: // Plugin
		if plugin, err := MakeNameVersion(portion); err == nil {
			pluginUserAgent.Plugin = &plugin.Name
			pluginUserAgent.Version = &plugin.Version
		}
	default:
		fmt.Println("No matching index")
	}
}

func NewPluginUserAgent(agent, user_id string) (*PluginUserAgent, error) {
	parsed := useragent.Parse(agent)
	ua := &PluginUserAgent{
		UserID:             user_id,
		Value:              agent,
		IsDesktopApp:       parsed.Desktop,
		ID:                 uuid.NewV4().String(),
		IsBrowserExtension: IsBrowserUserAgent(agent),
	}

	parts := strings.Split(agent, " ")
	if len(parts) == 5 {
		for index, portion := range parts {
			parseSection(index, portion, ua)
		}
		return ua, nil
	}
	return ParseWakatimeUserAgent(agent, user_id)
}
