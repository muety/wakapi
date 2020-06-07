package routes

import (
	"fmt"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"html/template"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
)

func init() {
	loadTemplates()
}

var templates map[string]*template.Template

func loadTemplates() {
	tplPath := "views"
	tpls := template.New("").Funcs(template.FuncMap{
		"json":       utils.Json,
		"date":       utils.FormatDateHuman,
		"title":      strings.Title,
		"capitalize": utils.Capitalize,
		"getBasePath": func() string {
			return models.GetConfig().BasePath
		},
		"getVersion": func() string {
			return models.GetConfig().Version
		},
		"htmlSafe": func(html string) template.HTML {
			return template.HTML(html)
		},
	})
	templates = make(map[string]*template.Template)

	files, err := ioutil.ReadDir(tplPath)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		tplName := file.Name()
		if file.IsDir() || path.Ext(tplName) != ".html" {
			continue
		}

		tpl, err := tpls.New(tplName).ParseFiles(fmt.Sprintf("%s/%s", tplPath, tplName))
		if err != nil {
			panic(err)
		}

		templates[tplName] = tpl
	}
}

func respondAlert(w http.ResponseWriter, error, success, tplName string, status int) {
	w.WriteHeader(status)
	if tplName == "" {
		tplName = "index.tpl.html"
	}
	templates[tplName].Execute(w, struct {
		Error   string
		Success string
	}{Error: error, Success: success})
}

// TODO: do better
func handleAlerts(w http.ResponseWriter, r *http.Request, tplName string) bool {
	if err := r.URL.Query().Get("error"); err != "" {
		if err == "unauthorized" {
			respondAlert(w, err, "", tplName, http.StatusUnauthorized)
		} else {
			respondAlert(w, err, "", tplName, http.StatusInternalServerError)
		}
		return true
	}

	if success := r.URL.Query().Get("success"); success != "" {
		respondAlert(w, "", success, tplName, http.StatusOK)
		return true
	}

	return false
}
