// SPDX-License-Identifier: Apache-2.0 OR MIT

package cryptoservice_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	// Named after the package, as doc.go writes it, so the examples on
	// pkg.go.dev read like caller code, and so gobco (make
	// branch-coverage) finds the import for its black-box test bridge.
	cryptoservice "github.com/sebastienrousseau/crypto-service-sdk-go"
)

// A client is created once and shared; every call takes a context.
func ExampleClient_Hash() {
	client, err := cryptoservice.NewClient(os.Getenv("CRYPTO_SERVICE_URL"),
		cryptoservice.WithAPIKey(os.Getenv("CRYPTO_SERVICE_API_KEY")),
		cryptoservice.WithRetry(cryptoservice.RetryPolicy{MaxRetries: 2}))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := client.Hash(ctx, cryptoservice.HashParams{Algorithm: cryptoservice.SHA256, Data: "abc"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res.Digest)
}

// Concurrency is goroutines over one shared client: no async API is
// needed, and the client pools its connections.
func Example_concurrent() {
	client, err := cryptoservice.NewClient(os.Getenv("CRYPTO_SERVICE_URL"),
		cryptoservice.WithAPIKey(os.Getenv("CRYPTO_SERVICE_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	inputs := []string{"a", "b", "c"}
	digests := make([]cryptoservice.Hex, len(inputs))
	errs := make([]error, len(inputs))
	var wg sync.WaitGroup
	for i, in := range inputs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := client.Hash(context.Background(), cryptoservice.HashParams{Algorithm: cryptoservice.BLAKE3, Data: in})
			if err != nil {
				errs[i] = err
				return
			}
			digests[i] = res.Digest
		}()
	}
	wg.Wait()
	if err := errors.Join(errs...); err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(digests))
}

// Problems come back as *APIError; no response at all is a
// *TransportError.
func ExampleAPIError() {
	client, err := cryptoservice.NewClient(os.Getenv("CRYPTO_SERVICE_URL"),
		cryptoservice.WithAPIKey(os.Getenv("CRYPTO_SERVICE_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	_, err = client.Sign(context.Background(), cryptoservice.SignParams{KeyID: "k_AAAAAAAAAAAAAAAAAAAAAA", Message: "m"})
	var apiErr *cryptoservice.APIError
	var transport *cryptoservice.TransportError
	switch {
	case errors.As(err, &apiErr):
		fmt.Println(apiErr.Status, apiErr.Problem.Slug(), apiErr.Problem.Code)
	case errors.As(err, &transport):
		fmt.Println("no response; timed out:", transport.TimedOut)
	}
}
