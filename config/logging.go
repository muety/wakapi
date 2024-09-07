package config

import (
	"log/slog"
	"os"
)

func InitLogger(isDev bool) {
	var handler slog.Handler
	if isDev {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	slog.SetDefault(slog.New(handler))
}
