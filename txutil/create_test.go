package txutil

import (
	"fmt"
	"log"
	"testing"
)

func TestRun(t *testing.T) {
	rawTx, err := CreateTx("932u6Q4xEC9UYRb3rS2BWrSpSPEt5KaU8NNP7EWy7zSkWmfBiGe",
		"n4kkk9H2jGj7t8LA4vxK4DHM7Lq95VaEXC", 0.01 * 1e8)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("raw signed transaction is: ", rawTx)
}
