package models

import (
	"fmt"
	"log/slog"
	"sort"
	"time"
	"unicode"

	"github.com/mitchellh/hashstructure/v2"
)

type ProcessHeartbeatsArgs struct {
	Heartbeats             []*Heartbeat
	Start                  time.Time
	End                    time.Time
	User                   *User
	LastHeartbeatYesterday *Heartbeat
	FirstHeartbeatTomorrow *Heartbeat
	SliceBy                string
}

type Duration struct {
	UserID          string        `json:"-"`
	Time            CustomTime    `json:"time" hash:"ignore"`
	Duration        time.Duration `json:"-" hash:"ignore"`
	Project         string        `json:"project,omitempty"`
	Language        string        `json:"language,omitempty"`
	Editor          string        `json:"editor,omitempty"`
	OperatingSystem string        `json:"operating_system,omitempty"`
	Machine         string        `json:"machine,omitempty"`
	Category        string        `json:"category,omitempty"`
	Branch          string        `json:"branch,omitempty"`
	Entity          string        `json:"entity,omitempty"`
	NumHeartbeats   int           `json:"-" hash:"ignore"`
	GroupHash       string        `json:"-" hash:"ignore"`
	excludeEntity   bool          `json:"-" hash:"ignore"`
	Color           *string       `json:"color" hash:"ignore"`
	DurationSecs    float64       `json:"duration" hash:"ignore"`
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

type MiniDurationHeartbeat struct {
	Heartbeat               // Embed original heartbeat fields
	Duration  time.Duration // Duration calculated in HeartbeatsToMiniDurations step (seconds)
}

type SummariesGrandTotal struct {
	Digital      string  `json:"digital"`
	Hours        int     `json:"hours"`
	Minutes      int     `json:"minutes"`
	Text         string  `json:"text"`
	TotalSeconds float64 `json:"total_seconds"`
}

type DurationResult struct {
	Data       Durations            `json:"data"`
	Start      time.Time            `json:"start"`
	End        time.Time            `json:"end"`
	Timezone   string               `json:"timezone"`
	GrandTotal *SummariesGrandTotal `json:"grand_total"`
}

func (d *DurationResult) TotalTime() time.Duration {
	var total time.Duration
	for _, item := range d.Data {
		total += item.Duration
	}
	return total
}

type Durations []*Duration

func (durations Durations) TotalTime() time.Duration {
	var total time.Duration
	for _, item := range durations {
		total += item.Duration
	}
	return total
}

func (durations Durations) GrandTotal() *SummariesGrandTotal {
	durationTotal := durations.TotalTime()
	d := durationTotal.Round(time.Minute)

	totalSeconds := durationTotal.Round(time.Second).Seconds()
	hours := int(d / time.Hour)
	minutes := int((d % time.Hour) / time.Minute)

	digital := fmt.Sprintf("%02d:%02d", hours, minutes)
	text := fmt.Sprintf("%d hrs %d mins", hours, minutes)

	return &SummariesGrandTotal{
		Digital:      digital,
		Hours:        hours,
		Minutes:      minutes,
		Text:         text,
		TotalSeconds: totalSeconds,
	}
}

func (d Durations) Len() int {
	return len(d)
}

func (d Durations) Less(i, j int) bool {
	return d[i].Time.T().Before(d[j].Time.T())
}

func (d Durations) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d Durations) TotalNumHeartbeats() int {
	var total int
	for _, e := range d {
		total += e.NumHeartbeats
	}
	return total
}

func (d Durations) Sorted() Durations {
	sort.Sort(d)
	return d
}

func (d *Durations) First() *Duration {
	// assumes slice to be sorted
	if d.Len() == 0 {
		return nil
	}
	return (*d)[0]
}

func (d *Durations) Last() *Duration {
	// assumes slice to be sorted
	if d.Len() == 0 {
		return nil
	}
	return (*d)[d.Len()-1]
}

type MakeDurationOptions struct {
	ExcludeUnknownProjects bool
	Filters                *Filters
	HeartBeats             []*Heartbeat
	HeartbeatsTimeout      time.Duration
	From                   time.Time
	To                     time.Time
	Timezone               *time.Location
	LastHeartbeatYesterday *Heartbeat
	FirstHeartbeatTomorrow *Heartbeat
}
