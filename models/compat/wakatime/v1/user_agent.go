package v1

import (
	"time"

	"github.com/muety/wakapi/models"
)

type UserAgentsViewModel struct {
	Data       []*UserAgentEntry `json:"data"`
	TotalPages int               `json:"total_pages"`
}

type UserAgentEntry struct {
	Id                 string `json:"id"`
	Editor             string `json:"editor"`
	Os                 string `json:"os"`
	Value              string `json:"value"`
	Version            string `json:"version"`              // currently not implemented
	IsBrowserExtension bool   `json:"is_browser_extension"` // currently not implemented
	IsDesktopApp       bool   `json:"is_desktop_app"`       // currently not implemented
	FirstSeen          string `json:"first_seen"`
	LastSeen           string `json:"last_seen"`
}

func (e *UserAgentEntry) FromModel(userAgent *models.UserAgent) *UserAgentEntry {
	e.Id = userAgent.Id
	e.Editor = userAgent.Editor
	e.Os = userAgent.Os
	e.Value = userAgent.Value
	e.FirstSeen = userAgent.FirstSeen.Format(time.RFC3339)
	e.LastSeen = userAgent.LastSeen.Format(time.RFC3339)
	return e
}
