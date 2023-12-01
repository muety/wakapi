package routes

import (
	"github.com/go-chi/chi/v5"
)

type Handler interface {
	RegisterRoutes(router chi.Router)
}

type NoopHandler struct {
}

func (n *NoopHandler) RegisterRoutes(r chi.Router) {
}

func NewNoopHandler() *NoopHandler {
	return &NoopHandler{}
}
