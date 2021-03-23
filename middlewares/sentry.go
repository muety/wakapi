package middlewares

import (
	sentryhttp "github.com/getsentry/sentry-go/http"
	"net/http"
)

func NewSentryMiddleware() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return sentryhttp.New(sentryhttp.Options{
			Repanic: true,
		}).Handle(h)
	}
}
