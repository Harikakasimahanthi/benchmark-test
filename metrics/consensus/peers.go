package consensus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/Harikakasimahanthi/benchmark-test/internal/platform/logger"
	"github.com/Harikakasimahanthi/benchmark-test/internal/platform/metric"
)

const (
	PeerCountMeasurement = "PeerCount"
)

type PeerMetric struct {
	metric.Base[uint32]
	url      string
	interval time.Duration
}

func NewPeerMetric(url, name string, interval time.Duration, healthCondition []metric.HealthCondition[uint32]) *PeerMetric {
	return &PeerMetric{
		url: url,
		Base: metric.Base[uint32]{
			HealthConditions: healthCondition,
			Name:             name,
		},
		interval: interval,
	}
}

func (p *PeerMetric) Measure(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.With("metric_name", p.Name).Debug("metric was stopped")
			return
		case <-ticker.C:
			p.measure(ctx)
		}
	}
}

func (p *PeerMetric) measure(ctx context.Context) {
	var (
		resp struct {
			Data struct {
				Connected string `json:"connected"`
			} `json:"data"`
		}
	)
	// Context timeout set for the request
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/eth/v1/node/peer_count", p.url), nil)
	if err != nil {
		logger.WriteError(metric.ConsensusGroup, p.Name, err)
		return
	}

	// Make HTTP request to fetch peer count
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		p.AddDataPoint(map[string]uint32{
			PeerCountMeasurement: 0,
		})
		logger.WriteError(metric.ConsensusGroup, p.Name, err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		p.AddDataPoint(map[string]uint32{
			PeerCountMeasurement: 0,
		})
		p.logErrorResponse(res)
		return
	}

	// Decode response body
	if err = json.NewDecoder(res.Body).Decode(&resp); err != nil {
		p.AddDataPoint(map[string]uint32{
			PeerCountMeasurement: 0,
		})
		logger.WriteError(metric.ConsensusGroup, p.Name, err)
		return
	}

	// Convert connected peer count to an integer
	peerCount, err := strconv.Atoi(resp.Data.Connected)
	if err != nil {
		p.AddDataPoint(map[string]uint32{
			PeerCountMeasurement: 0,
		})
		logger.WriteError(metric.ConsensusGroup, p.Name, err)
		return
	}

	// Record the peer count metric
	p.writeMetric(peerCount)
}

func (p *PeerMetric) logErrorResponse(res *http.Response) {
	var responseString string
	if res.Header.Get("Content-Type") == "application/json" {
		var errorResponse any
		if err := json.NewDecoder(res.Body).Decode(&errorResponse); err != nil {
			logger.WriteError(
				metric.ConsensusGroup,
				p.Name,
				errors.Join(err, fmt.Errorf("received unsuccessful status code. Code: '%s'. Failed to JSON decode response", res.Status)))
			return
		}
		jsonErrResponse, err := json.Marshal(errorResponse)
		if err != nil {
			logger.WriteError(
				metric.ConsensusGroup,
				p.Name,
				errors.Join(err, fmt.Errorf("received unsuccessful status code. Code: '%s'. Failed to marshal response", res.Status)))
			return
		}
		responseString = string(jsonErrResponse)
	} else {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			logger.WriteError(
				metric.ConsensusGroup,
				p.Name,
				errors.Join(err, fmt.Errorf("received unsuccessful status code. Code: '%s'. Failed to decode response", res.Status)))
			return
		}
		responseString = string(body)
	}

	logger.WriteError(
		metric.ConsensusGroup,
		p.Name,
		fmt.Errorf("received unsuccessful status code. Code: '%s'. Response: '%s'", res.Status, responseString))
}

func (p *PeerMetric) writeMetric(peerCount int) {
	p.AddDataPoint(map[string]uint32{
		PeerCountMeasurement: uint32(peerCount),
	})

	// Assuming Prometheus metric tracking
	peerCountMetric.Set(float64(peerCount))

	logger.WriteMetric(metric.ConsensusGroup, p.Name, map[string]any{PeerCountMeasurement: peerCount})
}

func (p *PeerMetric) AggregateResults() string {
	var values []uint32
	for _, point := range p.DataPoints {
		values = append(values, point.Values[PeerCountMeasurement])
	}

	// Calculate percentiles for peer count
	percentiles := metric.CalculatePercentiles(values, 0, 10, 50, 90, 100)

	return metric.FormatPercentiles(
		percentiles[0],
		percentiles[10],
		percentiles[50],
		percentiles[90],
		percentiles[100])
}
