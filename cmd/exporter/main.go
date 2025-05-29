package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bbcbear/scd41-exporter/internal/app"
	"bbcbear/scd41-exporter/internal/config"
	"log/slog"
)

func main() {
	config.SetupLogger()
	slog.Info("Application starting")
	
	addr := config.GetEnv("METRICS_ADDR", ":2113")
	pollInterval := config.GetEnvDuration("POLL_INTERVAL", 2*time.Second)

	appInstance, err := app.New(addr, pollInterval)
	if err != nil {
		slog.Error("Failed to initialize app", "error", err)
		os.Exit(1)
	}
	defer func() {
		appInstance.Shutdown()
		if err := appInstance.Bus.Close(); err != nil {
			slog.Warn("Failed to close I2C bus", "error", err)
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go appInstance.StartPolling(ctx, cancel)
	go func() {
		if err := appInstance.StartHTTPServer(ctx); err != nil {
			slog.Error("HTTP server exited with error", "error", err)
			cancel()
		}
	}()

	slog.Info("Application started")
	<-ctx.Done()
	slog.Info("Shutting down application")
}
