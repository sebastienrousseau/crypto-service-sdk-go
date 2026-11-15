// SPDX-License-Identifier: Apache-2.0 OR MIT

// Command mlkem creates a server-held ML-KEM-768 key, encapsulates a
// shared secret to its public key, and decapsulates it by keyId; both
// sides must agree.
//
//	CRYPTO_SERVICE_URL=http://127.0.0.1:3000 CRYPTO_SERVICE_API_KEY=... go run ./examples/mlkem
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

	key, err := client.MlKemGenerateKeyPair(ctx)
	if err != nil {
		log.Fatal(err)
	}
	enc, err := client.MlKemEncapsulate(ctx, cs.MlKemEncapsulateParams{PublicKey: key.PublicKey})
	if err != nil {
		log.Fatal(err)
	}
	dec, err := client.MlKemDecapsulate(ctx, cs.MlKemDecapsulateParams{KeyID: key.KeyID, Ciphertext: enc.Ciphertext})
	if err != nil {
		log.Fatal(err)
	}
	ct, _ := enc.Ciphertext.Bytes()
	fmt.Printf("%s: %d-byte ciphertext, secrets match: %v\n", key.Algorithm, len(ct), enc.SharedSecret == dec.SharedSecret)
	if enc.SharedSecret != dec.SharedSecret {
		log.Fatal("shared secrets differ")
	}
	if _, err := client.KeysDelete(ctx, cs.KeyIDParams{KeyID: key.KeyID}); err != nil {
		log.Fatal(err)
	}
}
