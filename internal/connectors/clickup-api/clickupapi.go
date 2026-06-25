// Package clickupapi implements the native pm ClickUp connector. It follows the
// stripe declarative-HTTP template: a thin package that composes the connsdk
// toolkit (Requester + personal-token header auth + RecordsAt extraction) with
// ClickUp-specific stream definitions, endpoints, and page-based pagination.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
//
// ClickUp is read-only here: the v2 API's writes (creating tasks, etc.) are not
// a natural reverse-ETL target, so Capabilities.Write is false.
package clickupapi

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
	registryName          = "clickup-api"
	clickupDefaultBaseURL = "https://api.clickup.com/api/v2"
	clickupUserAgent      = "polymetrics-go-cli"
	clickupMaxPages       = 1000 // hard safety bound on tasks pagination
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the ClickUp connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm ClickUp connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "ClickUp",
		IntegrationType: "api",
		Description:     "Reads ClickUp workspaces (teams), spaces, folders, lists, and tasks through the ClickUp v2 REST API using a personal API token.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to ClickUp. In
// fixture mode it short-circuits without a network call. Otherwise it lists the
// authenticated user's workspaces, which confirms auth and connectivity without
// mutating anything.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := clickupBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(clickupToken(cfg)) == "" {
		return errors.New("clickup-api connector requires secret api_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, "team", nil, nil, nil); err != nil {
		return fmt.Errorf("check clickup-api: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: clickupStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "teams"
	}
	shape, ok := clickupStreamShapes[stream]
	if !ok {
		return fmt.Errorf("clickup-api stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, shape, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	path, err := resolveEndpoint(stream, req.Config)
	if err != nil {
		return err
	}
	if shape.paginated {
		return c.harvestPaged(ctx, r, path, shape, emit, req.Config)
	}
	return c.fetchOnce(ctx, r, path, shape, emit, req.Config)
}

// Write is unsupported: clickup-api is a read-only source connector. It exists
// to satisfy the connectors.Connector interface.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// fetchOnce reads a single (non-paginated) ClickUp list endpoint and emits each
// record. ClickUp's space/folder/list endpoints return the whole collection in
// one response.
func (c Connector) fetchOnce(ctx context.Context, r *connsdk.Requester, path string, shape streamShape, emit func(connectors.Record) error, cfg connectors.RuntimeConfig) error {
	query := url.Values{}
	if includeArchived(cfg) {
		query.Set("archived", "true")
	}
	resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		return fmt.Errorf("read clickup-api %s: %w", path, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, shape.recordsPath)
	if err != nil {
		return fmt.Errorf("decode clickup-api %s: %w", path, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(shape.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// harvestPaged drives ClickUp's page-based pagination for the tasks endpoint.
// Responses are {tasks:[...], last_page:bool}; pages are 0-indexed via ?page=N
// and the loop stops when last_page is true (or an empty page is returned).
func (c Connector) harvestPaged(ctx context.Context, r *connsdk.Requester, path string, shape streamShape, emit func(connectors.Record) error, cfg connectors.RuntimeConfig) error {
	base := url.Values{}
	if includeClosed(cfg) {
		base.Set("include_closed", "true")
	}
	if !includeArchived(cfg) {
		base.Set("archived", "false")
	} else {
		base.Set("archived", "true")
	}

	for page := 0; page < clickupMaxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("page", strconv.Itoa(page))
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read clickup-api %s page %d: %w", path, page, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, shape.recordsPath)
		if err != nil {
			return fmt.Errorf("decode clickup-api %s page %d: %w", path, page, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(shape.mapRecord(item)); err != nil {
				return err
			}
		}
		lastPage, err := connsdk.StringAt(resp.Body, "last_page")
		if err != nil {
			return fmt.Errorf("decode clickup-api %s last_page: %w", path, err)
		}
		if lastPage == "true" || len(records) == 0 {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise clickup-api credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, shape streamShape, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":           fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":         fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"color":        "#4287f5",
			"private":      false,
			"archived":     false,
			"hidden":       false,
			"orderindex":   i,
			"task_count":   i,
			"date_created": "1767225600000",
			"date_updated": fmt.Sprintf("17672256%02d000", i),
			"date_closed":  nil,
			"url":          fmt.Sprintf("https://app.clickup.com/t/%s_fixture_%d", stream, i),
			"status":       map[string]any{"status": "open"},
			"space":        map[string]any{"id": "space_fixture_1"},
			"list":         map[string]any{"id": "list_fixture_1"},
			"folder":       map[string]any{"id": "folder_fixture_1"},
			"creator":      map[string]any{"id": "user_fixture_1"},
		}
		record := shape.mapRecord(item)
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with personal-token header auth and
// the resolved base URL. ClickUp personal tokens are sent raw in the
// Authorization header with no "Bearer " prefix, so APIKeyHeader with an empty
// prefix is used. The secret only ever flows into the authenticator; it is never
// logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := clickupBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := clickupToken(cfg)
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("clickup-api connector requires secret api_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", token, ""),
		UserAgent: clickupUserAgent,
	}, nil
}

// resolveEndpoint maps a stream + config to its ClickUp v2 resource path. The
// teams stream is unparameterised; the others require a team_id (tasks) or
// space_id (spaces require team_id; folders/lists require space_id) supplied via
// config, mirroring the upstream connector's parent-stream scoping.
func resolveEndpoint(stream string, cfg connectors.RuntimeConfig) (string, error) {
	teamID := strings.TrimSpace(cfg.Config["team_id"])
	spaceID := strings.TrimSpace(cfg.Config["space_id"])
	switch stream {
	case "teams":
		return "team", nil
	case "spaces":
		if teamID == "" {
			return "", errors.New("clickup-api stream spaces requires config team_id")
		}
		return "team/" + url.PathEscape(teamID) + "/space", nil
	case "folders":
		if spaceID == "" {
			return "", errors.New("clickup-api stream folders requires config space_id")
		}
		return "space/" + url.PathEscape(spaceID) + "/folder", nil
	case "lists":
		if spaceID == "" {
			return "", errors.New("clickup-api stream lists requires config space_id")
		}
		return "space/" + url.PathEscape(spaceID) + "/list", nil
	case "tasks":
		if teamID == "" {
			return "", errors.New("clickup-api stream tasks requires config team_id")
		}
		return "team/" + url.PathEscape(teamID) + "/task", nil
	default:
		return "", fmt.Errorf("clickup-api stream %q has no endpoint", stream)
	}
}

func clickupToken(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_token"]
}

// clickupBaseURL resolves and validates the base URL. The default is
// api.clickup.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func clickupBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return clickupDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("clickup-api config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("clickup-api config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("clickup-api config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func includeClosed(cfg connectors.RuntimeConfig) bool {
	return boolConfig(cfg, "include_closed_tasks")
}

func includeArchived(cfg connectors.RuntimeConfig) bool {
	return boolConfig(cfg, "include_archived")
}

func boolConfig(cfg connectors.RuntimeConfig, key string) bool {
	if cfg.Config == nil {
		return false
	}
	v := strings.TrimSpace(strings.ToLower(cfg.Config[key]))
	return v == "true" || v == "1" || v == "yes"
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
