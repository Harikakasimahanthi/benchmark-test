package infrastructure

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/mackerelio/go-osstat/cpu"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/ssvlabs/ssv-pulse/internal/platform/logger"
	"github.com/ssvlabs/ssv-pulse/internal/platform/metric"
)

const (
	SystemCPUMeasurement = "System"
	UserCPUMeasurement   = "User"
)

type CPUMetric struct {
	metric.Base[float64]
	prevUser, prevSystem, total uint64
	interval                    time.Duration
}

func NewCPUMetric(name string, interval time.Duration, healthCondition []metric.HealthCondition[float64]) *CPUMetric {
	return &CPUMetric{
		Base: metric.Base[float64]{
			Name:             name,
			HealthConditions: healthCondition,
		},
		interval: interval,
	}
}

func (c *CPUMetric) Measure(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.With("metric_name", c.Name).Debug("metric was stopped")
			return
		case <-ticker.C:
			c.measure()
		}
	}
}

func (c *CPUMetric) measure() {
	cpu, err := cpu.Get()
	if err != nil {
		logger.WriteError(metric.InfrastructureGroup, c.Name, err)
		return
	}
	systemPercent := float64(cpu.System-c.prevSystem) / float64(cpu.Total-c.total) * 100
	userPercent := float64(cpu.User-c.prevUser) / float64(cpu.Total-c.total) * 100

	c.prevUser = cpu.User
	c.prevSystem = cpu.System
	c.total = cpu.Total

	c.writeMetric(systemPercent, userPercent)
}

func (c *CPUMetric) writeMetric(systemPercent, userPercent float64) {
	c.AddDataPoint(map[string]float64{
		SystemCPUMeasurement: systemPercent,
		UserCPUMeasurement:   userPercent,
	})

	cpuUsageMetric.With(prometheus.Labels{cpuUsageTypeLabel: "system"}).Set(float64(systemPercent))
	cpuUsageMetric.With(prometheus.Labels{cpuUsageTypeLabel: "user"}).Set(float64(userPercent))

	logger.WriteMetric(metric.InfrastructureGroup, c.Name, map[string]any{
		SystemCPUMeasurement: systemPercent,
		UserCPUMeasurement:   userPercent,
	})
}

func (c *CPUMetric) AggregateResults() string {
	var values map[string][]float64 = make(map[string][]float64)

	for _, point := range c.DataPoints {
		values[SystemCPUMeasurement] = append(values[SystemCPUMeasurement], point.Values[SystemCPUMeasurement])
		values[UserCPUMeasurement] = append(values[UserCPUMeasurement], point.Values[UserCPUMeasurement])
	}

	return fmt.Sprintf("user_P50=%.2f%%, system_P50=%.2f%%, total=%v",
		metric.CalculatePercentiles(values[UserCPUMeasurement], 50)[50],
		metric.CalculatePercentiles(values[SystemCPUMeasurement], 50)[50], c.total)
}
