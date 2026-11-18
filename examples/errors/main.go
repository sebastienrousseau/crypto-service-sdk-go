// SPDX-License-Identifier: Apache-2.0 OR MIT
// SPDX-FileCopyrightText: 2026 Sebastien Rousseau

// Command errors shows the three kinds of failure: an RFC 9457 problem
// for an unknown key, a validation problem with per-field errors, and a
// transport error for a server that is not there.
//
//	CRYPTO_SERVICE_URL=http://127.0.0.1:3000 CRYPTO_SERVICE_API_KEY=... go run ./examples/errors
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"

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

	_, err = client.Sign(ctx, cs.SignParams{KeyID: "k_AAAAAAAAAAAAAAAAAAAAAA", Message: "m"})
	var apiErr *cs.APIError
	if !errors.As(err, &apiErr) || apiErr.Status != 404 {
		log.Fatalf("want a 404 problem, got %v", err)
	}
	fmt.Println(err)
	fmt.Println("  type:", apiErr.Problem.Type, "code:", apiErr.Problem.Code, "detail:", apiErr.Problem.Detail)

	_, err = client.Hash(ctx, cs.HashParams{Algorithm: "md5", Data: "abc"})
	if !errors.As(err, &apiErr) || apiErr.Problem.Slug() != "validation-failed" {
		log.Fatalf("want a validation problem, got %v", err)
	}
	fmt.Println(err)
	for _, field := range apiErr.Problem.Errors {
		fmt.Println("  field:", field.Field, "-", field.Message)
	}

	nowhere, err := cs.NewClient("http://"+closedPort(), cs.WithTimeout(2*time.Second))
	if err != nil {
		log.Fatal(err)
	}
	_, err = nowhere.Health(ctx)
	var transport *cs.TransportError
	if !errors.As(err, &transport) {
		log.Fatalf("want a transport error, got %v", err)
	}
	var base cs.Error
	fmt.Println("no response, timed out:", transport.TimedOut, "- one base type:", errors.As(err, &base))
}

// closedPort returns a loopback address nothing listens on.
func closedPort() string {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	return addr
}
