package models

import (
	"crypto/md5"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dchest/captcha"
	"github.com/duke-git/lancet/v2/random"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/utils"
)

const (
	DefaultHeartbeatsTimeout       = 10 * time.Minute
	DefaultHeartbeatsTimeoutLegacy = 2 * time.Minute
	MinHeartbeatsTimeout           = 1 * time.Minute
	MaxHeartbeatsTimeout           = 1 * time.Hour
)

func init() {
	mailRegex = regexp.MustCompile(MailPattern)
}

type User struct {
	ID                     string      `json:"id" gorm:"primary_key"`
	ApiKey                 string      `json:"api_key" gorm:"unique; default:NULL"`
	Email                  string      `json:"email" gorm:"index:idx_user_email; size:255"`
	Location               string      `json:"location"`
	StartOfWeek            int         `json:"start_of_week" gorm:"default:1"`
	Password               string      `json:"-"`
	CreatedAt              CustomTime  `swaggertype:"string" format:"date" example:"2006-01-02 15:04:05.000"` // filled by gorm, see https://gorm.io/docs/conventions.html#CreatedAt
	LastLoggedInAt         CustomTime  `swaggertype:"string" format:"date" example:"2006-01-02 15:04:05.000"` // filled by gorm, see https://gorm.io/docs/conventions.html#CreatedAt
	ShareDataMaxDays       int         `json:"-"`
	ShareEditors           bool        `json:"-" gorm:"default:false; type:bool"`
	ShareLanguages         bool        `json:"-" gorm:"default:false; type:bool"`
	ShareProjects          bool        `json:"-" gorm:"default:false; type:bool"`
	ShareOSs               bool        `json:"-" gorm:"default:false; type:bool; column:share_oss"`
	ShareMachines          bool        `json:"-" gorm:"default:false; type:bool"`
	ShareLabels            bool        `json:"-" gorm:"default:false; type:bool"`
	ShareActivityChart     bool        `json:"-" gorm:"default:false; type:bool"`
	IsAdmin                bool        `json:"-" gorm:"default:false; type:bool"`
	HasData                bool        `json:"-" gorm:"default:false; type:bool"`
	WakatimeApiKey         string      `json:"-"` // for relay middleware and imports
	WakatimeApiUrl         string      `json:"-"` // for relay middleware and imports
	ResetToken             string      `json:"-"`
	UnsubscribeToken       string      `json:"-"`
	ReportsWeekly          bool        `json:"-" gorm:"default:false; type:bool"`
	PublicLeaderboard      bool        `json:"-" gorm:"default:false; type:bool"`
	SubscribedUntil        *CustomTime `json:"-" swaggertype:"string" format:"date" example:"2006-01-02 15:04:05.000"`
	SubscriptionRenewal    *CustomTime `json:"-" swaggertype:"string" format:"date" example:"2006-01-02 15:04:05.000"`
	StripeCustomerId       string      `json:"-"`
	InvitedBy              string      `json:"-"`
	ExcludeUnknownProjects bool        `json:"-"`
	HeartbeatsTimeoutSec   int         `json:"-" gorm:"default:600"` // https://github.com/muety/wakapi/issues/156
	ReadmeStatsBaseUrl     string      `json:"-" gorm:"default:''"`
	AuthType               string      `json:"auth_type" gorm:"default:local; size:255"`
	Sub                    string      `json:"sub" gorm:"size:255"` // openid connect subject
}

type Login struct {
	Username string `schema:"username"`
	Password string `schema:"password"`
}

type Signup struct {
	Username       string `schema:"username"`
	Email          string `schema:"email"`
	Password       string `schema:"password"`
	PasswordRepeat string `schema:"password_repeat"`
	Location       string `schema:"location"`
	CaptchaId      string `schema:"captcha_id"`
	Captcha        string `schema:"captcha"`
	InviteCode     string `schema:"invite_code"`
	InvitedBy      string `schema:"-"`
	OidcProvider   string `schema:"-"`
	OidcSubject    string `schema:"-"`
}

