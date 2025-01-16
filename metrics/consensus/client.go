package consensus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ssvlabs/ssv-pulse/internal/platform/logger"
	"github.com/ssvlabs/ssv-pulse/internal/platform/metric"
)

const (
	VersionMeasurement    = "Version"
	NodeHealthMeasurement = "NodeHealth"
	SyncStatusMeasurement = "SyncStatus"
	LatencyMeasurement    = "Latency"
)

type ClientMetric struct {
	metric.Base[string]
	url             string
	measureInterval time.Duration
}

func NewClientMetric(url, name string, healthCondition []metric.HealthCondition[string], measureInterval time.Duration) *ClientMetric {
	return &ClientMetric{
		url:             url,
		measureInterval: measureInterval,
		Base: metric.Base[string]{
			HealthConditions: healthCondition,
			Name:             name,
		},
	}
}

func (c *ClientMetric) Measure(ctx context.Context) {
	ticker := time.NewTicker(c.measureInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Fetch version and health data
			c.measureNodeHealth(ctx)
			c.measureNodeVersion(ctx)

			// Measure additional metrics like sync status and latency
			c.measureSyncStatus(ctx)
			c.measureLatency(ctx)
		case <-ctx.Done():
			logger.WriteError(metric.ConsensusGroup, c.Name, fmt.Errorf("client metric measurement stopped"))
			return
		}
	}
}

func (c *ClientMetric) measureNodeHealth(ctx context.Context) {
	// Check the health of the node (replace with actual health check endpoint if available)
	res, err := http.Get(fmt.Sprintf("%s/eth/v1/node/health", c.url))
	if err != nil || res.StatusCode != http.StatusOK {
		c.AddDataPoint(map[string]string{
			NodeHealthMeasurement: "Unhealthy",
		})
		logger.WriteError(metric.ConsensusGroup, c.Name, err)
		return
	}
	defer res.Body.Close()

	c.AddDataPoint(map[string]string{
		NodeHealthMeasurement: "Healthy",
	})
	logger.WriteMetric(metric.ConsensusGroup, c.Name, map[string]any{
		NodeHealthMeasurement: "Healthy",
	})
}

func (c *ClientMetric) measureNodeVersion(ctx context.Context) {
	var resp struct {
		Data struct {
			Version string `json:"version"`
		} `json:"data"`
	}
	res, err := http.Get(fmt.Sprintf("%s/eth/v1/node/version", c.url))
	if err != nil {
		c.AddDataPoint(map[string]string{
			VersionMeasurement: "",
		})
		logger.WriteError(metric.ConsensusGroup, c.Name, err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		c.AddDataPoint(map[string]string{
			VersionMeasurement: "",
		})
		var errorResponse any
		_ = json.NewDecoder(res.Body).Decode(&errorResponse)
		jsonErrResponse, _ := json.Marshal(errorResponse)
		logger.WriteError(
			metric.ConsensusGroup,
			c.Name,
			fmt.Errorf("received unsuccessful status code. Code: '%s'. Response: '%s'", res.Status, jsonErrResponse))
		return
	}

	if err = json.NewDecoder(res.Body).Decode(&resp); err != nil {
		c.AddDataPoint(map[string]string{
			VersionMeasurement: "",
		})
		logger.WriteError(metric.ConsensusGroup, c.Name, err)
		return
	}

	c.AddDataPoint(map[string]string{
		VersionMeasurement: resp.Data.Version,
	})

	logger.WriteMetric(metric.ConsensusGroup, c.Name, map[string]any{VersionMeasurement: resp.Data.Version})
}

func (c *ClientMetric) measureSyncStatus(ctx context.Context) {
	// Measure sync status (replace with actual sync status endpoint if available)
	res, err := http.Get(fmt.Sprintf("%s/eth/v1/node/syncing", c.url))
	if err != nil || res.StatusCode != http.StatusOK {
		c.AddDataPoint(map[string]string{
			SyncStatusMeasurement: "Not Synced",
		})
		logger.WriteError(metric.ConsensusGroup, c.Name, err)
		return
	}
	defer res.Body.Close()

	c.AddDataPoint(map[string]string{
		SyncStatusMeasurement: "Synced",
	})
	logger.WriteMetric(metric.ConsensusGroup, c.Name, map[string]any{
		SyncStatusMeasurement: "Synced",
	})
}

func (c *ClientMetric) measureLatency(ctx context.Context) {
	startTime := time.Now()
	res, err := http.Get(fmt.Sprintf("%s/eth/v1/node/health", c.url)) // Using health endpoint for latency check
	if err != nil {
		c.AddDataPoint(map[string]string{
			LatencyMeasurement: "Error",
		})
		logger.WriteError(metric.ConsensusGroup, c.Name, err)
		return
	}
	defer res.Body.Close()

	latency := time.Since(startTime).Milliseconds()
	c.AddDataPoint(map[string]string{
		LatencyMeasurement: fmt.Sprintf("%dms", latency),
	})
	logger.WriteMetric(metric.ConsensusGroup, c.Name, map[string]any{
		LatencyMeasurement: fmt.Sprintf("%dms", latency),
	})
}

func (c *ClientMetric) AggregateResults() string {
	var version, health, syncStatus, latency string

	if len(c.DataPoints) != 0 {
		for _, point := range c.DataPoints {
			if versionValue, ok := point.Values[VersionMeasurement]; ok {
				version = versionValue
			}
			if healthValue, ok := point.Values[NodeHealthMeasurement]; ok {
				health = healthValue
			}
			if syncValue, ok := point.Values[SyncStatusMeasurement]; ok {
				syncStatus = syncValue
			}
			if latencyValue, ok := point.Values[LatencyMeasurement]; ok {
				latency = latencyValue
			}
		}
	}

	return fmt.Sprintf(
		"Version: %s, Node Health: %s, Sync Status: %s, Latency: %s",
		version, health, syncStatus, latency)
}
