package models

type Alias struct {
	ID     uint   `gorm:"primary_key"`
	Type   uint8  `gorm:"not null; index:idx_alias_type_key"`
	User   *User  `json:"-" gorm:"not null; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	UserID string `gorm:"not null; index:idx_alias_user"`
	Key    string `gorm:"not null; index:idx_alias_type_key"`
	Value  string `gorm:"not null"`
}

func (a *Alias) IsValid() bool {
	return a.Key != "" && a.Value != "" && a.validateType()
}

func (a *Alias) validateType() bool {
	for _, t := range SummaryTypes() {
		if a.Type == t {
			return true
		}
	}
	return false
}
