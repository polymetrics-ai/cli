// Package appfollow implements the native pm AppFollow connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit: an
// APIKeyHeader authenticator (X-AppFollow-API-Token), root/dotted-path record
// extraction, and AppFollow-specific stream definitions and endpoints.
//
// It mirrors the stripe reference connector's shape. Like the others, it
// self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// AppFollow is read-only here: the v2 reporting API exposed by this connector
// has no safe reverse-ETL writes, so Capabilities.Write is false.
package appfollow

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	appfollowDefaultBaseURL = "https://api.appfollow.io/api/v2"
	appfollowUserAgent      = "polymetrics-go-cli"
	appfollowTokenHeader    = "X-AppFollow-API-Token"
)

func init() {
	connectors.RegisterFactory("appfollow", New)
}

// New returns the AppFollow connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm AppFollow connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "appfollow" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "appfollow",
		DisplayName:     "Appfollow",
		IntegrationType: "api",
		Description:     "Reads AppFollow account users, app collections, app lists, and per-app rating breakdowns through the AppFollow REST API v2.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to AppFollow. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := appfollowBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(appfollowSecret(cfg)) == "" {
		return errors.New("appfollow connector requires secret api_secret")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// Listing the account users confirms auth and connectivity without mutating
	// anything (the users endpoint takes no required parameters).
	if err := r.DoJSON(ctx, http.MethodGet, "account/users", nil, nil, nil); err != nil {
		return fmt.Errorf("check appfollow: %w", err)
	}
	return nil
}

// Write is unsupported: AppFollow is a read-only reporting source here, so the
// connector does not implement WriteValidator/DryRunWriter and Write always
// returns ErrUnsupportedOperation.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: appfollowStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "users"
	}
	endpoint, ok := appfollowStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("appfollow stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	switch stream {
	case "ratings":
		return c.readRatings(ctx, r, endpoint, req.Config, emit)
	case "app_lists":
		return c.readAppLists(ctx, r, endpoint, req.Config, emit)
	default:
		return c.readSimple(ctx, r, endpoint, nil, emit)
	}
}

