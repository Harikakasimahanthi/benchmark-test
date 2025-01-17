package execution

import (
	"context"
	"log/slog"
	"net"
	"time"

	"github.com/Harikakasimahanthi/benchmark-test/internal/platform/logger"
	"github.com/Harikakasimahanthi/benchmark-test/internal/platform/metric"
)

const (
	DurationMinMeasurement = "DurationMin"
	DurationP10Measurement = "DurationP10"
	DurationP50Measurement = "DurationP50"
	DurationP90Measurement = "DurationP90"
	DurationMaxMeasurement = "DurationMax"
)

type LatencyMetric struct {
	metric.Base[time.Duration]
	host              string
	interval, timeout time.Duration
	durations         []time.Duration
}

func NewLatencyMetric(host, name string, interval time.Duration, healthCondition []metric.HealthCondition[time.Duration]) *LatencyMetric {
	return &LatencyMetric{
		host: host,
		Base: metric.Base[time.Duration]{
			HealthConditions: healthCondition,
			Name:             name,
		},
		interval: interval,
		timeout:  time.Duration(float64(interval) * 0.75),
	}
}

func (l *LatencyMetric) Measure(ctx context.Context) {
	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.With("metric_name", l.Name).Debug("metric was stopped")
			return
		case <-ticker.C:
			l.measure()
		}
	}
}

func (l *LatencyMetric) measure() {
	var latency time.Duration
	start := time.Now()

	// Measure the latency between the execution layer and the host
	conn, err := net.DialTimeout("tcp", l.host, l.timeout)
	if err != nil {
		// Log error if the connection fails
		logger.WriteError(metric.ExecutionGroup, l.Name, err)
		return
	}
	defer conn.Close()

	latency = time.Since(start)

	// Store the latency measurements
	l.durations = append(l.durations, latency)

	// Report the latency metric
	l.writeMetric(latency)
}

func (l *LatencyMetric) writeMetric(latency time.Duration) {
	// Calculate percentiles for latency (e.g., min, p10, p50, p90, max)
	percentiles := metric.CalculatePercentiles(l.durations, 0, 10, 50, 90, 100)

	// Record latency percentiles as data points
	l.AddDataPoint(map[string]time.Duration{
		DurationMinMeasurement: percentiles[0],
		DurationP10Measurement: percentiles[10],
		DurationP50Measurement: percentiles[50],
		DurationP90Measurement: percentiles[90],
		DurationMaxMeasurement: percentiles[100],
	})

	// Assuming Prometheus metric tracking is used for latency
	latencyMetric.Observe(latency.Seconds())

	// Log the latency metric
	logger.WriteMetric(metric.ExecutionGroup, l.Name, map[string]any{
		DurationMinMeasurement: percentiles[0],
		DurationP10Measurement: percentiles[10],
		DurationP50Measurement: percentiles[50],
		DurationP90Measurement: percentiles[90],
		DurationMaxMeasurement: percentiles[100],
	})
}

func (l *LatencyMetric) AggregateResults() string {
	// Retrieve the last recorded latency values
	var min, p10, p50, p90, max time.Duration

	min = l.DataPoints[len(l.DataPoints)-1].Values[DurationMinMeasurement]
	p10 = l.DataPoints[len(l.DataPoints)-1].Values[DurationP10Measurement]
	p50 = l.DataPoints[len(l.DataPoints)-1].Values[DurationP50Measurement]
	p90 = l.DataPoints[len(l.DataPoints)-1].Values[DurationP90Measurement]
	max = l.DataPoints[len(l.DataPoints)-1].Values[DurationMaxMeasurement]

	// Return formatted latency percentiles
	return metric.FormatPercentiles(min, p10, p50, p90, max)
}
