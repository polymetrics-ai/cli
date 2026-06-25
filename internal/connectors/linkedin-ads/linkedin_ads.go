// Package linkedinads implements the native pm LinkedIn Ads connector. It follows
// the declarative-HTTP shape established by the stripe connector: a thin package
// that composes the connsdk toolkit (Requester + Bearer auth + RecordsAt
// extraction + cursor state) with LinkedIn-Marketing-API-specific stream
// definitions, endpoints, and offset pagination.
//
// LinkedIn's Marketing API has a few traits versus a vanilla bearer API:
//   - Every request must carry a LinkedIn-Version header (e.g. 202601) and the
//     X-Restli-Protocol-Version: 2.0.0 header.
//   - List endpoints return {"elements":[...]} and are paged with start/count
//     offset parameters; a short page (fewer than count) signals the last page.
//   - The access token is a member OAuth2 token; this connector reads the
//     long-lived access_token secret directly (the refresh_token grant exchange
//     is left to the operator/agent layer). client_id/client_secret/refresh_token
//     are accepted as secrets but only the resolved access_token is used here.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package linkedinads

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
	registryName = "linkedin-ads"

	linkedinDefaultBaseURL = "https://api.linkedin.com/rest"
	// linkedinDefaultVersion is the LinkedIn-Version header sent when the operator
	// does not override it via config. LinkedIn versions are monthly (YYYYMM).
	linkedinDefaultVersion = "202601"
	linkedinRestliVersion  = "2.0.0"
	linkedinUserAgent      = "polymetrics-go-cli"

	linkedinDefaultPageSize = 100
	linkedinMaxPageSize     = 1000

	// linkedinFixtureModified is the deterministic lastModified epoch-millis used
	// by fixture-mode records (2026-01-01T00:00:00Z in unix millis).
	linkedinFixtureModified int64 = 1767225600000
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the LinkedIn Ads connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm LinkedIn Ads connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "LinkedIn Ads",
		IntegrationType: "api",
		Description:     "Reads LinkedIn Ads accounts, campaign groups, campaigns, and creatives through the LinkedIn Marketing REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to LinkedIn. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := linkedinBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(linkedinSecret(cfg)) == "" {
		return errors.New("linkedin-ads connector requires secret credentials.access_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the adAccounts list confirms auth and connectivity
	// without mutating anything.
	query := url.Values{"q": []string{"search"}, "start": []string{"0"}, "count": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "adAccounts", query, nil, nil); err != nil {
		return fmt.Errorf("check linkedin-ads: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: linkedinStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a LinkedIn stream starts with
// an empty incremental cursor (full sync).
func (c Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return connsdk.WithCursor(map[string]string{"stream": stream}, ""), nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "accounts"
	}
	endpoint, ok := linkedinStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("linkedin-ads stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := linkedinPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := linkedinMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives LinkedIn's start/count offset pagination. List endpoints return
// {"elements":[...]}; the next page is requested by advancing start by count. A
// page shorter than count signals the last page. The loop is built on
// connsdk.Requester + connsdk.RecordsAt so it shares retry/rate-limit handling.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	start := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		// "search" is the canonical finder for LinkedIn ad-entity list endpoints.
		query.Set("q", "search")
		query.Set("start", strconv.Itoa(start))
		query.Set("count", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read linkedin-ads %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "elements")
		if err != nil {
			return fmt.Errorf("decode linkedin-ads %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
		start += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise linkedin-ads credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	endpoint := linkedinStreamEndpoints[stream]
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":             int64(1000 + i),
			"name":           fmt.Sprintf("Fixture %s %d", stream, i),
			"status":         "ACTIVE",
			"type":           "BUSINESS",
			"currency":       "USD",
			"account":        "urn:li:sponsoredAccount:1001",
			"campaignGroup":  "urn:li:sponsoredCampaignGroup:2001",
			"campaign":       "urn:li:sponsoredCampaign:3001",
			"costType":       "CPM",
			"objectiveType":  "BRAND_AWARENESS",
			"format":         "STANDARD_UPDATE",
			"isServing":      true,
			"intendedStatus": "ACTIVE",
			"createdAt":      linkedinFixtureModified,
			"lastModifiedAt": linkedinFixtureModified + int64(i),
			"changeAuditStamps": map[string]any{
				"created":      map[string]any{"time": linkedinFixtureModified},
				"lastModified": map[string]any{"time": linkedinFixtureModified + int64(i)},
			},
		}
		// Creatives use string urn ids; other streams use numeric ids.
		if stream == "creatives" {
			item["id"] = fmt.Sprintf("urn:li:sponsoredCreative:%d", 3000+i)
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

// requester builds a connsdk.Requester wired with Bearer auth, the resolved base
// URL, and the mandatory LinkedIn-Version + X-Restli-Protocol-Version headers.
// The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := linkedinBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := linkedinSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("linkedin-ads connector requires secret credentials.access_token")
	}
	headers := map[string]string{
		"LinkedIn-Version":          linkedinVersion(cfg),
		"X-Restli-Protocol-Version": linkedinRestliVersion,
	}
	return &connsdk.Requester{
		Client:         c.Client,
		BaseURL:        base,
		Auth:           connsdk.Bearer(secret),
		UserAgent:      linkedinUserAgent,
		DefaultHeaders: headers,
	}, nil
}

// linkedinSecret resolves the access token from the credentials.access_token
// secret. The OAuth2 path stores the bare token under the same dotted key after
// the refresh exchange, so a single lookup covers both auth methods.
func linkedinSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	if v := strings.TrimSpace(cfg.Secrets["credentials.access_token"]); v != "" {
		return v
	}
	// Tolerate the un-prefixed key for convenience.
	return strings.TrimSpace(cfg.Secrets["access_token"])
}

func linkedinVersion(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return linkedinDefaultVersion
	}
	if v := strings.TrimSpace(cfg.Config["linkedin_version"]); v != "" {
		return v
	}
	return linkedinDefaultVersion
}

// linkedinBaseURL resolves and validates the base URL. The default is
// api.linkedin.com/rest; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func linkedinBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return linkedinDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("linkedin-ads config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("linkedin-ads config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("linkedin-ads config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func linkedinPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return linkedinDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("linkedin-ads config page_size must be an integer: %w", err)
	}
	if value < 1 || value > linkedinMaxPageSize {
		return 0, fmt.Errorf("linkedin-ads config page_size must be between 1 and %d", linkedinMaxPageSize)
	}
	return value, nil
}

func linkedinMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("linkedin-ads config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("linkedin-ads config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write satisfies the connectors.Connector interface. LinkedIn Ads is read-only
// in pm: there is no approved reverse-ETL write surface, so any write is
// rejected as an unsupported operation.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{RecordsFailed: len(records)}, connectors.ErrUnsupportedOperation
}
