// Package openrouter implements the native pm OpenRouter connector. OpenRouter
// (https://openrouter.ai/api/v1) is an OpenAI-compatible gateway to many model
// vendors; this connector makes it a first-class, vault-backed LLM provider for
// pm (used by the RLM router's Tier-2 classifier and as a documented credential
// source for the RLM agent container).
//
// It is a thin declarative-HTTP connector modeled on nebius-ai: connsdk Requester
// + Bearer auth + RecordsAt extraction. It self-registers via RegisterFactory in
// init(); registryset blank-imports it in the production binary.
//
// The directory/system name is "openrouter"; the Go package identifier is
// "openrouter".
package openrouter

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
	openrouterDefaultBaseURL = "https://openrouter.ai/api/v1"
	openrouterUserAgent      = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("openrouter", New)
}

// New returns the OpenRouter connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm OpenRouter connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "openrouter" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "openrouter",
		DisplayName:     "OpenRouter",
		IntegrationType: "api",
		Description:     "OpenAI-compatible gateway to many LLM vendors via the OpenRouter API; lists available models and serves as pm's default RLM LLM provider.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies configuration and, outside fixture mode, that the API key lists
// models successfully.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := openrouterBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(openrouterSecret(cfg)) == "" {
		return errors.New("openrouter connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, "models", nil, nil, nil); err != nil {
		return fmt.Errorf("check openrouter: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: openrouterStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader.
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
		stream = "models"
	}
	if stream != "models" {
		return fmt.Errorf("openrouter stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	resp, err := r.Do(ctx, http.MethodGet, "models", nil, nil)
	if err != nil {
		return fmt.Errorf("read openrouter models: %w", err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "data")
	if err != nil {
		return fmt.Errorf("decode openrouter models: %w", err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(openrouterModelRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// Write is unsupported: the OpenRouter connector is read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) readFixture(ctx context.Context, emit func(connectors.Record) error) error {
	for _, id := range []string{"anthropic/claude-fable-5", "openai/gpt-5.4"} {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{
			"id":             id,
			"name":           id,
			"context_length": int64(200000),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := openrouterBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := openrouterSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("openrouter connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: openrouterUserAgent,
	}, nil
}

func openrouterSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

func openrouterBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return openrouterDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("openrouter config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("openrouter config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("openrouter config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func openrouterStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "models",
			Description: "Models available through OpenRouter.",
			PrimaryKey:  []string{"id"},
			Fields: []connectors.Field{
				{Name: "id", Type: "string"},
				{Name: "name", Type: "string"},
				{Name: "context_length", Type: "integer"},
			},
		},
	}
}

func openrouterModelRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"name":           item["name"],
		"context_length": item["context_length"],
	}
}
