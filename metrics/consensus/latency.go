package consensus

import (
	"context"
	"log/slog"
	"net/http"
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
	url               string
	interval, timeout time.Duration
	durations         []time.Duration
}

func NewLatencyMetric(url, name string, interval time.Duration, healthCondition []metric.HealthCondition[time.Duration]) *LatencyMetric {
	return &LatencyMetric{
		url: url,
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

	// Measure latency for the solo staking node’s key endpoint
	client := http.Client{
		Timeout: l.timeout,
	}
	res, err := client.Get(l.url) // Replace with specific solo staking node endpoint if required
	if err != nil {
		logger.WriteError(metric.ConsensusGroup, l.Name, err)
		return
	}
	defer res.Body.Close()

	latency = time.Since(start)

	l.durations = append(l.durations, latency)

	l.writeMetric(latency)
}

func (l *LatencyMetric) writeMetric(latency time.Duration) {
	// Calculate percentiles for latency
	percentiles := metric.CalculatePercentiles(l.durations, 0, 10, 50, 90, 100)

	// Record latency metrics for reporting
	l.AddDataPoint(map[string]time.Duration{
		DurationMinMeasurement: percentiles[0],
		DurationP10Measurement: percentiles[10],
		DurationP50Measurement: percentiles[50],
		DurationP90Measurement: percentiles[90],
		DurationMaxMeasurement: percentiles[100],
	})

	// Assuming there is a Prometheus metric being used here for latency
	latencyMetric.Observe(latency.Seconds())

	// Log the measured metrics
	logger.WriteMetric(metric.ConsensusGroup, l.Name, map[string]any{
		DurationMinMeasurement: percentiles[0],
		DurationP10Measurement: percentiles[10],
		DurationP50Measurement: percentiles[50],
		DurationP90Measurement: percentiles[90],
		DurationMaxMeasurement: percentiles[100],
	})
}

func (l *LatencyMetric) AggregateResults() string {
	// Extract and return the percentiles for latency measurements
	var min, p10, p50, p90, max time.Duration

	if len(l.DataPoints) > 0 {
		min = l.DataPoints[len(l.DataPoints)-1].Values[DurationMinMeasurement]
		p10 = l.DataPoints[len(l.DataPoints)-1].Values[DurationP10Measurement]
		p50 = l.DataPoints[len(l.DataPoints)-1].Values[DurationP50Measurement]
		p90 = l.DataPoints[len(l.DataPoints)-1].Values[DurationP90Measurement]
		max = l.DataPoints[len(l.DataPoints)-1].Values[DurationMaxMeasurement]
	}

	return metric.FormatPercentiles(min, p10, p50, p90, max)
}
