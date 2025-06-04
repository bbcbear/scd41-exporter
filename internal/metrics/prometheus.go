package metrics

import (
	"github.com/bbcbear/scd41-exporter/internal/sensor"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	SensorMetrics *prometheus.GaugeVec
	readErrors    prometheus.Counter
)

func Register() {
	SensorMetrics = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "scd41_value",
			Help: "SCD41 sensor values with type and unit labels",
		},
		[]string{"type", "unit"},
	)
	readErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sensor_read_errors_total",
			Help: "Total number of failed sensor reads",
		},
	)

	prometheus.MustRegister(SensorMetrics)
	prometheus.MustRegister(readErrors)
}

func Unregister() {
	prometheus.Unregister(SensorMetrics)
	prometheus.Unregister(readErrors)
}

func IncReadError() {
	readErrors.Inc()
}

func Update(m sensor.Measurement) {
	SensorMetrics.WithLabelValues("co2", "ppm").Set(float64(m.CO2))
	SensorMetrics.WithLabelValues("temperature", "°C").Set(float64(m.Temperature))
	SensorMetrics.WithLabelValues("humidity", "%").Set(float64(m.Humidity))
}
