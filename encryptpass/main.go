package main

import (
	"context"
	"fmt"

	"github.com/alexballas/vaultlib"
)

func main() {
	ctx := context.Background()
	conf := vaultlib.NewConfig()
	
	transitclient, err := vaultlib.NewTransitClient(conf, "my-key")
	check(err)

	text := "Encode me please!"

	cipher, _, err := transitclient.Encrypt(ctx, text)
	check(err)

	dec, err := transitclient.Decrypt(ctx, cipher)
	check(err)

	fmt.Printf("Text     %s\n", text)
	fmt.Printf("Encoded: %s\n", cipher)
	fmt.Printf("Decoded: %s\n", dec)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
