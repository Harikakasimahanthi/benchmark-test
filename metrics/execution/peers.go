package execution

import (
	"bytes"
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
	PeerCountMeasurement = "Count"
)

var measuringErr = errors.New("UNABLE_TO_MEASURE")

type PeerMetric struct {
	metric.Base[uint32]
	url             string
	interval        time.Duration
	measuringErrors map[string]error
}

func NewPeerMetric(url, name string, interval time.Duration, healthCondition []metric.HealthCondition[uint32]) *PeerMetric {
	return &PeerMetric{
		url: url,
		Base: metric.Base[uint32]{
			HealthConditions: healthCondition,
			Name:             name,
		},
		interval:        interval,
		measuringErrors: make(map[string]error),
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
			Result string `json:"result"`
		}
	)

	// Prepare the JSON-RPC request for peer count
	request := struct {
		Jsonrpc string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  []any  `json:"params"`
		ID      int    `json:"id"`
	}{
		Jsonrpc: "2.0",
		Method:  "net_peerCount",
		Params:  []any{},
		ID:      1,
	}

	// Marshal the request into JSON
	requestBytes, err := json.Marshal(request)
	if err != nil {
		logger.WriteError(metric.ExecutionGroup, p.Name, err)
		return
	}

	// Set the request timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Create the HTTP POST request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.url, bytes.NewBuffer(requestBytes))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		logger.WriteError(metric.ExecutionGroup, p.Name, err)
		return
	}

	// Send the request to the Reth execution layer
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		p.writeMetric(0)
		logger.WriteError(metric.ExecutionGroup, p.Name, err)
		return
	}
	defer res.Body.Close()

	// Handle unsuccessful responses
	if res.StatusCode != http.StatusOK {
		p.writeMetric(0)
		p.logErrorResponse(res)
		return
	}

	// Decode the response body
	if err = json.NewDecoder(res.Body).Decode(&resp); err != nil {
		p.writeMetric(0)
		logger.WriteError(metric.ExecutionGroup, p.Name, err)
		return
	}

	// Parse the peer count from the response (hexadecimal string)
	peerCountHex := resp.Result
	if peerCountHex == "" {
		p.writeMetric(0)
		err := errors.New("peer count RPC response was empty. Most likely net_peerCount RPC method is not supported")
		logger.WriteError(metric.ExecutionGroup, p.Name, err)
		p.measuringErrors[PeerCountMeasurement] = errors.Join(measuringErr, err)
		return
	}

	// Convert the peer count from hex to integer
	peerCount, err := strconv.ParseInt(peerCountHex[2:], 16, 64)
	if err != nil {
		p.writeMetric(0)
		logger.WriteError(metric.ExecutionGroup, p.Name, err)
		return
	}

	// Write the measured peer count to the metric
	p.writeMetric(peerCount)
}

func (p *PeerMetric) logErrorResponse(res *http.Response) {
	var responseString string
	if res.Header.Get("Content-Type") == "application/json" {
		var errorResponse any
		if err := json.NewDecoder(res.Body).Decode(&errorResponse); err != nil {
			logger.WriteError(
				metric.ExecutionGroup,
				p.Name,
				errors.Join(err, fmt.Errorf("received unsuccessful status code. Code: '%s'. Failed to JSON decode response", res.Status)))
			return
		}
		jsonErrResponse, err := json.Marshal(errorResponse)
		if err != nil {
			logger.WriteError(
				metric.ExecutionGroup,
				p.Name,
				errors.Join(err, fmt.Errorf("received unsuccessful status code. Code: '%s'. Failed to marshal response", res.Status)))
			return
		}
		responseString = string(jsonErrResponse)
	} else {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			logger.WriteError(
				metric.ExecutionGroup,
				p.Name,
				errors.Join(err, fmt.Errorf("received unsuccessful status code. Code: '%s'. Failed to decode response", res.Status)))
			return
		}
		responseString = string(body)
	}

	// Log the error response from the server
	logger.WriteError(
		metric.ExecutionGroup,
		p.Name,
		fmt.Errorf("received unsuccessful status code. Code: '%s'. Response: '%s'", res.Status, responseString))
}

func (p *PeerMetric) writeMetric(value int64) {
	// Record the peer count in the metric system
	p.AddDataPoint(map[string]uint32{
		PeerCountMeasurement: uint32(value),
	})

	// Update the metric value (e.g., Prometheus)
	peerCountMetric.Set(float64(value))

	// Log the metric
	logger.WriteMetric(metric.ExecutionGroup, p.Name, map[string]any{PeerCountMeasurement: value})
}

func (p *PeerMetric) AggregateResults() string {
	// Check for any errors encountered during measurement
	for measurementName, err := range p.measuringErrors {
		slog.
			With("metric_name", p.Name).
			With("measurement_name", measurementName).
			With("err", err).
			Warn("error measuring metric")

		// Return the error message
		return err.Error()
	}

	// Collect the peer count values from all data points
	var values []uint32
	for _, point := range p.DataPoints {
		values = append(values, point.Values[PeerCountMeasurement])
	}

	// Calculate and format the percentiles (e.g., min, p10, p50, p90, max)
	percentiles := metric.CalculatePercentiles(values, 0, 10, 50, 90, 100)

	// Return the formatted percentiles
	return metric.FormatPercentiles(
		percentiles[0],
		percentiles[10],
		percentiles[50],
		percentiles[90],
		percentiles[100])
}
