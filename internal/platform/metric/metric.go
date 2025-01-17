package metric

import (
	"time"

	"golang.org/x/exp/constraints"
)

type (
	Metricable interface {
		constraints.Integer | constraints.Float | ~string
	}

	Base[T Metricable] struct {
		Name             string
		DataPoints       []DataPoint[T]
		HealthConditions []HealthCondition[T]
	}

	DataPoint[T Metricable] struct {
		Timestamp time.Time
		Values    map[string]T
	}
)

func (bm *Base[T]) GetName() string {
	return bm.Name
}

func (bm *Base[T]) AddDataPoint(values map[string]T) {
	bm.DataPoints = append(bm.DataPoints, DataPoint[T]{
		Timestamp: time.Now(),
		Values:    values,
	})
}

func (bm *Base[T]) EvaluateMetric() (HealthStatus, map[string]SeverityLevel) {
	overallHealth := Healthy
	maxSeverities := make(map[string]SeverityLevel)

	for _, dp := range bm.DataPoints {
		for name := range dp.Values {
			maxSeverities[name] = SeverityNone
		}
	}

	for _, dp := range bm.DataPoints {
		for name, value := range dp.Values {
			for _, condition := range bm.HealthConditions {
				if condition.Name == name {
					if condition.Evaluate((value)) {
						overallHealth = Unhealthy
						if CompareSeverities(condition.Severity, maxSeverities[name]) > 0 {
							maxSeverities[name] = condition.Severity
						}
					}
				}
			}
		}
	}

	return overallHealth, maxSeverities
}
