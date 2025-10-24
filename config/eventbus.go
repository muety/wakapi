package config

import "github.com/leandro-lugaresi/hub"

type ApplicationEvent struct {
	Type    string
	Payload interface{}
}

const (
	TopicUser                    = "user.*"
	TopicHeartbeat               = "heartbeat.*"
	TopicProjectLabel            = "project_label.*"
	EventUserUpdate              = "user.update"
	EventUserDelete              = "user.delete"
	EventHeartbeatCreate         = "heartbeat.create"
	EventProjectLabelCreate      = "project_label.create"
	EventProjectLabelDelete      = "project_label.delete"
	EventWakatimeFailure         = "wakatime.failure"
	EventLanguageMappingsChanged = "language_mappings.changed"
	EventApiKeyCreate            = "api_key.create"
	EventApiKeyDelete            = "api_key.delete"
	FieldPayload                 = "payload"
	FieldUser                    = "user"
	FieldUserId                  = "user.id"
)

var eventHub *hub.Hub

func init() {
	eventHub = hub.New()
}

func EventBus() *hub.Hub {
	return eventHub
}
