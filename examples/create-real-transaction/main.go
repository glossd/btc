package main

import (
	"fmt"
	"github.com/glossd/btc/netchain"
	"github.com/glossd/btc/txutil"
	"os"
)

func main() {
	os.Setenv("BTC_API_KEY", "you_token") // set your Blockcypher Token e.g. 40f3102a0bbf409d1642a0d4ba31d3df
	rawTx, err := txutil.Create(txutil.CreateParams{
		PrivateKey:  "your key", // e.g. 932u6Q4xEC9UYRb3rS2BWrSpSPEt5KaU8NNP7EWy7zSkWmfBiGe
		Destination: "address",  // e.g. n4kkk9H2jGj7t8LA4vxK4DHM7Lq95VaEXC
		Amount:      150000,     // satoshi to send
		Net:         netchain.MainNet,
		MinerFee:    4000, // set good miner fee for the transaction to be picked up
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(rawTx) // copy the output and broadcast it anywhere e.g. https://www.blockchain.com/explorer/assets/btc/broadcast-transaction
}
