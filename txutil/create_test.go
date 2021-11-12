package txutil

import (
	"fmt"
	"github.com/glossd/btc/netchain"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreate(t *testing.T) {
	rawTx, err := Create(CreateParams{
		PrivateKey: "932u6Q4xEC9UYRb3rS2BWrSpSPEt5KaU8NNP7EWy7zSkWmfBiGe",
		Destination: "n4kkk9H2jGj7t8LA4vxK4DHM7Lq95VaEXC",
		Amount: 500000,
		Net: netchain.TestNet,
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
		PrivateKey: "932u6Q4xEC9UYRb3rS2BWrSpSPEt5KaU8NNP7EWy7zSkWmfBiGe",
		Destination: "n4kkk9H2jGj7t8LA4vxK4DHM7Lq95VaEXC",
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
		PrivateKeys: []string{"932u6Q4xEC9UYRb3rS2BWrSpSPEt5KaU8NNP7EWy7zSkWmfBiGe", "cMvRbsVJKjRkZTV7tosWEYEu1x8tQcnLEbC64RiKwPeeEz29j8QZ"},
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
