package middlewares

import (
	"net/http"
)

type RequireBasicAuthMiddleware struct{}

func (m *RequireBasicAuthMiddleware) Init() {}

func (m *RequireBasicAuthMiddleware) Handle(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
	next(w, r)
}
