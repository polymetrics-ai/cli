// Package jinaaireader implements a conservative read-only connector for Jina AI
// Reader. Reader converts supplied URLs into structured page content; the API is
// not a conventional paginated collection, so Read traverses an allow-listed
// config urls list and emits one record per fetched URL.
package jinaaireader

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	connectorName  = "jina-ai-reader"
	defaultBaseURL = "https://r.jina.ai"
	userAgent      = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory(connectorName, New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Jina AI Reader", IntegrationType: "api", Description: "Reads URL content through the Jina AI Reader API and emits one page record per configured URL.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	urls := targetURLs(cfg)
	if len(urls) == 0 {
		return errors.New("jina-ai-reader connector requires config urls")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if _, err := r.Do(ctx, http.MethodGet, readerPath(urls[0]), nil, nil); err != nil {
		return fmt.Errorf("check jina-ai-reader: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "pages"
	}
	if stream != "pages" {
		return fmt.Errorf("jina-ai-reader stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, emit)
	}
	urls := targetURLs(req.Config)
	if len(urls) == 0 {
		return errors.New("jina-ai-reader pages stream requires config urls")
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	for _, target := range urls {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, readerPath(target), nil, nil)
		if err != nil {
			return fmt.Errorf("read jina-ai-reader %s: %w", target, err)
		}
		rec := pageRecord(target, resp.Body)
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) readFixture(ctx context.Context, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{"url": fmt.Sprintf("https://example.com/fixture/%d", i), "title": fmt.Sprintf("Fixture Page %d", i), "content": "deterministic fixture content", "connector": connectorName, "fixture": true}
		if err := emit(rec); err != nil {
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
	var auth connsdk.Authenticator
	if token := strings.TrimSpace(secret(cfg)); token != "" {
		auth = connsdk.Bearer(token)
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: auth, UserAgent: userAgent, Accept: "application/json"}, nil
}

func pageRecord(target string, body []byte) connectors.Record {
	if rec, ok := jsonPageRecord(target, body); ok {
		return rec
	}
	return connectors.Record{"url": target, "content": string(bytes.TrimSpace(body))}
}

func jsonPageRecord(target string, body []byte) (connectors.Record, bool) {
	var root map[string]any
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	if err := dec.Decode(&root); err != nil {
		return nil, false
	}
	data, _ := root["data"].(map[string]any)
	if data == nil {
		data = root
	}
	urlValue := data["url"]
	if urlValue == nil {
		urlValue = target
	}
	return connectors.Record{"url": urlValue, "title": data["title"], "content": data["content"], "description": data["description"]}, true
}

func targetURLs(cfg connectors.RuntimeConfig) []string {
	return splitList(cfg.Config["urls"])
}

func readerPath(target string) string {
	return "/" + strings.TrimLeft(target, "/")
}

func secret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("jina-ai-reader config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("jina-ai-reader config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("jina-ai-reader config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func splitList(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
