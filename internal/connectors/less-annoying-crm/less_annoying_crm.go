// Package lessannoyingcrm implements the native pm Less Annoying CRM connector.
// It is a declarative-HTTP per-system connector that composes the connsdk
// toolkit (Requester + APIKeyHeader auth + RecordsAt extraction + cursor state)
// with Less Annoying CRM-specific stream definitions and record mappers,
// following the shape of the reference stripe connector.
//
// The Less Annoying CRM v2 API is RPC-style: every call is a POST to the root
// path whose JSON body carries a Function name (e.g. GetContacts). Records live
// under the "Results" key (or the root array for GetUsers). Pagination is page
// increment via the Page / MaxNumberOfResults request parameters.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package lessannoyingcrm

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
	lacrmConnectorName  = "less-annoying-crm"
	lacrmDefaultBaseURL = "https://api.lessannoyingcrm.com/v2"
	lacrmDefaultPageSz  = 500
	lacrmMaxPageSize    = 10000
	lacrmUserAgent      = "polymetrics-go-cli"
	// lacrmRequestPath is the single RPC endpoint; every Function POSTs here.
	lacrmRequestPath = "/"
	// lacrmRecordsPath is where list Functions return their array.
	lacrmRecordsPath = "Results"
)

func init() {
	connectors.RegisterFactory(lacrmConnectorName, New)
}

// New returns the Less Annoying CRM connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Less Annoying CRM connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk
	// Requester. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return lacrmConnectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            lacrmConnectorName,
		DisplayName:     "Less Annoying CRM",
		IntegrationType: "api",
		Description:     "Reads Less Annoying CRM users, contacts, tasks, notes, and events through the Less Annoying CRM v2 API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Less
// Annoying CRM. In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := lacrmBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(lacrmSecret(cfg)) == "" {
		return errors.New("less-annoying-crm connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// GetUsers is a cheap call that confirms auth and connectivity without
	// reading bulk data or mutating anything.
	body := map[string]any{"Function": "GetUsers"}
	if err := r.DoJSON(ctx, http.MethodPost, lacrmRequestPath, nil, body, nil); err != nil {
		return fmt.Errorf("check less-annoying-crm: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: lacrmStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a stream starts with an
// empty incremental cursor (full sync), which the start_date config can raise at
// read time for the incremental streams.
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
	endpoint, ok := lacrmStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("less-annoying-crm stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := lacrmPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := lacrmMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Less Annoying CRM's PageIncrement pagination. Each request is a
// POST whose body carries the Function; Page and MaxNumberOfResults are supplied
// as request parameters (mirroring the upstream declarative manifest). A page
// shorter than pageSize ends the loop.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	body := map[string]any{"Function": endpoint.function}
	recordsPath := endpoint.recordsPath
	if recordsPath == "" {
		recordsPath = lacrmRecordsPath
	}

	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("Page", strconv.Itoa(page))
		query.Set("MaxNumberOfResults", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodPost, lacrmRequestPath, query, body)
		if err != nil {
			return fmt.Errorf("read less-annoying-crm %s: %w", endpoint.function, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, recordsPath)
		if err != nil {
			return fmt.Errorf("decode less-annoying-crm %s page: %w", endpoint.function, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
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
		item := map[string]any{
			"UserId":      fmt.Sprintf("user_fixture_%d", i),
			"ContactId":   fmt.Sprintf("%s_contact_%d", stream, i),
			"TaskId":      fmt.Sprintf("%s_task_%d", stream, i),
			"NoteId":      fmt.Sprintf("%s_note_%d", stream, i),
			"EventId":     fmt.Sprintf("%s_event_%d", stream, i),
			"Name":        fmt.Sprintf("Fixture %d", i),
			"FirstName":   fmt.Sprintf("Fixture%d", i),
			"LastName":    "Example",
			"Note":        fmt.Sprintf("Fixture note %d", i),
			"Description": fmt.Sprintf("Fixture description %d", i),
			"DateCreated": "2026-01-01T00:00:0" + strconv.Itoa(i) + "Z",
			"DateUpdated": "2026-01-01T00:00:0" + strconv.Itoa(i) + "Z",
			"LastUpdate":  "2026-01-01T00:00:0" + strconv.Itoa(i) + "Z",
			"DueDate":     "2026-01-02",
			"StartDate":   "2026-01-01T09:00:00Z",
			"EndDate":     "2026-01-01T10:00:00Z",
			"IsCompleted": false,
			"IsCompany":   false,
			"IsAllDay":    false,
			"connector":   lacrmConnectorName,
			"fixture":     true,
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

// requester builds a connsdk.Requester wired with API-key header auth and the
// resolved base URL. The Less Annoying CRM v2 API authenticates with the raw API
// key in the Authorization header (no Bearer prefix). The secret only ever flows
// into connsdk.APIKeyHeader; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := lacrmBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := lacrmSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("less-annoying-crm connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", secret, ""),
		UserAgent: lacrmUserAgent,
	}, nil
}

func lacrmSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// lacrmBaseURL resolves and validates the base URL. The default is
// api.lessannoyingcrm.com/v2; any override must be an absolute https (or http
// for local test servers) URL with a host to bound SSRF risk.
func lacrmBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return lacrmDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("less-annoying-crm config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("less-annoying-crm config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("less-annoying-crm config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func lacrmPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return lacrmDefaultPageSz, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("less-annoying-crm config page_size must be an integer: %w", err)
	}
	if value < 1 || value > lacrmMaxPageSize {
		return 0, fmt.Errorf("less-annoying-crm config page_size must be between 1 and %d", lacrmMaxPageSize)
	}
	return value, nil
}

func lacrmMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("less-annoying-crm config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("less-annoying-crm config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// Write reports that this connector is read-only. The Less Annoying CRM port
// exposes no reverse-ETL write actions.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
