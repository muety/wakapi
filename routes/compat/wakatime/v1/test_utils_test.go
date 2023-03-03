package v1

import (
	"context"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strings"
)

func withUrlParam(r *http.Request, key, value string) *http.Request {
	r.URL.RawPath = strings.Replace(r.URL.RawPath, "{"+key+"}", value, 1)
	r.URL.Path = strings.Replace(r.URL.Path, "{"+key+"}", value, 1)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	return r
}