type SetPasswordRequest struct {
	Password       string `schema:"password"`
	PasswordRepeat string `schema:"password_repeat"`
	Token          string `schema:"token"`
}

type ResetPasswordRequest struct {
	Email string `schema:"email"`
}

type CredentialsReset struct {
	PasswordOld    string `schema:"password_old"`
	PasswordNew    string `schema:"password_new"`
	PasswordRepeat string `schema:"password_repeat"`
}

type UserDataUpdate struct {
	Email             string `schema:"email"`
	Location          string `schema:"location"`
	StartOfWeek       int    `schema:"start_of_week"`
	ReportsWeekly     bool   `schema:"reports_weekly"`
	PublicLeaderboard bool   `schema:"public_leaderboard"`
}

type RangeByUser struct {
	User  string
	First CustomTime
	Last  CustomTime
}

type TimeByUser struct {
	User string
	Time CustomTime
}

type CountByUser struct {
	User  string
	Count int64
}

func SignupFromOidcIdToken(idToken *conf.IdTokenPayload) *Signup {
	tempPassword := random.RandString(32) // users must not be able to log in with password, so we simply set it to a random string that we forget afterwards

	return &Signup{
		Username:       idToken.Username(),
		Email:          idToken.Email,
		Password:       tempPassword,
		PasswordRepeat: tempPassword,
		OidcProvider:   idToken.ProviderName,
		OidcSubject:    idToken.Subject,
	}
}

func (u *User) Identity() string {
	return u.ID
}

func (u *User) TZ() *time.Location {
	if u.Location == "" {
		u.Location = "Local"
	}
	tz, err := time.LoadLocation(u.Location)
	if err != nil {
		return time.Local
	}
	return tz
}

// TZOffset returns the time difference between the user's current time zone and UTC
// TODO: is this actually working??
func (u *User) TZOffset() time.Duration {
	_, offset := time.Now().In(u.TZ()).Zone()
	return time.Duration(offset * int(time.Second))
}

// StartOfWeekDay returns the user's preferred start of week as time.Weekday
func (u *User) StartOfWeekDay() time.Weekday {
	if u.StartOfWeek < 0 || u.StartOfWeek > 6 {
		u.StartOfWeek = 1 // Default to Monday
	}
	return time.Weekday(u.StartOfWeek)
}

func (u *User) AvatarURL(urlTemplate string) string {
	urlTemplate = strings.ReplaceAll(urlTemplate, "{username}", u.ID)
	urlTemplate = strings.ReplaceAll(urlTemplate, "{email}", u.Email)
	if strings.Contains(urlTemplate, "{username_hash}") {
		urlTemplate = strings.ReplaceAll(urlTemplate, "{username_hash}", fmt.Sprintf("%x", md5.Sum([]byte(u.ID))))
	}
	if strings.Contains(urlTemplate, "{email_hash}") {
		urlTemplate = strings.ReplaceAll(urlTemplate, "{email_hash}", fmt.Sprintf("%x", md5.Sum([]byte(u.Email))))
	}
	return urlTemplate
}

func (u *User) HeartbeatsTimeout() time.Duration {
	if u.HeartbeatsTimeoutSec > 0 {
		return time.Duration(u.HeartbeatsTimeoutSec) * time.Second
	}
	return DefaultHeartbeatsTimeout
}

func (u *User) HeartbeatsTimeoutMin() int {
	return int(u.HeartbeatsTimeout() / time.Minute)
}

// WakaTimeURL returns the user's effective WakaTime URL, i.e. a custom one (which could also point to another Wakapi instance) or fallback if not specified otherwise.
func (u *User) WakaTimeURL(fallback string) string {
	if u.WakatimeApiUrl != "" {
		return strings.TrimSuffix(u.WakatimeApiUrl, "/")
	}
	return fallback
}

