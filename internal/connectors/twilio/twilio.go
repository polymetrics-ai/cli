// Package twilio implements the native pm Twilio connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit
// (Requester + HTTP Basic auth + RecordsAt extraction + next_page_uri
// pagination) plus Twilio-specific stream definitions and endpoints. It copies
// the shape of the stripe reference connector.
//
// Twilio's REST API (2010-04-01) authenticates with HTTP Basic where the
// username is the Account SID and the password is the Auth Token. List
// endpoints are account-scoped (/Accounts/{sid}/<Resource>.json) and return a
// JSON object holding the records array under a resource-named key plus a
// host-relative "next_page_uri" used to walk subsequent pages.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect. The connector is read-only:
// Twilio writes (sending messages, placing calls) are side-effecting actions
// inappropriate for a generic reverse-ETL source, so Capabilities.Write=false.
package twilio

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
	twilioDefaultBaseURL  = "https://api.twilio.com/2010-04-01"
	twilioDefaultPageSize = 50
	twilioMaxPageSize     = 1000
	twilioUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("twilio", New)
}

// New returns the Twilio connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Twilio connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "twilio" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "twilio",
		DisplayName:     "Twilio",
		IntegrationType: "api",
		Description:     "Reads Twilio messages, calls, recordings, conferences, and usage records through the Twilio REST API (2010-04-01).",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Twilio. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := twilioBaseURL(cfg); err != nil {
		return err
	}
	sid, token := twilioSecrets(cfg)
	if strings.TrimSpace(sid) == "" || strings.TrimSpace(token) == "" {
		return errors.New("twilio connector requires secrets account_sid and auth_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the account's Messages list confirms auth and
	// connectivity without mutating anything.
	path := accountPath(sid, "Messages.json")
	if err := r.DoJSON(ctx, http.MethodGet, path, url.Values{"PageSize": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check twilio: %w", err)
	}
	return nil
}

// Write is unsupported: Twilio is a read-only source connector. Sending
// messages or placing calls are side-effecting actions inappropriate for a
// generic reverse-ETL destination, so the connector reports no Write capability
// and rejects writes here.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: twilioStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Twilio stream starts with
// an empty incremental cursor (full sync) which the start_date config can raise.
func (c Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return connsdk.WithCursor(map[string]string{"stream": stream}, ""), nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "messages"
	}
	endpoint, ok := twilioStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("twilio stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	base, err := twilioBaseURL(req.Config)
	if err != nil {
		return err
	}
	sid, _ := twilioSecrets(req.Config)
	pageSize, err := twilioPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := twilioMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, base, sid, endpoint, pageSize, maxPages, emit)
}

// harvest drives Twilio's next_page_uri pagination. A list response is shaped
// {"<recordsKey>":[...], "next_page_uri":"/2010-04-01/...?Page=N&..."}; the
// next page is fetched by following next_page_uri (host-relative) until it is
// null/empty. The loop lives here, built on connsdk.Requester + RecordsAt +
// StringAt, because Twilio's body-cursor shape has no exact connsdk paginator.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, base, sid string, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	// First page: account-scoped resource path with a PageSize.
	nextURL, err := absoluteURL(base, accountPath(sid, endpoint.resource))
	if err != nil {
		return err
	}
	firstQuery := url.Values{"PageSize": []string{strconv.Itoa(pageSize)}}

	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var query url.Values
		if page == 0 {
			query = firstQuery
		}
		resp, err := r.Do(ctx, http.MethodGet, nextURL, query, nil)
		if err != nil {
			return fmt.Errorf("read twilio %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsKey)
		if err != nil {
			return fmt.Errorf("decode twilio %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "next_page_uri")
		if err != nil {
			return fmt.Errorf("decode twilio %s next_page_uri: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		if next == "" || next == "null" {
			return nil
		}
		// next_page_uri is host-relative; resolve it against the base host.
		nextURL, err = absoluteURL(base, next)
		if err != nil {
			return err
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise twilio credential-free (mirrors stripe).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"sid":           fmt.Sprintf("FX%s%d", strings.ToUpper(stream[:1]), i),
			"account_sid":   "AC_fixture",
			"category":      fmt.Sprintf("%s-%d", stream, i),
			"date_created":  "Mon, 01 Jan 2024 00:00:00 +0000",
			"date_updated":  "Mon, 01 Jan 2024 00:00:00 +0000",
			"date_sent":     "Mon, 01 Jan 2024 00:00:00 +0000",
			"start_time":    "Mon, 01 Jan 2024 00:00:00 +0000",
			"start_date":    "2024-01-01",
			"end_date":      "2024-01-31",
			"from":          "+15550000001",
			"to":            fmt.Sprintf("+1555000100%d", i),
			"body":          fmt.Sprintf("fixture message %d", i),
			"status":        "delivered",
			"direction":     "outbound-api",
			"duration":      strconv.Itoa(30 * i),
			"price":         "-0.0075",
			"price_unit":    "USD",
			"friendly_name": fmt.Sprintf("fixture-%d", i),
			"channels":      1,
			"connector":     "twilio",
			"fixture":       true,
		}
		record := endpoint.mapRecord(item)
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with HTTP Basic auth and the
// resolved base URL. The Account SID and Auth Token only ever flow into
// connsdk.Basic; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := twilioBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	sid, token := twilioSecrets(cfg)
	if strings.TrimSpace(sid) == "" || strings.TrimSpace(token) == "" {
		return nil, errors.New("twilio connector requires secrets account_sid and auth_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(sid, token),
		UserAgent: twilioUserAgent,
	}, nil
}

func twilioSecrets(cfg connectors.RuntimeConfig) (sid, token string) {
	if cfg.Secrets == nil {
		return "", ""
	}
	return cfg.Secrets["account_sid"], cfg.Secrets["auth_token"]
}

// accountPath builds an account-scoped resource path relative to the base URL,
// e.g. accountPath("ACxxx", "Messages.json") => "Accounts/ACxxx/Messages.json".
func accountPath(sid, resource string) string {
	return "Accounts/" + url.PathEscape(strings.TrimSpace(sid)) + "/" + resource
}

// absoluteURL resolves a possibly-relative reference (e.g. a host-relative
// next_page_uri, or an account-scoped path segment) against the base URL's host
// and scheme. Twilio's next_page_uri already includes the /2010-04-01 prefix,
// so it is resolved against the host root rather than appended to base.
func absoluteURL(base, ref string) (string, error) {
	b, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("twilio base_url is invalid: %w", err)
	}
	r, err := url.Parse(ref)
	if err != nil {
		return "", fmt.Errorf("twilio path is invalid: %w", err)
	}
	if r.IsAbs() {
		return r.String(), nil
	}
	if strings.HasPrefix(ref, "/") {
		// Host-relative: keep scheme+host from base, replace path+query.
		resolved := *b
		resolved.Path = r.Path
		resolved.RawQuery = r.RawQuery
		return resolved.String(), nil
	}
	// Relative to the base path (e.g. "Accounts/.../Messages.json").
	basePath := strings.TrimRight(b.Path, "/")
	resolved := *b
	resolved.Path = basePath + "/" + ref
	resolved.RawQuery = r.RawQuery
	return resolved.String(), nil
}

// twilioBaseURL resolves and validates the base URL. The default is
// api.twilio.com/2010-04-01; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func twilioBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return twilioDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("twilio config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("twilio config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("twilio config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func twilioPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return twilioDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("twilio config page_size must be an integer: %w", err)
	}
	if value < 1 || value > twilioMaxPageSize {
		return 0, fmt.Errorf("twilio config page_size must be between 1 and %d", twilioMaxPageSize)
	}
	return value, nil
}

func twilioMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("twilio config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("twilio config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