// readSimple reads a single-response endpoint and emits each mapped record. The
// AppFollow v2 reporting endpoints return the full result in one body (no
// pagination), so a single request per call is correct.
func (c Connector) readSimple(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, query url.Values, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
	if err != nil {
		return fmt.Errorf("read appfollow %s: %w", endpoint.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return fmt.Errorf("decode appfollow %s: %w", endpoint.resource, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// readAppLists fans out one request per app collection, forwarding the
// collection id as the apps_id query parameter. Collection ids come from the
// app_collection_ids config, or are discovered from /account/apps when absent.
func (c Connector) readAppLists(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	ids := splitList(cfg.Config["app_collection_ids"])
	if len(ids) == 0 {
		discovered, err := c.discoverCollectionIDs(ctx, r)
		if err != nil {
			return err
		}
		ids = discovered
	}
	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{"apps_id": []string{id}}
		if err := c.readSimple(ctx, r, endpoint, query, func(rec connectors.Record) error {
			if rec["app_collection_id"] == nil {
				rec["app_collection_id"] = id
			}
			return emit(rec)
		}); err != nil {
			return err
		}
	}
	return nil
}

// discoverCollectionIDs lists the account's app collections and returns their
// ids, used when app_collection_ids is not configured.
func (c Connector) discoverCollectionIDs(ctx context.Context, r *connsdk.Requester) ([]string, error) {
	resp, err := r.Do(ctx, http.MethodGet, "account/apps", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("read appfollow account/apps: %w", err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "apps")
	if err != nil {
		return nil, fmt.Errorf("decode appfollow account/apps: %w", err)
	}
	ids := make([]string, 0, len(records))
	for _, item := range records {
		if id := stringField(item, "id"); id != "" {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// readRatings fans out one request per configured ext_id, flattening the
// ratings.list rows and stamping the enclosing ext_id/store onto each row.
func (c Connector) readRatings(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	extIDs := splitList(cfg.Config["ext_ids"])
	if len(extIDs) == 0 {
		return errors.New("appfollow ratings stream requires config ext_ids (comma-separated app external ids)")
	}
	for _, ext := range extIDs {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{"ext_id": []string{ext}}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read appfollow ratings ext_id=%s: %w", ext, err)
		}
		rows, encExtID, store := ratingRows(resp.Body)
		if encExtID == "" {
			encExtID = ext
		}
		for _, row := range rows {
			if err := ctx.Err(); err != nil {
				return err
			}
			if row["ext_id"] == nil {
				row["ext_id"] = encExtID
			}
			if row["store"] == nil {
				row["store"] = store
			}
			if err := emit(endpoint.mapRecord(row)); err != nil {
				return err
			}
		}
	}
	return nil
}

// ratingRows extracts the per-day rating rows from a /meta/ratings body. The
// payload nests the rows under ratings.list with ext_id/store as siblings.
func ratingRows(body []byte) (rows []map[string]any, extID, store string) {
	extID, _ = connsdk.StringAt(body, "ratings.ext_id")
	store, _ = connsdk.StringAt(body, "ratings.store")
	list, err := connsdk.RecordsAt(body, "ratings.list")
	if err != nil {
		return nil, extID, store
	}
	out := make([]map[string]any, 0, len(list))
	for _, rec := range list {
		out = append(out, map[string]any(rec))
	}
	return out, extID, store
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise appfollow credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var rec connectors.Record
		switch stream {
		case "users":
			rec = userRecord(map[string]any{
				"id": i, "email": fmt.Sprintf("user+%d@example.com", i),
				"name": fmt.Sprintf("Fixture User %d", i), "role": "admin",
				"status": "active", "updated": "2026-01-01T00:00:00Z",
			})
		case "app_collections":
			rec = appCollectionRecord(map[string]any{
				"id": i, "title": fmt.Sprintf("Fixture Collection %d", i),
				"title_normalized": fmt.Sprintf("fixture collection %d", i),
				"count_apps":       i, "countries": "US", "languages": "en",
				"created": "2026-01-01T00:00:00Z",
			})
		case "app_lists":
			rec = appListRecord(map[string]any{
				"app_id": i, "app_collection_id": 1,
				"ext_id": fmt.Sprintf("ios:%d", 1000+i), "store": "ios",
				"count_reviews": 10 * i, "count_whatsnew": i, "is_favorite": 0,
				"watch_url": "https://apps.apple.com/app/id" + fmt.Sprintf("%d", 1000+i),
				"created":   "2026-01-01T00:00:00Z",
			})
		case "ratings":
			rec = ratingRecord(map[string]any{
				"ext_id": fmt.Sprintf("ios:%d", 1000+i), "store": "ios",
				"date": fmt.Sprintf("2026-01-0%d", i), "country": "US",
				"version": "1.0.0", "rating": 4.0 + float64(i)/10,
				"stars1": 1, "stars2": 2, "stars3": 3, "stars4": 4, "stars5": 5,
				"stars_total": 15 * i,
			})
		default:
			return fmt.Errorf("appfollow fixture stream %q not found", stream)
		}
		rec["connector"] = "appfollow"
		rec["fixture"] = true
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with APIKeyHeader auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := appfollowBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := appfollowSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("appfollow connector requires secret api_secret")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(appfollowTokenHeader, secret, ""),
		UserAgent: appfollowUserAgent,
	}, nil
}

func appfollowSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_secret"]
}

// appfollowBaseURL resolves and validates the base URL. The default is
// api.appfollow.io; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func appfollowBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return appfollowDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("appfollow config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("appfollow config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("appfollow config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// splitList parses a comma-separated config value into trimmed, non-empty parts.
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

func stringField(item map[string]any, key string) string {
	switch v := item[key].(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}
