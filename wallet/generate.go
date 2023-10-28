package wallet

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil"
	"github.com/glossd/btc/netchain"
	"log"
)

func New(net netchain.Net, compress bool) (wifString string, bitcoinAddress string, privateKeyHex string) {
	// errors shouldn't happen
	priv, err := btcec.NewPrivateKey(btcec.S256())
	check(err)
	wif, err := btcutil.NewWIF(priv, net.GetBtcdNetParams(), compress)
	check(err)
	addr := GetBitcoinAddress(priv, net, compress)
	return wif.String(), addr.EncodeAddress(), hex.EncodeToString(priv.Serialize())
}

func GetWallet(hexPriv string, net netchain.Net, compress bool) (wifString string, bitcoinAddress string) {
	privBytes, err := hex.DecodeString(hexPriv)
	check(err)
	priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), privBytes)
	wif, err := btcutil.NewWIF(priv, net.GetBtcdNetParams(), compress)
	check(err)
	addr := GetBitcoinAddress(priv, net, compress)
	return wif.String(), addr.EncodeAddress()
}

func GetBitcoinAddress(priv *btcec.PrivateKey, net netchain.Net, compress bool) *btcutil.AddressPubKey {
	var addr *btcutil.AddressPubKey
	var err error
	if compress {
		addr, err = btcutil.NewAddressPubKey(priv.PubKey().SerializeCompressed(), net.GetBtcdNetParams())
	} else {
		addr, err = btcutil.NewAddressPubKey(priv.PubKey().SerializeUncompressed(), net.GetBtcdNetParams())
	}
	check(err)
	return addr
}

func NewAddress(net netchain.Net, compress bool) string {
	_, address, _ := New(net, compress)
	return address
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
