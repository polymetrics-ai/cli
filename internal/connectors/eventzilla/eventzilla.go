// Package eventzilla implements the native pm Eventzilla connector. It follows
// the declarative-HTTP template established by the stripe package: a thin
// package that composes the connsdk toolkit (Requester + x-api-key header auth +
// field-path record extraction + offset/limit pagination) with Eventzilla
// stream definitions and endpoints.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// Eventzilla exposes only full-refresh reads, so the connector is read-only.
package eventzilla

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
	eventzillaDefaultBaseURL  = "https://www.eventzillaapi.net/api/v2"
	eventzillaDefaultPageSize = 100
	eventzillaMaxPageSize     = 100
	eventzillaUserAgent       = "polymetrics-go-cli"
	// eventzillaAPIKeyHeader is the header Eventzilla expects the API key in.
	eventzillaAPIKeyHeader = "x-api-key"
)

func init() {
	connectors.RegisterFactory("eventzilla", New)
}

// New returns the Eventzilla connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Eventzilla connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "eventzilla" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "eventzilla",
		DisplayName:     "Eventzilla",
		IntegrationType: "api",
		Description:     "Reads Eventzilla events, categories, users, attendees, and ticket types through the Eventzilla v2 REST API (read-only; full refresh).",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Eventzilla.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := eventzillaBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(eventzillaSecret(cfg)) == "" {
		return errors.New("eventzilla connector requires secret x-api-key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the events list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "events", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check eventzilla: %w", err)
	}
	return nil
}

// Write is unsupported: Eventzilla is a read-only source connector (full
// refresh only). It satisfies the connectors.Connector interface but always
// returns ErrUnsupportedOperation.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: eventzillaStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "events"
	}
	endpoint, ok := eventzillaStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("eventzilla stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := eventzillaPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := eventzillaMaxPages(req.Config)
	if err != nil {
		return err
	}

	if endpoint.parentScoped {
		return c.readSubstream(ctx, r, endpoint, pageSize, maxPages, emit)
	}
	return c.harvest(ctx, r, endpoint.resource, endpoint.fieldPath, pageSize, maxPages, endpoint.mapRecord, emit)
}

// readSubstream fans a parent-scoped child stream (attendees, tickets) out over
// every event id. It first lists events, then reads
// /events/{event_id}/<child> for each, stamping the parent event_id onto each
// emitted record so the child rows remain joinable.
func (c Connector) readSubstream(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	var eventIDs []string
	collect := func(rec connectors.Record) error {
		id := stringField(rec, "id")
		if id != "" {
			eventIDs = append(eventIDs, id)
		}
		return nil
	}
	if err := c.harvest(ctx, r, "events", "events", pageSize, maxPages, eventRecord, collect); err != nil {
		return fmt.Errorf("eventzilla list events for %s: %w", endpoint.resource, err)
	}

	for _, eventID := range eventIDs {
		if err := ctx.Err(); err != nil {
			return err
		}
		path := "events/" + url.PathEscape(eventID) + "/" + endpoint.resource
		mapper := stampParent(endpoint.mapRecord, eventID)
		if err := c.harvest(ctx, r, path, endpoint.fieldPath, pageSize, maxPages, mapper, emit); err != nil {
			return fmt.Errorf("eventzilla read %s for event %s: %w", endpoint.resource, eventID, err)
		}
	}
	return nil
}

// harvest drives Eventzilla's offset/limit pagination over a single endpoint.
// Eventzilla returns {<field>:[...]} with no total/has_more marker, so the loop
// stops when a page returns fewer than pageSize records (OffsetIncrement
// semantics). The loop lives here, built on connsdk.Requester + RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path, fieldPath string, pageSize, maxPages int, mapRecord func(map[string]any) connectors.Record, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(offset))

		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read eventzilla %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, fieldPath)
		if err != nil {
			return fmt.Errorf("decode eventzilla %s page: %w", path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
		offset += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise eventzilla credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                 int64(i),
			"event_id":           int64(900),
			"title":              fmt.Sprintf("%s fixture %d", endpoint.resource, i),
			"status":             "live",
			"url":                fmt.Sprintf("https://events.example/%s/%d", endpoint.resource, i),
			"venue":              "Fixture Hall",
			"currency":           "USD",
			"start_date":         "2026-01-01",
			"start_time":         "09:00",
			"end_date":           "2026-01-01",
			"end_time":           "17:00",
			"time_zone":          "UTC",
			"categories":         "Conference",
			"tickets_sold":       int64(10 * i),
			"tickets_total":      int64(100),
			"category":           fmt.Sprintf("Category %d", i),
			"username":           fmt.Sprintf("user%d", i),
			"first_name":         fmt.Sprintf("First%d", i),
			"last_name":          fmt.Sprintf("Last%d", i),
			"email":              fmt.Sprintf("fixture+%d@example.com", i),
			"company":            "Fixture Co",
			"user_type":          "organizer",
			"phone_primary":      "+10000000000",
			"timezone":           "UTC",
			"last_seen":          "2026-01-01T00:00:00Z",
			"ticket_type":        "general",
			"refno":              fmt.Sprintf("REF-%d", i),
			"transaction_amount": float64(25 * i),
			"transaction_status": "complete",
			"transaction_date":   "2026-01-01",
			"is_attended":        "false",
			"price":              float64(25 * i),
			"quantity_total":     int64(100),
			"is_visible":         true,
			"sales_start_date":   "2025-12-01",
			"sales_end_date":     "2026-01-01",
		}
		record := endpoint.mapRecord(item)
		if endpoint.parentScoped {
			record["event_id"] = int64(900)
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with x-api-key header auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := eventzillaBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := eventzillaSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("eventzilla connector requires secret x-api-key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(eventzillaAPIKeyHeader, secret, ""),
		UserAgent: eventzillaUserAgent,
	}, nil
}

func eventzillaSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["x-api-key"]
}

// eventzillaBaseURL resolves and validates the base URL. The default is
// www.eventzillaapi.net; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func eventzillaBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return eventzillaDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("eventzilla config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("eventzilla config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("eventzilla config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func eventzillaPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return eventzillaDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("eventzilla config page_size must be an integer: %w", err)
	}
	if value < 1 || value > eventzillaMaxPageSize {
		return 0, fmt.Errorf("eventzilla config page_size must be between 1 and %d", eventzillaMaxPageSize)
	}
	return value, nil
}

func eventzillaMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("eventzilla config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("eventzilla config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// stampParent wraps a child mapper so each emitted record carries its parent
// event_id, keeping child rows joinable even when the API omits the field.
func stampParent(mapRecord func(map[string]any) connectors.Record, eventID string) func(map[string]any) connectors.Record {
	return func(item map[string]any) connectors.Record {
		record := mapRecord(item)
		if stringField(record, "event_id") == "" {
			record["event_id"] = eventID
		}
		return record
	}
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
