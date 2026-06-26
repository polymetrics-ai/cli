// Package boxdataextract implements a read-only Box Data Extract connector. It
// uses Box OAuth2 client credentials, bounded offset pagination, deterministic
// fixtures, and no write actions.
package boxdataextract

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
	defaultBaseURL  = "https://api.box.com/2.0"
	defaultTokenURL = "https://api.box.com/oauth2/token"
	defaultPageSize = 100
	defaultMaxPages = 100
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("box-data-extract", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "box-data-extract" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "box-data-extract",
		DisplayName:     "Box Data Extract",
		IntegrationType: "api",
		Description:     "Reads Box folder files and extracted file text through the Box REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if err := requireCredentials(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, folderItemsResource(cfg), url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check box-data-extract: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: []connectors.Stream{
		{Name: "files", Description: "Box folder items.", Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "type", Type: "string"}, {Name: "name", Type: "string"}}, PrimaryKey: []string{"id"}},
		{Name: "file_text", Description: "Extracted Box file text metadata.", Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "file_id", Type: "string"}, {Name: "text", Type: "string"}}, PrimaryKey: []string{"id"}},
	}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "files"
	}
	if stream != "files" && stream != "file_text" {
		return fmt.Errorf("box-data-extract stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, emit)
	}
	if stream == "file_text" {
		return errors.New("box-data-extract file_text live read requires fixture mode or a safe extraction service")
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := intConfig(req.Config, "page_size", defaultPageSize)
	if err != nil {
		return err
	}
	maxPages, err := intConfig(req.Config, "max_pages", defaultMaxPages)
	if err != nil {
		return err
	}
	return readOffset(ctx, r, folderItemsResource(req.Config), pageSize, maxPages, emit)
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func readOffset(ctx context.Context, r *connsdk.Requester, resource string, pageSize, maxPages int, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; page < maxPages; page++ {
		query := url.Values{"limit": []string{strconv.Itoa(pageSize)}, "offset": []string{strconv.Itoa(offset)}}
		resp, err := r.Do(ctx, http.MethodGet, resource, query, nil)
		if err != nil {
			return fmt.Errorf("read box-data-extract files: %w", err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "entries")
		if err != nil {
			return fmt.Errorf("decode box-data-extract files: %w", err)
		}
		for _, rec := range records {
			if err := emit(connectors.Record(rec)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
		offset += pageSize
		if total, ok := intAt(resp.Body, "total_count"); ok && offset >= total {
			return nil
		}
	}
	return nil
}

func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "fixture": true}
		if stream == "files" {
			rec["type"] = "file"
			rec["name"] = fmt.Sprintf("fixture-%d.txt", i)
		} else {
			rec["file_id"] = fmt.Sprintf("file_fixture_%d", i)
			rec["text"] = fmt.Sprintf("fixture extracted text %d", i)
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg, "base_url", defaultBaseURL)
	if err != nil {
		return nil, err
	}
	if err := requireCredentials(cfg); err != nil {
		return nil, err
	}
	extra := url.Values{}
	subjectType := strings.TrimSpace(cfg.Config["box_subject_type"])
	if subjectType == "" {
		subjectType = "enterprise"
	}
	extra.Set("box_subject_type", subjectType)
	if id := strings.TrimSpace(cfg.Config["box_subject_id"]); id != "" {
		extra.Set("box_subject_id", id)
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: &connsdk.OAuth2ClientCredentials{TokenURL: tokenURL(cfg), ClientID: secretOrConfig(cfg, "client_id"), ClientSecret: secretOrConfig(cfg, "client_secret"), ExtraParams: extra, Client: c.Client}, UserAgent: userAgent}, nil
}

func folderItemsResource(cfg connectors.RuntimeConfig) string {
	folderID := strings.TrimSpace(cfg.Config["box_folder_id"])
	if folderID == "" {
		folderID = "0"
	}
	return "folders/" + url.PathEscape(folderID) + "/items"
}

func requireCredentials(cfg connectors.RuntimeConfig) error {
	if strings.TrimSpace(secretOrConfig(cfg, "client_id")) == "" {
		return errors.New("box-data-extract connector requires secret client_id")
	}
	if strings.TrimSpace(secretOrConfig(cfg, "client_secret")) == "" {
		return errors.New("box-data-extract connector requires secret client_secret")
	}
	return nil
}

func tokenURL(cfg connectors.RuntimeConfig) string {
	if raw := strings.TrimSpace(cfg.Config["token_url"]); raw != "" {
		return raw
	}
	return defaultTokenURL
}

func baseURL(cfg connectors.RuntimeConfig, key, fallback string) (string, error) {
	base := strings.TrimSpace(cfg.Config[key])
	if base == "" {
		base = fallback
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("box-data-extract config %s is invalid: %w", key, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("box-data-extract config %s must use http or https", key)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("box-data-extract config %s must include a host", key)
	}
	return strings.TrimRight(base, "/"), nil
}

func intConfig(cfg connectors.RuntimeConfig, key string, fallback int) (int, error) {
	raw := strings.TrimSpace(cfg.Config[key])
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, fmt.Errorf("box-data-extract config %s must be a positive integer", key)
	}
	return value, nil
}

func intAt(body []byte, path string) (int, bool) {
	raw, err := connsdk.StringAt(body, path)
	if err != nil || strings.TrimSpace(raw) == "" {
		return 0, false
	}
	value, err := strconv.Atoi(raw)
	return value, err == nil
}

func secretOrConfig(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets != nil && cfg.Secrets[key] != "" {
		return cfg.Secrets[key]
	}
	return cfg.Config[key]
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
