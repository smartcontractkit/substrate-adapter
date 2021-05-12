package main

import (
	"fmt"
	"os"

	"github.com/smartcontractkit/substrate-adapter/adapter"
)

func main() {
	fmt.Println("Starting Substrate adapter")

	privkey := os.Getenv("SA_PRIVATE_KEY")
	txType := os.Getenv("SA_TX_TYPE")
	endpoint := os.Getenv("SA_ENDPOINT")
	port := os.Getenv("SA_PORT")

	adapterClient, err := adapter.NewSubstrateAdapter(privkey, txType, endpoint)
	if err != nil {
		fmt.Println("Failed starting Substrate adapter:", err)
		return
	}

	adapter.RunWebserver(adapterClient.Handle, port)
}
