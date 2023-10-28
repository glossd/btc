package main

import (
	"fmt"
	"github.com/glossd/btc/netchain"
	"github.com/glossd/btc/wallet"
)

func main() {
	compress := false
	net := netchain.TestNet
	format := wallet.GetFormat(compress)
	hexPriv := "your-key" // e.g. d84b85293493810f912754dd53192dac77c5e2b5e077b3d3233ef5683bda2d82
	wif, address := wallet.GetWallet(hexPriv, net, compress)
	fmt.Printf(
		"Wallet initialized from private key.\n\n"+
			"WIF: %s\n"+
			"Bitcoin address: %s\n"+
			"Private Key: %s\n\n"+
			"WIF and Bitcoin address are in %s format!\n",
		wif, address, hexPriv, format)
}
