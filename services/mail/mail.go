package mail

import (
	"bytes"
	"fmt"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/routes"
	"html/template"
	"io/ioutil"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/views"
)

const (
	tplNamePasswordReset      = "reset_password"
	tplNameImportNotification = "import_finished"
	tplNameReport             = "report"
	subjectPasswordReset      = "Wakapi – Password Reset"
	subjectImportNotification = "Wakapi – Data Import Finished"
	subjectReport             = "Wakapi – Your Latest Report"
)

type PasswordResetTplData struct {
	ResetLink string
}

type ImportNotificationTplData struct {
	PublicUrl     string
	Duration      string
	NumHeartbeats int
}

type ReportTplData struct {
	Report *models.Report
}

// Factory
func NewMailService() services.IMailService {
	config := conf.Get()
	if config.Mail.Enabled {
		if config.Mail.Provider == conf.MailProviderMailWhale {
			return NewMailWhaleService(config.Mail.MailWhale, config.Server.PublicUrl)
		} else if config.Mail.Provider == conf.MailProviderSmtp {
			return NewSMTPMailService(config.Mail.Smtp, config.Server.PublicUrl)
		}
	}
	return &NoopMailService{}
}

func getPasswordResetTemplate(data PasswordResetTplData) (*bytes.Buffer, error) {
	tpl, err := loadTemplate(tplNamePasswordReset)
	if err != nil {
		return nil, err
	}
	var rendered bytes.Buffer
	if err := tpl.Execute(&rendered, data); err != nil {
		return nil, err
	}
	return &rendered, nil
}

func getImportNotificationTemplate(data ImportNotificationTplData) (*bytes.Buffer, error) {
	tpl, err := loadTemplate(tplNameImportNotification)
	if err != nil {
		return nil, err
	}
	var rendered bytes.Buffer
	if err := tpl.Execute(&rendered, data); err != nil {
		return nil, err
	}
	return &rendered, nil
}

func getReportTemplate(data ReportTplData) (*bytes.Buffer, error) {
	tpl, err := loadTemplate(tplNameReport)
	if err != nil {
		return nil, err
	}
	var rendered bytes.Buffer
	if err := tpl.Execute(&rendered, data); err != nil {
		return nil, err
	}
	return &rendered, nil
}

func loadTemplate(tplName string) (*template.Template, error) {
	tplFile, err := views.TemplateFiles.Open(fmt.Sprintf("mail/%s.tpl.html", tplName))
	if err != nil {
		return nil, err
	}
	defer tplFile.Close()

	tplData, err := ioutil.ReadAll(tplFile)
	if err != nil {
		return nil, err
	}

	return template.
		New(tplName).
		Funcs(routes.DefaultTemplateFuncs()).
		Parse(string(tplData))
}
