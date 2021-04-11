package mail

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"text/template"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/views"
)

const (
	tplNamePasswordReset      = "reset_password"
	tplNameImportNotification = "import_finished"
	subjectPasswordReset      = "Wakapi – Password Reset"
	subjectImportNotification = "Wakapi – Data Import Finished"
)

type PasswordResetTplData struct {
	ResetLink string
}

type ImportNotificationTplData struct {
	PublicUrl     string
	Duration      string
	NumHeartbeats int
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

	return template.New(tplName).Parse(string(tplData))
}
