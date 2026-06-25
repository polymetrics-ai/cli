// Package coassemble implements the native pm Coassemble connector. It follows
// the declarative-HTTP shape established by the stripe connector: a thin package
// that composes the connsdk toolkit (Requester + a custom Authorization header +
// root-array extraction) with Coassemble-specific stream definitions and the
// page-increment pagination its headless API uses.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// Coassemble's headless API has no incremental cursor, so the connector is full
// refresh and read-only. Authentication uses a single Authorization header of the
// form "COASSEMBLE-V1-SHA256 UserId=<user_id>, UserToken=<user_token>"; neither
// secret is ever logged.
package coassemble

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
	coassembleDefaultBaseURL  = "https://api.coassemble.com"
	coassembleDefaultPageSize = 20
	coassembleMaxPageSize     = 100
	coassembleUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("coassemble", New)
}

// New returns the Coassemble connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Coassemble connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "coassemble" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "coassemble",
		DisplayName:     "Coassemble",
		IntegrationType: "api",
		Description:     "Reads Coassemble courses, screen types, and learner tracking records through the Coassemble headless REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Coassemble. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := coassembleBaseURL(cfg); err != nil {
		return err
	}
	if err := requireSecrets(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the first page of courses confirms auth and connectivity
	// without mutating anything.
	query := url.Values{"page": []string{"1"}, "length": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "/api/v1/headless/courses", query, nil, nil); err != nil {
		return fmt.Errorf("check coassemble: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: coassembleStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "courses"
	}
	endpoint, ok := coassembleStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("coassemble stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := coassemblePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := coassembleMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Coassemble's page-increment pagination. The headless list
// endpoints return a top-level JSON array; the next page is requested with an
// incremented page number (starting at 1) and a length page size. A page that
// returns fewer than length records (or zero) ends the walk. screen_types is not
// paginated, so it is read as a single request.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 1; ; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if maxPages > 0 && page > maxPages {
			return nil
		}
		query := url.Values{}
		if endpoint.paginated {
			query.Set("page", strconv.Itoa(page))
			query.Set("length", strconv.Itoa(pageSize))
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.path, query, nil)
		if err != nil {
			return fmt.Errorf("read coassemble %s: %w", endpoint.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode coassemble %s page: %w", endpoint.path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// Un-paginated endpoints are complete after one request. Paginated ones
		// stop on a short page (the last page) or an empty page.
		if !endpoint.paginated || len(records) < pageSize {
			return nil
		}
	}
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise coassemble credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":          i,
			"title":       fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"name":        fmt.Sprintf("fixture_%d", i),
			"description": "Deterministic fixture record.",
			"key":         fmt.Sprintf("%s_fixture_%d", stream, i),
			"icon":        "fixture",
			"image":       "",
			"active":      true,
			"private":     false,
			"paid":        false,
			"identified":  false,
			"is_sharable": true,
			"premium":     false,
			"course_id":   1000 + i,
			"identifier":  fmt.Sprintf("learner_%d", i),
			"status":      "in_progress",
			"progress":    50 * i,
			"completed":   false,
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the Coassemble custom
// Authorization header and the resolved base URL. The secrets only ever flow into
// the header value; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := coassembleBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	if err := requireSecrets(cfg); err != nil {
		return nil, err
	}
	userID, userToken := coassembleSecrets(cfg)
	authValue := fmt.Sprintf("COASSEMBLE-V1-SHA256 UserId=%s, UserToken=%s", userID, userToken)
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", authValue, ""),
		UserAgent: coassembleUserAgent,
	}, nil
}

func coassembleSecrets(cfg connectors.RuntimeConfig) (userID, userToken string) {
	if cfg.Secrets == nil {
		return "", ""
	}
	return strings.TrimSpace(cfg.Secrets["user_id"]), strings.TrimSpace(cfg.Secrets["user_token"])
}

func requireSecrets(cfg connectors.RuntimeConfig) error {
	userID, userToken := coassembleSecrets(cfg)
	if userID == "" {
		return errors.New("coassemble connector requires secret user_id")
	}
	if userToken == "" {
		return errors.New("coassemble connector requires secret user_token")
	}
	return nil
}

// coassembleBaseURL resolves and validates the base URL. The default is
// api.coassemble.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func coassembleBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return coassembleDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("coassemble config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("coassemble config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("coassemble config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func coassemblePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return coassembleDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("coassemble config page_size must be an integer: %w", err)
	}
	if value < 1 || value > coassembleMaxPageSize {
		return 0, fmt.Errorf("coassemble config page_size must be between 1 and %d", coassembleMaxPageSize)
	}
	return value, nil
}

func coassembleMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("coassemble config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("coassemble config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is not supported: Coassemble's headless API is read-oriented for this
// connector, so the connector is read-only (Capabilities.Write=false).
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
