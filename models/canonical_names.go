package models

import (
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/muety/wakapi/config"
	"regexp"
	"strings"
)

// special treatment for system-wide entities (language, editors, os) that are known to commonly cause confusion
// due to being sent by different plugins in different use of capital and small letters in their spelling, e.g. "JAVA"

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9]+`)
var canonicalNames map[uint8]map[string]string

func initLookup() {
	cfg := config.Get()
	canonicalNames = map[uint8]map[string]string{
		SummaryLanguage: cfg.App.GetCanonicalLanguageNames(),
		SummaryEditor:   {},
		SummaryOS:       {},
	}
}

func CanonicalName(value string, entityType uint8) string {
	if canonicalNames == nil {
		initLookup()
	}

	if _, ok := canonicalNames[entityType]; !ok {
		return value
	}

	lookupKey := nonAlphanumericRegex.ReplaceAllString(strings.ToLower(value), "")
	if canonical, ok := canonicalNames[entityType][lookupKey]; ok {
		return canonical
	}

	return strutil.Capitalize(value) // even if no specific canonical name is provided, still always capitalize languages, editors and os for consistency
}
