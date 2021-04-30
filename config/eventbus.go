package config

import "github.com/leandro-lugaresi/hub"

type ApplicationEvent struct {
	Type    string
	Payload interface{}
}

const (
	TopicUser       = "user.*"
	EventUserUpdate = "user.update"
	FieldPayload    = "payload"
)

var eventHub *hub.Hub

func init() {
	eventHub = hub.New()
}

func EventBus() *hub.Hub {
	return eventHub
}
