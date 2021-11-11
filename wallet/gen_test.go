package wallet

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"log"
	"testing"
)

func TestWalletCompatibility(t *testing.T) {
	realHexPriv := "beb0beef2bd9bbb51ca2282b7a27825aa0cebfeca77820bb1e975170a47af4b3"
	decodeString, err := hex.DecodeString(realHexPriv)
	if err != nil {
		log.Fatal(err)
	}
	priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), decodeString)

	wif, err := btcutil.NewWIF(priv, &chaincfg.TestNet3Params, false)
	if err != nil {
		log.Fatal(err)
	}
	realWif := "932u6Q4xEC9UYRb3rS2BWrSpSPEt5KaU8NNP7EWy7zSkWmfBiGe"
	if wif.String() != realWif {
		log.Fatal("wif not right")
	}



	decodeWIF, err := btcutil.DecodeWIF(realWif)
	if err != nil {
		log.Fatal(err)
	}

	hexPriv := hex.EncodeToString(decodeWIF.PrivKey.Serialize())
	if hexPriv != realHexPriv {
		log.Fatal("wrong hex priv")
	}
}
