package models

import (
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode"

	"github.com/cespare/xxhash/v2"
	"github.com/gohugoio/hashstructure"
	"github.com/muety/wakapi/models/lib"
)

// TODO: support multiple durations per time per user for different heartbeat timeouts
// see discussion at https://github.com/muety/wakapi/issues/675
type Duration struct {
	ID     int64  `json:"-" gorm:"primaryKey; autoIncrement"` // https://github.com/muety/wakapi/issues/777
	UserID string `json:"user_id" gorm:"not null; index:idx_time_duration_user"`
	// note: on sqlite, table will have an additional column `time_real`, introduced "manually" by migration 20260111
	// see https://github.com/muety/wakapi/issues/882 for details
	Time            CustomTime    `json:"time" hash:"ignore" gorm:"not null; index:idx_time_duration; index:idx_time_duration_user"` // time of first heartbeat of this duration
	Duration        time.Duration `json:"duration" hash:"ignore" gorm:"not null"`
	Project         string        `json:"project"`
	Language        string        `json:"language"`
	Editor          string        `json:"editor"`
	OperatingSystem string        `json:"operating_system"`
	Machine         string        `json:"machine"`
	Category        string        `json:"category"`
	Branch          string        `json:"branch"`
	Entity          string        `json:"Entity"`
	NumHeartbeats   int           `json:"-" hash:"ignore"`
	GroupHash       string        `json:"-" hash:"ignore" gorm:"type:varchar(17)"`
	Timeout         time.Duration `json:"-" gorm:"not null; default:600000000000"` // heartbeat timeout preference, see DefaultHeartbeatsTimeout
	excludeEntity   bool          `json:"-" hash:"ignore"`
}

func (d *Duration) TimeEnd() time.Time {
	return d.Time.T().Add(d.Duration)
}

func (d *Duration) HashInclude(field string, v interface{}) (bool, error) {
	if field == "Entity" {
		return !d.excludeEntity, nil
	}
	if field == "Time" ||
		field == "Duration" ||
		field == "NumHeartbeats" ||
		field == "GroupHash" ||
		field == "ID" ||
		field == "Timeout" ||
		unicode.IsLower(rune(field[0])) {
		return false, nil
	}
	return true, nil
}

func NewDurationFromHeartbeat(h *Heartbeat) *Duration {
	var interval = DefaultHeartbeatsTimeout
	if h.User != nil && h.User.HeartbeatsTimeout() > 0 {
		interval = h.User.HeartbeatsTimeout()
	}

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
		Timeout:         interval,
	}
	return d
}

func (d *Duration) WithEntityIgnored() *Duration {
	d.excludeEntity = true
	return d
}

func (d *Duration) WithTimeout(interval time.Duration) *Duration {
	d.Timeout = interval
	return d
}

func (d *Duration) Hashed() *Duration {
	hash, err := hashstructure.Hash(d, &hashstructure.HashOptions{Hasher: xxhash.New()})
	if err != nil {
		slog.Error("CRITICAL ERROR: failed to hash struct", "error", err)
	}
	d.GroupHash = fmt.Sprintf("%x", hash)
	return d
}

func (d *Duration) Augmented(languageMappings map[string]string) *Duration {
	for ext, targetLang := range languageMappings {
		langs, ok := lib.LanguagesByExtension["."+ext]
		if !ok {
			continue
		}
		if lang := langs[0]; strings.ToLower(d.Language) == strings.ToLower(lang) {
			d.Language = targetLang
		}
	}
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
