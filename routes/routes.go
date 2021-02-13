package routes

import (
	"fmt"
	"github.com/markbates/pkger"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"html/template"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
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
		"title":          strings.Title,
		"join":           strings.Join,
		"add":            utils.Add,
		"capitalize":     utils.Capitalize,
		"toRunes":        utils.ToRunes,
		"entityTypes":    models.SummaryTypes,
		"typeName":       typeName,
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

	dir, err := pkger.Open(tplPath)
	if err != nil {
		panic(err)
	}
	defer dir.Close()
	files, err := dir.Readdir(0)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		tplName := file.Name()
		if file.IsDir() || path.Ext(tplName) != ".html" {
			continue
		}

		templateFile, err := pkger.Open(fmt.Sprintf("%s/%s", tplPath, tplName))
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
