// Package sonarcloud implements a read-only SonarCloud API connector.
package sonarcloud

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
	defaultBaseURL  = "https://sonarcloud.io"
	defaultPageSize = 100
	maxPageSize     = 500
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("sonar-cloud", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "sonar-cloud" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "sonar-cloud", DisplayName: "SonarCloud", IntegrationType: "api", Description: "Reads SonarCloud issues, components, quality gates, and measures through the Web API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	q, err := queryParams("issues", cfg)
	if err != nil {
		return err
	}
	q.Set("ps", "1")
	if err := r.DoJSON(ctx, http.MethodGet, "api/issues/search", q, nil, nil); err != nil {
		return fmt.Errorf("check sonar-cloud: %w", err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, _ connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: "sonar-cloud", Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "issues"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("sonar-cloud stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	q, err := queryParams(stream, req.Config)
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
	"issues":        {"api/issues/search", "issues"},
	"components":    {"api/components/search", "components"},
	"quality_gates": {"api/qualitygates/list", "qualitygates"},
	"measures":      {"api/measures/search", "measures"},
}

func streams() []connectors.Stream {
	names := []string{"issues", "components", "quality_gates", "measures"}
	out := make([]connectors.Stream, 0, len(names))
	for _, name := range names {
		out = append(out, connectors.Stream{Name: name, Description: "SonarCloud " + name + ".", PrimaryKey: []string{"key"}, CursorFields: []string{"createdAt"}, Fields: []connectors.Field{{Name: "key", Type: "string"}, {Name: "component", Type: "string"}, {Name: "createdAt", Type: "string"}, {Name: "severity", Type: "string"}}})
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
		if err := emit(connectors.Record{"key": fmt.Sprintf("%s-%d", stream, i), "component": "fixture:main.go", "severity": "INFO", "createdAt": fmt.Sprintf("2026-01-0%dT00:00:00+0000", i)}); err != nil {
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
	token := secret(cfg, "user_token")
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("sonar-cloud connector requires secret user_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

func queryParams(stream string, cfg connectors.RuntimeConfig) (url.Values, error) {
	size, err := pageSize(cfg)
	if err != nil {
		return nil, err
	}
	q := url.Values{"p": []string{"1"}, "ps": []string{strconv.Itoa(size)}}
	copyConfig(q, cfg, "organization", "organization")
	if keys := firstConfig(cfg, "component_keys", "componentKeys"); keys != "" {
		if stream == "components" {
			q.Set("component", strings.Split(keys, ",")[0])
		} else {
			q.Set("componentKeys", keys)
		}
	}
	copyConfig(q, cfg, "start_date", "createdAfter")
	copyConfig(q, cfg, "end_date", "createdBefore")
	return q, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	return validatedBaseURL("sonar-cloud", configValue(cfg, "base_url"), defaultBaseURL)
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := configValue(cfg, "page_size")
	if raw == "" {
		return defaultPageSize, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 1 || v > maxPageSize {
		return 0, fmt.Errorf("sonar-cloud config page_size must be an integer between 1 and %d", maxPageSize)
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

func firstConfig(cfg connectors.RuntimeConfig, names ...string) string {
	for _, name := range names {
		if v := configValue(cfg, name); v != "" {
			return v
		}
	}
	return ""
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
