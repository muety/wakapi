package models

import (
	"fmt"
	"github.com/emvi/logbuch"
	"github.com/mitchellh/hashstructure/v2"
	"strings"
	"time"
)

type Heartbeat struct {
	ID              uint       `gorm:"primary_key" hash:"ignore"`
	User            *User      `json:"-" gorm:"not null; constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" hash:"ignore"`
	UserID          string     `json:"-" gorm:"not null; index:idx_time_user"`
	Entity          string     `json:"entity" gorm:"not null; index:idx_entity"`
	Type            string     `json:"type"`
	Category        string     `json:"category"`
	Project         string     `json:"project"`
	Branch          string     `json:"branch"`
	Language        string     `json:"language" gorm:"index:idx_language"`
	IsWrite         bool       `json:"is_write"`
	Editor          string     `json:"editor" hash:"ignore"`           // ignored because editor might be parsed differently by wakatime
	OperatingSystem string     `json:"operating_system" hash:"ignore"` // ignored because os might be parsed differently by wakatime
	Machine         string     `json:"machine" hash:"ignore"`          // ignored because wakatime api doesn't return machines currently
	UserAgent       string     `json:"user_agent" hash:"ignore"`
	Time            CustomTime `json:"time" gorm:"type:timestamp; index:idx_time,idx_time_user" swaggertype:"primitive,number"`
	Hash            string     `json:"-" gorm:"type:varchar(17); uniqueIndex"`
	Origin          string     `json:"-" hash:"ignore"`
	OriginId        string     `json:"-" hash:"ignore"`
	CreatedAt       CustomTime `json:"created_at" gorm:"type:timestamp" swaggertype:"primitive,number" hash:"ignore"` // https://gorm.io/docs/conventions.html#CreatedAt
}

func (h *Heartbeat) Valid() bool {
	return h.User != nil && h.UserID != "" && h.User.ID == h.UserID && h.Time != CustomTime(time.Time{})
}

func (h *Heartbeat) Normalize() {
	h.Language = strings.Title(strings.ToLower(h.Language))
	h.OperatingSystem = strings.Title(strings.ToLower(h.OperatingSystem))
	h.Editor = strings.Title(strings.ToLower(h.Editor))
}

func (h *Heartbeat) Augment(languageMappings map[string]string) {
	maxPrec := -1 // precision / mapping complexity -> more concrete ones shall take precedence
	for ending, value := range languageMappings {
		if ok, prec := strings.HasSuffix(h.Entity, "."+ending), strings.Count(ending, "."); ok && prec > maxPrec {
			h.Language = value
			maxPrec = prec
		}
	}
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

func (h *Heartbeat) String() string {
	return fmt.Sprintf(
		"Heartbeat {user=%s, entity=%s, type=%s, category=%s, project=%s, branch=%s, language=%s, iswrite=%v, editor=%s, os=%s, machine=%s, time=%d}",
		h.UserID,
		h.Entity,
		h.Type,
		h.Category,
		h.Project,
		h.Branch,
		h.Language,
		h.IsWrite,
		h.Editor,
		h.OperatingSystem,
		h.Machine,
		(time.Time(h.Time)).UnixNano(),
	)
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
