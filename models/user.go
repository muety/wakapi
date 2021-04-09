package models

import "regexp"

func init() {
	mailRegex = regexp.MustCompile(MailPattern)
}

type User struct {
	ID               string     `json:"id" gorm:"primary_key"`
	ApiKey           string     `json:"api_key" gorm:"unique"`
	Email            string     `json:"email" gorm:"uniqueIndex:idx_user_email"`
	Password         string     `json:"-"`
	CreatedAt        CustomTime `gorm:"type:timestamp; default:CURRENT_TIMESTAMP" swaggertype:"string" format:"date" example:"2006-01-02 15:04:05.000"`
	LastLoggedInAt   CustomTime `gorm:"type:timestamp; default:CURRENT_TIMESTAMP" swaggertype:"string" format:"date" example:"2006-01-02 15:04:05.000"`
	ShareDataMaxDays int        `json:"-" gorm:"default:0"`
	ShareEditors     bool       `json:"-" gorm:"default:false; type:bool"`
	ShareLanguages   bool       `json:"-" gorm:"default:false; type:bool"`
	ShareProjects    bool       `json:"-" gorm:"default:false; type:bool"`
	ShareOSs         bool       `json:"-" gorm:"default:false; type:bool; column:share_oss"`
	ShareMachines    bool       `json:"-" gorm:"default:false; type:bool"`
	IsAdmin          bool       `json:"-" gorm:"default:false; type:bool"`
	HasData          bool       `json:"-" gorm:"default:false; type:bool"`
	WakatimeApiKey   string     `json:"-"`
	ResetToken       string     `json:"-"`
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
	Email string `schema:"email"`
}

type TimeByUser struct {
	User string
	Time CustomTime
}

type CountByUser struct {
	User  string
	Count int64
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
	return ValidateEmail(r.Email)
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
