package models

import (
	"crypto/md5"
	"fmt"
	"regexp"
	"strings"
	"time"
)

func init() {
	mailRegex = regexp.MustCompile(MailPattern)
}

type User struct {
	ID                string     `json:"id" gorm:"primary_key"`
	ApiKey            string     `json:"api_key" gorm:"unique; default:NULL"`
	Email             string     `json:"email" gorm:"index:idx_user_email; size:255"`
	Location          string     `json:"location"`
	Password          string     `json:"-"`
	CreatedAt         CustomTime `gorm:"type:timestamp; default:CURRENT_TIMESTAMP" swaggertype:"string" format:"date" example:"2006-01-02 15:04:05.000"`
	LastLoggedInAt    CustomTime `gorm:"type:timestamp; default:CURRENT_TIMESTAMP" swaggertype:"string" format:"date" example:"2006-01-02 15:04:05.000"`
	ShareDataMaxDays  int        `json:"-" gorm:"default:0"`
	ShareEditors      bool       `json:"-" gorm:"default:false; type:bool"`
	ShareLanguages    bool       `json:"-" gorm:"default:false; type:bool"`
	ShareProjects     bool       `json:"-" gorm:"default:false; type:bool"`
	ShareOSs          bool       `json:"-" gorm:"default:false; type:bool; column:share_oss"`
	ShareMachines     bool       `json:"-" gorm:"default:false; type:bool"`
	ShareLabels       bool       `json:"-" gorm:"default:false; type:bool"`
	IsAdmin           bool       `json:"-" gorm:"default:false; type:bool"`
	HasData           bool       `json:"-" gorm:"default:false; type:bool"`
	WakatimeApiKey    string     `json:"-"` // for relay middleware and imports
	WakatimeApiUrl    string     `json:"-"` // for relay middleware and imports
	ResetToken        string     `json:"-"`
	ReportsWeekly     bool       `json:"-" gorm:"default:false; type:bool"`
	PublicLeaderboard bool       `json:"-" gorm:"default:false; type:bool"`
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
	ReportsWeekly     bool   `schema:"reports_weekly"`
	PublicLeaderboard bool   `schema:"public_leaderboard"`
}

type TimeByUser struct {
	User string
	Time CustomTime
}

type CountByUser struct {
	User  string
	Count int64
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

// WakaTimeURL returns the user's effective WakaTime URL, i.e. a custom one (which could also point to another Wakapi instance) or fallback if not specified otherwise.
func (u *User) WakaTimeURL(fallback string) string {
	if u.WakatimeApiUrl != "" {
		return strings.TrimSuffix(u.WakatimeApiUrl, "/")
	}
	return fallback
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
	return ValidateUsername(s.Username) &&
		ValidateEmail(s.Email) &&
		ValidatePassword(s.Password) &&
		s.Password == s.PasswordRepeat
}

func (r *UserDataUpdate) IsValid() bool {
	return ValidateEmail(r.Email) && ValidateTimezone(r.Location)
}

func ValidateUsername(username string) bool {
	return len(username) >= 1 && username != "current"
}

func ValidatePassword(password string) bool {
	return len(password) >= 6
}

func ValidateEmail(email string) bool {
	return email == "" || mailRegex.Match([]byte(email))
}

func ValidateTimezone(tz string) bool {
	_, err := time.LoadLocation(tz)
	return err == nil
}
