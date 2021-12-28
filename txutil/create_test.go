package txutil

import (
	"github.com/btcsuite/btcd/wire"
	"github.com/glossd/btc/addressinfo"
	"github.com/glossd/btc/netchain"
	"github.com/stretchr/testify/assert"
	"testing"
)

const privateKey1 = "932u6Q4xEC9UYRb3rS2BWrSpSPEt5KaU8NNP7EWy7zSkWmfBiGe"
const privateKey2 = "cMvRbsVJKjRkZTV7tosWEYEu1x8tQcnLEbC64RiKwPeeEz29j8QZ"
const privateKey3 = "93UVjiGYyB6q16iMPuKjYePdLesaYvdMyP3EjE1PjZEqzd456h1"
// destination of each private key
const destination1 = "mgFv6afUVhrdd3D6mY2iyWzHVk5b64qTok"
const destination2 = "n4kkk9H2jGj7t8LA4vxK4DHM7Lq95VaEXC"
const destination3 = "mwRL1TpsRSFy5KXbxEd2KrHiD16VvbbAdj"


func TestCreate_SendAll(t *testing.T) {
	t.Run("Through amount", func(t *testing.T) {
		rawTx, err := Create(CreateParams{
			PrivateKey:  privateKey1,
			Destination: destination2,
			Amount:      addressinfo.MockAddressBalance - defaultMinerFee,
			Fetch: addressinfo.FetchMock,
			Net:         netchain.TestNet,
		})
		assert.Nil(t, err)

		tx := decodeTx(t, rawTx)
		assert.EqualValues(t, 1, len(tx.TxIn))
		assert.EqualValues(t, 1, len(tx.TxOut))
	})
	t.Run("SendAll flag true", func(t *testing.T) {
		rawTx, err := Create(CreateParams{
			PrivateKey: privateKey1,
			Destination: destination2,
			SendAll: true,
			Fetch: addressinfo.FetchMock,
			Net: netchain.TestNet,
		})
		assert.Nil(t, err)

		tx := decodeTx(t, rawTx)
		assert.EqualValues(t, 1, len(tx.TxIn))
		assert.EqualValues(t, 1, len(tx.TxOut))
	})
}

func TestCreate_MultiplePrivateKeys(t *testing.T) {
	t.Run("SendAll", func(t *testing.T) {
		rawTx, err := Create(CreateParams{
			PrivateKeys: []string{privateKey1, privateKey2},
			Destination: destination3,
			SendAll: true,
			Fetch: addressinfo.FetchMock,
			Net: netchain.TestNet,
		})
		assert.Nil(t, err)

		tx := decodeTx(t, rawTx)
		assert.GreaterOrEqual(t, len(tx.TxIn), 2)
	})
	t.Run("WithRemainder", func(t *testing.T) {
		amount := addressinfo.MockAddressBalance*3/2 - defaultMinerFee
		rawTx, err := Create(CreateParams{
			PrivateKeys: []string{privateKey1, privateKey2},
			Destination: destination3,
			Amount:      amount,
			Fetch:       addressinfo.FetchMock,
			Net:         netchain.TestNet,
		})
		assert.Nil(t, err)

		tx := decodeTx(t, rawTx)
		assert.EqualValues(t, 2, len(tx.TxIn))
		assert.EqualValues(t, 2, len(tx.TxOut))
		assert.EqualValues(t, amount, tx.TxOut[0].Value)
		assert.EqualValues(t, addressinfo.MockAddressBalance/2, tx.TxOut[1].Value)
		assert.EqualValues(t, tx.TxOut[1].PkScript, addressPkScript(t, destination2))
	})
}


func TestCreate_ToMultipleDestinations(t *testing.T) {
	rawTx, err := Create(CreateParams{
		PrivateKey: privateKey1,
		Destinations: []Destination{{Address: destination2, Amount: 200000}, {Address: destination3, Amount: 300000}},
		Fetch: addressinfo.FetchMock,
		Net: netchain.TestNet,
	})
	assert.Nil(t, err)

	tx := decodeTx(t, rawTx)
	assert.EqualValues(t, 3, len(tx.TxOut))
}

func TestCreate_Validation(t *testing.T) {
	type test struct {
		input CreateParams
	}
	var okAmount int64 = 50000
	dests := func(addresses ...string) []Destination {
		var result []Destination
		for _, a := range addresses {
			result = append(result, Destination{Address: a, Amount: okAmount})
		}
		return result
	}
	shouldntPass := []test{
		{input: CreateParams{Destination: destination2, Amount: okAmount}},
		{input: CreateParams{PrivateKey: privateKey1, Amount: okAmount}},
		{input: CreateParams{PrivateKey: privateKey1, Destination: destination2}},
		{input: CreateParams{PrivateKey: privateKey1, Destination: destination2, Amount: 10}},
		{input: CreateParams{PrivateKey: privateKey1, Destinations: dests(destination2, destination3), SendAll: true}},
		{input: CreateParams{PrivateKey: privateKey1, Destinations: []Destination{{Address: destination2}}}},
		{input: CreateParams{PrivateKey: privateKey1, Destinations: []Destination{{Amount: okAmount}}}},
	}

	for _, test := range shouldntPass {
		test.input.Net = netchain.TestNet
		test.input.Fetch = addressinfo.FetchMock
		_, err := Create(test.input)
		assert.NotNil(t, err)
	}

	shouldPass := []test{
		{input: CreateParams{PrivateKey: privateKey1, Destination: destination2, Amount: okAmount}},
		{input: CreateParams{PrivateKey: privateKey1, Destination: destination2, SendAll: true}},
		{input: CreateParams{PrivateKey: privateKey1, Destinations: dests(destination2)}},
		{input: CreateParams{PrivateKey: privateKey1, Destinations: dests(destination2, destination3)}},
		{input: CreateParams{PrivateKey: privateKey1, Destinations: []Destination{{Address: destination2}}, SendAll: true}},
		{input: CreateParams{PrivateKeys: []string{privateKey1, privateKey2}, Destination: destination2, Amount: okAmount}},
	}
	for _, test := range shouldPass {
		test.input.Fetch = addressinfo.FetchMock
		test.input.Net = netchain.TestNet
		_, err := Create(test.input)
		assert.Nil(t, err)
	}
}

func decodeTx(t *testing.T, rawTx string) *wire.MsgTx {
	tx, err := hexDecodeTx(rawTx)
	assert.Nil(t, err)
	return tx
}

func addressPkScript(t *testing.T, address string) []byte {
	script, err := addressToPkScript(address, netchain.TestNet)
	assert.Nil(t, err)
	return script
}


