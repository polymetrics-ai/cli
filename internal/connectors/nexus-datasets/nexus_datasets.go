// Package nexusdatasets implements the native pm Infor Nexus Datasets connector.
// It follows the declarative-HTTP per-system shape established by the stripe
// connector: a thin package that composes the connsdk toolkit (Requester +
// custom HMAC auth + RecordsAt extraction + cursor state) with Infor Nexus Data
// API (v3.1) dataset export semantics.
//
// The connector self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the
// production binary to run that side effect.
//
// The directory is internal/connectors/nexus-datasets (hyphenated, the bare
// system name); the Go package identifier is nexusdatasets and the registry key
// is the exact bare name "nexus-datasets".
package nexusdatasets

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	nexusDefaultPageSize = 100
	nexusMaxPageSize     = 1000
	nexusUserAgent       = "polymetrics-go-cli"
	// nexusFixtureUpdated is the deterministic updated_at base used by the
	// fixture-mode records.
	nexusFixtureUpdated = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("nexus-datasets", New)
}

// New returns the Infor Nexus Datasets connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Infor Nexus Datasets connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "nexus-datasets" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "nexus-datasets",
		DisplayName:     "Infor Nexus Datasets",
		IntegrationType: "api",
		Description:     "Reads records from a configured Infor Nexus export dataset through the Infor Nexus Data API (v3.1) using HMAC authentication. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Infor Nexus.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := nexusBaseURL(cfg); err != nil {
		return err
	}
	if err := requireCredentials(cfg); err != nil {
		return err
	}
	dataset := datasetName(cfg)
	if dataset == "" {
		return errors.New("nexus-datasets connector requires config dataset_name")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the dataset confirms auth and connectivity.
	query := url.Values{"limit": []string{"1"}, "offset": []string{"0"}}
	if _, err := r.Do(ctx, http.MethodGet, datasetPath(dataset), query, nil); err != nil {
		return fmt.Errorf("check nexus-datasets: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: nexusStreams()}, nil
}

// Write is unsupported: Infor Nexus dataset export is a read-only source. The
// method exists to satisfy the connectors.Connector interface.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: a dataset stream starts with
// an empty incremental cursor (full sync), which the start_date config can raise
// at read time.
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
		stream = "datasets"
	}
	if stream != "datasets" {
		return fmt.Errorf("nexus-datasets stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, req, emit)
	}

	dataset := datasetName(req.Config)
	if dataset == "" {
		return errors.New("nexus-datasets connector requires config dataset_name")
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := nexusPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := nexusMaxPages(req.Config)
	if err != nil {
		return err
	}
	lower := incrementalLowerBound(req)
	return c.harvest(ctx, r, dataset, pageSize, maxPages, lower, emit)
}

// harvest drives offset/limit pagination over the dataset export endpoint. Infor
// Nexus dataset responses wrap rows under "records"; pages are advanced by
// offset until a short (or empty) page is returned. The loop lives here, built
// on connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, dataset string, pageSize, maxPages int, modifiedSince string, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(offset))
		if modifiedSince != "" {
			query.Set("modifiedSince", modifiedSince)
		}
		resp, err := r.Do(ctx, http.MethodGet, datasetPath(dataset), query, nil)
		if err != nil {
			return fmt.Errorf("read nexus-datasets %s: %w", dataset, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "records")
		if err != nil {
			return fmt.Errorf("decode nexus-datasets %s page: %w", dataset, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			rec := nexusDatasetRecord(item)
			rec["dataset_name"] = dataset
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
func (c Connector) readFixture(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	dataset := datasetName(req.Config)
	if dataset == "" {
		dataset = "fixture_dataset"
	}
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		payload := map[string]any{
			"order_id": fmt.Sprintf("ord_%d", i),
			"amount":   int64(1000 * i),
			"status":   "confirmed",
		}
		raw, _ := json.Marshal(payload)
		item := map[string]any{
			"id":              fmt.Sprintf("%s_fixture_%d", dataset, i),
			"raw_data":        payload,
			"raw_data_string": string(raw),
			"updated_at":      nexusFixtureUpdated,
		}
		rec := nexusDatasetRecord(item)
		rec["dataset_name"] = dataset
		rec["fixture"] = true
		if cursor := req.State["cursor"]; cursor != "" {
			rec["previous_cursor"] = cursor
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with HMAC auth, the resolved base
// URL, and the Infor Nexus credential headers. Secrets only ever flow into the
// authenticator; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := nexusBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	if err := requireCredentials(cfg); err != nil {
		return nil, err
	}
	accessKeyID := strings.TrimSpace(cfg.Config["access_key_id"])
	userID := strings.TrimSpace(cfg.Config["user_id"])
	headers := map[string]string{
		"X-Infor-AccessKeyId": accessKeyID,
		"X-Infor-UserId":      userID,
		"X-Infor-ApiKey":      apiKey(cfg),
	}
	return &connsdk.Requester{
		Client:         c.Client,
		BaseURL:        base,
		Auth:           hmacAuth{accessKeyID: accessKeyID, secretKey: secretKey(cfg)},
		UserAgent:      nexusUserAgent,
		DefaultHeaders: headers,
	}, nil
}

// hmacAuth signs each request with an HMAC-SHA256 signature over the canonical
// request (method, path, timestamp) keyed by the secret key, and sets it on the
// Authorization header. It never logs secret material. The exact upstream
// canonicalization may differ across Infor Nexus deployments; this implements
// the common HMAC-SHA256 scheme and keeps the secret confined to this type.
type hmacAuth struct {
	accessKeyID string
	secretKey   string
}

func (a hmacAuth) Apply(_ context.Context, req *http.Request) error {
	ts := strconv.FormatInt(time.Now().UTC().Unix(), 10)
	canonical := strings.Join([]string{req.Method, req.URL.Path, ts}, "\n")
	mac := hmac.New(sha256.New, []byte(a.secretKey))
	mac.Write([]byte(canonical))
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	req.Header.Set("X-Infor-Timestamp", ts)
	req.Header.Set("Authorization", fmt.Sprintf("InforNexus %s:%s", a.accessKeyID, sig))
	return nil
}

// incrementalLowerBound returns the modifiedSince lower bound derived from the
// incremental cursor (if any) or else the start_date config. An empty result
// means no lower bound (full sync).
func incrementalLowerBound(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	return strings.TrimSpace(req.Config.Config["start_date"])
}

func requireCredentials(cfg connectors.RuntimeConfig) error {
	if strings.TrimSpace(cfg.Config["access_key_id"]) == "" {
		return errors.New("nexus-datasets connector requires config access_key_id")
	}
	if strings.TrimSpace(cfg.Config["user_id"]) == "" {
		return errors.New("nexus-datasets connector requires config user_id")
	}
	if strings.TrimSpace(secretKey(cfg)) == "" {
		return errors.New("nexus-datasets connector requires secret secret_key")
	}
	if strings.TrimSpace(apiKey(cfg)) == "" {
		return errors.New("nexus-datasets connector requires secret api_key")
	}
	return nil
}

// secretKey resolves the HMAC secret key, preferring the secrets map (where it
// belongs) but tolerating it in config for compatibility.
func secretKey(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets != nil {
		if v := strings.TrimSpace(cfg.Secrets["secret_key"]); v != "" {
			return v
		}
	}
	return strings.TrimSpace(cfg.Config["secret_key"])
}

// apiKey resolves the Infor Data API key, preferring the secrets map.
func apiKey(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets != nil {
		if v := strings.TrimSpace(cfg.Secrets["api_key"]); v != "" {
			return v
		}
	}
	return strings.TrimSpace(cfg.Config["api_key"])
}

func datasetName(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return ""
	}
	return strings.TrimSpace(cfg.Config["dataset_name"])
}

func datasetPath(dataset string) string {
	return "datasets/" + url.PathEscape(dataset)
}

// nexusBaseURL resolves and validates the base URL. base_url is required by the
// API spec; any value must be an absolute https (or http for local test servers)
// URL with a host to bound SSRF risk.
func nexusBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return "", errors.New("nexus-datasets connector requires config base_url")
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("nexus-datasets config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("nexus-datasets config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("nexus-datasets config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func nexusPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return nexusDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("nexus-datasets config page_size must be an integer: %w", err)
	}
	if value < 1 || value > nexusMaxPageSize {
		return 0, fmt.Errorf("nexus-datasets config page_size must be between 1 and %d", nexusMaxPageSize)
	}
	return value, nil
}

func nexusMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("nexus-datasets config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("nexus-datasets config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// fixtureMode reports whether the connector should run credential-free with
// canned records. The catalog's mode field accepts Full/Incremental; "fixture"
// is the pm-specific deterministic mode used by conformance.
func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

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

// firstNonEmpty returns the first non-empty stringified value among keys.
func firstNonEmpty(item map[string]any, keys ...string) string {
	for _, k := range keys {
		if v := stringField(item, k); v != "" {
			return v
		}
	}
	return ""
}
