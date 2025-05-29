package config

import (
	"log/slog"
	"os"
)

func SetupLogger() {
	logFormat := GetEnv("LOG_FORMAT", "json")
	var handler slog.Handler

	if logFormat == "text" {
		handler = slog.NewTextHandler(os.Stdout, nil)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	}

	slog.SetDefault(slog.New(handler))
}
