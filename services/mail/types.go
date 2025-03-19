package mail

import (
	"github.com/muety/wakapi/models"
)

type PasswordResetTplData struct {
	ResetLink string
}

type LoginOtpTplData struct {
	Otp        string
	ExpiryTime string
}

type ImportNotificationTplData struct {
	PublicUrl     string
	Duration      string
	NumHeartbeats int
}

type WakatimeFailureNotificationNotificationTplData struct {
	PublicUrl   string
	NumFailures int
}

type ReportTplData struct {
	Report *models.Report
}

type SubscriptionNotificationTplData struct {
	PublicUrl           string
	HasExpired          bool
	DataRetentionMonths int
}
