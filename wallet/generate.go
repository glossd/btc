package wallet

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil"
	"github.com/glossd/btc/netchain"
	"log"
)

func New(net netchain.Net) (privateKeyWif, bitcoinAddress string) {
	// errors shouldn't happen
	priv, err := btcec.NewPrivateKey(btcec.S256())
	check(err)
	wif, err := btcutil.NewWIF(priv, net.GetBtcdNetParams(), false)
	check(err)
	addr, err := btcutil.NewAddressPubKey(priv.PubKey().SerializeUncompressed(), net.GetBtcdNetParams())
	check(err)
	return wif.String(), addr.EncodeAddress()
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
