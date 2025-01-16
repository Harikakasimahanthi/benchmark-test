package benchmark

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ssvlabs/ssv-pulse/configs"
	"github.com/ssvlabs/ssv-pulse/internal/benchmark/report"
	"github.com/ssvlabs/ssv-pulse/internal/platform/lifecycle"
	"github.com/ssvlabs/ssv-pulse/internal/platform/server/host"
	"github.com/ssvlabs/ssv-pulse/internal/platform/server/route"
)

const (
	durationFlag             = "duration"
	defaultExecutionDuration = time.Minute * 15

	serverPortFlag    = "port"
	defaultServerPort = 8080

	consensusAddrFlag              = "consensus-addr"
	consensusMetricClientFlag      = "consensus-metric-client-enabled"
	consensusMetricLatencyFlag     = "consensus-metric-latency-enabled"
	consensusMetricPeersFlag       = "consensus-metric-peers-enabled"
	consensusMetricAttestationFlag = "consensus-metric-attestation-enabled"

	executionAddrFlag          = "execution-addr"
	executionMetricPeersFlag   = "execution-metric-peers-enabled"
	executionMetricLatencyFlag = "execution-metric-latency-enabled"

	ssvAddrFlag              = "ssv-addr"
	ssvMetricPeersFlag       = "ssv-metric-peers-enabled"
	ssvMetricConnectionsFlag = "ssv-metric-connections-enabled"

	infraMetricCPUFlag    = "infra-metric-cpu-enabled"
	infraMetricMemoryFlag = "infra-metric-memory-enabled"

	networkFlag = "network"
)

func init() {
	addFlags(CMD)
	if err := bindFlags(CMD); err != nil {
		panic(err.Error())
	}
}

var CMD = &cobra.Command{
	Use:   "benchmark",
	Short: "Run benchmarks of ssv node",
	Run: func(cobraCMD *cobra.Command, args []string) {
		var (
			ctx    context.Context
			cancel context.CancelFunc
		)
		if configs.Values.Benchmark.Duration == 0 {
			ctx, cancel = context.WithCancel(context.Background())
		} else {
			ctx, cancel = context.WithTimeout(context.Background(), configs.Values.Benchmark.Duration)
		}

		isValid, err := configs.Values.Benchmark.Validate()
		if !isValid {
			panic(err.Error())
		}

		metrics, err := LoadEnabledMetrics(configs.Values)
		if err != nil {
			panic(err.Error())
		}

		benchmarkService := New(metrics, report.New())

		go benchmarkService.Start(ctx)

		slog.With("port", configs.Values.Benchmark.Server.Port).Info("running web host")
		host := host.New(configs.Values.Benchmark.Server.Port,
			route.
				NewRouter().
				WithMetrics().
				Router())
		host.Run()

		lifecycle.ListenForApplicationShutDown(ctx, func() {
			cancel()
			slog.Warn("terminating the application")
		}, make(chan os.Signal))
	},
}

