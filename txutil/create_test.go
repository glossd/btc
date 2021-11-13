package txutil

import (
	"fmt"
	"github.com/glossd/btc/addressinfo"
	"github.com/glossd/btc/netchain"
	"github.com/stretchr/testify/assert"
	"testing"
)

const privateKey1 = "932u6Q4xEC9UYRb3rS2BWrSpSPEt5KaU8NNP7EWy7zSkWmfBiGe"
const privateKey2 = "cMvRbsVJKjRkZTV7tosWEYEu1x8tQcnLEbC64RiKwPeeEz29j8QZ"
const destination1 = "mgFv6afUVhrdd3D6mY2iyWzHVk5b64qTok"
const destination2 = "n4kkk9H2jGj7t8LA4vxK4DHM7Lq95VaEXC"
const destination3 = "mwRL1TpsRSFy5KXbxEd2KrHiD16VvbbAdj"

func TestCreate(t *testing.T) {
	rawTx, err := Create(CreateParams{
		PrivateKey:  privateKey1,
		Destination: destination2,
		Amount:      500000,
		Net:         netchain.TestNet,
	})
	assert.Nil(t, err)

	tx, err := hexDecodeTx(rawTx)
	assert.Nil(t, err)
	assert.Positive(t, len(tx.TxIn))
	assert.EqualValues(t, 1, len(tx.TxOut))

	fmt.Println("raw signed transaction is: ", rawTx)
}

func TestCreate_SendAll(t *testing.T) {
	rawTx, err := Create(CreateParams{
		PrivateKey: privateKey1,
		Destination: destination2,
		SendAll: true,
		Net: netchain.TestNet,
	})
	assert.Nil(t, err)

	tx, err := hexDecodeTx(rawTx)
	assert.Nil(t, err)
	assert.Positive(t, len(tx.TxIn))

	fmt.Println("raw signed transaction is: ", rawTx)
}

func TestCreate_MultiplePrivateKeys(t *testing.T) {
	rawTx, err := Create(CreateParams{
		PrivateKeys: []string{privateKey1, "cMvRbsVJKjRkZTV7tosWEYEu1x8tQcnLEbC64RiKwPeeEz29j8QZ"},
		Destination: "mwRL1TpsRSFy5KXbxEd2KrHiD16VvbbAdj",
		SendAll: true,
		Net: netchain.TestNet,
	})
	assert.Nil(t, err)

	tx, err := hexDecodeTx(rawTx)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(tx.TxIn), 2)
	fmt.Println("raw signed transaction is: ", rawTx)
}
func TestCreate_ToMultipleDestinations(t *testing.T) {
	rawTx, err := Create(CreateParams{
		PrivateKey: privateKey1,
		Destinations: []Destination{{Address: destination2, Amount: 200000}, {Address: destination3, Amount: 300000}},
		Net: netchain.TestNet,
	})
	assert.Nil(t, err)

	tx, err := hexDecodeTx(rawTx)
	assert.Nil(t, err)
	assert.EqualValues(t, 3, len(tx.TxOut))
	fmt.Println("raw signed transaction is: ", rawTx)
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
