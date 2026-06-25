// Package activecampaign implements the native pm ActiveCampaign connector. It
// is a declarative-HTTP per-system connector built on the connsdk toolkit
// (Requester + Api-Token header auth + RecordsAt extraction) with
// ActiveCampaign-specific stream definitions and limit/offset pagination,
// following the shape of the stripe reference connector.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// ActiveCampaign API v3 reference: base URL https://<account>.api-us1.com/api/3,
// authentication via the Api-Token request header, and limit/offset pagination.
package activecampaign

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
	// acDefaultPageSize is ActiveCampaign's default list page size; acMaxPageSize
	// is the documented maximum.
	acDefaultPageSize = 20
	acMaxPageSize     = 100
	acUserAgent       = "polymetrics-go-cli"
	// acAuthHeader is the ActiveCampaign API token header.
	acAuthHeader = "Api-Token"
	// acFixtureCDate is the deterministic creation timestamp used by fixture-mode
	// records.
	acFixtureCDate = "2026-01-01T00:00:00-05:00"
)

func init() {
	connectors.RegisterFactory("activecampaign", New)
}

// New returns the ActiveCampaign connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm ActiveCampaign connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "activecampaign" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "activecampaign",
		DisplayName:     "ActiveCampaign",
		IntegrationType: "api",
		Description:     "Reads ActiveCampaign contacts, lists, deals, and campaigns through the ActiveCampaign v3 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true},
	}
}

// Check verifies the connector is configured well enough to talk to
// ActiveCampaign. In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := acBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(acSecret(cfg)) == "" {
		return errors.New("activecampaign connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the contacts list confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "contacts", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check activecampaign: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams()}, nil
}

// Write satisfies the connectors.Connector interface. ActiveCampaign is
// exposed read-only here, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader. ActiveCampaign only supports
// full-refresh syncs, so a stream starts with an empty cursor.
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
		stream = "contacts"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("activecampaign stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := acPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := acMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives ActiveCampaign's limit/offset pagination. List responses wrap
// the records array under a resource-named key (e.g. {"contacts":[...]}). A page
// shorter than pageSize signals the end. The loop lives here, built on
// connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(offset))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read activecampaign %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsKey)
		if err != nil {
			return fmt.Errorf("decode activecampaign %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short page means we have read everything.
		if len(records) < pageSize {
			return nil
		}
		offset += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise activecampaign credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":               strconv.Itoa(i),
			"email":            fmt.Sprintf("fixture+%d@example.com", i),
			"firstName":        fmt.Sprintf("Fixture%d", i),
			"lastName":         "Example",
			"phone":            "",
			"name":             fmt.Sprintf("%s fixture %d", stream, i),
			"stringid":         fmt.Sprintf("%s-fixture-%d", stream, i),
			"sender_url":       "https://example.com",
			"subscriber_count": strconv.Itoa(i * 10),
			"title":            fmt.Sprintf("Deal %d", i),
			"contact":          "1",
			"value":            strconv.Itoa(i * 1000),
			"currency":         "usd",
			"status":           "0",
			"stage":            "1",
			"owner":            "1",
			"type":             "single",
			"subject":          fmt.Sprintf("Subject %d", i),
			"send_amt":         strconv.Itoa(i * 100),
			"opens":            strconv.Itoa(i * 5),
			"uniqueopens":      strconv.Itoa(i * 4),
			"linkclicks":       strconv.Itoa(i),
			"orgid":            "",
			"deleted":          "0",
			"userid":           "1",
			"cdate":            acFixtureCDate,
			"udate":            acFixtureCDate,
			"mdate":            acFixtureCDate,
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

// requester builds a connsdk.Requester wired with Api-Token header auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := acBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := acSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("activecampaign connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(acAuthHeader, secret, ""),
		UserAgent: acUserAgent,
	}, nil
}

func acSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// acBaseURL resolves and validates the base URL. The default is derived from the
// account_username config as https://<account>.api-us1.com/api/3. Any base_url
// override must be an absolute https (or http for local test servers) URL with a
// host to bound SSRF risk.
func acBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		account := strings.TrimSpace(cfg.Config["account_username"])
		if account == "" {
			return "", errors.New("activecampaign config requires account_username or base_url")
		}
		if !validAccount(account) {
			return "", fmt.Errorf("activecampaign config account_username %q is invalid", account)
		}
		return fmt.Sprintf("https://%s.api-us1.com/api/3", account), nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("activecampaign config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("activecampaign config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("activecampaign config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

// validAccount restricts account_username to a conservative subdomain-safe set
// so it cannot inject extra host/path segments into the derived base URL.
func validAccount(account string) bool {
	for _, r := range account {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_':
		default:
			return false
		}
	}
	return account != ""
}

func acPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return acDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("activecampaign config page_size must be an integer: %w", err)
	}
	if value < 1 || value > acMaxPageSize {
		return 0, fmt.Errorf("activecampaign config page_size must be between 1 and %d", acMaxPageSize)
	}
	return value, nil
}

func acMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("activecampaign config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("activecampaign config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
