// Package granola implements the native pm Granola connector. It follows the
// stripe reference shape: a thin declarative-HTTP package that composes the
// connsdk toolkit (Requester + Bearer auth + RecordsAt extraction + cursor
// state) with Granola-specific stream definitions and endpoints.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// The Granola public API (https://public-api.granola.ai/v1) is authenticated
// with a Bearer API key (grn_ prefix), exposes a GET /notes list with cursor
// pagination ({notes:[...], hasMore, cursor}), and a GET /notes/{id} detail
// endpoint. Only notes that have a generated AI summary and transcript are
// returned. This connector is read-only.
package granola

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
	granolaDefaultBaseURL  = "https://public-api.granola.ai/v1"
	granolaDefaultPageSize = 30
	granolaMaxPageSize     = 30
	granolaUserAgent       = "polymetrics-go-cli"
	// granolaFixtureCreated is the deterministic created_at used by fixture-mode
	// records.
	granolaFixtureCreated = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("granola", New)
}

// New returns the Granola connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Granola connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "granola" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "granola",
		DisplayName:     "Granola",
		IntegrationType: "api",
		Description:     "Reads Granola meeting notes and full note detail (summaries, owners, attendees, calendar events) through the Granola public API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Granola. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := granolaBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(granolaSecret(cfg)) == "" {
		return errors.New("granola connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the notes list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "notes", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check granola: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: granolaStreams()}, nil
}

// Write is unsupported: the Granola connector is read-only. It satisfies the
// connectors.Connector interface while reporting the operation as unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: a Granola stream starts with
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
		stream = "notes"
	}
	if stream != "notes" && stream != "detailed_notes" {
		return fmt.Errorf("granola stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	createdAfter, err := incrementalLowerBound(req)
	if err != nil {
		return err
	}
	pageSize, err := granolaPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := granolaMaxPages(req.Config)
	if err != nil {
		return err
	}

	if stream == "detailed_notes" {
		return c.harvestDetailed(ctx, r, pageSize, maxPages, createdAfter, emit)
	}
	return c.harvest(ctx, r, pageSize, maxPages, createdAfter, granolaNoteRecord, emit)
}

// harvest drives Granola's cursor pagination over the notes list. List responses
// are {notes:[...], hasMore:bool, cursor:string}; the next page is requested with
// cursor=<cursor>. The loop is built on connsdk.Requester + connsdk.RecordsAt +
// connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, pageSize, maxPages int, createdAfter string, mapRecord func(map[string]any) connectors.Record, emit func(connectors.Record) error) error {
	for note := range c.notesSeq(ctx, r, pageSize, maxPages, createdAfter) {
		if note.err != nil {
			return note.err
		}
		if err := emit(mapRecord(note.item)); err != nil {
			return err
		}
	}
	return nil
}

// harvestDetailed fans out from the notes list to a per-note GET /notes/{id}
// detail fetch, emitting the richer detailed_notes record for each.
func (c Connector) harvestDetailed(ctx context.Context, r *connsdk.Requester, pageSize, maxPages int, createdAfter string, emit func(connectors.Record) error) error {
	for note := range c.notesSeq(ctx, r, pageSize, maxPages, createdAfter) {
		if note.err != nil {
			return note.err
		}
		id := stringField(note.item, "id")
		if id == "" {
			continue
		}
		var detail map[string]any
		if err := r.DoJSON(ctx, http.MethodGet, "notes/"+url.PathEscape(id), url.Values{"include": []string{"transcript"}}, nil, &detail); err != nil {
			return fmt.Errorf("read granola note %s: %w", id, err)
		}
		if err := emit(granolaDetailedNoteRecord(detail)); err != nil {
			return err
		}
	}
	return nil
}

// noteItem carries a single decoded note or a terminal error through the
// iterator sequence.
type noteItem struct {
	item map[string]any
	err  error
}

