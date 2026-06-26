// Package surveycto implements the native pm SurveyCTO connector.
package surveycto

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
	connectorName   = "surveycto"
	defaultPageSize = 100
	maxPageSize     = 1000
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

type streamEndpoint struct {
	resource    string
	recordsPath string
	needsForm   bool
	fields      []connectors.Field
	mapRecord   func(map[string]any) connectors.Record
}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "SurveyCTO", IntegrationType: "api", Description: "Reads SurveyCTO forms, submissions, datasets, and cases through the SurveyCTO API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, "forms", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check surveycto: %w", err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, _ connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: connectorName, Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "forms"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("surveycto stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, endpoint, emit)
	}
	resource, err := resolveResource(endpoint, req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := pageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := maxPages(req.Config)
	if err != nil {
		return err
	}
	paginator := &connsdk.OffsetPaginator{LimitParam: "limit", OffsetParam: "offset", PageSize: pageSize}
	return connsdk.Harvest(ctx, r, http.MethodGet, resource, nil, paginator, endpoint.recordsPath, maxPages, func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	})
}

func (Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	username := secret(cfg, "username")
	password := secret(cfg, "password")
	if strings.TrimSpace(username) == "" || strings.TrimSpace(password) == "" {
		return nil, errors.New("surveycto connector requires secrets username and password")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Basic(username, password), UserAgent: userAgent}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		server := strings.TrimSpace(cfg.Config["server_name"])
		if server == "" {
			return "", errors.New("surveycto connector requires config base_url or server_name")
		}
		if strings.ContainsAny(server, "/?#@:") {
			return "", errors.New("surveycto config server_name must be a bare SurveyCTO server name")
		}
		base = "https://" + server + ".surveycto.com/api/v2"
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("surveycto config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("surveycto config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("surveycto config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func resolveResource(endpoint streamEndpoint, cfg connectors.RuntimeConfig) (string, error) {
	if !endpoint.needsForm {
		return endpoint.resource, nil
	}
	formID := strings.TrimSpace(cfg.Config["form_id"])
	if formID == "" {
		return "", fmt.Errorf("surveycto stream requires config form_id for path %q", endpoint.resource)
	}
	return strings.ReplaceAll(endpoint.resource, "{form_id}", url.PathEscape(formID)), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return boundedInt(cfg.Config["page_size"], defaultPageSize, maxPageSize, "surveycto config page_size")
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("surveycto config max_pages must be a non-negative integer: %w", err)
	}
	return value, nil
}

func boundedInt(raw string, def, max int, name string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, err)
	}
	if value < 1 || value > max {
		return 0, fmt.Errorf("%s must be between 1 and %d", name, max)
	}
	return value, nil
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "title": fmt.Sprintf("Fixture %s %d", stream, i), "version": strconv.Itoa(i), "form_id": "fixture_form", "submissionDate": "2026-01-01T00:00:00Z", "caseid": fmt.Sprintf("case-%d", i)}
		rec := endpoint.mapRecord(item)
		rec["connector"] = connectorName
		rec["fixture"] = true
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "forms", Description: "SurveyCTO forms.", PrimaryKey: []string{"id"}, Fields: streamEndpoints["forms"].fields},
		{Name: "submissions", Description: "SurveyCTO submissions for a form_id.", PrimaryKey: []string{"id"}, CursorFields: []string{"submissionDate"}, Fields: streamEndpoints["submissions"].fields},
		{Name: "datasets", Description: "SurveyCTO datasets.", PrimaryKey: []string{"id"}, Fields: streamEndpoints["datasets"].fields},
		{Name: "cases", Description: "SurveyCTO cases.", PrimaryKey: []string{"caseid"}, Fields: streamEndpoints["cases"].fields},
	}
}

var streamEndpoints = map[string]streamEndpoint{
	"forms":       {resource: "forms", recordsPath: "forms", fields: formFields(), mapRecord: copyRecord("id", "title", "version")},
	"submissions": {resource: "forms/{form_id}/submissions", recordsPath: "submissions", needsForm: true, fields: submissionFields(), mapRecord: copyRecord("id", "form_id", "submissionDate")},
	"datasets":    {resource: "datasets", recordsPath: "datasets", fields: formFields(), mapRecord: copyRecord("id", "title", "version")},
	"cases":       {resource: "cases", recordsPath: "cases", fields: caseFields(), mapRecord: copyRecord("caseid", "form_id", "status")},
}

func formFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "string"}, {Name: "title", Type: "string"}, {Name: "version", Type: "string"}}
}

func submissionFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "string"}, {Name: "form_id", Type: "string"}, {Name: "submissionDate", Type: "timestamp"}}
}

func caseFields() []connectors.Field {
	return []connectors.Field{{Name: "caseid", Type: "string"}, {Name: "form_id", Type: "string"}, {Name: "status", Type: "string"}}
}

func copyRecord(keys ...string) func(map[string]any) connectors.Record {
	return func(item map[string]any) connectors.Record {
		rec := connectors.Record{}
		for _, key := range keys {
			rec[key] = item[key]
		}
		return rec
	}
}
