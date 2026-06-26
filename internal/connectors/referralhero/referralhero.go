// Package referralhero implements a conservative read-only ReferralHero API v2
// connector. ReferralHero deployments expose several list-scoped resources; this
// package supports documented list-style endpoints with bearer auth and keeps
// writes unsupported.
package referralhero

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
	referralHeroDefaultBaseURL  = "https://app.referralhero.com/api/v2"
	referralHeroDefaultPageSize = 100
	referralHeroMaxPageSize     = 250
	referralHeroUserAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("referralhero", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "referralhero" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "referralhero", DisplayName: "ReferralHero", IntegrationType: "api", Description: "Reads ReferralHero lists, subscribers, referrals, and rewards through the API v2 list endpoints. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "lists", url.Values{"page": []string{"1"}, "per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check referralhero: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: referralHeroStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "lists"
	}
	endpoint, ok := referralHeroEndpoints[stream]
	if !ok {
		return fmt.Errorf("referralhero stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, emit)
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
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	page := 1
	for fetched := 0; maxPages == 0 || fetched < maxPages; fetched++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("per_page", strconv.Itoa(pageSize))
		resp, err := r.Do(ctx, http.MethodGet, endpoint.path, query, nil)
		if err != nil {
			return fmt.Errorf("read referralhero %s: %w", endpoint.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode referralhero %s: %w", endpoint.path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "pagination.next_page")
		if err != nil {
			return fmt.Errorf("decode referralhero %s next_page: %w", endpoint.path, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		nextPage, err := strconv.Atoi(next)
		if err != nil || nextPage <= page {
			return nil
		}
		page = nextPage
	}
	return nil
}

func (c Connector) readFixture(ctx context.Context, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": fmt.Sprintf("%s_fixture_%d", endpoint.path, i), "name": fmt.Sprintf("Fixture %d", i), "email": fmt.Sprintf("fixture+%d@example.com", i), "status": "active", "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-02T00:00:00Z", "referral_code": fmt.Sprintf("FIX%d", i)}
		if err := emit(endpoint.mapRecord(item)); err != nil {
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
	key := strings.TrimSpace(secret(cfg, "api_key"))
	if key == "" {
		return nil, errors.New("referralhero connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(key), UserAgent: referralHeroUserAgent}, nil
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct {
	path        string
	recordsPath string
	mapRecord   func(map[string]any) connectors.Record
}

var referralHeroEndpoints = map[string]streamEndpoint{
	"lists":       {path: "lists", recordsPath: "data", mapRecord: listRecord},
	"subscribers": {path: "subscribers", recordsPath: "data", mapRecord: subscriberRecord},
	"referrals":   {path: "referrals", recordsPath: "data", mapRecord: referralRecord},
	"rewards":     {path: "rewards", recordsPath: "data", mapRecord: rewardRecord},
}

func referralHeroStreams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "lists", Description: "ReferralHero lists.", PrimaryKey: []string{"id"}, Fields: fields("id", "name", "status", "created_at")},
		{Name: "subscribers", Description: "ReferralHero subscribers.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields("id", "email", "name", "status", "referral_code", "updated_at")},
		{Name: "referrals", Description: "ReferralHero referrals.", PrimaryKey: []string{"id"}, CursorFields: []string{"created_at"}, Fields: fields("id", "subscriber_id", "email", "status", "created_at")},
		{Name: "rewards", Description: "ReferralHero rewards.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields("id", "name", "status", "updated_at")},
	}
}

func listRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"], "status": item["status"], "created_at": item["created_at"]}
}
func subscriberRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "email": item["email"], "name": item["name"], "status": item["status"], "referral_code": item["referral_code"], "updated_at": item["updated_at"]}
}
func referralRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "subscriber_id": item["subscriber_id"], "email": item["email"], "status": item["status"], "created_at": item["created_at"]}
}
func rewardRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"], "status": item["status"], "updated_at": item["updated_at"]}
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
		return referralHeroDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("referralhero config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("referralhero config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("referralhero config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return referralHeroDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > referralHeroMaxPageSize {
		return 0, fmt.Errorf("referralhero config page_size must be between 1 and %d", referralHeroMaxPageSize)
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
		return 0, errors.New("referralhero config max_pages must be a non-negative integer, all, or unlimited")
	}
	return value, nil
}
