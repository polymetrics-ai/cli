// Package mailjetsms implements the native pm Mailjet SMS connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit: a
// connsdk.Requester wired with Bearer auth, offset-based pagination over the
// Mailjet SMS API, and Data[]-array record extraction.
//
// The Mailjet SMS API (https://dev.mailjet.com/sms/) exposes outbound SMS
// messages and a count endpoint, both full-refresh only. The connector is
// read-only.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect. Note the directory/registry
// name keeps the hyphen ("mailjet-sms") while the Go package identifier is the
// sanitized "mailjetsms".
package mailjetsms

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
	mailjetDefaultBaseURL  = "https://api.mailjet.com/v4"
	mailjetDefaultPageSize = 100
	mailjetMaxPageSize     = 1000
	mailjetUserAgent       = "polymetrics-go-cli"
	// mailjetFixtureCreated is the deterministic CreationTS used by fixture-mode
	// records (2026-01-01T00:00:00Z in unix seconds).
	mailjetFixtureCreated int64 = 1767225600
)

func init() {
	connectors.RegisterFactory("mailjet-sms", New)
}

// New returns the Mailjet SMS connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Mailjet SMS connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "mailjet-sms" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "mailjet-sms",
		DisplayName:     "Mailjet SMS",
		IntegrationType: "api",
		Description:     "Reads outbound SMS messages and SMS counts from the Mailjet SMS API (full refresh, read-only).",
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
	if strings.TrimSpace(mailjetSecret(cfg)) == "" {
		return errors.New("mailjet-sms connector requires secret token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the SMS count endpoint confirms auth and connectivity
	// without listing any records.
	if err := r.DoJSON(ctx, http.MethodGet, "sms/count", nil, nil, nil); err != nil {
		return fmt.Errorf("check mailjet-sms: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "sms"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("mailjet-sms stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}

	base := url.Values{}
	if from := strings.TrimSpace(req.Config.Config["start_date"]); from != "" {
		base.Set("FromTS", from)
	}
	if to := strings.TrimSpace(req.Config.Config["end_date"]); to != "" {
		base.Set("ToTS", to)
	}

	// Single-object endpoints (sms/count) are a single bounded GET.
	if !endpoint.paginated {
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, base, nil)
		if err != nil {
			return fmt.Errorf("read mailjet-sms %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode mailjet-sms %s: %w", endpoint.resource, err)
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

	pageSize, err := mailjetPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := mailjetMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, base, pageSize, maxPages, emit)
}

// Write is unsupported: the Mailjet SMS API is read-only for this connector.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// harvest drives Mailjet's offset/limit pagination. SMS lists return
// {"Data":[...],"Count":N}; pages are requested with Limit + an Offset that
// advances by Limit until a page returns fewer than Limit records.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, base url.Values, pageSize, maxPages int, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("Limit", strconv.Itoa(pageSize))
		query.Set("Offset", strconv.Itoa(offset))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read mailjet-sms %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode mailjet-sms %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short page (fewer than the requested Limit) means we have reached the
		// end of the result set.
		if len(records) < pageSize {
			return nil
		}
		offset += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise mailjet-sms credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if stream == "sms_count" {
		return emit(endpoint.mapRecord(map[string]any{"Count": int64(2)}))
	}
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"ID":         fmt.Sprintf("sms_fixture_%d", i),
			"From":       "MAILJET",
			"To":         fmt.Sprintf("+1555000%04d", i),
			"MessageId":  fmt.Sprintf("msg_fixture_%d", i),
			"CreationTS": mailjetFixtureCreated + int64(i),
			"SentTS":     mailjetFixtureCreated + int64(i) + 1,
			"SMSCount":   int64(1),
			"Status":     map[string]any{"Code": int64(2), "Name": "sent", "Description": "Message sent"},
			"Cost":       map[string]any{"Value": 0.05, "Currency": "EUR"},
		}
		record := endpoint.mapRecord(item)
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := mailjetBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := mailjetSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("mailjet-sms connector requires secret token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: mailjetUserAgent,
	}, nil
}

func mailjetSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["token"]
}

// mailjetBaseURL resolves and validates the base URL. The default is
// api.mailjet.com/v4; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func mailjetBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return mailjetDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("mailjet-sms config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("mailjet-sms config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("mailjet-sms config base_url must include a host")
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
		return 0, fmt.Errorf("mailjet-sms config page_size must be an integer: %w", err)
	}
	if value < 1 || value > mailjetMaxPageSize {
		return 0, fmt.Errorf("mailjet-sms config page_size must be between 1 and %d", mailjetMaxPageSize)
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
		return 0, fmt.Errorf("mailjet-sms config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("mailjet-sms config max_pages must be 0 for unlimited or a positive integer")
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
