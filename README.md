The library helps you to transfer bitcoins.
 
```shell script
go get github.com/glossd/btc
```

First we need to create transaction then broadcast it to blockchain.  
 

```go
package main

import (
	"fmt"
	"github.com/glossd/btc/netchain"
	"github.com/glossd/btc/txutil"
)

func main() {
	net := netchain.TestNet
	rawTx, err := txutil.Create(txutil.CreateParams{
		PrivateKey: "your-key", // e.g. 932u6Q4xEC9UYRb3rS2BWrSpSPEt5KaU8NNP7EWy7zSkWmfBiGe
		Destination: "address", // e.g. n4kkk9H2jGj7t8LA4vxK4DHM7Lq95VaEXC
		Amount: 500000, // amount of satoshi to send
		Net: net,
	})
	if err != nil {
		panic(err)
	}

	// you can broadcast transaction yourself.
	// copy rawTx and paste it to https://www.blockchain.com/btc/pushtx
	txID, err := txutil.Broadcast(rawTx, net)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Transaction ID: %s\n", txID)
	// check out your transaction at https://www.blockchain.com/btc-testnet/tx/{txID}
}
```

For testing purposes we use `netchain.TestNet`. If you want to send real bitcoins to blockchain you need to specify BTC_API_KEY env var for blockcypher or you could pass your own txutil.CreateParams.Fetch function to txutil.Create.