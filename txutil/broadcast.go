package txutil

import (
	"github.com/blockcypher/gobcy"
	"github.com/glossd/btc/netchain"
	"os"
)

// Returns the hash of the broadcasted transaction.
func Broadcast(rawTx string, net netchain.Net) (string, error) {
	btc := gobcy.API{Token: os.Getenv("BTC_API_KEY"), Coin: "btc", Chain: net.GetBlockcypherChain()}
	tx, err := btc.PushTX(rawTx)
	if err != nil {
		return "", err
	}
	return tx.Trans.Hash, nil
}
