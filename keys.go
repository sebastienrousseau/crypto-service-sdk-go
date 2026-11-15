// SPDX-License-Identifier: Apache-2.0 OR MIT

package cryptoservice

import "context"

// KeyAlgorithm names a server-held key type of /v2/keys/generate.
type KeyAlgorithm string

// The algorithms /v2/keys/generate accepts.
const (
	Ed25519   KeyAlgorithm = "ed25519" // the server's default
	X25519    KeyAlgorithm = "x25519"
	Ed448     KeyAlgorithm = "ed448"
	X448      KeyAlgorithm = "x448"
	P256      KeyAlgorithm = "p256"
	P384      KeyAlgorithm = "p384"
	MLKEM512  KeyAlgorithm = "ml-kem-512"
	MLKEM768  KeyAlgorithm = "ml-kem-768"
	MLKEM1024 KeyAlgorithm = "ml-kem-1024"
	MLDSA44   KeyAlgorithm = "ml-dsa-44"
	MLDSA65   KeyAlgorithm = "ml-dsa-65"
	MLDSA87   KeyAlgorithm = "ml-dsa-87"
)

// KeyMetadata is the JOSE-style metadata of a server-held key.
type KeyMetadata struct {
	// Kid is the key's caller-chosen identifier.
	Kid string `json:"kid,omitempty"`
	// Use is "sig" or "enc".
	Use string `json:"use,omitempty"`
	// Exp is an expiry timestamp.
	Exp string `json:"exp,omitempty"`
}

// GenerateKeyPairParams is the body of POST /v2/keys/generate. The
// zero value asks for the server's default, an Ed25519 key.
type GenerateKeyPairParams struct {
	Algorithm KeyAlgorithm `json:"algorithm,omitempty"`
	Metadata  *KeyMetadata `json:"metadata,omitempty"`
}

// KeyPair is the data of POST /v2/keys/generate: the public half and
// the keyId the private half is held under on the server.
type KeyPair struct {
	KeyID     string       `json:"keyId"`
	Algorithm KeyAlgorithm `json:"algorithm"`
	PublicKey Hex          `json:"publicKey"`
	Kid       string       `json:"kid"`
	Metadata  KeyMetadata  `json:"metadata"`
}

// KeyIDParams names a server-held key, as returned by a key-generation
// method ("k_" and 22 base64url characters).
type KeyIDParams struct {
	KeyID string `json:"keyId"`
}

// ExportedKey is the data of POST /v2/keys/export. Which members are
// set depends on the key's algorithm; the private ones are secrets.
type ExportedKey struct {
	KeyID            string `json:"keyId"`
	Algorithm        string `json:"algorithm"`
	PublicKey        Hex    `json:"publicKey,omitempty"`
	PrivateKey       Hex    `json:"privateKey,omitempty"`
	SecretKey        Hex    `json:"secretKey,omitempty"`
	X25519PublicKey  Hex    `json:"x25519PublicKey,omitempty"`
	X25519PrivateKey Hex    `json:"x25519PrivateKey,omitempty"`
	MlKemPublicKey   Hex    `json:"mlKemPublicKey,omitempty"`
	MlKemSecretKey   Hex    `json:"mlKemSecretKey,omitempty"`
}

// KeyDeleted is the data of POST /v2/keys/delete.
type KeyDeleted struct {
	Deleted bool   `json:"deleted"`
	KeyID   string `json:"keyId"`
}

// SignParams is the body of the signing routes: a server-held key and
// the message, signed as its UTF-8 bytes.
type SignParams struct {
	KeyID   string `json:"keyId"`
	Message string `json:"message"`
}

// SignResult is the data of a signing route.
type SignResult struct {
	Signature Hex    `json:"signature"`
	Algorithm string `json:"algorithm"`
}

// VerifyParams is the body of POST /v2/verify (Ed25519).
type VerifyParams struct {
	// PublicKey is 32 bytes (64 hex characters).
	PublicKey Hex    `json:"publicKey"`
	Message   string `json:"message"`
	// Signature is 64 bytes (128 hex characters).
	Signature Hex `json:"signature"`
}

