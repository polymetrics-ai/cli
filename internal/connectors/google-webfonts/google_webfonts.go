// Package googlewebfonts implements the native pm Google Web Fonts connector. It
// is a declarative-HTTP per-system connector following the stripe template: a
// thin package that composes the connsdk toolkit (Requester + API-key query
// auth + RecordsAt extraction) with Google-Web-Fonts-specific stream
// definitions and endpoints.
//
// The Google Fonts Developer API exposes a single list resource,
// https://www.googleapis.com/webfonts/v1/webfonts, authenticated with an API
// key supplied as the `key` query parameter. The response is a JSON object with
// the font families under the `items` array. The published streams are different
// sorted views of that list (default, popularity, trending, date, alpha) routed
// by the per-stream `sort` parameter. The API does not document pagination, but
// the read loop honours an optional `nextPageToken`/`pageToken` cursor so the
// connector keeps working if Google adds paging. The source is read-only
// (full-refresh): there are no reverse-ETL writes, so Capabilities.Write=false.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package googlewebfonts

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/httpsource"
)

const (
	defaultBaseURL = "https://www.googleapis.com/webfonts/v1"
	listResource   = "webfonts"
	userAgent      = "polymetrics-go-cli"
	maxPagesCap    = 100 // safety bound on the optional pageToken loop
)

func init() {
	connectors.RegisterFactory("google-webfonts", New)
}

// New returns the Google Web Fonts connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Google Web Fonts connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "google-webfonts" }

func (c Connector) source() httpsource.Source {
	return httpsource.Source{Client: c.Client, Spec: httpsource.Spec{
		Name:            "google-webfonts",
		DisplayName:     "Google Webfonts",
		IntegrationType: "api",
		Description:     "Reads Google Web Fonts families (default, popular, trending, newest, and alphabetical views) through the Google Fonts Developer API. Read-only.",
		DefaultBaseURL:  defaultBaseURL,
		DefaultStream:   "webfonts",
		UserAgent:       userAgent,
		Auth:            httpsource.AuthSpec{Type: httpsource.AuthAPIKeyQuery, SecretName: "api_key", QueryParam: "key"},
		DefaultMaxPages: maxPagesCap,
		Streams:         httpStreams(),
	}}
}

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "google-webfonts",
		DisplayName:     "Google Webfonts",
		IntegrationType: "api",
		Description:     "Reads Google Web Fonts families (default, popular, trending, newest, and alphabetical views) through the Google Fonts Developer API. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to the Google
// Web Fonts API. In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	return c.source().Check(ctx, cfg)
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	return c.source().Catalog(ctx, cfg)
}

// InitialState satisfies connectors.StatefulReader: a stream starts with an
// empty incremental cursor (full sync). The Google Web Fonts API does not accept
// a server-side lastModified filter, so the cursor is advisory only.
func (c Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	return c.source().InitialState(ctx, stream, cfg)
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	return c.source().Read(ctx, req, emit)
}

func httpStreams() []httpsource.StreamSpec {
	catalog := streams()
	out := make([]httpsource.StreamSpec, 0, len(catalog))
	for _, stream := range catalog {
		endpoint := streamEndpoints[stream.Name]
		sort := endpoint.sort
		mapper := endpoint.mapRecord
		out = append(out, httpsource.StreamSpec{
			Name:           stream.Name,
			Description:    stream.Description,
			Fields:         stream.Fields,
			PrimaryKey:     stream.PrimaryKey,
			CursorFields:   stream.CursorFields,
			Method:         http.MethodGet,
			Path:           listResource,
			RecordsPath:    "items",
			FixtureRecords: googleFixtureRecords(),
			Query: func(cfg connectors.RuntimeConfig, pageSize int) url.Values {
				query := url.Values{}
				if sort != "" {
					query.Set("sort", sort)
				}
				for k, v := range optionalQuery(cfg) {
					query.Set(k, v)
				}
				return query
			},
			Paginator: func(pageSize int) connsdk.Paginator {
				return &connsdk.CursorPaginator{CursorParam: "pageToken", TokenPath: "nextPageToken"}
			},
			Map: func(record connectors.Record) (connectors.Record, error) {
				return mapper(map[string]any(record)), nil
			},
		})
	}
	return out
}

