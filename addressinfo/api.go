package addressinfo

import "github.com/glossd/btc/netchain"

var utxoMock = UTXO{
	TxID:     "5dcb1e3a6fcd02e9e216113deda9a914b2b2191a0bb42383817b737c4c3280e2",
	Balance:  0.02237644 * 1e8,
	Pbscript: "76a914fee7132bbe9201c4f1a0f846b5f714d9335e263088ac",
	TxOutIdx: 1,
}

type Address struct {
	Balance int64
	UTXOs []UTXO
}

type UTXO struct {
	TxID     string
	Pbscript string
	Balance  int64
	TxOutIdx uint32
}

type Fetch func(address string, net netchain.Net) (Address, error)