// Package dixa implements the native pm Dixa connector. It follows the
// declarative-HTTP template established by the stripe connector: a thin package
// that composes the connsdk toolkit (Requester + Bearer auth + RecordsAt
// extraction + cursor state) with Dixa-specific stream definitions and the
// conversation_export endpoint.
//
// Dixa exposes a single public export endpoint, conversation_export, which
// returns a top-level JSON array of rich conversation objects. There is no
// per-request paginator; instead the export is walked by advancing a
// [updated_after, updated_before) date window by batch_size days, mirroring the
// upstream Airbyte DatetimeBasedCursor. The connector projects that one payload
// into a few focused, well-keyed streams (conversations, queue, rating,
// assignment).
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package dixa

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	dixaDefaultBaseURL   = "https://exports.dixa.io/v1"
	dixaDefaultBatchDays = 31
	dixaMaxBatchDays     = 31
	dixaUserAgent        = "polymetrics-go-cli"
	// dixaMaxWindows bounds the number of date windows walked in a single read so
	// a wide start_date..end_date range cannot loop unbounded.
	dixaMaxWindows = 4096
	// dixaFixtureUpdated is the deterministic millisecond updated_at used by the
	// fixture-mode records (2026-01-01T00:00:00Z).
	dixaFixtureUpdated int64 = 1767225600000
)

func init() {
	connectors.RegisterFactory("dixa", New)
}

// New returns the Dixa connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Dixa connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "dixa" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "dixa",
		DisplayName:     "Dixa",
		IntegrationType: "api",
		Description:     "Reads Dixa conversations (and their queue, rating, and assignment projections) from the Dixa conversation_export API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Dixa. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := dixaBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(dixaSecret(cfg)) == "" {
		return errors.New("dixa connector requires secret api_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded one-day window confirms auth and connectivity without exporting
	// the whole history. The export endpoint requires both window bounds.
	end := time.Now().UTC().Truncate(24 * time.Hour)
	start := end.Add(-24 * time.Hour)
	query := url.Values{}
	query.Set("updated_after", millis(start))
	query.Set("updated_before", millis(end))
	if err := r.DoJSON(ctx, http.MethodGet, "conversation_export", query, nil, nil); err != nil {
		return fmt.Errorf("check dixa: %w", err)
	}
	return nil
}

// Write is unsupported: Dixa's public API surface here is a read-only export, so
// the connector advertises Write=false and rejects writes.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: dixaStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Dixa stream starts with an
// empty incremental cursor (the start_date config supplies the lower bound at
// read time).
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
		stream = "conversations"
	}
	endpoint, ok := dixaStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("dixa stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	start, end, err := dixaWindowBounds(req)
	if err != nil {
		return err
	}
	batch, err := dixaBatchDays(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, start, end, batch, emit)
}

