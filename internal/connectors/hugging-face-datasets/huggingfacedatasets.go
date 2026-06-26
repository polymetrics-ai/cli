// Package huggingfacedatasets implements the native pm Hugging Face Datasets
// connector. It is a declarative-HTTP per-system connector following the stripe
// reference template: a thin package that composes the connsdk toolkit
// (Requester + optional Bearer auth + RecordsAt extraction) with Hugging Face
// dataset-viewer stream definitions, endpoints, and offset pagination.
//
// The connector reads the public Hugging Face dataset-viewer REST API
// (https://datasets-server.huggingface.co). Authentication is OPTIONAL: public
// datasets are readable without credentials, while gated/private datasets require
// a user access token supplied as the api_token secret and sent as a Bearer
// token. It is read-only; there is no reverse-ETL surface for this API.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package huggingfacedatasets

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
	connectorName       = "hugging-face-datasets"
	defaultBaseURL      = "https://datasets-server.huggingface.co"
	defaultRowsPageSize = 100
	maxRowsPageSize     = 100
	userAgent           = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory(connectorName, New)
}

// New returns the Hugging Face Datasets connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Hugging Face Datasets connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            connectorName,
		DisplayName:     "Hugging Face - Datasets",
		IntegrationType: "api",
		Description:     "Reads dataset splits, per-split sizes, and rows from the Hugging Face dataset-viewer REST API. Read-only; an optional user access token unlocks gated and private datasets.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to the
// dataset-viewer API. In fixture mode it short-circuits without a network call.
// It requires a dataset_name and confirms the dataset is loadable via /is-valid.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := baseURL(cfg); err != nil {
		return err
	}
	dataset, err := datasetName(cfg)
	if err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// /is-valid is a cheap, read-only probe that confirms both connectivity and
	// (for gated datasets) credential validity without fetching any rows.
	if err := r.DoJSON(ctx, http.MethodGet, "is-valid", url.Values{"dataset": []string{dataset}}, nil, nil); err != nil {
		return fmt.Errorf("check hugging-face-datasets: %w", err)
	}
	return nil
}

// Write is unsupported: the Hugging Face dataset-viewer API is read-only, so the
// connector advertises Capabilities.Write=false and rejects writes.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: datasetStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "splits"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("hugging-face-datasets stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	dataset, err := datasetName(req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}

	switch endpoint.kind {
	case kindRows:
		return c.readRows(ctx, r, dataset, endpoint, req, emit)
	default:
		return c.readList(ctx, r, dataset, endpoint, emit)
	}
}

// readList fetches a single non-paginated object (e.g. /splits, /size) and emits
// the array of records found at endpoint.recordsPath.
func (c Connector) readList(ctx context.Context, r *connsdk.Requester, dataset string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	query := url.Values{"dataset": []string{dataset}}
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
	if err != nil {
		return fmt.Errorf("read hugging-face-datasets %s: %w", endpoint.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return fmt.Errorf("decode hugging-face-datasets %s: %w", endpoint.resource, err)
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

// readRows drives the /rows endpoint's offset pagination. The endpoint takes
// dataset/config/split/offset/length and returns {"rows":[...]}; pages advance by
// length until a short page (fewer than length records) is returned. The connsdk
// OffsetPaginator does not carry the row response shape's extra dataset/config/
// split params cleanly here, so the bounded loop lives in-package, built on
// connsdk.Requester + connsdk.RecordsAt.
func (c Connector) readRows(ctx context.Context, r *connsdk.Requester, dataset string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	config := strings.TrimSpace(req.Config.Config["config"])
	split := strings.TrimSpace(req.Config.Config["split"])
	if config == "" || split == "" {
		return errors.New("hugging-face-datasets rows stream requires config and split in config (a (config, split) slice to read)")
	}
	pageSize, err := rowsPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := maxPagesConfig(req.Config)
	if err != nil {
		return err
	}

	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{
			"dataset": []string{dataset},
			"config":  []string{config},
			"split":   []string{split},
			"offset":  []string{strconv.Itoa(offset)},
			"length":  []string{strconv.Itoa(pageSize)},
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read hugging-face-datasets rows: %w", err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode hugging-face-datasets rows: %w", err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			rec := endpoint.mapRecord(item)
			// /rows elements do not echo the slice identity; stamp it on so the
			// emitted records are self-describing.
			rec["dataset"] = dataset
			rec["config"] = config
			rec["split"] = split
			if err := emit(rec); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
		offset += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	dataset := strings.TrimSpace(req.Config.Config["dataset_name"])
	if dataset == "" {
		dataset = "fixture/dataset"
	}
	for i := 0; i < 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var item map[string]any
		switch endpoint.kind {
		case kindRows:
			item = map[string]any{
				"row_idx": int64(i),
				"row": map[string]any{
					"text":  fmt.Sprintf("fixture row %d", i),
					"label": int64(i % 2),
				},
				"truncated_cells": []any{},
			}
		default:
			item = map[string]any{
				"dataset":                 dataset,
				"config":                  "default",
				"split":                   []string{"train", "test"}[i],
				"num_rows":                int64(100 * (i + 1)),
				"num_columns":             int64(3),
				"num_bytes_parquet_files": int64(1024 * (i + 1)),
				"num_bytes_memory":        int64(4096 * (i + 1)),
			}
		}
		rec := endpoint.mapRecord(item)
		if endpoint.kind == kindRows {
			rec["dataset"] = dataset
			rec["config"] = "default"
			rec["split"] = "train"
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the resolved base URL and,
// when a token is configured, optional Bearer auth. The secret only ever flows
// into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	r := &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		UserAgent: userAgent,
	}
	if token := apiToken(cfg); token != "" {
		r.Auth = connsdk.Bearer(token)
	}
	return r, nil
}

// apiToken resolves the optional Hugging Face user access token. Public datasets
// do not need one; gated/private datasets do.
func apiToken(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	for _, key := range []string{"api_token", "access_token", "token"} {
		if v := strings.TrimSpace(cfg.Secrets[key]); v != "" {
			return v
		}
	}
	return ""
}

// datasetName resolves the required dataset identifier (e.g. "ibm/duorc").
func datasetName(cfg connectors.RuntimeConfig) (string, error) {
	if cfg.Config == nil {
		return "", errors.New("hugging-face-datasets connector requires config dataset_name")
	}
	name := strings.TrimSpace(cfg.Config["dataset_name"])
	if name == "" {
		return "", errors.New("hugging-face-datasets connector requires config dataset_name")
	}
	return name, nil
}

// baseURL resolves and validates the base URL. The default is
// datasets-server.huggingface.co; any override must be an absolute https (or
// http for local test servers) URL with a host to bound SSRF risk.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	if cfg.Config == nil {
		return defaultBaseURL, nil
	}
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return defaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("hugging-face-datasets config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("hugging-face-datasets config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("hugging-face-datasets config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func rowsPageSize(cfg connectors.RuntimeConfig) (int, error) {
	if cfg.Config == nil {
		return defaultRowsPageSize, nil
	}
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultRowsPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("hugging-face-datasets config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxRowsPageSize {
		return 0, fmt.Errorf("hugging-face-datasets config page_size must be between 1 and %d", maxRowsPageSize)
	}
	return value, nil
}

func maxPagesConfig(cfg connectors.RuntimeConfig) (int, error) {
	if cfg.Config == nil {
		return 0, nil
	}
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("hugging-face-datasets config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("hugging-face-datasets config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
