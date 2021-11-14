package addressinfo

import (
	"github.com/btcsuite/btcd/wire"
	"github.com/glossd/btc/netchain"
	mathrand "math/rand"
)

func FetchMock(address string, net netchain.Net) (Address, error) {
	var utxoMock = UTXO{
		TxID:     wire.NewMsgTx(wire.TxVersion).TxHash().String(),
		Balance:  mathrand.Int63n(50000000) + 1000000,
		Pbscript: "76a914fee7132bbe9201c4f1a0f846b5f714d9335e263088ac",
		TxOutIdx: 1,
	}
	return Address{Balance: utxoMock.Balance, UTXOs: []UTXO{utxoMock}}, nil
}
