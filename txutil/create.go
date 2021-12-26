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
	"strconv"
)

const defaultMinerFee = 5000
// https://support.blockchain.com/hc/en-us/articles/210354003-What-is-the-minimum-amount-I-can-send-
const minSatoshiToSend = 546

type CreateParams struct {
	// WIF-format. Will be omitted if PrivateKeys are specified.
	PrivateKey string
	// Iteratively includes each key in transaction until the full amount can be transferred.
	// The remainder of bitcoins will be sent to the first private key.
	PrivateKeys []string
	// Bitcoin address of the receiver. Amount or SendAll must be set. Will be omitted if Destinations are specified.
	Destination string
	// Parameter for Destination. Measured in satoshi. Will be omitted if SendAll is true.
	Amount int64
	Destinations []Destination
	// If true, all satoshi will be sent. Only works if you specified only one destination.
	SendAll bool
	// In satoshi, defaults to defaultMinerFee.
	MinerFee int64
	// defaults to netchain.MainNet.
	Net netchain.Net
	// defaults to addressinfo.FetchFromBlockcypher.
	Fetch addressinfo.Fetch

	pkInfos   []privateKeyInfo
	destInfos []destinationInfo
}

type Destination struct {
	// Bitcoin address of one of the receivers.
	Address string
	// Measured in satoshi.
	Amount int64
}

func (cp CreateParams) fullCost() int64 {
	return cp.fullAmount() + cp.MinerFee
}

func (cp CreateParams) fullAmount() int64 {
	var result int64
	for _, info := range cp.destInfos {
		result += info.Amount
	}
	return result
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

func checkCreateParams(p CreateParams) (CreateParams, error) {
	if p.MinerFee == 0 {
		p.MinerFee = defaultMinerFee
	}
	if p.Net == "" {
		p.Net = netchain.MainNet
	}
	if p.Fetch == nil {
		p.Fetch = addressinfo.FetchFromBlockcypher
	}

	if len(p.Destinations) == 0 {
		if p.Destination == "" {
			return CreateParams{}, fmt.Errorf("destination must be specified")
		}
		if p.Amount < minSatoshiToSend && !p.SendAll {
			return CreateParams{}, fmt.Errorf("amount of satoshi can't be less than %d", minSatoshiToSend)
		}
		payAddress, err := toPayAddress(p.Destination, p.Net)
		if err != nil {
			return CreateParams{}, err
		}
		p.destInfos = []destinationInfo{{
			Destination: Destination{Address: p.Destination, Amount: p.Amount},
			payAddress:  payAddress,
		}}
	} else {
		if len(p.Destinations) > 1 && p.SendAll {
			return CreateParams{}, fmt.Errorf("SendAll works with only one destination")
		}
		var fullAmount int64
		var dInfos []destinationInfo
		for _, d := range p.Destinations {
			fullAmount += d.Amount
			info, err := toDestInfo(d, p.Net)
			if err != nil {
				return CreateParams{}, err
			}
			dInfos = append(dInfos, info)
		}
		sendAllToOneDest := len(dInfos) == 1 && p.SendAll
		if fullAmount < minSatoshiToSend && !sendAllToOneDest {
			return CreateParams{}, fmt.Errorf("full amount of satoshi can't be less than %d", minSatoshiToSend)
		}
		p.destInfos = dInfos
	}

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
		return CreateParams{}, fmt.Errorf("must specify either PrivateKey or PrivateKeys")
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
		if !params.SendAll && satoshiSum >= params.fullCost() {
			return addrsToWithdrawFrom, nil
		}
	}
	if params.SendAll {
		return addrsToWithdrawFrom, nil
	} else {
		return nil, fmt.Errorf("not enough satoshi to send, amount+fee=%d, balance=%d", params.fullCost(), satoshiSum)
	}
}

func addUTXOsToTxInputs(tx *wire.MsgTx, addrs []address, params CreateParams,) (satoshiRemainder int64, err error) {
	amountLeftToRedeem := params.fullCost()
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
		outPoint := wire.NewOutPoint(utxoHash, uint32(utxo.TxOutIdx))
		txIn := wire.NewTxIn(outPoint, nil, nil)
		tx.AddTxIn(txIn)
	}
	return nil
}

func addTxOutputs(tx *wire.MsgTx, params CreateParams, satoshiRemainder int64, addrs []address) {
	if params.SendAll {
		fullBalance := calcBalanceOfAddresses(addrs)
		tx.AddTxOut(wire.NewTxOut(fullBalance-params.MinerFee, params.destInfos[0].payAddress))
	} else {
		for _, info := range params.destInfos {
			tx.AddTxOut(wire.NewTxOut(info.Amount, info.payAddress))
		}
		if satoshiRemainder > 0 {
			tx.AddTxOut(wire.NewTxOut(satoshiRemainder, params.pkInfos[0].payAddress))
		}
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

type destinationInfo struct {
	Destination
	payAddress []byte
}

func toDestInfo(d Destination, net netchain.Net) (destinationInfo, error) {
	payAddress, err := toPayAddress(d.Address, net)
	if err != nil {
		return destinationInfo{}, err
	}
	return destinationInfo{
		Destination: d,
		payAddress:  payAddress,
	}, err
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
		if accBalance >= amountToSend {
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
			utxosToSpendMap[h.String() + strconv.Itoa(u.TxOutIdx)] = utxoWithKey{UTXO: u, wif: wif}
		}
	}

	for i, in := range tx.TxIn {
		utxoOfIn := utxosToSpendMap[in.PreviousOutPoint.Hash.String() + strconv.Itoa(int(in.PreviousOutPoint.Index))]
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
