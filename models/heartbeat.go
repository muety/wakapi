package models

import (
	"fmt"
	"github.com/emvi/logbuch"
	"github.com/mitchellh/hashstructure/v2"
	"regexp"
	"time"
)

type Heartbeat struct {
	ID              uint       `gorm:"primary_key"`
	User            *User      `json:"-" gorm:"not null; constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UserID          string     `json:"-" gorm:"not null; index:idx_time_user"`
	Entity          string     `json:"entity" gorm:"not null; index:idx_entity"`
	Type            string     `json:"type"`
	Category        string     `json:"category"`
	Project         string     `json:"project"`
	Branch          string     `json:"branch"`
	Language        string     `json:"language" gorm:"index:idx_language"`
	IsWrite         bool       `json:"is_write"`
	Editor          string     `json:"editor"`
	OperatingSystem string     `json:"operating_system"`
	Machine         string     `json:"machine"`
	Time            CustomTime `json:"time" gorm:"type:timestamp; default:CURRENT_TIMESTAMP; index:idx_time,idx_time_user"`
	Hash            string     `json:"-" gorm:"type:varchar(17); uniqueIndex"`
	languageRegex   *regexp.Regexp
}

func (h *Heartbeat) Valid() bool {
	return h.User != nil && h.UserID != "" && h.User.ID == h.UserID && h.Time != CustomTime(time.Time{})
}

func (h *Heartbeat) Augment(languageMappings map[string]string) {
	if h.languageRegex == nil {
		h.languageRegex = regexp.MustCompile(`^.+\.(.+)$`)
	}
	groups := h.languageRegex.FindAllStringSubmatch(h.Entity, -1)
	if len(groups) == 0 || len(groups[0]) != 2 {
		return
	}
	ending := groups[0][1]
	if _, ok := languageMappings[ending]; !ok {
		return
	}
	h.Language, _ = languageMappings[ending]
}

func (h *Heartbeat) GetKey(t uint8) (key string) {
	switch t {
	case SummaryProject:
		key = h.Project
	case SummaryEditor:
		key = h.Editor
	case SummaryLanguage:
		key = h.Language
	case SummaryOS:
		key = h.OperatingSystem
	case SummaryMachine:
		key = h.Machine
	}

	if key == "" {
		key = UnknownSummaryKey
	}

	return key
}

// Hash is used to prevent duplicate heartbeats
// Using a UNIQUE INDEX over all relevant columns would be more straightforward,
// whereas manually computing this kind of hash is quite cumbersome. However,
// such a unique index would, according to https://stackoverflow.com/q/65980064/3112139,
// essentially double the space required for heartbeats, so we decided to go this way.

func (h *Heartbeat) Hashed() *Heartbeat {
	hash, err := hashstructure.Hash(h, hashstructure.FormatV2, nil)
	if err != nil {
		logbuch.Error("CRITICAL ERROR: failed to hash struct – %v", err)
	}
	h.Hash = fmt.Sprintf("%x", hash) // "uint64 values with high bit set are not supported"
	return h
}
