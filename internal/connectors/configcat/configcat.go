// Package configcat implements the native pm ConfigCat connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit: an HTTP
// Basic authenticator over the ConfigCat Public Management API, root-array record
// extraction, and a product fan-out for the nested config/environment/tag
// resources.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
//
// The ConfigCat Public Management API is read-only here (reverse ETL writes are
// not exposed): see https://api.configcat.com/docs/.
package configcat

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
	configcatDefaultBaseURL = "https://api.configcat.com"
	configcatUserAgent      = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("configcat", New)
}

// New returns the ConfigCat connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm ConfigCat connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "configcat" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "configcat",
		DisplayName:     "ConfigCat",
		IntegrationType: "api",
		Description:     "Reads ConfigCat organizations, products, configs, environments, and tags through the ConfigCat Public Management API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to ConfigCat. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := configcatBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(configcatPassword(cfg)) == "" {
		return errors.New("configcat connector requires secret password")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the organizations list confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "/v1/organizations", nil, nil, nil); err != nil {
		return fmt.Errorf("check configcat: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: configcatStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "products"
	}
	def, ok := configcatStreamDefs[stream]
	if !ok {
		return fmt.Errorf("configcat stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, def, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	if def.nestedUnderProduct {
		return c.readNested(ctx, r, def, emit)
	}
	return c.readList(ctx, r, def.path, def.mapRecord, "", emit)
}

// readList reads one ConfigCat list endpoint (a root JSON array) and emits each
// element through the stream's mapper. productID, when non-empty, is injected as
// product_id on every record so fan-out records carry their owning product even
// when the API payload omits the embedded product object.
func (c Connector) readList(ctx context.Context, r *connsdk.Requester, path string, mapRecord func(map[string]any) connectors.Record, productID string, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return fmt.Errorf("read configcat %s: %w", path, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode configcat %s: %w", path, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := mapRecord(item)
		if productID != "" {
			if rec["product_id"] == nil {
				rec["product_id"] = productID
			}
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

// readNested fans a nested resource (configs/environments/tags) out across every
// accessible product: it first lists products, then reads the resource for each
// product id and annotates every record with that product_id.
func (c Connector) readNested(ctx context.Context, r *connsdk.Requester, def streamDef, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, "/v1/products", nil, nil)
	if err != nil {
		return fmt.Errorf("read configcat products for %s fan-out: %w", def.name, err)
	}
	products, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode configcat products for %s fan-out: %w", def.name, err)
	}
	for _, product := range products {
		if err := ctx.Err(); err != nil {
			return err
		}
		productID := stringField(product, "productId")
		if productID == "" {
			continue
		}
		path := fmt.Sprintf(def.path, url.PathEscape(productID))
		if err := c.readList(ctx, r, path, def.mapRecord, productID, emit); err != nil {
			return err
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise configcat credential-free.
func (c Connector) readFixture(ctx context.Context, def streamDef, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"organizationId":    fmt.Sprintf("org_fixture_%d", i),
			"productId":         fmt.Sprintf("prod_fixture_%d", i),
			"configId":          fmt.Sprintf("config_fixture_%d", i),
			"environmentId":     fmt.Sprintf("env_fixture_%d", i),
			"tagId":             i,
			"name":              fmt.Sprintf("Fixture %d", i),
			"description":       fmt.Sprintf("fixture %s %d", def.name, i),
			"color":             "panther",
			"order":             i,
			"reasonRequired":    false,
			"approveRequired":   false,
			"evaluationVersion": "v2",
			"product":           map[string]any{"productId": fmt.Sprintf("prod_fixture_%d", i)},
			"organization":      map[string]any{"organizationId": fmt.Sprintf("org_fixture_%d", i)},
		}
		rec := def.mapRecord(item)
		rec["fixture"] = true
		rec["connector"] = "configcat"
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with HTTP Basic auth and the
// resolved base URL. The secret only ever flows into connsdk.Basic; it is never
// logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := configcatBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	password := configcatPassword(cfg)
	if strings.TrimSpace(password) == "" {
		return nil, errors.New("configcat connector requires secret password")
	}
	username := configcatUsername(cfg)
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(username, password),
		UserAgent: configcatUserAgent,
	}, nil
}

// configcatPassword resolves the Basic auth password from secrets.
func configcatPassword(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["password"]
}

// configcatUsername resolves the Basic auth username. ConfigCat's Public API
// username is not secret; it may live in config or (defensively) in secrets.
func configcatUsername(cfg connectors.RuntimeConfig) string {
	if cfg.Config != nil {
		if v := strings.TrimSpace(cfg.Config["username"]); v != "" {
			return v
		}
	}
	if cfg.Secrets != nil {
		return strings.TrimSpace(cfg.Secrets["username"])
	}
	return ""
}

// configcatBaseURL resolves and validates the base URL. The default is
// api.configcat.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func configcatBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return configcatDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("configcat config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("configcat config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("configcat config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
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

// Write is unsupported: the ConfigCat connector is read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
