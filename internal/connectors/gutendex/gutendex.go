// Package gutendex implements the native pm Gutendex connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit, modelled on
// the stripe reference connector.
//
// Gutendex (https://gutendex.com) is a free, public JSON API over the Project
// Gutenberg book catalog. It requires no authentication, so the connector carries
// no secret; reads are unauthenticated GETs against /books with the standard
// Django REST framework page pagination (count / next / previous / results),
// where "next" is an absolute URL to the following page.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package gutendex

import (
	"context"
	"encoding/json"
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
	gutendexDefaultBaseURL = "https://gutendex.com"
	gutendexUserAgent      = "polymetrics-go-cli"
	// gutendexMaxPagesDefault bounds an unbounded public crawl by default so a
	// naive read does not walk the entire 78k-book catalog; callers can raise it
	// via the max_pages config (or "all"/"unlimited" for no bound).
	gutendexMaxPagesDefault = 3
)

func init() {
	connectors.RegisterFactory("gutendex", New)
}

// New returns the Gutendex connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Gutendex connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "gutendex" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "gutendex",
		DisplayName:     "Gutendex",
		IntegrationType: "api",
		Description:     "Reads Project Gutenberg books from the free, public Gutendex JSON API (books, popular, latest, and English-language views). Read-only; no credentials required.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector can reach Gutendex. The API is public, so there is
// no credential to validate; in fixture mode it short-circuits without a network
// call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := gutendexBaseURL(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the first books page confirms connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "books/", url.Values{"page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check gutendex: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: gutendexStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "books"
	}
	endpoint, ok := gutendexStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("gutendex stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := gutendexMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, req.Config, maxPages, emit)
}

// Write is unsupported: the Gutendex API is read-only (Capabilities.Write is
// false). It satisfies the connectors.Connector interface.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// harvest drives Gutendex's DRF page pagination. List responses are
// {count, next, previous, results:[...]}; "next" is an absolute URL to the
// following page (or null when exhausted). connsdk.Requester treats an absolute
// http(s) path as-is, so we feed "next" straight back in. The loop lives here
// (rather than connsdk.Harvest) so it can stop on a nil next token and honour the
// max_pages bound.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, cfg connectors.RuntimeConfig, maxPages int, emit func(connectors.Record) error) error {
	query := gutendexQuery(endpoint, cfg)
	path := endpoint.resource + "/"

	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read gutendex %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "results")
		if err != nil {
			return fmt.Errorf("decode gutendex %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "next")
		if err != nil {
			return fmt.Errorf("decode gutendex %s next: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		if next == "" {
			return nil
		}
		// The next URL already carries page + filter params; use it verbatim and
		// drop the now-redundant base query.
		path = next
		query = nil
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise gutendex credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	fixtures := []map[string]any{
		{
			"id":             json.Number("2701"),
			"title":          "Moby Dick; Or, The Whale",
			"authors":        []any{map[string]any{"name": "Melville, Herman", "birth_year": json.Number("1819"), "death_year": json.Number("1891")}},
			"translators":    []any{},
			"subjects":       []any{"Sea stories", "Whaling -- Fiction"},
			"bookshelves":    []any{"Best Books Ever Listings"},
			"languages":      []any{"en"},
			"copyright":      false,
			"media_type":     "Text",
			"download_count": json.Number("135445"),
		},
		{
			"id":             json.Number("1342"),
			"title":          "Pride and Prejudice",
			"authors":        []any{map[string]any{"name": "Austen, Jane", "birth_year": json.Number("1775"), "death_year": json.Number("1817")}},
			"translators":    []any{},
			"subjects":       []any{"Love stories", "England -- Fiction"},
			"bookshelves":    []any{"Harvard Classics"},
			"languages":      []any{"en"},
			"copyright":      false,
			"media_type":     "Text",
			"download_count": json.Number("117126"),
		},
	}
	for _, item := range fixtures {
		if err := ctx.Err(); err != nil {
			return err
		}
		record := endpoint.mapRecord(item)
		record["stream"] = stream
		record["fixture"] = true
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester for the Gutendex API. There is no
// authenticator: the API is public.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := gutendexBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		UserAgent: gutendexUserAgent,
	}, nil
}

// gutendexQuery merges the stream's fixed view params with any user-supplied
// filters from config (search, languages, topic, sort, copyright, author year
// range). Stream params win over config so e.g. popular_books always sorts by
// popularity.
func gutendexQuery(endpoint streamEndpoint, cfg connectors.RuntimeConfig) url.Values {
	out := url.Values{}
	for _, key := range []string{"search", "languages", "topic", "sort", "copyright", "author_year_start", "author_year_end", "ids"} {
		if v := strings.TrimSpace(cfg.Config[key]); v != "" {
			out.Set(key, v)
		}
	}
	for k, vs := range endpoint.baseQuery {
		out.Del(k)
		for _, v := range vs {
			out.Add(k, v)
		}
	}
	return out
}

// gutendexBaseURL resolves and validates the base URL. The default is
// gutendex.com; any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func gutendexBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return gutendexDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("gutendex config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("gutendex config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("gutendex config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

// gutendexMaxPages reads the max_pages config bound. Empty defaults to
// gutendexMaxPagesDefault; "all"/"unlimited"/0 mean no bound.
func gutendexMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" {
		return gutendexMaxPagesDefault, nil
	}
	if raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("gutendex config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("gutendex config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
