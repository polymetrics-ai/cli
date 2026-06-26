// Package mailosaur implements the native pm Mailosaur connector. Mailosaur is an
// email and SMS testing service; this connector reads its virtual servers, the
// message summaries within a server, and account usage transactions through the
// Mailosaur REST API.
//
// It follows the declarative-HTTP template established by the stripe connector: a
// thin package that composes the connsdk toolkit (Requester + Basic auth +
// RecordsAt extraction) with Mailosaur-specific stream definitions and endpoints.
// It is read-only: Mailosaur has no reverse-ETL-shaped writes.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package mailosaur

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
	mailosaurDefaultBaseURL  = "https://mailosaur.com/api"
	mailosaurDefaultUsername = "api"
	mailosaurDefaultPageSize = 50
	mailosaurMaxPageSize     = 1000
	mailosaurUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("mailosaur", New)
}

// New returns the Mailosaur connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Mailosaur connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "mailosaur" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "mailosaur",
		DisplayName:     "Mailosaur",
		IntegrationType: "api",
		Description:     "Reads Mailosaur virtual servers, message summaries, and account usage transactions through the Mailosaur REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Mailosaur. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := mailosaurBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(mailosaurSecret(cfg)) == "" {
		return errors.New("mailosaur connector requires secret password (api key)")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// Listing servers confirms auth and connectivity without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "servers", nil, nil, nil); err != nil {
		return fmt.Errorf("check mailosaur: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: mailosaurStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Mailosaur stream starts with
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
		stream = "servers"
	}
	endpoint, ok := mailosaurStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("mailosaur stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}

	base := url.Values{}
	if endpoint.needsServer {
		server := strings.TrimSpace(req.Config.Config["server"])
		if server == "" {
			return fmt.Errorf("mailosaur stream %q requires config server (server id)", stream)
		}
		base.Set("server", server)
	}
	if receivedAfter := strings.TrimSpace(req.Config.Config["received_after"]); receivedAfter != "" && stream == "messages" {
		base.Set("receivedAfter", receivedAfter)
	}

	if !endpoint.paginated {
		return c.readSinglePage(ctx, r, endpoint, base, emit)
	}

	pageSize, err := mailosaurPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := mailosaurMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, base, pageSize, maxPages, emit)
}

// readSinglePage fetches a non-paginated endpoint once and emits its records.
func (c Connector) readSinglePage(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, base url.Values, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, base, nil)
	if err != nil {
		return fmt.Errorf("read mailosaur %s: %w", endpoint.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return fmt.Errorf("decode mailosaur %s: %w", endpoint.resource, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// harvest drives Mailosaur's page/itemsPerPage pagination. Pages are zero-indexed;
// a page returning fewer records than itemsPerPage signals the end.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, base url.Values, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("page", strconv.Itoa(page))
		query.Set("itemsPerPage", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read mailosaur %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode mailosaur %s page: %w", endpoint.resource, err)
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
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise mailosaur credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var record connectors.Record
		switch stream {
		case "servers":
			record = connectors.Record{
				"id":       fmt.Sprintf("srv_fixture_%d", i),
				"name":     fmt.Sprintf("Fixture Server %d", i),
				"users":    []any{},
				"messages": int64(i),
			}
		case "transactions":
			record = connectors.Record{
				"timestamp": fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
				"email":     int64(i),
				"sms":       int64(0),
				"previews":  int64(0),
			}
		default: // messages
			record = connectors.Record{
				"id":       fmt.Sprintf("msg_fixture_%d", i),
				"received": fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
				"type":     "Email",
				"subject":  fmt.Sprintf("Fixture message %d", i),
				"server":   "srv_fixture_1",
				"from":     []any{map[string]any{"email": "sender@example.com"}},
				"to":       []any{map[string]any{"email": "rcpt@example.com"}},
				"cc":       []any{},
				"bcc":      []any{},
			}
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Basic auth and the resolved base
// URL. Mailosaur accepts the API key as the Basic auth value with the literal
// username "api" (overridable via config). The secret only ever flows into
// connsdk.Basic; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := mailosaurBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := mailosaurSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("mailosaur connector requires secret password (api key)")
	}
	username := strings.TrimSpace(cfg.Config["username"])
	if username == "" {
		username = mailosaurDefaultUsername
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(username, secret),
		UserAgent: mailosaurUserAgent,
	}, nil
}

func mailosaurSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["password"]
}

// mailosaurBaseURL resolves and validates the base URL. The default is
// mailosaur.com/api; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func mailosaurBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return mailosaurDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("mailosaur config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("mailosaur config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("mailosaur config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func mailosaurPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["items_per_page"])
	if raw == "" {
		return mailosaurDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("mailosaur config items_per_page must be an integer: %w", err)
	}
	if value < 1 || value > mailosaurMaxPageSize {
		return 0, fmt.Errorf("mailosaur config items_per_page must be between 1 and %d", mailosaurMaxPageSize)
	}
	return value, nil
}

func mailosaurMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("mailosaur config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("mailosaur config max_pages must be 0 for unlimited or a positive integer")
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

// Write is unsupported: Mailosaur is a read-only source for pm. It satisfies the
// connectors.Connector interface by returning ErrUnsupportedOperation.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
