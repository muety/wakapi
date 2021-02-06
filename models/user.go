package models

type User struct {
	ID               string     `json:"id" gorm:"primary_key"`
	ApiKey           string     `json:"api_key" gorm:"unique"`
	Password         string     `json:"-"`
	CreatedAt        CustomTime `gorm:"type:timestamp; default:CURRENT_TIMESTAMP"`
	LastLoggedInAt   CustomTime `gorm:"type:timestamp; default:CURRENT_TIMESTAMP"`
	ShareDataMaxDays uint       `json:"-" gorm:"default:0"`
	ShareEditors     bool       `json:"-" gorm:"default:false; type:bool"`
	ShareLanguages   bool       `json:"-" gorm:"default:false; type:bool"`
	ShareProjects    bool       `json:"-" gorm:"default:false; type:bool"`
	ShareOSs         bool       `json:"-" gorm:"default:false; type:bool; column:share_oss"`
	ShareMachines    bool       `json:"-" gorm:"default:false; type:bool"`
	WakatimeApiKey   string     `json:"-"`
}

type Login struct {
	Username string `schema:"username"`
	Password string `schema:"password"`
}

type Signup struct {
	Username       string `schema:"username"`
	Password       string `schema:"password"`
	PasswordRepeat string `schema:"password_repeat"`
}

type CredentialsReset struct {
	PasswordOld    string `schema:"password_old"`
	PasswordNew    string `schema:"password_new"`
	PasswordRepeat string `schema:"password_repeat"`
}

type TimeByUser struct {
	User string
	Time CustomTime
}

func (c *CredentialsReset) IsValid() bool {
	return validatePassword(c.PasswordNew) &&
		c.PasswordNew == c.PasswordRepeat
}

func (s *Signup) IsValid() bool {
	return validateUsername(s.Username) &&
		validatePassword(s.Password) &&
		s.Password == s.PasswordRepeat
}

func validateUsername(username string) bool {
	return len(username) >= 1 && username != "current"
}

func validatePassword(password string) bool {
	return len(password) >= 6
}
