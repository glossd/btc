package netchain

import (
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
)

type Net string

const MainNet Net = "mainnet"
const TestNet Net = "testnet3"

func (n Net) GetBtcdNetParams() *chaincfg.Params {
	switch n {
	case MainNet: return &chaincfg.MainNetParams
	case TestNet: return &chaincfg.TestNet3Params
	default: panic(fmt.Sprintf("net chain '%s' is not supported", n))
	}
}

func (n Net) GetBlockcypherChain() string {
	switch n {
	case MainNet: return "main"
	case TestNet: return "test3"
	default: panic(fmt.Sprintf("net chain '%s' is not supported", n))
	}
}

func (n Net) String() string {
	return string(n)
}
