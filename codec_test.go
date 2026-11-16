// SPDX-License-Identifier: Apache-2.0 OR MIT

package cryptoservice

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"testing/iotest"
)

func TestEncodeKeepsTextAsWritten(t *testing.T) {
	got, err := encode(HashParams{Algorithm: SHA256, Data: "<a&b>"})
	if err != nil || string(got) != `{"algorithm":"sha256","data":"<a&b>"}` {
		t.Errorf("%s %v", got, err)
	}
	if _, err := encode(func() {}); err == nil {
		t.Error("encoded a func")
	}
}

func TestUnencodableParamsFailBeforeSending(t *testing.T) {
	c, _ := NewClient("http://127.0.0.1:1")
	_, err := c.StreamIso20022(context.Background(), Iso20022VerifyParams{Payload: make(chan int)})
	if err == nil || !strings.Contains(err.Error(), "encoding the request") {
		t.Errorf("err %v", err)
	}
}

func TestDecode(t *testing.T) {
	var out map[string]any
	if err := decode(strings.NewReader(` {"a":1} `+"\n"), &out); err != nil || out["a"] == nil {
		t.Errorf("%v %v", out, err)
	}
	for _, body := range []string{`{"a":1} x`, `{"a":`, ``, `nope`, `{"a":1}{"b":2}`} {
		if err := decode(strings.NewReader(body), &out); !errors.Is(err, errDecode) {
			t.Errorf("%q: %v", body, err)
		}
	}
	var n struct{ A int }
	if err := decode(strings.NewReader(`{"A":"x"}`), &n); !errors.Is(err, errDecode) {
		t.Errorf("type mismatch: %v", err)
	}
	broken := iotest.ErrReader(io.ErrClosedPipe)
	if err := decode(broken, &out); !errors.Is(err, io.ErrClosedPipe) || errors.Is(err, errDecode) {
		t.Errorf("read failure: %v", err)
	}
}

func TestResponsesOverTheCapAreRefused(t *testing.T) {
	saved := maxResponseBytes
	maxResponseBytes = 8
	t.Cleanup(func() { maxResponseBytes = saved })
	var out any
	if err := decode(strings.NewReader(`"0123456789"`), &out); !errors.Is(err, ErrResponseTooLarge) {
		t.Errorf("%v", err)
	}
	if err := decode(strings.NewReader(`"012345"`), &out); err != nil {
		t.Errorf("at the cap: %v", err)
	}
	if _, err := readAll(strings.NewReader("0123456789")); !errors.Is(err, ErrResponseTooLarge) || !errors.Is(err, errDecode) {
		t.Errorf("%v", err)
	}
	c, _ := fakeServer(t, 200, `{"data":{"digest":"0011223344"}}`)
	if _, err := c.Hash(context.Background(), HashParams{}); !errors.Is(err, ErrResponseTooLarge) {
		t.Errorf("%v", err)
	}
}

func TestMissingDataIsADecodeError(t *testing.T) {
	for _, body := range []string{`{}`, `{"data":null}`, `{"data":`} {
		c, _ := fakeServer(t, 200, body)
		_, err := c.Hash(context.Background(), HashParams{})
		var de *DecodeError
		if !errors.As(err, &de) || de.Status != http.StatusOK {
			t.Errorf("%q: %v", body, err)
		}
	}
	c, _ := fakeServer(t, 200, `{"data":5}`)
	if _, err := c.SecretboxOpen(context.Background(), SecretboxOpenParams{}); err == nil {
		t.Error("a number decoded as a string")
	}
	if _, err := c.Algorithms(context.Background()); err == nil {
		t.Error("a number decoded as the algorithm map")
	}
}

func TestHexAndBase64(t *testing.T) {
	b := []byte{0, 0xff, 'a'}
	if h := HexOf(b); h != "00ff61" {
		t.Error(h)
	}
	if got, err := Hex("00FF61").Bytes(); err != nil || string(got) != string(b) {
		t.Error(got, err)
	}
	if _, err := Hex("zz").Bytes(); err == nil {
		t.Error("decoded zz")
	}
	if s := Base64Of(b); s != "AP9h" {
		t.Error(s)
	}
	if got, err := Base64("AP9h").Bytes(); err != nil || string(got) != string(b) {
		t.Error(got, err)
	}
}
