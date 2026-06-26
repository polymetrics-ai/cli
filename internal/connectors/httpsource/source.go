// Package httpsource provides a small template for read-only HTTP source
// connectors built on top of connsdk.
package httpsource

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

const defaultUserAgent = "polymetrics-go-cli"

// AuthType selects how Source authenticates outgoing HTTP requests.
type AuthType string

const (
	// AuthNone sends no authentication.
	AuthNone AuthType = ""
	// AuthBearer sends Authorization: Bearer <secret>.
	AuthBearer AuthType = "bearer"
	// AuthAPIKeyHeader sends an API key in a configured request header.
	AuthAPIKeyHeader AuthType = "api_key_header"
	// AuthAPIKeyQuery sends an API key as a configured query parameter.
	AuthAPIKeyQuery AuthType = "api_key_query"
)

// AuthSpec describes how to turn RuntimeConfig.Secrets into a connsdk auth
// strategy. SecretName defaults to "api_key" when omitted.
type AuthSpec struct {
	Type       AuthType
	SecretName string
	Header     string
	QueryParam string
	Prefix     string
}

// Spec describes a read-only HTTP source connector.
type Spec struct {
	Name            string
	DisplayName     string
	IntegrationType string
	Description     string

	DefaultBaseURL string
	DefaultStream  string
	UserAgent      string
	Auth           AuthSpec

	DefaultPageSize int
	MaxPageSize     int
	DefaultMaxPages int

	Streams []StreamSpec
}

// StreamSpec describes one HTTP endpoint exposed as a connector stream.
type StreamSpec struct {
	Name         string
	Description  string
	Fields       []connectors.Field
	PrimaryKey   []string
	CursorFields []string

	Method      string
	Path        string
	RecordsPath string

	// Paginator builds the paginator after page_size parsing. When nil, Source
	// reads exactly one page through connsdk.Harvest.
	Paginator func(pageSize int) connsdk.Paginator
	// Query returns base query params shared by every page.
	Query func(cfg connectors.RuntimeConfig, pageSize int) url.Values

	// FixtureRecords, when set, replaces the generated deterministic fixture rows.
	FixtureRecords []connectors.Record
	// Map, when set, converts each extracted JSON object before emit.
	Map func(connectors.Record) (connectors.Record, error)
}

// Source implements connectors.Connector and connectors.StatefulReader for a
// Spec. It is intended to be embedded or returned by per-system connector
// packages, not registered by this template package itself.
type Source struct {
	Spec   Spec
	Client *http.Client
}

var _ connectors.Connector = Source{}
var _ connectors.StatefulReader = Source{}

// New returns a Source for spec.
func New(spec Spec) Source { return Source{Spec: spec} }

func (s Source) Name() string { return s.Spec.Name }

func (s Source) Metadata() connectors.Metadata {
	name := s.Name()
	display := strings.TrimSpace(s.Spec.DisplayName)
	if display == "" {
		display = name
	}
	integrationType := strings.TrimSpace(s.Spec.IntegrationType)
	if integrationType == "" {
		integrationType = "api"
	}
	return connectors.Metadata{
		Name:            name,
		DisplayName:     display,
		IntegrationType: integrationType,
		Description:     s.Spec.Description,
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

func (s Source) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := s.validateSpec(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := s.baseURL(cfg); err != nil {
		return err
	}
	if _, err := s.authenticator(cfg); err != nil {
		return err
	}
	if _, err := s.pageSize(cfg); err != nil {
		return err
	}
	if _, err := s.maxPages(cfg); err != nil {
		return err
	}
	return nil
}

func (s Source) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	streams := make([]connectors.Stream, 0, len(s.Spec.Streams))
	for _, stream := range s.Spec.Streams {
		streams = append(streams, connectors.Stream{
			Name:         stream.Name,
			Description:  stream.Description,
			Fields:       append([]connectors.Field(nil), stream.Fields...),
			PrimaryKey:   append([]string(nil), stream.PrimaryKey...),
			CursorFields: append([]string(nil), stream.CursorFields...),
		})
	}
	return connectors.Catalog{Connector: s.Name(), Streams: streams}, nil
}

func (s Source) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	spec, err := s.streamSpec(stream)
	if err != nil {
		return nil, err
	}
	return connsdk.WithCursor(map[string]string{"stream": spec.Name}, ""), nil
}

