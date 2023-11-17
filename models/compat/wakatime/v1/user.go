package v1

import (
	"fmt"
	"time"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
)

const DefaultWakaUserDisplayName = "Anonymous User"

// partially compatible with https://wakatime.com/developers#users

type UserViewModel struct {
	Data *User `json:"data"`
}

type User struct {
	ID               string            `json:"id"`
	DisplayName      string            `json:"display_name"`
	FullName         string            `json:"full_name"`
	Email            string            `json:"email"`
	IsEmailPublic    bool              `json:"is_email_public"`
	IsEmailConfirmed bool              `json:"is_email_confirmed"`
	TimeZone         string            `json:"timezone"`
	LastHeartbeatAt  models.CustomTime `json:"last_heartbeat_at"`
	LastProject      string            `json:"last_project"`
	LastPluginName   string            `json:"last_plugin_name"`
	Username         string            `json:"username"`
	Website          string            `json:"website"`
	CreatedAt        models.CustomTime `json:"created_at"`
	ModifiedAt       models.CustomTime `json:"modified_at"`
	Photo            string            `json:"photo"`
}

func NewFromUser(user *models.User) *User {
	tz, _ := time.Now().Zone()
	if user.Location != "" {
		tz = user.Location
	}

	config := config.Get()

	return &User{
		ID:          user.ID,
		DisplayName: user.ID,
		Email:       user.Email,
		TimeZone:    tz,
		Username:    user.ID,
		CreatedAt:   user.CreatedAt,
		ModifiedAt:  user.CreatedAt,
		Photo:       fmt.Sprintf("%s%s/%s", config.Server.GetPublicUrl(), config.Server.BasePath, user.AvatarURL(config.App.AvatarURLTemplate)),
	}
}

func (u *User) WithLatestHeartbeat(h *models.Heartbeat) *User {
	u.LastHeartbeatAt = h.Time
	u.LastProject = h.Project
	u.LastPluginName = h.Editor
	return u
}
