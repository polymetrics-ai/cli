// Package dockerhub implements the native pm Docker Hub source connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit, modeled on
// the stripe reference connector.
//
// Docker Hub's public registry API (https://hub.docker.com/v2) requires no
// authentication for public repositories: the connector reads repositories, image
// tags, and the namespace/user profile for a configured docker_username. List
// endpoints return a {count,next,previous,results} envelope where `next` is an
// absolute URL; the connector follows it until null.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package dockerhub

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
	dockerhubDefaultBaseURL  = "https://hub.docker.com/v2"
	dockerhubDefaultPageSize = 100
	dockerhubMaxPageSize     = 100
	dockerhubUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("dockerhub", New)
}

// New returns the Docker Hub connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Docker Hub source connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "dockerhub" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "dockerhub",
		DisplayName:     "Docker Hub",
		IntegrationType: "api",
		Description:     "Reads public Docker Hub repositories, image tags, and namespace profiles for a configured username or organization via the Docker Hub registry API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Docker Hub.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := dockerhubBaseURL(cfg); err != nil {
		return err
	}
	username, err := dockerhubUsername(cfg)
	if err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the repositories list confirms the username resolves and
	// the API is reachable. The public API needs no credentials.
	q := url.Values{"page_size": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "repositories/"+username+"/", q, nil, nil); err != nil {
		return fmt.Errorf("check dockerhub: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: dockerhubStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Docker Hub stream starts
// with an empty incremental cursor (full sync). Docker Hub only supports full
// refresh, but exposing the cursor keeps the state shape consistent with other
// connectors.
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
		stream = "repositories"
	}
	def, ok := dockerhubStreamDefs[stream]
	if !ok {
		return fmt.Errorf("dockerhub stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, def, emit)
	}

	username, err := dockerhubUsername(req.Config)
	if err != nil {
		return err
	}
	repository := strings.TrimSpace(req.Config.Config["repository"])
	if def.requiresRepository && repository == "" {
		return fmt.Errorf("dockerhub stream %q requires config repository", stream)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	path := def.path(username, repository)

	switch def.kind {
	case kindSingle:
		return c.readSingle(ctx, r, path, def, emit)
	default:
		pageSize, err := dockerhubPageSize(req.Config)
		if err != nil {
			return err
		}
		maxPages, err := dockerhubMaxPages(req.Config)
		if err != nil {
			return err
		}
		return c.harvest(ctx, r, path, pageSize, maxPages, def, emit)
	}
}

// Write is unsupported: Docker Hub's public registry API exposes no safe
// reverse-ETL write actions, so this connector is read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// harvest drives Docker Hub's {count,next,previous,results} pagination. The first
// page uses the relative path with page/page_size; each subsequent page follows
// the absolute `next` URL verbatim (connsdk.Requester treats absolute http(s)
// paths as-is). There is no body-token paginator in connsdk for the absolute-next
// shape, so the loop lives here, built on Requester + RecordsAt + StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, pageSize, maxPages int, def streamDef, emit func(connectors.Record) error) error {
	next := path
	first := true
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var query url.Values
		if first {
			query = url.Values{
				"page":      []string{"1"},
				"page_size": []string{strconv.Itoa(pageSize)},
			}
			first = false
		}
		resp, err := r.Do(ctx, http.MethodGet, next, query, nil)
		if err != nil {
			return fmt.Errorf("read dockerhub %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "results")
		if err != nil {
			return fmt.Errorf("decode dockerhub %s page: %w", path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(def.mapRecord(item)); err != nil {
				return err
			}
		}
		nextURL, err := connsdk.StringAt(resp.Body, "next")
		if err != nil {
			return fmt.Errorf("decode dockerhub %s next: %w", path, err)
		}
		if strings.TrimSpace(nextURL) == "" {
			return nil
		}
		next = nextURL
	}
	return nil
}

// readSingle reads a single-object endpoint (namespace profile) into one record.
func (c Connector) readSingle(ctx context.Context, r *connsdk.Requester, path string, def streamDef, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return fmt.Errorf("read dockerhub %s: %w", path, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode dockerhub %s: %w", path, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(def.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise dockerhub credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, def streamDef, emit func(connectors.Record) error) error {
	count := 2
	if def.kind == kindSingle {
		count = 1
	}
	for i := 1; i <= count; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"name":                  fmt.Sprintf("%s-fixture-%d", stream, i),
			"namespace":             "upstream",
			"repository_type":       "image",
			"status":                int64(1),
			"status_description":    "active",
			"description":           fmt.Sprintf("Fixture %s record %d", stream, i),
			"is_private":            false,
			"star_count":            int64(i),
			"pull_count":            int64(1000 * i),
			"storage_size":          int64(2048 * i),
			"last_updated":          fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"last_modified":         fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"date_registered":       "2026-01-01T00:00:00Z",
			"id":                    fmt.Sprintf("id_fixture_%d", i),
			"uuid":                  fmt.Sprintf("uuid-fixture-%d", i),
			"orgname":               "upstream",
			"full_name":             "upstream",
			"company":               "upstream",
			"location":              "",
			"type":                  "Organization",
			"badge":                 "verified_publisher",
			"is_active":             true,
			"date_joined":           "2020-09-18T22:53:58.901735Z",
			"repository":            int64(i),
			"full_size":             int64(4096 * i),
			"digest":                fmt.Sprintf("sha256:fixture%d", i),
			"media_type":            "application/vnd.docker.distribution.manifest.v2+json",
			"content_type":          "image",
			"tag_status":            "active",
			"last_pushed":           fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"last_updater_username": "upstream",
		}
		if err := emit(def.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the resolved base URL. Docker
// Hub's public registry API needs no authenticator.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := dockerhubBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		UserAgent: dockerhubUserAgent,
	}, nil
}

// dockerhubUsername resolves and validates the required docker_username config.
func dockerhubUsername(cfg connectors.RuntimeConfig) (string, error) {
	username := strings.TrimSpace(cfg.Config["docker_username"])
	if username == "" {
		return "", errors.New("dockerhub connector requires config docker_username")
	}
	if !validUsername(username) {
		return "", fmt.Errorf("dockerhub config docker_username %q is invalid", username)
	}
	return username, nil
}

// validUsername matches Docker Hub's namespace pattern (lowercase alphanumerics,
// underscores, hyphens) and guards against path-injection in the request URL.
func validUsername(name string) bool {
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '_' || r == '-':
		default:
			return false
		}
	}
	return name != ""
}

// dockerhubBaseURL resolves and validates the base URL. The default is
// hub.docker.com/v2; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func dockerhubBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return dockerhubDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("dockerhub config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("dockerhub config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("dockerhub config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func dockerhubPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return dockerhubDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("dockerhub config page_size must be an integer: %w", err)
	}
	if value < 1 || value > dockerhubMaxPageSize {
		return 0, fmt.Errorf("dockerhub config page_size must be between 1 and %d", dockerhubMaxPageSize)
	}
	return value, nil
}

func dockerhubMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("dockerhub config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("dockerhub config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
