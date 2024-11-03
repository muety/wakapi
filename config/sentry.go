package config

import (
	"github.com/getsentry/sentry-go"
	slogmulti "github.com/samber/slog-multi"
	slogsentry "github.com/samber/slog-sentry/v2"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

// How to: Logging
// Use slog.[Debug|Info|Warn|Error|Fatal]() by default
// Use config.Log().[Debug|Info|Warn|Error|Fatal]() when wanting the log to appear in Sentry as well

// SentryLogger wraps slog.Logger and provides a Fatal method
type SentryLogger struct {
	*slog.Logger
}

var sentryLogger *SentryLogger

func Log() *SentryLogger {
	if sentryLogger != nil {
		return sentryLogger
	}

	level := slog.LevelInfo
	if IsDev(env) {
		level = slog.LevelDebug
	}

	filterRequestInfo := slogmulti.NewWithAttrsInlineMiddleware(func(attrs []slog.Attr, next func([]slog.Attr) slog.Handler) slog.Handler {
		attrsNew := []slog.Attr{}
		for _, attr := range attrs {
			if attr.Key != "request" {
				attrsNew = append(attrsNew, attr)
			}
		}
		return next(attrsNew)
	})

	sentryLogger = &SentryLogger{Logger: slog.New(
		slogmulti.Fanout(
			slogmulti.Pipe(filterRequestInfo).Handler(slog.Default().Handler()),
			slogsentry.Option{Level: level}.NewSentryHandler(),
		),
	)}

	return sentryLogger
}

func (l *SentryLogger) Fatal(msg string, args ...any) {
	l.Error(msg, args...)
	sentry.Flush(2 * time.Second)
	os.Exit(1)
}

func (l *SentryLogger) Request(r *http.Request) *slog.Logger {
	l.Logger = l.Logger.With("request", r)
	if uid := getPrincipal(r); uid != "" {
		l.Logger = l.Logger.With(slog.Group("user", slog.String("id", uid)))
	}
	return l.Logger
}

var excludedRoutes = []string{
	"GET /assets",
	"GET /api/health",
	"GET /swagger-ui",
	"GET /docs",
}

func initSentry(config sentryConfig, debug bool, releaseVersion string) {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              config.Dsn,
		Debug:            debug,
		Environment:      config.Environment,
		Release:          releaseVersion,
		AttachStacktrace: true,
		EnableTracing:    config.EnableTracing,
		TracesSampler: func(ctx sentry.SamplingContext) float64 {
			txName := ctx.Span.Name
			for _, ex := range excludedRoutes {
				if strings.HasPrefix(txName, ex) {
					return 0.0
				}
			}
			if txName == "POST /api/heartbeat" {
				return float64(config.SampleRateHeartbeats)
			}
			return float64(config.SampleRate)
		},
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			// optional pre-processing before sending the event off
			return event
		},
	}); err != nil {
		Log().Fatal("failed to initialized sentry", "error", err)
	}
}

// returns a user id
func getPrincipal(r *http.Request) string {
	type principalIdentityGetter interface {
		GetPrincipalIdentity() string
	}

	if p := r.Context().Value("principal"); p != nil {
		return p.(principalIdentityGetter).GetPrincipalIdentity()
	}
	return ""
}
