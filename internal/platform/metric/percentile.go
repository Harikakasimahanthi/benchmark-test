package metric

import (
	"fmt"
	"sort"
)

type (
	stringable interface {
		~int | ~int64 | ~float64 | ~uint64 | ~uint16 | ~uint32 | string
	}
)

func CalculatePercentiles[T Metricable](values []T, percentiles ...float64) map[float64]T {
	result := make(map[float64]T)

	if len(percentiles) == 0 {
		return result
	}

	if len(values) == 0 {
		var zero T
		for _, p := range percentiles {
			result[p] = zero
		}
		return result
	}

	sort.Slice(values, func(i, j int) bool {
		return values[i] < values[j]
	})

	for _, percentile := range percentiles {
		index := int(float64(len(values)-1) * percentile / 100.0)
		result[percentile] = values[index]
	}

	return result
}

func FormatPercentiles[T stringable](min, p10, p50, p90, max T) string {
	return fmt.Sprintf("min=%v, p10=%v, p50=%v, p90=%v, max=%v", min, p10, p50, p90, max)
}