func addFlags(cobraCMD *cobra.Command) {
	cobraCMD.Flags().Duration(durationFlag, defaultExecutionDuration, "Duration for which the application will run to gather metrics, e.g. '5m'")
	cobraCMD.Flags().Uint16(serverPortFlag, defaultServerPort, "Web server port with metrics endpoint exposed, e.g. '8080'")
	cobraCMD.Flags().String(consensusAddrFlag, "", "Consensus client address (beacon node API) with scheme (HTTP/HTTPS) and port, e.g. https://lighthouse:5052")
	cobraCMD.Flags().Bool(consensusMetricClientFlag, true, "Enable consensus client metric")
	cobraCMD.Flags().Bool(consensusMetricLatencyFlag, true, "Enable consensus client latency metric")
	cobraCMD.Flags().Bool(consensusMetricPeersFlag, true, "Enable consensus client peers metric")
	cobraCMD.Flags().Bool(consensusMetricAttestationFlag, true, "Enable consensus client attestation metric")

	cobraCMD.Flags().String(executionAddrFlag, "", "Execution client address with scheme (HTTP/HTTPS) and port, e.g. https://geth:8545")
	cobraCMD.Flags().Bool(executionMetricPeersFlag, true, "Enable execution client peers metric")
	cobraCMD.Flags().Bool(executionMetricLatencyFlag, true, "Enable execution client latency metric")

	cobraCMD.Flags().String(ssvAddrFlag, "", "SSV API address with scheme (HTTP/HTTPS) and port, e.g. http://ssv-node:16000")
	cobraCMD.Flags().Bool(ssvMetricPeersFlag, true, "Enable SSV client peers metric")
	cobraCMD.Flags().Bool(ssvMetricConnectionsFlag, true, "Enable SSV client connections metric")

	cobraCMD.Flags().Bool(infraMetricCPUFlag, true, "Enable infrastructure CPU metric")
	cobraCMD.Flags().Bool(infraMetricMemoryFlag, true, "Enable infrastructure memory metric")

	cobraCMD.Flags().String(networkFlag, "", "Ethereum network to use, either 'mainnet' or 'holesky'")
}

func bindFlags(cmd *cobra.Command) error {
	if err := viper.BindPFlag("benchmark.duration", cmd.Flags().Lookup(durationFlag)); err != nil {
		return err
	}
	if err := viper.BindPFlag("benchmark.server.port", cmd.Flags().Lookup(serverPortFlag)); err != nil {
		return err
	}
	if err := viper.BindPFlag("benchmark.consensus.address", cmd.Flags().Lookup(consensusAddrFlag)); err != nil {
		return err
	}
	if err := viper.BindPFlag("benchmark.execution.address", cmd.Flags().Lookup(executionAddrFlag)); err != nil {
		return err
	}
	if err := viper.BindPFlag("benchmark.ssv.address", cmd.Flags().Lookup(ssvAddrFlag)); err != nil {
		return err
	}
	if err := viper.BindPFlag("benchmark.network", cmd.Flags().Lookup(networkFlag)); err != nil {
		return err
	}
	if err := viper.BindPFlag("benchmark.consensus.metrics.client.enabled", cmd.Flags().Lookup(consensusMetricClientFlag)); err != nil {
		return err
	}
	if err := viper.BindPFlag("benchmark.consensus.metrics.latency.enabled", cmd.Flags().Lookup(consensusMetricLatencyFlag)); err != nil {
		return err
	}
	if err := viper.BindPFlag("benchmark.consensus.metrics.peers.enabled", cmd.Flags().Lookup(consensusMetricPeersFlag)); err != nil {
		return err
	}
	if err := viper.BindPFlag("benchmark.consensus.metrics.attestation.enabled", cmd.Flags().Lookup(consensusMetricAttestationFlag)); err != nil {
		return err
	}
	if err := viper.BindPFlag("benchmark.execution.metrics.peers.enabled", cmd.Flags().Lookup(executionMetricPeersFlag)); err != nil {
		return err
	}
	if err := viper.BindPFlag("benchmark.execution.metrics.latency.enabled", cmd.Flags().Lookup(executionMetricLatencyFlag)); err != nil {
		return err
	}
	if err := viper.BindPFlag("benchmark.ssv.metrics.peers.enabled", cmd.Flags().Lookup(ssvMetricPeersFlag)); err != nil {
		return err
	}
	if err := viper.BindPFlag("benchmark.ssv.metrics.connections.enabled", cmd.Flags().Lookup(ssvMetricConnectionsFlag)); err != nil {
		return err
	}
	if err := viper.BindPFlag("benchmark.infrastructure.metrics.cpu.enabled", cmd.Flags().Lookup(infraMetricCPUFlag)); err != nil {
		return err
	}
	if err := viper.BindPFlag("benchmark.infrastructure.metrics.memory.enabled", cmd.Flags().Lookup(infraMetricMemoryFlag)); err != nil {
		return err
	}

	return nil
}