// HasActiveSubscription returns true if subscriptions are enabled on the server and the user has got one
func (u *User) HasActiveSubscription() bool {
	return conf.Get().Subscriptions.Enabled && u.HasActiveSubscriptionStrict()
}

func (u *User) HasActiveSubscriptionStrict() bool {
	return u.SubscribedUntil != nil && u.SubscribedUntil.T().After(time.Now())
}

// SubscriptionExpiredSince returns if a user's subscription has expiration and the duration since when that happened.
// Returns (false, <negative duration>), if subscription hasn't expired, yet.
// Returns (false, 0), if subscriptions are not enabled in the first place.
// Returns (true, <very long duration>), if the user has never had a subscription.
func (u *User) SubscriptionExpiredSince() (bool, time.Duration) {
	cfg := conf.Get()
	if !cfg.Subscriptions.Enabled {
		return false, 0
	}
	if u.SubscribedUntil == nil {
		return true, 99 * 365 * 24 * time.Hour
	}
	diff := time.Now().Sub(u.SubscribedUntil.T())
	return diff >= 0, diff
}

func (u *User) MinDataAge() time.Time {
	retentionMonths := conf.Get().App.DataRetentionMonths
	if retentionMonths <= 0 || u.HasActiveSubscription() {
		return time.Time{}
	}
	// this is not exactly precise, because of summer / winter time, etc.
	return time.Now().AddDate(0, -retentionMonths, 0)
}

func (u *User) AnyDataShared() bool {
	return u.ShareDataMaxDays != 0 && (u.ShareEditors || u.ShareLanguages || u.ShareProjects || u.ShareOSs || u.ShareMachines || u.ShareLabels)
}

func (u *User) UnsubscribeLink() string {
	return fmt.Sprintf("%s/unsubscribe?token=%s", conf.Get().Server.GetPublicUrl(), u.UnsubscribeToken)
}

func (c *CredentialsReset) IsValid() bool {
	return ValidatePassword(c.PasswordNew) &&
		c.PasswordNew == c.PasswordRepeat
}

func (c *SetPasswordRequest) IsValid() bool {
	return ValidatePassword(c.Password) &&
		c.Password == c.PasswordRepeat
}

func (s *Signup) IsValid() bool {
	config := conf.Get()

	captchaValid := s.OidcProvider != "" || !config.Security.SignupCaptcha
	if !captchaValid && config.Security.SignupCaptcha {
		captchaValid = ValidateCaptcha(s.CaptchaId, s.Captcha)
	}

	return ValidateUsername(s.Username) &&
		ValidateEmail(s.Email) &&
		ValidatePassword(s.Password) &&
		s.Password == s.PasswordRepeat &&
		captchaValid
}

func (r *UserDataUpdate) IsValid() bool {
	return ValidateEmail(r.Email) && ValidateTimezone(r.Location) && ValidateStartOfWeek(r.StartOfWeek)
}

func ValidateUsername(username string) bool {
	return len(username) >= 1 && username != "current" && !strings.Contains(username, " ")
}

func ValidatePassword(password string) bool {
	return len(password) >= 6
}

func ValidateCaptcha(captchaId, captchaValue string) bool {
	return captcha.VerifyString(captchaId, captchaValue)
}

// ValidateEmail checks that, if an email address is given, it has proper syntax and (if not in dev mode) an MX record exists for the domain
func ValidateEmail(email string) bool {
	cfg := conf.Get()
	verifyMx := !cfg.IsDev() && !cfg.Mail.SkipVerifyMXRecord
	return email == "" || (mailRegex.MatchString(email) && (!verifyMx || utils.CheckEmailMX(email)))
}

func ValidateTimezone(tz string) bool {
	_, err := time.LoadLocation(tz)
	return err == nil
}

func ValidateStartOfWeek(startOfWeek int) bool {
	return startOfWeek >= 0 && startOfWeek <= 6
}