func (s Source) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream, err := s.streamSpec(req.Stream)
	if err != nil {
		return err
	}
	if fixtureMode(req.Config) {
		return s.readFixture(ctx, stream, emit)
	}
	r, err := s.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := s.pageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := s.maxPages(req.Config)
	if err != nil {
		return err
	}
	query := url.Values{}
	if stream.Query != nil {
		query = stream.Query(req.Config, pageSize)
		if query == nil {
			query = url.Values{}
		}
	}
	paginator := connsdk.Paginator(singlePagePaginator{})
	if stream.Paginator != nil {
		if p := stream.Paginator(pageSize); p != nil {
			paginator = p
		}
	}
	method := strings.TrimSpace(stream.Method)
	if method == "" {
		method = http.MethodGet
	}
	if err := connsdk.Harvest(ctx, r, method, stream.Path, query, paginator, stream.RecordsPath, maxPages, func(rec connsdk.Record) error {
		mapped, err := stream.mapRecord(connectors.Record(rec))
		if err != nil {
			return err
		}
		return emit(mapped)
	}); err != nil {
		return fmt.Errorf("read %s %s: %w", s.errorName(), stream.Name, err)
	}
	return nil
}

func (stream StreamSpec) mapRecord(record connectors.Record) (connectors.Record, error) {
	if stream.Map == nil {
		return record, nil
	}
	return stream.Map(record)
}

func (s Source) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (s Source) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := s.baseURL(cfg)
	if err != nil {
		return nil, err
	}
	auth, err := s.authenticator(cfg)
	if err != nil {
		return nil, err
	}
	userAgent := strings.TrimSpace(s.Spec.UserAgent)
	if userAgent == "" {
		userAgent = defaultUserAgent
	}
	return &connsdk.Requester{Client: s.Client, BaseURL: base, Auth: auth, UserAgent: userAgent}, nil
}

func (s Source) baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = strings.TrimSpace(s.Spec.DefaultBaseURL)
	}
	if base == "" {
		return "", fmt.Errorf("%s connector requires config base_url", s.errorName())
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("%s config base_url is invalid: %w", s.errorName(), err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("%s config base_url must use http or https, got %q", s.errorName(), parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("%s config base_url must include a host", s.errorName())
	}
	return strings.TrimRight(base, "/"), nil
}

func (s Source) authenticator(cfg connectors.RuntimeConfig) (connsdk.Authenticator, error) {
	auth := s.Spec.Auth
	switch auth.Type {
	case AuthNone:
		return nil, nil
	case AuthBearer:
		value, err := s.secret(cfg, auth.secretName())
		if err != nil {
			return nil, err
		}
		return connsdk.Bearer(value), nil
	case AuthAPIKeyHeader:
		header := strings.TrimSpace(auth.Header)
		if header == "" {
			return nil, fmt.Errorf("%s auth api_key_header requires header", s.errorName())
		}
		value, err := s.secret(cfg, auth.secretName())
		if err != nil {
			return nil, err
		}
		return connsdk.APIKeyHeader(header, value, auth.Prefix), nil
	case AuthAPIKeyQuery:
		param := strings.TrimSpace(auth.QueryParam)
		if param == "" {
			return nil, fmt.Errorf("%s auth api_key_query requires query param", s.errorName())
		}
		value, err := s.secret(cfg, auth.secretName())
		if err != nil {
			return nil, err
		}
		return connsdk.APIKeyQuery(param, value), nil
	default:
		return nil, fmt.Errorf("%s auth type %q is not supported", s.errorName(), auth.Type)
	}
}

func (a AuthSpec) secretName() string {
	name := strings.TrimSpace(a.SecretName)
	if name == "" {
		return "api_key"
	}
	return name
}

func (s Source) secret(cfg connectors.RuntimeConfig, name string) (string, error) {
	value := ""
	if cfg.Secrets != nil {
		value = cfg.Secrets[name]
	}
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%s connector requires secret %s", s.errorName(), name)
	}
	return value, nil
}

func (s Source) pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return s.defaultPageSize(), nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s config page_size must be an integer: %w", s.errorName(), err)
	}
	if value < 1 {
		return 0, fmt.Errorf("%s config page_size must be a positive integer", s.errorName())
	}
	if max := s.Spec.MaxPageSize; max > 0 && value > max {
		return 0, fmt.Errorf("%s config page_size must be between 1 and %d", s.errorName(), max)
	}
	return value, nil
}

