package addressinfo

import (
	"encoding/json"
	"fmt"
	"github.com/glossd/btc/netchain"
	"io/ioutil"
	"net/http"
)

type blockchainResponse struct {
	UnspentOutputs []blockchainUTXO `json:"unspent_outputs"`
}

type blockchainUTXO struct {
	TxID string `json:"tx_hash_big_endian"`
	TxOutputN int `json:"tx_output_n"`
	Script string `json:"script"`
	Value int64 `json:"value"`
}

func FetchFromBlockchain(address string, net netchain.Net) (Address, error) {
	if net != netchain.MainNet {
		return Address{}, fmt.Errorf("only mainnet is supported fetching UTXOs from blockchain.info")
	}
	resp, err := http.Get(fmt.Sprintf("https://blockchain.info/unspent?active=%s", address))
	if err != nil {
		return Address{}, err
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Address{}, err
	}

	var data blockchainResponse
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		return Address{}, err
	}
	utxos := make([]UTXO, 0, len(data.UnspentOutputs))
	var balance int64
	for _, output := range data.UnspentOutputs {
		utxos = append(utxos, UTXO{
			TxID:     output.TxID,
			Pbscript: output.Script,
			Balance:  output.Value,
			TxOutIdx: uint32(output.TxOutputN),
		})
		balance += output.Value
	}
	return Address{UTXOs: utxos, Balance: balance}, nil
}
