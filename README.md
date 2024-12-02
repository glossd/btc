This library helps you transfer your bitcoins, saving your time from managing UTXOs, inputs, outputs and signatures.
 
```shell script
go get github.com/glossd/btc
```

## Generating wallet
```go
package main

import (
	"fmt"
	"github.com/glossd/btc/netchain"
	"github.com/glossd/btc/wallet"
)

func main() {
	privateKey, address := wallet.New(netchain.TestNet)
	fmt.Printf("Private Key: %s\nBitcoin address: %s\n", privateKey, address)
}
```
If you just started learning about bitcoins and blockchain, you probably **don't have any testnet bitcoins**, wondering where I can get some.
People on [bitcoin.stackexchange](https://bitcoin.stackexchange.com/questions/17690/is-there-any-where-to-get-free-testnet-bitcoins) provided a lot of links.    

## Sending transaction
First we need to create a transaction then broadcast it to blockchain.  

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
		Amount: 500000, // satoshi to send
		Net: net,
	})
	if err != nil {
		panic(err)
	}

	// you can broadcast transaction yourself on any blockchain website.
	// e.g. copy rawTx and paste it to https://live.blockcypher.com/btc-testnet/pushtx/
	txID, err := txutil.Broadcast(rawTx, net)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Transaction ID: %s\n", txID)
	// check out your transaction at https://www.blockchain.com/btc-testnet/tx/{txID}
}
```
*all the code can be found in the 'examples' folder.*

**For testing purposes** I used `netchain.TestNet`.
If you want to send real bitcoins to blockchain you need to specify BTC_API_KEY env var for blockcypher or you could pass your own txutil.CreateParams.Fetch function to txutil.Create,
refer to [examples/create-real-transaction](https://github.com/glossd/btc/blob/master/examples/create-real-transaction/main.go)

---   
### More options
`txutil.CreateParams`:

| Field         | Type                 | Usage  |
|:-------------:|:--------------------:|------- |
| PrivateKeys  | []string              | send your bitcoins from multiple wallets |
| Destinations | []txutil.Destination  | send your bitcoins to multiple addresses |
| SendAll      | bool                  | send all your bitcoins from your private key or keys, but it only works if you specified just one destination |

For the full list of the transaction parameters look inside `txutil.CreateParams`.
