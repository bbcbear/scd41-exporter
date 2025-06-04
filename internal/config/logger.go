package config

import (
	"log/slog"
	"os"
)

func SetupLogger() {
	logFormat := GetEnv("LOG_FORMAT", "json")
	logLevel := getLogLevel()

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	var handler slog.Handler
	if logFormat == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}

func getLogLevel() slog.Level {
	switch GetEnv("LOG_LEVEL", "info") {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}