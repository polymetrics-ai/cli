// Package pocket implements a read-only Pocket API connector.
package pocket

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	defaultBaseURL  = "https://getpocket.com/v3"
	defaultPageSize = 100
	defaultMaxPages = 3
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("pocket", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "pocket" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "pocket",
		DisplayName:     "Pocket",
		IntegrationType: "api",
		Description:     "Reads saved Pocket items through the v3 retrieve API.",
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
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	body, err := requestBody(cfg, 1, 0)
	if err != nil {
		return err
	}
	if _, err := r.Do(ctx, http.MethodPost, "get", nil, body); err != nil {
		return fmt.Errorf("check pocket: %w", err)
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
		stream = "items"
	}
	if stream != "items" {
		return fmt.Errorf("pocket stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, emit)
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
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		body, err := requestBody(req.Config, pageSize, offset)
		if err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodPost, "get", nil, body)
		if err != nil {
			return fmt.Errorf("read pocket items: %w", err)
		}
		records, err := itemRecords(resp.Body)
		if err != nil {
			return fmt.Errorf("decode pocket items: %w", err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(itemRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
		offset += pageSize
	}
	return nil
}

func (Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func streams() []connectors.Stream {
	return []connectors.Stream{{Name: "items", Description: "Saved Pocket items.", PrimaryKey: []string{"item_id"}, CursorFields: []string{"updated_at"}, Fields: []connectors.Field{{Name: "item_id", Type: "string"}, {Name: "title", Type: "string"}, {Name: "url", Type: "string"}, {Name: "excerpt", Type: "string"}, {Name: "updated_at", Type: "string"}}}}
}

func itemRecord(item map[string]any) connectors.Record {
	return connectors.Record{"item_id": item["item_id"], "title": first(item, "resolved_title", "given_title"), "url": first(item, "resolved_url", "given_url"), "excerpt": item["excerpt"], "updated_at": first(item, "time_updated", "time_added")}
}

func readFixture(ctx context.Context, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"item_id": fmt.Sprintf("fixture-%d", i), "resolved_title": fmt.Sprintf("Fixture Item %d", i), "given_url": fmt.Sprintf("https://example.com/%d", i), "time_updated": fmt.Sprintf("176000000%d", i)}
		if err := emit(itemRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

func itemRecords(body []byte) ([]map[string]any, error) {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var root struct {
		List map[string]map[string]any `json:"list"`
	}
	if err := dec.Decode(&root); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(root.List))
	for key := range root.List {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]map[string]any, 0, len(keys))
	for _, key := range keys {
		item := root.List[key]
		if item["item_id"] == nil {
			item["item_id"] = key
		}
		out = append(out, item)
	}
	return out, nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, UserAgent: userAgent}, nil
}

func requestBody(cfg connectors.RuntimeConfig, count, offset int) (map[string]any, error) {
	consumerKey := strings.TrimSpace(cfg.Secrets["consumer_key"])
	accessToken := strings.TrimSpace(cfg.Secrets["access_token"])
	if consumerKey == "" || accessToken == "" {
		return nil, errors.New("pocket connector requires secrets consumer_key and access_token")
	}
	body := map[string]any{"consumer_key": consumerKey, "access_token": accessToken, "count": count, "offset": offset, "detailType": valueOrDefault(cfg.Config["detail_type"], "complete")}
	for _, key := range []string{"state", "favorite", "tag", "contentType", "sort", "since"} {
		if v := strings.TrimSpace(cfg.Config[key]); v != "" {
			body[key] = v
		}
	}
	return body, nil
}

func valueOrDefault(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return strings.TrimSpace(v)
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("pocket config base_url is invalid: %w", err)
	}
	if (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return "", errors.New("pocket config base_url must be an absolute http or https URL")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return intConfig(cfg, "page_size", defaultPageSize, 1, 500)
}
func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	return intConfig(cfg, "max_pages", defaultMaxPages, 0, 10000)
}

func intConfig(cfg connectors.RuntimeConfig, key string, def, min, max int) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config[key]))
	if raw == "" {
		return def, nil
	}
	if key == "max_pages" && (raw == "all" || raw == "unlimited") {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < min {
		return 0, fmt.Errorf("pocket config %s must be an integer >= %d", key, min)
	}
	if max > 0 && value > max {
		return max, nil
	}
	return value, nil
}

func first(m map[string]any, keys ...string) any {
	for _, key := range keys {
		if v, ok := m[key]; ok {
			return v
		}
	}
	return nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
