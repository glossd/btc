package addressinfo

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFetchBlockcypher(t *testing.T) {
	got, err := fetchFromBlockcypher("mgFv6afUVhrdd3D6mY2iyWzHVk5b64qTok")
	assert.Nil(t, err)
	fmt.Println(got)
}