// harvest walks the conversation_export endpoint one date window at a time. Dixa
// has no per-request paginator; each request returns a top-level JSON array for
// the [updated_after, updated_before) window. Windows advance by batch days until
// the end bound is reached. Records are deduplicated by id across windows so a
// conversation updated near a window boundary is emitted once.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, start, end time.Time, batchDays int, emit func(connectors.Record) error) error {
	if !start.Before(end) {
		return nil
	}
	step := time.Duration(batchDays) * 24 * time.Hour
	seen := map[string]bool{}
	windowStart := start
	for window := 0; windowStart.Before(end); window++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if window >= dixaMaxWindows {
			return fmt.Errorf("dixa export exceeded %d date windows; narrow start_date/end_date or raise batch_size", dixaMaxWindows)
		}
		windowEnd := windowStart.Add(step)
		if windowEnd.After(end) {
			windowEnd = end
		}
		query := url.Values{}
		query.Set("updated_after", millis(windowStart))
		query.Set("updated_before", millis(windowEnd))
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read dixa %s: %w", endpoint.resource, err)
		}
		// conversation_export returns a top-level array (field_path: []).
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode dixa %s window: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if id := stringField(item, "id"); id != "" {
				if seen[id] {
					continue
				}
				seen[id] = true
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		windowStart = windowEnd
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise dixa credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                      int64(i),
			"created_at":              dixaFixtureUpdated - 60000,
			"updated_at":              dixaFixtureUpdated + int64(i),
			"closed_at":               dixaFixtureUpdated + 120000,
			"status":                  "Closed",
			"direction":               "Inbound",
			"initial_channel":         "Email",
			"subject":                 fmt.Sprintf("Fixture conversation %d", i),
			"requester_id":            fmt.Sprintf("usr_fixture_%d", i),
			"requester_name":          fmt.Sprintf("Fixture User %d", i),
			"requester_email":         fmt.Sprintf("fixture+%d@example.com", i),
			"total_duration":          int64(600 * i),
			"handling_duration":       int64(120 * i),
			"last_message_created_at": dixaFixtureUpdated,
			"originating_country":     "US",
			"queue_id":                "queue_fixture_1",
			"queue_name":              "Support",
			"queued_at":               dixaFixtureUpdated - 30000,
			"rating_score":            int64(5),
			"rating_message":          "Great",
			"assigned_at":             dixaFixtureUpdated - 10000,
			"assignee_id":             "agt_fixture_1",
			"assignee_name":           "Fixture Agent",
			"assignee_email":          "agent@example.com",
			"transferee_name":         "",
			"transfer_time":           int64(0),
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
	base, err := dixaBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := dixaSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("dixa connector requires secret api_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: dixaUserAgent,
	}, nil
}

// dixaWindowBounds resolves the [start, end) export window. The lower bound comes
// from the incremental cursor (a millisecond updated_at) if present, otherwise
// from the start_date config; the upper bound is the end_date config or now.
func dixaWindowBounds(req connectors.ReadRequest) (time.Time, time.Time, error) {
	var start time.Time
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		ms, err := strconv.ParseInt(cursor, 10, 64)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("dixa cursor must be millisecond updated_at: %w", err)
		}
		start = time.UnixMilli(ms).UTC()
	} else {
		raw := strings.TrimSpace(req.Config.Config["start_date"])
		if raw == "" {
			return time.Time{}, time.Time{}, errors.New("dixa connector requires config start_date")
		}
		t, err := parseDixaDate(raw)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("dixa config start_date: %w", err)
		}
		start = t
	}

	end := time.Now().UTC()
	if raw := strings.TrimSpace(req.Config.Config["end_date"]); raw != "" {
		t, err := parseDixaDate(raw)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("dixa config end_date: %w", err)
		}
		end = t
	}
	return start, end, nil
}

// parseDixaDate accepts both a plain date (YYYY-MM-DD) and a full RFC3339
// timestamp, normalising to UTC. Dixa's spec documents YYYY-MM-DD but
// format=date-time, so both are tolerated.
func parseDixaDate(raw string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("must be YYYY-MM-DD or RFC3339, got %q", raw)
}

func dixaSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_token"]
}

// dixaBaseURL resolves and validates the base URL. The default is
// exports.dixa.io; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func dixaBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return dixaDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("dixa config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("dixa config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("dixa config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

// dixaBatchDays resolves the date-window step in days, clamped to the documented
// maximum of 31.
func dixaBatchDays(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["batch_size"])
	if raw == "" {
		return dixaDefaultBatchDays, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("dixa config batch_size must be an integer: %w", err)
	}
	if value < 1 || value > dixaMaxBatchDays {
		return 0, fmt.Errorf("dixa config batch_size must be between 1 and %d", dixaMaxBatchDays)
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// millis renders a time as Dixa's millisecond unix timestamp string.
func millis(t time.Time) string {
	return strconv.FormatInt(t.UTC().UnixMilli(), 10)
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
