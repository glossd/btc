package txutil

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
	"github.com/glossd/btc/netchain"
	"github.com/glossd/btc/wallet"
	"sort"
)

const DefaultMinerFee = 5000
var netParams = &chaincfg.TestNet3Params

type CreateParams struct {
	// WIF-format.
	PrivateKey string
	// Bitcoin address of the receiver.
	Destination        string
	// In satoshi, zero value is rejected. Can be omitted if SendAll is true.
	Amount int64
	// if true, it sends all satoshi.
	SendAll bool
	// In satoshi, defaults to DefaultMinerFee.
	MinerFee int64
	// defaults to netchain.MainNet.
	Net netchain.Net
	// defaults to addressinfo.FetchFromBlockcypher.
	Fetch addressinfo.Fetch

	sourceAddr string
	sourcePayAddr []byte
	destinationPayAddr []byte
}

func Create(params CreateParams) (string, error) {
	params, err := checkCreateParams(params)
	if err != nil {
		return "", err
	}

	btcAddr, err := params.Fetch(params.sourceAddr, params.Net)
	if err != nil {
		return "", err
	}
	if len(btcAddr.UTXOs) == 0 {
		return "", fmt.Errorf("no utxos available on your address")
	}
	if params.Amount + params.MinerFee > btcAddr.Balance {
		return "", fmt.Errorf("not enough satishi, balance=%d, fee+amount=%d", btcAddr.Balance, params.Amount + params.MinerFee)
	}


	var utxosToSpend []addressinfo.UTXO
	var balanceToSpend int64
	if params.SendAll {
		utxosToSpend, balanceToSpend = btcAddr.UTXOs, btcAddr.Balance
	} else {
		utxosToSpend, balanceToSpend = chooseUTXOs(btcAddr.UTXOs, params.Amount+DefaultMinerFee)
	}

	// creating a new bitcoin transaction, different sections of the tx, including
	// input list (contain UTXOs) and outputlist (contain destination address and usually our address)
	// in next steps, sections will be field and pass to sign
	redeemTx := wire.NewMsgTx(wire.TxVersion)

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
	if params.SendAll && params.Amount+params.MinerFee == balanceToSpend {
		redeemTx.AddTxOut(wire.NewTxOut(balanceToSpend-params.MinerFee, params.destinationPayAddr))
	} else {
		redeemTx.AddTxOut(wire.NewTxOut(params.Amount, params.destinationPayAddr))
		redeemTx.AddTxOut(wire.NewTxOut(balanceToSpend-params.Amount-params.MinerFee, params.sourcePayAddr))
	}

	err = SignTx(params.PrivateKey, utxosToSpend, redeemTx)
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

func checkCreateParams(p CreateParams) (CreateParams, error){
	if p.Amount == 0 && !p.SendAll {
		return CreateParams{}, fmt.Errorf("amount of satoshi is not specified")
	}

	destinationPayAddr, err := payAddress(p.Destination)
	if err != nil {
		return CreateParams{}, fmt.Errorf("wrong destination: %v", err)
	}
	p.destinationPayAddr = destinationPayAddr

	if p.MinerFee == 0 {
		p.MinerFee = DefaultMinerFee
	}
	if p.Net == "" {
		p.Net = netchain.MainNet
	}
	if p.Fetch == nil {
		p.Fetch = addressinfo.FetchFromBlockcypher
	}

	sourceAddr, err := wallet.AddressFromPrivateKey(p.PrivateKey, p.Net)
	if err != nil {
		return CreateParams{}, err
	}
	p.sourceAddr = sourceAddr

	sourcePayAddr, err := payAddress(sourceAddr)
	if err != nil {
		return CreateParams{}, err
	}
	p.sourcePayAddr = sourcePayAddr

	return p, nil
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

func payAddress(destination string) ([]byte, error) {
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
		sourcePkString, err := hex.DecodeString(utxoOfIn.Pbscript)
		if err != nil {
			return err
		}
		signature, err := txscript.SignatureScript(redeemTx, i, sourcePkString, txscript.SigHashAll, wif.PrivKey, false)
		if err != nil {
			return err
		}
		in.SignatureScript = signature
	}

	return nil
}