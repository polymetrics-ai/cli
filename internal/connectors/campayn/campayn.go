// Package campayn implements the native pm Campayn connector, a declarative-HTTP
// per-system connector built on the connsdk toolkit. It mirrors the stripe
// reference connector's shape: a thin package that composes a connsdk Requester
// (with Campayn's TRUEREST apikey header auth) plus Campayn-specific stream
// definitions and endpoints.
//
// Campayn's API is read-only for our purposes (its write endpoints are
// documented as TODO upstream), so this connector exposes Check/Catalog/Read
// only and reports Write=false.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
package campayn

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/connsdk"
)

const (
	// campaignDefaultBaseURL is the placeholder base. In production the base URL
	// is derived from the per-account sub_domain config (see campaignBaseURL); the
	// default is only used as a fallback for fixture mode / catalog calls.
	campaignDefaultBaseURL = "https://app.campayn.com/api/v1"
	campaignUserAgent      = "polymetrics-go-cli"
	// campaignAuthPrefix is Campayn's custom Authorization header scheme:
	// "Authorization: TRUEREST apikey=<key>".
	campaignAuthPrefix = "TRUEREST apikey="
)

func init() {
	connectors.RegisterFactory("campayn", New)
}

// New returns the Campayn connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Campayn connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "campayn" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "campayn",
		DisplayName:     "Campayn",
		IntegrationType: "api",
		Description:     "Reads Campayn subscriber lists, signup forms, contacts, email campaigns, and calendar reports through the Campayn REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Campayn. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := campaignBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(campaignSecret(cfg)) == "" {
		return errors.New("campayn connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the lists collection confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "lists.json", nil, nil, nil); err != nil {
		return fmt.Errorf("check campayn: %w", err)
	}
	return nil
}

// Write is unsupported: Campayn's documented write surface is incomplete (marked
// TODO upstream), so this connector is read-only. It satisfies the Connector
// interface by reporting the shared unsupported-operation error.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: campaignStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "lists"
	}
	endpoint, ok := campaignStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("campayn stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	if endpoint.isSubstream() {
		return c.readSubstream(ctx, r, endpoint, emit)
	}
	return c.readResource(ctx, r, endpoint.resource, endpoint.mapRecord, emit)
}

