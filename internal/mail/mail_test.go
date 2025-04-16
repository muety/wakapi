package mail

import (
	"fmt"
	"testing"

	"github.com/muety/wakapi/routes"
	"github.com/muety/wakapi/utils"
	"github.com/stretchr/testify/assert"
)

func TestTemplatesLoaded(t *testing.T) {
	templates, err := utils.LoadTemplates(TemplateFiles, routes.DefaultTemplateFuncs())
	if err != nil {
		fmt.Println("ERROR LOADING templates:", err)
		panic(err)
	}

	_, hasReport := templates[fmt.Sprintf("%s.tpl.html", tplNameReport)]
	assert.True(t, hasReport, "The 'report.tpl.html' template should be loaded")
}
