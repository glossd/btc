package txutil

import (
	"bytes"
	"encoding/hex"
	"fmt"
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

type CreateParams struct {
	// WIF-format. Will be omitted if PrivateKeys are specified.
	PrivateKey string
	// If Amount is specified the remainder will be sent to the first private key.
	PrivateKeys []string
	// Bitcoin address of the receiver.
	Destination        string
	// Measured in satoshi. Will be omitted if SendAll is true.
	Amount int64
	// if true, all satoshi will be sent.
	SendAll bool
	// In satoshi, defaults to DefaultMinerFee.
	MinerFee int64
	// defaults to netchain.MainNet.
	Net netchain.Net
	// defaults to addressinfo.FetchFromBlockcypher.
	Fetch addressinfo.Fetch

	pkInfos []privateKeyInfo
	destinationPayAddr []byte
}

func (cp CreateParams) FullCost() int64 {
	return cp.Amount + cp.MinerFee
}

func Create(params CreateParams) (string, error) {
	params, err := checkCreateParams(params)
	if err != nil {
		return "", err
	}

	addrs, err := getAddressesToWithdrawFrom(params)
	if err != nil {
		return "", err
	}

	tx := wire.NewMsgTx(wire.TxVersion)

	satoshiRemainder, err := addUTXOsToTxInputs(tx, addrs, params)
	if err != nil {
		return "", err
	}

	addTxOutputs(tx, params, satoshiRemainder, addrs)

	err = signTx(tx, addrs)
	if err != nil {
		return "", err
	}

	return hexEncodeTx(tx)
}

func checkCreateParams(p CreateParams) (CreateParams, error){
	if p.Amount == 0 && !p.SendAll {
		return CreateParams{}, fmt.Errorf("amount of satoshi is not specified")
	}

	if p.MinerFee == 0 {
		p.MinerFee = DefaultMinerFee
	}
	if p.Net == "" {
		p.Net = netchain.MainNet
	}
	if p.Fetch == nil {
		p.Fetch = addressinfo.FetchFromBlockcypher
	}

	destinationPayAddr, err := toPayAddress(p.Destination, p.Net)
	if err != nil {
		return CreateParams{}, fmt.Errorf("wrong destination: %v", err)
	}
	p.destinationPayAddr = destinationPayAddr

	if len(p.PrivateKeys) > 0 {
		for _, key := range p.PrivateKeys {
			pkInfo, err := toPkInfo(key, p.Net)
			if err != nil {
				return CreateParams{}, fmt.Errorf("one of the private keys is malformed: %s", err)
			}
			p.pkInfos = append(p.pkInfos, pkInfo)
		}
	} else if p.PrivateKey != "" {
		pkInfo, err := toPkInfo(p.PrivateKey, p.Net)
		if err != nil {
			return CreateParams{}, err
		}
		p.pkInfos = []privateKeyInfo{pkInfo}
	} else {
		return CreateParams{}, fmt.Errorf("must specify either PrivateKey or PrivateKeys: %v", err)
	}

	return p, nil
}

type address struct {
	addressinfo.Address
	privateKey string
}

func getAddressesToWithdrawFrom(params CreateParams) ([]address, error){
	var addrsToWithdrawFrom []address
	var satoshiSum int64
	for _, pkInfo := range params.pkInfos {
		addr, err := params.Fetch(pkInfo.address, params.Net)
		if err != nil {
			return nil, err
		}
		addrsToWithdrawFrom = append(addrsToWithdrawFrom, address{Address: addr, privateKey: pkInfo.key})
		satoshiSum += addr.Balance
		if !params.SendAll && satoshiSum > params.FullCost() {
			return addrsToWithdrawFrom, nil
		}
	}
	if params.SendAll {
		return addrsToWithdrawFrom, nil
	} else {
		return nil, fmt.Errorf("not enough satoshi to send, amount+fee=%d, balance=%d", params.FullCost(), satoshiSum)
	}
}

func addUTXOsToTxInputs(tx *wire.MsgTx, addrs []address, params CreateParams,) (satoshiRemainder int64, err error) {
	amountLeftToRedeem := params.FullCost()
	for i, addr := range addrs {
		isLastAddr := i == len(addrs)-1
		if isLastAddr && !params.SendAll {
			lastUTXOs, theirBalance := chooseUTXOs(addr.UTXOs, amountLeftToRedeem)
			satoshiRemainder = theirBalance - amountLeftToRedeem
			err := addInputs(tx, lastUTXOs)
			return satoshiRemainder, err
		}
		err := addInputs(tx, addr.UTXOs)
		if err != nil {
			return 0, err
		}
		amountLeftToRedeem -= addr.Balance
	}
	return satoshiRemainder, nil
}

func addInputs(tx *wire.MsgTx, utxos []addressinfo.UTXO) error {
	for _, utxo := range utxos {
		utxoHash, err := chainhash.NewHashFromStr(utxo.TxID)
		if err != nil {
			return err
		}
		outPoint := wire.NewOutPoint(utxoHash, utxo.TxOutIdx)
		txIn := wire.NewTxIn(outPoint, nil, nil)
		tx.AddTxIn(txIn)
	}
	return nil
}

func addTxOutputs(tx *wire.MsgTx, params CreateParams, satoshiRemainder int64, addrs []address) {
	if params.SendAll || satoshiRemainder == 0 {
		fullBalance := calcBalanceOfAddresses(addrs)
		tx.AddTxOut(wire.NewTxOut(fullBalance-params.MinerFee, params.destinationPayAddr))
	} else {
		tx.AddTxOut(wire.NewTxOut(params.Amount, params.destinationPayAddr))
		tx.AddTxOut(wire.NewTxOut(satoshiRemainder, params.pkInfos[0].payAddress))
	}
}

type privateKeyInfo struct {
	key string
	address string
	payAddress []byte
}

func toPkInfo(privKey string, net netchain.Net) (privateKeyInfo, error) {
	addr, err := wallet.AddressFromPrivateKey(privKey, net)
	if err != nil {
		return privateKeyInfo{}, err
	}
	payAddress, err := toPayAddress(addr, net)
	if err != nil {
		return privateKeyInfo{}, err
	}
	return privateKeyInfo{key: privKey, address: addr, payAddress: payAddress}, nil
}

func calcBalanceOfAddresses(addresses []address) (balance int64) {
	for _, a := range addresses {
		balance += a.Balance
	}
	return
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

func toPayAddress(address string, net netchain.Net) ([]byte, error) {
	// extracting address as []byte from function argument
	destinationAddr, err := btcutil.DecodeAddress(address, net.GetBtcdNetParams())
	if err != nil {
		return nil, err
	}

	destinationAddrByte, err := txscript.PayToAddrScript(destinationAddr)
	if err != nil {
		return nil, err
	}
	return destinationAddrByte, nil
}

func signTx(tx *wire.MsgTx, addresses []address) error {

	type utxoWithKey struct {
		addressinfo.UTXO
		wif *btcutil.WIF
	}

	utxosToSpendMap := make(map[string]utxoWithKey)
	for _, a := range addresses {
		wif, err := btcutil.DecodeWIF(a.privateKey)
		if err != nil {
			return err
		}
		for _, u := range a.UTXOs {
			h, err := chainhash.NewHashFromStr(u.TxID)
			if err != nil {
				return fmt.Errorf("signing transaction failed, could compute hash utxo=%v", u)
			}
			utxosToSpendMap[h.String()] = utxoWithKey{UTXO: u, wif: wif}
		}
	}

	for i, in := range tx.TxIn {
		utxoOfIn := utxosToSpendMap[in.PreviousOutPoint.Hash.String()]
		sourcePkString, err := hex.DecodeString(utxoOfIn.Pbscript)
		if err != nil {
			return err
		}
		signature, err := txscript.SignatureScript(tx, i, sourcePkString, txscript.SigHashAll, utxoOfIn.wif.PrivKey, false)
		if err != nil {
			return err
		}
		in.SignatureScript = signature
	}

	return nil
}

func hexEncodeTx(tx *wire.MsgTx) (string, error) {
	var txBytes bytes.Buffer
	err := tx.Serialize(&txBytes)
	if err != nil {
		return "", err
	}

	hexSignedTx := hex.EncodeToString(txBytes.Bytes())
	return hexSignedTx, nil
}

func hexDecodeTx(rawTx string) (*wire.MsgTx, error) {
	txBytes, err := hex.DecodeString(rawTx)
	if err != nil {
		return nil, err
	}
	tx := wire.NewMsgTx(wire.TxVersion)
	err = tx.Deserialize(bytes.NewReader(txBytes))
	if err != nil {
		return nil, err
	}
	return tx, nil
}
