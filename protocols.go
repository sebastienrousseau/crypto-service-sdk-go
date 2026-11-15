// SPDX-License-Identifier: Apache-2.0 OR MIT

package cryptoservice

import "context"

// HpkeKem names an HPKE key-encapsulation mechanism.
type HpkeKem string

// The KEMs of /v2/hpke/keygen and /v2/hpke/seal.
const (
	HpkeX25519         HpkeKem = "x25519"
	HpkeP256           HpkeKem = "p256"
	HpkeXWing          HpkeKem = "x-wing" // the server's default
	HpkeX25519MlKem768 HpkeKem = "x25519-ml-kem-768"
)

// HpkeKeygenParams is the body of POST /v2/hpke/keygen.
type HpkeKeygenParams struct {
	KEM HpkeKem `json:"kem,omitempty"`
}

// HpkeSealParams is the body of POST /v2/hpke/seal.
//
// The server reads Plaintext, Info and AAD as hex when the string is
// even-length hexadecimal, and as UTF-8 text otherwise. The client
// sends them unchanged and does not guess: text that happens to look
// like hex ("cafe") is sealed as the bytes it spells. Send hex (HexOf)
// whenever the input could be mistaken for it.
type HpkeSealParams struct {
	RecipientPublicKey Hex     `json:"recipientPublicKey"`
	Plaintext          string  `json:"plaintext"`
	KEM                HpkeKem `json:"kem,omitempty"`
	// AEAD is "chacha20-poly1305" or "aes-128-gcm".
	AEAD  string `json:"aead,omitempty"`
	Info  string `json:"info,omitempty"`
	AAD   string `json:"aad,omitempty"`
	PSK   string `json:"psk,omitempty"`
	PSKID string `json:"pskId,omitempty"`
}

// HpkeSealResult is the data of POST /v2/hpke/seal.
type HpkeSealResult struct {
	EncapsulatedKey Hex `json:"encapsulatedKey"`
	Ciphertext      Hex `json:"ciphertext"`
}

// HpkeOpenParams is the body of POST /v2/hpke/open; Info, AAD, PSK and
// PSKID must repeat what HpkeSeal was given.
type HpkeOpenParams struct {
	KeyID           string `json:"keyId"`
	EncapsulatedKey Hex    `json:"encapsulatedKey"`
	Ciphertext      Hex    `json:"ciphertext"`
	AEAD            string `json:"aead,omitempty"`
	Info            string `json:"info,omitempty"`
	AAD             string `json:"aad,omitempty"`
	PSK             string `json:"psk,omitempty"`
	PSKID           string `json:"pskId,omitempty"`
}

// OpenedText is a decrypted plaintext in both of the server's forms:
// UTF-8 text (invalid sequences become U+FFFD) and the exact bytes.
type OpenedText struct {
	Plaintext string `json:"plaintext"`
	Hex       Hex    `json:"hex"`
}

// SignedEnvelopeKey is one signature of an ISO 20022 envelope.
type SignedEnvelopeKey struct {
	Algorithm string `json:"algorithm"`
	PublicKey Hex    `json:"publicKey"`
	Signature Hex    `json:"signature"`
}

// Iso20022Envelope is a dual-signed (classical and post-quantum)
// envelope over an ISO 20022 payload digest.
type Iso20022Envelope struct {
	MessageID string `json:"messageId"`
	// MessageType is pacs.008, pain.001, camt.053 or generic.
	MessageType     string            `json:"messageType"`
	PayloadDigest   Hex               `json:"payloadDigest"`
	DigestAlgorithm HashAlgorithm     `json:"digestAlgorithm"`
	Timestamp       string            `json:"timestamp"`
	Classical       SignedEnvelopeKey `json:"classical"`
	PostQuantum     SignedEnvelopeKey `json:"postQuantum"`
}

// Iso20022TrustedKeys are the keys the envelope must be signed with.
type Iso20022TrustedKeys struct {
	ClassicalPublicKey   Hex `json:"classicalPublicKey"`
	PostQuantumPublicKey Hex `json:"postQuantumPublicKey"`
}

// Iso20022VerifyParams is the body of POST /v2/stream/iso20022.
type Iso20022VerifyParams struct {
	Envelope Iso20022Envelope `json:"envelope"`
	// Payload is any JSON value; the server digests its canonical form.
	Payload     any                 `json:"payload"`
	TrustedKeys Iso20022TrustedKeys `json:"trustedKeys"`
}

// Iso20022VerifyResult is the data of POST /v2/stream/iso20022.
type Iso20022VerifyResult struct {
	Valid            bool   `json:"valid"`
	ClassicalValid   bool   `json:"classicalValid"`
	PostQuantumValid bool   `json:"postQuantumValid"`
	DigestMatches    bool   `json:"digestMatches"`
	MessageID        string `json:"messageId"`
}

// OpaqueSuite names an OPAQUE cipher suite.
type OpaqueSuite string

