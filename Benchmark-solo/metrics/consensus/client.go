package consensus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ssvlabs/ssv-pulse/internal/platform/logger"
	"github.com/ssvlabs/ssv-pulse/internal/platform/metric"
)

const (
	VersionMeasurement = "Version"
)

type ClientMetric struct {
	metric.Base[string]
	url string
}

func NewClientMetric(url, name string, healthCondition []metric.HealthCondition[string]) *ClientMetric {
	return &ClientMetric{
		url: url,
		Base: metric.Base[string]{
			HealthConditions: healthCondition,
			Name:             name,
		},
	}
}

func (c *ClientMetric) Measure(ctx context.Context) {
	var (
		resp struct {
			Data struct {
				Version string `json:"version"`
			} `json:"data"`
		}
	)
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

func (c *ClientMetric) AggregateResults() string {
	if len(c.DataPoints) != 0 {
		return c.DataPoints[0].Values[VersionMeasurement]
	}
	return ""
}
