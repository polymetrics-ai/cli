// Package tmdb implements a read-only native Go connector for The Movie Database API.
package tmdb

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
	connectorName   = "tmdb"
	defaultBaseURL  = "https://api.themoviedb.org/3"
	defaultPageSize = 20
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "TMDb", IntegrationType: "api", Description: "Reads movies and search results from The Movie Database API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	q, err := baseQuery(cfg)
	if err != nil {
		return err
	}
	q.Set("page", "1")
	if err := r.DoJSON(ctx, http.MethodGet, "movie/popular", q, nil, nil); err != nil {
		return safeError("check tmdb", err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, _ connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: connectorName, Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "popular_movies"
	}
	spec, ok := streamSpecs[stream]
	if !ok {
		return fmt.Errorf("%s stream %q not found", connectorName, stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, spec, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	path, err := spec.path(req.Config)
	if err != nil {
		return err
	}
	q, err := baseQuery(req.Config)
	if err != nil {
		return err
	}
	if spec.name == "search_movies" {
		query := strings.TrimSpace(req.Config.Config["query"])
		if query == "" {
			return errors.New("tmdb search_movies stream requires config query")
		}
		q.Set("query", query)
	}
	if !spec.paginated {
		resp, err := r.Do(ctx, http.MethodGet, path, q, nil)
		if err != nil {
			return safeError("read tmdb", err)
		}
		records, err := connsdk.RecordsAt(resp.Body, spec.recordsPath)
		if err != nil {
			return err
		}
		for _, rec := range records {
			if err := emit(connectors.Record(rec)); err != nil {
				return err
			}
		}
		return nil
	}
	size, err := pageSize(req.Config)
	if err != nil {
		return err
	}
	max, err := maxPages(req.Config)
	if err != nil {
		return err
	}
	p := &connsdk.PageNumberPaginator{PageParam: "page", StartPage: 1, PageSize: size}
	if err := connsdk.Harvest(ctx, r, http.MethodGet, path, q, p, spec.recordsPath, max, func(rec connsdk.Record) error { return emit(connectors.Record(rec)) }); err != nil {
		return safeError("read tmdb", err)
	}
	return nil
}

func (Connector) Write(ctx context.Context, _ connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	if err := ctx.Err(); err != nil {
		return connectors.WriteResult{}, err
	}
	return connectors.WriteResult{RecordsFailed: len(records)}, fmt.Errorf("%s connector is read-only: %w", connectorName, connectors.ErrUnsupportedOperation)
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, UserAgent: userAgent}, nil
}

type streamSpec struct {
	name, description, recordsPath string
	paginated                      bool
	path                           func(connectors.RuntimeConfig) (string, error)
	fields                         []connectors.Field
}

var streamSpecs = map[string]streamSpec{
	"popular_movies":     {name: "popular_movies", description: "Popular movies.", recordsPath: "results", paginated: true, path: staticPath("movie/popular"), fields: fields("id", "title", "overview", "release_date", "vote_average")},
	"now_playing_movies": {name: "now_playing_movies", description: "Now-playing movies.", recordsPath: "results", paginated: true, path: staticPath("movie/now_playing"), fields: fields("id", "title", "overview", "release_date", "vote_average")},
	"search_movies":      {name: "search_movies", description: "Movie search results.", recordsPath: "results", paginated: true, path: staticPath("search/movie"), fields: fields("id", "title", "overview", "release_date", "vote_average")},
	"movie_details":      {name: "movie_details", description: "Details for a configured movie_id.", recordsPath: ".", paginated: false, path: movieDetailPath, fields: fields("id", "title", "overview", "release_date", "runtime")},
}

func streams() []connectors.Stream {
	order := []string{"popular_movies", "now_playing_movies", "search_movies", "movie_details"}
	out := make([]connectors.Stream, 0, len(order))
	for _, name := range order {
		s := streamSpecs[name]
		out = append(out, connectors.Stream{Name: s.name, Description: s.description, Fields: s.fields, PrimaryKey: []string{"id"}})
	}
	return out
}
func fields(names ...string) []connectors.Field {
	out := make([]connectors.Field, 0, len(names))
	for _, name := range names {
		out = append(out, connectors.Field{Name: name, Type: "string"})
	}
	return out
}
func staticPath(path string) func(connectors.RuntimeConfig) (string, error) {
	return func(connectors.RuntimeConfig) (string, error) { return path, nil }
}
func movieDetailPath(cfg connectors.RuntimeConfig) (string, error) {
	id := strings.TrimSpace(cfg.Config["movie_id"])
	if id == "" {
		return "", errors.New("tmdb movie_details stream requires config movie_id")
	}
	return "movie/" + url.PathEscape(id), nil
}
func readFixture(ctx context.Context, spec streamSpec, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{"id": i, "title": fmt.Sprintf("Fixture Movie %d", i), "overview": "Deterministic fixture movie.", "fixture": true}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}
func baseQuery(cfg connectors.RuntimeConfig) (url.Values, error) {
	key := secret(cfg, "api_key")
	if key == "" {
		return nil, errors.New("tmdb connector requires secret api_key")
	}
	q := url.Values{"api_key": {key}}
	if lang := strings.TrimSpace(cfg.Config["language"]); lang != "" {
		q.Set("language", lang)
	}
	return q, nil
}
func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	raw := strings.TrimSpace(cfg.Config["base_url"])
	if raw == "" {
		raw = defaultBaseURL
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid %s base_url", connectorName)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return "", fmt.Errorf("invalid %s base_url scheme %q", connectorName, u.Scheme)
	}
	return strings.TrimRight(raw, "/"), nil
}
func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	value := strings.TrimSpace(cfg.Config["page_size"])
	if value == "" {
		return defaultPageSize, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("%s page_size must be a positive integer", connectorName)
	}
	return n, nil
}
func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	value := strings.TrimSpace(cfg.Config["max_pages"])
	if value == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("%s max_pages must be a non-negative integer", connectorName)
	}
	return n, nil
}
func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets != nil {
		if value := strings.TrimSpace(cfg.Secrets[key]); value != "" {
			return value
		}
	}
	return strings.TrimSpace(cfg.Config[key])
}
func safeError(prefix string, err error) error {
	var httpErr *connsdk.HTTPError
	if errors.As(err, &httpErr) {
		return fmt.Errorf("%s: http %d", prefix, httpErr.Status)
	}
	return fmt.Errorf("%s: %w", prefix, err)
}
