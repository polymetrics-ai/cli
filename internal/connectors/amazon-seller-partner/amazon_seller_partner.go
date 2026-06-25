// Package amazonsellerpartner implements the native pm Amazon Selling Partner
// API (SP-API) connector. It follows the stripe declarative-HTTP template: a
// thin package composing the connsdk toolkit (Requester + retry + RecordsAt /
// StringAt extraction + cursor state) with SP-API specific stream definitions,
// endpoints, and a Login with Amazon (LWA) token authenticator.
//
// SP-API auth differs from a plain bearer flow: the connector exchanges a
// long-lived LWA refresh_token for a short-lived access_token at the LWA token
// endpoint, then passes that token on every data request via the
// x-amz-access-token header (not Authorization: Bearer). That exchange lives in
// lwaAuthenticator below.
//
// The package lives in directory amazon-seller-partner (a hyphen is not a legal
// Go identifier, so the package name is amazonsellerpartner) and self-registers
// under the registry key "amazon-seller-partner" via RegisterFactory in init().
// The registryset package blank-imports this directory in the production binary
// to run that side effect.
package amazonsellerpartner

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	connectorName = "amazon-seller-partner"

	// spDefaultBaseURL is the North America SP-API endpoint. Sellers in other
	// regions override base_url (or set region) to the matching endpoint.
	spDefaultBaseURL = "https://sellingpartnerapi-na.amazon.com"
	// lwaDefaultTokenURL is the Login with Amazon token exchange endpoint.
	lwaDefaultTokenURL = "https://api.amazon.com/auth/o2/token"
	// defaultMarketplaceID is the US marketplace, used when marketplace_id is
	// unset. Required by Orders and FBA Inventory.
	defaultMarketplaceID = "ATVPDKIKX0DER"

	spDefaultPageSize = 100
	spMaxPageSize     = 100
	spUserAgent       = "polymetrics-go-cli"

	// spFixtureUpdated is the deterministic update timestamp used by fixture-mode
	// records (2026-01-01T00:00:00Z).
	spFixtureUpdated = "2026-01-01T00:00:00Z"
)

// spRegionBaseURLs maps the configured AWS region to its SP-API endpoint. The
// region keys mirror the catalog's enum; everything else defaults to NA.
var spRegionBaseURLs = map[string]string{
	"NA": "https://sellingpartnerapi-na.amazon.com",
	"US": "https://sellingpartnerapi-na.amazon.com",
	"CA": "https://sellingpartnerapi-na.amazon.com",
	"MX": "https://sellingpartnerapi-na.amazon.com",
	"BR": "https://sellingpartnerapi-na.amazon.com",
	"EU": "https://sellingpartnerapi-eu.amazon.com",
	"GB": "https://sellingpartnerapi-eu.amazon.com",
	"UK": "https://sellingpartnerapi-eu.amazon.com",
	"DE": "https://sellingpartnerapi-eu.amazon.com",
	"FR": "https://sellingpartnerapi-eu.amazon.com",
	"ES": "https://sellingpartnerapi-eu.amazon.com",
	"IT": "https://sellingpartnerapi-eu.amazon.com",
	"NL": "https://sellingpartnerapi-eu.amazon.com",
	"SE": "https://sellingpartnerapi-eu.amazon.com",
	"PL": "https://sellingpartnerapi-eu.amazon.com",
	"BE": "https://sellingpartnerapi-eu.amazon.com",
	"TR": "https://sellingpartnerapi-eu.amazon.com",
	"AE": "https://sellingpartnerapi-eu.amazon.com",
	"SA": "https://sellingpartnerapi-eu.amazon.com",
	"EG": "https://sellingpartnerapi-eu.amazon.com",
	"IN": "https://sellingpartnerapi-eu.amazon.com",
	"JP": "https://sellingpartnerapi-fe.amazon.com",
	"AU": "https://sellingpartnerapi-fe.amazon.com",
	"SG": "https://sellingpartnerapi-fe.amazon.com",
}

func init() {
	connectors.RegisterFactory(connectorName, New)
}

// New returns the Amazon Seller Partner connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Amazon Selling Partner API connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the LWA token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            connectorName,
		DisplayName:     "Amazon Seller Partner",
		IntegrationType: "api",
		Description:     "Reads Amazon Selling Partner API orders, order items, FBA inventory summaries, and financial event groups via Login with Amazon (LWA) authentication.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to SP-API. In
