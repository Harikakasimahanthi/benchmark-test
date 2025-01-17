package metric

type (
	HealthStatus  string
	SeverityLevel string
	Operator      string
)

const (
	Healthy   HealthStatus = "Healthy✅"
	Unhealthy HealthStatus = "Unhealthy⚠️"

	SeverityNone   SeverityLevel = "None"
	SeverityLow    SeverityLevel = "Low"
	SeverityMedium SeverityLevel = "Medium"
	SeverityHigh   SeverityLevel = "High"

	OperatorGreaterThan        Operator = ">"
	OperatorLessThan           Operator = "<"
	OperatorGreaterThanOrEqual Operator = ">="
	OperatorLessThanOrEqual    Operator = "<="
	OperatorEqual              Operator = "=="
)

type HealthCondition[T Metricable] struct {
	Name      string
	Threshold T
	Operator  Operator
	Severity  SeverityLevel
}

func (c HealthCondition[T]) Evaluate(value T) bool {
	switch c.Operator {
	case OperatorGreaterThan:
		return value > c.Threshold
	case OperatorLessThan:
		return value < c.Threshold
	case OperatorGreaterThanOrEqual:
		return value >= c.Threshold
	case OperatorLessThanOrEqual:
		return value <= c.Threshold
	case OperatorEqual:
		return value == c.Threshold
	default:
		return false
	}
}

var severityOrder = map[SeverityLevel]int{
	SeverityLow:    1,
	SeverityMedium: 2,
	SeverityHigh:   3,
}

func CompareSeverities(a, b SeverityLevel) int {
	return severityOrder[a] - severityOrder[b]
}
