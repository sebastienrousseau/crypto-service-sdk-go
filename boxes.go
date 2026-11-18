// SPDX-License-Identifier: Apache-2.0 OR MIT
// SPDX-FileCopyrightText: 2026 Sebastien Rousseau

package cryptoservice

import "context"

// SealedboxSealParams is the body of POST /v2/sealedbox/seal.
type SealedboxSealParams struct {
	// RecipientPublicKey is an X25519 public key.
	RecipientPublicKey Hex    `json:"recipientPublicKey"`
	Plaintext          string `json:"plaintext"`
}

// SealedboxSealPqParams is the body of POST /v2/sealedbox/seal-pq.
type SealedboxSealPqParams struct {
	X25519PublicKey Hex    `json:"x25519PublicKey"`
	MlKemPublicKey  Hex    `json:"mlKemPublicKey"`
	Plaintext       string `json:"plaintext"`
}

// SealedboxSealResult is the data of both sealed-box seal routes.
type SealedboxSealResult struct {
	Sealed    Base64 `json:"sealed"`
	Algorithm string `json:"algorithm"`
}

// SealedboxOpenParams is the body of both sealed-box open routes: the
// server-held key of the matching algorithm and the sealed box.
type SealedboxOpenParams struct {
	KeyID  string `json:"keyId"`
	Sealed Base64 `json:"sealed"`
}

// MultiRecipient is one recipient of POST /v2/multi-recipient/encrypt.
type MultiRecipient struct {
	// Type is "classical" (X25519) or "pq" (X25519 + ML-KEM-768).
	Type string `json:"type"`
	// PublicKey is the X25519 public key.
	PublicKey Hex `json:"publicKey"`
	// MlKemPublicKey is required for type "pq".
	MlKemPublicKey Hex `json:"mlKemPublicKey,omitempty"`
}

// MultiRecipientParams is the body of POST /v2/multi-recipient/encrypt.
type MultiRecipientParams struct {
	Plaintext  string           `json:"plaintext"`
	Recipients []MultiRecipient `json:"recipients"`
}

// MultiRecipientSlot is the content key wrapped for one recipient.
type MultiRecipientSlot struct {
	Type               string `json:"type"`
	EphemeralPublicKey Hex    `json:"ephemeralPublicKey"`
	WrappedKey         Base64 `json:"wrappedKey"`
	MlKemCiphertext    Hex    `json:"mlKemCiphertext,omitempty"`
}

// MultiRecipientResult is the data of POST /v2/multi-recipient/encrypt.
type MultiRecipientResult struct {
	Algorithm  string               `json:"algorithm"`
	Ciphertext Base64               `json:"ciphertext"`
	Recipients []MultiRecipientSlot `json:"recipients"`
}

// StreamPqEncryptParams is the body of POST /v2/stream/pq-encrypt.
type StreamPqEncryptParams struct {
	X25519PublicKey Hex    `json:"x25519PublicKey"`
	MlKemPublicKey  Hex    `json:"mlKemPublicKey"`
	Plaintext       string `json:"plaintext"`
	// ChunkSize is in bytes; zero takes the server's 65536.
	ChunkSize int `json:"chunkSize,omitempty"`
}

// StreamEncryptResult is the data of both stream encryption routes.
type StreamEncryptResult struct {
	Ciphertext Base64 `json:"ciphertext"`
	Algorithm  string `json:"algorithm"`
	// RecipientCount is set by the multi-recipient route only.
	RecipientCount int `json:"recipientCount,omitempty"`
}

// StreamPqDecryptParams is the body of POST /v2/stream/pq-decrypt.
type StreamPqDecryptParams struct {
	KeyID      string `json:"keyId"`
	Ciphertext Base64 `json:"ciphertext"`
	ChunkSize  int    `json:"chunkSize,omitempty"`
}

// StreamDecryptResult is the data of both stream decryption routes.
type StreamDecryptResult struct {
	Plaintext string `json:"plaintext"`
	// RecipientID is set by the multi-recipient route only.
	RecipientID string `json:"recipientId,omitempty"`
}

// StreamRecipient is one recipient of a multi-recipient stream.
type StreamRecipient struct {
	RecipientID     string `json:"recipientId"`
	X25519PublicKey Hex    `json:"x25519PublicKey"`
	MlKemPublicKey  Hex    `json:"mlKemPublicKey"`
}

