package utils

import (
	"encoding/json"
	"html/template"
	"io/fs"
	"io/ioutil"
	"path"
)

type TemplateMap map[string]*template.Template

func Json(data interface{}) template.JS {
	d, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return template.JS(d)
}

func ToRunes(s string) (r []string) {
	for _, c := range []rune(s) {
		r = append(r, string(c))
	}
	return r
}

func LoadTemplates(templateFs fs.FS, funcs template.FuncMap) (TemplateMap, error) {
	tpls := template.New("").Funcs(funcs)
	templates := make(map[string]*template.Template)

	files, err := fs.ReadDir(templateFs, ".")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		tplName := file.Name()
		if file.IsDir() || path.Ext(tplName) != ".html" {
			continue
		}

		templateFile, err := templateFs.Open(tplName)
		if err != nil {
			return nil, err
		}
		templateData, err := ioutil.ReadAll(templateFile)
		if err != nil {
			return nil, err
		}

		templateFile.Close()

		tpl, err := tpls.New(tplName).Parse(string(templateData))
		if err != nil {
			return nil, err
		}

		templates[tplName] = tpl
	}

	return templates, nil
}
