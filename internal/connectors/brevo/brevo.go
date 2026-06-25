// Package brevo implements the native pm Brevo (formerly Sendinblue) connector.
// It is a declarative-HTTP per-system connector that composes the connsdk
// toolkit (Requester + api-key header auth + RecordsAt extraction + offset
// pagination + cursor state) with Brevo-specific stream definitions and
// endpoints. It follows the stripe reference connector's shape.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
package brevo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/connsdk"
)

const (
	brevoDefaultBaseURL  = "https://api.brevo.com/v3"
	brevoDefaultPageSize = 100
	brevoMaxPageSize     = 1000
	brevoUserAgent       = "polymetrics-go-cli"
	// brevoFixtureModified is the deterministic modifiedAt timestamp used by the
	// fixture-mode records.
	brevoFixtureModified = "2026-01-01T00:00:00.000+00:00"
)

func init() {
	connectors.RegisterFactory("brevo", New)
}

// New returns the Brevo connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Brevo connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "brevo" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "brevo",
		DisplayName:     "Brevo",
		IntegrationType: "api",
		Description:     "Reads Brevo (formerly Sendinblue) contacts, email campaigns, contact lists, and senders through the Brevo REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Brevo. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := brevoBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(brevoSecret(cfg)) == "" {
		return errors.New("brevo connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the account endpoint confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "account", nil, nil, nil); err != nil {
		return fmt.Errorf("check brevo: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: brevoStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Brevo stream starts with an
// empty incremental cursor (full sync), which the start_date config can raise at
// read time.
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
		stream = "contacts"
	}
	endpoint, ok := brevoStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("brevo stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := brevoPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := brevoMaxPages(req.Config)
	if err != nil {
		return err
	}
	modifiedSince := incrementalLowerBound(req)
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, modifiedSince, emit)
}

// harvest drives Brevo's offset/limit pagination. List endpoints return
// {<recordsPath>:[...], count:N}; the next page is requested by advancing offset
// by the number of records returned, stopping when a short page (fewer than
// pageSize records) comes back. Non-paginated endpoints (e.g. senders) make a
// single request. There is no exact connsdk paginator for this body shape, so
// the loop lives here, built on connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, modifiedSince string, emit func(connectors.Record) error) error {
	base := url.Values{}
	if endpoint.paginated {
		base.Set("limit", strconv.Itoa(pageSize))
	}
	if endpoint.supportsModifiedSince && modifiedSince != "" {
		base.Set("modifiedSince", modifiedSince)
	}

	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if endpoint.paginated {
			query.Set("offset", strconv.Itoa(offset))
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read brevo %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode brevo %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// Non-paginated endpoints return everything in one request.
		if !endpoint.paginated {
			return nil
		}
		// A short (or empty) page means we have exhausted the resource.
		if len(records) < pageSize {
			return nil
		}
		offset += len(records)
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise brevo credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                int64(i),
			"email":             fmt.Sprintf("fixture+%d@example.com", i),
			"emailBlacklisted":  false,
			"smsBlacklisted":    false,
			"createdAt":         brevoFixtureModified,
			"modifiedAt":        brevoFixtureModified,
			"listIds":           []any{int64(1)},
			"attributes":        map[string]any{"connector": "brevo", "fixture": true},
			"name":              fmt.Sprintf("Fixture %s %d", stream, i),
			"subject":           fmt.Sprintf("Fixture subject %d", i),
			"type":              "classic",
			"status":            "sent",
			"totalSubscribers":  int64(10 * i),
			"totalBlacklisted":  int64(0),
			"uniqueSubscribers": int64(10 * i),
			"folderId":          int64(1),
			"active":            true,
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

// requester builds a connsdk.Requester wired with api-key header auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := brevoBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := brevoSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("brevo connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("api-key", secret, ""),
		UserAgent: brevoUserAgent,
	}, nil
}

// incrementalLowerBound returns the modifiedSince lower bound, derived from the
// incremental cursor (if any) or else the start_date config. An empty result
// means no lower bound (full sync).
func incrementalLowerBound(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	return strings.TrimSpace(req.Config.Config["start_date"])
}

func brevoSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// brevoBaseURL resolves and validates the base URL. The default is api.brevo.com;
// any override must be an absolute https (or http for local test servers) URL
// with a host to bound SSRF risk.
func brevoBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return brevoDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("brevo config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("brevo config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("brevo config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func brevoPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return brevoDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("brevo config page_size must be an integer: %w", err)
	}
	if value < 1 || value > brevoMaxPageSize {
		return 0, fmt.Errorf("brevo config page_size must be between 1 and %d", brevoMaxPageSize)
	}
	return value, nil
}

func brevoMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("brevo config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("brevo config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// Write satisfies the connectors.Connector interface. Brevo is read-only in this
// connector (no allow-listed reverse-ETL actions are shipped), so writes are
// unsupported and the metadata reports Write=false.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
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
