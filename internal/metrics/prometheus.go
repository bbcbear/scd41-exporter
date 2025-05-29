
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"bbcbear/sps30-exporter/internal/sensor"
)

var (
	SensorMetrics *prometheus.GaugeVec
	readErrors    prometheus.Counter
)

func Register() {
	SensorMetrics = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sps30_value",
			Help: "SPS30 sensor values with type and unit labels",
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
	SensorMetrics.WithLabelValues("co2", "ppm").Set(m.CO2)
	SensorMetrics.WithLabelValues("temperature", "°C").Set(m.Temperature)
	SensorMetrics.WithLabelValues("humidity", "%").Set(m.Humidity)
}
