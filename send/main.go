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
	"log"
)

const minerFee = 5000
var netParams = &chaincfg.TestNet3Params
var theUTXO = UTXO{
	TxID: "5dcb1e3a6fcd02e9e216113deda9a914b2b2191a0bb42383817b737c4c3280e2",
	Balance: 0.02237644 * 1e8,
	Pbscript: "76a914fee7132bbe9201c4f1a0f846b5f714d9335e263088ac",
	SourceOutIdx: 1,
}

func main()  {
	rawTx, err := CreateTx("91fPLgXt3tJPZGyDSLEFnD4btsZ9UZ86ibUtShVPsPMJxP15qJP", "n4kkk9H2jGj7t8LA4vxK4DHM7Lq95VaEXC",
		"mgFv6afUVhrdd3D6mY2iyWzHVk5b64qTok", 0.005 * 1e8)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("raw signed transaction is: ", rawTx)
}



func NewTx() (*wire.MsgTx, error) {
	return wire.NewMsgTx(wire.TxVersion), nil
}

type UTXO struct {
	TxID string
	Pbscript string
	Balance int64
	SourceOutIdx uint32
}

func FetchLatestUTXO(address string) (UTXO, error) {

	// Provide your url to get UTXOs, read the response
	// unmarshal it, and extract necessary data
	// newURL := fmt.Sprintf("https://your.favorite.block-explorer/%s", address)

	//response, err := http.Get(newURL)
	//if err != nil {
	// fmt.Println("error in FetchLatestUTXO, http.Get")
	// return nil, 0, "", err
	//}
	//defer response.Body.Close()
	//body, err := ioutil.ReadAll(response.Body)

	// based on the response you get, should define a struct
	// so before unmarshaling check your JSON response model

	//var blockChairResp = model.BlockChairResp{}
	//err = json.Unmarshal(body, &blockChairResp)
	//if err != nil {
	// fmt.Println("error in FetchLatestUTXO, json.Unmarshal")
	// return  nil, 0, "", err
	//}

	return theUTXO, nil
}

func CreateTx(privKey string, source string, destination string, amount int64) (string, error) {

	wif, err := btcutil.DecodeWIF(privKey)
	if err != nil {
		return "", err
	}

	addrPubKey, err := btcutil.NewAddressPubKey(wif.PrivKey.PubKey().SerializeUncompressed(), netParams)
	if err != nil {
		return "", err
	}

	utxo, err := FetchLatestUTXO(addrPubKey.EncodeAddress())
	if err != nil {
		return "", err
	}


	/*
	 * 1 or unit-amount in Bitcoin is equal to 1 satoshi and 1 Bitcoin = 100000000 satoshi
	 */

	// checking for sufficiency of account
	if utxo.Balance < amount {
		return "", fmt.Errorf("the balance of the account is not sufficient")
	}

	// creating a new bitcoin transaction, different sections of the tx, including
	// input list (contain UTXOs) and outputlist (contain destination address and usually our address)
	// in next steps, sections will be field and pass to sign
	redeemTx, err := NewTx()
	if err != nil {
		return "", err
	}

	utxoHash, err := chainhash.NewHashFromStr(utxo.TxID)
	if err != nil {
		return "", err
	}

	// the second argument is vout or Tx-index, which is the index
	// of spending UTXO in the transaction that Txid referred to
	// in this case is 1, but can vary different numbers
	outPoint := wire.NewOutPoint(utxoHash, utxo.SourceOutIdx)

	// making the input, and adding it to transaction
	txIn := wire.NewTxIn(outPoint, nil, nil)
	redeemTx.AddTxIn(txIn)

	// adding the destination address and the amount to
	// the transaction as output
	if amount+minerFee >= utxo.Balance {
		redeemTx.AddTxOut(wire.NewTxOut(utxo.Balance-minerFee, outAddressFatal(destination)))
	} else {
		redeemTx.AddTxOut(wire.NewTxOut(amount, outAddressFatal(destination)))
		redeemTx.AddTxOut(wire.NewTxOut(utxo.Balance-amount-minerFee, outAddressFatal(source)))
	}

	// now sign the transaction
	err = SignTx(privKey, utxo.Pbscript, redeemTx)
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

func SignTx(privKey string, pkScript string, redeemTx *wire.MsgTx) error {

	wif, err := btcutil.DecodeWIF(privKey)
	if err != nil {
		return err
	}

	sourcePKScript, err := hex.DecodeString(pkScript)
	if err != nil {
		return err
	}

	// since there is only one input in our transaction
	// we use 0 as second argument, if the transaction
	// has more args, should pass related index
	signature, err := txscript.SignatureScript(redeemTx, 0, sourcePKScript, txscript.SigHashAll, wif.PrivKey, false)
	if err != nil {
		return err
	}

	// since there is only one input, and want to add
	// signature to it use 0 as index
	redeemTx.TxIn[0].SignatureScript = signature
	return nil
}