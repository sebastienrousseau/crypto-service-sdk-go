// SPDX-License-Identifier: Apache-2.0 OR MIT
// SPDX-FileCopyrightText: 2026 Sebastien Rousseau

// Command concurrent fans four hash requests out over goroutines that
// share one client, with the standard library only. The client is safe
// for concurrent use and reuses its pooled connections; a context
// deadline bounds the whole batch.
//
//	CRYPTO_SERVICE_URL=http://127.0.0.1:3000 CRYPTO_SERVICE_API_KEY=... go run ./examples/concurrent
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	cs "github.com/sebastienrousseau/crypto-service-sdk-go"
)

func main() {
	client, err := cs.NewClient(os.Getenv("CRYPTO_SERVICE_URL"),
		cs.WithAPIKey(os.Getenv("CRYPTO_SERVICE_API_KEY")),
		cs.WithRetry(cs.RetryPolicy{MaxRetries: 2}))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	inputs := []string{"alpha", "beta", "gamma", "delta"}
	digests := make([]cs.Hex, len(inputs))
	errs := make([]error, len(inputs))
	var wg sync.WaitGroup
	for i, in := range inputs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := client.Hash(ctx, cs.HashParams{Algorithm: cs.BLAKE3, Data: in})
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
	for i, d := range digests {
		fmt.Printf("%-6s %s\n", inputs[i], d)
	}
}