// StreamMultiPqEncryptParams is the body of POST
// /v2/stream/multi-pq-encrypt.
type StreamMultiPqEncryptParams struct {
	Recipients []StreamRecipient `json:"recipients"`
	Plaintext  string            `json:"plaintext"`
	ChunkSize  int               `json:"chunkSize,omitempty"`
}

// StreamMultiPqDecryptParams is the body of POST
// /v2/stream/multi-pq-decrypt.
type StreamMultiPqDecryptParams struct {
	KeyID string `json:"keyId"`
	// RecipientID selects the slot directly; without it the server
	// tries each.
	RecipientID string `json:"recipientId,omitempty"`
	Ciphertext  Base64 `json:"ciphertext"`
	ChunkSize   int    `json:"chunkSize,omitempty"`
}

// SealedboxSeal seals text to an X25519 public key, anonymous sender
// (POST /v2/sealedbox/seal, scope crypto:encrypt).
func (c *Client) SealedboxSeal(ctx context.Context, p SealedboxSealParams) (*SealedboxSealResult, error) {
	return post[SealedboxSealResult](ctx, c, "/v2/sealedbox/seal", p)
}

// SealedboxOpen opens a sealed box with a server-held X25519 key and
// returns the plaintext (POST /v2/sealedbox/open, scope crypto:decrypt).
func (c *Client) SealedboxOpen(ctx context.Context, p SealedboxOpenParams) (string, error) {
	return postString(ctx, c, "/v2/sealedbox/open", p)
}

// SealedboxSealPq seals text to an X25519 + ML-KEM-768 key pair (POST
// /v2/sealedbox/seal-pq, scope crypto:encrypt).
func (c *Client) SealedboxSealPq(ctx context.Context, p SealedboxSealPqParams) (*SealedboxSealResult, error) {
	return post[SealedboxSealResult](ctx, c, "/v2/sealedbox/seal-pq", p)
}

// SealedboxOpenPq opens a hybrid sealed box with a server-held hybrid
// key and returns the plaintext (POST /v2/sealedbox/open-pq, scope
// crypto:decrypt).
func (c *Client) SealedboxOpenPq(ctx context.Context, p SealedboxOpenParams) (string, error) {
	return postString(ctx, c, "/v2/sealedbox/open-pq", p)
}

// MultiRecipientEncrypt encrypts once for several recipients (POST
// /v2/multi-recipient/encrypt, scope crypto:encrypt).
func (c *Client) MultiRecipientEncrypt(ctx context.Context, p MultiRecipientParams) (*MultiRecipientResult, error) {
	return post[MultiRecipientResult](ctx, c, "/v2/multi-recipient/encrypt", p)
}

// StreamPqEncrypt encrypts with the hybrid post-quantum STREAM
// construction (POST /v2/stream/pq-encrypt, scope crypto:encrypt).
func (c *Client) StreamPqEncrypt(ctx context.Context, p StreamPqEncryptParams) (*StreamEncryptResult, error) {
	return post[StreamEncryptResult](ctx, c, "/v2/stream/pq-encrypt", p)
}

// StreamPqDecrypt decrypts a STREAM ciphertext with a server-held
// hybrid key (POST /v2/stream/pq-decrypt, scope crypto:decrypt).
func (c *Client) StreamPqDecrypt(ctx context.Context, p StreamPqDecryptParams) (*StreamDecryptResult, error) {
	return post[StreamDecryptResult](ctx, c, "/v2/stream/pq-decrypt", p)
}

// StreamMultiPqEncrypt encrypts a STREAM for several recipients (POST
// /v2/stream/multi-pq-encrypt, scope crypto:encrypt).
func (c *Client) StreamMultiPqEncrypt(ctx context.Context, p StreamMultiPqEncryptParams) (*StreamEncryptResult, error) {
	return post[StreamEncryptResult](ctx, c, "/v2/stream/multi-pq-encrypt", p)
}

// StreamMultiPqDecrypt decrypts a multi-recipient STREAM with a
// server-held hybrid key (POST /v2/stream/multi-pq-decrypt, scope
// crypto:decrypt).
func (c *Client) StreamMultiPqDecrypt(ctx context.Context, p StreamMultiPqDecryptParams) (*StreamDecryptResult, error) {
	return post[StreamDecryptResult](ctx, c, "/v2/stream/multi-pq-decrypt", p)
}
