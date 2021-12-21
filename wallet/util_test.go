package wallet

import (
	"github.com/glossd/btc/netchain"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddressFromPrivateKey(t *testing.T) {
	address, err := AddressFromPrivateKey("932u6Q4xEC9UYRb3rS2BWrSpSPEt5KaU8NNP7EWy7zSkWmfBiGe", netchain.TestNet)
	assert.Nil(t, err)
	assert.EqualValues(t, "mgFv6afUVhrdd3D6mY2iyWzHVk5b64qTok", address)
}

func TestIsAddressValid(t *testing.T) {
	var valid = []string{
		"16ftSEQ4ctQFDtVZiUBusQUjRrGhM3JYwe",
		"3D2oetdNuZUqQHPJmcMDDHYoqkyNVsFk9r",
		"16rCmCmbuWDhPjWTrpQGaU3EPdZF7MTdUk",
		"3Cbq7aT1tY8kMxWLbitaG7yT6bPbKChq64",
		"3Nxwenay9Z8Lc9JBiywExpnEFiLp6Afp8v",
	}
	for _, a := range valid {
		assert.True(t, IsAddressValid(a, netchain.MainNet))
	}
	var invalid = []string{
		"",
		"a",
		"address",
		"1234567890",
		"mgFv6afUVhrdd3D6mY2iyWzHVk5b64qTok",
		"ghfteEc4gtQFDtVZiUBusQUjRrGhM3JYwe",
	}
	for _, a := range invalid {
		assert.False(t, IsAddressValid(a, netchain.MainNet))
	}
}
