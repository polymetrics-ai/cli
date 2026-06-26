// Package factorial implements the native pm Factorial (FactorialHR) connector.
// It is a declarative-HTTP per-system connector built on the same shape as the
// reference stripe package: a thin package that composes the connsdk toolkit
// (Requester + APIKeyHeader auth + RecordsAt extraction + page-increment
// pagination + cursor state) with Factorial-specific stream definitions and
// endpoints.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
//
// API: https://apidoc.factorialhr.com/  Auth: X-API-KEY header.
package factorial

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
	factorialDefaultBaseURL  = "https://api.factorialhr.com/api/v2/resources"
	factorialDefaultPageSize = 50
	factorialMaxPageSize     = 1000
	factorialUserAgent       = "polymetrics-go-cli"
	// factorialCheckResource is a cheap, side-effect-free list used by Check to
	// confirm auth + connectivity.
	factorialCheckResource = "api_public/credentials"
	// factorialAPIKeyHeader is the header carrying the API key.
	factorialAPIKeyHeader = "X-API-KEY"
)

func init() {
	connectors.RegisterFactory("factorial", New)
}

// New returns the Factorial connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Factorial connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "factorial" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "factorial",
		DisplayName:     "Factorial",
		IntegrationType: "api",
		Description:     "Reads FactorialHR employees, teams, time-off leaves, leave types, and locations through the Factorial REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Factorial.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := factorialBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(factorialSecret(cfg)) == "" {
		return errors.New("factorial connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the public credentials list confirms auth and
	// connectivity without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, factorialCheckResource, url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check factorial: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: factorialStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Factorial stream starts
// with an empty incremental cursor (full sync), which the start_date config can
// raise at read time.
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
		stream = "employees"
	}
	endpoint, ok := factorialStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("factorial stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := factorialPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := factorialMaxPages(req.Config)
	if err != nil {
		return err
	}

	// Factorial uses page-increment pagination: ?page=N&limit=size, with records
	// under data[]. PageNumberPaginator stops on the first short page.
	paginator := &connsdk.PageNumberPaginator{
		PageParam: "page",
		SizeParam: "limit",
		StartPage: 1,
		PageSize:  pageSize,
	}
	// The cursor maps to the largest updated_at seen so far for incremental
	// streams; mapping is data-driven so the harvest loop stays generic.
	maxCursor := connsdk.Cursor(req.State)
	wrap := func(item connsdk.Record) error {
		rec := endpoint.mapRecord(item)
		if endpoint.incremental {
			if v := stringField(item, "updated_at"); v != "" {
				maxCursor = connsdk.MaxCursor(maxCursor, v)
			}
			if maxCursor != "" {
				rec["_cursor"] = maxCursor
			}
		}
		return emit(rec)
	}

	if err := connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, url.Values{}, paginator, "data", maxPages, wrap); err != nil {
		return fmt.Errorf("read factorial %s: %w", stream, err)
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise factorial credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		record := endpoint.mapRecord(fixtureItem(stream, i))
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with X-API-KEY header auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := factorialBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := factorialSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("factorial connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(factorialAPIKeyHeader, secret, ""),
		UserAgent: factorialUserAgent,
	}, nil
}

func factorialSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// factorialBaseURL resolves and validates the base URL. The default is
// api.factorialhr.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func factorialBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return factorialDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("factorial config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("factorial config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("factorial config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

// factorialPageSize resolves the per-page limit from the `limit` config (the
// Factorial catalog field name).
func factorialPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["limit"])
	if raw == "" {
		raw = strings.TrimSpace(cfg.Config["page_size"])
	}
	if raw == "" {
		return factorialDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("factorial config limit must be an integer: %w", err)
	}
	if value < 1 || value > factorialMaxPageSize {
		return 0, fmt.Errorf("factorial config limit must be between 1 and %d", factorialMaxPageSize)
	}
	return value, nil
}

func factorialMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("factorial config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("factorial config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func stringField(item map[string]any, key string) string {
	switch v := item[key].(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}

// Write is unsupported: Factorial is read-only here. It satisfies the Connector
// interface but always returns ErrUnsupportedOperation.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
