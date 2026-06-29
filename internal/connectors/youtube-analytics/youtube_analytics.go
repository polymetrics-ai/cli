// Package youtubeanalytics implements the native pm YouTube Analytics connector.
// It follows the declarative-HTTP shape established by the stripe connector: a
// thin package that composes the connsdk toolkit (Requester + cursor pagination +
// RecordsAt extraction) with YouTube-Reporting-API-specific stream definitions
// and endpoints.
//
// The upstream source-youtube-analytics connector is built on YouTube's bulk
// Reporting API (https://youtubereporting.googleapis.com/v1), authenticated with
// a Google OAuth 2.0 refresh-token grant. This connector exposes the three core
// control-plane resources of that API as streams — reporting jobs, report types,
// and the generated reports for a job — which together drive the bulk report
// pipeline the upstream source reads from. It is read-only.
//
// Auth: client_id, client_secret, and refresh_token secrets are exchanged for a
// short-lived access token at the Google token endpoint (see auth.go); only the
// resolved access token is sent (as a Bearer header) to the Reporting API. The
// secret values are never logged.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package youtubeanalytics

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
	registryName          = "youtube-analytics"
	defaultBaseURL        = "https://youtubereporting.googleapis.com/v1"
	defaultTokenURL       = "https://oauth2.googleapis.com/token"
	defaultScope          = "https://www.googleapis.com/auth/yt-analytics.readonly"
	defaultPageSize       = 100
	maxPageSize           = 100
	userAgent             = "polymetrics-go-cli"
	pageTokenParam        = "pageToken"
	nextPageTokenPath     = "nextPageToken"
	contentOwnerParam     = "onBehalfOfContentOwner"
	fixtureCreateTimeBase = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the YouTube Analytics connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm YouTube Analytics connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the OAuth token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "YouTube Analytics",
		IntegrationType: "api",
		Description:     "Reads YouTube Reporting API jobs, report types, and generated reports via the Google OAuth 2.0 refresh-token grant.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true},
	}
}

// Check verifies the connector is configured well enough to talk to the YouTube
// Reporting API. In fixture mode it short-circuits without a network call.
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
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the reportTypes list confirms auth and connectivity
	// without mutating anything.
	query := url.Values{"pageSize": []string{"1"}}
	applyContentOwner(query, cfg)
	if err := r.DoJSON(ctx, http.MethodGet, "reportTypes", query, nil, nil); err != nil {
		return fmt.Errorf("check youtube-analytics: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: youtubeStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: each stream starts with an
// empty incremental cursor (full sync). The connector advertises only
// full_refresh upstream, but the cursor scaffolding is kept so an incremental
// mode can be layered on without an interface change.
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
		stream = "jobs"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("youtube-analytics stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	resourcePath, err := resolveResource(endpoint, req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := pageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := maxPages(req.Config)
	if err != nil {
		return err
	}

	base := url.Values{}
	base.Set("pageSize", strconv.Itoa(pageSize))
	applyContentOwner(base, req.Config)

	paginator := &connsdk.CursorPaginator{
		CursorParam: pageTokenParam,
		TokenPath:   nextPageTokenPath,
		FirstQuery:  base,
	}
	return connsdk.Harvest(ctx, r, http.MethodGet, resourcePath, nil, paginator, endpoint.recordsPath, maxPages, func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	})
}

// Write reports that this connector is read-only. The YouTube Reporting API has
// no safe reverse-ETL surface, so Capabilities.Write is false and Write always
// returns ErrUnsupportedOperation.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free (mirrors the
// stripe fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":            fmt.Sprintf("%s_fixture_%d", stream, i),
			"reportTypeId":  "channel_basic_a3",
			"name":          fmt.Sprintf("Fixture %s %d", stream, i),
			"createTime":    fixtureCreateTimeBase,
			"expireTime":    fixtureCreateTimeBase,
			"deprecateTime": "",
			"systemManaged": false,
			"jobId":         "job_fixture_1",
			"startTime":     fixtureCreateTimeBase,
			"endTime":       fixtureCreateTimeBase,
			"jobExpireTime": fixtureCreateTimeBase,
			"downloadUrl":   fmt.Sprintf("https://youtubereporting.googleapis.com/v1/media/fixture_%d", i),
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the Google OAuth refresh-token
// authenticator and the resolved base URL. Secrets only ever flow into the
// authenticator; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	tokenURL, err := tokenURL(cfg)
	if err != nil {
		return nil, err
	}
	clientID := secret(cfg, "client_id")
	clientSecret := secret(cfg, "client_secret")
	refreshToken := secret(cfg, "refresh_token")
	if strings.TrimSpace(clientID) == "" || strings.TrimSpace(refreshToken) == "" {
		return nil, errors.New("youtube-analytics connector requires secrets client_id and refresh_token")
	}
	auth := &oauthRefreshAuth{
		tokenURL:     tokenURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		refreshToken: refreshToken,
		scope:        defaultScope,
		client:       c.Client,
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: userAgent,
	}, nil
}

// resolveResource fills in the {jobId} segment for the reports stream from the
// job_id config; other streams use their static resource path.
func resolveResource(endpoint streamEndpoint, cfg connectors.RuntimeConfig) (string, error) {
	if !endpoint.needsJobID {
		return endpoint.resource, nil
	}
	jobID := strings.TrimSpace(cfg.Config["job_id"])
	if jobID == "" {
		return "", errors.New("youtube-analytics reports stream requires config job_id")
	}
	return fmt.Sprintf(endpoint.resource, url.PathEscape(jobID)), nil
}

// applyContentOwner adds the onBehalfOfContentOwner query parameter when a
// content_owner_id is configured (used by content-owner-scoped accounts).
func applyContentOwner(query url.Values, cfg connectors.RuntimeConfig) {
	if owner := strings.TrimSpace(cfg.Config["content_owner_id"]); owner != "" {
		query.Set(contentOwnerParam, owner)
	}
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

// baseURL resolves and validates the base URL. The default is the YouTube
// Reporting API host; any override must be an absolute http/https URL with a
// host to bound SSRF risk.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	return validatedURL(cfg.Config["base_url"], defaultBaseURL, "base_url")
}

// tokenURL resolves the OAuth token endpoint. A token_url override wins;
// otherwise the Google default is used.
func tokenURL(cfg connectors.RuntimeConfig) (string, error) {
	return validatedURL(cfg.Config["token_url"], defaultTokenURL, "token_url")
}

func validatedURL(raw, fallback, field string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, nil
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("youtube-analytics config %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("youtube-analytics config %s must use http or https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("youtube-analytics config %s must include a host", field)
	}
	return strings.TrimRight(raw, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("youtube-analytics config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("youtube-analytics config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("youtube-analytics config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("youtube-analytics config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
