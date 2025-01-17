package main

import (
	"errors"
	"log/slog"

	"github.com/ssvlabs/ssv-pulse/internal/loki"

	"github.com/Harikakasimahanthi/benchmark-test/"
	"github.com/Harikakasimahanthi/benchmark-test/configs"
	"github.com/Harikakasimahanthi/benchmark-test/internal/platform/cmd"
	_ "github.com/Harikakasimahanthi/benchmark-test/internal/platform/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	appName = "solostaking-benchmark"
	version = "1.0"
)

var rootCmd = &cobra.Command{
	Use:   "solostaking-benchmark",
	Short: "CLI for analyzing and benchmarking ssv node",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")

		if err := viper.ReadInConfig(); err != nil {
			const errMsg = "error reading config file"
			slog.With("err", err.Error()).Error(errMsg)
			return errors.Join(err, errors.New(errMsg))
		}
		if err := viper.Unmarshal(&configs.Values); err != nil {
			const errMsg = "unable to decode application config"
			slog.With("err", err.Error()).Error(errMsg)
			return errors.Join(err, errors.New(errMsg))
		}

		slog.
			With("config_file", viper.ConfigFileUsed()).
			With("config", configs.Values).
			Debug("configurations loaded")
		return nil
	},
}

func main() {
	rootCmd.Short = appName
	rootCmd.Version = version

	rootCmd.AddCommand(analyzer.CMD)
	rootCmd.AddCommand(benchmark.CMD)
	rootCmd.AddCommand(cmd.Version)
	rootCmd.AddCommand(loki.CMD)
	if err := rootCmd.Execute(); err != nil {
		slog.With("err", err.Error()).Error("failed to execute root command")
		panic(err.Error())
	}
}
