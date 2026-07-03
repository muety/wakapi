package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/Marlliton/slogpretty"
)

func InitLogger(logFormat string) {
	var handler slog.Handler
	if logFormat == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Value.Kind() == slog.KindDuration {
					// log durations in milliseconds instead of nanoseconds
					// relates to https://go-review.googlesource.com/c/go/+/480735, which proposed to log durations as their string representation but apparently was rejected
					return slog.Int64(a.Key, int64(a.Value.Duration()/time.Millisecond))
				}
				return a
			}})
	} else {
		handler = slogpretty.New(os.Stdout, &slogpretty.Options{
			Level:    slog.LevelDebug,
			Colorful: true,
		})
	}
	slog.SetDefault(slog.New(handler))
}
