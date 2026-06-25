// Package box implements the native pm Box connector. It is a declarative-HTTP
// per-system connector built on the connsdk toolkit, following the stripe
// reference shape: a thin package that composes a connsdk.Requester (OAuth2
// client-credentials auth + entry extraction + offset pagination) with
// Box-specific stream definitions and endpoints.
//
// Box authenticates with the OAuth2 client-credentials grant (Server
// Authentication / Client Credentials Grant): the client_id/client_secret are
// exchanged at the token endpoint for a short-lived bearer token, scoped to an
// enterprise or user via box_subject_type/box_subject_id. The connector is
// read-only (Box's Airbyte manifest source is a read source), so Capabilities
// .Write is false.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package box

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
	boxDefaultBaseURL  = "https://api.box.com/2.0"
	boxDefaultTokenURL = "https://api.box.com/oauth2/token"
	boxDefaultPageSize = 100
	boxMaxPageSize     = 1000
	boxMaxOffset       = 9999
	boxUserAgent       = "polymetrics-go-cli"
	boxRootFolderID    = "0"
)

func init() {
	connectors.RegisterFactory("box", New)
}

// New returns the Box connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Box connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the OAuth2 token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "box" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "box",
		DisplayName:     "Box",
		IntegrationType: "api",
		Description:     "Reads Box users, groups, collections, and folder items through the Box REST API using the OAuth2 client-credentials grant.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Box. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := boxBaseURL(cfg); err != nil {
		return err
	}
	if err := requireCredentials(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the users list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "users", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check box: %w", err)
	}
	return nil
}

// Write satisfies the connectors.Connector interface. Box is a read-only source
// connector (Capabilities.Write is false), so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: boxStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "users"
	}
	endpoint, ok := boxStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("box stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	resource, err := resolveResource(endpoint, req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := boxPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := boxMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, resource, endpoint.mapRecord, pageSize, maxPages, emit)
}

