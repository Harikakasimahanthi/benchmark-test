package configs

import (
	"errors"
	"net/url"
	"time"

	"github.com/ssvlabs/ssv-pulse/internal/platform/network"
)

type Metric struct {
	Enabled bool `mapstructure:"enabled"`
}

// Consensus layer (Beacon Node) metrics
type BeaconMetrics struct {
	Client      Metric `mapstructure:"client"`
	Latency     Metric `mapstructure:"latency"`
	Peers       Metric `mapstructure:"peers"`
	Attestation Metric `mapstructure:"attestation"`
	SyncStatus  Metric `mapstructure:"sync_status"`
}

// Execution layer metrics
type ExecutionMetrics struct {
	Peers   Metric `mapstructure:"peers"`
	Latency Metric `mapstructure:"latency"`
}

// Validator client metrics
type ValidatorMetrics struct {
	Proposals    Metric `mapstructure:"proposals"`
	Attestations Metric `mapstructure:"attestations"`
	Duties       Metric `mapstructure:"duties"`
}

// Infrastructure metrics (System Monitoring)
type InfrastructureMetrics struct {
	CPU    Metric `mapstructure:"cpu"`
	Memory Metric `mapstructure:"memory"`
	Disk   Metric `mapstructure:"disk"`
}

type BeaconNode struct {
	Address string        `mapstructure:"address"`
	Metrics BeaconMetrics `mapstructure:"metrics"`
}

func (b BeaconNode) AddrURL() (*url.URL, error) {
	parsedURL, err := url.Parse(b.Address)
	if err != nil {
		return nil, errors.Join(err, errors.New("error parsing Beacon Node address to URL type"))
	}
	return parsedURL, nil
}

type ExecutionNode struct {
	Address string           `mapstructure:"address"`
	Metrics ExecutionMetrics `mapstructure:"metrics"`
}

func (e ExecutionNode) AddrURL() (*url.URL, error) {
	parsedURL, err := url.Parse(e.Address)
	if err != nil {
		return nil, errors.Join(err, errors.New("error parsing Execution Node address to URL type"))
	}
	return parsedURL, nil
}

type ValidatorClient struct {
	Address string           `mapstructure:"address"`
	Metrics ValidatorMetrics `mapstructure:"metrics"`
}

func (v ValidatorClient) AddrURL() (*url.URL, error) {
	parsedURL, err := url.Parse(v.Address)
	if err != nil {
		return nil, errors.Join(err, errors.New("error parsing Validator Client address to URL type"))
	}
	return parsedURL, nil
}

type Infrastructure struct {
	Metrics InfrastructureMetrics `mapstructure:"metrics"`
}

type Server struct {
	Port uint16 `mapstructure:"port"`
}

type Benchmark struct {
	BeaconNode      BeaconNode      `mapstructure:"beacon_node"`
	ExecutionNode   ExecutionNode   `mapstructure:"execution_node"`
	ValidatorClient ValidatorClient `mapstructure:"validator_client"`
	Infrastructure  Infrastructure  `mapstructure:"infrastructure"`
	Server          Server          `mapstructure:"server"`
	Duration        time.Duration   `mapstructure:"duration"`
	Network         string          `mapstructure:"network"`
}

func (b *Benchmark) Validate() (bool, error) {
	// Validate beacon node if relevant metrics are enabled
	if b.BeaconNode.Metrics.Peers.Enabled ||
		b.BeaconNode.Metrics.Attestation.Enabled ||
		b.BeaconNode.Metrics.Client.Enabled ||
		b.BeaconNode.Metrics.Latency.Enabled ||
		b.BeaconNode.Metrics.SyncStatus.Enabled {
		url, err := sanitizeURL(b.BeaconNode.Address)
		if err != nil {
			return false, errors.Join(err, errors.New("beacon node address was not a valid URL"))
		}
		b.BeaconNode.Address = url
	}

	// Validate execution node if relevant metrics are enabled
	if b.ExecutionNode.Metrics.Peers.Enabled ||
		b.ExecutionNode.Metrics.Latency.Enabled {
		url, err := sanitizeURL(b.ExecutionNode.Address)
		if err != nil {
			return false, errors.Join(err, errors.New("execution node address was not a valid URL"))
		}
		b.ExecutionNode.Address = url
	}

	// Validate validator client if relevant metrics are enabled
	if b.ValidatorClient.Metrics.Proposals.Enabled ||
		b.ValidatorClient.Metrics.Attestations.Enabled ||
		b.ValidatorClient.Metrics.Duties.Enabled {
		url, err := sanitizeURL(b.ValidatorClient.Address)
		if err != nil {
			return false, errors.Join(err, errors.New("validator client address was not a valid URL"))
		}
		b.ValidatorClient.Address = url
	}

	// Validate network name
	network := network.Name(b.Network)
	if err := network.Validate(); err != nil {
		return false, errors.Join(err, errors.New("network name was not valid"))
	}

	return true, nil
}
