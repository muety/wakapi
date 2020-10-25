package models


type CustomRule struct {
	ID             uint       `json:"id" gorm:"primary_key"`
	User           *User      `json:"-" gorm:"not null"`
	UserID         string     `json:"-" gorm:"not null; index:idx_customrule_user"`
	Extension      string     `json:"extension"`
	Language       string     `json:"language"`
}

func validateLanguage(language string) bool {
	return len(language) >= 1
}

func validateExtension(extension string) bool {
	return len(extension) >= 2
}
