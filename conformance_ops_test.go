// SPDX-License-Identifier: Apache-2.0 OR MIT

package cryptoservice_test

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"

	cs "github.com/sebastienrousseau/crypto-service-sdk-go"
	"github.com/sebastienrousseau/crypto-service-sdk-go/internal/kit"
)

// op runs a vector's request through a public method. It reports
// ok=false when the method cannot send exactly the vector's body (an
// unknown member, a missing required one), and the vector then goes
// through Raw instead. result is the method's answer in wire form.
type op func(ctx context.Context, c *cs.Client, body []byte) (result any, ok bool, err error)

// typed adapts a method taking params P; wrap puts the result back
// under "data", as the server sent it.
func typed[P, R any](m func(*cs.Client, context.Context, P) (R, error), wrap bool, defaults map[string]any) op {
	return func(ctx context.Context, c *cs.Client, body []byte) (any, bool, error) {
		var p P
		if !sendsExactly(body, &p, defaults) {
			return nil, false, nil
		}
		r, err := m(c, ctx, p)
		return wireForm(r, wrap), true, err
	}
}

// bare adapts a method without params: the vector must send no body
// or an empty object.
func bare[R any](m func(*cs.Client, context.Context) (R, error), wrap bool) op {
	return func(ctx context.Context, c *cs.Client, body []byte) (any, bool, error) {
		if len(body) > 0 && !bytes.Equal(body, []byte("{}")) {
			return nil, false, nil
		}
		r, err := m(c, ctx)
		return wireForm(r, wrap), true, err
	}
}

func wireForm(r any, wrap bool) any {
	if wrap {
		return map[string]any{"data": r}
	}
	return r
}

// sendsExactly decodes body into p and reports whether encoding p
// again gives the same JSON value, after the method's documented
// defaults are applied to body (KmsEncrypt always sends its encoding).
func sendsExactly(body []byte, p any, defaults map[string]any) bool {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if dec.Decode(p) != nil {
		return false
	}
	want, err := kit.Decode(body)
	if err != nil {
		return false
	}
	if obj, ok := want.(map[string]any); ok {
		for k, v := range defaults {
			if _, set := obj[k]; !set {
				obj[k] = v
			}
		}
	}
	again, _ := json.Marshal(p)
	got, _ := kit.Decode(again)
	return reflect.DeepEqual(want, got)
}

type client = cs.Client