// VerifyResult is the data of a verification route. A bad signature is
// Valid false, not an error.
type VerifyResult struct {
	Valid     bool   `json:"valid"`
	Algorithm string `json:"algorithm"`
}

// StreamSignItem is one message of a batch signature request.
type StreamSignItem struct {
	ID      string `json:"id"`
	KeyID   string `json:"keyId"`
	Message string `json:"message"`
}

// StreamSignParams is the body of POST /v2/stream/sign.
type StreamSignParams struct {
	Items []StreamSignItem `json:"items"`
}

// StreamSignature is one signature of a batch.
type StreamSignature struct {
	ID        string `json:"id"`
	Signature Hex    `json:"signature"`
	Algorithm string `json:"algorithm"`
}

// StreamSignResult is the data of POST /v2/stream/sign.
type StreamSignResult struct {
	Count      int               `json:"count"`
	Signatures []StreamSignature `json:"signatures"`
}

// StreamVerifyItem is one Ed25519 signature of a batch to verify.
type StreamVerifyItem struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	PublicKey Hex    `json:"publicKey"`
	Signature Hex    `json:"signature"`
}

// StreamVerifyParams is the body of POST /v2/stream/verify.
type StreamVerifyParams struct {
	Items []StreamVerifyItem `json:"items"`
}

// StreamVerification is one item's outcome.
type StreamVerification struct {
	ID    string `json:"id"`
	Valid bool   `json:"valid"`
}

// StreamVerifyResult is the data of POST /v2/stream/verify.
type StreamVerifyResult struct {
	AllValid   bool                 `json:"allValid"`
	Count      int                  `json:"count"`
	ValidCount int                  `json:"validCount"`
	Results    []StreamVerification `json:"results"`
}

// GenerateKeyPair creates a server-held key and returns its keyId and
// public key; the private key stays on the server (POST
// /v2/keys/generate, scope crypto:keys). Stateful: retried only after
// 429 or 503.
func (c *Client) GenerateKeyPair(ctx context.Context, p GenerateKeyPairParams) (*KeyPair, error) {
	return post[KeyPair](ctx, c, "/v2/keys/generate", p)
}

// ExportKey returns a server-held key, private parts included (POST
// /v2/keys/export). It needs the crypto:keys:export scope, which
// crypto:admin does not imply.
func (c *Client) ExportKey(ctx context.Context, p KeyIDParams) (*ExportedKey, error) {
	return post[ExportedKey](ctx, c, "/v2/keys/export", p)
}

// KeysDelete deletes a server-held key (POST /v2/keys/delete, scope
// crypto:keys). Stateful: retried only after 429 or 503.
func (c *Client) KeysDelete(ctx context.Context, p KeyIDParams) (*KeyDeleted, error) {
	return post[KeyDeleted](ctx, c, "/v2/keys/delete", p)
}

// Sign signs a message with a server-held Ed25519 key (POST /v2/sign,
// scope crypto:sign).
func (c *Client) Sign(ctx context.Context, p SignParams) (*SignResult, error) {
	return post[SignResult](ctx, c, "/v2/sign", p)
}

// Verify checks an Ed25519 signature against a public key (POST
// /v2/verify, scope crypto:verify).
func (c *Client) Verify(ctx context.Context, p VerifyParams) (*VerifyResult, error) {
	return post[VerifyResult](ctx, c, "/v2/verify", p)
}

// StreamSign signs a batch of messages with server-held Ed25519 keys
// (POST /v2/stream/sign, scope crypto:sign).
func (c *Client) StreamSign(ctx context.Context, p StreamSignParams) (*StreamSignResult, error) {
	return post[StreamSignResult](ctx, c, "/v2/stream/sign", p)
}

// StreamVerify checks a batch of Ed25519 signatures (POST
// /v2/stream/verify, scope crypto:verify).
func (c *Client) StreamVerify(ctx context.Context, p StreamVerifyParams) (*StreamVerifyResult, error) {
	return post[StreamVerifyResult](ctx, c, "/v2/stream/verify", p)
}