func googleFixtureRecords() []connectors.Record {
	fixtures := []map[string]any{
		{
			"family":       "Roboto",
			"category":     "sans-serif",
			"version":      "v30",
			"lastModified": "2026-01-01",
			"kind":         "webfonts#webfont",
			"menu":         "https://fonts.gstatic.com/s/roboto/menu.ttf",
			"variants":     []any{"100", "300", "regular", "500", "700"},
			"subsets":      []any{"latin", "latin-ext", "cyrillic"},
			"files":        map[string]any{"regular": "https://fonts.gstatic.com/s/roboto/regular.ttf"},
		},
		{
			"family":       "Open Sans",
			"category":     "sans-serif",
			"version":      "v40",
			"lastModified": "2026-01-02",
			"kind":         "webfonts#webfont",
			"menu":         "https://fonts.gstatic.com/s/opensans/menu.ttf",
			"variants":     []any{"300", "regular", "600", "700"},
			"subsets":      []any{"latin", "greek"},
			"files":        map[string]any{"regular": "https://fonts.gstatic.com/s/opensans/regular.ttf"},
		},
	}
	out := make([]connectors.Record, 0, len(fixtures))
	for _, fixture := range fixtures {
		record := fontRecord(fixture)
		record["connector"] = "google-webfonts"
		record["fixture"] = true
		out = append(out, record)
	}
	return out
}

// harvest reads the Google Web Fonts list endpoint. The API returns the full
// font list in a single response under "items"; there is no documented
// pagination. The loop nonetheless follows an optional "nextPageToken" in the
// body (sent back as "pageToken") so the connector keeps working if Google adds
// paging, and is bounded by maxPagesCap.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	base := url.Values{}
	if endpoint.sort != "" {
		base.Set("sort", endpoint.sort)
	}
	for k, v := range optionalQuery(cfg) {
		base.Set(k, v)
	}

	pageToken := ""
	for page := 0; page < maxPagesCap; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if pageToken != "" {
			query.Set("pageToken", pageToken)
		}
		resp, err := r.Do(ctx, http.MethodGet, listResource, query, nil)
		if err != nil {
			return fmt.Errorf("read google-webfonts %s: %w", listResource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "items")
		if err != nil {
			return fmt.Errorf("decode google-webfonts %s page: %w", listResource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "nextPageToken")
		if err != nil {
			return fmt.Errorf("decode google-webfonts %s nextPageToken: %w", listResource, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		pageToken = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free.
func (c Connector) readFixture(ctx context.Context, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	fixtures := []map[string]any{
		{
			"family":       "Roboto",
			"category":     "sans-serif",
			"version":      "v30",
			"lastModified": "2026-01-01",
			"kind":         "webfonts#webfont",
			"menu":         "https://fonts.gstatic.com/s/roboto/menu.ttf",
			"variants":     []any{"100", "300", "regular", "500", "700"},
			"subsets":      []any{"latin", "latin-ext", "cyrillic"},
			"files":        map[string]any{"regular": "https://fonts.gstatic.com/s/roboto/regular.ttf"},
		},
		{
			"family":       "Open Sans",
			"category":     "sans-serif",
			"version":      "v40",
			"lastModified": "2026-01-02",
			"kind":         "webfonts#webfont",
			"menu":         "https://fonts.gstatic.com/s/opensans/menu.ttf",
			"variants":     []any{"300", "regular", "600", "700"},
			"subsets":      []any{"latin", "greek"},
			"files":        map[string]any{"regular": "https://fonts.gstatic.com/s/opensans/regular.ttf"},
		},
	}
	for _, item := range fixtures {
		if err := ctx.Err(); err != nil {
			return err
		}
		record := endpoint.mapRecord(item)
		record["connector"] = "google-webfonts"
		record["fixture"] = true
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with API-key query auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyQuery; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	key := secret(cfg)
	if strings.TrimSpace(key) == "" {
		return nil, errors.New("google-webfonts connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery("key", key),
		UserAgent: userAgent,
	}, nil
}

// optionalQuery collects the API's optional passthrough query parameters from
// config (alt, prettyPrint, family, subset, category, capability). The `sort`
// param is owned by the stream definition and not read here.
func optionalQuery(cfg connectors.RuntimeConfig) map[string]string {
	out := map[string]string{}
	for _, key := range []string{"alt", "prettyPrint", "family", "subset", "category", "capability"} {
		if v := strings.TrimSpace(cfg.Config[key]); v != "" {
			out[key] = v
		}
	}
	return out
}

func secret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// baseURL resolves and validates the base URL. The default is
// www.googleapis.com/webfonts/v1; any override must be an absolute https (or
// http for local test servers) URL with a host to bound SSRF risk.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return defaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("google-webfonts config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("google-webfonts config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("google-webfonts config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func cloneValues(in url.Values) url.Values {
	out := url.Values{}
	for k, vs := range in {
		for _, v := range vs {
			out.Add(k, v)
		}
	}
	return out
}

// Write satisfies the connectors.Connector interface. The Google Web Fonts API
// is read-only, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return c.source().Write(ctx, req, records)
}