// ops maps "METHOD path" to the public method for it. Every v2
// operation and the probes are here; v1 and /metrics are not wrapped.
var ops = map[string]op{
	"GET /":                            bare((*client).Root, false),
	"GET /health":                      bare((*client).Health, false),
	"GET /live":                        bare((*client).Live, false),
	"GET /ready":                       bare((*client).Ready, false),
	"GET /v2/algorithms":               bare((*client).Algorithms, true),
	"GET /v2/compliance/dora":          bare((*client).GetDoraCompliance, true),
	"GET /v2/compliance/cbom":          bare((*client).GetCbom, false),
	"POST /v2/encrypt":                 typed((*client).Encrypt, true, nil),
	"POST /v2/decrypt":                 typed((*client).Decrypt, true, nil),
	"POST /v2/hash":                    typed((*client).Hash, true, nil),
	"POST /v2/hmac":                    typed((*client).Mac, true, nil),
	"POST /v2/hmac/verify":             typed((*client).MacVerify, true, nil),
	"POST /v2/kdf":                     typed((*client).Kdf, true, nil),
	"POST /v2/password/hash":           typed((*client).PasswordHash, true, nil),
	"POST /v2/password/verify":         typed((*client).PasswordVerify, true, nil),
	"POST /v2/password/encrypt":        typed((*client).PasswordEncrypt, true, nil),
	"POST /v2/password/decrypt":        typed((*client).PasswordDecrypt, true, nil),
	"POST /v2/secretbox/seal":          typed((*client).SecretboxSeal, true, nil),
	"POST /v2/secretbox/open":          typed((*client).SecretboxOpen, true, nil),
	"POST /v2/keys/wrap":               typed((*client).KeyWrap, true, nil),
	"POST /v2/keys/unwrap":             typed((*client).KeyUnwrap, true, nil),
	"POST /v2/keys/generate":           typed((*client).GenerateKeyPair, true, nil),
	"POST /v2/keys/export":             typed((*client).ExportKey, true, nil),
	"POST /v2/keys/delete":             typed((*client).KeysDelete, true, nil),
	"POST /v2/sign":                    typed((*client).Sign, true, nil),
	"POST /v2/verify":                  typed((*client).Verify, true, nil),
	"POST /v2/stream/sign":             typed((*client).StreamSign, true, nil),
	"POST /v2/stream/verify":           typed((*client).StreamVerify, true, nil),
	"POST /v2/pq/keygen":               bare((*client).MlKemGenerateKeyPair, true),
	"POST /v2/pq/encapsulate":          typed((*client).MlKemEncapsulate, true, nil),
	"POST /v2/pq/decapsulate":          typed((*client).MlKemDecapsulate, true, nil),
	"POST /v2/pq/hybrid/keygen":        bare((*client).PqGenerateKeyPair, true),
	"POST /v2/pq/hybrid/encapsulate":   typed((*client).PqEncapsulate, true, nil),
	"POST /v2/pq/hybrid/decapsulate":   typed((*client).PqDecapsulate, true, nil),
	"POST /v2/pq/dsa/keygen":           typed((*client).PqSignKeygen, true, nil),
	"POST /v2/pq/dsa/sign":             typed((*client).PqSign, true, nil),
	"POST /v2/pq/dsa/verify":           typed((*client).PqVerify, true, nil),
	"POST /v2/pq/slh-dsa/keygen":       typed((*client).PqHashSignKeygen, true, nil),
	"POST /v2/pq/slh-dsa/sign":         typed((*client).PqHashSign, true, nil),
	"POST /v2/pq/slh-dsa/verify":       typed((*client).PqHashVerify, true, nil),
	"POST /v2/sealedbox/seal":          typed((*client).SealedboxSeal, true, nil),
	"POST /v2/sealedbox/open":          typed((*client).SealedboxOpen, true, nil),
	"POST /v2/sealedbox/seal-pq":       typed((*client).SealedboxSealPq, true, nil),
	"POST /v2/sealedbox/open-pq":       typed((*client).SealedboxOpenPq, true, nil),
	"POST /v2/multi-recipient/encrypt": typed((*client).MultiRecipientEncrypt, true, nil),
	"POST /v2/stream/pq-encrypt":       typed((*client).StreamPqEncrypt, true, nil),
	"POST /v2/stream/pq-decrypt":       typed((*client).StreamPqDecrypt, true, nil),
	"POST /v2/stream/multi-pq-encrypt": typed((*client).StreamMultiPqEncrypt, true, nil),
	"POST /v2/stream/multi-pq-decrypt": typed((*client).StreamMultiPqDecrypt, true, nil),
	"POST /v2/stream/iso20022":         typed((*client).StreamIso20022, true, nil),
	"POST /v2/hpke/keygen":             typed((*client).HpkeGenerateKeyPair, true, nil),
	"POST /v2/hpke/seal":               typed((*client).HpkeSeal, true, nil),
	"POST /v2/hpke/open":               typed((*client).HpkeOpen, true, nil),
	"POST /v2/kms/create-key":          typed((*client).KmsCreateKey, true, nil),
	"POST /v2/kms/wrap":                typed((*client).KmsWrap, true, nil),
	"POST /v2/kms/unwrap":              typed((*client).KmsUnwrap, true, nil),
	"POST /v2/kms/generate-data-key":   typed((*client).KmsGenerateDataKey, true, nil),
	"POST /v2/kms/encrypt":             typed((*client).KmsEncrypt, true, map[string]any{"encoding": "utf8"}),
	"POST /v2/kms/decrypt":             typed((*client).KmsDecrypt, true, nil),
	"POST /v2/opaque/register/init":    typed((*client).OpaqueRegisterInit, true, nil),
	"POST /v2/opaque/register/finish":  typed((*client).OpaqueRegisterFinish, true, nil),
	"POST /v2/opaque/login/init":       typed((*client).OpaqueLoginInit, true, nil),
	"POST /v2/opaque/login/finish":     typed((*client).OpaqueLoginFinish, true, nil),
}
