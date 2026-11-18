// SPDX-License-Identifier: Apache-2.0 OR MIT
// SPDX-FileCopyrightText: 2026 Sebastien Rousseau

package cryptoservice

import "context"

// GeneratedKey is the data of a keygen route: the keyId the private
// key is held under on the server, and the public key.
type GeneratedKey struct {
	KeyID     string `json:"keyId"`
	Algorithm string `json:"algorithm"`
	PublicKey Hex    `json:"publicKey"`
}

// MlKemEncapsulateParams is the body of POST /v2/pq/encapsulate.
type MlKemEncapsulateParams struct {
	// PublicKey is an ML-KEM-768 public key.
	PublicKey Hex `json:"publicKey"`
}

// MlKemEncapsulateResult is the data of POST /v2/pq/encapsulate.
type MlKemEncapsulateResult struct {
	Ciphertext   Hex    `json:"ciphertext"`
	SharedSecret Hex    `json:"sharedSecret"`
	Algorithm    string `json:"algorithm"`
}

// MlKemDecapsulateParams is the body of POST /v2/pq/decapsulate.
type MlKemDecapsulateParams struct {
	KeyID      string `json:"keyId"`
	Ciphertext Hex    `json:"ciphertext"`
}

// KemDecapsulateResult is the data of POST /v2/pq/decapsulate.
type KemDecapsulateResult struct {
	SharedSecret Hex    `json:"sharedSecret"`
	Algorithm    string `json:"algorithm"`
}

// HybridKeyPair is the data of POST /v2/pq/hybrid/keygen.
type HybridKeyPair struct {
	KeyID           string `json:"keyId"`
	Algorithm       string `json:"algorithm"`
	X25519PublicKey Hex    `json:"x25519PublicKey"`
	MlKemPublicKey  Hex    `json:"mlKemPublicKey"`
}

// HybridEncapsulateParams is the body of POST /v2/pq/hybrid/encapsulate.
type HybridEncapsulateParams struct {
	X25519PublicKey Hex `json:"x25519PublicKey"`
	MlKemPublicKey  Hex `json:"mlKemPublicKey"`
}

// HybridEncapsulateResult is the data of POST /v2/pq/hybrid/encapsulate.
type HybridEncapsulateResult struct {
	Algorithm             string `json:"algorithm"`
	Combiner              string `json:"combiner"`
	SharedSecret          Hex    `json:"sharedSecret"`
	X25519EphemeralPublic Hex    `json:"x25519EphemeralPublic"`
	MlKemCiphertext       Hex    `json:"mlKemCiphertext"`
}

// HybridDecapsulateParams is the body of POST /v2/pq/hybrid/decapsulate.
type HybridDecapsulateParams struct {
	KeyID                 string `json:"keyId"`
	X25519EphemeralPublic Hex    `json:"x25519EphemeralPublic"`
	MlKemCiphertext       Hex    `json:"mlKemCiphertext"`
}

// HybridDecapsulateResult is the data of POST /v2/pq/hybrid/decapsulate.
type HybridDecapsulateResult struct {
	Algorithm    string `json:"algorithm"`
	Combiner     string `json:"combiner"`
	SharedSecret Hex    `json:"sharedSecret"`
}

// MlDsaLevel is an ML-DSA parameter set: 44, 65 or 87.
type MlDsaLevel int

// MlDsaKeygenParams is the body of POST /v2/pq/dsa/keygen.
type MlDsaKeygenParams struct {
	Level MlDsaLevel `json:"level"`
}

// MlDsaVerifyParams is the body of POST /v2/pq/dsa/verify.
type MlDsaVerifyParams struct {
	Level     MlDsaLevel `json:"level"`
	PublicKey Hex        `json:"publicKey"`
	Message   string     `json:"message"`
	Signature Hex        `json:"signature"`
}

// SlhDsaVariant is an SLH-DSA parameter set, such as "sha2-128f".
type SlhDsaVariant string

// SlhDsaKeygenParams is the body of POST /v2/pq/slh-dsa/keygen.
type SlhDsaKeygenParams struct {
	Variant SlhDsaVariant `json:"variant"`
}

// SlhDsaVerifyParams is the body of POST /v2/pq/slh-dsa/verify.
type SlhDsaVerifyParams struct {
	Variant   SlhDsaVariant `json:"variant"`
	PublicKey Hex           `json:"publicKey"`
	Message   string        `json:"message"`
	Signature Hex           `json:"signature"`
}

