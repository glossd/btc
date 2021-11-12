package main

import (
	"fmt"
	"github.com/glossd/btc/netchain"
	"github.com/glossd/btc/wallet"
)

func main() {
	privateKey, address := wallet.New(netchain.TestNet)
	fmt.Printf("Private Key: %s\nBitcoin address: %s\n", privateKey, address)
}
