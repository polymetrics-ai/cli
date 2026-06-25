// Package breezyhr implements the native pm Breezy HR connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit, modeled on
// the stripe reference connector: a thin package that composes a connsdk
// Requester (raw-API-key Authorization header) with Breezy-specific stream
// definitions, endpoints, and record mappers.
//
// Breezy's REST API is read-oriented for the applicant-tracking objects we
// surface (positions, candidates, pipelines), so this connector is read-only.
// It self-registers with the connectors registry via RegisterFactory in init()
// under the key "breezy-hr"; the registryset package blank-imports this package
// in the production binary to run that side effect.
package breezyhr

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
	breezyConnectorName = "breezy-hr"
	// breezyDefaultBaseRoot is the Breezy API root; the per-company base URL is
	// breezyDefaultBaseRoot + "/company/<company_id>".
	breezyDefaultBaseRoot = "https://api.breezy.hr/v3"
	breezyDefaultPageSize = 100
	breezyMaxPageSize     = 100
	breezyUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory(breezyConnectorName, New)
}

// New returns the Breezy HR connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Breezy HR connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return breezyConnectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            breezyConnectorName,
		DisplayName:     "Breezy HR",
		IntegrationType: "api",
		Description:     "Reads Breezy HR positions, candidates, and pipelines through the Breezy v3 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Breezy. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := breezyBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(breezySecret(cfg)) == "" {
		return errors.New("breezy-hr connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// Listing positions confirms auth, the company_id, and connectivity without
	// mutating anything.
	if _, err := r.Do(ctx, http.MethodGet, "positions", nil, nil); err != nil {
		return fmt.Errorf("check breezy-hr: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: breezyStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "positions"
	}
	if !breezyKnownStream(stream) {
		return fmt.Errorf("breezy-hr stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := breezyPageSize(req.Config)
	if err != nil {
		return err
	}

	switch stream {
	case "positions":
		return c.readPositions(ctx, r, pageSize, emit)
	case "pipelines":
		return c.readSimpleList(ctx, r, "pipelines", breezyPipelineRecord, emit)
	case "candidates":
		return c.readCandidates(ctx, r, pageSize, emit)
	default:
		return fmt.Errorf("breezy-hr stream %q not found", stream)
	}
}

// readPositions reads the positions list. Breezy returns a top-level JSON array
// and supports page/limit query params; the loop advances pages until a short
// (or empty) page is returned.
func (c Connector) readPositions(ctx context.Context, r *connsdk.Requester, pageSize int, emit func(connectors.Record) error) error {
	_, err := c.harvestPositions(ctx, r, pageSize, func(item map[string]any) error {
		return emit(breezyPositionRecord(item))
	})
	return err
}

// readCandidates is a substream of positions: for each position it reads the
// candidates list and propagates the parent position_id onto each record.
func (c Connector) readCandidates(ctx context.Context, r *connsdk.Requester, pageSize int, emit func(connectors.Record) error) error {
	positionIDs, err := c.harvestPositions(ctx, r, pageSize, nil)
	if err != nil {
		return err
	}
	for _, pid := range positionIDs {
		if err := ctx.Err(); err != nil {
			return err
		}
		path := "position/" + url.PathEscape(pid) + "/candidates"
		query := url.Values{"sort": []string{"updated_date"}}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read breezy-hr candidates for %s: %w", pid, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode breezy-hr candidates for %s: %w", pid, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(breezyCandidateRecord(item, pid)); err != nil {
				return err
			}
		}
	}
	return nil
}

// readSimpleList reads a single top-level-array endpoint with no pagination.
func (c Connector) readSimpleList(ctx context.Context, r *connsdk.Requester, resource string, mapRecord func(map[string]any) connectors.Record, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, resource, nil, nil)
	if err != nil {
		return fmt.Errorf("read breezy-hr %s: %w", resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode breezy-hr %s: %w", resource, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// harvestPositions drives page-based pagination over the positions list. When
// emit is non-nil it is invoked for every raw position; the returned slice always
// collects the position ids (used to drive the candidates substream). The loop
// stops on a short page (len < pageSize) or an empty page.
func (c Connector) harvestPositions(ctx context.Context, r *connsdk.Requester, pageSize int, emit func(map[string]any) error) ([]string, error) {
	var ids []string
	for page := 1; ; page++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("limit", strconv.Itoa(pageSize))
		resp, err := r.Do(ctx, http.MethodGet, "positions", query, nil)
		if err != nil {
			return nil, fmt.Errorf("read breezy-hr positions: %w", err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return nil, fmt.Errorf("decode breezy-hr positions page %d: %w", page, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			if id := stringField(item, "_id"); id != "" {
				ids = append(ids, id)
			}
			if emit != nil {
				if err := emit(item); err != nil {
					return nil, err
				}
			}
		}
		if len(records) < pageSize || len(records) == 0 {
			return ids, nil
		}
	}
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise breezy-hr credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var rec connectors.Record
		switch stream {
		case "candidates":
			rec = breezyCandidateRecord(map[string]any{
				"_id":           fmt.Sprintf("cand_fixture_%d", i),
				"name":          fmt.Sprintf("Fixture Candidate %d", i),
				"email_address": fmt.Sprintf("fixture+%d@example.com", i),
				"stage":         map[string]any{"name": "Applied"},
				"creation_date": "2026-01-01T00:00:00Z",
				"updated_date":  "2026-01-02T00:00:00Z",
			}, "pos_fixture_1")
		case "pipelines":
			rec = breezyPipelineRecord(map[string]any{
				"_id":  fmt.Sprintf("pipe_fixture_%d", i),
				"name": fmt.Sprintf("Fixture Pipeline %d", i),
			})
		default: // positions
			rec = breezyPositionRecord(map[string]any{
				"_id":           fmt.Sprintf("pos_fixture_%d", i),
				"name":          fmt.Sprintf("Fixture Position %d", i),
				"type":          map[string]any{"name": "Full-Time"},
				"state":         "published",
				"creation_date": "2026-01-01T00:00:00Z",
				"updated_date":  "2026-01-02T00:00:00Z",
			})
		}
		rec["connector"] = breezyConnectorName
		rec["fixture"] = true
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with raw-API-key Authorization auth
// and the resolved per-company base URL. The api_key only ever flows into
// connsdk.APIKeyHeader; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := breezyBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := breezySecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("breezy-hr connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", secret, ""),
		UserAgent: breezyUserAgent,
	}, nil
}

func breezySecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// breezyCompanyID resolves the company id. It is a "secret" field per the catalog
// but is part of the URL path rather than a credential, so it may also be set via
// config for tests/automation; secrets take precedence.
func breezyCompanyID(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets != nil {
		if v := strings.TrimSpace(cfg.Secrets["company_id"]); v != "" {
			return v
		}
	}
	if cfg.Config != nil {
		return strings.TrimSpace(cfg.Config["company_id"])
	}
	return ""
}

// breezyBaseURL resolves and validates the per-company base URL. The default
// root is api.breezy.hr; any base_url override must be an absolute https (or http
// for local test servers) URL with a host to bound SSRF risk. The company_id is
// appended as /company/<company_id>.
func breezyBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	root := strings.TrimSpace(cfg.Config["base_url"])
	if root == "" {
		root = breezyDefaultBaseRoot
	} else {
		parsed, err := url.Parse(root)
		if err != nil {
			return "", fmt.Errorf("breezy-hr config base_url is invalid: %w", err)
		}
		if parsed.Scheme != "https" && parsed.Scheme != "http" {
			return "", fmt.Errorf("breezy-hr config base_url must use http or https, got %q", parsed.Scheme)
		}
		if parsed.Host == "" {
			return "", errors.New("breezy-hr config base_url must include a host")
		}
	}
	companyID := breezyCompanyID(cfg)
	if companyID == "" {
		return "", errors.New("breezy-hr connector requires company_id")
	}
	return strings.TrimRight(root, "/") + "/company/" + companyID, nil
}

func breezyPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return breezyDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("breezy-hr config page_size must be an integer: %w", err)
	}
	if value < 1 || value > breezyMaxPageSize {
		return 0, fmt.Errorf("breezy-hr config page_size must be between 1 and %d", breezyMaxPageSize)
	}
	return value, nil
}

func breezyKnownStream(stream string) bool {
	switch stream {
	case "positions", "candidates", "pipelines":
		return true
	default:
		return false
	}
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// stringField reads a string-able value at key, stringifying non-string scalars
// and returning "" for nil/missing.
func stringField(item map[string]any, key string) string {
	switch v := item[key].(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}

// Write satisfies the connectors.Connector interface. Breezy HR is read-only in
// this connector, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{RecordsFailed: len(records)}, connectors.ErrUnsupportedOperation
}
