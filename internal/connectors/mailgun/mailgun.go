// Package mailgun implements the native pm Mailgun connector. It is a
// declarative-HTTP per-system connector built on the stripe template: a thin
// package that composes the connsdk toolkit (Requester + Basic auth + RecordsAt
// extraction + cursor state) with Mailgun-specific stream definitions and
// endpoints.
//
// Mailgun authenticates with HTTP Basic auth using the username "api" and the
// account private API key as the password. List endpoints under the v3 API
// return an {"items":[...]} envelope; /v3/domains pages with skip/limit while
// most sub-resources page by following the absolute paging.next URL.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package mailgun

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
	mailgunDefaultBaseURLUS = "https://api.mailgun.net"
	mailgunDefaultBaseURLEU = "https://api.eu.mailgun.net"
	mailgunDefaultPageSize  = 100
	mailgunMaxPageSize      = 1000
	mailgunUserAgent        = "polymetrics-go-cli"
	mailgunBasicUser        = "api"
)

func init() {
	connectors.RegisterFactory("mailgun", New)
}

// New returns the Mailgun connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Mailgun connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "mailgun" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "mailgun",
		DisplayName:     "Mailgun",
		IntegrationType: "api",
		Description:     "Reads Mailgun domains, email events, mailing lists, and tags through the Mailgun v3 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Mailgun. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := mailgunBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(mailgunSecret(cfg)) == "" {
		return errors.New("mailgun connector requires secret private_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the domains list confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "v3/domains", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check mailgun: %w", err)
	}
	return nil
}

// Write is unsupported: Mailgun is a read-only source connector. It returns
// ErrUnsupportedOperation to satisfy the connectors.Connector interface.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: mailgunStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Mailgun stream starts with
// an empty incremental cursor (full sync), which the start_date config can raise
// at read time.
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
		stream = "domains"
	}
	endpoint, ok := mailgunStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("mailgun stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	resource, err := resolveResource(endpoint, req.Config)
	if err != nil {
		return err
	}
	pageSize, err := mailgunPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := mailgunMaxPages(req.Config)
	if err != nil {
		return err
	}

	switch endpoint.pagination {
	case paginationOffset:
		return c.harvestOffset(ctx, r, resource, endpoint, pageSize, maxPages, emit)
	default:
		return c.harvestPagingNext(ctx, r, resource, endpoint, pageSize, maxPages, emit)
	}
}

// harvestOffset drives Mailgun's skip/limit pagination over the {items,...}
// envelope (used by /v3/domains). It stops on a short page or when maxPages is
// reached.
func (c Connector) harvestOffset(ctx context.Context, r *connsdk.Requester, resource string, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	skip := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("skip", strconv.Itoa(skip))
		resp, err := r.Do(ctx, http.MethodGet, resource, query, nil)
		if err != nil {
			return fmt.Errorf("read mailgun %s: %w", resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "items")
		if err != nil {
			return fmt.Errorf("decode mailgun %s page: %w", resource, err)
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
		skip += pageSize
	}
	return nil
}

// harvestPagingNext drives Mailgun's paging.next pagination over the
// {items,paging:{next}} envelope (used by events, mailing lists, and tags). The
// next page is the absolute URL at paging.next; the loop stops when next is
// empty, repeats, or yields no items.
func (c Connector) harvestPagingNext(ctx context.Context, r *connsdk.Requester, resource string, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	// First request carries a limit; subsequent requests use the absolute next
	// URL (which already encodes its own cursor/limit).
	next := ""
	firstQuery := url.Values{}
	firstQuery.Set("limit", strconv.Itoa(pageSize))
	seen := map[string]bool{}

	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		path := resource
		var query url.Values
		if next == "" {
			query = firstQuery
		} else {
			if seen[next] {
				return nil
			}
			seen[next] = true
			path = next
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read mailgun %s: %w", resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "items")
		if err != nil {
			return fmt.Errorf("decode mailgun %s page: %w", resource, err)
		}
		if len(records) == 0 {
			return nil
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		nextURL, err := connsdk.StringAt(resp.Body, "paging.next")
		if err != nil {
			return fmt.Errorf("decode mailgun %s paging: %w", resource, err)
		}
		nextURL = strings.TrimSpace(nextURL)
		if nextURL == "" || nextURL == path {
			return nil
		}
		next = nextURL
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise mailgun credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":            fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":          fmt.Sprintf("fixture-%d.example.com", i),
			"address":       fmt.Sprintf("list%d@example.com", i),
			"tag":           fmt.Sprintf("tag_%d", i),
			"state":         "active",
			"type":          "custom",
			"event":         "delivered",
			"timestamp":     1577836800 + int64(i),
			"recipient":     fmt.Sprintf("user+%d@example.com", i),
			"message-id":    fmt.Sprintf("<msg-%d@example.com>", i),
			"log-level":     "info",
			"description":   fmt.Sprintf("Fixture %d", i),
			"access_level":  "readonly",
			"members_count": int64(i),
			"is_disabled":   false,
			"created_at":    "Wed, 01 Jan 2020 00:00:00 GMT",
			"first-seen":    "Wed, 01 Jan 2020 00:00:00 GMT",
			"last-seen":     "Wed, 01 Jan 2020 00:00:00 GMT",
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

// requester builds a connsdk.Requester wired with Basic auth and the resolved
// base URL. The secret only ever flows into connsdk.Basic; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := mailgunBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := mailgunSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("mailgun connector requires secret private_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(mailgunBasicUser, secret),
		UserAgent: mailgunUserAgent,
	}, nil
}

// resolveResource substitutes the {domain} token in domain-scoped resource
// paths with the configured domain_name.
func resolveResource(endpoint streamEndpoint, cfg connectors.RuntimeConfig) (string, error) {
	if !endpoint.needsDomain {
		return endpoint.resource, nil
	}
	domain := strings.TrimSpace(cfg.Config["domain_name"])
	if domain == "" {
		return "", errors.New("mailgun config domain_name is required for this stream")
	}
	return strings.ReplaceAll(endpoint.resource, "{domain}", domain), nil
}

func mailgunSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["private_key"]
}

// mailgunBaseURL resolves and validates the base URL. The default is the US or
// EU Mailgun host selected by domain_region; any base_url override must be an
// absolute https (or http for local test servers) URL with a host to bound SSRF
// risk.
func mailgunBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		region := strings.ToUpper(strings.TrimSpace(cfg.Config["domain_region"]))
		if region == "EU" {
			return mailgunDefaultBaseURLEU, nil
		}
		return mailgunDefaultBaseURLUS, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("mailgun config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("mailgun config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("mailgun config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func mailgunPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return mailgunDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("mailgun config page_size must be an integer: %w", err)
	}
	if value < 1 || value > mailgunMaxPageSize {
		return 0, fmt.Errorf("mailgun config page_size must be between 1 and %d", mailgunMaxPageSize)
	}
	return value, nil
}

func mailgunMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("mailgun config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("mailgun config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
