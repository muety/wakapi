package config

import (
	"github.com/emvi/logbuch"
	"github.com/getsentry/sentry-go"
	"github.com/muety/wakapi/models"
	"net/http"
	"os"
	"strings"
)

type SentryErrorWriter struct{}

// TODO: extend sentry error logging to include context and stacktrace
// see https://github.com/muety/wakapi/issues/169
func (s *SentryErrorWriter) Write(p []byte) (n int, err error) {
	sentry.CaptureMessage(string(p))
	return os.Stderr.Write(p)
}

func init() {
	logbuch.SetOutput(os.Stdout, &SentryErrorWriter{})
}

func initSentry(config sentryConfig, debug bool) {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:   config.Dsn,
		Debug: debug,
		TracesSampler: sentry.TracesSamplerFunc(func(ctx sentry.SamplingContext) sentry.Sampled {
			if !config.EnableTracing {
				return sentry.SampledFalse
			}

			hub := sentry.GetHubFromContext(ctx.Span.Context())
			txName := hub.Scope().Transaction()

			if strings.HasPrefix(txName, "GET /assets") || strings.HasPrefix(txName, "GET /api/health") {
				return sentry.SampledFalse
			}
			if txName == "POST /api/heartbeat" {
				return sentry.UniformTracesSampler(config.SampleRateHeartbeats).Sample(ctx)
			}
			return sentry.UniformTracesSampler(config.SampleRate).Sample(ctx)
		}),
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			type principalGetter interface {
				GetPrincipal() *models.User
			}
			if hint.Context != nil {
				if req, ok := hint.Context.Value(sentry.RequestContextKey).(*http.Request); ok {
					if p := req.Context().Value("principal"); p != nil {
						event.User.ID = p.(principalGetter).GetPrincipal().ID
					}
				}
			}
			return event
		},
	}); err != nil {
		logbuch.Fatal("failed to initialized sentry – %v", err)
	}
}
