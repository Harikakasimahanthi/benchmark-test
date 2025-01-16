package benchmark

import (
	"context"
	"log/slog"

	"github.com/ssvlabs/ssv-pulse/internal/benchmark/report"
	"github.com/ssvlabs/ssv-pulse/internal/platform/metric"
)

type (
	metricService interface {
		Measure(context.Context)
		GetName() string
		AggregateResults() string
		EvaluateMetric() (metric.HealthStatus, map[string]metric.SeverityLevel)
	}
	reportService interface {
		AddRecord(metric report.Record)
		Render()
	}

	Service struct {
		metrics map[metric.Group][]metricService
		report  reportService
	}
)

func New(
	metrics map[metric.Group][]metricService,
	reportService reportService,
) *Service {
	return &Service{
		metrics: metrics,
		report:  reportService,
	}
}

func (s *Service) Start(ctx context.Context) {
	slog.With("metrics", s.metrics).Debug("starting benchmark service")

	for _, groupMetrics := range s.metrics {
		for _, metric := range groupMetrics {
			go metric.Measure(ctx)
		}
	}

	<-ctx.Done()

	for metricGroup, groupMetrics := range s.metrics {
		for _, m := range groupMetrics {
			health, severity := m.EvaluateMetric()

			slog.With("metric_group", metricGroup).With("metric_name", m.GetName()).Info("adding report record")
			s.report.AddRecord(report.Record{
				GroupName:  metricGroup,
				MetricName: m.GetName(),
				Value:      m.AggregateResults(),
				Health:     health,
				Severity:   severity,
			})
		}
	}

	slog.Info("rendering")
	s.report.Render()
}
