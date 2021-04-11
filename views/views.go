package views

import "embed"

//go:embed *.html mail/*.html
var TemplateFiles embed.FS
