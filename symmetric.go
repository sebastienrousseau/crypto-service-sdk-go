// SPDX-License-Identifier: Apache-2.0 OR MIT

package cryptoservice

import "context"

// EncryptParams is the body of POST /v2/encrypt.
type EncryptParams struct {
	// Key is the 256-bit key, as 64 hex characters.
	Key Hex `json:"key"`
	// Plaintext is UTF-8 text, 1 byte to 10 MiB.
	Plaintext string `json:"plaintext"`
}

// EncryptResult is the data of POST /v2/encrypt.
type EncryptResult struct {
	// Ciphertext is the nonce, ciphertext and tag.
	Ciphertext Base64 `json:"ciphertext"`
	// Algorithm is "xchacha20-poly1305".
	Algorithm string `json:"algorithm"`
}

// DecryptParams is the body of POST /v2/decrypt.
type DecryptParams struct {
	Key        Hex    `json:"key"`
	Ciphertext Base64 `json:"ciphertext"`
}

// DecryptResult is the data of POST /v2/decrypt.
type DecryptResult struct {
	// Plaintext is the recovered text; invalid UTF-8 becomes U+FFFD.
	Plaintext string `json:"plaintext"`
}

// SecretboxSealParams is the body of POST /v2/secretbox/seal.
type SecretboxSealParams struct {
	Key       Hex    `json:"key"`
	Plaintext string `json:"plaintext"`
	// AAD is optional associated data, authenticated but not encrypted.
	AAD string `json:"aad,omitempty"`
}

// SecretboxSealResult is the data of POST /v2/secretbox/seal.
type SecretboxSealResult struct {
	Sealed    Base64 `json:"sealed"`
	Algorithm string `json:"algorithm"`
}

// SecretboxOpenParams is the body of POST /v2/secretbox/open.
type SecretboxOpenParams struct {
	Key        Hex    `json:"key"`
	Ciphertext Base64 `json:"ciphertext"`
	AAD        string `json:"aad,omitempty"`
}

// PasswordEncryptParams is the body of POST /v2/password/encrypt.
type PasswordEncryptParams struct {
	Password  string `json:"password"`
	Plaintext string `json:"plaintext"`
}

// PasswordEncryptResult is the data of POST /v2/password/encrypt.
type PasswordEncryptResult struct {
	// Encrypted holds the parameters, salt, nonce and ciphertext.
	Encrypted Base64 `json:"encrypted"`
	// Algorithm is "argon2id-xchacha20-poly1305".
	Algorithm string `json:"algorithm"`
}

// PasswordDecryptParams is the body of POST /v2/password/decrypt.
type PasswordDecryptParams struct {
	Password string `json:"password"`
	// Ciphertext is PasswordEncryptResult.Encrypted.
	Ciphertext Base64 `json:"ciphertext"`
}

// KeyWrapAlgorithm names an AES key-wrap mode.
type KeyWrapAlgorithm string

// The key-wrap modes of /v2/keys/wrap and /v2/keys/unwrap.
const (
	AESKW  KeyWrapAlgorithm = "aes-kw"  // RFC 3394, the server's default
	AESKWP KeyWrapAlgorithm = "aes-kwp" // RFC 5649, with padding
)

// KeyWrapParams is the body of POST /v2/keys/wrap.
type KeyWrapParams struct {
	// KEK is the key-encryption key.
	KEK       Hex              `json:"kek"`
	KeyToWrap Hex              `json:"keyToWrap"`
	Algorithm KeyWrapAlgorithm `json:"algorithm,omitempty"`
}

// KeyWrapResult is the data of POST /v2/keys/wrap.
type KeyWrapResult struct {
	Wrapped   Base64           `json:"wrapped"`
	Algorithm KeyWrapAlgorithm `json:"algorithm"`
}

// KeyUnwrapParams is the body of POST /v2/keys/unwrap.
type KeyUnwrapParams struct {
	KEK        Hex              `json:"kek"`
	WrappedKey Base64           `json:"wrappedKey"`
	Algorithm  KeyWrapAlgorithm `json:"algorithm,omitempty"`
}

// Encrypt encrypts text with XChaCha20-Poly1305 under a 256-bit key
// (POST /v2/encrypt, scope crypto:encrypt).
func (c *Client) Encrypt(ctx context.Context, p EncryptParams) (*EncryptResult, error) {
	return post[EncryptResult](ctx, c, "/v2/encrypt", p)
}

// Decrypt reverses Encrypt (POST /v2/decrypt, scope crypto:decrypt). A
// wrong key or a tampered ciphertext is an *APIError, status 400.
func (c *Client) Decrypt(ctx context.Context, p DecryptParams) (*DecryptResult, error) {
	return post[DecryptResult](ctx, c, "/v2/decrypt", p)
}

// SecretboxSeal seals text with XChaCha20-Poly1305 and optional
// associated data (POST /v2/secretbox/seal, scope crypto:encrypt).
func (c *Client) SecretboxSeal(ctx context.Context, p SecretboxSealParams) (*SecretboxSealResult, error) {
	return post[SecretboxSealResult](ctx, c, "/v2/secretbox/seal", p)
}

// SecretboxOpen opens a secret box and returns the plaintext, which the
// server answers as a bare string (POST /v2/secretbox/open, scope
// crypto:decrypt).
func (c *Client) SecretboxOpen(ctx context.Context, p SecretboxOpenParams) (string, error) {
	return postString(ctx, c, "/v2/secretbox/open", p)
}

// PasswordEncrypt encrypts text under a password with Argon2id and
// XChaCha20-Poly1305 (POST /v2/password/encrypt, scope crypto:encrypt).
func (c *Client) PasswordEncrypt(ctx context.Context, p PasswordEncryptParams) (*PasswordEncryptResult, error) {
	return post[PasswordEncryptResult](ctx, c, "/v2/password/encrypt", p)
}

// PasswordDecrypt reverses PasswordEncrypt and returns the plaintext
// (POST /v2/password/decrypt, scope crypto:decrypt).
func (c *Client) PasswordDecrypt(ctx context.Context, p PasswordDecryptParams) (string, error) {
	return postString(ctx, c, "/v2/password/decrypt", p)
}

// KeyWrap wraps a key with AES-KW or AES-KWP (POST /v2/keys/wrap, scope
// crypto:keys).
func (c *Client) KeyWrap(ctx context.Context, p KeyWrapParams) (*KeyWrapResult, error) {
	return post[KeyWrapResult](ctx, c, "/v2/keys/wrap", p)
}

// KeyUnwrap reverses KeyWrap and returns the key (POST /v2/keys/unwrap,
// scope crypto:keys).
func (c *Client) KeyUnwrap(ctx context.Context, p KeyUnwrapParams) (Hex, error) {
	s, err := postString(ctx, c, "/v2/keys/unwrap", p)
	return Hex(s), err
}
