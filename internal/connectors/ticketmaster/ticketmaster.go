// Package ticketmaster implements a read-only native Go connector for the Ticketmaster Discovery API.
package ticketmaster

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
	connectorName   = "ticketmaster"
	defaultBaseURL  = "https://app.ticketmaster.com/discovery/v2"
	defaultPageSize = 200
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Ticketmaster", IntegrationType: "api", Description: "Reads events, venues, attractions, and classifications from the Ticketmaster Discovery API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	q.Set("page", "0")
	q.Set("size", "1")
	if err := r.DoJSON(ctx, http.MethodGet, "classifications.json", q, nil, nil); err != nil {
		return safeError("check ticketmaster", err)
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
		stream = "events"
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
	q, err := baseQuery(req.Config)
	if err != nil {
		return err
	}
	for _, pair := range []struct{ param, key string }{{"keyword", "keyword"}, {"countryCode", "country_code"}, {"locale", "locale"}} {
		if v := strings.TrimSpace(req.Config.Config[pair.key]); v != "" {
			q.Set(pair.param, v)
		}
	}
	size, err := pageSize(req.Config)
	if err != nil {
		return err
	}
	max, err := maxPages(req.Config)
	if err != nil {
		return err
	}
	return harvestPages(ctx, r, spec, q, size, max, emit)
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

func harvestPages(ctx context.Context, r *connsdk.Requester, spec streamSpec, base url.Values, size, max int, emit func(connectors.Record) error) error {
	for page := 0; max == 0 || page < max; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		q := cloneValues(base)
		q.Set("page", strconv.Itoa(page))
		q.Set("size", strconv.Itoa(size))
		resp, err := r.Do(ctx, http.MethodGet, spec.path, q, nil)
		if err != nil {
			return safeError("read ticketmaster", err)
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
		if len(records) < size {
			return nil
		}
	}
	return nil
}

type streamSpec struct {
	name, description, path, recordsPath string
	fields                               []connectors.Field
}

var streamSpecs = map[string]streamSpec{
	"events":          {"events", "Ticketmaster events.", "events.json", "_embedded.events", fields("id", "name", "url", "type", "locale")},
	"venues":          {"venues", "Ticketmaster venues.", "venues.json", "_embedded.venues", fields("id", "name", "url", "city", "country")},
	"attractions":     {"attractions", "Ticketmaster attractions.", "attractions.json", "_embedded.attractions", fields("id", "name", "url", "type", "locale")},
	"classifications": {"classifications", "Ticketmaster classifications.", "classifications.json", "_embedded.classifications", fields("id", "name", "segment", "genre", "subGenre")},
}

func streams() []connectors.Stream {
	order := []string{"events", "venues", "attractions", "classifications"}
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

func readFixture(ctx context.Context, spec streamSpec, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{"id": fmt.Sprintf("%s_fixture_%d", spec.name, i), "name": fmt.Sprintf("Fixture %s %d", spec.name, i), "url": fmt.Sprintf("https://example.com/%s/%d", spec.name, i), "fixture": true}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func baseQuery(cfg connectors.RuntimeConfig) (url.Values, error) {
	key := secret(cfg, "api_key")
	if key == "" {
		return nil, errors.New("ticketmaster connector requires secret api_key")
	}
	return url.Values{"apikey": {key}}, nil
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
func cloneValues(in url.Values) url.Values {
	out := url.Values{}
	for k, vs := range in {
		for _, v := range vs {
			out.Add(k, v)
		}
	}
	return out
}

func safeError(prefix string, err error) error {
	var httpErr *connsdk.HTTPError
	if errors.As(err, &httpErr) {
		return fmt.Errorf("%s: http %d", prefix, httpErr.Status)
	}
	return fmt.Errorf("%s: %w", prefix, err)
}
