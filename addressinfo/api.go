package addressinfo

import "github.com/glossd/btc/netchain"

type Address struct {
	Balance int64
	UTXOs   []UTXO
}

type UTXO struct {
	TxID     string
	Pbscript string
	Balance  int64
	TxOutIdx int
}

type Fetch func(address string, net netchain.Net) (Address, error)

// GetSatoshiPerByte returns minimum 'good-enough' satoshi per byte rate.
type GetSatoshiPerByte func(net netchain.Net) (int, error)