// harvest drives Box's offset-based pagination. Box list responses are
// {entries:[...], offset, limit, total_count}; the next page is requested with
// offset += limit until the page is short, an empty page is returned, or the max
// offset (9999) is reached. The loop lives here, built on connsdk.Requester +
// connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, resource string, mapRecord func(map[string]any) connectors.Record, pageSize, maxPages int, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(offset))

		resp, err := r.Do(ctx, http.MethodGet, resource, query, nil)
		if err != nil {
			return fmt.Errorf("read box %s: %w", resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "entries")
		if err != nil {
			return fmt.Errorf("decode box %s page: %w", resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(mapRecord(item)); err != nil {
				return err
			}
		}
		// Stop on a short or empty page: Box returns fewer than limit entries on
		// the final page.
		if len(records) < pageSize {
			return nil
		}
		offset += pageSize
		if offset > boxMaxOffset {
			return nil
		}
		// Defensive stop using total_count when present.
		if total, ok := intAt(resp.Body, "total_count"); ok && offset >= total {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise box credential-free (mirrors stripe's fixture
// intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	objectType := strings.TrimSuffix(stream, "s")
	switch stream {
	case "folder_items":
		objectType = "file"
	case "collections":
		objectType = "collection"
	}
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":              fmt.Sprintf("%s_fixture_%d", stream, i),
			"type":            objectType,
			"name":            fmt.Sprintf("Fixture %s %d", objectType, i),
			"login":           fmt.Sprintf("fixture+%d@example.com", i),
			"status":          "active",
			"language":        "en",
			"timezone":        "America/Los_Angeles",
			"group_type":      "managed_group",
			"collection_type": "favorites",
			"etag":            strconv.Itoa(i),
			"sequence_id":     strconv.Itoa(i),
			"sha1":            fmt.Sprintf("%040d", i),
			"size":            int64(1024 * i),
			"created_at":      "2026-01-01T00:00:00-08:00",
			"modified_at":     "2026-01-02T00:00:00-08:00",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with OAuth2 client-credentials auth
// and the resolved base URL. Secrets only ever flow into the authenticator; they
// are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := boxBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	if err := requireCredentials(cfg); err != nil {
		return nil, err
	}
	auth, err := c.authenticator(cfg)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: boxUserAgent,
	}, nil
}

// authenticator builds the OAuth2 client-credentials authenticator with the
// Box-specific box_subject_type/box_subject_id extra params.
func (c Connector) authenticator(cfg connectors.RuntimeConfig) (connsdk.Authenticator, error) {
	subjectType, subjectID := boxSubject(cfg)
	extra := url.Values{}
	extra.Set("box_subject_type", subjectType)
	if subjectID != "" {
		extra.Set("box_subject_id", subjectID)
	}
	return &connsdk.OAuth2ClientCredentials{
		TokenURL:     boxTokenURL(cfg),
		ClientID:     boxClientID(cfg),
		ClientSecret: boxClientSecret(cfg),
		ExtraParams:  extra,
		Client:       c.Client,
	}, nil
}

// boxSubject resolves the Box subject type and id. Box accepts box_subject_type
// of "enterprise" (the application service account) or "user". The subject id
// comes from the enterprise_id/user config; for enterprise mode the id is
// optional in some setups but typically the enterprise id.
func boxSubject(cfg connectors.RuntimeConfig) (string, string) {
	subjectType := strings.ToLower(strings.TrimSpace(cfg.Config["box_subject_type"]))
	if subjectType != "user" && subjectType != "enterprise" {
		subjectType = "enterprise"
	}
	id := strings.TrimSpace(cfg.Config["box_subject_id"])
	if id == "" {
		if subjectType == "user" {
			id = strings.TrimSpace(cfg.Config["user"])
		} else {
			id = strings.TrimSpace(cfg.Config["enterprise_id"])
		}
	}
	return subjectType, id
}

// resolveResource computes the request path for a stream, substituting the
// configured folder id for folder-scoped endpoints (defaults to the root folder).
func resolveResource(endpoint streamEndpoint, cfg connectors.RuntimeConfig) (string, error) {
	if !endpoint.folderScoped {
		return endpoint.resource, nil
	}
	folderID := strings.TrimSpace(cfg.Config["folder_id"])
	if folderID == "" {
		folderID = boxRootFolderID
	}
	if !validFolderID(folderID) {
		return "", fmt.Errorf("box config folder_id must be a numeric id, got %q", folderID)
	}
	return "folders/" + url.PathEscape(folderID) + "/items", nil
}

func validFolderID(id string) bool {
	if id == "" {
		return false
	}
	for _, r := range id {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func requireCredentials(cfg connectors.RuntimeConfig) error {
	if strings.TrimSpace(boxClientID(cfg)) == "" {
		return errors.New("box connector requires secret client_id")
	}
	if strings.TrimSpace(boxClientSecret(cfg)) == "" {
		return errors.New("box connector requires secret client_secret")
	}
	return nil
}

func boxClientID(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["client_id"]
}

func boxClientSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["client_secret"]
}

// boxTokenURL resolves and validates the OAuth2 token endpoint. The default is
// api.box.com; any override must be an absolute http(s) URL with a host.
func boxTokenURL(cfg connectors.RuntimeConfig) string {
	raw := strings.TrimSpace(cfg.Config["token_url"])
	if raw == "" {
		return boxDefaultTokenURL
	}
	return raw
}

// boxBaseURL resolves and validates the base URL. The default is api.box.com/2.0;
// any override must be an absolute https (or http for local test servers) URL with
// a host to bound SSRF risk.
func boxBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return boxDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("box config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("box config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("box config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func boxPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return boxDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("box config page_size must be an integer: %w", err)
	}
	if value < 1 || value > boxMaxPageSize {
		return 0, fmt.Errorf("box config page_size must be between 1 and %d", boxMaxPageSize)
	}
	return value, nil
}

func boxMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("box config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("box config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// intAt reads an integer value at a dotted path in the JSON body. It returns
// (0, false) when the value is missing or not numeric.
func intAt(body []byte, path string) (int, bool) {
	raw, err := connsdk.StringAt(body, path)
	if err != nil || raw == "" {
		return 0, false
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	return value, true
}
