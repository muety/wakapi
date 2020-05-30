package models

type User struct {
	ID             string     `json:"id" gorm:"primary_key"`
	ApiKey         string     `json:"api_key" gorm:"unique"`
	Password       string     `json:"-"`
	CreatedAt      CustomTime `gorm:"type:timestamp; default:CURRENT_TIMESTAMP"`
	LastLoggedInAt CustomTime `gorm:"type:timestamp; default:CURRENT_TIMESTAMP"`
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

func (s *Signup) IsValid() bool {
	return len(s.Username) >= 3 &&
		len(s.Password) >= 6 &&
		s.Password == s.PasswordRepeat
}
