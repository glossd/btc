package wallet

import (
	"fmt"
	"github.com/btcsuite/btcutil"
	"github.com/glossd/btc/netchain"
)

func AddressFromPrivateKey(privKey string, net netchain.Net) (string, error) {
	wif, err := btcutil.DecodeWIF(privKey)
	if err != nil {
		return "", fmt.Errorf("couldn't decode private key")
	}
	addr, err := btcutil.NewAddressPubKey(wif.PrivKey.PubKey().SerializeUncompressed(), net.GetBtcdNetParams())
	if err != nil {
		return "", fmt.Errorf("couldn't extract address from private key")
	}
	return addr.EncodeAddress(), nil
}