func (s Source) defaultPageSize() int {
	if s.Spec.DefaultPageSize > 0 {
		return s.Spec.DefaultPageSize
	}
	return 100
}

func (s Source) maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" {
		if s.Spec.DefaultMaxPages > 0 {
			return s.Spec.DefaultMaxPages, nil
		}
		return 0, nil
	}
	if raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s config max_pages must be an integer, all, or unlimited: %w", s.errorName(), err)
	}
	if value < 0 {
		return 0, fmt.Errorf("%s config max_pages must be 0 for unlimited or a positive integer", s.errorName())
	}
	return value, nil
}

func (s Source) streamSpec(name string) (StreamSpec, error) {
	if err := s.validateSpec(); err != nil {
		return StreamSpec{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = strings.TrimSpace(s.Spec.DefaultStream)
	}
	if name == "" {
		name = s.Spec.Streams[0].Name
	}
	for _, stream := range s.Spec.Streams {
		if stream.Name == name {
			return stream, nil
		}
	}
	return StreamSpec{}, fmt.Errorf("%s stream %q not found", s.errorName(), name)
}

func (s Source) validateSpec() error {
	if strings.TrimSpace(s.Spec.Name) == "" {
		return errors.New("httpsource spec name is required")
	}
	if len(s.Spec.Streams) == 0 {
		return fmt.Errorf("%s spec requires at least one stream", s.errorName())
	}
	seen := map[string]bool{}
	for _, stream := range s.Spec.Streams {
		name := strings.TrimSpace(stream.Name)
		if name == "" {
			return fmt.Errorf("%s stream name is required", s.errorName())
		}
		if seen[name] {
			return fmt.Errorf("%s stream %q is duplicated", s.errorName(), name)
		}
		seen[name] = true
		if strings.TrimSpace(stream.Path) == "" {
			return fmt.Errorf("%s stream %q path is required", s.errorName(), name)
		}
	}
	return nil
}

func (s Source) readFixture(ctx context.Context, stream StreamSpec, emit func(connectors.Record) error) error {
	if len(stream.FixtureRecords) > 0 {
		for _, record := range stream.FixtureRecords {
			if err := ctx.Err(); err != nil {
				return err
			}
			out := cloneRecord(record)
			if _, ok := out["fixture"]; !ok {
				out["fixture"] = true
			}
			if _, ok := out["connector"]; !ok {
				out["connector"] = s.Name()
			}
			if err := emit(out); err != nil {
				return err
			}
		}
		return nil
	}

	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		id := fmt.Sprintf("%s_fixture_%d", stream.Name, i)
		record := connectors.Record{"id": id, "connector": s.Name(), "fixture": true}
		if len(stream.PrimaryKey) > 0 && strings.TrimSpace(stream.PrimaryKey[0]) != "" {
			record[stream.PrimaryKey[0]] = id
		}
		for _, field := range stream.Fields {
			if _, exists := record[field.Name]; exists {
				continue
			}
			record[field.Name] = fixtureValue(stream.Name, field, i)
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

func fixtureValue(stream string, field connectors.Field, index int) any {
	switch strings.ToLower(strings.TrimSpace(field.Type)) {
	case "integer", "int":
		return int64(index)
	case "number", "float":
		return float64(index)
	case "boolean", "bool":
		return index%2 == 1
	case "timestamp", "datetime":
		return fmt.Sprintf("2026-01-%02dT00:00:00Z", index)
	case "date":
		return fmt.Sprintf("2026-01-%02d", index)
	case "object":
		return map[string]any{"fixture": true}
	case "array":
		return []any{}
	default:
		return fmt.Sprintf("%s_%s_fixture_%d", stream, field.Name, index)
	}
}

func cloneRecord(in connectors.Record) connectors.Record {
	out := make(connectors.Record, len(in)+2)
	for k, v := range in {
		out[k] = v
	}
	return out
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func (s Source) errorName() string {
	name := strings.TrimSpace(s.Name())
	if name == "" {
		return "httpsource"
	}
	return name
}

type singlePagePaginator struct{}

func (singlePagePaginator) Start() *connsdk.NextPage { return &connsdk.NextPage{} }

func (singlePagePaginator) Next(resp *connsdk.Response, recordCount int) *connsdk.NextPage {
	return nil
}
