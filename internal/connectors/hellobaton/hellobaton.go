// Package hellobaton implements the native pm Hellobaton source connector. It is
// a declarative-HTTP per-system connector that composes the connsdk toolkit
// (Requester + api_key query authentication + RecordsAt extraction) with
// Hellobaton-specific stream definitions and endpoints. It follows the same
// shape as the stripe reference connector.
//
// Hellobaton is a customer-onboarding / project-management API. Its base URL is
// derived from the customer's company name (https://<company>.hellobaton.com/api/)
// and it authenticates with an api_key query parameter. List endpoints are
// Django-REST-Framework style: {count, next, previous, results:[...]} where
// `next` is the absolute URL of the following page. Only full-refresh reads are
// supported upstream, so the connector is read-only.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package hellobaton

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
	hellobatonDefaultPageSize = 100
	hellobatonMaxPageSize     = 100
	hellobatonUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("hellobaton", New)
}

// New returns the Hellobaton connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Hellobaton connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "hellobaton" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "hellobaton",
		DisplayName:     "Hellobaton",
		IntegrationType: "api",
		Description:     "Reads Hellobaton projects, milestones, tasks, phases, companies, and users through the Hellobaton REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Hellobaton.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := hellobatonBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(hellobatonSecret(cfg)) == "" {
		return errors.New("hellobaton connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the projects list confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "projects", url.Values{"page_size": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check hellobaton: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: hellobatonStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "projects"
	}
	endpoint, ok := hellobatonStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("hellobaton stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := hellobatonPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := hellobatonMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Hellobaton's DRF-style pagination. List responses are
// {count, next, previous, results:[...]} where `next` is the absolute URL of the
// next page (or null when exhausted). connsdk.Requester treats an http(s) path
// as absolute, so each page after the first re-requests the `next` URL directly.
// The api_key authenticator re-injects the credential on every request.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := endpoint.resource
	query := url.Values{}
	query.Set("page_size", strconv.Itoa(pageSize))

	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read hellobaton %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "results")
		if err != nil {
			return fmt.Errorf("decode hellobaton %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "next")
		if err != nil {
			return fmt.Errorf("decode hellobaton %s next: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		if next == "" {
			return nil
		}
		// Subsequent pages are addressed by the absolute next URL; page_size is
		// already encoded there, so clear the explicit query to avoid duplication.
		path = next
		query = url.Values{}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise hellobaton credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                    i,
			"name":                  fmt.Sprintf("%s fixture %d", strings.TrimSuffix(stream, "s"), i),
			"created":               "2026-01-01T00:00:00Z",
			"modified":              "2026-01-02T00:00:00Z",
			"_self":                 fmt.Sprintf("/api/%s/%d", endpoint.resource, i),
			"archived":              false,
			"completed_datetime":    "2026-01-03T00:00:00Z",
			"cost":                  int64(1000 * i),
			"annual_contract_value": "10000",
			"creator":               "fixture@example.com",
			"description":           fmt.Sprintf("fixture %s %d", stream, i),
			"deadline_datetime":     "2026-02-01T00:00:00Z",
			"deadline_fixed":        true,
			"duration":              int64(i),
			"finish_datetime":       "2026-02-02T00:00:00Z",
			"project":               "fixture project",
			"first_name":            fmt.Sprintf("First%d", i),
			"last_name":             fmt.Sprintf("Last%d", i),
		}
		record := endpoint.mapRecord(item)
		record["connector"] = "hellobaton"
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

// requester builds a connsdk.Requester wired with api_key query authentication
// and the resolved base URL. The secret only ever flows into
// connsdk.APIKeyQuery; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := hellobatonBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := hellobatonSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("hellobaton connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery("api_key", secret),
		UserAgent: hellobatonUserAgent,
	}, nil
}

func hellobatonSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// hellobatonBaseURL resolves and validates the base URL. By default it is built
// from the company config field (https://<company>.hellobaton.com/api). An
// explicit base_url override takes precedence; any override (or derived URL)
// must be an absolute https (or http for local test servers) URL with a host to
// bound SSRF risk.
func hellobatonBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		company := strings.TrimSpace(cfg.Config["company"])
		if company == "" {
			return "", errors.New("hellobaton connector requires config company or base_url")
		}
		if !validCompany(company) {
			return "", fmt.Errorf("hellobaton config company %q is invalid", company)
		}
		base = "https://" + company + ".hellobaton.com/api"
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("hellobaton config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("hellobaton config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("hellobaton config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

// validCompany guards the company-derived subdomain against injection: only
// letters, digits, and hyphens are allowed (DNS label charset).
func validCompany(company string) bool {
	for _, r := range company {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-':
		default:
			return false
		}
	}
	return true
}

func hellobatonPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return hellobatonDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("hellobaton config page_size must be an integer: %w", err)
	}
	if value < 1 || value > hellobatonMaxPageSize {
		return 0, fmt.Errorf("hellobaton config page_size must be between 1 and %d", hellobatonMaxPageSize)
	}
	return value, nil
}

func hellobatonMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("hellobaton config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("hellobaton config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write satisfies the connectors.Connector interface. Hellobaton is read-only
// (the upstream API exposes no safe reverse-ETL writes), so writes are rejected.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
