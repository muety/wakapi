package view

import (
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
)

type BasicViewModel interface {
	SetError(string)
	SetSuccess(string)
}

type Messages struct {
	Success string
	Error   string
}

type SharedViewModel struct {
	Messages
	LeaderboardEnabled bool
	InvitesEnabled     bool
}

type SharedLoggedInViewModel struct {
	SharedViewModel
	User   *models.User
	ApiKey string
}

func NewSharedViewModel(c *conf.Config, messages *Messages) SharedViewModel {
	vm := SharedViewModel{
		LeaderboardEnabled: c.App.LeaderboardEnabled,
		InvitesEnabled:     c.Security.InviteCodes,
	}
	if messages != nil {
		vm.Messages = *messages
	}
	return vm
}

func (m *Messages) SetError(message string) {
	m.Error = message
}

func (m *Messages) SetSuccess(message string) {
	m.Success = message
}
