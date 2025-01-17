package network

import (
	"fmt"
	"strings"
	"time"
)

var (
	GenesisTime = map[Name]time.Time{
		Holesky: time.Unix(1695902400, 0),
		Mainnet: time.Unix(1606824023, 0),
	}
)

type Name string

const (
	Holesky Name = "holesky"
	Mainnet Name = "mainnet"
)

func (n Name) Validate() error {
	if !strings.EqualFold(string(n), string(Holesky)) && !strings.EqualFold(string(n), string(Mainnet)) {
		return fmt.Errorf("network name should be either '%s' or '%s'", Holesky, Mainnet)
	}
	return nil
}
