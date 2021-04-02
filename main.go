package main

import (
	"fmt"
	"github.com/smartcontractkit/substrate-adapter/adapter"
	"os"
)

func main() {
	fmt.Println("Starting Substrate adapter")

	privkey := os.Getenv("SA_PRIVATE_KEY")
	txType := os.Getenv("SA_TX_TYPE")
	endpoint := os.Getenv("SA_ENDPOINT")

	adapterClient, err := adapter.NewSubstrateAdapter(privkey, txType, endpoint)
	if err != nil {
		fmt.Println("Failed starting Substrate adapter:", err)
		return
	}

	adapter.RunWebserver(adapterClient.Handle)
}
