package models

import (
	"time"

	"github.com/muety/wakapi/utils"
)

type UserAgent struct {
	Id        string    `json:"id"`
	Value     string    `json:"value"`
	Os        string    `json:"os"`
	Editor    string    `json:"editor"`
	FirstSeen time.Time `gorm:"column:first_seen"`
	LastSeen  time.Time `gorm:"column:last_seen"`
}

func (ua *UserAgent) WithId() *UserAgent {
	ua.Id, _ = utils.UUIDFromSeed(ua.Value)
	return ua
}
