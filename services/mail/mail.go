package mail

import (
	"bytes"
	"fmt"
	"time"

	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/routes"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"github.com/muety/wakapi/views/mail"

	conf "github.com/muety/wakapi/config"
)

const (
	tplNamePasswordReset               = "reset_password"
	tplNameImportNotification          = "import_finished"
	tplNameWakatimeFailureNotification = "wakatime_connection_failure"
	tplNameReport                      = "report"
	tplOtp                             = "otp"
	tplNameSubscriptionNotification    = "subscription_expiring"
	subjectPasswordReset               = "Wakana - Password Reset"
	subjectWakanaOtp                   = "Wakana - OTP"
	subjectImportNotification          = "Wakana - Data Import Finished"
	subjectWakatimeFailureNotification = "Wakana - WakaTime Connection Failure"
	subjectReport                      = "Wakana - Report from %s"
	subjectSubscriptionNotification    = "Wakana - Subscription expiring / expired"
)

type SendingService interface {
	Send(*models.Mail) error
}

type MailService struct {
	config         *conf.Config
	sendingService SendingService
	templates      utils.TemplateMap
}

func NewMailService() services.IMailService {
	config := conf.Get()

	var sendingService SendingService
	sendingService = &NoopSendingService{}

	if config.Mail.Enabled {
		if config.Mail.Provider == conf.MailProviderSmtp {
			sendingService = NewSMTPSendingService(config.Mail.Smtp)
		}
	}

	// Use local file system when in 'dev' environment, go embed file system otherwise
	templateFs := conf.ChooseFS("views/mail", mail.TemplateFiles)
	templates, err := utils.LoadTemplates(templateFs, routes.DefaultTemplateFuncs())
	if err != nil {
		panic(err)
	}

	return &MailService{sendingService: sendingService, config: config, templates: templates}
}

func (m *MailService) SendLoginOtp(recipient string, otp string, expiryTime time.Time) error {
	tpl, err := m.getOtpTemplate(LoginOtpTplData{Otp: otp, ExpiryTime: expiryTime.Format("Jan 2, 2006 15:04 (MST -07:00)")})
	if err != nil {
		return err
	}
	mail := &models.Mail{
		From:    models.MailAddress(m.config.Mail.Sender),
		To:      models.MailAddresses([]models.MailAddress{models.MailAddress(recipient)}),
		Subject: subjectWakanaOtp,
	}
	mail.WithHTML(tpl.String())
	return m.sendingService.Send(mail)
}

func (m *MailService) SendPasswordReset(recipient *models.User, resetLink string) error {
	tpl, err := m.getPasswordResetTemplate(PasswordResetTplData{ResetLink: resetLink})
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

func (m *MailService) SendWakatimeFailureNotification(recipient *models.User, numFailures int) error {
	tpl, err := m.getWakatimeFailureNotificationTemplate(WakatimeFailureNotificationNotificationTplData{
		PublicUrl:   m.config.Server.PublicUrl,
		NumFailures: numFailures,
	})
	if err != nil {
		return err
	}
	mail := &models.Mail{
		From:    models.MailAddress(m.config.Mail.Sender),
		To:      models.MailAddresses([]models.MailAddress{models.MailAddress(recipient.Email)}),
		Subject: subjectWakatimeFailureNotification,
	}
	mail.WithHTML(tpl.String())
	return m.sendingService.Send(mail)
}

func (m *MailService) SendImportNotification(recipient *models.User, duration time.Duration, numHeartbeats int) error {
	tpl, err := m.getImportNotificationTemplate(ImportNotificationTplData{
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
	tpl, err := m.getReportTemplate(ReportTplData{report})
	if err != nil {
		return err
	}
	mail := &models.Mail{
		From:    models.MailAddress(m.config.Mail.Sender),
		To:      models.MailAddresses([]models.MailAddress{models.MailAddress(recipient.Email)}),
		Subject: fmt.Sprintf(subjectReport, helpers.FormatDateHuman(time.Now().In(recipient.TZ()))),
	}
	mail.WithHTML(tpl.String())
	return m.sendingService.Send(mail)
}

func (m *MailService) SendSubscriptionNotification(recipient *models.User, hasExpired bool) error {
	tpl, err := m.getSubscriptionNotificationTemplate(SubscriptionNotificationTplData{
		PublicUrl:           m.config.Server.PublicUrl,
		DataRetentionMonths: m.config.App.DataRetentionMonths,
		HasExpired:          hasExpired,
	})
	if err != nil {
		return err
	}
	mail := &models.Mail{
		From:    models.MailAddress(m.config.Mail.Sender),
		To:      models.MailAddresses([]models.MailAddress{models.MailAddress(recipient.Email)}),
		Subject: subjectSubscriptionNotification,
	}
	mail.WithHTML(tpl.String())
	return m.sendingService.Send(mail)
}

func (m *MailService) getPasswordResetTemplate(data PasswordResetTplData) (*bytes.Buffer, error) {
	var rendered bytes.Buffer
	if err := m.templates[m.fmtName(tplNamePasswordReset)].Execute(&rendered, data); err != nil {
		return nil, err
	}
	return &rendered, nil
}

func (m *MailService) getOtpTemplate(data LoginOtpTplData) (*bytes.Buffer, error) {
	var rendered bytes.Buffer
	if err := m.templates[m.fmtName(tplOtp)].Execute(&rendered, data); err != nil {
		return nil, err
	}
	return &rendered, nil
}

func (m *MailService) getWakatimeFailureNotificationTemplate(data WakatimeFailureNotificationNotificationTplData) (*bytes.Buffer, error) {
	var rendered bytes.Buffer
	if err := m.templates[m.fmtName(tplNameWakatimeFailureNotification)].Execute(&rendered, data); err != nil {
		return nil, err
	}
	return &rendered, nil
}

func (m *MailService) getImportNotificationTemplate(data ImportNotificationTplData) (*bytes.Buffer, error) {
	var rendered bytes.Buffer
	if err := m.templates[m.fmtName(tplNameImportNotification)].Execute(&rendered, data); err != nil {
		return nil, err
	}
	return &rendered, nil
}

func (m *MailService) getReportTemplate(data ReportTplData) (*bytes.Buffer, error) {
	var rendered bytes.Buffer
	if err := m.templates[m.fmtName(tplNameReport)].Execute(&rendered, data); err != nil {
		return nil, err
	}
	return &rendered, nil
}

func (m *MailService) getSubscriptionNotificationTemplate(data SubscriptionNotificationTplData) (*bytes.Buffer, error) {
	var rendered bytes.Buffer
	if err := m.templates[m.fmtName(tplNameSubscriptionNotification)].Execute(&rendered, data); err != nil {
		return nil, err
	}
	return &rendered, nil
}

func (m *MailService) fmtName(name string) string {
	return fmt.Sprintf("%s.tpl.html", name)
}
