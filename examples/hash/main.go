// SPDX-License-Identifier: Apache-2.0 OR MIT

// Command hash digests "abc" with SHA-256 on the server and checks the
// FIPS 180-2 test vector.
//
//	CRYPTO_SERVICE_URL=http://127.0.0.1:3000 CRYPTO_SERVICE_API_KEY=... go run ./examples/hash
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
	res, err := client.Hash(context.Background(), cs.HashParams{Algorithm: cs.SHA256, Data: "abc"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res.Algorithm, res.Length, res.Digest)
	if res.Digest != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		log.Fatal("unexpected digest")
	}
}
