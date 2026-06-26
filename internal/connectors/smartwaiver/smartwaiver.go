// Package smartwaiver implements a read-only Smartwaiver API connector.
package smartwaiver

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
	defaultBaseURL  = "https://api.smartwaiver.com"
	defaultPageSize = 100
	maxPageSize     = 100
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("smartwaiver", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "smartwaiver" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "smartwaiver", DisplayName: "Smartwaiver", IntegrationType: "api", Description: "Reads Smartwaiver waivers, checkins, templates, published keys, and user info.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "v4/me", nil, nil, nil); err != nil {
		return fmt.Errorf("check smartwaiver: %w", err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, _ connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: "smartwaiver", Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "waivers"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("smartwaiver stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	q, err := queryParams(req.Config)
	if err != nil {
		return err
	}
	return readRecords(ctx, r, endpoint.resource, endpoint.recordsPath, q, emit)
}

func (Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct{ resource, recordsPath string }

var streamEndpoints = map[string]streamEndpoint{
	"waivers":        {"v4/waivers", "waivers.waivers"},
	"checkins":       {"v4/checkins", "checkins.checkins"},
	"templates":      {"v4/templates", "templates.templates"},
	"published_keys": {"v4/keys/published", "published_keys.keys"},
	"user_info":      {"v4/info", ""},
}

func streams() []connectors.Stream {
	names := []string{"waivers", "checkins", "templates", "published_keys", "user_info"}
	out := make([]connectors.Stream, 0, len(names))
	for _, name := range names {
		out = append(out, connectors.Stream{Name: name, Description: "Smartwaiver " + name + ".", PrimaryKey: []string{"waiverId"}, Fields: []connectors.Field{{Name: "waiverId", Type: "string"}, {Name: "templateId", Type: "string"}, {Name: "createdAt", Type: "string"}}})
	}
	return out
}

func readRecords(ctx context.Context, r *connsdk.Requester, resource, recordsPath string, q url.Values, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, resource, q, nil)
	if err != nil {
		return err
	}
	records, err := connsdk.RecordsAt(resp.Body, recordsPath)
	if err != nil {
		return err
	}
	for _, rec := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record(rec)); err != nil {
			return err
		}
	}
	return nil
}

func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"waiverId": fmt.Sprintf("waiver-%d", i), "templateId": fmt.Sprintf("template-%d", i), "stream": stream}); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	key := secret(cfg, "api_key")
	if strings.TrimSpace(key) == "" {
		return nil, errors.New("smartwaiver connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(key), UserAgent: userAgent}, nil
}

func queryParams(cfg connectors.RuntimeConfig) (url.Values, error) {
	size, err := pageSize(cfg)
	if err != nil {
		return nil, err
	}
	q := url.Values{"limit": []string{strconv.Itoa(size)}, "offset": []string{"0"}}
	copyConfig(q, cfg, "start_date", "fromDts")
	copyConfig(q, cfg, "start_date_2", "fromDts")
	copyConfig(q, cfg, "end_date", "toDts")
	return q, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	return validatedBaseURL("smartwaiver", configValue(cfg, "base_url"), defaultBaseURL)
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := configValue(cfg, "page_size")
	if raw == "" {
		return defaultPageSize, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 1 || v > maxPageSize {
		return 0, fmt.Errorf("smartwaiver config page_size must be an integer between 1 and %d", maxPageSize)
	}
	return v, nil
}

func validatedBaseURL(connector, raw, def string) (string, error) {
	base := strings.TrimSpace(raw)
	if base == "" {
		return def, nil
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("%s config base_url is invalid: %w", connector, err)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return "", fmt.Errorf("%s config base_url must use http or https", connector)
	}
	if u.Host == "" {
		return "", fmt.Errorf("%s config base_url must include a host", connector)
	}
	return strings.TrimRight(base, "/"), nil
}

func configValue(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Config == nil {
		return ""
	}
	return strings.TrimSpace(cfg.Config[key])
}

func copyConfig(q url.Values, cfg connectors.RuntimeConfig, from, to string) {
	if v := configValue(cfg, from); v != "" {
		q.Set(to, v)
	}
}

func secret(cfg connectors.RuntimeConfig, name string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[name]
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(configValue(cfg, "mode"), "fixture")
}
