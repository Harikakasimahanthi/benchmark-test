package report

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/aquasecurity/table"
	"github.com/ssvlabs/ssv-pulse/internal/platform/metric"
)

var headers = []string{"Group Name", "Metric Name", "Value", "Health", "Severity"}

type Record struct {
	GroupName  metric.Group
	MetricName string
	Value      string
	Health     metric.HealthStatus
	Severity   map[string]metric.SeverityLevel
}

type Report struct {
	t     *table.Table
	mutex sync.Mutex
}

func New() *Report {
	t := table.New(os.Stdout)

	t.SetHeaders(headers...)

	var alignments []table.Alignment
	for i := 0; i < len(headers); i++ {
		alignments = append(alignments, table.AlignCenter)
	}
	t.SetAlignment(alignments...)

	return &Report{
		t: t,
	}
}

func (r *Report) AddRecord(metric Record) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.t.AddRow(
		string(metric.GroupName),
		string(metric.MetricName),
		metric.Value,
		string(metric.Health),
		formatSeverityMap(metric.Severity),
	)
}

func (r *Report) Render() {
	r.t.Render()
}

func formatSeverityMap(severityMap map[string]metric.SeverityLevel) string {
	var builder strings.Builder

	for name, severity := range severityMap {
		builder.WriteString(fmt.Sprintf("%s: %s, ", name, severity))
	}

	// Remove the trailing comma and space, if necessary
	result := builder.String()
	if len(result) > 2 {
		result = result[:len(result)-2]
	}

	return result
}
