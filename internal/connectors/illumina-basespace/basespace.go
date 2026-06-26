// Package basespace implements the native pm Illumina BaseSpace connector
// (registry key "illumina-basespace"). It is a declarative-HTTP per-system
// connector following the stripe template: a thin package composing the connsdk
// toolkit (Requester + x-access-token header auth + Response.Items extraction +
// Offset/Limit pagination) with BaseSpace-specific stream definitions.
//
// BaseSpace is read-only here: the v1pre3 API exposes sequencing projects, runs,
// samples, app sessions, and datasets, and there is no safe reverse-ETL write
// surface, so Capabilities.Write is false.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package basespace

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
	connectorName          = "illumina-basespace"
	basespaceAPIPrefix     = "v1pre3"
	basespaceDefaultUser   = "current"
	basespaceDefaultPageSz = 100
	basespaceMaxPageSize   = 1000
	basespaceUserAgent     = "polymetrics-go-cli"
	// basespaceFixtureCreated is the deterministic DateCreated used by fixture
	// records (2026-01-01T00:00:00Z).
	basespaceFixtureCreated = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory(connectorName, New)
}

// New returns the BaseSpace connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Illumina BaseSpace connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            connectorName,
		DisplayName:     "Illumina BaseSpace",
		IntegrationType: "api",
		Description:     "Reads Illumina BaseSpace projects, runs, samples, app sessions, and datasets through the BaseSpace v1pre3 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to BaseSpace.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := basespaceBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(basespaceToken(cfg)) == "" {
		return errors.New("illumina-basespace connector requires secret access_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the current user confirms auth and connectivity without
	// listing any resource.
	user := basespaceUser(cfg)
	path := fmt.Sprintf("%s/users/%s", basespaceAPIPrefix, url.PathEscape(user))
	if err := r.DoJSON(ctx, http.MethodGet, path, nil, nil, nil); err != nil {
		return fmt.Errorf("check illumina-basespace: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: basespaceStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a BaseSpace stream starts
// with an empty incremental cursor (full sync). The upstream source is
// full-refresh only, but publishing a cursor keeps the state shape consistent
// with other connectors.
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
		stream = "projects"
	}
	endpoint, ok := basespaceStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("illumina-basespace stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := basespacePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := basespaceMaxPages(req.Config)
	if err != nil {
		return err
	}
	user := basespaceUser(req.Config)
	path := fmt.Sprintf("%s/users/%s/%s", basespaceAPIPrefix, url.PathEscape(user), endpoint.resource)
	return c.harvest(ctx, r, path, endpoint, pageSize, maxPages, emit)
}

// harvest drives BaseSpace's Offset/Limit pagination over the Response.Items
// envelope. Each page returns {"Response":{"Items":[...],"Offset":N,"Limit":N,
// "TotalCount":N}}; the loop advances Offset by the page size until a short page
// (fewer than Limit items) is returned. The loop lives here rather than using
// connsdk.OffsetPaginator because the records live under a two-segment path
// (Response.Items) that OffsetPaginator's recordsPath handles, but we also want
// to flatten each item through the stream mapper, which Harvest does not do.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("Limit", strconv.Itoa(pageSize))
		query.Set("Offset", strconv.Itoa(offset))
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read illumina-basespace %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "Response.Items")
		if err != nil {
			return fmt.Errorf("decode illumina-basespace %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// Stop when the API returned a short (or empty) page.
		if len(records) < pageSize {
			return nil
		}
		offset += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free (mirrors the
// stripe fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"Id":             fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"Name":           fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"Href":           fmt.Sprintf("v1pre3/%s/%s_fixture_%d", endpoint.resource, endpoint.resource, i),
			"DateCreated":    basespaceFixtureCreated,
			"DateModified":   basespaceFixtureCreated,
			"DateCompleted":  basespaceFixtureCreated,
			"Status":         "Complete",
			"StatusSummary":  "Completed",
			"ExperimentName": fmt.Sprintf("Experiment %d", i),
			"InstrumentName": "NovaSeq",
			"SampleId":       fmt.Sprintf("S%d", i),
			"TotalSize":      int64(1000 * i),
			"NumReadsRaw":    int64(500 * i),
			"NumReadsPF":     int64(450 * i),
			"Application":    map[string]any{"Id": "app_1", "Name": "FixtureApp"},
			"DatasetType":    map[string]any{"Id": "common.fastq", "Name": "FASTQ"},
			"Project":        map[string]any{"Id": "projects_fixture_1", "Name": "Fixture project 1"},
			"UserOwnedBy":    map[string]any{"Id": "user_1", "Name": "Fixture User"},
		}
		record := endpoint.mapRecord(item)
		record["connector"] = connectorName
		record["fixture"] = true
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with x-access-token header auth and
// the resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it
// is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := basespaceBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := basespaceToken(cfg)
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("illumina-basespace connector requires secret access_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("x-access-token", token, ""),
		UserAgent: basespaceUserAgent,
	}, nil
}

func basespaceToken(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["access_token"]
}

func basespaceUser(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return basespaceDefaultUser
	}
	user := strings.TrimSpace(cfg.Config["user"])
	if user == "" {
		return basespaceDefaultUser
	}
	return user
}

// basespaceBaseURL resolves and validates the base URL. BaseSpace is
// domain-scoped (e.g. https://api.basespace.illumina.com or a regional domain
// like https://euw2.sh.basespace.illumina.com), so there is no single fixed
// default — the base_url (or domain) config must be supplied. Any value must be
// an absolute https (or http for local test servers) URL with a host to bound
// SSRF risk.
func basespaceBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		// Fall back to the domain config field (the catalog's required field),
		// promoting a bare host to an https URL.
		domain := strings.TrimSpace(cfg.Config["domain"])
		if domain == "" {
			return "", errors.New("illumina-basespace config requires base_url or domain")
		}
		if !strings.Contains(domain, "://") {
			domain = "https://" + domain
		}
		base = domain
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("illumina-basespace config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("illumina-basespace config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("illumina-basespace config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func basespacePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return basespaceDefaultPageSz, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("illumina-basespace config page_size must be an integer: %w", err)
	}
	if value < 1 || value > basespaceMaxPageSize {
		return 0, fmt.Errorf("illumina-basespace config page_size must be between 1 and %d", basespaceMaxPageSize)
	}
	return value, nil
}

func basespaceMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("illumina-basespace config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("illumina-basespace config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: BaseSpace is read-only in this connector.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
