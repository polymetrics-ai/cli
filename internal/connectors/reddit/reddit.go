// Package reddit implements a read-only connector for Reddit OAuth API listing
// endpoints. It reads subreddit posts and comments using bearer tokens supplied
// by the caller; OAuth token acquisition is intentionally out of scope.
package reddit

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
	redditDefaultBaseURL  = "https://oauth.reddit.com"
	redditDefaultPageSize = 100
	redditMaxPageSize     = 100
	redditUserAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("reddit", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "reddit" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "reddit", DisplayName: "Reddit", IntegrationType: "api", Description: "Reads subreddit posts and comments through the Reddit OAuth API. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	r, subreddit, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, "r/"+subreddit+"/new", url.Values{"limit": []string{"1"}, "raw_json": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check reddit: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: redditStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "posts"
	}
	endpoint, ok := redditEndpoints[stream]
	if !ok {
		return fmt.Errorf("reddit stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, emit)
	}
	r, subreddit, err := c.requester(req.Config)
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
	return c.harvest(ctx, r, subreddit, endpoint, pageSize, maxPages, emit)
}

func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, subreddit string, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	after := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("raw_json", "1")
		if after != "" {
			query.Set("after", after)
		}
		resp, err := r.Do(ctx, http.MethodGet, "r/"+subreddit+"/"+endpoint.path, query, nil)
		if err != nil {
			return fmt.Errorf("read reddit %s: %w", endpoint.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data.children")
		if err != nil {
			return fmt.Errorf("decode reddit %s: %w", endpoint.path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "data.after")
		if err != nil {
			return fmt.Errorf("decode reddit %s after: %w", endpoint.path, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		after = next
	}
	return nil
}

func (c Connector) readFixture(ctx context.Context, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"kind": "t3", "data": map[string]any{"id": fmt.Sprintf("fixture_%d", i), "name": fmt.Sprintf("t3_fixture_%d", i), "title": fmt.Sprintf("Fixture post %d", i), "body": fmt.Sprintf("Fixture comment %d", i), "subreddit": "fixture", "author": "fixture_user", "created_utc": int64(1767225600 + i)}}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, string, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, "", err
	}
	subreddit := cleanSegment(strings.TrimSpace(cfg.Config["subreddit"]))
	if subreddit == "" {
		return nil, "", errors.New("reddit connector requires config subreddit")
	}
	token := strings.TrimSpace(secret(cfg, "access_token"))
	if token == "" {
		return nil, "", errors.New("reddit connector requires secret access_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: redditUserAgent}, subreddit, nil
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct {
	path      string
	mapRecord func(map[string]any) connectors.Record
}

var redditEndpoints = map[string]streamEndpoint{
	"posts":    {path: "new", mapRecord: postRecord},
	"comments": {path: "comments", mapRecord: commentRecord},
}

func redditStreams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "posts", Description: "Recent subreddit posts.", PrimaryKey: []string{"id"}, CursorFields: []string{"created_utc"}, Fields: fields("id", "name", "title", "subreddit", "author", "created_utc")},
		{Name: "comments", Description: "Recent subreddit comments.", PrimaryKey: []string{"id"}, CursorFields: []string{"created_utc"}, Fields: fields("id", "name", "body", "subreddit", "author", "created_utc")},
	}
}

func postRecord(item map[string]any) connectors.Record {
	data := childData(item)
	return connectors.Record{"id": data["id"], "name": data["name"], "title": data["title"], "subreddit": data["subreddit"], "author": data["author"], "created_utc": data["created_utc"], "permalink": data["permalink"]}
}

func commentRecord(item map[string]any) connectors.Record {
	data := childData(item)
	return connectors.Record{"id": data["id"], "name": data["name"], "body": first(data, "body", "title"), "subreddit": data["subreddit"], "author": data["author"], "created_utc": data["created_utc"], "permalink": data["permalink"]}
}

func childData(item map[string]any) map[string]any {
	if data, ok := item["data"].(map[string]any); ok {
		return data
	}
	return item
}

func first(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if v := item[key]; v != nil {
			return v
		}
	}
	return nil
}

func fields(names ...string) []connectors.Field {
	out := make([]connectors.Field, 0, len(names))
	for _, name := range names {
		out = append(out, connectors.Field{Name: name, Type: "string"})
	}
	return out
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return redditDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("reddit config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("reddit config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("reddit config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func cleanSegment(value string) string {
	if value == "" || strings.ContainsAny(value, "/?#") || strings.Contains(value, "..") {
		return ""
	}
	return value
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return redditDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > redditMaxPageSize {
		return 0, fmt.Errorf("reddit config page_size must be between 1 and %d", redditMaxPageSize)
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
		return 0, errors.New("reddit config max_pages must be a non-negative integer, all, or unlimited")
	}
	return value, nil
}
