package utils

import (
	"encoding/json"
	"html/template"
)

func Json(data interface{}) template.JS {
	d, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return template.JS(d)
}
