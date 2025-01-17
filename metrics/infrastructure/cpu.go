package infrastructure

import (
	"context"
	"time"

	"github.com/Harikakasimahanthi/benchmark-test/internal/platform/logger"
	"github.com/Harikakasimahanthi/benchmark-test/internal/platform/metric"
	"github.com/mackerelio/go-osstat/cpu"
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

	logger.WriteMetric(metric.InfrastructureGroup, c.Name, map[string]any{
		SystemCPUMeasurement: systemPercent,
		UserCPUMeasurement:   userPercent,
	})
}
