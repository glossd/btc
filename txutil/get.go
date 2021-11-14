package txutil

import (
	"github.com/blockcypher/gobcy"
	"github.com/glossd/btc/netchain"
	"os"
)

func GetConfirmations(txID string, net netchain.Net) (int, error) {
	btc := gobcy.API{Token: os.Getenv("BTC_API_KEY"), Coin: "btc", Chain: net.GetBlockcypherChain()}
	tx, err := btc.GetTX(txID, map[string]string{"limit":"1"})
	if err != nil {
		return 0, err
	}
	return tx.Confirmations, nil
}
