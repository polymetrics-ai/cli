// Package drift implements the native pm Drift connector. It follows the stripe
// reference shape for declarative-HTTP connectors: a thin package that composes
// the connsdk toolkit (Requester + Bearer auth + RecordsAt/StringAt extraction)
// with Drift-specific stream definitions, endpoints, and pagination.
//
// Drift is a conversational-marketing platform. This connector reads users,
// accounts, conversations, and contacts from the Drift REST API
// (https://driftapi.com) using a Bearer access token. It is read-only.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package drift

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
	driftDefaultBaseURL = "https://driftapi.com"
	driftUserAgent      = "polymetrics-go-cli"
	// driftAccountsPageSize is the per-page batch size for the accounts stream
	// (Drift caps size at 65, default 10).
	driftAccountsPageSize = 65
	// driftConversationsPageSize is the per-page batch size for conversations
	// (Drift caps limit at 50, default 25).
	driftConversationsPageSize = 50
	// driftMaxPages bounds pagination to avoid unbounded loops on a
	// misbehaving server. 0 would mean unlimited; we use a generous ceiling.
	driftMaxPages = 10000
	// driftFixtureCreated is the deterministic timestamp used by fixture-mode
	// records (2026-01-01T00:00:00Z in unix seconds).
	driftFixtureCreated int64 = 1767225600
)

func init() {
	connectors.RegisterFactory("drift", New)
}

// New returns the Drift connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Drift connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "drift" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "drift",
		DisplayName:     "Drift",
		IntegrationType: "api",
		Description:     "Reads Drift users, accounts, conversations, and contacts through the Drift REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Drift. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := driftBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(driftSecret(cfg)) == "" {
		return errors.New("drift connector requires secret credentials.access_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// Listing users is a cheap, non-mutating read that confirms auth and
	// connectivity.
	if err := r.DoJSON(ctx, http.MethodGet, "users/list", nil, nil, nil); err != nil {
		return fmt.Errorf("check drift: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: driftStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader. Drift only supports full
// refresh today, so a stream starts with an empty cursor.
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
		stream = "users"
	}
	endpoint, ok := driftStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("drift stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, req.Config, emit)
}

// harvest drives the per-stream pagination for the Drift API. Each of the three
// pagination shapes is handled in one loop, built on connsdk.Requester +
// connsdk.RecordsAt + connsdk.StringAt; no single connsdk paginator covers all
// three Drift styles.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	query := url.Values{}
	switch endpoint.pagination {
	case paginationNextURL:
		query.Set("size", strconv.Itoa(driftAccountsPageSize))
		query.Set("index", "0")
	case paginationCursor:
		query.Set("limit", strconv.Itoa(driftConversationsPageSize))
	}
	// The contacts stream is looked up by email; pass it through when provided.
	if endpoint.resource == "contacts" {
		if email := strings.TrimSpace(cfg.Config["email"]); email != "" {
			query.Set("email", email)
		}
	}

	path := endpoint.resource
	for page := 0; page < driftMaxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read drift %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode drift %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}

		switch endpoint.pagination {
		case paginationNone:
			return nil
		case paginationNextURL:
			next, err := connsdk.StringAt(resp.Body, endpoint.nextPath)
			if err != nil {
				return fmt.Errorf("decode drift %s next: %w", endpoint.resource, err)
			}
			if strings.TrimSpace(next) == "" {
				return nil
			}
			// next is an absolute URL; the Requester treats http(s) paths as
			// absolute, and query already lives in the URL so clear it.
			path = next
			query = url.Values{}
		case paginationCursor:
			more, err := connsdk.StringAt(resp.Body, endpoint.morePath)
			if err != nil {
				return fmt.Errorf("decode drift %s more: %w", endpoint.resource, err)
			}
			next, err := connsdk.StringAt(resp.Body, endpoint.nextPath)
			if err != nil {
				return fmt.Errorf("decode drift %s next: %w", endpoint.resource, err)
			}
			if more != "true" || strings.TrimSpace(next) == "" {
				return nil
			}
			query.Set("next", next)
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise drift credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":             int64(i),
			"accountId":      fmt.Sprintf("acct_fixture_%d", i),
			"orgId":          int64(42),
			"name":           fmt.Sprintf("Fixture %d", i),
			"alias":          fmt.Sprintf("fix%d", i),
			"email":          fmt.Sprintf("fixture+%d@example.com", i),
			"domain":         "example.com",
			"role":           "member",
			"availability":   "AVAILABLE",
			"status":         "open",
			"contactId":      int64(100 + i),
			"inboxId":        int64(7),
			"bot":            false,
			"verified":       true,
			"deleted":        false,
			"targeted":       false,
			"createdAt":      driftFixtureCreated + int64(i),
			"updatedAt":      driftFixtureCreated + int64(i),
			"createDateTime": driftFixtureCreated + int64(i),
			"updateDateTime": driftFixtureCreated + int64(i),
			"attributes":     map[string]any{"email": fmt.Sprintf("fixture+%d@example.com", i)},
		}
		record := endpoint.mapRecord(item)
		record["connector"] = "drift"
		record["stream"] = stream
		record["fixture"] = true
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
	base, err := driftBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := driftSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("drift connector requires secret credentials.access_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: driftUserAgent,
	}, nil
}

// driftSecret resolves the Drift access token from the runtime secrets. It
// accepts both the canonical dotted key and a couple of common aliases so the
// connector is robust to how the secret is surfaced.
func driftSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	for _, key := range []string{"credentials.access_token", "access_token", "credentials_access_token"} {
		if v := strings.TrimSpace(cfg.Secrets[key]); v != "" {
			return v
		}
	}
	return ""
}

// driftBaseURL resolves and validates the base URL. The default is
// driftapi.com; any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func driftBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return driftDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("drift config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("drift config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("drift config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

// Write satisfies the connectors.Connector interface. Drift is read-only in
// this connector, so writes are unsupported.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
