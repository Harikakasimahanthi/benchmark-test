package metric

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGivenRangeOfPercentilesWhenCalculatePercentilesThenCalculatesCorrectly(t *testing.T) {
	tests := []struct {
		name        string
		values      []int
		percentiles []float64
		expected    map[float64]int
	}{
		{
			name:        "Valid percentiles",
			values:      []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			percentiles: []float64{10, 50, 90},
			expected:    map[float64]int{10: 1, 50: 5, 90: 9},
		},
		{
			name:        "No percentiles provided",
			values:      []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			percentiles: []float64{},
			expected:    map[float64]int{},
		},
		{
			name:        "Empty values list",
			values:      []int{},
			percentiles: []float64{10, 50, 90},
			expected:    map[float64]int{10: 0, 50: 0, 90: 0},
		},
		{
			name:        "Empty values and no percentiles",
			values:      []int{},
			percentiles: []float64{},
			expected:    map[float64]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculatePercentiles(tt.values, tt.percentiles...)
			assert.Equal(t, tt.expected, result)
		})
	}
}
