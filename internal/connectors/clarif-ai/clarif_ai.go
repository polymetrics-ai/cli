// Package clarifai implements the native pm Clarif-ai (Clarifai) connector. It is
// a declarative-HTTP per-system connector built on the stripe template: a thin
// package that composes the connsdk toolkit (Requester + Clarifai "Authorization:
// Key <pat>" auth + RecordsAt extraction) with Clarifai-specific stream
// definitions, endpoints, and page/per_page pagination.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect. The Clarifai source is read-only (full refresh).
package clarifai

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
	registryName            = "clarif-ai"
	clarifaiDefaultBaseURL  = "https://api.clarifai.com/v2"
	clarifaiDefaultPageSize = 100
	clarifaiMaxPageSize     = 1000
	clarifaiUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the Clarif-ai connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Clarif-ai connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "Clarif-ai",
		IntegrationType: "api",
		Description:     "Reads Clarifai applications, datasets, models, model versions, and workflows through the Clarifai v2 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Clarifai. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := clarifaiBaseURL(cfg); err != nil {
		return err
	}
	userID, err := clarifaiUserID(cfg)
	if err != nil {
		return err
	}
	if strings.TrimSpace(clarifaiSecret(cfg)) == "" {
		return errors.New(registryName + " connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the apps list confirms auth and connectivity without
	// mutating anything.
	path := "users/" + url.PathEscape(userID) + "/apps"
	if err := r.DoJSON(ctx, http.MethodGet, path, url.Values{"per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check %s: %w", registryName, err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: clarifaiStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "applications"
	}
	endpoint, ok := clarifaiStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("%s stream %q not found", registryName, stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	userID, err := clarifaiUserID(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := clarifaiPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := clarifaiMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, userID, endpoint, pageSize, maxPages, emit)
}

// harvest drives Clarifai's page/per_page pagination. Clarifai list responses
// return {status:{...}, <resource>:[...]} with no total or has_more flag, so the
// loop advances the page number until a short (or empty) page is returned. There
// is no connsdk paginator for this exact short-page shape over a custom records
// key, so the loop lives here, built on connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, userID string, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := "users/" + url.PathEscape(userID) + "/" + endpoint.resource
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("per_page", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read %s %s: %w", registryName, endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsKey)
		if err != nil {
			return fmt.Errorf("decode %s %s page: %w", registryName, endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A page smaller than the requested size (including an empty page) means
		// there are no more results to fetch.
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// Write is unsupported: the Clarifai source connector is read-only (full
// refresh). It satisfies the connectors.Connector interface by returning
// ErrUnsupportedOperation, and Metadata reports Write=false.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
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
			"id":                      fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"name":                    fmt.Sprintf("Fixture %s %d", stream, i),
			"user_id":                 "fixture-user",
			"app_id":                  "fixture-app",
			"description":             fmt.Sprintf("fixture %s record %d", stream, i),
			"default_language":        "en",
			"model_type_id":           "visual-classifier",
			"created_at":              "2026-01-01T00:00:00Z",
			"modified_at":             "2026-01-02T00:00:00Z",
			"visibility":              map[string]any{"gettable": 10},
			"status":                  map[string]any{"code": 21100},
			"version":                 map[string]any{"id": "v1"},
			"default_processing_info": map[string]any{"frame_info": map[string]any{}},
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Clarifai "Authorization: Key
// <pat>" auth and the resolved base URL. The secret only ever flows into the
// authenticator; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := clarifaiBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := clarifaiSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New(registryName + " connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", secret, "Key "),
		UserAgent: clarifaiUserAgent,
	}, nil
}

func clarifaiSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

func clarifaiUserID(cfg connectors.RuntimeConfig) (string, error) {
	if cfg.Config == nil {
		return "", errors.New(registryName + " connector requires config user_id")
	}
	userID := strings.TrimSpace(cfg.Config["user_id"])
	if userID == "" {
		return "", errors.New(registryName + " connector requires config user_id")
	}
	return userID, nil
}

// clarifaiBaseURL resolves and validates the base URL. The default is
// api.clarifai.com/v2; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func clarifaiBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return clarifaiDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("%s config base_url is invalid: %w", registryName, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("%s config base_url must use http or https, got %q", registryName, parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New(registryName + " config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func clarifaiPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return clarifaiDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s config page_size must be an integer: %w", registryName, err)
	}
	if value < 1 || value > clarifaiMaxPageSize {
		return 0, fmt.Errorf("%s config page_size must be between 1 and %d", registryName, clarifaiMaxPageSize)
	}
	return value, nil
}

func clarifaiMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s config max_pages must be an integer, all, or unlimited: %w", registryName, err)
	}
	if value < 0 {
		return 0, errors.New(registryName + " config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
