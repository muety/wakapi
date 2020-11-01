package models

type LanguageMapping struct {
	ID        uint   `json:"id" gorm:"primary_key"`
	User      *User  `json:"-" gorm:"not null"`
	UserID    string `json:"-" gorm:"not null; index:idx_language_mapping_user; uniqueIndex:idx_language_mapping_composite"`
	Extension string `json:"extension" gorm:"uniqueIndex:idx_language_mapping_composite"`
	Language  string `json:"language"`
}

func validateLanguage(language string) bool {
	return len(language) >= 1
}

func validateExtension(extension string) bool {
	return len(extension) >= 1
}
