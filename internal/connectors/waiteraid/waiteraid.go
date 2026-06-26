package waiteraid

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
	connectorName  = "waiteraid"
	defaultBaseURL = "https://api.waiteraid.com"
	userAgent      = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "WaiterAid", IntegrationType: "api", Description: "Reads WaiterAid reservations.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := baseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Secrets["auth_hash"]) == "" || strings.TrimSpace(cfg.Secrets["restid"]) == "" {
		return errors.New("waiteraid connector requires auth_hash and restid secrets")
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: []connectors.Stream{{Name: "reservations", Description: "WaiterAid reservations.", PrimaryKey: []string{"id"}, CursorFields: []string{"date"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "guest_name", Type: "string"}, {Name: "date", Type: "string"}, {Name: "status", Type: "string"}}}}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "reservations"
	}
	if stream != "reservations" {
		return fmt.Errorf("waiteraid stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return emitFixture(ctx, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	q := url.Values{}
	if start := strings.TrimSpace(req.Config.Config["start_date"]); start != "" {
		q.Set("start_date", start)
	}
	resp, err := r.Do(ctx, http.MethodGet, "reservations", q, nil)
	if err != nil {
		return fmt.Errorf("read waiteraid reservations: %w", err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "reservations")
	if err != nil {
		return err
	}
	for _, item := range records {
		if err := emit(connectors.Record{"id": item["id"], "guest_name": item["guest_name"], "date": item["date"], "status": item["status"]}); err != nil {
			return err
		}
	}
	return nil
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	auth := strings.TrimSpace(cfg.Secrets["auth_hash"])
	restaurant := strings.TrimSpace(cfg.Secrets["restid"])
	if auth == "" || restaurant == "" {
		return nil, errors.New("waiteraid connector requires auth_hash and restid secrets")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, DefaultHeaders: map[string]string{"X-Auth-Hash": auth, "X-Restaurant-ID": restaurant}, UserAgent: userAgent}, nil
}

func emitFixture(ctx context.Context, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return emit(connectors.Record{"id": "res_fixture_1", "guest_name": "Fixture Guest", "date": "2026-01-01", "status": "confirmed", "fixture": true})
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	raw := strings.TrimSpace(cfg.Config["base_url"])
	if raw == "" {
		raw = defaultBaseURL
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("waiteraid config base_url is invalid: %w", err)
	}
	if parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return "", errors.New("waiteraid config base_url must be an absolute http(s) URL")
	}
	return strings.TrimRight(raw, "/"), nil
}
