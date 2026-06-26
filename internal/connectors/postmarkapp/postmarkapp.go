// Package postmarkapp implements a read-only native connector for the Postmark
// HTTP API. It uses Postmark's account token for account-level resources and the
// server token for message resources.
package postmarkapp

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
	defaultBaseURL  = "https://api.postmarkapp.com"
	defaultPageSize = 100
	maxPageSize     = 500
	userAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("postmarkapp", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

type streamSpec struct {
	path       string
	recordsKey string
	tokenName  string
}

var streamSpecs = map[string]streamSpec{
	"servers":           {path: "servers", recordsKey: "Servers", tokenName: "X-Postmark-Account-Token"},
	"outbound_messages": {path: "messages/outbound", recordsKey: "Messages", tokenName: "X-Postmark-Server-Token"},
	"inbound_messages":  {path: "messages/inbound", recordsKey: "InboundMessages", tokenName: "X-Postmark-Server-Token"},
}

func (Connector) Name() string { return "postmarkapp" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "postmarkapp",
		DisplayName:     "Postmark App",
		IntegrationType: "api",
		Description:     "Reads Postmark servers and message activity through the Postmark HTTP API. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	r, err := c.requester(cfg, streamSpecs["servers"])
	if err != nil {
		return err
	}
	query := url.Values{"count": []string{"1"}, "offset": []string{"0"}}
	if err := r.DoJSON(ctx, http.MethodGet, "servers", query, nil, nil); err != nil {
		return fmt.Errorf("check postmarkapp: %w", err)
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
		stream = "outbound_messages"
	}
	spec, ok := streamSpecs[stream]
	if !ok {
		return fmt.Errorf("postmarkapp stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config, spec)
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
	for offset, page := 0, 0; maxPages == 0 || page < maxPages; offset, page = offset+pageSize, page+1 {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{"count": []string{strconv.Itoa(pageSize)}, "offset": []string{strconv.Itoa(offset)}}
		resp, err := r.Do(ctx, http.MethodGet, spec.path, query, nil)
		if err != nil {
			return fmt.Errorf("read postmarkapp %s: %w", spec.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, spec.recordsKey)
		if err != nil {
			return fmt.Errorf("decode postmarkapp %s: %w", spec.path, err)
		}
		for _, item := range records {
			if err := emit(mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig, spec streamSpec) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := secret(cfg, spec.tokenName)
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("postmarkapp connector requires secret %s", spec.tokenName)
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(spec.tokenName, token, ""),
		UserAgent: userAgent,
	}, nil
}

func streams() []connectors.Stream {
	common := []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "subject", Type: "string"}, {Name: "created_at", Type: "timestamp"}, {Name: "updated_at", Type: "timestamp"}}
	return []connectors.Stream{
		{Name: "servers", Description: "Postmark account servers.", Fields: common, PrimaryKey: []string{"id"}},
		{Name: "outbound_messages", Description: "Postmark outbound message activity.", Fields: common, PrimaryKey: []string{"id"}, CursorFields: []string{"received_at"}},
		{Name: "inbound_messages", Description: "Postmark inbound message activity.", Fields: common, PrimaryKey: []string{"id"}, CursorFields: []string{"received_at"}},
	}
}

func mapRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":          first(item, "MessageID", "ID", "ID"),
		"name":        item["Name"],
		"subject":     item["Subject"],
		"from":        item["From"],
		"to":          item["To"],
		"status":      first(item, "Status", "MessageStatus"),
		"created_at":  item["CreatedAt"],
		"updated_at":  item["UpdatedAt"],
		"received_at": item["ReceivedAt"],
	}
	return rec
}

func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": fmt.Sprintf("%s-%d", stream, i), "name": fmt.Sprintf("Fixture %s %d", stream, i), "subject": fmt.Sprintf("Fixture subject %d", i), "received_at": fmt.Sprintf("2026-01-0%dT00:00:00Z", i)}); err != nil {
			return err
		}
	}
	return nil
}

func first(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if v := item[key]; v != nil {
			return v
		}
	}
	return nil
}

func secret(cfg connectors.RuntimeConfig, name string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[name]
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("postmarkapp config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("postmarkapp config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("postmarkapp config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("postmarkapp config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, errors.New("postmarkapp config max_pages must be 0, all, unlimited, or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
