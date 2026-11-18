// SPDX-License-Identifier: Apache-2.0 OR MIT
// SPDX-FileCopyrightText: 2026 Sebastien Rousseau

package cryptoservice

import (
	"encoding/base64"
	"encoding/hex"
)

// Hex is a string the API carries as lowercase hexadecimal: keys,
// signatures, digests, MACs, shared secrets and KEM ciphertexts. It is
// sent and received exactly as written; Bytes decodes it and HexOf
// encodes bytes into it.
type Hex string

// HexOf returns b as lowercase hexadecimal.
func HexOf(b []byte) Hex { return Hex(hex.EncodeToString(b)) }

// Bytes decodes h (either case is accepted).
func (h Hex) Bytes() ([]byte, error) { return hex.DecodeString(string(h)) }

// Base64 is a string the API carries as standard base64 with padding:
// AEAD, secretbox, sealed-box, stream, password-encryption and AES-KW
// outputs. It is sent and received exactly as written; Bytes decodes it
// and Base64Of encodes bytes into it.
type Base64 string

// Base64Of returns b as standard padded base64.
func Base64Of(b []byte) Base64 { return Base64(base64.StdEncoding.EncodeToString(b)) }

// Bytes decodes b.
func (b Base64) Bytes() ([]byte, error) { return base64.StdEncoding.DecodeString(string(b)) }
