package models

import "strings"

// AliasResolver returns the alias of an Entity, given its original name. I.e., it returns Alias.Key, given an Alias.Value
type AliasResolver func(t uint8, k string) string

// AliasReverseResolver returns all original names, which have the given alias as mapping target. I.e., it returns a list of Alias.Value, given an Alias.Key
type AliasReverseResolver func(t uint8, k string) []string

type Alias struct {
	ID     uint   `gorm:"primary_key"`
	Type   uint8  `gorm:"not null; index:idx_alias_type_key"`
	User   *User  `json:"-" gorm:"not null; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	UserID string `gorm:"not null; index:idx_alias_user"`
	Key    string `gorm:"not null; index:idx_alias_type_key"`
	Value  string `gorm:"not null"`
}

func (a *Alias) IsValid() bool {
	return a.Key != "" &&
		a.Value != "" &&
		a.validateType() &&
		a.validateWildcard()
}

func (a *Alias) validateType() bool {
	for _, t := range SummaryTypes() {
		if a.Type == t {
			return true
		}
	}
	return false
}

func (a *Alias) validateWildcard() bool {
	if !strings.Contains(a.Value, "*") && !strings.Contains(a.Value, "?") {
		return true
	}
	v := a.Value
	v = strings.ReplaceAll(v, "*", "")
	v = strings.ReplaceAll(v, "?", "")
	return len(v) >= 3 // don't allow "*" or "a*" or sth.
}
