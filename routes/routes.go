package routes

import (
	"fmt"
	"github.com/muety/wakapi/views"
	"html/template"
	"net/http"
	"strings"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
)

type action func(w http.ResponseWriter, r *http.Request) (int, string, string)

var templates map[string]*template.Template

func Init() {
	loadTemplates()
}

func DefaultTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"json":           utils.Json,
		"date":           utils.FormatDateHuman,
		"datetime":       utils.FormatDateTimeHuman,
		"simpledate":     utils.FormatDate,
		"simpledatetime": utils.FormatDateTime,
		"duration":       utils.FmtWakatimeDuration,
		"floordate":      utils.FloorDate,
		"ceildate":       utils.CeilDate,
		"title":          strings.Title,
		"join":           strings.Join,
		"add":            utils.Add,
		"capitalize":     utils.Capitalize,
		"toRunes":        utils.ToRunes,
		"entityTypes":    models.SummaryTypes,
		"typeName":       typeName,
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
		"avatarUrlTemplate": func() string {
			return config.Get().App.AvatarURLTemplate
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
	return "unknown"
}

func loadTemplates() {
	// Use local file system when in 'dev' environment, go embed file system otherwise
	templateFs := config.ChooseFS("views", views.TemplateFiles)
	if tpls, err := utils.LoadTemplates(templateFs, DefaultTemplateFuncs()); err == nil {
		templates = tpls
	} else {
		panic(err)
	}
}

func defaultErrorRedirectTarget() string {
	return fmt.Sprintf("%s/?error=unauthorized", config.Get().Server.BasePath)
}