// OpaqueRegisterInitParams is the body of POST /v2/opaque/register/init.
type OpaqueRegisterInitParams struct {
	CredentialIdentifier string      `json:"credentialIdentifier"`
	Request              Hex         `json:"request"`
	Suite                OpaqueSuite `json:"suite,omitempty"`
}

// OpaqueRegisterInitResult is the data of POST /v2/opaque/register/init.
type OpaqueRegisterInitResult struct {
	Response Hex `json:"response"`
}

// OpaqueRegisterFinishParams is the body of POST
// /v2/opaque/register/finish.
type OpaqueRegisterFinishParams struct {
	CredentialIdentifier string      `json:"credentialIdentifier"`
	Record               Hex         `json:"record"`
	Suite                OpaqueSuite `json:"suite,omitempty"`
}

// OpaqueRegisterFinishResult is the data of POST
// /v2/opaque/register/finish.
type OpaqueRegisterFinishResult struct {
	CredentialIdentifier string `json:"credentialIdentifier"`
	Status               string `json:"status"`
}

// OpaqueLoginInitParams is the body of POST /v2/opaque/login/init.
type OpaqueLoginInitParams struct {
	CredentialIdentifier string      `json:"credentialIdentifier"`
	KE1                  Hex         `json:"ke1"`
	Suite                OpaqueSuite `json:"suite,omitempty"`
}

// OpaqueLoginInitResult is the data of POST /v2/opaque/login/init.
type OpaqueLoginInitResult struct {
	KE2       Hex    `json:"ke2"`
	SessionID string `json:"sessionId"`
}

// OpaqueLoginFinishParams is the body of POST /v2/opaque/login/finish.
type OpaqueLoginFinishParams struct {
	SessionID string `json:"sessionId"`
	KE3       Hex    `json:"ke3"`
}

// OpaqueLoginFinishResult is the data of POST /v2/opaque/login/finish.
type OpaqueLoginFinishResult struct {
	SessionKey Hex `json:"sessionKey"`
}

// HpkeGenerateKeyPair creates a server-held HPKE key (POST
// /v2/hpke/keygen, scope crypto:keys). Stateful.
func (c *Client) HpkeGenerateKeyPair(ctx context.Context, p HpkeKeygenParams) (*GeneratedKey, error) {
	return post[GeneratedKey](ctx, c, "/v2/hpke/keygen", p)
}

// HpkeSeal encrypts to an HPKE public key (POST /v2/hpke/seal, scope
// crypto:encrypt). See HpkeSealParams on how the server reads inputs.
func (c *Client) HpkeSeal(ctx context.Context, p HpkeSealParams) (*HpkeSealResult, error) {
	return post[HpkeSealResult](ctx, c, "/v2/hpke/seal", p)
}

// HpkeOpen decrypts with a server-held HPKE key (POST /v2/hpke/open,
// scope crypto:decrypt).
func (c *Client) HpkeOpen(ctx context.Context, p HpkeOpenParams) (*OpenedText, error) {
	return post[OpenedText](ctx, c, "/v2/hpke/open", p)
}

// StreamIso20022 verifies a dual-signed ISO 20022 envelope (POST
// /v2/stream/iso20022, scope crypto:verify).
func (c *Client) StreamIso20022(ctx context.Context, p Iso20022VerifyParams) (*Iso20022VerifyResult, error) {
	return post[Iso20022VerifyResult](ctx, c, "/v2/stream/iso20022", p)
}

// OpaqueRegisterInit starts an OPAQUE registration for the caller's own
// credential (POST /v2/opaque/register/init, authenticated). Stateful.
func (c *Client) OpaqueRegisterInit(ctx context.Context, p OpaqueRegisterInitParams) (*OpaqueRegisterInitResult, error) {
	return post[OpaqueRegisterInitResult](ctx, c, "/v2/opaque/register/init", p)
}

// OpaqueRegisterFinish stores an OPAQUE registration record (POST
// /v2/opaque/register/finish, authenticated). Stateful.
func (c *Client) OpaqueRegisterFinish(ctx context.Context, p OpaqueRegisterFinishParams) (*OpaqueRegisterFinishResult, error) {
	return post[OpaqueRegisterFinishResult](ctx, c, "/v2/opaque/register/finish", p)
}

// OpaqueLoginInit starts an OPAQUE login (POST /v2/opaque/login/init,
// no credentials required).
func (c *Client) OpaqueLoginInit(ctx context.Context, p OpaqueLoginInitParams) (*OpaqueLoginInitResult, error) {
	return post[OpaqueLoginInitResult](ctx, c, "/v2/opaque/login/init", p)
}

// OpaqueLoginFinish completes an OPAQUE login and returns the session
// key (POST /v2/opaque/login/finish, no credentials required).
func (c *Client) OpaqueLoginFinish(ctx context.Context, p OpaqueLoginFinishParams) (*OpaqueLoginFinishResult, error) {
	return post[OpaqueLoginFinishResult](ctx, c, "/v2/opaque/login/finish", p)
}
