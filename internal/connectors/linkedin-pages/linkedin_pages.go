// Package linkedinpages implements the native pm LinkedIn Pages connector. It
// follows the declarative-HTTP shape established by the stripe connector (and its
// LinkedIn Ads sibling): a thin package that composes the connsdk toolkit
// (Requester + Bearer auth + RecordsAt extraction) with LinkedIn Community
// Management API stream definitions and endpoints.
//
// LinkedIn's Pages (Community Management) API has a few traits versus a vanilla
// bearer API:
//   - Every request must carry a LinkedIn-Version header (e.g. 202601) and the
//     X-Restli-Protocol-Version: 2.0.0 header.
//   - Streams are heterogeneous: the organization lookup and the network-size
//     follower count are single objects, while follower/share statistics are
//     finder endpoints (q=organizationalEntity) returning {"elements":[...]}
//     paged with start/count offset parameters.
//   - Everything is scoped to one organization id (org_id), which is treated as a
//     secret and stamped onto every emitted record.
//   - The access token is a member OAuth2 token; this connector reads the
//     long-lived credentials.access_token secret directly (the refresh_token
//     grant exchange is left to the operator/agent layer). client_id,
//     client_secret, and refresh_token are accepted as secrets but only the
//     resolved access_token is used here.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package linkedinpages

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
	registryName = "linkedin-pages"

	linkedinDefaultBaseURL = "https://api.linkedin.com/rest"
	// linkedinDefaultVersion is the LinkedIn-Version header sent when the operator
	// does not override it via config. LinkedIn versions are monthly (YYYYMM).
	linkedinDefaultVersion = "202601"
	linkedinRestliVersion  = "2.0.0"
	linkedinUserAgent      = "polymetrics-go-cli"

	linkedinDefaultPageSize = 100
	linkedinMaxPageSize     = 1000

	// linkedinFixtureFollowers is the deterministic follower count used by
	// fixture-mode records.
	linkedinFixtureFollowers int64 = 12345
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the LinkedIn Pages connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm LinkedIn Pages connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "LinkedIn Pages",
		IntegrationType: "api",
		Description:     "Reads LinkedIn organization (company page) profile, follower statistics, share statistics, and total follower count through the LinkedIn Community Management REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to LinkedIn. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := linkedinBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(linkedinSecret(cfg)) == "" {
		return errors.New("linkedin-pages connector requires secret credentials.access_token")
	}
	orgID := linkedinOrgID(cfg)
	if orgID == "" {
		return errors.New("linkedin-pages connector requires secret org_id")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the organization profile confirms auth, the org id, and
	// connectivity without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "organizations/"+url.PathEscape(orgID), nil, nil, nil); err != nil {
		return fmt.Errorf("check linkedin-pages: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: linkedinPagesStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: LinkedIn Pages streams are
// full-refresh only (lifetime statistics), so they start with an empty cursor.
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
		stream = "organizations"
	}
	endpoint, ok := linkedinPagesStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("linkedin-pages stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	orgID := linkedinOrgID(req.Config)
	if orgID == "" {
		return errors.New("linkedin-pages connector requires secret org_id")
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := linkedinMaxPages(req.Config)
	if err != nil {
		return err
	}

	// Stamp org_id onto every emitted record so downstream records are
	// self-describing (statistics elements do not always echo the entity).
	stamp := func(rec connectors.Record) error {
		rec["org_id"] = orgID
		return emit(rec)
	}

	if endpoint.paged {
		pageSize, err := linkedinPageSize(req.Config)
		if err != nil {
			return err
		}
		return c.harvest(ctx, r, endpoint, orgID, pageSize, maxPages, stamp)
	}
	return c.readSingle(ctx, r, endpoint, orgID, stamp)
}

// orgURN renders the organizationalEntity urn for an org id.
func orgURN(orgID string) string { return "urn:li:organization:" + orgID }

// buildPath/buildQuery assemble the request path and query for an endpoint.
func buildPath(endpoint streamEndpoint, orgID string) string {
	switch {
	case endpoint.idInPath:
		return endpoint.resource + "/" + url.PathEscape(orgID)
	case endpoint.entityInPath:
		return endpoint.resource + "/" + url.PathEscape(orgURN(orgID))
	default:
		return endpoint.resource
	}
}

func buildQuery(endpoint streamEndpoint, orgID string) url.Values {
	q := url.Values{}
	for k, vs := range endpoint.extraQuery {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	if endpoint.finder != "" {
		q.Set("q", endpoint.finder)
	}
	if endpoint.entityInQuery {
		q.Set("organizationalEntity", orgURN(orgID))
	}
	return q
}

// readSingle reads a single-object endpoint (organization lookup, network size)
// and emits its flattened record.
func (c Connector) readSingle(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, orgID string, emit func(connectors.Record) error) error {
	path := buildPath(endpoint, orgID)
	query := buildQuery(endpoint, orgID)
	resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		return fmt.Errorf("read linkedin-pages %s: %w", endpoint.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return fmt.Errorf("decode linkedin-pages %s: %w", endpoint.resource, err)
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

// harvest drives LinkedIn's start/count offset pagination over an elements[]
// finder endpoint. A page shorter than count signals the last page. The loop is
// built on connsdk.Requester + connsdk.RecordsAt so it shares retry/rate-limit
// handling.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, orgID string, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := buildPath(endpoint, orgID)
	base := buildQuery(endpoint, orgID)
	start := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("start", strconv.Itoa(start))
		query.Set("count", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read linkedin-pages %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode linkedin-pages %s page: %w", endpoint.resource, err)
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
		start += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise linkedin-pages credential-free (mirrors
// stripe's fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	orgID := linkedinOrgID(req.Config)
	if orgID == "" {
		orgID = "1000"
	}
	entity := orgURN(orgID)

	var items []map[string]any
	switch stream {
	case "organizations":
		items = []map[string]any{{
			"id":                      1000,
			"$URN":                    entity,
			"vanityName":              "fixture-org",
			"localizedName":           "Fixture Organization",
			"localizedWebsite":        "https://fixture.example",
			"primaryOrganizationType": "NONE",
			"organizationType":        "SELF_OWNED",
			"versionTag":              "1",
			"staffCountRange":         "SIZE_2_TO_10",
			"name":                    map[string]any{"localized": map[string]any{"en_US": "Fixture Organization"}},
			"locations":               []any{},
			"industries":              []any{"urn:li:industry:4"},
		}}
	case "follower_statistics":
		items = []map[string]any{{
			"organizationalEntity": entity,
			"followerCountsByAssociationType": []any{
				map[string]any{"associationType": "SPONSORED", "followerCounts": map[string]any{"organicFollowerGain": 5, "paidFollowerGain": 1}},
			},
			"followerCountsByCountry": []any{
				map[string]any{"country": "urn:li:country:us", "followerCounts": map[string]any{"organicFollowerGain": 3, "paidFollowerGain": 0}},
			},
		}}
	case "share_statistics":
		items = []map[string]any{{
			"organizationalEntity": entity,
			"totalShareStatistics": map[string]any{
				"impressionCount":        1000,
				"clickCount":             50,
				"likeCount":              20,
				"commentCount":           5,
				"shareCount":             3,
				"engagement":             0.078,
				"uniqueImpressionsCount": 800,
			},
		}}
	case "total_follower_count":
		items = []map[string]any{{
			"firstDegreeSize": linkedinFixtureFollowers,
		}}
	default:
		return fmt.Errorf("linkedin-pages stream %q not found", stream)
	}

	for _, item := range items {
		if err := ctx.Err(); err != nil {
			return err
		}
		record := endpoint.mapRecord(item)
		record["org_id"] = orgID
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer auth, the resolved base
// URL, and the mandatory LinkedIn-Version + X-Restli-Protocol-Version headers.
// The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := linkedinBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := linkedinSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("linkedin-pages connector requires secret credentials.access_token")
	}
	headers := map[string]string{
		"LinkedIn-Version":          linkedinVersion(cfg),
		"X-Restli-Protocol-Version": linkedinRestliVersion,
	}
	return &connsdk.Requester{
		Client:         c.Client,
		BaseURL:        base,
		Auth:           connsdk.Bearer(secret),
		UserAgent:      linkedinUserAgent,
		DefaultHeaders: headers,
	}, nil
}

// linkedinSecret resolves the access token from the credentials.access_token
// secret. The OAuth2 path stores the bare token under the same dotted key after
// the refresh exchange, so a single lookup covers both auth methods.
func linkedinSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	if v := strings.TrimSpace(cfg.Secrets["credentials.access_token"]); v != "" {
		return v
	}
	// Tolerate the un-prefixed key for convenience.
	return strings.TrimSpace(cfg.Secrets["access_token"])
}

// linkedinOrgID resolves the organization id. It is declared a secret in the
// catalog, so the secret map wins; config is tolerated as a fallback.
func linkedinOrgID(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets != nil {
		if v := strings.TrimSpace(cfg.Secrets["org_id"]); v != "" {
			return v
		}
	}
	if cfg.Config != nil {
		return strings.TrimSpace(cfg.Config["org_id"])
	}
	return ""
}

func linkedinVersion(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return linkedinDefaultVersion
	}
	if v := strings.TrimSpace(cfg.Config["linkedin_version"]); v != "" {
		return v
	}
	return linkedinDefaultVersion
}

// linkedinBaseURL resolves and validates the base URL. The default is
// api.linkedin.com/rest; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func linkedinBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return linkedinDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("linkedin-pages config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("linkedin-pages config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("linkedin-pages config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func linkedinPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return linkedinDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("linkedin-pages config page_size must be an integer: %w", err)
	}
	if value < 1 || value > linkedinMaxPageSize {
		return 0, fmt.Errorf("linkedin-pages config page_size must be between 1 and %d", linkedinMaxPageSize)
	}
	return value, nil
}

func linkedinMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("linkedin-pages config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("linkedin-pages config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func cloneValues(in url.Values) url.Values {
	out := url.Values{}
	for k, vs := range in {
		for _, v := range vs {
			out.Add(k, v)
		}
	}
	return out
}

// Write satisfies the connectors.Connector interface. LinkedIn Pages is read-only
// in pm: there is no approved reverse-ETL write surface, so any write is rejected
// as an unsupported operation.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{RecordsFailed: len(records)}, connectors.ErrUnsupportedOperation
}
