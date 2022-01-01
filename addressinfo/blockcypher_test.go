package addressinfo

import (
	"github.com/glossd/btc/netchain"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFetchBlockcypher(t *testing.T) {
	t.Run(netchain.TestNet.String(), func(t *testing.T) {
		got, err := FetchFromBlockcypher("mop76RFpxCMpNBx2M2NtAJsZEmo6qu5PSa", netchain.TestNet)
		assert.Nil(t, err)
		assert.Positive(t, got.Balance)
		assert.Positive(t, len(got.UTXOs))
	})

	t.Run(netchain.MainNet.String(), func(t *testing.T) {
		got, err := FetchFromBlockcypher("3LQUu4v9z6KNch71j7kbj8GPeAGUo1FW6a", netchain.MainNet)
		assert.Nil(t, err)
		assert.Positive(t, got.Balance)
		assert.Positive(t, len(got.UTXOs))
	})
}
