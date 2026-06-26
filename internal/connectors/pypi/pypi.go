// Package pypi implements a read-only native connector for PyPI's JSON API.
package pypi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	defaultBaseURL = "https://pypi.org"
	userAgent      = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("pypi", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

func (Connector) Name() string { return "pypi" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "pypi",
		DisplayName:     "PyPI",
		IntegrationType: "api",
		Description:     "Reads PyPI project metadata and release files through the PyPI JSON API. Read-only and credential-free.",
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
	project, err := projectName(cfg)
	if err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, projectPath(project, cfg.Config["version"]), nil, nil, nil); err != nil {
		return fmt.Errorf("check pypi: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "project"
	}
	if stream != "project" && stream != "releases" {
		return fmt.Errorf("pypi stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
	}
	project, err := projectName(req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	resp, err := r.Do(ctx, http.MethodGet, projectPath(project, req.Config.Config["version"]), nil, nil)
	if err != nil {
		return fmt.Errorf("read pypi %s: %w", project, err)
	}
	if stream == "project" {
		return emitProject(ctx, resp.Body, emit)
	}
	return emitReleases(ctx, project, strings.TrimSpace(req.Config.Config["version"]), resp.Body, emit)
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, UserAgent: userAgent}, nil
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "project", Description: "PyPI project metadata.", Fields: []connectors.Field{{Name: "name", Type: "string"}, {Name: "version", Type: "string"}, {Name: "summary", Type: "string"}}, PrimaryKey: []string{"name"}},
		{Name: "releases", Description: "PyPI release file metadata.", Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "project_name", Type: "string"}, {Name: "version", Type: "string"}, {Name: "filename", Type: "string"}, {Name: "upload_time", Type: "timestamp"}}, PrimaryKey: []string{"id"}, CursorFields: []string{"upload_time"}},
	}
}

func emitProject(ctx context.Context, body []byte, emit func(connectors.Record) error) error {
	records, err := connsdk.RecordsAt(body, "info")
	if err != nil {
		return fmt.Errorf("decode pypi project: %w", err)
	}
	for _, rec := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record(rec)); err != nil {
			return err
		}
	}
	return nil
}

func emitReleases(ctx context.Context, project, version string, body []byte, emit func(connectors.Record) error) error {
	var root struct {
		Info     map[string]any              `json:"info"`
		Releases map[string][]map[string]any `json:"releases"`
		URLs     []map[string]any            `json:"urls"`
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	if err := dec.Decode(&root); err != nil {
		return fmt.Errorf("decode pypi releases: %w", err)
	}
	if version != "" {
		if len(root.URLs) == 0 {
			return nil
		}
		for _, file := range root.URLs {
			if err := emit(releaseRecord(project, version, file)); err != nil {
				return err
			}
		}
		return nil
	}
	versions := make([]string, 0, len(root.Releases))
	for v := range root.Releases {
		versions = append(versions, v)
	}
	sort.Strings(versions)
	for _, v := range versions {
		for _, file := range root.Releases[v] {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(releaseRecord(project, v, file)); err != nil {
				return err
			}
		}
	}
	return nil
}

func releaseRecord(project, version string, file map[string]any) connectors.Record {
	filename, _ := file["filename"].(string)
	return connectors.Record{
		"id":             project + ":" + version + ":" + filename,
		"project_name":   project,
		"version":        version,
		"filename":       file["filename"],
		"url":            file["url"],
		"packagetype":    file["packagetype"],
		"python_version": file["python_version"],
		"size":           file["size"],
		"upload_time":    first(file, "upload_time_iso_8601", "upload_time"),
	}
}

func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	if stream == "project" {
		return emit(connectors.Record{"name": "sampleproject", "version": "1.0.0", "summary": "Fixture PyPI project"})
	}
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": fmt.Sprintf("sampleproject:1.0.%d:file.tar.gz", i), "project_name": "sampleproject", "version": fmt.Sprintf("1.0.%d", i), "filename": "file.tar.gz", "upload_time": fmt.Sprintf("2026-01-0%dT00:00:00Z", i)}); err != nil {
			return err
		}
	}
	return nil
}

func first(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if v := item[key]; v != nil {
			return v
		}
	}
	return nil
}

func projectName(cfg connectors.RuntimeConfig) (string, error) {
	project := strings.TrimSpace(cfg.Config["project_name"])
	if project == "" {
		return "", errors.New("pypi connector requires config project_name")
	}
	if strings.ContainsAny(project, "/?#") || strings.Contains(project, "..") {
		return "", fmt.Errorf("pypi config project_name %q is invalid", project)
	}
	return project, nil
}

func projectPath(project, version string) string {
	project = url.PathEscape(strings.TrimSpace(project))
	version = strings.TrimSpace(version)
	if version != "" {
		return "pypi/" + project + "/" + url.PathEscape(version) + "/json"
	}
	return "pypi/" + project + "/json"
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("pypi config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("pypi config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("pypi config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
