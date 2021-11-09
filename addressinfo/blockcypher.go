package addressinfo

import (
	"github.com/blockcypher/gobcy"
	"os"
)

func fetchFromBlockcypher(address string) ([]UTXO, error) {
	btc := gobcy.API{Token: os.Getenv("BTC_API_KEY"), Coin: "btc", Chain: "test3"}
	info, err := btc.GetAddrFull(address, map[string]string{"unspentOnly":"true"})
	if err != nil {
		return nil, err
	}
	var result []UTXO
	for _, tx := range info.TXs {
		for outputIdx, output := range tx.Outputs {
			if len(output.Addresses) == 1 && output.Addresses[0] == address {
				if output.SpentBy == "" {
					result = append(result, UTXO{TxID: tx.Hash, Balance: output.Value.Int64(), Pbscript: output.Script, SourceOutIdx: uint32(outputIdx)})
				}
			}
		}
	}
	return result, nil
}


