// Package ip2whois implements the native pm IP2WHOIS connector. It follows the
// declarative-HTTP template established by the stripe package: a thin package
// that composes the connsdk toolkit (Requester + APIKeyQuery auth + JSON
// extraction) with IP2WHOIS-specific stream definitions and record mappers.
//
// IP2WHOIS exposes a single domain-lookup endpoint
// (https://api.ip2whois.com/v2?key=<key>&domain=<domain>) that returns one
// WHOIS object per domain rather than a paginated list. The connector therefore
// iterates across a configured set of domains, performing one lookup each, and
// fans the resulting object out into the whois, nameservers, and contacts
// streams.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package ip2whois

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	ip2whoisDefaultBaseURL = "https://api.ip2whois.com/v2"
	ip2whoisUserAgent      = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("ip2whois", New)
}

// New returns the IP2WHOIS connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm IP2WHOIS connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "ip2whois" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "ip2whois",
		DisplayName:     "IP2WHOIS",
		IntegrationType: "api",
		Description:     "Looks up WHOIS records for configured domains via the IP2WHOIS API, exposing flattened whois, nameservers, and contacts streams.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to IP2WHOIS. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := ip2whoisBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(ip2whoisSecret(cfg)) == "" {
		return errors.New("ip2whois connector requires secret api_key")
	}
	domains := configuredDomains(cfg)
	if len(domains) == 0 {
		return errors.New("ip2whois connector requires at least one domain (config domain or domains)")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A single bounded lookup of the first domain confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "", url.Values{"domain": []string{domains[0]}}, nil, nil); err != nil {
		return fmt.Errorf("check ip2whois: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: ip2whoisStreams()}, nil
}

// Write is unsupported: IP2WHOIS is a read-only lookup API.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "whois"
	}
	fanout, ok := streamFanout[stream]
	if !ok {
		return fmt.Errorf("ip2whois stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, fanout, req, emit)
	}

	domains := configuredDomains(req.Config)
	if len(domains) == 0 {
		return errors.New("ip2whois connector requires at least one domain (config domain or domains)")
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, fanout, domains, emit)
}

// streamFanout maps each stream to the function that fans a single raw lookup
// object out into that stream's records. whois yields exactly one record per
// domain; nameservers and contacts yield zero or more.
var streamFanout = map[string]func(map[string]any) []connectors.Record{
	"whois":       func(item map[string]any) []connectors.Record { return []connectors.Record{whoisRecord(item)} },
	"nameservers": nameserverRecords,
	"contacts":    contactRecords,
}

// harvest performs one lookup per configured domain. IP2WHOIS has no list
// pagination; iterating across the domain set is the connector's natural
// "pages". Each lookup returns a single WHOIS object that the stream's fan-out
// function turns into records.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, fanout func(map[string]any) []connectors.Record, domains []string, emit func(connectors.Record) error) error {
	for _, domain := range domains {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{"domain": []string{domain}}
		resp, err := r.Do(ctx, http.MethodGet, "", query, nil)
		if err != nil {
			return fmt.Errorf("read ip2whois domain %q: %w", domain, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode ip2whois domain %q: %w", domain, err)
		}
		for _, item := range records {
			for _, rec := range fanout(map[string]any(item)) {
				if err := emit(rec); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise ip2whois credential-free.
func (c Connector) readFixture(ctx context.Context, fanout func(map[string]any) []connectors.Record, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	domains := configuredDomains(req.Config)
	if len(domains) == 0 {
		domains = []string{"example.com", "example.org"}
	}
	for _, domain := range domains {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := fixtureLookup(domain)
		for _, rec := range fanout(item) {
			if cursor := req.State["cursor"]; cursor != "" {
				rec["previous_cursor"] = cursor
			}
			if err := emit(rec); err != nil {
				return err
			}
		}
	}
	return nil
}

// fixtureLookup builds a deterministic raw IP2WHOIS lookup object for a domain.
func fixtureLookup(domain string) map[string]any {
	return map[string]any{
		"domain":       domain,
		"domain_id":    domain + "-fixture",
		"status":       "clientTransferProhibited",
		"create_date":  "2020-01-01T00:00:00Z",
		"update_date":  "2026-01-01T00:00:00Z",
		"expire_date":  "2030-01-01T00:00:00Z",
		"domain_age":   int64(2190),
		"whois_server": "whois.fixture-registrar.com",
		"registrar":    map[string]any{"iana_id": "9999", "name": "Fixture Registrar", "url": "https://registrar.fixture"},
		"registrant":   map[string]any{"name": "Fixture Registrant", "organization": "Fixture Org", "email": "registrant@" + domain, "country": "US"},
		"admin":        map[string]any{"name": "Fixture Admin", "email": "admin@" + domain, "country": "US"},
		"tech":         map[string]any{"name": "Fixture Tech", "email": "tech@" + domain, "country": "US"},
		"billing":      map[string]any{"name": "Fixture Billing", "email": "billing@" + domain, "country": "US"},
		"nameservers":  []any{"ns1." + domain, "ns2." + domain},
	}
}

// requester builds a connsdk.Requester wired with API-key query auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyQuery; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := ip2whoisBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := ip2whoisSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("ip2whois connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery("key", secret),
		UserAgent: ip2whoisUserAgent,
	}, nil
}

func ip2whoisSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// configuredDomains resolves the set of domains to look up from config. It
// accepts a comma-separated "domains" list and/or a single "domain" field,
// de-duplicating while preserving order.
func configuredDomains(cfg connectors.RuntimeConfig) []string {
	if cfg.Config == nil {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0)
	add := func(raw string) {
		d := strings.TrimSpace(raw)
		if d == "" || seen[d] {
			return
		}
		seen[d] = true
		out = append(out, d)
	}
	for _, part := range strings.Split(cfg.Config["domains"], ",") {
		add(part)
	}
	add(cfg.Config["domain"])
	return out
}

// ip2whoisBaseURL resolves and validates the base URL. The default is
// api.ip2whois.com/v2; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func ip2whoisBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return ip2whoisDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("ip2whois config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("ip2whois config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("ip2whois config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
