// Package harness implements the native pm Harness connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference: a thin package that composes the connsdk toolkit (Requester +
// x-api-key auth + RecordsAt extraction) with Harness NextGen stream
// definitions, endpoints, and page-index pagination.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// The Harness NextGen platform API authenticates with an `x-api-key` header,
// scopes every read to an account via the `accountIdentifier` query parameter,
// and returns list responses in the envelope
// {"data":{"content":[{<wrapper>:{...}}],"pageIndex":N,"totalPages":N,"empty":bool}}.
// This connector is read-only: organization/project/service listings have no
// safe reverse-ETL write semantics, so Capabilities.Write is false.
package harness

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
	harnessDefaultBaseURL  = "https://app.harness.io"
	harnessDefaultPageSize = 50
	harnessMaxPageSize     = 100
	harnessUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("harness", New)
}

// New returns the Harness connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Harness connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "harness" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "harness",
		DisplayName:     "Harness",
		IntegrationType: "api",
		Description:     "Reads Harness NextGen organizations, projects, services, connectors, and pipelines through the Harness platform REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Harness. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := harnessBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(harnessSecret(cfg)) == "" {
		return errors.New("harness connector requires secret api_key")
	}
	account, err := harnessAccountID(cfg)
	if err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the organizations list confirms auth and connectivity
	// without mutating anything.
	query := url.Values{"accountIdentifier": []string{account}, "pageIndex": []string{"0"}, "pageSize": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "ng/api/organizations", query, nil, nil); err != nil {
		return fmt.Errorf("check harness: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: harnessStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "organizations"
	}
	endpoint, ok := harnessStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("harness stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	account, err := harnessAccountID(req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := harnessPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := harnessMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, account, pageSize, maxPages, emit)
}

// harvest drives Harness NextGen page-index pagination. Each list response
// carries data.content[] plus data.pageIndex and data.totalPages; the loop
// requests pageIndex=0,1,... until pageIndex+1 >= totalPages (or a short/empty
// page is returned). It is built on connsdk.Requester + connsdk.RecordsAt +
// connsdk.StringAt rather than a connsdk paginator because the stop condition
// reads two body fields (this exact shape has no off-the-shelf paginator).
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, account string, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("accountIdentifier", account)
		query.Set("pageIndex", strconv.Itoa(page))
		query.Set("pageSize", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read harness %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data.content")
		if err != nil {
			return fmt.Errorf("decode harness %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(unwrap(item, endpoint.wrapper))); err != nil {
				return err
			}
		}

		if len(records) == 0 {
			return nil
		}
		totalPages, _ := connsdk.StringAt(resp.Body, "data.totalPages")
		total, perr := strconv.Atoi(strings.TrimSpace(totalPages))
		if perr != nil || total <= 0 {
			// totalPages absent/unparseable: stop on a short page to be safe.
			if len(records) < pageSize {
				return nil
			}
			continue
		}
		if page+1 >= total {
			return nil
		}
	}
	return nil
}

// unwrap returns the inner object addressed by wrapper inside a content element.
// NextGen wraps each list item under a singular key (content[i].organization);
// when wrapper is empty the element itself is the record. A missing/typed-wrong
// wrapper falls back to the element so a payload shape change degrades to raw
// fields rather than panicking.
func unwrap(item map[string]any, wrapper string) map[string]any {
	if wrapper == "" {
		return item
	}
	if inner, ok := item[wrapper].(map[string]any); ok {
		return inner
	}
	return item
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise harness credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"identifier":        fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":              fmt.Sprintf("Fixture %s %d", stream, i),
			"accountIdentifier": "acct_fixture",
			"orgIdentifier":     "org_fixture",
			"projectIdentifier": "proj_fixture",
			"description":       fmt.Sprintf("fixture %s record %d", stream, i),
			"color":             "#0063F7",
			"type":              "Github",
			"deleted":           false,
			"modules":           []any{"CD", "CI"},
			"numOfStages":       int64(i),
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with x-api-key auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := harnessBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := harnessSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("harness connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("x-api-key", secret, ""),
		UserAgent: harnessUserAgent,
	}, nil
}

func harnessSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// harnessAccountID resolves the required Harness account identifier from config.
// It is the accountIdentifier query param every NextGen read is scoped to.
func harnessAccountID(cfg connectors.RuntimeConfig) (string, error) {
	if cfg.Config == nil {
		return "", errors.New("harness connector requires config account_id")
	}
	account := strings.TrimSpace(cfg.Config["account_id"])
	if account == "" {
		account = strings.TrimSpace(cfg.Config["account_identifier"])
	}
	if account == "" {
		return "", errors.New("harness connector requires config account_id")
	}
	return account, nil
}

// harnessBaseURL resolves and validates the base URL. The default is
// app.harness.io; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk. The `api_url` config key is
// accepted as an alias matching the upstream spec.
func harnessBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = strings.TrimSpace(cfg.Config["api_url"])
	}
	if base == "" {
		return harnessDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("harness config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("harness config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("harness config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func harnessPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return harnessDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("harness config page_size must be an integer: %w", err)
	}
	if value < 1 || value > harnessMaxPageSize {
		return 0, fmt.Errorf("harness config page_size must be between 1 and %d", harnessMaxPageSize)
	}
	return value, nil
}

func harnessMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("harness config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("harness config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write satisfies the connectors.Connector interface. Harness is a read-only
// source connector (organization/project/service listings have no safe
// reverse-ETL write semantics), so every write is rejected.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	if err := ctx.Err(); err != nil {
		return connectors.WriteResult{}, err
	}
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
