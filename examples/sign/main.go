// SPDX-License-Identifier: Apache-2.0 OR MIT
// SPDX-FileCopyrightText: 2026 Sebastien Rousseau

// Command sign creates a server-held Ed25519 key, signs with it by
// keyId, verifies the signature with the public key, and deletes the
// key. The private key never leaves the server.
//
//	CRYPTO_SERVICE_URL=http://127.0.0.1:3000 CRYPTO_SERVICE_API_KEY=... go run ./examples/sign
package main

import (
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

	key, err := client.GenerateKeyPair(ctx, cs.GenerateKeyPairParams{
		Algorithm: cs.Ed25519,
		Metadata:  &cs.KeyMetadata{Kid: "example", Use: "sig"},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if _, err := client.KeysDelete(ctx, cs.KeyIDParams{KeyID: key.KeyID}); err != nil {
			log.Print(err)
		}
	}()
	fmt.Println("key:", key.KeyID, key.Algorithm)

	sig, err := client.Sign(ctx, cs.SignParams{KeyID: key.KeyID, Message: "pay 10 EUR"})
	if err != nil {
		log.Fatal(err)
	}
	ok, err := client.Verify(ctx, cs.VerifyParams{PublicKey: key.PublicKey, Message: "pay 10 EUR", Signature: sig.Signature})
	if err != nil {
		log.Fatal(err)
	}
	tampered, err := client.Verify(ctx, cs.VerifyParams{PublicKey: key.PublicKey, Message: "pay 99 EUR", Signature: sig.Signature})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("valid:", ok.Valid, "tampered valid:", tampered.Valid)
	if !ok.Valid || tampered.Valid {
		log.Fatal("unexpected verification result")
	}
}
