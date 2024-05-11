package models

import (
	"fmt"
	"github.com/emvi/logbuch"
	"github.com/mitchellh/hashstructure/v2"
	"time"
	"unicode"
)

type Duration struct {
	UserID          string        `json:"user_id"`
	Time            CustomTime    `json:"time" hash:"ignore"`
	Duration        time.Duration `json:"duration" hash:"ignore"`
	Project         string        `json:"project"`
	Language        string        `json:"language"`
	Editor          string        `json:"editor"`
	OperatingSystem string        `json:"operating_system"`
	Machine         string        `json:"machine"`
	Category        string        `json:"category"`
	Branch          string        `json:"branch"`
	Entity          string        `json:"Entity"`
	NumHeartbeats   int           `json:"-" hash:"ignore"`
	GroupHash       string        `json:"-" hash:"ignore"`
	excludeEntity   bool          `json:"-" hash:"ignore"`
}

func (d *Duration) HashInclude(field string, v interface{}) (bool, error) {
	if field == "Entity" {
		return !d.excludeEntity, nil
	}
	if field == "Time" ||
		field == "Duration" ||
		field == "NumHeartbeats" ||
		field == "GroupHash" ||
		unicode.IsLower(rune(field[0])) {
		return false, nil
	}
	return true, nil
}

func NewDurationFromHeartbeat(h *Heartbeat) *Duration {
	d := &Duration{
		UserID:          h.UserID,
		Time:            h.Time,
		Duration:        0,
		Project:         h.Project,
		Language:        h.Language,
		Editor:          h.Editor,
		OperatingSystem: h.OperatingSystem,
		Machine:         h.Machine,
		Category:        h.Category,
		Branch:          h.Branch,
		Entity:          h.Entity,
		NumHeartbeats:   1,
	}
	return d.Hashed()
}

func (d *Duration) WithEntityIgnored() *Duration {
	d.excludeEntity = true
	return d
}

func (d *Duration) Hashed() *Duration {
	hash, err := hashstructure.Hash(d, hashstructure.FormatV2, nil)
	if err != nil {
		logbuch.Error("CRITICAL ERROR: failed to hash struct - %v", err)
	}
	d.GroupHash = fmt.Sprintf("%x", hash)
	return d
}

func (d *Duration) GetKey(t uint8) (key string) {
	switch t {
	case SummaryProject:
		key = d.Project
	case SummaryEditor:
		key = d.Editor
	case SummaryLanguage:
		key = d.Language
	case SummaryOS:
		key = d.OperatingSystem
	case SummaryMachine:
		key = d.Machine
	case SummaryBranch:
		key = d.Branch
	case SummaryEntity:
		key = d.Entity
	case SummaryCategory:
		key = d.Category
	}

	if key == "" {
		key = UnknownSummaryKey
	}

	return key
}
