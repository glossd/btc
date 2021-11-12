package addressinfo

import (
	"github.com/blockcypher/gobcy"
	"github.com/glossd/btc/netchain"
	"os"
)

func FetchFromBlockcypher(address string, net netchain.Net) (Address, error) {
	btc := gobcy.API{Token: os.Getenv("BTC_API_KEY"), Coin: "btc", Chain: net.GetBlockcypherChain()}
	info, err := btc.GetAddrFull(address, map[string]string{})
	if err != nil {
		return Address{}, err
	}
	var utxos []UTXO
	for _, tx := range info.TXs {
		for outputIdx, output := range tx.Outputs {
			if len(output.Addresses) == 1 && output.Addresses[0] == address {
				if output.SpentBy == "" {
					utxos = append(utxos, UTXO{TxID: tx.Hash, Balance: output.Value.Int64(), Pbscript: output.Script, TxOutIdx: uint32(outputIdx)})
				}
			}
		}
	}

	return Address{UTXOs: utxos, Balance: info.Balance.Int64()}, nil
}


