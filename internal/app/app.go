package app

import (
	"context"
	"github.com/bbcbear/scd41-exporter/internal/handlers"
	"github.com/bbcbear/scd41-exporter/internal/metrics"
	"github.com/bbcbear/scd41-exporter/internal/sensor"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"fmt"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

type App struct {
	Sensor          sensor.Sensor
	Bus             interface{ Close() error }
	Addr            string
	Interval        time.Duration
	isSensorHealthy atomic.Bool
}

// I2CBusAdapter адаптирует *i2c.Dev к интерфейсу sensor.Bus
type I2CBusAdapter struct {
	Dev *i2c.Dev
}

func (b *I2CBusAdapter) Tx(w, r []byte) error {
	return b.Dev.Tx(w, r)
}

func initHardware() error {
	if _, err := host.Init(); err != nil {
		return fmt.Errorf("failed to initialize periph host: %w", err)
	}
	return nil
}

func openI2CBus() i2c.BusCloser {
	bus, err := i2creg.Open("")
	if err != nil {
		slog.Error("Failed to open I2C bus", "error", err)
	}
	return bus
}

func New(addr string, interval time.Duration) (*App, error) {
	if err := initHardware(); err != nil {
		return nil, fmt.Errorf("sensor init failed: %w", err)
	}
	bus := openI2CBus()

	dev := &i2c.Dev{Addr: 0x62, Bus: bus}
	adapter := &I2CBusAdapter{Dev: dev}
	s := sensor.New(adapter)
	if err := s.Init(); err != nil {
		_ = bus.Close()
		return nil, fmt.Errorf("sensor init failed: %w", err)
	}
	metrics.Register()

	return &App{
		Sensor:   s,
		Bus:      bus,
		Addr:     addr,
		Interval: interval,
	}, nil
}

func (a *App) StartPolling(ctx context.Context, cancel context.CancelFunc) {
	slog.Info("Sensor polling started")
	defer slog.Info("Sensor polling stopped")

	ticker := time.NewTicker(a.Interval)
	defer ticker.Stop()

	consecutiveFails := 0
	const maxFails = 5

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ok := a.readAndUpdate()
			if !ok {
				consecutiveFails++
				slog.Warn("Sensor read failed", "consecutiveFails", consecutiveFails)
				if consecutiveFails >= maxFails {
					if a.recoverSensor() {
						consecutiveFails = 0
					} else {
						slog.Warn("Sensor recovery failed, will retry later")
					}
				}
			} else {
				consecutiveFails = 0
			}
		}
	}
}

func (a *App) readAndUpdate() bool {
	measuring, err := a.Sensor.IsMeasuring()
	if err != nil {
		metrics.IncReadError()
		a.isSensorHealthy.Store(false)
		slog.Error("Sensor status check failed", "error", err)
		return false
	}
	if !measuring {
		a.isSensorHealthy.Store(false)
		slog.Warn("Sensor is not measuring, skipping update")
		return false
	}

	data, err := a.Sensor.Read()
	if err != nil {
		metrics.IncReadError()
		a.isSensorHealthy.Store(false)
		slog.Error("Failed to read sensor data", "error", err)
		return false
	}

	a.isSensorHealthy.Store(true)
	metrics.Update(data)
	slog.Info("Sensor data updated")
	return true
}

func (a *App) recoverSensor() bool {
	slog.Warn("Attempting to recover sensor")
	_ = a.Sensor.Stop()
	time.Sleep(300 * time.Millisecond)
	if err := a.Sensor.Init(); err != nil {
		slog.Error("Sensor re-init failed", "error", err)
		return false
	}
	// Проверим, действительно ли сенсор активен после повторной инициализации
	measuring, err := a.Sensor.IsMeasuring()
	if err != nil || !measuring {
		slog.Error("Sensor still not measuring after re-init", "error", err)
		return false
	}

	slog.Info("Sensor re-initialized successfully")
	return true
}

func (a *App) StartHTTPServer(ctx context.Context) error {
	router := handlers.Init(a.Sensor, &a.isSensorHealthy)

	srv := &http.Server{
		Addr:    a.Addr,
		Handler: router,
	}

	slog.Info("Exporter listening", "addr", a.Addr)

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		slog.Info("Shutting down HTTP server")
		return srv.Shutdown(context.Background())
	case err := <-errCh:
		return err
	}
}

func (a *App) Shutdown() {
	if err := a.Sensor.Stop(); err != nil {
		slog.Error("Sensor stop failed", "error", err)
	} else {
		slog.Info("Sensor stopped successfully")
	}
}
