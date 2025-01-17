package infrastructure

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Harikakasimahanthi/benchmark-test/internal/platform/logger"
	"github.com/Harikakasimahanthi/benchmark-test/internal/platform/metric"
	"github.com/mackerelio/go-osstat/memory"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	UsedMemoryMeasurement   = "Used"
	TotalMemoryMeasurement  = "Total"
	CachedMemoryMeasurement = "Cached"
	FreeMemoryMeasurement   = "Free"
)

type MemoryMetric struct {
	metric.Base[uint64]
	interval time.Duration
}

func NewMemoryMetric(name string, interval time.Duration, healthCondition []metric.HealthCondition[uint64]) *MemoryMetric {
	return &MemoryMetric{
		Base: metric.Base[uint64]{
			HealthConditions: healthCondition,
			Name:             name,
		},
		interval: interval,
	}
}

func (m *MemoryMetric) Measure(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.With("metric_name", m.Name).Debug("memory metric was stopped")
			return
		case <-ticker.C:
			m.measure()
		}
	}
}

func (m *MemoryMetric) measure() {
	// Get the memory stats from the system
	memoryStats, err := memory.Get()
	if err != nil {
		logger.WriteError(metric.InfrastructureGroup, m.Name, err)
		return
	}

	// Log and record the memory metrics
	m.writeMetric(memoryStats.Cached, memoryStats.Used, memoryStats.Free, memoryStats.Total)
}

func (m *MemoryMetric) writeMetric(cached, used, free, total uint64) {
	// Record the data points in memory metrics
	m.AddDataPoint(map[string]uint64{
		CachedMemoryMeasurement: cached,
		UsedMemoryMeasurement:   used,
		FreeMemoryMeasurement:   free,
		TotalMemoryMeasurement:  total,
	})

	// Push memory metrics to Prometheus
	memoryUsageMetric.With(prometheus.Labels{memoryUsageTypeLabel: "cached"}).Set(float64(cached))
	memoryUsageMetric.With(prometheus.Labels{memoryUsageTypeLabel: "used"}).Set(float64(used))
	memoryUsageMetric.With(prometheus.Labels{memoryUsageTypeLabel: "free"}).Set(float64(free))
	memoryUsageMetric.With(prometheus.Labels{memoryUsageTypeLabel: "total"}).Set(float64(total))

	// Log the memory usage data
	logger.WriteMetric(metric.InfrastructureGroup, m.Name, map[string]any{
		TotalMemoryMeasurement:  toMegabytes(total),
		UsedMemoryMeasurement:   toMegabytes(used),
		CachedMemoryMeasurement: toMegabytes(cached),
		FreeMemoryMeasurement:   toMegabytes(free),
	})
}

func (m *MemoryMetric) AggregateResults() string {
	// Prepare to calculate and display the percentiles
	var values map[string][]float64 = make(map[string][]float64)

	for _, point := range m.DataPoints {
		values[TotalMemoryMeasurement] = append(values[TotalMemoryMeasurement], toMegabytes(point.Values[TotalMemoryMeasurement]))
		values[FreeMemoryMeasurement] = append(values[FreeMemoryMeasurement], toMegabytes(point.Values[FreeMemoryMeasurement]))
		values[UsedMemoryMeasurement] = append(values[UsedMemoryMeasurement], toMegabytes(point.Values[UsedMemoryMeasurement]))
		values[CachedMemoryMeasurement] = append(values[CachedMemoryMeasurement], toMegabytes(point.Values[CachedMemoryMeasurement]))
	}

	// Return a formatted string with the 50th percentile (P50) for each memory category
	return fmt.Sprintf("total_P50=%.2fMB, used_P50=%.2fMB, cached_P50=%.2fMB, free_P50=%.2fMB",
		metric.CalculatePercentiles(values[TotalMemoryMeasurement], 50)[50],
		metric.CalculatePercentiles(values[UsedMemoryMeasurement], 50)[50],
		metric.CalculatePercentiles(values[CachedMemoryMeasurement], 50)[50],
		metric.CalculatePercentiles(values[FreeMemoryMeasurement], 50)[50])
}

func toMegabytes(bytes uint64) float64 {
	// Convert bytes to megabytes
	return float64(bytes) / (1024 * 1024)
}
