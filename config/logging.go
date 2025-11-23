package config

import (
	"log/slog"
	"os"
	"time"
)

func InitLogger(isDev bool) {
	var handler slog.Handler
	if isDev {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	} else {
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
	}
	slog.SetDefault(slog.New(handler))
}
