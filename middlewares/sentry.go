package middlewares

import (
	"context"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"net/http"
)

type SentryMiddleware struct {
	handler http.Handler
}

func NewSentryMiddleware() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return sentryhttp.New(sentryhttp.Options{
			Repanic: true,
		}).Handle(&SentryMiddleware{handler: h})
	}
}

func (h *SentryMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), "-", "-")
	h.handler.ServeHTTP(w, r.WithContext(ctx))
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		if user := GetPrincipal(r); user != nil {
			hub.Scope().SetUser(sentry.User{ID: user.ID})
		}
	}
}
