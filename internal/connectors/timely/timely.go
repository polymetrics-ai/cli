// Package timely implements a read-only native Go connector for Timely.
package timely

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
	connectorName  = "timely"
	defaultBaseURL = "https://api.timelyapp.com/1.1"
	userAgent      = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Timely", IntegrationType: "api", Description: "Reads users, projects, clients, and events from the Timely API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	path, err := accountPath(cfg, "users")
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, path, nil, nil, nil); err != nil {
		return fmt.Errorf("check %s: %w", connectorName, err)
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
		stream = "users"
	}
	spec, ok := streamSpecs[stream]
	if !ok {
		return fmt.Errorf("%s stream %q not found", connectorName, stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, spec, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	path, err := accountPath(req.Config, spec.path)
	if err != nil {
		return err
	}
	q := url.Values{}
	if spec.name == "events" {
		if start := strings.TrimSpace(req.Config.Config["start_date"]); start != "" {
			q.Set("since", start)
		}
	}
	resp, err := r.Do(ctx, http.MethodGet, path, q, nil)
	if err != nil {
		return fmt.Errorf("read %s %s: %w", connectorName, stream, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, spec.recordsPath)
	if err != nil {
		return err
	}
	for _, rec := range records {
		if err := emit(connectors.Record(rec)); err != nil {
			return err
		}
	}
	return nil
}

func (Connector) Write(ctx context.Context, _ connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	if err := ctx.Err(); err != nil {
		return connectors.WriteResult{}, err
	}
	return connectors.WriteResult{RecordsFailed: len(records)}, fmt.Errorf("%s connector is read-only: %w", connectorName, connectors.ErrUnsupportedOperation)
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := secret(cfg, "bearer_token")
	if token == "" {
		return nil, errors.New("timely connector requires secret bearer_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

type streamSpec struct {
	name, description, path, recordsPath string
	fields                               []connectors.Field
}

var streamSpecs = map[string]streamSpec{
	"users":    {"users", "Timely users.", "users", ".", fields("id", "name", "email", "created_at", "updated_at")},
	"projects": {"projects", "Timely projects.", "projects", ".", fields("id", "name", "client_id", "created_at", "updated_at")},
	"clients":  {"clients", "Timely clients.", "clients", ".", fields("id", "name", "created_at", "updated_at")},
	"events":   {"events", "Timely calendar and time events.", "events", ".", fields("id", "project_id", "user_id", "duration", "created_at", "updated_at")},
}

func streams() []connectors.Stream {
	order := []string{"users", "projects", "clients", "events"}
	out := make([]connectors.Stream, 0, len(order))
	for _, name := range order {
		s := streamSpecs[name]
		out = append(out, connectors.Stream{Name: s.name, Description: s.description, Fields: s.fields, PrimaryKey: []string{"id"}})
	}
	return out
}
func fields(names ...string) []connectors.Field {
	out := make([]connectors.Field, 0, len(names))
	for _, name := range names {
		out = append(out, connectors.Field{Name: name, Type: "string"})
	}
	return out
}
func readFixture(ctx context.Context, spec streamSpec, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{"id": fmt.Sprintf("%s_fixture_%d", spec.name, i), "name": fmt.Sprintf("Fixture %s %d", spec.name, i), "email": fmt.Sprintf("fixture+%d@example.com", i), "fixture": true}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}
func accountPath(cfg connectors.RuntimeConfig, resource string) (string, error) {
	account := strings.TrimSpace(cfg.Config["account_id"])
	if account == "" {
		return "", errors.New("timely connector requires config account_id")
	}
	return url.PathEscape(account) + "/" + strings.TrimLeft(resource, "/"), nil
}
func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	raw := strings.TrimSpace(cfg.Config["base_url"])
	if raw == "" {
		raw = defaultBaseURL
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid %s base_url", connectorName)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return "", fmt.Errorf("invalid %s base_url scheme %q", connectorName, u.Scheme)
	}
	return strings.TrimRight(raw, "/"), nil
}
func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets != nil {
		if value := strings.TrimSpace(cfg.Secrets[key]); value != "" {
			return value
		}
	}
	return strings.TrimSpace(cfg.Config[key])
}
