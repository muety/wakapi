package mail

import (
	"bytes"
	"fmt"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/routes"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"html/template"
	"io/ioutil"
	"time"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/views"
)

const (
	tplNamePasswordReset      = "reset_password"
	tplNameImportNotification = "import_finished"
	tplNameReport             = "report"
	subjectPasswordReset      = "Wakapi - Password Reset"
	subjectImportNotification = "Wakapi - Data Import Finished"
	subjectReport             = "Wakapi - Report from %s"
)

type SendingService interface {
	Send(*models.Mail) error
}

type MailService struct {
	config         *conf.Config
	sendingService SendingService
}

func NewMailService() services.IMailService {
	config := conf.Get()

	var sendingService SendingService
	sendingService = &NoopSendingService{}

	if config.Mail.Enabled {
		if config.Mail.Provider == conf.MailProviderMailWhale {
			sendingService = NewMailWhaleSendingService(config.Mail.MailWhale)
		} else if config.Mail.Provider == conf.MailProviderSmtp {
			sendingService = NewSMTPSendingService(config.Mail.Smtp)
		}
	}

	return &MailService{sendingService: sendingService, config: config}
}

func (m *MailService) SendPasswordReset(recipient *models.User, resetLink string) error {
	tpl, err := getPasswordResetTemplate(PasswordResetTplData{ResetLink: resetLink})
	if err != nil {
		return err
	}
	mail := &models.Mail{
		From:    models.MailAddress(m.config.Mail.Sender),
		To:      models.MailAddresses([]models.MailAddress{models.MailAddress(recipient.Email)}),
		Subject: subjectPasswordReset,
	}
	mail.WithHTML(tpl.String())
	return m.sendingService.Send(mail)
}

func (m *MailService) SendImportNotification(recipient *models.User, duration time.Duration, numHeartbeats int) error {
	tpl, err := getImportNotificationTemplate(ImportNotificationTplData{
		PublicUrl:     m.config.Server.PublicUrl,
		Duration:      fmt.Sprintf("%.0f seconds", duration.Seconds()),
		NumHeartbeats: numHeartbeats,
	})
	if err != nil {
		return err
	}
	mail := &models.Mail{
		From:    models.MailAddress(m.config.Mail.Sender),
		To:      models.MailAddresses([]models.MailAddress{models.MailAddress(recipient.Email)}),
		Subject: subjectImportNotification,
	}
	mail.WithHTML(tpl.String())
	return m.sendingService.Send(mail)
}

func (m *MailService) SendReport(recipient *models.User, report *models.Report) error {
	tpl, err := getReportTemplate(ReportTplData{report})
	if err != nil {
		return err
	}
	mail := &models.Mail{
		From:    models.MailAddress(m.config.Mail.Sender),
		To:      models.MailAddresses([]models.MailAddress{models.MailAddress(recipient.Email)}),
		Subject: fmt.Sprintf(subjectReport, utils.FormatDateHuman(time.Now().In(recipient.TZ()))),
	}
	mail.WithHTML(tpl.String())
	return m.sendingService.Send(mail)
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
