// Package gainsightpx implements the native pm Gainsight PX connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference connector: a thin package that composes the connsdk toolkit
// (Requester + APIKeyHeader auth + RecordsAt extraction + scrollId cursor
// pagination) with Gainsight-PX-specific stream definitions and endpoints.
//
// The Gainsight PX (aptrinsic) REST API uses header API-key auth
// (X-APTRINSIC-API-KEY: <api_key>), nests list payloads under a per-stream key
// (e.g. {"accounts":[...],"scrollId":"..."}), and paginates with a scrollId
// cursor carried in the response body and replayed as a request parameter. It
// supports only full-refresh sync, so no incremental cursor is published, and it
// exposes no reverse-ETL writes, so Capabilities.Write is false.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package gainsightpx

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	registryName          = "gainsight-px"
	defaultBaseURL        = "https://api.aptrinsic.com/v1"
	defaultPageSize       = 100
	maxPageSize           = 500
	userAgent             = "polymetrics-go-cli"
	authHeader            = "X-APTRINSIC-API-KEY"
	scrollParam           = "scrollId"
	pageSizeParam         = "pageSize"
	defaultMaxPagesUnset  = 0
	fixtureModeRecordsLen = 2
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the Gainsight PX connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Gainsight PX connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "Gainsight PX",
		IntegrationType: "api",
		Description:     "Reads Gainsight PX accounts, users, features, and segments through the aptrinsic REST API (read-only).",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Gainsight PX.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := resolveBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(secretAPIKey(cfg)) == "" {
		return errors.New("gainsight-px connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the accounts list confirms auth and connectivity without
	// mutating anything.
	q := url.Values{pageSizeParam: []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "accounts", q, nil, nil); err != nil {
		return fmt.Errorf("check gainsight-px: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: gainsightStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "accounts"
	}
	endpoint, ok := gainsightStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("gainsight-px stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := resolvePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := resolveMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// Write satisfies the connectors.Connector interface. The Gainsight PX source is
// read-only (no safe reverse-ETL write surface), so it rejects writes.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// harvest drives Gainsight PX's scrollId pagination. List responses carry the
// records under endpoint.recordsKey and a scrollId token; the next page is
// requested with ?scrollId=<token>. An empty/missing scrollId (or an empty page)
// ends the scan. There is no connsdk paginator for this exact body-token shape,
// so the loop lives here, built on connsdk.Requester + connsdk.RecordsAt +
// connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set(pageSizeParam, strconv.Itoa(pageSize))

	scrollID := ""
	seen := map[string]bool{}
	for page := 0; maxPages == defaultMaxPagesUnset || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if scrollID != "" {
			query.Set(scrollParam, scrollID)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read gainsight-px %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsKey)
		if err != nil {
			return fmt.Errorf("decode gainsight-px %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, scrollParam)
		if err != nil {
			return fmt.Errorf("decode gainsight-px %s scrollId: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		// Stop on an empty token, an empty page (no further data), or a repeated
		// token (defensive guard against a server that echoes the same cursor).
		if next == "" || len(records) == 0 || seen[next] {
			return nil
		}
		seen[next] = true
		scrollID = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise gainsight-px credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= fixtureModeRecordsLen; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                    fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"name":                  fmt.Sprintf("Fixture %s %d", stream, i),
			"type":                  "FEATURE",
			"accountId":             "acc_fixture_1",
			"aptrinsicId":           fmt.Sprintf("apt_fixture_%d", i),
			"email":                 fmt.Sprintf("fixture+%d@example.com", i),
			"firstName":             "Fixture",
			"lastName":              fmt.Sprintf("User %d", i),
			"title":                 "Engineer",
			"role":                  "member",
			"score":                 float64(i),
			"numberOfVisits":        int64(i),
			"trackedSubscriptionId": "sub_fixture_1",
			"sfdcId":                "sfdc_fixture_1",
			"industry":              "Software",
			"plan":                  "enterprise",
			"location":              "Remote",
			"website":               "https://example.com",
			"status":                "ACTIVE",
			"priority":              "1",
			"productId":             "prod_fixture_1",
			"productName":           "PX",
			"parentFeatureId":       "",
			"propertyKey":           "AP-FIXTURE-2",
			"description":           "deterministic fixture record",
			"createdBy":             "fixture",
			"modifiedBy":            "fixture",
			"createDate":            int64(1767225600 + i),
			"createdDate":           "2026-01-01T00:00:00Z",
			"modifiedDate":          "2026-01-02T00:00:00Z",
			"lastModifiedDate":      int64(1767225600 + i),
			"lastSeenDate":          int64(1767225600 + i),
			"signUpDate":            int64(1767225600),
		}
		record := endpoint.mapRecord(item)
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with X-APTRINSIC-API-KEY header auth
// and the resolved base URL. The secret only ever flows into connsdk.APIKeyHeader;
// it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := resolveBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := secretAPIKey(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("gainsight-px connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(authHeader, secret, ""),
		UserAgent: userAgent,
	}, nil
}

func secretAPIKey(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// resolveBaseURL resolves and validates the base URL. The default is
// api.aptrinsic.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func resolveBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return defaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("gainsight-px config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("gainsight-px config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("gainsight-px config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func resolvePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("gainsight-px config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("gainsight-px config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

func resolveMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return defaultMaxPagesUnset, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("gainsight-px config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("gainsight-px config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func cloneValues(in url.Values) url.Values {
	out := url.Values{}
	for k, vs := range in {
		for _, v := range vs {
			out.Add(k, v)
		}
	}
	return out
}
