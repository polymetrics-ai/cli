// Package facebookpages implements a read-only Facebook Pages Graph API
// connector. It covers page metadata and posts; post insight fan-out is not
// implemented because it requires per-post metric selection and permissions.
package facebookpages

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
	connectorName  = "facebook-pages"
	defaultBaseURL = "https://graph.facebook.com/v19.0"
	userAgent      = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Facebook Pages", IntegrationType: "api", Description: "Reads Facebook Page metadata and posts from the Graph API. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	q := url.Values{"fields": []string{"id,name"}}
	if err := r.DoJSON(ctx, http.MethodGet, pagePath(cfg), q, nil, nil); err != nil {
		return fmt.Errorf("check facebook-pages: %w", err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: connectorName, Streams: []connectors.Stream{
		{Name: "page", Description: "Facebook Page metadata.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "category", Type: "string"}, {Name: "fan_count", Type: "integer"}, {Name: "link", Type: "string"}}},
		{Name: "posts", Description: "Posts published by the configured Facebook Page.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_time"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "message", Type: "string"}, {Name: "created_time", Type: "timestamp"}, {Name: "updated_time", Type: "timestamp"}, {Name: "permalink_url", Type: "string"}}},
	}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "posts"
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	switch stream {
	case "page":
		q := url.Values{"fields": []string{"id,name,category,fan_count,link"}}
		resp, err := r.Do(ctx, http.MethodGet, pagePath(req.Config), q, nil)
		if err != nil {
			return fmt.Errorf("read facebook page: %w", err)
		}
		items, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return err
		}
		for _, item := range items {
			if err := emit(record(item)); err != nil {
				return err
			}
		}
		return nil
	case "posts":
		return c.readPosts(ctx, r, req.Config, emit)
	default:
		return fmt.Errorf("facebook-pages stream %q not found", stream)
	}
}

func (c Connector) readPosts(ctx context.Context, r *connsdk.Requester, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	q := url.Values{}
	q.Set("fields", "id,message,created_time,updated_time,permalink_url")
	q.Set("limit", pageSize(cfg))
	path := pagePath(cfg) + "/posts"
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, q, nil)
		if err != nil {
			return fmt.Errorf("read facebook posts: %w", err)
		}
		items, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return err
		}
		for _, item := range items {
			if err := emit(record(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "paging.next")
		if err != nil {
			return err
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		path = next
		q = nil
	}
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(cfg.Config["page_id"]) == "" {
		return nil, errors.New("facebook-pages connector requires config page_id")
	}
	token := strings.TrimSpace(cfg.Secrets["access_token"])
	if token == "" {
		return nil, errors.New("facebook-pages connector requires secret access_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

func pagePath(cfg connectors.RuntimeConfig) string {
	return "/" + url.PathEscape(strings.TrimSpace(cfg.Config["page_id"]))
}

func pageSize(cfg connectors.RuntimeConfig) string {
	n, err := strconv.Atoi(strings.TrimSpace(cfg.Config["page_size"]))
	if err != nil || n < 1 || n > 100 {
		n = 100
	}
	return strconv.Itoa(n)
}

func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	if stream == "page" {
		return emit(connectors.Record{"id": "page_fixture", "name": "Fixture Page", "fixture": true})
	}
	if stream != "posts" {
		return fmt.Errorf("facebook-pages stream %q not found", stream)
	}
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": fmt.Sprintf("post_fixture_%d", i), "message": fmt.Sprintf("Fixture post %d", i), "created_time": fmt.Sprintf("2026-01-%02dT00:00:00+0000", i), "fixture": true}); err != nil {
			return err
		}
	}
	return nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	raw := strings.TrimSpace(cfg.Config["base_url"])
	if raw == "" {
		raw = defaultBaseURL
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return "", fmt.Errorf("facebook-pages config base_url must be an absolute http(s) URL")
	}
	return strings.TrimRight(raw, "/"), nil
}

func record(in map[string]any) connectors.Record {
	out := connectors.Record{}
	for k, v := range in {
		out[k] = v
	}
	return out
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
