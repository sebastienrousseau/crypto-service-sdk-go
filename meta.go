// SPDX-License-Identifier: Apache-2.0 OR MIT
// SPDX-FileCopyrightText: 2026 Sebastien Rousseau

package cryptoservice

import (
	"context"
	"net/http"
)

// HealthResult is the body of GET /health.
type HealthResult struct {
	Status     string `json:"status"`
	StatusCode int    `json:"statusCode"`
	// Uptime is the process uptime in seconds.
	Uptime float64 `json:"uptime"`
}

// LiveResult is the body of GET /live.
type LiveResult struct {
	Status string `json:"status"`
	Uptime int64  `json:"uptime"`
}

// ReadyResult is the body of GET /ready, which the server sends with
// 200 when ready and 503 when not; Ready returns it for both.
type ReadyResult struct {
	// Status is "ready" or "not_ready".
	Status string `json:"status"`
	// Checks maps each check (accepting, keyStore, ...) to whether it
	// passes.
	Checks map[string]bool `json:"checks"`
}

// ServiceInfo is the body of GET /.
type ServiceInfo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Details     string `json:"details"`
	License     string `json:"license"`
	URL         string `json:"url"`
	Version     string `json:"version"`
}

// Algorithms maps each category (encryption, hashing, kdf, mac,
// signing, keyExchange, postQuantum, and any newer one) to its
// algorithm identifiers.
type Algorithms map[string][]string

// DoraDeprecation is one entry of the DORA deprecation schedule.
type DoraDeprecation struct {
	Algorithm            string `json:"algorithm"`
	Category             string `json:"category"`
	SunsetDate           string `json:"sunsetDate"`
	RecommendedMigration string `json:"recommendedMigration"`
}

// DoraInventoryItem is one package of the cryptographic inventory.
type DoraInventoryItem struct {
	Package        string `json:"package"`
	Version        string `json:"version"`
	Status         string `json:"status"`
	FipsCompliance string `json:"fipsCompliance"`
}

// DoraComplianceScorecard is the data of GET /v2/compliance/dora: a
// self-assessment of the algorithms the suite implements, not a
// compliance verdict.
type DoraComplianceScorecard struct {
	Standard                     string              `json:"standard"`
	Article                      string              `json:"article"`
	Status                       string              `json:"status"`
	Disclaimer                   string              `json:"disclaimer"`
	QuantumResistanceRatio       float64             `json:"quantumResistanceRatio"`
	ActivePrimitivesCount        int                 `json:"activePrimitivesCount"`
	PostQuantumPrimitivesCount   int                 `json:"postQuantumPrimitivesCount"`
	AlgorithmDeprecationSchedule []DoraDeprecation   `json:"algorithmDeprecationSchedule"`
	CryptographicInventory       []DoraInventoryItem `json:"cryptographicInventory"`
	Timestamp                    string              `json:"timestamp"`
}

// Cbom is the body of GET /v2/compliance/cbom: a CycloneDX 1.6
// Cryptographic Bill of Materials, not wrapped in data. Metadata and
// Components keep the CycloneDX structure as decoded JSON.
type Cbom struct {
	BomFormat    string           `json:"bomFormat"`
	SpecVersion  string           `json:"specVersion"`
	SerialNumber string           `json:"serialNumber"`
	Version      int              `json:"version"`
	Metadata     map[string]any   `json:"metadata"`
	Components   []map[string]any `json:"components"`
}

// get fetches path and decodes its body, not wrapped in data.
func get[T any](ctx context.Context, c *Client, rq *request) (*T, error) {
	rq.method = http.MethodGet
	out := new(T)
	if err := c.call(ctx, rq, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Health reports the server's health (GET /health). It sends no
// credentials: the route is public.
func (c *Client) Health(ctx context.Context) (*HealthResult, error) {
	return get[HealthResult](ctx, c, &request{path: "/health", public: true})
}

// Live reports liveness (GET /live, public).
func (c *Client) Live(ctx context.Context) (*LiveResult, error) {
	return get[LiveResult](ctx, c, &request{path: "/live", public: true})
}

// Ready reports readiness (GET /ready, public). A server that is not
// ready answers 503 with the same document; Ready returns it with
// Status "not_ready" and no error, and does not retry it.
func (c *Client) Ready(ctx context.Context) (*ReadyResult, error) {
	return get[ReadyResult](ctx, c, &request{path: "/ready", public: true, accept: http.StatusServiceUnavailable})
}

// Root returns the service metadata (GET /, authenticated).
func (c *Client) Root(ctx context.Context) (*ServiceInfo, error) {
	return get[ServiceInfo](ctx, c, &request{path: "/"})
}

// Algorithms lists the algorithm identifiers by category (GET
// /v2/algorithms, any authenticated principal).
func (c *Client) Algorithms(ctx context.Context) (Algorithms, error) {
	a, err := wrapped[Algorithms](ctx, c, http.MethodGet, "/v2/algorithms", nil)
	if err != nil {
		return nil, err
	}
	return *a, nil
}

// GetDoraCompliance returns the DORA Article 9 and 13 self-assessment
// (GET /v2/compliance/dora, any authenticated principal).
func (c *Client) GetDoraCompliance(ctx context.Context) (*DoraComplianceScorecard, error) {
	return wrapped[DoraComplianceScorecard](ctx, c, http.MethodGet, "/v2/compliance/dora", nil)
}

// GetCbom returns the CycloneDX 1.6 Cryptographic Bill of Materials
// (GET /v2/compliance/cbom, any authenticated principal).
func (c *Client) GetCbom(ctx context.Context) (*Cbom, error) {
	return get[Cbom](ctx, c, &request{path: "/v2/compliance/cbom"})
}
