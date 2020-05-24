package models

type User struct {
	ID       string `json:"id" gorm:"primary_key"`
	ApiKey   string `json:"api_key" gorm:"unique"`
	Password string `json:"-"`
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
