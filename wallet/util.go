package wallet

import (
	"fmt"
	"github.com/btcsuite/btcutil"
	"github.com/glossd/btc/netchain"
)

func AddressFromWif(wifString string, net netchain.Net, compress bool) (string, error) {
	wif, err := btcutil.DecodeWIF(wifString)
	if err != nil {
		return "", fmt.Errorf("couldn't decode WIF")
	}
	addr := GetBitcoinAddress(wif.PrivKey, net, compress)
	return addr.EncodeAddress(), nil
}

func IsAddressValid(address string, net netchain.Net) bool {
	_, err := btcutil.DecodeAddress(address, net.GetBtcdNetParams())
	return err == nil
}

func GetFormat(compress bool) string {
	format := "uncompressed"
	if compress {
		format = "compressed"
	}
	return format
}
