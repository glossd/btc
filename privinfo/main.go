package main

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"log"
	"os"
)

func main() {
	wifStr := os.Args[1]
	wif, err := btcutil.DecodeWIF(wifStr)
	if err != nil {
		log.Fatalf("wif decode: %v", err)
	}
	addrPubKey, err := btcutil.NewAddressPubKey(wif.PrivKey.PubKey().SerializeUncompressed(), &chaincfg.TestNet3Params)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(addrPubKey.EncodeAddress())
}
