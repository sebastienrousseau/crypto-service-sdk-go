// SPDX-License-Identifier: Apache-2.0 OR MIT
// SPDX-FileCopyrightText: 2026 Sebastien Rousseau

package cryptoservice

import "context"

// HashAlgorithm names a digest. The constants are the server's set
// today; any other string converts, so a newer server's algorithms
// work without a new client.
type HashAlgorithm string

// The digests of /v2/hash. The HMAC ones are the SHA-2 and SHA-3 ones.
const (
	SHA256  HashAlgorithm = "sha256"
	SHA384  HashAlgorithm = "sha384"
	SHA512  HashAlgorithm = "sha512"
	SHA3256 HashAlgorithm = "sha3-256"
	SHA3512 HashAlgorithm = "sha3-512"
	BLAKE2b HashAlgorithm = "blake2b"
	BLAKE3  HashAlgorithm = "blake3"
)

// HashParams is the body of POST /v2/hash.
type HashParams struct {
	Algorithm HashAlgorithm `json:"algorithm"`
	// Data is hashed as its UTF-8 bytes.
	Data string `json:"data"`
}

// HashResult is the data of POST /v2/hash.
type HashResult struct {
	Digest    Hex           `json:"digest"`
	Algorithm HashAlgorithm `json:"algorithm"`
	// Length is the digest length in bytes.
	Length int `json:"length"`
}

// MacParams is the body of POST /v2/hmac.
type MacParams struct {
	Algorithm HashAlgorithm `json:"algorithm"`
	Key       Hex           `json:"key"`
	Data      string        `json:"data"`
}

// MacResult is the data of POST /v2/hmac.
type MacResult struct {
	Mac       Hex           `json:"mac"`
	Algorithm HashAlgorithm `json:"algorithm"`
}

// MacVerifyParams is the body of POST /v2/hmac/verify.
type MacVerifyParams struct {
	Algorithm HashAlgorithm `json:"algorithm"`
	Key       Hex           `json:"key"`
	Data      string        `json:"data"`
	Mac       Hex           `json:"mac"`
}

// MacVerifyResult is the data of POST /v2/hmac/verify.
type MacVerifyResult struct {
	Valid     bool          `json:"valid"`
	Algorithm HashAlgorithm `json:"algorithm"`
}

// KdfAlgorithm names a key-derivation function.
type KdfAlgorithm string

// The functions of /v2/kdf.
const (
	Scrypt       KdfAlgorithm = "scrypt"
	HKDFSHA256   KdfAlgorithm = "hkdf-sha256"
	PBKDF2SHA256 KdfAlgorithm = "pbkdf2-sha256"
)

// KdfParams is the body of POST /v2/kdf.
type KdfParams struct {
	Algorithm KdfAlgorithm `json:"algorithm"`
	Password  string       `json:"password"`
	// Salt is optional; the server draws 16 random bytes without it.
	Salt Hex `json:"salt,omitempty"`
	// KeyLength is the derived key's length in bytes (16 to 64).
	KeyLength int `json:"keyLength,omitempty"`
	// Params tunes the function; the server enforces OWASP floors.
	Params *KdfOptions `json:"params,omitempty"`
}

// KdfOptions are the per-function parameters of KdfParams.
type KdfOptions struct {
	// N is the scrypt cost (the server accepts 131072 only).
	N int `json:"N,omitempty"`
	// R is the scrypt block size (8 only).
	R int `json:"r,omitempty"`
	// P is the scrypt parallelism (1 to 4).
	P int `json:"p,omitempty"`
	// Iterations is the PBKDF2 count (600000 to 1000000).
	Iterations int `json:"iterations,omitempty"`
	// Info is the HKDF info string.
	Info string `json:"info,omitempty"`
}

// KdfResult is the data of POST /v2/kdf.
type KdfResult struct {
	DerivedKey Hex          `json:"derivedKey"`
	Salt       Hex          `json:"salt"`
	Algorithm  KdfAlgorithm `json:"algorithm"`
	KeyLength  int          `json:"keyLength"`
}

// Argon2Params are Argon2id costs.
type Argon2Params struct {
	// T is the time cost (passes).
	T int `json:"t"`
	// M is the memory cost in KiB.
	M int `json:"m"`
	// P is the parallelism.
	P int `json:"p"`
}

// PasswordHashParams is the body of POST /v2/password/hash. Zero costs
// take the server's defaults; it requires at least two passes and
// 19456 KiB.
type PasswordHashParams struct {
	Password    string `json:"password"`
	TimeCost    int    `json:"timeCost,omitempty"`
	MemoryCost  int    `json:"memoryCost,omitempty"`
	Parallelism int    `json:"parallelism,omitempty"`
}

// PasswordHashResult is the data of POST /v2/password/hash.
type PasswordHashResult struct {
	Hash      Hex          `json:"hash"`
	Salt      Hex          `json:"salt"`
	Params    Argon2Params `json:"params"`
	Algorithm string       `json:"algorithm"`
	// PHC is the hash in PHC string format ($argon2id$v=19$...).
	PHC string `json:"phc"`
}

// PasswordVerifyParams is the body of POST /v2/password/verify.
type PasswordVerifyParams struct {
	Password string       `json:"password"`
	Hash     Hex          `json:"hash"`
	Salt     Hex          `json:"salt"`
	Params   Argon2Params `json:"params"`
}

// PasswordVerifyResult is the data of POST /v2/password/verify.
type PasswordVerifyResult struct {
	Valid bool `json:"valid"`
}

// Hash digests data (POST /v2/hash, scope crypto:hash).
func (c *Client) Hash(ctx context.Context, p HashParams) (*HashResult, error) {
	return post[HashResult](ctx, c, "/v2/hash", p)
}

// Mac computes an HMAC (POST /v2/hmac, scope crypto:sign).
func (c *Client) Mac(ctx context.Context, p MacParams) (*MacResult, error) {
	return post[MacResult](ctx, c, "/v2/hmac", p)
}

// MacVerify checks an HMAC (POST /v2/hmac/verify, scope crypto:verify).
// A wrong MAC is Valid false, not an error.
func (c *Client) MacVerify(ctx context.Context, p MacVerifyParams) (*MacVerifyResult, error) {
	return post[MacVerifyResult](ctx, c, "/v2/hmac/verify", p)
}

// Kdf derives a key from a password (POST /v2/kdf, scope crypto:kdf).
func (c *Client) Kdf(ctx context.Context, p KdfParams) (*KdfResult, error) {
	return post[KdfResult](ctx, c, "/v2/kdf", p)
}

// PasswordHash hashes a password with Argon2id (POST /v2/password/hash,
// scope crypto:kdf).
func (c *Client) PasswordHash(ctx context.Context, p PasswordHashParams) (*PasswordHashResult, error) {
	return post[PasswordHashResult](ctx, c, "/v2/password/hash", p)
}

// PasswordVerify checks a password against an Argon2id hash
// (POST /v2/password/verify, scope crypto:kdf).
func (c *Client) PasswordVerify(ctx context.Context, p PasswordVerifyParams) (*PasswordVerifyResult, error) {
	return post[PasswordVerifyResult](ctx, c, "/v2/password/verify", p)
}