// readResource reads a single Campayn collection endpoint. Campayn returns each
// collection as a bare top-level JSON array, so records are extracted at the root
// path ("") via connsdk.RecordsAt.
func (c Connector) readResource(ctx context.Context, r *connsdk.Requester, resource string, mapRecord func(map[string]any) connectors.Record, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, resource, nil, nil)
	if err != nil {
		return fmt.Errorf("read campayn %s: %w", resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode campayn %s: %w", resource, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// readSubstream reads a per-list substream (forms, contacts) by first listing
// every subscriber list, then fanning out one request per list id. Each emitted
// record is annotated with its parent list_id so the partition is recoverable
// downstream. This multi-request fan-out is Campayn's equivalent of pagination:
// the full result spans several upstream requests.
func (c Connector) readSubstream(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	listIDs, err := c.listIDs(ctx, r)
	if err != nil {
		return err
	}
	for _, listID := range listIDs {
		if err := ctx.Err(); err != nil {
			return err
		}
		resource := "lists/" + url.PathEscape(listID) + "/" + endpoint.listResource
		resp, err := r.Do(ctx, http.MethodGet, resource, nil, nil)
		if err != nil {
			return fmt.Errorf("read campayn %s: %w", resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode campayn %s: %w", resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			rec := endpoint.mapRecord(item)
			rec["list_id"] = listID
			if err := emit(rec); err != nil {
				return err
			}
		}
	}
	return nil
}

// listIDs fetches the parent list ids used to drive substream reads.
func (c Connector) listIDs(ctx context.Context, r *connsdk.Requester) ([]string, error) {
	resp, err := r.Do(ctx, http.MethodGet, "lists.json", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("read campayn lists.json: %w", err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return nil, fmt.Errorf("decode campayn lists.json: %w", err)
	}
	ids := make([]string, 0, len(records))
	for _, item := range records {
		if id := stringField(item, "id"); id != "" {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise campayn credential-free (mirrors the stripe
// fixture intent). Substream records carry the synthetic parent list_id.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                fmt.Sprintf("%s_fixture_%d", stream, i),
			"list_name":         fmt.Sprintf("Fixture List %d", i),
			"contact_count":     float64(10 * i),
			"tags":              "fixture",
			"contact_list_id":   "list_fixture_1",
			"form_title":        fmt.Sprintf("Fixture Form %d", i),
			"form_type":         "signup",
			"form_html":         "<form></form>",
			"signup_count":      fmt.Sprintf("%d", i),
			"email":             fmt.Sprintf("fixture+%d@example.com", i),
			"first_name":        fmt.Sprintf("First%d", i),
			"last_name":         fmt.Sprintf("Last%d", i),
			"confirmed":         "true",
			"image_url":         "",
			"name":              fmt.Sprintf("Fixture Campaign %d", i),
			"status":            "sent",
			"send_count":        fmt.Sprintf("%d", 100*i),
			"send_now":          false,
			"unique_views":      float64(5 * i),
			"percent_views":     float64(i),
			"unique_responses":  float64(i),
			"percent_responses": float64(i),
			"preview_url":       "https://example.com/preview",
			"preview_thumb":     "https://example.com/thumb.png",
			"scheduled_date":    "2026-01-01",
		}
		rec := endpoint.mapRecord(item)
		if endpoint.isSubstream() {
			rec["list_id"] = "list_fixture_1"
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Campayn's TRUEREST apikey
// header auth and the resolved base URL. The secret only ever flows into
// connsdk.APIKeyHeader; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := campaignBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := campaignSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("campayn connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", secret, campaignAuthPrefix),
		UserAgent: campaignUserAgent,
	}, nil
}

func campaignSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// campaignBaseURL resolves and validates the base URL. Precedence:
//  1. an explicit base_url config override (validated for SSRF: absolute
//     http/https with a host),
//  2. otherwise https://<sub_domain>.campayn.com/api/v1 when sub_domain is set,
//  3. otherwise the package default.
func campaignBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base != "" {
		parsed, err := url.Parse(base)
		if err != nil {
			return "", fmt.Errorf("campayn config base_url is invalid: %w", err)
		}
		if parsed.Scheme != "https" && parsed.Scheme != "http" {
			return "", fmt.Errorf("campayn config base_url must use http or https, got %q", parsed.Scheme)
		}
		if parsed.Host == "" {
			return "", errors.New("campayn config base_url must include a host")
		}
		return strings.TrimRight(base, "/"), nil
	}
	if sub := campaignSubDomain(cfg); sub != "" {
		if !validSubDomain(sub) {
			return "", fmt.Errorf("campayn config sub_domain %q must be a single DNS label (letters, digits, hyphen)", sub)
		}
		return "https://" + sub + ".campayn.com/api/v1", nil
	}
	return campaignDefaultBaseURL, nil
}

// validSubDomain reports whether sub is a safe single DNS label, preventing host
// injection (e.g. "evil.com/x" or "a.b") from corrupting the templated base URL.
func validSubDomain(sub string) bool {
	if sub == "" || len(sub) > 63 {
		return false
	}
	for _, r := range sub {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-':
		default:
			return false
		}
	}
	return sub[0] != '-' && sub[len(sub)-1] != '-'
}

// campaignSubDomain returns the configured sub_domain (the catalog also names it
// "domain"), sanitized to a bare host label to keep the templated base URL safe.
func campaignSubDomain(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return ""
	}
	sub := strings.TrimSpace(cfg.Config["sub_domain"])
	if sub == "" {
		sub = strings.TrimSpace(cfg.Config["domain"])
	}
	return sub
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
