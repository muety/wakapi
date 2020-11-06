package models

type LanguageMapping struct {
	ID        uint   `json:"id" gorm:"primary_key"`
	User      *User  `json:"-" gorm:"not null"`
	UserID    string `json:"-" gorm:"not null; index:idx_language_mapping_user; uniqueIndex:idx_language_mapping_composite"`
	Extension string `json:"extension" gorm:"uniqueIndex:idx_language_mapping_composite; type:varchar(16)"`
	Language  string `json:"language" gorm:"type:varchar(64)"`
}

func (m *LanguageMapping) IsValid() bool {
	return m.validateLanguage() && m.validateExtension()
}

func (m *LanguageMapping) validateLanguage() bool {
	return len(m.Language) >= 1
}

func (m *LanguageMapping) validateExtension() bool {
	return len(m.Extension) >= 1
}
