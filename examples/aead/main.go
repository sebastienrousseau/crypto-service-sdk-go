// SPDX-License-Identifier: Apache-2.0 OR MIT

// Command aead encrypts a message with XChaCha20-Poly1305 under a
// 256-bit key and decrypts it again. The key here is a fixed synthetic
// test value; never hard-code a real one.
//
//	CRYPTO_SERVICE_URL=http://127.0.0.1:3000 CRYPTO_SERVICE_API_KEY=... go run ./examples/aead
package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"

	cs "github.com/sebastienrousseau/crypto-service-sdk-go"
)

func main() {
	client, err := cs.NewClient(os.Getenv("CRYPTO_SERVICE_URL"),
		cs.WithAPIKey(os.Getenv("CRYPTO_SERVICE_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	ctx := context.Background()
	key := cs.HexOf(bytes.Repeat([]byte{0x42}, 32)) // synthetic test key

	sealed, err := client.Encrypt(ctx, cs.EncryptParams{Key: key, Plaintext: "round trip"})
	if err != nil {
		log.Fatal(err)
	}
	raw, err := sealed.Ciphertext.Bytes()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s: %d bytes of nonce, ciphertext and tag\n", sealed.Algorithm, len(raw))

	opened, err := client.Decrypt(ctx, cs.DecryptParams{Key: key, Ciphertext: sealed.Ciphertext})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("decrypted:", opened.Plaintext)
	if opened.Plaintext != "round trip" {
		log.Fatal("round trip failed")
	}
}
