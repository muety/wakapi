package models

import (
	"fmt"
	"log/slog"
	"math"
	"time"
	"unicode"

	"github.com/mitchellh/hashstructure/v2"
)

type DurationBlock struct {
	UserID       string        `json:"-"`
	Time         float64       `json:"time"`
	Project      string        `json:"project,omitempty"`
	Language     string        `json:"language,omitempty"`
	Entity       string        `json:"entity,omitempty"`
	Os           string        `json:"os,omitempty"`
	Editor       string        `json:"editor,omitempty"`
	Category     string        `json:"category,omitempty"`
	Machine      string        `json:"machine,omitempty"`
	DurationSecs float64       `json:"duration"`
	Duration     time.Duration `json:"-"`
	Color        *string       `json:"color"`
	Branch       string        `json:"branch,omitempty"`
}

type ProcessHeartbeatsArgs struct {
	Heartbeats             []*Heartbeat
	Start                  time.Time
	End                    time.Time
	User                   *User
	LastHeartbeatYesterday *Heartbeat
	FirstHeartbeatTomorrow *Heartbeat
	SliceBy                string
}

func (d *DurationBlock) AddDurationSecs() {
	d.DurationSecs = d.Duration.Seconds()
}

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

func round(number float64, precision int) float64 {
	factor := math.Pow(10, float64(precision))
	return math.Round(number*factor) / factor
}

func customTimeFromFloat(timeFloat float64) time.Time {
	return time.Unix(0, int64(round(timeFloat, 9)*1e9))
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

func NewDurationFromBlock(b *DurationBlock) *Duration {
	blockTime := customTimeFromFloat(b.Time)
	d := &Duration{
		UserID:          b.UserID,
		Time:            CustomTime(blockTime),
		Duration:        b.Duration,
		Project:         b.Project,
		Language:        b.Language,
		Editor:          b.Editor,
		OperatingSystem: b.Os,
		Machine:         b.Machine,
		Category:        b.Category,
		Branch:          b.Branch,
		Entity:          b.Entity,
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
		slog.Error("CRITICAL ERROR: failed to hash struct", "error", err)
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
