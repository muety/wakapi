package routes

import (
	"github.com/go-chi/chi/v5"
)

type Handler interface {
	RegisterRoutes(router chi.Router)
}
