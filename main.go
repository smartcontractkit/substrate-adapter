package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Starting Substrate adapter")

	privkey := os.Getenv("SA_PRIVATE_KEY")
	txType := os.Getenv("SA_TX_TYPE")
	endpoint := os.Getenv("SA_ENDPOINT")

	adapter, err := newSubstrateAdapter(privkey, txType, endpoint)
	if err != nil {
		fmt.Println("Failed starting Substrate adapter:", err)
		return
	}

	RunWebserver(adapter.handle)
}
