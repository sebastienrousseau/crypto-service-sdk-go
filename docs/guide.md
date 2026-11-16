<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Using the client

## Create one client and share it

`NewClient` builds a client with its own pooled transport. Create it
once per server and share it between goroutines; `Close` releases idle
connections when you are done.

```go
client, err := cryptoservice.NewClient("https://crypto.example.internal",
    cryptoservice.WithToken(token),
    cryptoservice.WithRetry(cryptoservice.RetryPolicy{MaxRetries: 2}),
    cryptoservice.WithDeadline(15*time.Second))
```

## Keys stay on the server

Key-generation methods (`GenerateKeyPair`, `MlKemGenerateKeyPair`,
`PqGenerateKeyPair`, `PqSignKeygen`, `PqHashSignKeygen`,
`HpkeGenerateKeyPair`) return a `KeyID` and the public key. Pass the
`KeyID` to `Sign`, `MlKemDecapsulate`, `SealedboxOpen`, `HpkeOpen` and
the other private-key operations; delete it with `KeysDelete`.

## Encodings

| Go type | Wire form | Used for |
| :--- | :--- | :--- |
| `Hex` | lowercase hexadecimal | keys, signatures, digests, MACs, KEM ciphertexts and shared secrets, HPKE output |
| `Base64` | standard base64 with padding | AEAD, secretbox, sealed-box, stream, password-encryption and AES-KW output |
| `string` | UTF-8 text | plaintexts, messages, opaque KMS ciphertexts and wrapped keys |

`HpkeSeal` passes its inputs through unchanged; the server reads an
even-length hex string as hex and anything else as text, so send
`HexOf(data)` when the input could be mistaken for hex. `KmsEncrypt`
always sends its `Encoding` (`KmsUTF8` unless you choose `KmsHex`).
`HpkeOpen` and `KmsDecrypt` return the plaintext both as text and as
`Hex`.

## Errors

```go
var apiErr *cryptoservice.APIError
var transport *cryptoservice.TransportError
switch {
case errors.As(err, &apiErr):
    // apiErr.Status, apiErr.Problem.Slug(), apiErr.Problem.Code, apiErr.Problem.Errors
case errors.As(err, &transport):
    // transport.TimedOut, errors.Is(err, context.Canceled)
}
```

## Retries

`WithRetry` retries 429, 503, 504 and lost connections for calls that
cannot change server state twice, and only 429 and 503 for key
generation, deletion, KMS key creation and OPAQUE registration. It waits
with full jitter below a cap that doubles from `InitialDelay` to
`MaxDelay`, or for `Retry-After` when the server sends one; a
`Retry-After` longer than `MaxDelay` ends the retries.
