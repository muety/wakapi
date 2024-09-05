package config

import (
	"github.com/getsentry/sentry-go"
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
	if Get().IsDev() {
		level = slog.LevelDebug
	}
	handler := slogsentry.Option{Level: level}.NewSentryHandler()
	logger := slog.New(handler)

	sentryLogger = &SentryLogger{Logger: logger}

	return sentryLogger
}

func (l *SentryLogger) Fatal(msg string, args ...any) {
	l.Error(msg, args...)
	sentry.Flush(2 * time.Second)
	os.Exit(1)
}

func (l *SentryLogger) Request(r *http.Request) *slog.Logger {
	return l.Logger.With(slog.Any("http_request", r))
}

var excludedRoutes = []string{
	"GET /assets",
	"GET /api/health",
	"GET /swagger-ui",
	"GET /docs",
}

func initSentry(config sentryConfig, debug bool) {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              config.Dsn,
		Debug:            debug,
		Environment:      config.Environment,
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
			r, ok := event.Contexts["extra"]["http_request"]
			if ok {
				if uid := getPrincipal(r.(*http.Request)); uid != "" {
					event.User.ID = uid
				}
			}
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
