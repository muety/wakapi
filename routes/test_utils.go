package routes

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// https://github.com/go-chi/chi/issues/76#issuecomment-370145140
func WithUrlParam(r *http.Request, key, value string) *http.Request {
	r.URL.RawPath = strings.Replace(r.URL.RawPath, "{"+key+"}", value, 1)
	r.URL.Path = strings.Replace(r.URL.Path, "{"+key+"}", value, 1)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	return r
}
