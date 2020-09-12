package models

type User struct {
	ID             string     `json:"id" gorm:"primary_key"`
	ApiKey         string     `json:"api_key" gorm:"unique"`
	Password       string     `json:"-"`
	CreatedAt      CustomTime `gorm:"type:timestamp; default:CURRENT_TIMESTAMP"`
	LastLoggedInAt CustomTime `gorm:"type:timestamp; default:CURRENT_TIMESTAMP"`
	BadgesEnabled  bool       `json:"-" gorm:"not null; default:false; type: bool"`
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
	return len(username) >= 3 && username != "current"
}

func validatePassword(password string) bool {
	return len(password) >= 6
}
