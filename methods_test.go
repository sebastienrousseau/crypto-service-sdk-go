// SPDX-License-Identifier: Apache-2.0 OR MIT
// SPDX-FileCopyrightText: 2026 Sebastien Rousseau

package cryptoservice

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// exchange is what the fake server saw for one call.
type exchange struct {
	method, path string
	body         map[string]any
	header       http.Header
}

// fakeServer answers every request with status and body, recording
// what it received.
func fakeServer(t *testing.T, status int, body string) (*Client, *exchange) {
	t.Helper()
	seen := &exchange{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.method, seen.path, seen.header = r.Method, r.URL.Path, r.Header.Clone()
		raw, _ := io.ReadAll(r.Body)
		seen.body = nil
		_ = json.Unmarshal(raw, &seen.body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)
	c, err := NewClient(srv.URL, WithAPIKey("test-key"))
	if err != nil {
		t.Fatal(err)
	}
	return c, seen
}

// call is one method under test, adapted to a common shape.
type call struct {
	path string
	run  func(context.Context, *Client) (any, error)
}

func p[T any](v *T, err error) (any, error) { return v, err }

func methodCalls() []call {
	return append(append(symmetricCalls(), keyCalls()...), protocolCalls()...)
}

func symmetricCalls() []call {
	return []call{
		{"/v2/encrypt", func(ctx context.Context, c *Client) (any, error) { return p(c.Encrypt(ctx, EncryptParams{})) }},
		{"/v2/decrypt", func(ctx context.Context, c *Client) (any, error) { return p(c.Decrypt(ctx, DecryptParams{})) }},
		{"/v2/secretbox/seal", func(ctx context.Context, c *Client) (any, error) {
			return p(c.SecretboxSeal(ctx, SecretboxSealParams{}))
		}},
		{"/v2/password/encrypt", func(ctx context.Context, c *Client) (any, error) {
			return p(c.PasswordEncrypt(ctx, PasswordEncryptParams{}))
		}},
		{"/v2/keys/wrap", func(ctx context.Context, c *Client) (any, error) { return p(c.KeyWrap(ctx, KeyWrapParams{})) }},
		{"/v2/hash", func(ctx context.Context, c *Client) (any, error) { return p(c.Hash(ctx, HashParams{})) }},
		{"/v2/hmac", func(ctx context.Context, c *Client) (any, error) { return p(c.Mac(ctx, MacParams{})) }},
		{"/v2/hmac/verify", func(ctx context.Context, c *Client) (any, error) { return p(c.MacVerify(ctx, MacVerifyParams{})) }},
		{"/v2/kdf", func(ctx context.Context, c *Client) (any, error) { return p(c.Kdf(ctx, KdfParams{})) }},
		{"/v2/password/hash", func(ctx context.Context, c *Client) (any, error) { return p(c.PasswordHash(ctx, PasswordHashParams{})) }},
		{"/v2/password/verify", func(ctx context.Context, c *Client) (any, error) {
			return p(c.PasswordVerify(ctx, PasswordVerifyParams{}))
		}},
		{"/v2/kms/create-key", func(ctx context.Context, c *Client) (any, error) { return p(c.KmsCreateKey(ctx, KmsCreateKeyParams{})) }},
		{"/v2/kms/wrap", func(ctx context.Context, c *Client) (any, error) { return p(c.KmsWrap(ctx, KmsWrapParams{})) }},
		{"/v2/kms/unwrap", func(ctx context.Context, c *Client) (any, error) { return p(c.KmsUnwrap(ctx, KmsUnwrapParams{})) }},
		{"/v2/kms/generate-data-key", func(ctx context.Context, c *Client) (any, error) {
			return p(c.KmsGenerateDataKey(ctx, KmsGenerateDataKeyParams{}))
		}},
		{"/v2/kms/encrypt", func(ctx context.Context, c *Client) (any, error) { return p(c.KmsEncrypt(ctx, KmsEncryptParams{})) }},
		{"/v2/kms/decrypt", func(ctx context.Context, c *Client) (any, error) { return p(c.KmsDecrypt(ctx, KmsDecryptParams{})) }},
	}
}

func keyCalls() []call {
	return []call{
		{"/v2/keys/generate", func(ctx context.Context, c *Client) (any, error) {
			return p(c.GenerateKeyPair(ctx, GenerateKeyPairParams{}))
		}},
		{"/v2/keys/export", func(ctx context.Context, c *Client) (any, error) { return p(c.ExportKey(ctx, KeyIDParams{})) }},
		{"/v2/keys/delete", func(ctx context.Context, c *Client) (any, error) { return p(c.KeysDelete(ctx, KeyIDParams{})) }},
		{"/v2/sign", func(ctx context.Context, c *Client) (any, error) { return p(c.Sign(ctx, SignParams{})) }},
		{"/v2/verify", func(ctx context.Context, c *Client) (any, error) { return p(c.Verify(ctx, VerifyParams{})) }},
		{"/v2/stream/sign", func(ctx context.Context, c *Client) (any, error) { return p(c.StreamSign(ctx, StreamSignParams{})) }},
		{"/v2/stream/verify", func(ctx context.Context, c *Client) (any, error) { return p(c.StreamVerify(ctx, StreamVerifyParams{})) }},
		{"/v2/pq/keygen", func(ctx context.Context, c *Client) (any, error) { return p(c.MlKemGenerateKeyPair(ctx)) }},
		{"/v2/pq/encapsulate", func(ctx context.Context, c *Client) (any, error) {
			return p(c.MlKemEncapsulate(ctx, MlKemEncapsulateParams{}))
		}},
		{"/v2/pq/decapsulate", func(ctx context.Context, c *Client) (any, error) {
			return p(c.MlKemDecapsulate(ctx, MlKemDecapsulateParams{}))
		}},
		{"/v2/pq/hybrid/keygen", func(ctx context.Context, c *Client) (any, error) { return p(c.PqGenerateKeyPair(ctx)) }},
		{"/v2/pq/hybrid/encapsulate", func(ctx context.Context, c *Client) (any, error) {
			return p(c.PqEncapsulate(ctx, HybridEncapsulateParams{}))
		}},
		{"/v2/pq/hybrid/decapsulate", func(ctx context.Context, c *Client) (any, error) {
			return p(c.PqDecapsulate(ctx, HybridDecapsulateParams{}))
		}},
		{"/v2/pq/dsa/keygen", func(ctx context.Context, c *Client) (any, error) { return p(c.PqSignKeygen(ctx, MlDsaKeygenParams{})) }},
		{"/v2/pq/dsa/sign", func(ctx context.Context, c *Client) (any, error) { return p(c.PqSign(ctx, SignParams{})) }},
		{"/v2/pq/dsa/verify", func(ctx context.Context, c *Client) (any, error) { return p(c.PqVerify(ctx, MlDsaVerifyParams{})) }},
		{"/v2/pq/slh-dsa/keygen", func(ctx context.Context, c *Client) (any, error) {
			return p(c.PqHashSignKeygen(ctx, SlhDsaKeygenParams{}))
		}},
		{"/v2/pq/slh-dsa/sign", func(ctx context.Context, c *Client) (any, error) { return p(c.PqHashSign(ctx, SignParams{})) }},
		{"/v2/pq/slh-dsa/verify", func(ctx context.Context, c *Client) (any, error) {
			return p(c.PqHashVerify(ctx, SlhDsaVerifyParams{}))
		}},
	}
}

func protocolCalls() []call {
	return []call{
		{"/v2/sealedbox/seal", func(ctx context.Context, c *Client) (any, error) {
			return p(c.SealedboxSeal(ctx, SealedboxSealParams{}))
		}},
		{"/v2/sealedbox/seal-pq", func(ctx context.Context, c *Client) (any, error) {
			return p(c.SealedboxSealPq(ctx, SealedboxSealPqParams{}))
		}},
		{"/v2/multi-recipient/encrypt", func(ctx context.Context, c *Client) (any, error) {
			return p(c.MultiRecipientEncrypt(ctx, MultiRecipientParams{}))
		}},
		{"/v2/stream/pq-encrypt", func(ctx context.Context, c *Client) (any, error) {
			return p(c.StreamPqEncrypt(ctx, StreamPqEncryptParams{}))
		}},
		{"/v2/stream/pq-decrypt", func(ctx context.Context, c *Client) (any, error) {
			return p(c.StreamPqDecrypt(ctx, StreamPqDecryptParams{}))
		}},
		{"/v2/stream/multi-pq-encrypt", func(ctx context.Context, c *Client) (any, error) {
			return p(c.StreamMultiPqEncrypt(ctx, StreamMultiPqEncryptParams{}))
		}},
		{"/v2/stream/multi-pq-decrypt", func(ctx context.Context, c *Client) (any, error) {
			return p(c.StreamMultiPqDecrypt(ctx, StreamMultiPqDecryptParams{}))
		}},
		{"/v2/stream/iso20022", func(ctx context.Context, c *Client) (any, error) {
			return p(c.StreamIso20022(ctx, Iso20022VerifyParams{}))
		}},
		{"/v2/hpke/keygen", func(ctx context.Context, c *Client) (any, error) {
			return p(c.HpkeGenerateKeyPair(ctx, HpkeKeygenParams{}))
		}},
		{"/v2/hpke/seal", func(ctx context.Context, c *Client) (any, error) { return p(c.HpkeSeal(ctx, HpkeSealParams{})) }},
		{"/v2/hpke/open", func(ctx context.Context, c *Client) (any, error) { return p(c.HpkeOpen(ctx, HpkeOpenParams{})) }},
		{"/v2/opaque/register/init", func(ctx context.Context, c *Client) (any, error) {
			return p(c.OpaqueRegisterInit(ctx, OpaqueRegisterInitParams{}))
		}},
		{"/v2/opaque/register/finish", func(ctx context.Context, c *Client) (any, error) {
			return p(c.OpaqueRegisterFinish(ctx, OpaqueRegisterFinishParams{}))
		}},
		{"/v2/opaque/login/init", func(ctx context.Context, c *Client) (any, error) {
			return p(c.OpaqueLoginInit(ctx, OpaqueLoginInitParams{}))
		}},
		{"/v2/opaque/login/finish", func(ctx context.Context, c *Client) (any, error) {
			return p(c.OpaqueLoginFinish(ctx, OpaqueLoginFinishParams{}))
		}},
		{"/v2/compliance/dora", func(ctx context.Context, c *Client) (any, error) { return p(c.GetDoraCompliance(ctx)) }},
	}
}

// stringCalls are the routes whose data is a bare string.
func stringCalls() []call {
	return []call{
		{"/v2/secretbox/open", func(ctx context.Context, c *Client) (any, error) { return c.SecretboxOpen(ctx, SecretboxOpenParams{}) }},
		{"/v2/password/decrypt", func(ctx context.Context, c *Client) (any, error) {
			return c.PasswordDecrypt(ctx, PasswordDecryptParams{})
		}},
		{"/v2/sealedbox/open", func(ctx context.Context, c *Client) (any, error) { return c.SealedboxOpen(ctx, SealedboxOpenParams{}) }},
		{"/v2/sealedbox/open-pq", func(ctx context.Context, c *Client) (any, error) {
			return c.SealedboxOpenPq(ctx, SealedboxOpenParams{})
		}},
		{"/v2/keys/unwrap", func(ctx context.Context, c *Client) (any, error) { return c.KeyUnwrap(ctx, KeyUnwrapParams{}) }},
	}
}

func TestEveryMethodCallsItsRoute(t *testing.T) {
	ctx := context.Background()
	c, seen := fakeServer(t, 200, `{"data":{}}`)
	for _, m := range methodCalls() {
		got, err := m.run(ctx, c)
		if err != nil || got == nil {
			t.Errorf("%s: %v, %v", m.path, got, err)
		}
		if seen.path != m.path {
			t.Errorf("called %s, want %s", seen.path, m.path)
		}
	}
	s, _ := fakeServer(t, 200, `{"data":"text"}`)
	for _, m := range stringCalls() {
		got, err := m.run(ctx, s)
		if err != nil || got == nil {
			t.Errorf("%s: %v, %v", m.path, got, err)
		}
	}
}

func TestEveryMethodReturnsTheAPIError(t *testing.T) {
	ctx := context.Background()
	c, _ := fakeServer(t, 404, `{"type":"urn:crypto-service:problem:key-not-found","title":"Key not found","status":404,"detail":"d"}`)
	for _, m := range append(methodCalls(), stringCalls()...) {
		_, err := m.run(ctx, c)
		var apiErr *APIError
		if !errors.As(err, &apiErr) || apiErr.Status != 404 || apiErr.Path != m.path {
			t.Errorf("%s: %v", m.path, err)
		}
	}
}

func TestKmsEncryptAlwaysSendsTheEncoding(t *testing.T) {
	c, seen := fakeServer(t, 200, `{"data":{}}`)
	if _, err := c.KmsEncrypt(context.Background(), KmsEncryptParams{KeyID: "k"}); err != nil {
		t.Fatal(err)
	}
	if seen.body["encoding"] != "utf8" {
		t.Errorf("encoding %v, want utf8", seen.body["encoding"])
	}
	if _, err := c.KmsEncrypt(context.Background(), KmsEncryptParams{KeyID: "k", Encoding: KmsHex}); err != nil {
		t.Fatal(err)
	}
	if seen.body["encoding"] != "hex" {
		t.Errorf("encoding %v, want hex", seen.body["encoding"])
	}
}

func TestKeygenRoutesSendAnEmptyObject(t *testing.T) {
	c, seen := fakeServer(t, 200, `{"data":{}}`)
	if _, err := c.PqGenerateKeyPair(context.Background()); err != nil {
		t.Fatal(err)
	}
	if seen.body == nil || len(seen.body) != 0 || seen.header.Get("Content-Type") != "application/json" {
		t.Errorf("body %v, content type %q", seen.body, seen.header.Get("Content-Type"))
	}
}
