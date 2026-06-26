package zohobigin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	connectorName   = "zoho-bigin"
	defaultBaseURL  = "https://www.zohoapis.com/bigin/v2"
	defaultTokenURL = "https://accounts.zoho.com/oauth/v2/token"
	userAgent       = "polymetrics-go-cli"
)

type streamSpec struct {
	path       string
	recordPath string
	fields     []connectors.Field
	mapRecord  func(map[string]any) connectors.Record
}

var streams = map[string]streamSpec{
	"pipelines": {path: "Pipelines", recordPath: "data", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "display_value", Type: "string"}}, mapRecord: mapPipeline},
	"records":   {path: "", recordPath: "data", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}}, mapRecord: mapRecord},
	"fields":    {path: "settings/fields", recordPath: "fields", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "api_name", Type: "string"}, {Name: "display_label", Type: "string"}}, mapRecord: mapField},
}

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Zoho Bigin", IntegrationType: "api", Description: "Reads Zoho Bigin pipelines, records, and fields.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true}}
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
	if _, err := tokenURL(cfg); err != nil {
		return err
	}
	if err := requireOAuth(cfg); err != nil {
		return err
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	out := make([]connectors.Stream, 0, len(streams))
	for name, spec := range streams {
		out = append(out, connectors.Stream{Name: name, Description: "Zoho Bigin " + name + ".", PrimaryKey: []string{"id"}, Fields: spec.fields})
	}
	return connectors.Catalog{Connector: c.Name(), Streams: out}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "pipelines"
	}
	spec, ok := streams[stream]
	if !ok {
		return fmt.Errorf("zoho-bigin stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return emitFixture(ctx, stream, spec, emit)
	}
	token, err := c.refreshToken(ctx, req.Config)
	if err != nil {
		return err
	}
	base, err := baseURL(req.Config)
	if err != nil {
		return err
	}
	path := spec.path
	if stream == "records" {
		path = strings.Trim(strings.TrimSpace(req.Config.Config["module_name"]), "/")
		if path == "" {
			path = "Deals"
		}
	}
	r := &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}
	resp, err := r.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return fmt.Errorf("read zoho-bigin %s: %w", stream, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, spec.recordPath)
	if err != nil {
		return err
	}
	for _, item := range records {
		if err := emit(spec.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) refreshToken(ctx context.Context, cfg connectors.RuntimeConfig) (string, error) {
	if err := requireOAuth(cfg); err != nil {
		return "", err
	}
	tokURL, err := tokenURL(cfg)
	if err != nil {
		return "", err
	}
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", strings.TrimSpace(cfg.Secrets["client_id"]))
	form.Set("client_secret", strings.TrimSpace(cfg.Secrets["client_secret"]))
	form.Set("refresh_token", strings.TrimSpace(cfg.Secrets["client_refresh_token"]))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("build zoho-bigin token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request zoho-bigin token: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("zoho-bigin token endpoint returned %s", resp.Status)
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := dec.Decode(&out); err != nil {
		return "", fmt.Errorf("decode zoho-bigin token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("zoho-bigin token response missing access_token")
	}
	return out.AccessToken, nil
}

func mapPipeline(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"], "display_value": item["display_value"]}
}

func mapRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": first(item["name"], item["Deal_Name"], item["display_value"])}
}

func mapField(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item["id"], item["api_name"]), "api_name": item["api_name"], "display_label": item["display_label"]}
}

func emitFixture(ctx context.Context, stream string, spec streamSpec, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	item := map[string]any{"id": stream + "_fixture_1", "name": "Fixture " + stream, "display_value": "Fixture " + stream, "Deal_Name": "Fixture Deal", "api_name": "Fixture_Field", "display_label": "Fixture Field"}
	rec := spec.mapRecord(item)
	rec["fixture"] = true
	return emit(rec)
}

func requireOAuth(cfg connectors.RuntimeConfig) error {
	if strings.TrimSpace(cfg.Secrets["client_id"]) == "" || strings.TrimSpace(cfg.Secrets["client_secret"]) == "" || strings.TrimSpace(cfg.Secrets["client_refresh_token"]) == "" {
		return errors.New("zoho-bigin connector requires client_id, client_secret, and client_refresh_token secrets")
	}
	return nil
}

func first(values ...any) any {
	for _, v := range values {
		if s := fmt.Sprint(v); v != nil && s != "" && s != "<nil>" {
			return v
		}
	}
	return nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func tokenURL(cfg connectors.RuntimeConfig) (string, error) {
	raw := strings.TrimSpace(cfg.Config["token_url"])
	if raw == "" {
		raw = defaultTokenURL
	}
	return validateURL(connectorName, "token_url", raw)
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	raw := strings.TrimSpace(cfg.Config["base_url"])
	if raw == "" {
		raw = defaultBaseURL
	}
	return validateURL(connectorName, "base_url", raw)
}

func validateURL(connector, field, raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("%s config %s is invalid: %w", connector, field, err)
	}
	if parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return "", fmt.Errorf("%s config %s must be an absolute http(s) URL", connector, field)
	}
	return strings.TrimRight(raw, "/"), nil
}
