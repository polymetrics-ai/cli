// Package freightview implements the native pm Freightview connector. It follows
// the declarative-HTTP template established by the stripe connector: a thin
// package composing the connsdk toolkit (Requester + extraction) with
// Freightview-specific stream definitions and a session-token authenticator.
//
// Auth is the Freightview v2.0 "session token" flow: POST client_id +
// client_secret + grant_type=client_credentials to /auth/token, read
// access_token, then send Authorization: Bearer <token> on data requests. The
// token is cached and refreshed on demand. Secrets only ever flow into the
// authenticator; they are never logged.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package freightview

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	freightviewDefaultBaseURL = "https://api.freightview.com/v2.0"
	freightviewUserAgent      = "polymetrics-go-cli"
	// freightviewMaxPages bounds substream fan-out and root pagination so a
	// misbehaving server cannot loop forever; overridable via max_pages config.
	freightviewDefaultMaxPages = 0 // 0 = unlimited
	// freightviewFixtureDate is the deterministic timestamp used by fixture-mode
	// records (2026-01-01T00:00:00Z).
	freightviewFixtureDate = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("freightview", New)
}

// New returns the Freightview connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Freightview connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the session-token login. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "freightview" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "freightview",
		DisplayName:     "Freightview",
		IntegrationType: "api",
		Description:     "Reads Freightview shipments, quotes, and tracking events through the Freightview v2.0 REST API using the client-credentials session-token flow.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Freightview.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := freightviewBaseURL(cfg); err != nil {
		return err
	}
	if clientID, clientSecret := freightviewCreds(cfg); clientID == "" || clientSecret == "" {
		return errors.New("freightview connector requires secrets client_id and client_secret")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the shipments list confirms auth and connectivity.
	q := url.Values{"limit": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "shipments", q, nil, nil); err != nil {
		return fmt.Errorf("check freightview: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: freightviewStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "shipments"
	}
	def, ok := freightviewStreamDefs[stream]
	if !ok {
		return fmt.Errorf("freightview stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, def, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := freightviewMaxPages(req.Config)
	if err != nil {
		return err
	}
	if def.substream {
		return c.readSubstream(ctx, r, def, maxPages, emit)
	}
	return c.readShipments(ctx, r, def, maxPages, emit)
}

// readShipments drives Freightview's continuationToken pagination over the root
// /shipments endpoint. The response carries a continuationToken; when it is
// present (non-null) the next page is requested with continuationToken=<value>.
func (c Connector) readShipments(ctx context.Context, r *connsdk.Requester, def streamDef, maxPages int, emit func(connectors.Record) error) error {
	token := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		if token != "" {
			query.Set("continuationToken", token)
		}
		resp, err := r.Do(ctx, http.MethodGet, "shipments", query, nil)
		if err != nil {
			return fmt.Errorf("read freightview shipments: %w", err)
		}
		records, err := connsdk.RecordsAt(resp.Body, def.recordsPath)
		if err != nil {
			return fmt.Errorf("decode freightview shipments page: %w", err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(def.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "continuationToken")
		if err != nil {
			return fmt.Errorf("decode freightview continuationToken: %w", err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		token = next
	}
	return nil
}

// readSubstream fans out a per-shipment subresource (quotes, tracking). It first
// lists shipment ids via the root stream, then reads
// /shipments/{shipmentId}/<sub> for each.
func (c Connector) readSubstream(ctx context.Context, r *connsdk.Requester, def streamDef, maxPages int, emit func(connectors.Record) error) error {
	ids, err := c.shipmentIDs(ctx, r, maxPages)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return err
		}
		path := fmt.Sprintf("shipments/%s/%s", url.PathEscape(id), def.sub)
		resp, err := r.Do(ctx, http.MethodGet, path, nil, nil)
		if err != nil {
			return fmt.Errorf("read freightview %s for shipment %s: %w", def.name, id, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, def.recordsPath)
		if err != nil {
			return fmt.Errorf("decode freightview %s for shipment %s: %w", def.name, id, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(def.mapRecord(item)); err != nil {
				return err
			}
		}
	}
	return nil
}

// shipmentIDs collects the parent shipment ids needed by substreams, following
// continuationToken pagination.
func (c Connector) shipmentIDs(ctx context.Context, r *connsdk.Requester, maxPages int) ([]string, error) {
	var ids []string
	token := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		query := url.Values{}
		if token != "" {
			query.Set("continuationToken", token)
		}
		resp, err := r.Do(ctx, http.MethodGet, "shipments", query, nil)
		if err != nil {
			return nil, fmt.Errorf("list freightview shipments: %w", err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "shipments")
		if err != nil {
			return nil, fmt.Errorf("decode freightview shipments: %w", err)
		}
		for _, item := range records {
			if id := stringField(item, "shipmentId"); id != "" {
				ids = append(ids, id)
			}
		}
		next, err := connsdk.StringAt(resp.Body, "continuationToken")
		if err != nil {
			return nil, fmt.Errorf("decode freightview continuationToken: %w", err)
		}
		if strings.TrimSpace(next) == "" {
			break
		}
		token = next
	}
	return ids, nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise freightview credential-free.
func (c Connector) readFixture(ctx context.Context, def streamDef, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"shipmentId":  fmt.Sprintf("ship_fixture_%d", i),
			"quoteId":     fmt.Sprintf("quote_fixture_%d", i),
			"status":      "booked",
			"direction":   "outbound",
			"bookedBy":    "fixture@example.com",
			"quotedBy":    "fixture@example.com",
			"createdDate": freightviewFixtureDate,
			"bookedDate":  freightviewFixtureDate,
			"pickupDate":  freightviewFixtureDate,
			"isArchived":  false,
			"isLiveLoad":  false,
			"mode":        "ltl",
			"method":      "api",
			"amount":      float64(100 * i),
			"currency":    "USD",
			"summary":     "Shipment picked up",
			"eventType":   "pickup",
			"eventDate":   "2026-01-01",
			"eventTime":   "08:00",
		}
		if err := emit(def.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the session-token authenticator
// and the resolved base URL.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := freightviewBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	clientID, clientSecret := freightviewCreds(cfg)
	if clientID == "" || clientSecret == "" {
		return nil, errors.New("freightview connector requires secrets client_id and client_secret")
	}
	auth := &sessionTokenAuth{
		client:       c.Client,
		tokenURL:     base + "/auth/token",
		clientID:     clientID,
		clientSecret: clientSecret,
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: freightviewUserAgent,
	}, nil
}

// sessionTokenAuth implements connsdk.Authenticator for Freightview's
// client-credentials session-token flow. It POSTs the credentials to tokenURL,
// caches the returned access_token, and applies it as a Bearer header. Secrets
// are never logged.
type sessionTokenAuth struct {
	client       *http.Client
	tokenURL     string
	clientID     string
	clientSecret string

	mu      sync.Mutex
	token   string
	expires time.Time
}

func (a *sessionTokenAuth) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.accessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *sessionTokenAuth) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.token != "" && time.Now().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}

	body, err := json.Marshal(map[string]string{
		"client_id":     a.clientID,
		"client_secret": a.clientSecret,
		"grant_type":    "client_credentials",
	})
	if err != nil {
		return "", fmt.Errorf("freightview: encode token request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("freightview: build token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	client := a.client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("freightview: token request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("freightview: token endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		return "", fmt.Errorf("freightview: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("freightview: token response missing access_token")
	}
	a.token = out.AccessToken
	ttl := 24 * time.Hour
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = time.Now().Add(ttl)
	return a.token, nil
}

func freightviewCreds(cfg connectors.RuntimeConfig) (clientID, clientSecret string) {
	if cfg.Secrets == nil {
		return "", ""
	}
	return strings.TrimSpace(cfg.Secrets["client_id"]), strings.TrimSpace(cfg.Secrets["client_secret"])
}

// freightviewBaseURL resolves and validates the base URL. The default is
// api.freightview.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func freightviewBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return freightviewDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("freightview config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("freightview config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("freightview config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func freightviewMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return freightviewDefaultMaxPages, nil
	}
	value, err := parseNonNegativeInt(raw)
	if err != nil {
		return 0, fmt.Errorf("freightview config max_pages must be a non-negative integer, all, or unlimited: %w", err)
	}
	return value, nil
}

func parseNonNegativeInt(raw string) (int, error) {
	n := 0
	for _, r := range raw {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("invalid integer %q", raw)
		}
		n = n*10 + int(r-'0')
	}
	return n, nil
}

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

// Write satisfies the connectors.Connector interface. Freightview is read-only in
// this connector (no reverse-ETL writes are exposed), so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
