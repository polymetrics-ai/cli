// Package castoredc implements the native pm Castor EDC connector. It follows
// the declarative-HTTP per-system connector shape pioneered by the stripe
// package: a thin package that composes the connsdk toolkit (Requester +
// OAuth2 client-credentials auth + RecordsAt extraction + HAL page pagination)
// with Castor-specific stream definitions and endpoints.
//
// Castor EDC is a clinical-trial electronic data capture (EDC) platform. Its
// REST API is OAuth2 (client-credentials grant) and returns HAL+JSON: list
// collections live under "_embedded.<key>" and pagination is page-based with
// "page"/"page_count" fields plus "_links.next".
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package castoredc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/connsdk"
)

const (
	registryName          = "castor-edc"
	castorDefaultBaseURL  = "https://data.castoredc.com/api"
	castorDefaultTokenURL = "https://data.castoredc.com/oauth/token"
	castorDefaultPageSize = 100
	castorMaxPageSize     = 1000
	castorUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the Castor EDC connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Castor EDC connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the OAuth2 token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "Castor EDC",
		IntegrationType: "api",
		Description:     "Reads Castor EDC studies, users, countries, and audit-trail events through the Castor EDC OAuth2 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Castor EDC.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := castorBaseURL(cfg); err != nil {
		return err
	}
	clientID, clientSecret := castorCredentials(cfg)
	if strings.TrimSpace(clientID) == "" || strings.TrimSpace(clientSecret) == "" {
		return errors.New("castor-edc connector requires secrets client_id and client_secret")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the user endpoint confirms the OAuth2 exchange and
	// connectivity without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "user", url.Values{"page_size": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check castor-edc: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: castorStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Castor stream starts with
// an empty incremental cursor (full sync).
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
		stream = "study"
	}
	endpoint, ok := castorStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("castor-edc stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := castorPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := castorMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Castor's HAL page pagination. List responses look like
// {"_embedded":{"<key>":[...]},"page":N,"page_count":M,"_links":{"next":{...}}}.
// The loop advances page until the current page reaches page_count (or a page
// returns no records). It is built on connsdk.Requester + connsdk.RecordsAt +
// connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("page_size", strconv.Itoa(pageSize))

	recordsPath := "_embedded." + endpoint.embeddedKey
	page := 1
	for pageNum := 0; maxPages == 0 || pageNum < maxPages; pageNum++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("page", strconv.Itoa(page))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read castor-edc %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, recordsPath)
		if err != nil {
			return fmt.Errorf("decode castor-edc %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}

		// Stop when there are no more records on this page, or when we have
		// reached the last page reported by page_count.
		if len(records) == 0 {
			return nil
		}
		pageCount := pageCountOf(resp.Body)
		if pageCount > 0 {
			if page >= pageCount {
				return nil
			}
		} else if len(records) < pageSize {
			// No page_count reported; fall back to short-page detection.
			return nil
		}
		page++
	}
	return nil
}

// pageCountOf reads the integer "page_count" field from a HAL response body,
// returning 0 when absent or unparseable.
func pageCountOf(body []byte) int {
	raw, err := connsdk.StringAt(body, "page_count")
	if err != nil || strings.TrimSpace(raw) == "" {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0
	}
	return n
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise castor-edc credential-free (mirrors stripe).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"study_id":      fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"crf_id":        fmt.Sprintf("crf_%d", i),
			"id":            fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"uuid":          fmt.Sprintf("uuid_%d", i),
			"user_id":       fmt.Sprintf("user_%d", i),
			"name":          fmt.Sprintf("Fixture Study %d", i),
			"full_name":     fmt.Sprintf("Fixture User %d", i),
			"email_address": fmt.Sprintf("fixture+%d@example.com", i),
			"user_email":    fmt.Sprintf("fixture+%d@example.com", i),
			"user_name":     fmt.Sprintf("Fixture User %d", i),
			"country_name":  "Netherlands",
			"country_cca2":  "NL",
			"event_type":    "Study created",
			"datetime":      "2026-01-01T00:00:00Z",
			"created_on":    "2026-01-01T00:00:00Z",
			"updated_on":    "2026-01-01T00:00:00Z",
			"last_login":    "2026-01-01T00:00:00Z",
			"live":          true,
			"is_active":     true,
			"connector":     registryName,
			"fixture":       true,
		}
		record := endpoint.mapRecord(item)
		record["fixture"] = true
		record["connector"] = registryName
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with OAuth2 client-credentials
// auth and the resolved base URL. The secrets only ever flow into the
// connsdk.OAuth2ClientCredentials authenticator; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := castorBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	tokenURL, err := castorTokenURL(cfg)
	if err != nil {
		return nil, err
	}
	clientID, clientSecret := castorCredentials(cfg)
	if strings.TrimSpace(clientID) == "" || strings.TrimSpace(clientSecret) == "" {
		return nil, errors.New("castor-edc connector requires secrets client_id and client_secret")
	}
	auth := &connsdk.OAuth2ClientCredentials{
		TokenURL:     tokenURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Client:       c.Client,
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: castorUserAgent,
	}, nil
}

// Write is unsupported: Castor EDC is a clinical-trial system of record and
// this connector is intentionally read-only.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func castorCredentials(cfg connectors.RuntimeConfig) (string, string) {
	if cfg.Secrets == nil {
		return "", ""
	}
	return cfg.Secrets["client_id"], cfg.Secrets["client_secret"]
}

// castorBaseURL resolves and validates the base URL. The default is
// data.castoredc.com; a url_region config (uk/nl/us) selects the regional host,
// and an explicit base_url override wins over both. Any URL must be an absolute
// http(s) URL with a host to bound SSRF risk.
func castorBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		if region := strings.TrimSpace(strings.ToLower(cfg.Config["url_region"])); region != "" {
			base = fmt.Sprintf("https://%s.castoredc.com/api", region)
		} else {
			base = castorDefaultBaseURL
		}
	}
	if err := validateURL(base, "base_url"); err != nil {
		return "", err
	}
	return strings.TrimRight(base, "/"), nil
}

// castorTokenURL resolves the OAuth2 token endpoint. It defaults to the same
// host as the resolved base URL (data.castoredc.com or the regional host) and
// may be overridden with the token_url config.
func castorTokenURL(cfg connectors.RuntimeConfig) (string, error) {
	tokenURL := strings.TrimSpace(cfg.Config["token_url"])
	if tokenURL == "" {
		base, err := castorBaseURL(cfg)
		if err != nil {
			return "", err
		}
		parsed, err := url.Parse(base)
		if err != nil {
			return "", fmt.Errorf("castor-edc derive token_url: %w", err)
		}
		tokenURL = fmt.Sprintf("%s://%s/oauth/token", parsed.Scheme, parsed.Host)
	}
	if err := validateURL(tokenURL, "token_url"); err != nil {
		return "", err
	}
	return tokenURL, nil
}

func validateURL(raw, field string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("castor-edc config %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return fmt.Errorf("castor-edc config %s must use http or https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("castor-edc config %s must include a host", field)
	}
	return nil
}

func castorPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return castorDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("castor-edc config page_size must be an integer: %w", err)
	}
	if value < 1 || value > castorMaxPageSize {
		return 0, fmt.Errorf("castor-edc config page_size must be between 1 and %d", castorMaxPageSize)
	}
	return value, nil
}

func castorMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("castor-edc config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("castor-edc config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
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
