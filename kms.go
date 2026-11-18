// SPDX-License-Identifier: Apache-2.0 OR MIT
// SPDX-FileCopyrightText: 2026 Sebastien Rousseau

package cryptoservice

import "context"

// KmsEncoding says how KmsEncryptParams.Plaintext is to be read.
type KmsEncoding string

// The encodings of /v2/kms/encrypt. The server never guesses: it reads
// UTF-8 unless told hex, and the client always tells it.
const (
	KmsUTF8 KmsEncoding = "utf8"
	KmsHex  KmsEncoding = "hex"
)

// KmsCreateKeyParams is the body of POST /v2/kms/create-key. The zero
// value creates an encryption key with the default provider.
type KmsCreateKeyParams struct {
	// Provider is local, aws, gcp or vault, as the server enables them.
	Provider  string `json:"provider,omitempty"`
	Algorithm string `json:"algorithm,omitempty"`
	// Usage is encrypt, sign or wrap.
	Usage    string         `json:"usage,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// KmsKey is the data of POST /v2/kms/create-key.
type KmsKey struct {
	KeyID     string `json:"keyId"`
	Provider  string `json:"provider"`
	Algorithm string `json:"algorithm"`
	Usage     string `json:"usage"`
	Enabled   bool   `json:"enabled"`
	CreatedAt string `json:"createdAt"`
	// CurrentVersion is set for a versioned key.
	CurrentVersion *int `json:"currentVersion,omitempty"`
}

// KmsWrapParams is the body of POST /v2/kms/wrap.
type KmsWrapParams struct {
	KeyID        string            `json:"keyId"`
	UnwrappedKey Hex               `json:"unwrappedKey"`
	Provider     string            `json:"provider,omitempty"`
	Context      map[string]string `json:"context,omitempty"`
}

// KmsWrapResult is the data of POST /v2/kms/wrap.
type KmsWrapResult struct {
	// WrappedKey is opaque and provider-specific; pass it back as is.
	WrappedKey string `json:"wrappedKey"`
	KeyID      string `json:"keyId"`
	Provider   string `json:"provider"`
}

// KmsUnwrapParams is the body of POST /v2/kms/unwrap.
type KmsUnwrapParams struct {
	KeyID      string            `json:"keyId"`
	WrappedKey string            `json:"wrappedKey"`
	Provider   string            `json:"provider,omitempty"`
	Context    map[string]string `json:"context,omitempty"`
}

// KmsUnwrapResult is the data of POST /v2/kms/unwrap.
type KmsUnwrapResult struct {
	UnwrappedKey Hex    `json:"unwrappedKey"`
	KeyID        string `json:"keyId"`
	Provider     string `json:"provider"`
}

// KmsGenerateDataKeyParams is the body of POST /v2/kms/generate-data-key.
type KmsGenerateDataKeyParams struct {
	KeyID    string `json:"keyId"`
	KeySpec  string `json:"keySpec,omitempty"`
	Provider string `json:"provider,omitempty"`
}

// KmsDataKey is the data of POST /v2/kms/generate-data-key: the data
// key in the clear (a secret) and wrapped under the KMS key.
type KmsDataKey struct {
	Plaintext  Hex    `json:"plaintext"`
	Ciphertext string `json:"ciphertext"`
	KeyID      string `json:"keyId"`
	Provider   string `json:"provider"`
}

// KmsEncryptParams is the body of POST /v2/kms/encrypt. Encoding is
// always sent; left empty it is sent as KmsUTF8.
type KmsEncryptParams struct {
	KeyID     string            `json:"keyId"`
	Plaintext string            `json:"plaintext"`
	Encoding  KmsEncoding       `json:"encoding"`
	Provider  string            `json:"provider,omitempty"`
	Context   map[string]string `json:"context,omitempty"`
}

// KmsEncryptResult is the data of POST /v2/kms/encrypt.
type KmsEncryptResult struct {
	// Ciphertext is opaque and provider-specific; pass it back as is.
	Ciphertext string `json:"ciphertext"`
	KeyID      string `json:"keyId"`
	Provider   string `json:"provider"`
	// KeyVersion is set for a versioned key.
	KeyVersion *int `json:"keyVersion,omitempty"`
	// Context is set when the provider echoes it.
	Context map[string]string `json:"context,omitempty"`
}

// KmsDecryptParams is the body of POST /v2/kms/decrypt.
type KmsDecryptParams struct {
	KeyID      string            `json:"keyId"`
	Ciphertext string            `json:"ciphertext"`
	Provider   string            `json:"provider,omitempty"`
	Context    map[string]string `json:"context,omitempty"`
}

// KmsDecryptResult is the data of POST /v2/kms/decrypt: the plaintext
// as text and as hex, so bytes encrypted with KmsHex survive.
type KmsDecryptResult struct {
	Plaintext string `json:"plaintext"`
	Hex       Hex    `json:"hex"`
	KeyID     string `json:"keyId"`
	Provider  string `json:"provider"`
}

// KmsCreateKey creates a managed key (POST /v2/kms/create-key, scope
// crypto:keys). Stateful: retried only after 429 or 503.
func (c *Client) KmsCreateKey(ctx context.Context, p KmsCreateKeyParams) (*KmsKey, error) {
	return post[KmsKey](ctx, c, "/v2/kms/create-key", p)
}

// KmsWrap wraps key bytes under a managed key (POST /v2/kms/wrap, scope
// crypto:keys).
func (c *Client) KmsWrap(ctx context.Context, p KmsWrapParams) (*KmsWrapResult, error) {
	return post[KmsWrapResult](ctx, c, "/v2/kms/wrap", p)
}

// KmsUnwrap reverses KmsWrap (POST /v2/kms/unwrap, scope crypto:keys).
func (c *Client) KmsUnwrap(ctx context.Context, p KmsUnwrapParams) (*KmsUnwrapResult, error) {
	return post[KmsUnwrapResult](ctx, c, "/v2/kms/unwrap", p)
}

// KmsGenerateDataKey draws a data key under a managed key (POST
// /v2/kms/generate-data-key, scope crypto:keys).
func (c *Client) KmsGenerateDataKey(ctx context.Context, p KmsGenerateDataKeyParams) (*KmsDataKey, error) {
	return post[KmsDataKey](ctx, c, "/v2/kms/generate-data-key", p)
}

// KmsEncrypt encrypts under a managed key (POST /v2/kms/encrypt, scope
// crypto:encrypt). The encoding is always sent explicitly.
func (c *Client) KmsEncrypt(ctx context.Context, p KmsEncryptParams) (*KmsEncryptResult, error) {
	if p.Encoding == "" {
		p.Encoding = KmsUTF8
	}
	return post[KmsEncryptResult](ctx, c, "/v2/kms/encrypt", p)
}

// KmsDecrypt reverses KmsEncrypt (POST /v2/kms/decrypt, scope
// crypto:decrypt).
func (c *Client) KmsDecrypt(ctx context.Context, p KmsDecryptParams) (*KmsDecryptResult, error) {
	return post[KmsDecryptResult](ctx, c, "/v2/kms/decrypt", p)
}
