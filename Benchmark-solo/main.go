package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// Metric Structs
type Metric struct {
	Enabled bool `mapstructure:"enabled"`
}

type ConsensusMetrics struct {
	Client      Metric `mapstructure:"client"`
	Latency     Metric `mapstructure:"latency"`
	Peers       Metric `mapstructure:"peers"`
	Attestation Metric `mapstructure:"attestation"`
}

type ExecutionMetrics struct {
	Peers   Metric `mapstructure:"peers"`
	Latency Metric `mapstructure:"latency"`
}

type InfrastructureMetrics struct {
	CPU    Metric `mapstructure:"cpu"`
	Memory Metric `mapstructure:"memory"`
}

// Benchmark Structs
type Consensus struct {
	Address string           `mapstructure:"address"`
	Metrics ConsensusMetrics `mapstructure:"metrics"`
}

type Execution struct {
	Address string           `mapstructure:"address"`
	Metrics ExecutionMetrics `mapstructure:"metrics"`
}

type Infrastructure struct {
	Metrics InfrastructureMetrics `mapstructure:"metrics"`
}

type Benchmark struct {
	Consensus      Consensus      `mapstructure:"consensus"`
	Execution      Execution      `mapstructure:"execution"`
	Infrastructure Infrastructure `mapstructure:"infrastructure"`
}

// MetricValue Struct
type MetricValue struct {
	Min float64 `json:"min"`
	P10 float64 `json:"p10"`
	P50 float64 `json:"p50"`
	P90 float64 `json:"p90"`
	Max float64 `json:"max"`
}

// Fetch HTTP Metrics (Consensus/Execution)
func fetchMetrics(url string) (map[string]MetricValue, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metrics from %s: %v", url, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// If response starts with '#' (Prometheus-style comments) or HTML, skip
	if body[0] == '#' || body[0] == '<' {
		return nil, fmt.Errorf("unexpected response format (not JSON)")
	}

	// Try to parse the response as JSON
	var metrics map[string]MetricValue
	err = json.Unmarshal(body, &metrics)
	if err != nil {
		// If JSON parsing fails, check if the response is a simple number
		var singleMetric float64
		err = json.Unmarshal(body, &singleMetric)
		if err != nil {
			return nil, fmt.Errorf("failed to parse response as JSON or single number: %v", err)
		}

		// In case it's a single numeric value, create a placeholder metric
		metrics = map[string]MetricValue{
			"single_metric": {Min: singleMetric, P10: singleMetric, P50: singleMetric, P90: singleMetric, Max: singleMetric},
		}
	}

	return metrics, nil
}

// Fetch System Metrics (CPU & Memory)
func fetchSystemMetrics() (float64, float64, error) {
	cpuPercentages, err := cpu.Percent(0, false)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get CPU usage: %v", err)
	}
	memStats, err := mem.VirtualMemory()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get Memory usage: %v", err)
	}

	// Return CPU and Memory Usage
	return cpuPercentages[0], memStats.UsedPercent, nil
}

// Print Metrics Table
func printMetricsTable(benchmark *Benchmark) {
	fmt.Println("+-------------------+---------------------+---------------------------------------------------------------+------------------+------------------+")
	fmt.Println("| Group Name        | Metric Name         | Value                                                         | Health           | Severity         |")
	fmt.Println("+-------------------+---------------------+---------------------------------------------------------------+------------------+------------------+")

	displayMetrics(benchmark.Consensus.Address, "Consensus")
	displayMetrics(benchmark.Execution.Address, "Execution")
	displaySystemMetrics("Infrastructure")
}

// Display HTTP-based Metrics
func displayMetrics(url, groupName string) {
	metrics, err := fetchMetrics(url)
	if err != nil {
		fmt.Printf("Error fetching %s metrics: %v\n", groupName, err)
		return
	}

	for metricName, value := range metrics {
		health := "Healthy ✅"
		severity := "None"
		fmt.Printf("| %-17s | %-19s | min=%.3f, p10=%.3f, p50=%.3f, p90=%.3f, max=%.3f | %-16s | %-16s |\n",
			groupName, metricName, value.Min, value.P10, value.P50, value.P90, value.Max, health, severity)
	}
}

// Display Local System Metrics (CPU & Memory)
func displaySystemMetrics(groupName string) {
	cpuUsage, memUsage, err := fetchSystemMetrics()
	if err != nil {
		fmt.Printf("Error fetching %s metrics: %v\n", groupName, err)
		return
	}

	fmt.Printf("| %-17s | %-19s | Current Usage: %.2f%%                              | %-16s | %-16s |\n",
		groupName, "CPU", cpuUsage, "Healthy ✅", "None")
	fmt.Printf("| %-17s | %-19s | Current Usage: %.2f%%                              | %-16s | %-16s |\n",
		groupName, "Memory", memUsage, "Healthy ✅", "None")
}

// Run Metrics Loop
func runRealTimeMetrics(benchmark *Benchmark) {
	duration := 10 * time.Minute
	endTime := time.Now().Add(duration)

	for time.Now().Before(endTime) {
		printMetricsTable(benchmark)
		time.Sleep(10 * time.Second)
	}
}

// Main Function
func main() {
	benchmark := &Benchmark{
		Consensus: Consensus{
			Address: "http://localhost:5054/metrics",
			Metrics: ConsensusMetrics{
				Client:      Metric{Enabled: true},
				Latency:     Metric{Enabled: true},
				Peers:       Metric{Enabled: true},
				Attestation: Metric{Enabled: true},
			},
		},
		Execution: Execution{
			Address: "http://localhost:9001/metrics",
			Metrics: ExecutionMetrics{
				Peers:   Metric{Enabled: true},
				Latency: Metric{Enabled: true},
			},
		},
		Infrastructure: Infrastructure{
			Metrics: InfrastructureMetrics{
				CPU:    Metric{Enabled: true},
				Memory: Metric{Enabled: true},
			},
		},
	}

	runRealTimeMetrics(benchmark)
}
