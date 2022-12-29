package config

import (
	"github.com/emvi/logbuch"
	"github.com/getsentry/sentry-go"
	"io"
	"net/http"
	"os"
	"strings"
)

// How to: Logging
// Use logbuch.[Debug|Info|Warn|Error|Fatal]() by default
// Use config.Log().[Debug|Info|Warn|Error|Fatal]() when wanting the log to appear in Sentry as well

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

// SentryWrapperLogger is a wrapper around a logbuch.Logger that forwards events to Sentry in addition and optionally allows to attach a request context
type SentryWrapperLogger struct {
	*logbuch.Logger
	req       *http.Request
	outWriter *capturingWriter
	errWriter *capturingWriter
}

func Log() *SentryWrapperLogger {
	ow, ew := &capturingWriter{Writer: os.Stdout}, &capturingWriter{Writer: os.Stderr}
	return &SentryWrapperLogger{
		Logger:    logbuch.NewLogger(ow, ew),
		outWriter: ow,
		errWriter: ew,
	}
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
	l.Logger.Fatal(msg, params...)
	l.log(l.errWriter.Message, sentry.LevelFatal)
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
		TracesSampler: sentry.TracesSamplerFunc(func(ctx sentry.SamplingContext) sentry.Sampled {
			if !config.EnableTracing {
				return sentry.SampledFalse
			}

			hub := sentry.GetHubFromContext(ctx.Span.Context())
			txName := hub.Scope().Transaction()

			for _, ex := range excludedRoutes {
				if strings.HasPrefix(txName, ex) {
					return sentry.SampledFalse
				}
			}
			if txName == "POST /api/heartbeat" {
				return sentry.UniformTracesSampler(config.SampleRateHeartbeats).Sample(ctx)
			}
			return sentry.UniformTracesSampler(config.SampleRate).Sample(ctx)
		}),
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
		logbuch.Fatal("failed to initialized sentry - %v", err)
	}
}

// returns a user id
func getPrincipal(r *http.Request) string {
	type identifiable interface {
		Identity() string
	}
	type principalGetter interface {
		GetPrincipal() *identifiable
	}

	if p := r.Context().Value("principal"); p != nil {
		return (*p.(principalGetter).GetPrincipal()).Identity()
	}
	return ""
}
