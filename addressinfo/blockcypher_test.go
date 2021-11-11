package addressinfo

import (
	"fmt"
	"github.com/glossd/btc/netchain"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFetchBlockcypher(t *testing.T) {
	got, err := FetchFromBlockcypher("mgFv6afUVhrdd3D6mY2iyWzHVk5b64qTok", netchain.TestNet)
	assert.Nil(t, err)
	fmt.Printf("UTXOs: %+v", got)
}
