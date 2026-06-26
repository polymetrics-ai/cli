// Package deputy implements the native pm Deputy connector. It is a declarative-
// HTTP per-system connector built on the connsdk toolkit (Requester + Bearer auth
// + RecordsAt extraction) wired to Deputy-specific stream definitions and
// endpoints. It follows the stripe reference shape.
//
// Deputy is a workforce-management product (scheduling, timesheets, HR). Its API
// is read-only here: full-refresh streams keyed by the integer Id. The connector
// self-registers with the connectors registry via RegisterFactory in init(); the
// registryset package blank-imports this package in the production binary to run
// that side effect.
//
// Auth is a bearer access token (config secret api_key). The base URL is
// install-specific (https://{installname}.{geo}.deputy.com) and therefore
// required — there is no shared default host.
package deputy

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
	deputyDefaultPageSize = 500
	deputyMaxPageSize     = 500
	deputyUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("deputy", New)
}

// New returns the Deputy connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Deputy connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "deputy" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "deputy",
		DisplayName:     "Deputy",
		IntegrationType: "api",
		Description:     "Reads Deputy locations, employees, departments, timesheets, and tasks through the Deputy REST API (read-only, full refresh).",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Deputy. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := deputyBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(deputySecret(cfg)) == "" {
		return errors.New("deputy connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the locations endpoint confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "api/v1/resource/Company", url.Values{"max": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check deputy: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: deputyStreams()}, nil
}

// Write satisfies the connectors.Connector interface. Deputy is a read-only
// source here, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "locations"
	}
	endpoint, ok := deputyStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("deputy stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := deputyPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := deputyMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Deputy reads. Deputy resource endpoints return a bare top-level
// JSON array and accept ?start=N / ?max=N offset pagination; the loop advances
// start by the page size and stops on a short page. Non-paginated curated
// endpoints (my/*, supervise/*) return a single bounded page, so the loop exits
// after the first request.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	start := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		if endpoint.paginated {
			query.Set("start", strconv.Itoa(start))
			query.Set("max", strconv.Itoa(pageSize))
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.path, query, nil)
		if err != nil {
			return fmt.Errorf("read deputy %s: %w", endpoint.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode deputy %s page: %w", endpoint.path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// Stop when this endpoint cannot page, when a short page is returned, or
		// when the page came back empty.
		if !endpoint.paginated || len(records) < pageSize || len(records) == 0 {
			return nil
		}
		start += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise deputy credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"Id":                  int64(i),
			"CompanyName":         fmt.Sprintf("Fixture Location %d", i),
			"Code":                fmt.Sprintf("LOC-%d", i),
			"OperationalUnitName": fmt.Sprintf("Fixture Department %d", i),
			"DisplayName":         fmt.Sprintf("Fixture Employee %d", i),
			"FirstName":           "Fixture",
			"LastName":            strconv.Itoa(i),
			"Title":               fmt.Sprintf("Fixture Task %d", i),
			"Active":              true,
			"Completed":           false,
			"Company":             int64(1),
			"Employee":            int64(i),
			"OperationalUnit":     int64(1),
			"TotalTime":           float64(i),
			"IsInProgress":        false,
			"Date":                "2026-01-01",
			"DueTime":             "2026-01-02T00:00:00",
			"Priority":            int64(2),
			"Creator":             int64(1),
			"Created":             "2026-01-01T00:00:00",
			"Modified":            "2026-01-01T00:00:00",
		}
		record := endpoint.mapRecord(item)
		record["connector"] = "deputy"
		record["fixture"] = true
		record["stream"] = stream
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := deputyBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := deputySecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("deputy connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: deputyUserAgent,
	}, nil
}

func deputySecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// deputyBaseURL resolves and validates the install-specific base URL. Unlike a
// shared-host API there is no usable default, so base_url is required and must be
// an absolute http(s) URL with a host to bound SSRF risk.
func deputyBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return "", errors.New("deputy connector requires config base_url (https://{installname}.{geo}.deputy.com)")
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("deputy config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("deputy config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("deputy config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func deputyPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return deputyDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("deputy config page_size must be an integer: %w", err)
	}
	if value < 1 || value > deputyMaxPageSize {
		return 0, fmt.Errorf("deputy config page_size must be between 1 and %d", deputyMaxPageSize)
	}
	return value, nil
}

func deputyMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("deputy config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("deputy config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
