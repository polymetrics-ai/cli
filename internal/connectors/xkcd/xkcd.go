// Package xkcd implements a read-only native XKCD JSON connector.
package xkcd

import (
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
	defaultBaseURL = "https://xkcd.com"
	userAgent      = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("xkcd", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "xkcd" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "xkcd", DisplayName: "XKCD", IntegrationType: "api", Description: "Reads public XKCD comic metadata from the JSON API. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "info.0.json", nil, nil, nil); err != nil {
		return fmt.Errorf("check xkcd: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	fields := []connectors.Field{{Name: "num", Type: "number"}, {Name: "title", Type: "string"}, {Name: "safe_title", Type: "string"}, {Name: "year", Type: "string"}, {Name: "month", Type: "string"}, {Name: "day", Type: "string"}}
	streams := []connectors.Stream{
		{Name: "latest", Description: "Latest XKCD comic metadata.", Fields: fields, PrimaryKey: []string{"num"}},
		{Name: "comic", Description: "Specific XKCD comic metadata selected by config comic_number.", Fields: fields, PrimaryKey: []string{"num"}},
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "latest"
	}
	if stream != "latest" && stream != "comic" {
		return fmt.Errorf("xkcd stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	path := "info.0.json"
	if stream == "comic" {
		num := strings.TrimSpace(req.Config.Config["comic_number"])
		if num == "" || strings.ContainsAny(num, "/?#") {
			return errors.New("xkcd config comic_number is required for comic and must be a path segment")
		}
		path = url.PathEscape(num) + "/info.0.json"
	}
	resp, err := r.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return fmt.Errorf("read xkcd %s: %w", stream, err)
	}
	var rec connectors.Record
	if err := json.Unmarshal(resp.Body, &rec); err != nil {
		return fmt.Errorf("decode xkcd %s: %w", stream, err)
	}
	return emit(rec)
}

func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	rec := connectors.Record{"num": 1, "title": "Fixture XKCD", "safe_title": "Fixture XKCD", "year": "2026", "month": "1", "day": "1", "stream": stream, "fixture": true}
	return emit(rec)
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg, defaultBaseURL, "xkcd")
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, UserAgent: userAgent}, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func baseURL(cfg connectors.RuntimeConfig, fallback, connector string) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = fallback
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("%s config base_url is invalid: %w", connector, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("%s config base_url must use http or https", connector)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("%s config base_url must include a host", connector)
	}
	return strings.TrimRight(base, "/"), nil
}
