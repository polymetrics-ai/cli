// Package mailjetmail implements the native pm Mailjet Mail connector. It is a
// declarative-HTTP per-system connector following the stripe reference shape: a
// thin package that composes the connsdk toolkit (Requester + HTTP Basic auth +
// RecordsAt extraction + offset pagination) with Mailjet-specific stream
// definitions and endpoints.
//
// Mailjet's Email REST API (v3) authenticates with HTTP Basic where the username
// is the public API key and the password is the secret API key. List endpoints
// return a {Count, Total, Data:[...]} envelope and paginate with Limit/Offset
// query params. The API supports full-refresh reads only, so this connector is
// read-only.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package mailjetmail

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
	mailjetDefaultBaseURL  = "https://api.mailjet.com/v3/REST"
	mailjetDefaultPageSize = 100
	mailjetMaxPageSize     = 1000
	mailjetUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("mailjet-mail", New)
}

// New returns the Mailjet Mail connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Mailjet Mail connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "mailjet-mail" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "mailjet-mail",
		DisplayName:     "Mailjet Mail",
		IntegrationType: "api",
		Description:     "Reads Mailjet contacts, contact lists, messages, campaigns, and statistics through the Mailjet Email REST API (v3).",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Mailjet. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := mailjetBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(mailjetAPIKey(cfg)) == "" {
		return errors.New("mailjet-mail connector requires config api_key")
	}
	if strings.TrimSpace(mailjetSecret(cfg)) == "" {
		return errors.New("mailjet-mail connector requires secret api_key_secret")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the contact list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "contact", url.Values{"Limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check mailjet-mail: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: mailjetStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "contacts"
	}
	endpoint, ok := mailjetStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("mailjet-mail stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := mailjetPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := mailjetMaxPages(req.Config)
	if err != nil {
		return err
	}

	// Mailjet lists return {Count, Total, Data:[...]} and paginate with
	// Limit/Offset. OffsetPaginator advances by PageSize until a short page is
	// returned; records live under the "Data" key.
	base := url.Values{}
	paginator := &connsdk.OffsetPaginator{
		LimitParam:  "Limit",
		OffsetParam: "Offset",
		PageSize:    pageSize,
	}
	mapped := func(rec connsdk.Record) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		return emit(endpoint.mapRecord(rec))
	}
	if err := connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, base, paginator, "Data", maxPages, mapped); err != nil {
		return fmt.Errorf("read mailjet-mail %s: %w", endpoint.resource, err)
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise mailjet-mail credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"ID":                       int64(1000 + i),
			"Email":                    fmt.Sprintf("fixture+%d@example.com", i),
			"Name":                     fmt.Sprintf("Fixture %d", i),
			"Address":                  fmt.Sprintf("list-%d", i),
			"CreatedAt":                "2026-01-01T00:00:00Z",
			"LastActivityAt":           "2026-01-01T00:00:00Z",
			"LastUpdateAt":             "2026-01-01T00:00:00Z",
			"ArrivedAt":                "2026-01-01T00:00:00Z",
			"SendStartAt":              "2026-01-01T00:00:00Z",
			"DeliveredCount":           int64(i),
			"SubscriberCount":          int64(10 * i),
			"ContactID":                int64(2000 + i),
			"CampaignID":               int64(3000 + i),
			"Status":                   "active",
			"Subject":                  fmt.Sprintf("Fixture Campaign %d", i),
			"FromEmail":                "sender@example.com",
			"FromName":                 "Sender",
			"AttemptCount":             int64(1),
			"MessageSize":              int64(1024 * i),
			"IsExcludedFromCampaigns":  false,
			"IsOptInPending":           false,
			"IsSpamComplaining":        false,
			"IsDeleted":                false,
			"IsStarred":                false,
			"IsClickTracked":           true,
			"IsOpenTracked":            true,
			"MessageSentCount":         int64(100 * i),
			"MessageOpenedCount":       int64(50 * i),
			"MessageClickedCount":      int64(20 * i),
			"MessageDeliveredCount":    int64(95 * i),
			"MessageBouncedCount":      int64(5 * i),
			"MessageSpamCount":         int64(1),
			"MessageUnsubscribedCount": int64(2),
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with HTTP Basic auth (api_key as
// username, api_key_secret as password) and the resolved base URL. The secret
// only ever flows into connsdk.Basic; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := mailjetBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	apiKey := mailjetAPIKey(cfg)
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("mailjet-mail connector requires config api_key")
	}
	secret := mailjetSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("mailjet-mail connector requires secret api_key_secret")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(apiKey, secret),
		UserAgent: mailjetUserAgent,
	}, nil
}

func mailjetAPIKey(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return ""
	}
	return cfg.Config["api_key"]
}

func mailjetSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key_secret"]
}

// mailjetBaseURL resolves and validates the base URL. The default is
// api.mailjet.com/v3/REST; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func mailjetBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return mailjetDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("mailjet-mail config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("mailjet-mail config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("mailjet-mail config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func mailjetPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return mailjetDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("mailjet-mail config page_size must be an integer: %w", err)
	}
	if value < 1 || value > mailjetMaxPageSize {
		return 0, fmt.Errorf("mailjet-mail config page_size must be between 1 and %d", mailjetMaxPageSize)
	}
	return value, nil
}

func mailjetMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("mailjet-mail config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("mailjet-mail config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: the Mailjet Mail source is read-only.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
