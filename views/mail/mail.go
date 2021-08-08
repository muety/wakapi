package mail

import "embed"

//go:embed *.html
var TemplateFiles embed.FS
