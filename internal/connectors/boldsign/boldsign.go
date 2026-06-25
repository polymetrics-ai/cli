// Package boldsign implements the native pm BoldSign connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit (Requester
// + APIKeyHeader auth + RecordsAt extraction + page-number pagination), modelled
// on the stripe reference connector.
//
// BoldSign is an e-signature platform; this connector reads documents,
// templates, teams, contacts, and brands from the BoldSign REST API. It is
// read-only: BoldSign's write surface (sending signature requests, uploading
// documents) is not a safe generic reverse-ETL target, so Capabilities.Write is
// false.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package boldsign

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
	boldsignDefaultBaseURL  = "https://api.boldsign.com"
	boldsignDefaultPageSize = 50
	boldsignMaxPageSize     = 100
	boldsignUserAgent       = "polymetrics-go-cli"
	boldsignAPIKeyHeader    = "X-API-KEY"
)

func init() {
	connectors.RegisterFactory("boldsign", New)
}

// New returns the BoldSign connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm BoldSign connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "boldsign" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "boldsign",
		DisplayName:     "BoldSign",
		IntegrationType: "api",
		Description:     "Reads BoldSign documents, templates, teams, contacts, and brands through the BoldSign REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to BoldSign. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := boldsignBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(boldsignSecret(cfg)) == "" {
		return errors.New("boldsign connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the documents list confirms auth and connectivity
	// without mutating anything.
	query := url.Values{"Page": []string{"1"}, "PageSize": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "v1/document/list", query, nil, nil); err != nil {
		return fmt.Errorf("check boldsign: %w", err)
	}
	return nil
}

// Write satisfies the connectors.Connector interface. BoldSign is exposed
// read-only (Capabilities.Write=false): sending signature requests and uploading
// documents are not safe generic reverse-ETL targets, so every write is
// rejected.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{RecordsFailed: len(records)}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: boldsignStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "documents"
	}
	endpoint, ok := boldsignStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("boldsign stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := boldsignPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := boldsignMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives BoldSign's Page-number pagination. List responses are
// {result:[...], pageDetails:{...}} (teams uses results:[...]). The next page is
// requested by incrementing Page; the loop stops when a short page (fewer than
// PageSize records) is returned. The per-stream loop lives here because the
// records path differs by stream (result vs results), which connsdk.Harvest's
// single recordsPath cannot express across the table.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("Page", strconv.Itoa(page))
		query.Set("PageSize", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read boldsign %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode boldsign %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short page (fewer records than requested) means we have reached the
		// last page. An empty page also stops.
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise boldsign credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		idValue := fmt.Sprintf("%s_fixture_%d", stream, i)
		item := map[string]any{
			"documentId":          idValue,
			"teamId":              idValue,
			"brandId":             idValue,
			"id":                  idValue,
			"status":              "Completed",
			"senderEmail":         fmt.Sprintf("fixture+%d@example.com", i),
			"messageTitle":        fmt.Sprintf("Fixture Document %d", i),
			"templateName":        fmt.Sprintf("Fixture Template %d", i),
			"templateDescription": "fixture template",
			"teamName":            fmt.Sprintf("Fixture Team %d", i),
			"brandName":           fmt.Sprintf("Fixture Brand %d", i),
			"name":                fmt.Sprintf("Fixture Contact %d", i),
			"email":               fmt.Sprintf("contact+%d@example.com", i),
			"phoneNumber":         "+15555550100",
			"companyName":         "Fixture Co",
			"createdDate":         "2026-01-01T00:00:00Z",
			"expiryDate":          "2026-02-01T00:00:00Z",
			"isDeleted":           false,
			"isDefault":           i == 1,
			"isSharedTemplate":    false,
			"enableSigningOrder":  false,
			"backgroundColor":     "#ffffff",
			"buttonColor":         "#1a73e8",
		}
		record := endpoint.mapRecord(item)
		record["connector"] = "boldsign"
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

// requester builds a connsdk.Requester wired with X-API-KEY auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := boldsignBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := boldsignSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("boldsign connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(boldsignAPIKeyHeader, secret, ""),
		UserAgent: boldsignUserAgent,
	}, nil
}

func boldsignSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// boldsignBaseURL resolves and validates the base URL. The default is
// api.boldsign.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func boldsignBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return boldsignDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("boldsign config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("boldsign config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("boldsign config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func boldsignPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return boldsignDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("boldsign config page_size must be an integer: %w", err)
	}
	if value < 1 || value > boldsignMaxPageSize {
		return 0, fmt.Errorf("boldsign config page_size must be between 1 and %d", boldsignMaxPageSize)
	}
	return value, nil
}

func boldsignMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("boldsign config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("boldsign config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
