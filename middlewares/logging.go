package middlewares

import (
	"github.com/gorilla/handlers"
	"net/http"
	"os"
)

type LoggingMiddleware struct{}

func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{}
}

func (m *LoggingMiddleware) Handler(h http.Handler) http.Handler {
	return handlers.LoggingHandler(os.Stdout, h)
}
