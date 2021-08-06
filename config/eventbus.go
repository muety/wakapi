package config

import "github.com/leandro-lugaresi/hub"

type ApplicationEvent struct {
	Type    string
	Payload interface{}
}

const (
	TopicUser               = "user.*"
	TopicHeartbeat          = "heartbeat.*"
	TopicProjectLabel       = "project_label.*"
	EventUserUpdate         = "user.update"
	EventHeartbeatCreate    = "heartbeat.create"
	EventProjectLabelCreate = "project_label.create"
	EventProjectLabelDelete = "project_label.delete"
	FieldPayload            = "payload"
	FieldUser               = "user"
	FieldUserId             = "user.id"
)

var eventHub *hub.Hub

func init() {
	eventHub = hub.New()
}

func EventBus() *hub.Hub {
	return eventHub
}
