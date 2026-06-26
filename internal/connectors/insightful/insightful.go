// Package insightful implements the native pm Insightful connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference connector: a thin package that composes the connsdk toolkit
// (Requester + Bearer auth + record extraction + cursor state) with
// Insightful-specific stream definitions, endpoints, and pagination.
//
// Insightful (workforce analytics, formerly Workpuls) exposes a token-authed
// REST API at https://app.insightful.io/api/v1. List endpoints return a
// top-level JSON array; some responses are wrapped in an envelope that carries a
// `next` cursor token echoed back as a `next` request parameter. This connector
// is read-only (the Insightful API has no safe reverse-ETL writes for the core
// resources we sync).
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package insightful

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	insightfulDefaultBaseURL = "https://app.insightful.io/api/v1"
	insightfulUserAgent      = "polymetrics-go-cli"
	// insightfulMaxPages bounds the cursor loop as a safety net against a server
	// that never stops returning a `next` token.
	insightfulMaxPages = 10000
	// insightfulFixtureUpdated is the deterministic `updatedAt` (unix millis)
	// used by fixture-mode records.
	insightfulFixtureUpdated int64 = 1767225600000
)

func init() {
	connectors.RegisterFactory("insightful", New)
}

// New returns the Insightful connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Insightful connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "insightful" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "insightful",
		DisplayName:     "Insightful",
		IntegrationType: "api",
		Description:     "Reads Insightful workforce-analytics employees, teams, projects, and directory entries through the Insightful REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Insightful.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := insightfulBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(insightfulSecret(cfg)) == "" {
		return errors.New("insightful connector requires secret api_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the team list confirms auth and connectivity without
	// mutating anything.
	if _, err := r.Do(ctx, http.MethodGet, "team", nil, nil); err != nil {
		return fmt.Errorf("check insightful: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: insightfulStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: an Insightful stream starts
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
		stream = "employee"
	}
	endpoint, ok := insightfulStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("insightful stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	lower, err := incrementalLowerBound(req)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, lower, emit)
}

// Write satisfies the connectors.Connector interface. Insightful is exposed
// read-only (no safe reverse-ETL writes for the synced resources), so writes are
// rejected.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// harvest drives Insightful's cursor pagination. List endpoints return either a
// bare top-level JSON array (single page, no cursor) or an envelope
// {data:[...], next:"<token>"}; the next page is requested with next=<token>
// until the token is absent. The loop lives here, built on connsdk.Requester +
// connsdk.RecordsAt + connsdk.StringAt, because connsdk's CursorPaginator assumes
// a fixed records path while Insightful varies between root-array and enveloped
// shapes.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, startGTE string, emit func(connectors.Record) error) error {
	base := url.Values{}
	for k, v := range endpoint.query {
		base.Set(k, v)
	}
	if startGTE != "" {
		base.Set("start", startGTE)
	}

	next := ""
	for page := 0; page < insightfulMaxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if next != "" {
			query.Set("next", next)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read insightful %s: %w", endpoint.resource, err)
		}

		records, err := extractRecords(resp.Body)
		if err != nil {
			return fmt.Errorf("decode insightful %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}

		token, err := connsdk.StringAt(resp.Body, "next")
		if err != nil {
			token = ""
		}
		token = strings.TrimSpace(token)
		if token == "" || token == next {
			return nil
		}
		next = token
	}
	return nil
}

// extractRecords pulls the record array from an Insightful response, handling
// both the bare top-level array shape and the {data:[...]} envelope.
// connsdk.RecordsAt returns []map[string]any (connsdk.Record is a map alias),
// which the per-stream mappers consume directly.
//
// The shape is disambiguated by the first non-whitespace byte: a leading '['
// means a bare top-level array (the common list-resource shape), while a leading
// '{' means an envelope whose records live under data[]. Using RecordsAt("") on
// an object would wrongly wrap the whole envelope as a single record, so the
// distinction matters.
func extractRecords(body []byte) ([]map[string]any, error) {
	if isJSONArray(body) {
		return connsdk.RecordsAt(body, "")
	}
	return connsdk.RecordsAt(body, "data")
}

// isJSONArray reports whether body's first non-whitespace byte is '[' .
func isJSONArray(body []byte) bool {
	for _, b := range body {
		switch b {
		case ' ', '\t', '\n', '\r':
			continue
		case '[':
			return true
		default:
			return false
		}
	}
	return false
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free (mirrors the
// stripe fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":             fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"name":           fmt.Sprintf("Fixture %d", i),
			"email":          fmt.Sprintf("fixture+%d@example.com", i),
			"modelName":      titleCase(stream),
			"description":    "fixture record",
			"organizationId": "org_fixture_1",
			"creatorId":      "emp_fixture_1",
			"default":        false,
			"archived":       false,
			"billable":       true,
			"employees":      []any{"emp_fixture_1"},
			"projects":       []any{"project_fixture_1"},
			"createdAt":      insightfulFixtureUpdated,
			"updatedAt":      insightfulFixtureUpdated + int64(i),
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

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := insightfulBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := insightfulSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("insightful connector requires secret api_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: insightfulUserAgent,
	}, nil
}

// incrementalLowerBound returns the unix-millis lower bound for the `start`
// query param, derived from the incremental cursor (if any) or else the
// start_date config. An empty result means no lower bound (full sync).
func incrementalLowerBound(req connectors.ReadRequest) (string, error) {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor, nil
	}
	startDate := strings.TrimSpace(req.Config.Config["start_date"])
	if startDate == "" {
		return "", nil
	}
	t, err := time.Parse(time.RFC3339, startDate)
	if err != nil {
		return "", fmt.Errorf("insightful config start_date must be RFC3339: %w", err)
	}
	return fmt.Sprintf("%d", t.UnixMilli()), nil
}

func insightfulSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_token"]
}

// insightfulBaseURL resolves and validates the base URL. The default is
// app.insightful.io; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func insightfulBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return insightfulDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("insightful config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("insightful config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("insightful config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// titleCase upper-cases the first rune of s (ASCII), avoiding the deprecated
// strings.Title. Used only to label fixture records.
func titleCase(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	if r[0] >= 'a' && r[0] <= 'z' {
		r[0] -= 'a' - 'A'
	}
	return string(r)
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
