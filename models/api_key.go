package models

type ApiKey struct {
	ApiKey   string `json:"api_key" gorm:"primary_key"`
	User     *User  `json:"-" gorm:"not null; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	UserID   string `json:"-" gorm:"not null; index:idx_api_key_user"`
	ReadOnly bool   `json:"readonly" gorm:"default:false"`
	Label    string `json:"label" gorm:"type:varchar(64)"`
}

func (k *ApiKey) IsValid() bool {
	return k.ApiKey != "" && k.Label != ""
}
