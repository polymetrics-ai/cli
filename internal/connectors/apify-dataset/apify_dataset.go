// Package apifydataset implements the native pm Apify Dataset connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit, following
// the stripe reference shape: a thin package that composes a connsdk Requester +
// Bearer auth + RecordsAt extraction with Apify-specific stream definitions and
// offset/limit pagination.
//
// The connector is read-only: Apify datasets are an analytics/ETL source, so
// there are no safe reverse-ETL writes to expose. Like the other per-system
// connectors it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
//
// The directory and registry key keep the hyphen ("apify-dataset"); the Go
// package name strips it ("apifydataset") because hyphens are not valid package
// identifiers.
package apifydataset

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
	registryName             = "apify-dataset"
	apifyDefaultBaseURL      = "https://api.apify.com/v2"
	apifyDefaultPageSize     = 1000
	apifyMaxPageSize         = 50000
	apifyUserAgent           = "polymetrics-go-cli"
	apifyFixtureDatasetID    = "ds_fixture"
	apifyFixtureItemsPerPage = 2
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the Apify Dataset connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Apify Dataset connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "Apify Dataset",
		IntegrationType: "api",
		Description:     "Reads Apify dataset items and dataset metadata (item_collection, dataset_collection, dataset) through the Apify API v2.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Apify. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := apifyBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(apifySecret(cfg)) == "" {
		return errors.New("apify-dataset connector requires secret token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the datasets list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "datasets", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check apify-dataset: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: apifyStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "item_collection"
	}
	ep, ok := apifyStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("apify-dataset stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, ep, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	path, err := ep.path(req.Config)
	if err != nil {
		return err
	}
	if !ep.paginated {
		return c.readSingle(ctx, r, ep, path, emit)
	}
	pageSize, err := apifyPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := apifyMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, ep, path, pageSize, maxPages, emit)
}

// Write is unsupported: Apify Dataset is a read-only ETL source, so the
// connector advertises Write=false and rejects writes.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// readSingle reads an endpoint that returns a single object (the dataset
// metadata stream), wrapped under data.
func (c Connector) readSingle(ctx context.Context, r *connsdk.Requester, ep streamEndpoint, path string, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return fmt.Errorf("read apify-dataset %s: %w", ep.name, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, ep.recordsPath)
	if err != nil {
		return fmt.Errorf("decode apify-dataset %s: %w", ep.name, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(ep.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// harvest drives Apify's offset/limit pagination. Both the items endpoint (a
// top-level array) and the management list endpoints (records under data.items)
// page by advancing offset until a short page is returned. recordsPath selects
// which shape applies per stream.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, ep streamEndpoint, path string, pageSize, maxPages int, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(offset))

		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read apify-dataset %s: %w", ep.name, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, ep.recordsPath)
		if err != nil {
			return fmt.Errorf("decode apify-dataset %s page: %w", ep.name, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(ep.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short page (fewer than the requested limit) means we have reached the
		// end of the dataset.
		if len(records) < pageSize {
			return nil
		}
		offset += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise apify-dataset credential-free (mirrors the
// stripe fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, ep streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	datasetID := strings.TrimSpace(req.Config.Config["dataset_id"])
	if datasetID == "" {
		datasetID = apifyFixtureDatasetID
	}
	for i := 1; i <= apifyFixtureItemsPerPage; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var raw map[string]any
		switch stream {
		case "item_collection":
			// item_collection wraps each raw dataset item under "data".
			raw = map[string]any{
				"id":        fmt.Sprintf("item_fixture_%d", i),
				"value":     int64(i),
				"connector": registryName,
				"fixture":   true,
			}
		case "dataset_collection":
			raw = map[string]any{
				"id":        fmt.Sprintf("ds_fixture_%d", i),
				"name":      fmt.Sprintf("fixture-dataset-%d", i),
				"itemCount": int64(10 * i),
				"createdAt": "2026-01-01T00:00:00.000Z",
				"fixture":   true,
			}
		default: // dataset
			raw = map[string]any{
				"id":        datasetID,
				"name":      "fixture-dataset",
				"itemCount": int64(42),
				"createdAt": "2026-01-01T00:00:00.000Z",
				"fixture":   true,
			}
		}
		record := ep.mapRecord(raw)
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
		if !ep.paginated {
			break
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := apifyBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := apifySecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("apify-dataset connector requires secret token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: apifyUserAgent,
	}, nil
}

func apifySecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["token"]
}

// apifyDatasetID resolves the configured dataset id, required by the dataset and
// item_collection streams.
func apifyDatasetID(cfg connectors.RuntimeConfig) (string, error) {
	id := strings.TrimSpace(cfg.Config["dataset_id"])
	if id == "" {
		return "", errors.New("apify-dataset connector requires config dataset_id for this stream")
	}
	return url.PathEscape(id), nil
}

// apifyBaseURL resolves and validates the base URL. The default is
// api.apify.com/v2; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func apifyBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return apifyDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("apify-dataset config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("apify-dataset config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("apify-dataset config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func apifyPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return apifyDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("apify-dataset config page_size must be an integer: %w", err)
	}
	if value < 1 || value > apifyMaxPageSize {
		return 0, fmt.Errorf("apify-dataset config page_size must be between 1 and %d", apifyMaxPageSize)
	}
	return value, nil
}

func apifyMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("apify-dataset config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("apify-dataset config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