// fixture mode it short-circuits without a network call. Otherwise it confirms
// the base URL and credentials, then exercises the LWA token exchange and a
// bounded orders read to validate auth and connectivity without mutating data.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := spBaseURL(cfg); err != nil {
		return err
	}
	if err := requireSecrets(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	query := url.Values{}
	query.Set("MarketplaceIds", marketplaceID(cfg))
	query.Set("CreatedAfter", spDefaultCreatedAfter())
	query.Set("MaxResultsPerPage", "1")
	if err := r.DoJSON(ctx, http.MethodGet, spStreamEndpoints["orders"].resource, query, nil, nil); err != nil {
		return fmt.Errorf("check amazon-seller-partner: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: spStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a stream starts with an
// empty incremental cursor (full sync), which replication_start_date can raise
// at read time.
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
		stream = "orders"
	}
	endpoint, ok := spStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("amazon-seller-partner stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	if _, err := spBaseURL(req.Config); err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := spMaxPages(req.Config)
	if err != nil {
		return err
	}
	base, err := endpoint.baseQuery(req)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, base, maxPages, emit)
}

// harvest drives SP-API's NextToken pagination. SP-API list responses carry the
// next-page token in the body (at endpoint.tokenPath) and the caller resupplies
// it via the endpoint.tokenParam query parameter on the following request. The
// records array lives at endpoint.recordsPath. This shape has no exact connsdk
// paginator, so the loop lives here, built on connsdk.Requester + RecordsAt +
// StringAt (mirroring the stripe harvest).
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, base url.Values, maxPages int, emit func(connectors.Record) error) error {
	nextToken := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if nextToken != "" {
			// When paging, SP-API requires only the token (other filters must be
			// dropped); resend just the token param.
			query = url.Values{}
			query.Set(endpoint.tokenParam, nextToken)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read amazon-seller-partner %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode amazon-seller-partner %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		token, err := connsdk.StringAt(resp.Body, endpoint.tokenPath)
		if err != nil {
			return fmt.Errorf("decode amazon-seller-partner %s next token: %w", endpoint.resource, err)
		}
		if strings.TrimSpace(token) == "" {
			return nil
		}
		nextToken = token
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free (mirrors the
// stripe fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		record := endpoint.mapRecord(endpoint.fixtureItem(i))
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the LWA authenticator and the
// resolved base URL. The secrets only ever flow into the authenticator; they are
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := spBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	if err := requireSecrets(cfg); err != nil {
		return nil, err
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      c.authenticator(cfg),
		UserAgent: spUserAgent,
	}, nil
}

// authenticator builds the LWA token authenticator from config + secrets.
func (c Connector) authenticator(cfg connectors.RuntimeConfig) *lwaAuthenticator {
	return &lwaAuthenticator{
		tokenURL:     lwaTokenURL(cfg),
		clientID:     secret(cfg, "lwa_app_id"),
		clientSecret: secret(cfg, "lwa_client_secret"),
		refreshToken: secret(cfg, "refresh_token"),
		client:       c.Client,
	}
}

// requireSecrets confirms the mandatory LWA secrets are present. It never logs
// their values.
func requireSecrets(cfg connectors.RuntimeConfig) error {
	for _, field := range []string{"lwa_app_id", "lwa_client_secret", "refresh_token"} {
		if strings.TrimSpace(secret(cfg, field)) == "" {
			return fmt.Errorf("amazon-seller-partner connector requires secret %s", field)
		}
	}
	return nil
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

// spBaseURL resolves and validates the base URL. The default is the NA SP-API
// endpoint; region selects a regional endpoint; any explicit base_url override
// must be an absolute https (or http for local test servers) URL with a host to
// bound SSRF risk.
func spBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		if region := strings.ToUpper(strings.TrimSpace(cfg.Config["region"])); region != "" {
			if endpoint, ok := spRegionBaseURLs[region]; ok {
				return endpoint, nil
			}
		}
		return spDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("amazon-seller-partner config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("amazon-seller-partner config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("amazon-seller-partner config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func lwaTokenURL(cfg connectors.RuntimeConfig) string {
	if v := strings.TrimSpace(cfg.Config["lwa_token_url"]); v != "" {
		return v
	}
	return lwaDefaultTokenURL
}

func marketplaceID(cfg connectors.RuntimeConfig) string {
	if v := strings.TrimSpace(cfg.Config["marketplace_id"]); v != "" {
		return v
	}
	return defaultMarketplaceID
}

// spReplicationStart returns the RFC3339 lower bound for time-windowed streams,
// derived from the incremental cursor (if any) or else replication_start_date.
// An empty result means the endpoint default (two years ago) should be used.
func spReplicationStart(req connectors.ReadRequest) (string, error) {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor, nil
	}
	start := strings.TrimSpace(req.Config.Config["replication_start_date"])
	if start == "" {
		return "", nil
	}
	if _, err := time.Parse(time.RFC3339, start); err != nil {
		return "", fmt.Errorf("amazon-seller-partner config replication_start_date must be RFC3339: %w", err)
	}
	return start, nil
}

// spDefaultCreatedAfter returns a conservative two-years-ago RFC3339 timestamp,
// matching SP-API's own default window for orders.
func spDefaultCreatedAfter() string {
	return time.Now().UTC().AddDate(-2, 0, 0).Format(time.RFC3339)
}

func spMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("amazon-seller-partner config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("amazon-seller-partner config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func spPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return spDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("amazon-seller-partner config page_size must be an integer: %w", err)
	}
	if value < 1 || value > spMaxPageSize {
		return 0, fmt.Errorf("amazon-seller-partner config page_size must be between 1 and %d", spMaxPageSize)
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
