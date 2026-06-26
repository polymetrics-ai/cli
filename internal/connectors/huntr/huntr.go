// Package huntr implements the native pm Huntr connector. It is a declarative-
// HTTP per-system connector built on the stripe template shape: a thin package
// that composes the connsdk toolkit (Requester + Bearer auth + RecordsAt
// extraction) with Huntr-specific stream definitions, endpoints, and the
// organization REST API's next-cursor pagination.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// Huntr exposes only full_refresh reads over its organization API, so this
// connector is read-only (Capabilities.Write = false).
package huntr

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
	huntrDefaultBaseURL  = "https://api.huntr.co/org"
	huntrDefaultPageSize = 100
	huntrMaxPageSize     = 100
	huntrUserAgent       = "polymetrics-go-cli"
	// huntrFixtureCreated is the deterministic createdAt value used by the
	// fixture-mode records (2026-01-01T00:00:00Z in unix seconds).
	huntrFixtureCreated int64 = 1767225600
)

func init() {
	connectors.RegisterFactory("huntr", New)
}

// New returns the Huntr connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Huntr connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "huntr" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "huntr",
		DisplayName:     "Huntr",
		IntegrationType: "api",
		Description:     "Reads Huntr organization members, candidates, activities, notes, and actions through the Huntr REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Huntr. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := huntrBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(huntrSecret(cfg)) == "" {
		return errors.New("huntr connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the members list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "members", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check huntr: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: huntrStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "members"
	}
	endpoint, ok := huntrStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("huntr stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := huntrPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := huntrMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Huntr's next-cursor pagination. Huntr lists return
// {data:[...], next:<token|null>}; the next page is requested with next=<token>
// and a limit. The loop stops when the response `next` is empty or null. There
// is no exact connsdk paginator for the body-token shape with this stop
// condition, so the loop lives here, built on connsdk.Requester +
// connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	next := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		if next != "" {
			query.Set("next", next)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read huntr %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode huntr %s page: %w", endpoint.resource, err)
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
			return fmt.Errorf("decode huntr %s next: %w", endpoint.resource, err)
		}
		token = strings.TrimSpace(token)
		// Huntr signals exhaustion with an empty, null (""), or "inf" cursor.
		if token == "" || token == "inf" || token == next {
			return nil
		}
		next = token
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise huntr credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":               fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"email":            fmt.Sprintf("fixture+%d@example.com", i),
			"fullName":         fmt.Sprintf("Fixture %d", i),
			"givenName":        "Fixture",
			"familyName":       strconv.Itoa(i),
			"firstName":        "Fixture",
			"lastName":         strconv.Itoa(i),
			"isActive":         true,
			"createdAt":        huntrFixtureCreated + int64(i),
			"lastSeenAt":       huntrFixtureCreated + int64(i),
			"startAt":          huntrFixtureCreated + int64(i),
			"completedAt":      huntrFixtureCreated + int64(i),
			"date":             huntrFixtureCreated + int64(i),
			"completed":        false,
			"title":            fmt.Sprintf("Fixture activity %d", i),
			"activityCategory": "TASK",
			"actionType":       "MEMBER_DEACTIVATED",
			"memberId":         "mem_fixture_1",
			"candidateId":      "cand_fixture_1",
			"activityId":       "act_fixture_1",
			"text":             fmt.Sprintf("Fixture note %d", i),
			"htmlText":         fmt.Sprintf("<p>Fixture note %d</p>", i),
			"boardIds":         []any{"board_fixture_1"},
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
	base, err := huntrBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := huntrSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("huntr connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: huntrUserAgent,
	}, nil
}

func huntrSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// huntrBaseURL resolves and validates the base URL. The default is
// api.huntr.co/org; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func huntrBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return huntrDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("huntr config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("huntr config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("huntr config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func huntrPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return huntrDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("huntr config page_size must be an integer: %w", err)
	}
	if value < 1 || value > huntrMaxPageSize {
		return 0, fmt.Errorf("huntr config page_size must be between 1 and %d", huntrMaxPageSize)
	}
	return value, nil
}

func huntrMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("huntr config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("huntr config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: Huntr's organization API is read-only for this
// connector, so the Connector interface's Write returns ErrUnsupportedOperation.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
