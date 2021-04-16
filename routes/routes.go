package routes

import (
	"fmt"
	"html/template"
	"io/fs"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"github.com/muety/wakapi/views"
)

func Init() {
	loadTemplates()
}

type action func(w http.ResponseWriter, r *http.Request) (int, string, string)

var templates map[string]*template.Template

func loadTemplates() {
	const tplPath = "/views"
	tpls := template.New("").Funcs(template.FuncMap{
		"json":           utils.Json,
		"date":           utils.FormatDateHuman,
		"simpledate":     utils.FormatDate,
		"simpledatetime": utils.FormatDateTime,
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
	})
	templates = make(map[string]*template.Template)

	files, err := fs.ReadDir(views.TemplateFiles, ".")
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		tplName := file.Name()
		if file.IsDir() || path.Ext(tplName) != ".html" {
			continue
		}

		templateFile, err := views.TemplateFiles.Open(tplName)
		if err != nil {
			panic(err)
		}
		templateData, err := ioutil.ReadAll(templateFile)
		if err != nil {
			panic(err)
		}

		templateFile.Close()

		tpl, err := tpls.New(tplName).Parse(string(templateData))
		if err != nil {
			panic(err)
		}

		templates[tplName] = tpl
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
	return "unknown"
}

func defaultErrorRedirectTarget() string {
	return fmt.Sprintf("%s/?error=unauthorized", config.Get().Server.BasePath)
}
