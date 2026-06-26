// Package eventbrite implements the native pm Eventbrite connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference: a thin package that composes the connsdk toolkit (Requester +
// Bearer auth + RecordsAt extraction + cursor state) with Eventbrite-specific
// stream definitions, endpoints, and record mappers.
//
// Eventbrite is a read-only data source (no reverse-ETL writes), so it exposes
// Check/Catalog/Read but not Write. Like the other per-system connectors it
// self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
package eventbrite

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
	eventbriteDefaultBaseURL = "https://www.eventbriteapi.com/v3"
	eventbriteUserAgent      = "polymetrics-go-cli"
	// eventbriteFixtureChanged is the deterministic `changed` timestamp used by
	// the fixture-mode records.
	eventbriteFixtureChanged = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("eventbrite", New)
}

// New returns the Eventbrite connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Eventbrite connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "eventbrite" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "eventbrite",
		DisplayName:     "Eventbrite",
		IntegrationType: "api",
		Description:     "Reads Eventbrite organizations, events, attendees, orders, and ticket classes through the Eventbrite v3 REST API. Read-only source.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Eventbrite.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := eventbriteBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(eventbriteSecret(cfg)) == "" {
		return errors.New("eventbrite connector requires secret private_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the authenticated user's organizations confirms auth and
	// connectivity without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "users/me/organizations/", nil, nil, nil); err != nil {
		return fmt.Errorf("check eventbrite: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: eventbriteStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: an Eventbrite stream starts
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
		stream = "organizations"
	}
	endpoint, ok := eventbriteStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("eventbrite stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	path, err := resolvePath(endpoint, req.Config)
	if err != nil {
		return err
	}
	maxPages, err := eventbriteMaxPages(req.Config)
	if err != nil {
		return err
	}
	lower := incrementalLowerBound(req)
	return c.harvest(ctx, r, path, endpoint, maxPages, lower, emit)
}

// Write is unsupported: Eventbrite is a read-only source.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// harvest drives Eventbrite's continuation-token pagination. List responses
// carry {pagination:{has_more_items, continuation}, <key>:[...]}; the next page
// is requested with continuation=<token>. There is no exact connsdk paginator
// for the has_more_items+continuation gate, so the loop lives here, built on
// connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, endpoint streamEndpoint, maxPages int, changedSince string, emit func(connectors.Record) error) error {
	base := url.Values{}
	// expand attendee/order profiles so flattened name/email fields are
	// populated; harmless on streams that ignore it.
	base.Set("expand", "venue,ticket_classes")
	if changedSince != "" {
		base.Set("changed_since", changedSince)
	}

	continuation := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if continuation != "" {
			query.Set("continuation", continuation)
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read eventbrite %s: %w", endpoint.recordsKey, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsKey)
		if err != nil {
			return fmt.Errorf("decode eventbrite %s page: %w", endpoint.recordsKey, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		hasMore, err := connsdk.StringAt(resp.Body, "pagination.has_more_items")
		if err != nil {
			return fmt.Errorf("decode eventbrite %s pagination: %w", endpoint.recordsKey, err)
		}
		if hasMore != "true" {
			return nil
		}
		next, err := connsdk.StringAt(resp.Body, "pagination.continuation")
		if err != nil {
			return fmt.Errorf("decode eventbrite %s continuation: %w", endpoint.recordsKey, err)
		}
		if strings.TrimSpace(next) == "" || next == continuation {
			return nil
		}
		continuation = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise eventbrite credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":              map[string]any{"text": fmt.Sprintf("Fixture %s %d", stream, i)},
			"description":       map[string]any{"text": "Fixture record"},
			"changed":           eventbriteFixtureChanged,
			"created":           eventbriteFixtureChanged,
			"status":            "live",
			"vertical":          "default",
			"locale":            "en_US",
			"url":               "https://www.eventbrite.com/e/fixture",
			"currency":          "USD",
			"online_event":      false,
			"listed":            true,
			"organization_id":   "org_fixture_1",
			"venue_id":          "venue_fixture_1",
			"capacity":          int64(100),
			"event_id":          "events_fixture_1",
			"order_id":          "orders_fixture_1",
			"ticket_class_id":   "ticket_classes_fixture_1",
			"ticket_class_name": "General Admission",
			"quantity":          int64(i),
			"checked_in":        false,
			"cancelled":         false,
			"refunded":          false,
			"time_remaining":    nil,
			"profile":           map[string]any{"name": fmt.Sprintf("Fixture %d", i), "email": fmt.Sprintf("fixture+%d@example.com", i)},
			"cost":              map[string]any{"display": "$10.00"},
			"fee":               map[string]any{"display": "$1.00"},
			"quantity_total":    int64(100),
			"quantity_sold":     int64(i),
			"free":              false,
			"hidden":            false,
			"on_sale_status":    "AVAILABLE",
			"start":             map[string]any{"utc": eventbriteFixtureChanged},
			"end":               map[string]any{"utc": eventbriteFixtureChanged},
			"published":         eventbriteFixtureChanged,
		}
		record := endpoint.mapRecord(item)
		if cursor := connsdk.Cursor(req.State); cursor != "" {
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
	base, err := eventbriteBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := eventbriteSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("eventbrite connector requires secret private_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: eventbriteUserAgent,
	}, nil
}

// resolvePath builds the concrete endpoint path for a stream, splicing in the
// configured organization_id or event_id as the scope requires.
func resolvePath(endpoint streamEndpoint, cfg connectors.RuntimeConfig) (string, error) {
	switch endpoint.scope {
	case scopeUser:
		return endpoint.pathTemplate, nil
	case scopeOrg:
		id := strings.TrimSpace(cfg.Config["organization_id"])
		if id == "" {
			return "", errors.New("eventbrite stream requires config organization_id")
		}
		return fmt.Sprintf(endpoint.pathTemplate, url.PathEscape(id)), nil
	case scopeEvent:
		id := strings.TrimSpace(cfg.Config["event_id"])
		if id == "" {
			return "", errors.New("eventbrite stream requires config event_id")
		}
		return fmt.Sprintf(endpoint.pathTemplate, url.PathEscape(id)), nil
	default:
		return "", fmt.Errorf("eventbrite: unknown stream scope %d", endpoint.scope)
	}
}

// incrementalLowerBound returns the RFC3339 changed_since lower bound, derived
// from the incremental cursor (if any) or else the start_date config. An empty
// result means no lower bound (full sync).
func incrementalLowerBound(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	return strings.TrimSpace(req.Config.Config["start_date"])
}

func eventbriteSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["private_token"]
}

// eventbriteBaseURL resolves and validates the base URL. The default is
// www.eventbriteapi.com/v3; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func eventbriteBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return eventbriteDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("eventbrite config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("eventbrite config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("eventbrite config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func eventbriteMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("eventbrite config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("eventbrite config max_pages must be 0 for unlimited or a positive integer")
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
