package benchmark

import (
	"errors"
	"time"

	"github.com/Harikakasimahanthi/benchmark-test/configs"
	"github.com/Harikakasimahanthi/benchmark-test/internal/platform/network"
	"github.com/Harikakasimahanthi/benchmark-test/metrics/consensus"
	"github.com/Harikakasimahanthi/benchmark-test/metrics/execution"
	"github.com/Harikakasimahanthi/benchmark-test/metrics/infrastructure"
)

func LoadEnabledMetrics(config configs.Config) (map[metric.Group][]metricService, error) {
	enabledMetrics := make(map[metric.Group][]metricService)

	// Consensus metrics
	if config.Benchmark.Consensus.Metrics.Client.Enabled {
		enabledMetrics[metric.ConsensusGroup] = append(enabledMetrics[metric.ConsensusGroup], consensus.NewClientMetric(
			configs.Values.Benchmark.Consensus.Address,
			"Client",
			[]metric.HealthCondition[string]{
				{Name: consensus.VersionMeasurement, Threshold: "", Operator: metric.OperatorEqual, Severity: metric.SeverityHigh},
			}))
	}

	if config.Benchmark.Consensus.Metrics.Latency.Enabled {
		consensusClientURL, err := config.Benchmark.Consensus.AddrURL()
		if err != nil {
			return nil, errors.Join(err, errors.New("failed fetching Consensus client address as URL"))
		}
		enabledMetrics[metric.ConsensusGroup] = append(enabledMetrics[metric.ConsensusGroup], consensus.NewLatencyMetric(
			consensusClientURL.Host,
			"Latency",
			time.Second*3,
			[]metric.HealthCondition[time.Duration]{
				{Name: consensus.DurationP90Measurement, Threshold: time.Second, Operator: metric.OperatorGreaterThanOrEqual, Severity: metric.SeverityHigh},
			}))
	}

	if config.Benchmark.Consensus.Metrics.Peers.Enabled {
		enabledMetrics[metric.ConsensusGroup] = append(enabledMetrics[metric.ConsensusGroup], consensus.NewPeerMetric(
			configs.Values.Benchmark.Consensus.Address,
			"Peers",
			time.Second*10,
			[]metric.HealthCondition[uint32]{
				{Name: consensus.PeerCountMeasurement, Threshold: 5, Operator: metric.OperatorLessThanOrEqual, Severity: metric.SeverityHigh},
				{Name: consensus.PeerCountMeasurement, Threshold: 20, Operator: metric.OperatorLessThanOrEqual, Severity: metric.SeverityMedium},
				{Name: consensus.PeerCountMeasurement, Threshold: 40, Operator: metric.OperatorLessThanOrEqual, Severity: metric.SeverityLow},
			}))
	}

	if config.Benchmark.Consensus.Metrics.Attestation.Enabled {
		enabledMetrics[metric.ConsensusGroup] = append(enabledMetrics[metric.ConsensusGroup], consensus.NewAttestationMetric(
			configs.Values.Benchmark.Consensus.Address,
			"Attestation",
			network.GenesisTime[network.Name(config.Benchmark.Network)],
			[]metric.HealthCondition[float64]{
				{Name: consensus.CorrectnessMeasurement, Threshold: 97, Operator: metric.OperatorLessThanOrEqual, Severity: metric.SeverityHigh},
				{Name: consensus.CorrectnessMeasurement, Threshold: 98.5, Operator: metric.OperatorLessThanOrEqual, Severity: metric.SeverityMedium},
			},
		))
	}

	// Execution metrics
	if config.Benchmark.Execution.Metrics.Peers.Enabled {
		enabledMetrics[metric.ExecutionGroup] = append(enabledMetrics[metric.ExecutionGroup], execution.NewPeerMetric(
			configs.Values.Benchmark.Execution.Address,
			"Peers",
			time.Second*10,
			[]metric.HealthCondition[uint32]{
				{Name: execution.PeerCountMeasurement, Threshold: 5, Operator: metric.OperatorLessThanOrEqual, Severity: metric.SeverityHigh},
				{Name: execution.PeerCountMeasurement, Threshold: 20, Operator: metric.OperatorLessThanOrEqual, Severity: metric.SeverityMedium},
				{Name: execution.PeerCountMeasurement, Threshold: 40, Operator: metric.OperatorLessThanOrEqual, Severity: metric.SeverityLow},
			}))
	}

	if config.Benchmark.Execution.Metrics.Latency.Enabled {
		executionClientURL, err := config.Benchmark.Execution.AddrURL()
		if err != nil {
			return nil, errors.Join(err, errors.New("failed fetching Execution client address as URL"))
		}
		enabledMetrics[metric.ExecutionGroup] = append(enabledMetrics[metric.ExecutionGroup], execution.NewLatencyMetric(
			executionClientURL.Host,
			"Latency",
			time.Second*3,
			[]metric.HealthCondition[time.Duration]{
				{Name: execution.DurationP90Measurement, Threshold: time.Second, Operator: metric.OperatorGreaterThanOrEqual, Severity: metric.SeverityHigh},
			}))
	}

	// Infrastructure metrics
	if config.Benchmark.Infrastructure.Metrics.CPU.Enabled {
		enabledMetrics[metric.InfrastructureGroup] = append(enabledMetrics[metric.InfrastructureGroup],
			infrastructure.NewCPUMetric("CPU", time.Second*5, []metric.HealthCondition[float64]{}),
		)
	}

	if config.Benchmark.Infrastructure.Metrics.Memory.Enabled {
		enabledMetrics[metric.InfrastructureGroup] = append(enabledMetrics[metric.InfrastructureGroup],
			infrastructure.NewMemoryMetric("Memory", time.Second*10, []metric.HealthCondition[uint64]{
				{Name: infrastructure.FreeMemoryMeasurement, Threshold: 0, Operator: metric.OperatorEqual, Severity: metric.SeverityHigh},
			}),
		)
	}

	return enabledMetrics, nil
}
