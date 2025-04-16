package routes

import (
	"html/template"
	"net/url"
	"strings"

	"github.com/duke-git/lancet/v2/strutil"
	"github.com/muety/wakapi/helpers"

	"github.com/duke-git/lancet/v2/datetime"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
)

func DefaultTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"json":           utils.Json,
		"date":           helpers.FormatDateHuman,
		"datetime":       helpers.FormatDateTimeHuman,
		"simpledate":     helpers.FormatDate,
		"simpledatetime": helpers.FormatDateTime,
		"duration":       helpers.FmtWakatimeDurationV2,
		"floordate":      datetime.BeginOfDay,
		"ceildate":       utils.CeilDate,
		"title":          strings.Title,
		"join":           strings.Join,
		"add":            add,
		"capitalize":     strutil.Capitalize,
		"lower":          strings.ToLower,
		"toRunes":        utils.ToRunes,
		"localTZOffset":  utils.LocalTZOffset,
		"entityTypes":    models.SummaryTypes,
		"strslice":       utils.SubSlice[string],
		"typeName":       typeName,
		"frontendUri": func() string {
			return config.Get().Server.FrontendUri
		},
		"isDev": func() bool {
			return config.Get().IsDev()
		},
		"getBasePath": func() string {
			return config.Get().Server.BasePath
		},
		"getVersion": func() string {
			return config.Get().Version
		},
		"getDbType": func() string {
			return strings.ToLower(config.Get().Db.Type)
		},
		"htmlSafe": func(html string) template.HTML {
			return template.HTML(html)
		},
		"urlSafe": func(s string) template.URL {
			return template.URL(url.QueryEscape(s))
		},
		"cssSafe": func(s string) template.CSS {
			return template.CSS(s)
		},
		"avatarUrlTemplate": func() string {
			return config.Get().App.AvatarURLTemplate
		},
		"defaultWakatimeUrl": func() string {
			return config.WakatimeApiUrl
		},
	}
}

func typeName(t uint8) string {
	if t == models.SummaryProject {
		return "project"
	}
	if t == models.SummaryLanguage {
		return "language"
	}
	if t == models.SummaryEditor {
		return "editor"
	}
	if t == models.SummaryOS {
		return "operating system"
	}
	if t == models.SummaryMachine {
		return "machine"
	}
	if t == models.SummaryLabel {
		return "label"
	}
	if t == models.SummaryBranch {
		return "branch"
	}
	if t == models.SummaryEntity {
		return "entity"
	}
	if t == models.SummaryCategory {
		return "category"
	}
	return "unknown"
}

func add(i, j int) int {
	return i + j
}
