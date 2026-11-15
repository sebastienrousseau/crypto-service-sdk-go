// SPDX-License-Identifier: Apache-2.0 OR MIT

package cryptoservice_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	cs "github.com/sebastienrousseau/crypto-service-sdk-go"
)

// The typed live tests run as their own JWT principals, apart from the
// first, which uses the API key as an application would.

func TestLiveHashAndProbes(t *testing.T) {
	c := liveServer(t).apiKeyClient(t)
	ctx := context.Background()
	res, err := c.Hash(ctx, cs.HashParams{Algorithm: cs.SHA256, Data: "abc"})
	if err != nil || res.Digest != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" || res.Length != 32 {
		t.Fatalf("%+v %v", res, err)
	}
	if h, err := c.Health(ctx); err != nil || h.Status != "ok" {
		t.Fatalf("%+v %v", h, err)
	}
	if r, err := c.Ready(ctx); err != nil || r.Status != "ready" {
		t.Fatalf("%+v %v", r, err)
	}
}

func TestLiveKeyRoundTrips(t *testing.T) {
	c := liveServer(t).tokenClient(t, "live-keys", cs.WithAutoTraceparent())
	ctx := context.Background()
	kp, err := c.GenerateKeyPair(ctx, cs.GenerateKeyPairParams{Algorithm: cs.Ed25519})
	if err != nil {
		t.Fatal(err)
	}
	sig, err := c.Sign(ctx, cs.SignParams{KeyID: kp.KeyID, Message: "live"})
	if err != nil {
		t.Fatal(err)
	}
	v, err := c.Verify(ctx, cs.VerifyParams{PublicKey: kp.PublicKey, Message: "live", Signature: sig.Signature})
	if err != nil || !v.Valid {
		t.Fatalf("%+v %v", v, err)
	}
	if _, err := c.KeysDelete(ctx, cs.KeyIDParams{KeyID: kp.KeyID}); err != nil {
		t.Fatal(err)
	}
	_, err = c.Sign(ctx, cs.SignParams{KeyID: kp.KeyID, Message: "live"})
	var apiErr *cs.APIError
	if !errors.As(err, &apiErr) || apiErr.Status != 404 || apiErr.Problem.Code != "KEY_NOT_FOUND" {
		t.Fatalf("signing with a deleted key: %v", err)
	}
}

func TestLiveKmsKeepsBytes(t *testing.T) {
	c := liveServer(t).tokenClient(t, "live-kms")
	ctx := context.Background()
	key, err := c.KmsCreateKey(ctx, cs.KmsCreateKeyParams{})
	if err != nil {
		t.Fatal(err)
	}
	enc, err := c.KmsEncrypt(ctx, cs.KmsEncryptParams{KeyID: key.KeyID, Plaintext: string(cs.HexOf([]byte{0, 0xff})), Encoding: cs.KmsHex})
	if err != nil {
		t.Fatal(err)
	}
	dec, err := c.KmsDecrypt(ctx, cs.KmsDecryptParams{KeyID: key.KeyID, Ciphertext: enc.Ciphertext})
	if err != nil {
		t.Fatal(err)
	}
	if b, err := dec.Hex.Bytes(); err != nil || len(b) != 2 || b[1] != 0xff {
		t.Fatalf("%v %v", b, err)
	}
}

func TestLiveConcurrentCallsShareOneClient(t *testing.T) {
	c := liveServer(t).tokenClient(t, "live-concurrent")
	var wg sync.WaitGroup
	errs := make([]error, 8)
	for i := range errs {
		wg.Go(func() {
			_, errs[i] = c.Hash(context.Background(), cs.HashParams{Algorithm: cs.BLAKE3, Data: "concurrent"})
		})
	}
	wg.Wait()
	if err := errors.Join(errs...); err != nil {
		t.Fatal(err)
	}
}
