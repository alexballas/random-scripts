package main

import (
	"fmt"
	"os"

	"github.com/alexballas/vaultlib"
)

var (
	vaultaddr = os.Getenv("VAULT_ADDR")
	token     = os.Getenv("VAULT_TOKEN")
	namespace = os.Getenv("VAULT_NAMESPACE")
)

func main() {
	transitclient := vaultlib.NewTransitClient(vaultaddr, token, namespace)
	text := "Encode me please!"
	cipher, _, err := transitclient.Encrypt(text)
	check(err)
	dec, err := transitclient.Decrypt(cipher)
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
