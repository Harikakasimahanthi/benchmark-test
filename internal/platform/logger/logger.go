package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/ssvlabs/ssv-pulse/internal/platform/metric"
)

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)
}

func WriteMetric(metricGroup metric.Group, metricName string, nameValue map[string]any) {
	logger := slog.Default()

	logger.
		With("metric_group", strings.ToLower(string(metricGroup))).
		With("metric_name", strings.ToLower(string(metricName))).
		With("values", nameValue).
		Debug("measured")
}

func WriteError(metricGroup metric.Group, metricName string, err error) {
	slog.
		With("err", err.Error()).
		With("metric_group", strings.ToLower(string(metricGroup))).
		With("metric_name", strings.ToLower(string(metricName))).
		Error("error")
}