// MlKemGenerateKeyPair creates a server-held ML-KEM-768 key (POST
// /v2/pq/keygen, scope crypto:keys). Stateful.
func (c *Client) MlKemGenerateKeyPair(ctx context.Context) (*GeneratedKey, error) {
	return post[GeneratedKey](ctx, c, "/v2/pq/keygen", empty)
}

// MlKemEncapsulate derives a shared secret and its ML-KEM ciphertext
// for a public key (POST /v2/pq/encapsulate, scope crypto:encrypt).
func (c *Client) MlKemEncapsulate(ctx context.Context, p MlKemEncapsulateParams) (*MlKemEncapsulateResult, error) {
	return post[MlKemEncapsulateResult](ctx, c, "/v2/pq/encapsulate", p)
}

// MlKemDecapsulate recovers the shared secret with a server-held key
// (POST /v2/pq/decapsulate, scope crypto:decrypt).
func (c *Client) MlKemDecapsulate(ctx context.Context, p MlKemDecapsulateParams) (*KemDecapsulateResult, error) {
	return post[KemDecapsulateResult](ctx, c, "/v2/pq/decapsulate", p)
}

// PqGenerateKeyPair creates a server-held X25519 + ML-KEM-768 hybrid
// key (POST /v2/pq/hybrid/keygen, scope crypto:keys). Stateful.
func (c *Client) PqGenerateKeyPair(ctx context.Context) (*HybridKeyPair, error) {
	return post[HybridKeyPair](ctx, c, "/v2/pq/hybrid/keygen", empty)
}

// PqEncapsulate derives a hybrid shared secret for a public key pair
// (POST /v2/pq/hybrid/encapsulate, scope crypto:encrypt).
func (c *Client) PqEncapsulate(ctx context.Context, p HybridEncapsulateParams) (*HybridEncapsulateResult, error) {
	return post[HybridEncapsulateResult](ctx, c, "/v2/pq/hybrid/encapsulate", p)
}

// PqDecapsulate recovers a hybrid shared secret with a server-held key
// (POST /v2/pq/hybrid/decapsulate, scope crypto:decrypt).
func (c *Client) PqDecapsulate(ctx context.Context, p HybridDecapsulateParams) (*HybridDecapsulateResult, error) {
	return post[HybridDecapsulateResult](ctx, c, "/v2/pq/hybrid/decapsulate", p)
}

// PqSignKeygen creates a server-held ML-DSA key (POST
// /v2/pq/dsa/keygen, scope crypto:keys). Stateful.
func (c *Client) PqSignKeygen(ctx context.Context, p MlDsaKeygenParams) (*GeneratedKey, error) {
	return post[GeneratedKey](ctx, c, "/v2/pq/dsa/keygen", p)
}

// PqSign signs with a server-held ML-DSA key; the level comes from the
// key (POST /v2/pq/dsa/sign, scope crypto:sign).
func (c *Client) PqSign(ctx context.Context, p SignParams) (*SignResult, error) {
	return post[SignResult](ctx, c, "/v2/pq/dsa/sign", p)
}

// PqVerify checks an ML-DSA signature (POST /v2/pq/dsa/verify, scope
// crypto:verify).
func (c *Client) PqVerify(ctx context.Context, p MlDsaVerifyParams) (*VerifyResult, error) {
	return post[VerifyResult](ctx, c, "/v2/pq/dsa/verify", p)
}

// PqHashSignKeygen creates a server-held SLH-DSA key (POST
// /v2/pq/slh-dsa/keygen, scope crypto:keys). Stateful.
func (c *Client) PqHashSignKeygen(ctx context.Context, p SlhDsaKeygenParams) (*GeneratedKey, error) {
	return post[GeneratedKey](ctx, c, "/v2/pq/slh-dsa/keygen", p)
}

// PqHashSign signs with a server-held SLH-DSA key; the variant comes
// from the key (POST /v2/pq/slh-dsa/sign, scope crypto:sign).
func (c *Client) PqHashSign(ctx context.Context, p SignParams) (*SignResult, error) {
	return post[SignResult](ctx, c, "/v2/pq/slh-dsa/sign", p)
}

// PqHashVerify checks an SLH-DSA signature (POST /v2/pq/slh-dsa/verify,
// scope crypto:verify).
func (c *Client) PqHashVerify(ctx context.Context, p SlhDsaVerifyParams) (*VerifyResult, error) {
	return post[VerifyResult](ctx, c, "/v2/pq/slh-dsa/verify", p)
}
