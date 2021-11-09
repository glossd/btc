package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddressFromPrivateKey(t *testing.T) {
	address, err := AddressFromPrivateKey("932u6Q4xEC9UYRb3rS2BWrSpSPEt5KaU8NNP7EWy7zSkWmfBiGe")
	assert.Nil(t, err)
	assert.EqualValues(t, "mgFv6afUVhrdd3D6mY2iyWzHVk5b64qTok", address)
}
