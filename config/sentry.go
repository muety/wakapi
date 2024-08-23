package config

import (
	"github.com/getsentry/sentry-go"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

// How to: Logging
// Use slog.[Debug|Info|Warn|Error|Fatal]() by default
// Use config.Log().[Debug|Info|Warn|Error|Fatal]() when wanting the log to appear in Sentry as well

var sentryWrapperLogger *SentryWrapperLogger

type capturingWriter struct {
	Writer  io.Writer
	Message string
}

func (c *capturingWriter) Clear() {
	c.Message = ""
}

func (c *capturingWriter) Write(p []byte) (n int, err error) {
	c.Message = string(p)
	return c.Writer.Write(p)
}

// SentryWrapperLogger is a wrapper around a slog.Logger that forwards events to Sentry in addition and optionally allows to attach a request context
type SentryWrapperLogger struct {
	Logger    *slog.Logger
	req       *http.Request
	outWriter *capturingWriter
	errWriter *capturingWriter
}

func Log() *SentryWrapperLogger {
	if sentryWrapperLogger != nil {
		return sentryWrapperLogger
	}

	ow, ew := &capturingWriter{Writer: os.Stdout}, &capturingWriter{Writer: os.Stderr}
	var handler slog.Handler
	if Get().IsDev() {
		handler = slog.NewTextHandler(io.MultiWriter(ow, ew), &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	} else {
		handler = slog.NewJSONHandler(io.MultiWriter(ow, ew), &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	// Create a custom handler that writes to both output and error writers
	sentryWrapperLogger = &SentryWrapperLogger{
		Logger:    slog.New(handler),
		outWriter: ow,
		errWriter: ew,
	}
	return sentryWrapperLogger
}

func (l *SentryWrapperLogger) Request(req *http.Request) *SentryWrapperLogger {
	l.req = req
	return l
}

func (l *SentryWrapperLogger) Debug(msg string, params ...interface{}) {
	l.outWriter.Clear()
	l.Logger.Debug(msg, params...)
	l.log(l.errWriter.Message, sentry.LevelDebug)
}

func (l *SentryWrapperLogger) Info(msg string, params ...interface{}) {
	l.outWriter.Clear()
	l.Logger.Info(msg, params...)
	l.log(l.errWriter.Message, sentry.LevelInfo)
}

func (l *SentryWrapperLogger) Warn(msg string, params ...interface{}) {
	l.outWriter.Clear()
	l.Logger.Warn(msg, params...)
	l.log(l.errWriter.Message, sentry.LevelWarning)
}

func (l *SentryWrapperLogger) Error(msg string, params ...interface{}) {
	l.errWriter.Clear()
	l.Logger.Error(msg, params...)
	l.log(l.errWriter.Message, sentry.LevelError)
}

func (l *SentryWrapperLogger) Fatal(msg string, params ...interface{}) {
	l.errWriter.Clear()
	l.Logger.Error(msg, params...)
	l.log(l.errWriter.Message, sentry.LevelFatal)
	os.Exit(1)
}

func (l *SentryWrapperLogger) log(msg string, level sentry.Level) {
	event := sentry.NewEvent()
	event.Level = level
	event.Message = msg

	if l.req != nil {
		if h := l.req.Context().Value(sentry.HubContextKey); h != nil {
			hub := h.(*sentry.Hub)
			hub.Scope().SetRequest(l.req)
			if uid := getPrincipal(l.req); uid != "" {
				hub.Scope().SetUser(sentry.User{ID: uid})
			}
			hub.CaptureEvent(event)
			return
		}
	}

	sentry.CaptureEvent(event)
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
			if hint.Context != nil {
				if req, ok := hint.Context.Value(sentry.RequestContextKey).(*http.Request); ok {
					if uid := getPrincipal(req); uid != "" {
						event.User.ID = uid
					}
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
