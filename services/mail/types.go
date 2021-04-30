package mail

import "github.com/muety/wakapi/models"

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
