package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/glossd/btc/addressinfo"
	"log"
	"sort"
)

const minerFee = 5000
var netParams = &chaincfg.TestNet3Params

func main()  {
	rawTx, err := CreateTx("932u6Q4xEC9UYRb3rS2BWrSpSPEt5KaU8NNP7EWy7zSkWmfBiGe",
		"n4kkk9H2jGj7t8LA4vxK4DHM7Lq95VaEXC", 0.01 * 1e8)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("raw signed transaction is: ", rawTx)
}



func NewTx() (*wire.MsgTx, error) {
	return wire.NewMsgTx(wire.TxVersion), nil
}

func CreateTx(privKey string, destination string, amount int64) (string, error) {

	wif, err := btcutil.DecodeWIF(privKey)
	if err != nil {
		return "", err
	}

	addrPubKey, err := btcutil.NewAddressPubKey(wif.PrivKey.PubKey().SerializeUncompressed(), netParams)
	if err != nil {
		return "", err
	}

	utxos, err := addressinfo.FetchUTXOs(addrPubKey.EncodeAddress(), netParams.Name)
	if err != nil {
		return "", err
	}
	if len(utxos) == 0 {
		return "", fmt.Errorf("no utxos available on your address")
	}

	utxosToSpend, balanceToSpend := chooseUTXOs(utxos, amount+minerFee)

	// creating a new bitcoin transaction, different sections of the tx, including
	// input list (contain UTXOs) and outputlist (contain destination address and usually our address)
	// in next steps, sections will be field and pass to sign
	redeemTx, err := NewTx()
	if err != nil {
		return "", err
	}

	for _, u := range utxosToSpend {
		utxoHash, err := chainhash.NewHashFromStr(u.TxID)
		if err != nil {
			return "", err
		}
		outPoint := wire.NewOutPoint(utxoHash, u.SourceOutIdx)
		txIn := wire.NewTxIn(outPoint, nil, nil)
		redeemTx.AddTxIn(txIn)
	}

	// adding the destination address and the amount to
	// the transaction as output
	if amount+minerFee >= balanceToSpend {
		redeemTx.AddTxOut(wire.NewTxOut(balanceToSpend-minerFee, outAddressFatal(destination)))
	} else {
		redeemTx.AddTxOut(wire.NewTxOut(amount, outAddressFatal(destination)))
		redeemTx.AddTxOut(wire.NewTxOut(balanceToSpend-amount-minerFee, outAddressFatal(addressFromPrivateKeyFatal(privKey))))
	}

	err = SignTx(privKey, utxosToSpend, redeemTx)
	if err != nil {
		return "", err
	}

	var signedTx bytes.Buffer
	err = redeemTx.Serialize(&signedTx)
	if err != nil {
		return "", err
	}

	hexSignedTx := hex.EncodeToString(signedTx.Bytes())

	return hexSignedTx, nil
}

func chooseUTXOs(utxos []addressinfo.UTXO, amountToSend int64) (toSpend []addressinfo.UTXO, balance int64) {
	sort.Slice(utxos, func(i, j int) bool {
		return utxos[i].Balance < utxos[j].Balance
	})

	var accBalance int64
	for i, u := range utxos {
		accBalance += u.Balance
		if accBalance > amountToSend {
			return utxos[:i+1], accBalance
		}
	}

	panic("address doesn't have enough bitcoins for transfer")
}

func outAddressFatal(destination string) []byte {
	destinationAddrByte, err := outAddress(destination)
	if err != nil {
		log.Fatalf("wrong out address: %v", err)
	}
	return destinationAddrByte
}

func outAddress(destination string) ([]byte, error) {
	// extracting destination address as []byte from function argument (destination string)
	destinationAddr, err := btcutil.DecodeAddress(destination, netParams)
	if err != nil {
		return nil, err
	}

	destinationAddrByte, err := txscript.PayToAddrScript(destinationAddr)
	if err != nil {
		return nil, err
	}
	return destinationAddrByte, nil
}

func addressFromPrivateKeyFatal(privKey string) string {
	address, err := AddressFromPrivateKey(privKey)
	if err != nil {
		log.Fatal(err)
	}
	return address
}

func AddressFromPrivateKey(privKey string) (string, error) {
	wif, err := btcutil.DecodeWIF(privKey)
	if err != nil {
		return "", fmt.Errorf("couldn't decode private key")
	}
	addr, err := btcutil.NewAddressPubKey(wif.PrivKey.PubKey().SerializeUncompressed(), netParams)
	if err != nil {
		return "", fmt.Errorf("couldn't extract address from private key")
	}
	return addr.EncodeAddress(), nil
}

func SignTx(privKey string, utxoToSpend []addressinfo.UTXO, redeemTx *wire.MsgTx) error {

	wif, err := btcutil.DecodeWIF(privKey)
	if err != nil {
		return err
	}

	utxoToSpendMap := make(map[string]addressinfo.UTXO)
	for _, u := range utxoToSpend {
		h, err := chainhash.NewHashFromStr(u.TxID)
		if err != nil {
			return fmt.Errorf("signing transaction failed, could compute hash utxo=%v", u)
		}
		utxoToSpendMap[h.String()] = u
	}

	for i, in := range redeemTx.TxIn {
		utxoOfIn := utxoToSpendMap[in.PreviousOutPoint.Hash.String()]
		// since there is only one input in our transaction
		// we use 0 as second argument, if the transaction
		// has more args, should pass related index
		sourcePkString, err := hex.DecodeString(utxoOfIn.Pbscript)
		if err != nil {
			return err
		}
		signature, err := txscript.SignatureScript(redeemTx, i, sourcePkString, txscript.SigHashAll, wif.PrivKey, false)
		if err != nil {
			return err
		}

		// since there is only one input, and want to add
		// signature to it use 0 as index
		in.SignatureScript = signature
	}

	return nil
}