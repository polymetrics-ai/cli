// Package googletasks implements the native pm Google Tasks connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit: a
// connsdk.Requester with Bearer auth (the api_key secret) reads the Google Tasks
// REST API (https://tasks.googleapis.com/tasks/v1), paginating with
// nextPageToken/pageToken and extracting the "items" array per resource.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect. The API is read-only for our purposes, so the connector
// exposes Check/Catalog/Read and sets Capabilities.Write=false.
package googletasks

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
	registryName              = "google-tasks"
	googleTasksDefaultBaseURL = "https://tasks.googleapis.com/tasks/v1"
	googleTasksDefaultLimit   = 50
	googleTasksMaxLimit       = 100
	googleTasksUserAgent      = "polymetrics-go-cli"
	// googleTasksFixtureUpdated is the deterministic `updated` timestamp used by
	// the fixture-mode records.
	googleTasksFixtureUpdated = "2026-01-01T00:00:00.000Z"
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the Google Tasks connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Google Tasks connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "Google Tasks",
		IntegrationType: "api",
		Description:     "Reads Google task lists and tasks through the Google Tasks REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true},
	}
}

// Check verifies the connector is configured well enough to talk to the Google
// Tasks API. In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := googleTasksBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(googleTasksSecret(cfg)) == "" {
		return errors.New("google-tasks connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the task lists confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "users/@me/lists", url.Values{"maxResults": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check google-tasks: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: googleTasksStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Google Tasks stream starts
// with an empty incremental cursor (full sync). The Google Tasks source only
// supports full_refresh, so the cursor is informational.
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
		stream = "tasklists"
	}
	endpoint, ok := googleTasksStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("google-tasks stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	limit, err := googleTasksLimit(req.Config)
	if err != nil {
		return err
	}

	if endpoint.nested {
		return c.readTasks(ctx, r, endpoint, limit, emit)
	}
	return c.harvest(ctx, r, endpoint.resource, limit, endpoint.mapRecord, emit)
}

// harvest reads every page of a single Google Tasks list endpoint. Pages carry a
// nextPageToken in the body which is supplied as pageToken on the next request;
// records live under "items".
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, resource string, limit int, mapRecord func(map[string]any) connectors.Record, emit func(connectors.Record) error) error {
	paginator := &connsdk.CursorPaginator{
		CursorParam: "pageToken",
		TokenPath:   "nextPageToken",
		FirstQuery:  url.Values{"maxResults": []string{strconv.Itoa(limit)}},
	}
	return connsdk.Harvest(ctx, r, http.MethodGet, resource, nil, paginator, "items", 0, func(rec connsdk.Record) error {
		return emit(mapRecord(map[string]any(rec)))
	})
}

// readTasks fans out across every task list, harvesting that list's tasks. Each
// emitted task is tagged with its parent tasklist_id so downstream consumers can
// relate it back to its list.
func (c Connector) readTasks(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, limit int, emit func(connectors.Record) error) error {
	listIDs, err := c.collectTaskListIDs(ctx, r, limit)
	if err != nil {
		return err
	}
	for _, listID := range listIDs {
		if err := ctx.Err(); err != nil {
			return err
		}
		resource := "lists/" + url.PathEscape(listID) + "/tasks"
		if err := c.harvest(ctx, r, resource, limit, func(item map[string]any) connectors.Record {
			rec := endpoint.mapRecord(item)
			rec["tasklist_id"] = listID
			return rec
		}, emit); err != nil {
			return fmt.Errorf("read google-tasks tasks for list %s: %w", listID, err)
		}
	}
	return nil
}

// collectTaskListIDs reads every task list and returns its ids, used to fan out
// the nested tasks stream.
func (c Connector) collectTaskListIDs(ctx context.Context, r *connsdk.Requester, limit int) ([]string, error) {
	var ids []string
	paginator := &connsdk.CursorPaginator{
		CursorParam: "pageToken",
		TokenPath:   "nextPageToken",
		FirstQuery:  url.Values{"maxResults": []string{strconv.Itoa(limit)}},
	}
	err := connsdk.Harvest(ctx, r, http.MethodGet, "users/@me/lists", nil, paginator, "items", 0, func(rec connsdk.Record) error {
		if id := stringField(map[string]any(rec), "id"); id != "" {
			ids = append(ids, id)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list google-tasks task lists: %w", err)
	}
	return ids, nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise google-tasks credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var record connectors.Record
		switch stream {
		case "tasks":
			record = taskRecord(map[string]any{
				"kind":      "tasks#task",
				"id":        fmt.Sprintf("task_fixture_%d", i),
				"etag":      fmt.Sprintf("etag_%d", i),
				"title":     fmt.Sprintf("Fixture task %d", i),
				"updated":   googleTasksFixtureUpdated,
				"selfLink":  fmt.Sprintf("https://tasks.googleapis.com/tasks/v1/lists/list_fixture_1/tasks/task_fixture_%d", i),
				"parent":    "",
				"position":  fmt.Sprintf("%020d", i),
				"notes":     "fixture notes",
				"status":    "needsAction",
				"due":       "",
				"completed": "",
				"deleted":   false,
				"hidden":    false,
			})
			record["tasklist_id"] = "list_fixture_1"
		default:
			record = taskListRecord(map[string]any{
				"kind":     "tasks#taskList",
				"id":       fmt.Sprintf("list_fixture_%d", i),
				"etag":     fmt.Sprintf("etag_%d", i),
				"title":    fmt.Sprintf("Fixture list %d", i),
				"updated":  googleTasksFixtureUpdated,
				"selfLink": fmt.Sprintf("https://tasks.googleapis.com/tasks/v1/users/@me/lists/list_fixture_%d", i),
			})
		}
		record["connector"] = registryName
		record["fixture"] = true
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := googleTasksBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := googleTasksSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("google-tasks connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: googleTasksUserAgent,
	}, nil
}

func googleTasksSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// googleTasksBaseURL resolves and validates the base URL. The default is
// tasks.googleapis.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func googleTasksBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return googleTasksDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("google-tasks config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("google-tasks config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("google-tasks config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

// googleTasksLimit resolves the per-request page size from the records_limit
// config (upstream's field name), bounded by the API maximum.
func googleTasksLimit(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["records_limit"])
	if raw == "" {
		return googleTasksDefaultLimit, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("google-tasks config records_limit must be an integer: %w", err)
	}
	if value < 1 || value > googleTasksMaxLimit {
		return 0, fmt.Errorf("google-tasks config records_limit must be between 1 and %d", googleTasksMaxLimit)
	}
	return value, nil
}

// Write is unsupported: the Google Tasks connector is read-only. The method
// exists to satisfy the connectors.Connector interface.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
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
