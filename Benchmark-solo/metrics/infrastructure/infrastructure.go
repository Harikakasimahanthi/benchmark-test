package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/disk"
	"github.com/mackerelio/go-osstat/memory"
	"github.com/mackerelio/go-osstat/network"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Define Prometheus metrics
var (
	// CPU Metrics
	cpuUser   = prometheus.NewGauge(prometheus.GaugeOpts{Name: "cpu_user_percent", Help: "User CPU usage"})
	cpuSystem = prometheus.NewGauge(prometheus.GaugeOpts{Name: "cpu_system_percent", Help: "System CPU usage"})

	// Memory Metrics
	memUsed   = prometheus.NewGauge(prometheus.GaugeOpts{Name: "memory_used_mb", Help: "Memory used in MB"})
	memFree   = prometheus.NewGauge(prometheus.GaugeOpts{Name: "memory_free_mb", Help: "Memory free in MB"})
	memTotal  = prometheus.NewGauge(prometheus.GaugeOpts{Name: "memory_total_mb", Help: "Total system memory in MB"})
	memCached = prometheus.NewGauge(prometheus.GaugeOpts{Name: "memory_cached_mb", Help: "Cached memory in MB"})

	// Disk Metrics
	diskUsed  = prometheus.NewGauge(prometheus.GaugeOpts{Name: "disk_used_gb", Help: "Disk used in GB"})
	diskAvail = prometheus.NewGauge(prometheus.GaugeOpts{Name: "disk_available_gb", Help: "Disk available in GB"})

	// Network Metrics
	netSent     = prometheus.NewGauge(prometheus.GaugeOpts{Name: "network_sent_mb", Help: "Network bytes sent in MB"})
	netReceived = prometheus.NewGauge(prometheus.GaugeOpts{Name: "network_received_mb", Help: "Network bytes received in MB"})
)

// Variables for CPU tracking
var prevUser, prevSystem, prevTotal uint64

// Convert bytes to megabytes
func toMegabytes(bytes uint64) float64 {
	return float64(bytes) / (1024 * 1024)
}

// Convert bytes to gigabytes
func toGigabytes(bytes uint64) float64 {
	return float64(bytes) / (1024 * 1024 * 1024)
}

// Collects CPU usage
func collectCPUUsage() {
	cpuStat, err := cpu.Get()
	if err != nil {
		log.Println("Error getting CPU stats:", err)
		return
	}

	if prevTotal > 0 {
		totalDelta := cpuStat.Total - prevTotal
		if totalDelta > 0 {
			userPercent := float64(cpuStat.User-prevUser) / float64(totalDelta) * 100
			systemPercent := float64(cpuStat.System-prevSystem) / float64(totalDelta) * 100

			cpuUser.Set(userPercent)
			cpuSystem.Set(systemPercent)
		}
	}

	prevUser, prevSystem, prevTotal = cpuStat.User, cpuStat.System, cpuStat.Total
}

// Collects memory usage
func collectMemoryUsage() {
	memStat, err := memory.Get()
	if err != nil {
		log.Println("Error getting memory stats:", err)
		return
	}

	memUsed.Set(toMegabytes(memStat.Used))
	memFree.Set(toMegabytes(memStat.Free))
	memTotal.Set(toMegabytes(memStat.Total))
	memCached.Set(toMegabytes(memStat.Cached))
}

// Collects disk usage
func collectDiskUsage() {
	diskStat, err := disk.Get()
	if err != nil {
		log.Println("Error getting disk stats:", err)
		return
	}

	diskUsed.Set(toGigabytes(diskStat.Used))
	diskAvail.Set(toGigabytes(diskStat.Available))
}

// Collects network usage
func collectNetworkUsage() {
	netStats, err := network.Get()
	if err != nil {
		log.Println("Error getting network stats:", err)
		return
	}

	var totalSent, totalReceived uint64
	for _, iface := range netStats {
		totalSent += iface.TxBytes
		totalReceived += iface.RxBytes
	}

	netSent.Set(toMegabytes(totalSent))
	netReceived.Set(toMegabytes(totalReceived))
}

// Periodically collect system metrics
func startMonitoring(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Monitoring stopped")
			return
		case <-ticker.C:
			collectCPUUsage()
			collectMemoryUsage()
			collectDiskUsage()
			collectNetworkUsage()
		}
	}
}

func main() {
	// Register Prometheus metrics
	prometheus.MustRegister(
		cpuUser, cpuSystem,
		memUsed, memFree, memTotal, memCached,
		diskUsed, diskAvail,
		netSent, netReceived,
	)

	// Start background monitoring
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go startMonitoring(ctx, 10*time.Second)

	// Start HTTP server for Prometheus
	http.Handle("/metrics", promhttp.Handler())
	log.Println("Prometheus metrics server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
