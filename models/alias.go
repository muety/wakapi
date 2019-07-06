package models

type Alias struct {
	ID     uint   `gorm:"primary_key"`
	Type   uint8  `gorm:"not null; index:idx_alias_type_key"`
	UserID string `gorm:"not null; index:idx_alias_user"`
	Key    string `gorm:"not null; index:idx_alias_type_key"`
	Value  string `gorm:"not null"`
}
