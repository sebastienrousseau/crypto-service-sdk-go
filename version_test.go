// SPDX-License-Identifier: Apache-2.0 OR MIT

package cryptoservice

import (
	"context"
	"testing"
)

func TestEveryRequestCarriesTheUserAgent(t *testing.T) {
	c, seen := fakeServer(t, 200, `{"status":"ok","statusCode":200,"uptime":1}`)
	if _, err := c.Health(context.Background()); err != nil {
		t.Fatal(err)
	}
	if want := "crypto-service-sdk-go/" + Version; seen.header.Get("User-Agent") != want {
		t.Errorf("User-Agent %q, want %q", seen.header.Get("User-Agent"), want)
	}
}