// notesSeq yields every note across all pages of GET /notes, walking the cursor
// pagination. Errors are delivered as a final element carrying err and stop the
// range loop.
func (c Connector) notesSeq(ctx context.Context, r *connsdk.Requester, pageSize, maxPages int, createdAfter string) func(yield func(noteItem) bool) {
	return func(yield func(noteItem) bool) {
		base := url.Values{}
		base.Set("limit", strconv.Itoa(pageSize))
		if createdAfter != "" {
			base.Set("created_after", createdAfter)
		}

		cursor := ""
		for page := 0; maxPages == 0 || page < maxPages; page++ {
			if err := ctx.Err(); err != nil {
				yield(noteItem{err: err})
				return
			}
			query := cloneValues(base)
			if cursor != "" {
				query.Set("cursor", cursor)
			}
			resp, err := r.Do(ctx, http.MethodGet, "notes", query, nil)
			if err != nil {
				yield(noteItem{err: fmt.Errorf("read granola notes: %w", err)})
				return
			}
			records, err := connsdk.RecordsAt(resp.Body, "notes")
			if err != nil {
				yield(noteItem{err: fmt.Errorf("decode granola notes page: %w", err)})
				return
			}
			for _, item := range records {
				if err := ctx.Err(); err != nil {
					yield(noteItem{err: err})
					return
				}
				if !yield(noteItem{item: item}) {
					return
				}
			}
			hasMore, err := connsdk.StringAt(resp.Body, "hasMore")
			if err != nil {
				yield(noteItem{err: fmt.Errorf("decode granola hasMore: %w", err)})
				return
			}
			next, err := connsdk.StringAt(resp.Body, "cursor")
			if err != nil {
				yield(noteItem{err: fmt.Errorf("decode granola cursor: %w", err)})
				return
			}
			if hasMore != "true" || strings.TrimSpace(next) == "" {
				return
			}
			cursor = next
		}
	}
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise granola credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":         fmt.Sprintf("not_fixture_%d", i),
			"title":      fmt.Sprintf("Fixture meeting %d", i),
			"object":     "note",
			"created_at": granolaFixtureCreated,
			"updated_at": granolaFixtureCreated,
			"owner": map[string]any{
				"name":  fmt.Sprintf("Fixture Owner %d", i),
				"email": fmt.Sprintf("owner+%d@example.com", i),
			},
			"summary":        fmt.Sprintf("Fixture summary %d", i),
			"transcript":     []any{map[string]any{"speaker": "A", "text": "hello"}},
			"attendees":      []any{map[string]any{"name": "Attendee", "email": "att@example.com"}},
			"calendar_event": map[string]any{"id": fmt.Sprintf("cal_%d", i), "title": "Calendar event"},
			"folders":        []any{},
		}
		var record connectors.Record
		if stream == "detailed_notes" {
			record = granolaDetailedNoteRecord(item)
		} else {
			record = granolaNoteRecord(item)
		}
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
	base, err := granolaBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := granolaSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("granola connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: granolaUserAgent,
	}, nil
}

// incrementalLowerBound returns the RFC3339 lower bound for created_after,
// derived from the incremental cursor (if any) or else the start_date config.
// An empty result means no lower bound (full sync). The catalog config supplies
// start_date as YYYY-MM-DD, which is normalized to RFC3339 here.
func incrementalLowerBound(req connectors.ReadRequest) (string, error) {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor, nil
	}
	startDate := strings.TrimSpace(req.Config.Config["start_date"])
	if startDate == "" {
		return "", nil
	}
	if t, err := time.Parse(time.RFC3339, startDate); err == nil {
		return t.UTC().Format(time.RFC3339), nil
	}
	t, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return "", fmt.Errorf("granola config start_date must be YYYY-MM-DD or RFC3339: %w", err)
	}
	return t.UTC().Format(time.RFC3339), nil
}

func granolaSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// granolaBaseURL resolves and validates the base URL. The default is
// public-api.granola.ai; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func granolaBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return granolaDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("granola config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("granola config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("granola config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func granolaPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return granolaDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("granola config page_size must be an integer: %w", err)
	}
	if value < 1 || value > granolaMaxPageSize {
		return 0, fmt.Errorf("granola config page_size must be between 1 and %d", granolaMaxPageSize)
	}
	return value, nil
}

func granolaMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("granola config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("granola config max_pages must be 0 for unlimited or a positive integer")
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
