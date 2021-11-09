package addressinfo

import (
	"encoding/json"
	"fmt"
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

func fetchFromBlockchain(address string) ([]UTXO, error) {
	resp, err := http.Get(fmt.Sprintf("https://blockchain.info/unspent?active=%s", address))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data blockchainResponse
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		return nil, err
	}
	result := make([]UTXO, 0, len(data.UnspentOutputs))
	for _, output := range data.UnspentOutputs {
		result = append(result, UTXO{
			TxID:         output.TxID,
			Pbscript:     output.Script,
			Balance:      output.Value,
			SourceOutIdx: uint32(output.TxOutputN),
		})
	}
	return result, nil
}
