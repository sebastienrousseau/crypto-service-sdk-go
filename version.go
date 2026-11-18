// SPDX-License-Identifier: Apache-2.0 OR MIT
// SPDX-FileCopyrightText: 2026 Sebastien Rousseau

package cryptoservice

// Version is this module's release, in lockstep with the Crypto Service
// suite. scripts/verify-release-versions checks every file that
// restates it.
const Version = "0.0.40"

// userAgent identifies the client on every request, so a server's
// access log can tell which SDK and release sent a call.
const userAgent = "crypto-service-sdk-go/" + Version
