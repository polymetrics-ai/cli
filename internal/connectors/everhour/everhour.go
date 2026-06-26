// Package everhour implements the native pm Everhour connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference: a thin package that composes the connsdk toolkit (Requester +
// APIKeyHeader auth + RecordsAt extraction) with Everhour-specific stream
// definitions and endpoints.
//
// The Everhour API (https://api.everhour.com) authenticates with an X-Api-Key
// header and returns top-level JSON arrays. It is full-refresh only, so the
// connector is read-only with no incremental cursor. Tasks are read as a
// substream of projects (one /projects/<id>/tasks call per project).
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package everhour

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	everhourDefaultBaseURL = "https://api.everhour.com"
	everhourUserAgent      = "polymetrics-go-cli"
	everhourAPIKeyHeader   = "X-Api-Key"
)

func init() {
	connectors.RegisterFactory("everhour", New)
}

// New returns the Everhour connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Everhour connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "everhour" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "everhour",
		DisplayName:     "Everhour",
		IntegrationType: "api",
		Description:     "Reads Everhour projects, clients, team members, tasks, and time records through the Everhour REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Everhour. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := everhourBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(everhourSecret(cfg)) == "" {
		return errors.New("everhour connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// Listing the team members confirms auth and connectivity without mutating
	// anything (mirrors the Airbyte check stream).
	if err := r.DoJSON(ctx, http.MethodGet, "team/users", nil, nil, nil); err != nil {
		return fmt.Errorf("check everhour: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: everhourStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "projects"
	}
	endpoint, ok := everhourStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("everhour stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	if endpoint.parentResource != "" {
		return c.readSubstream(ctx, r, endpoint, emit)
	}
	return c.readTopLevel(ctx, r, endpoint.resource, endpoint.mapRecord, emit)
}

// readTopLevel reads a single Everhour array endpoint and emits each element.
func (c Connector) readTopLevel(ctx context.Context, r *connsdk.Requester, resource string, mapRecord func(map[string]any) connectors.Record, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, resource, nil, nil)
	if err != nil {
		return fmt.Errorf("read everhour %s: %w", resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode everhour %s: %w", resource, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// readSubstream lists the parent resource, then for each parent id requests the
// templated child path. The parent id is stitched onto each child record so
// downstream joins keep the relationship. This is the multi-page read path: one
// HTTP request per parent.
func (c Connector) readSubstream(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	parentResp, err := r.Do(ctx, http.MethodGet, endpoint.parentResource, nil, nil)
	if err != nil {
		return fmt.Errorf("read everhour %s (parents): %w", endpoint.parentResource, err)
	}
	parents, err := connsdk.RecordsAt(parentResp.Body, "")
	if err != nil {
		return fmt.Errorf("decode everhour %s (parents): %w", endpoint.parentResource, err)
	}
	for _, parent := range parents {
		if err := ctx.Err(); err != nil {
			return err
		}
		parentID := stringField(parent, "id")
		if parentID == "" {
			continue
		}
		childPath := endpoint.childPathFor(parentID)
		resp, err := r.Do(ctx, http.MethodGet, childPath, nil, nil)
		if err != nil {
			return fmt.Errorf("read everhour %s: %w", childPath, err)
		}
		children, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode everhour %s: %w", childPath, err)
		}
		for _, child := range children {
			if err := ctx.Err(); err != nil {
				return err
			}
			record := endpoint.mapRecord(child)
			if endpoint.parentIDField != "" {
				record[endpoint.parentIDField] = parentID
			}
			if err := emit(record); err != nil {
				return err
			}
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise everhour credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":              fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":            fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"type":            "board",
			"platform":        "everhour",
			"status":          "open",
			"workspaceId":     "ws_fixture",
			"workspaceName":   "Fixture Workspace",
			"favorite":        false,
			"foreign":         false,
			"email":           fmt.Sprintf("fixture+%d@example.com", i),
			"headline":        "Engineer",
			"role":            "member",
			"capacity":        40,
			"isEmailVerified": true,
			"completed":       false,
			"url":             fmt.Sprintf("https://app.everhour.com/%s/%d", stream, i),
			"date":            "2026-01-01",
			"time":            int64(3600 * i),
			"user":            int64(100 + i),
			"createdAt":       "2026-01-01T00:00:00Z",
		}
		record := endpoint.mapRecord(item)
		if endpoint.parentIDField != "" {
			record[endpoint.parentIDField] = "projects_fixture_1"
		}
		record["fixture"] = true
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with X-Api-Key auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := everhourBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := everhourSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("everhour connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(everhourAPIKeyHeader, secret, ""),
		UserAgent: everhourUserAgent,
	}, nil
}

func everhourSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// everhourBaseURL resolves and validates the base URL. The default is
// api.everhour.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func everhourBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return everhourDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("everhour config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("everhour config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("everhour config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: Everhour is exposed read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
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
