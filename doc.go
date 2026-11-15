// SPDX-License-Identifier: Apache-2.0 OR MIT

// Package cryptoservice is a typed client for the Crypto Service REST
// API (crypto-server), version 2.
//
// The server does the cryptography; this package only sends JSON
// requests and decodes the answers. It holds no keys and implements no
// algorithm: key-generation methods return a keyId with the public
// key, and signing, decapsulation and opening take that keyId, so
// private keys never leave the server.
//
// # Calls
//
// Every method takes a context.Context first and blocks until the
// server answers, the context ends or the client's timeout or deadline
// passes. That blocking, context-aware call is the whole API: there is
// no separate asynchronous variant. Run calls concurrently by starting
// goroutines; a single *Client is safe for concurrent use and shares one
// pool of keep-alive connections (HTTP/2 over TLS where the server
// offers it).
//
//	client, err := cryptoservice.NewClient("http://127.0.0.1:3000",
//		cryptoservice.WithAPIKey(os.Getenv("CRYPTO_SERVICE_API_KEY")))
//	if err != nil {
//		return err
//	}
//	defer client.Close()
//	res, err := client.Hash(ctx, cryptoservice.HashParams{
//		Algorithm: cryptoservice.SHA256, Data: "abc"})
//
// # Encodings
//
// Fields are typed by the encoding the API documents: Hex for keys,
// signatures, digests and KEM output, Base64 for AEAD, box, stream and
// key-wrap output, string for UTF-8 text and for opaque KMS values.
// Hex and Base64 are strings, sent exactly as written; their Bytes
// methods and the HexOf and Base64Of constructors convert.
//
// # Errors
//
// A call returns an *APIError for an error status (with the server's
// RFC 9457 problem), a *TransportError when no response arrived (with
// TimedOut), or a *DecodeError for a success body that is not the
// documented JSON. All three implement Error. No error message carries
// a credential, a key, a plaintext or a body.
//
// # Retries
//
// WithRetry opts into retries with exponential backoff, full jitter
// and Retry-After. Calls that create or delete server state are retried
// only when the server certainly did not act on them (429 or 503).
//
// # Security
//
// The client refuses redirects, keeps TLS certificate verification on
// (it refuses an injected transport that turns it off), logs nothing,
// and redacts its credentials from String and GoString.
package cryptoservice
