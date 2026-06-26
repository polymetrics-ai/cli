// Package smartsheets implements a read-only Smartsheet API connector for sheet
// metadata and rows. OAuth refresh is intentionally not implemented; callers pass
// an access token secret, matching Smartsheet's documented Bearer-token requests.
package smartsheets

import (
	"bytes"
	"context"
	"encoding/json"
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
	connectorName  = "smartsheets"
	defaultBaseURL = "https://api.smartsheet.com/2.0"
	userAgent      = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Smartsheets", IntegrationType: "api", Description: "Reads Smartsheet sheet metadata and rows. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	q := url.Values{"pageSize": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "/sheets", q, nil, nil); err != nil {
		return fmt.Errorf("check smartsheets: %w", err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: connectorName, Streams: []connectors.Stream{
		{Name: "sheets", Description: "Sheets accessible to the token.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "name", Type: "string"}, {Name: "permalink", Type: "string"}, {Name: "modifiedAt", Type: "timestamp"}}},
		{Name: "sheet_rows", Description: "Rows in the configured sheet or report.", PrimaryKey: []string{"row_id"}, CursorFields: []string{"modified_at"}, Fields: []connectors.Field{{Name: "sheet_id", Type: "integer"}, {Name: "sheet_name", Type: "string"}, {Name: "row_id", Type: "integer"}, {Name: "row_number", Type: "integer"}, {Name: "modified_at", Type: "timestamp"}, {Name: "cells", Type: "object"}}},
	}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "sheet_rows"
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	switch stream {
	case "sheets":
		return c.readSheets(ctx, r, req.Config, emit)
	case "sheet_rows":
		if strings.TrimSpace(req.Config.Config["spreadsheet_id"]) == "" {
			return errors.New("smartsheets connector requires config spreadsheet_id")
		}
		return c.readRows(ctx, r, req.Config, emit)
	default:
		return fmt.Errorf("smartsheets stream %q not found", stream)
	}
}

func (c Connector) readSheets(ctx context.Context, r *connsdk.Requester, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	for page := 1; ; page++ {
		q := pageQuery(cfg, page)
		resp, err := r.Do(ctx, http.MethodGet, "/sheets", q, nil)
		if err != nil {
			return fmt.Errorf("read smartsheets sheets: %w", err)
		}
		root, err := decode(resp.Body)
		if err != nil {
			return err
		}
		for _, item := range arrayAt(root, "data") {
			if err := ctx.Err(); err != nil {
				return err
			}
			if obj, ok := item.(map[string]any); ok {
				if err := emit(record(obj)); err != nil {
					return err
				}
			}
		}
		if page >= intAt(root, "totalPages") {
			return nil
		}
	}
}

func (c Connector) readRows(ctx context.Context, r *connsdk.Requester, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	path := "/sheets/" + url.PathEscape(strings.TrimSpace(cfg.Config["spreadsheet_id"]))
	for page := 1; ; page++ {
		q := pageQuery(cfg, page)
		q.Set("include", "rows")
		resp, err := r.Do(ctx, http.MethodGet, path, q, nil)
		if err != nil {
			return fmt.Errorf("read smartsheets rows: %w", err)
		}
		root, err := decode(resp.Body)
		if err != nil {
			return err
		}
		columns := columnsByID(root)
		for _, item := range arrayAt(root, "rows") {
			if err := ctx.Err(); err != nil {
				return err
			}
			row, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if err := emit(rowRecord(root, row, columns)); err != nil {
				return err
			}
		}
		if total := intAt(root, "totalPages"); total == 0 || page >= total {
			return nil
		}
	}
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := accessToken(cfg)
	if token == "" {
		return nil, errors.New("smartsheets connector requires secret access_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

func rowRecord(sheet, row map[string]any, columns map[string]string) connectors.Record {
	rec := connectors.Record{"sheet_id": sheet["id"], "sheet_name": sheet["name"], "row_id": row["id"], "row_number": row["rowNumber"], "modified_at": row["modifiedAt"], "cells": row["cells"]}
	for _, item := range arrayAt(row, "cells") {
		cell, ok := item.(map[string]any)
		if !ok {
			continue
		}
		key := columns[scalarKey(cell["columnId"])]
		if key == "" {
			key = "cell_" + scalarKey(cell["columnId"])
		}
		rec[key] = cell["value"]
	}
	return rec
}

func columnsByID(root map[string]any) map[string]string {
	out := map[string]string{}
	for _, item := range arrayAt(root, "columns") {
		col, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if title, ok := col["title"].(string); ok && title != "" {
			out[scalarKey(col["id"])] = title
		}
	}
	return out
}

func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	if stream == "sheets" {
		return emit(connectors.Record{"id": float64(900), "name": "Fixture Sheet", "fixture": true})
	}
	if stream != "sheet_rows" {
		return fmt.Errorf("smartsheets stream %q not found", stream)
	}
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"sheet_id": float64(900), "sheet_name": "Fixture Sheet", "row_id": float64(i), "row_number": float64(i), "Name": fmt.Sprintf("Fixture %d", i), "fixture": true}); err != nil {
			return err
		}
	}
	return nil
}

func decode(body []byte) (map[string]any, error) {
	dec := json.NewDecoder(bytes.NewReader(body))
	var out map[string]any
	if err := dec.Decode(&out); err != nil {
		return nil, fmt.Errorf("decode smartsheets response: %w", err)
	}
	return out, nil
}

func record(in map[string]any) connectors.Record {
	out := connectors.Record{}
	for k, v := range in {
		out[k] = v
	}
	return out
}

func arrayAt(m map[string]any, key string) []any {
	if v, ok := m[key].([]any); ok {
		return v
	}
	return nil
}

func intAt(m map[string]any, key string) int {
	switch v := m[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	}
	return 0
}

func scalarKey(v any) string {
	switch t := v.(type) {
	case float64:
		return strconv.FormatInt(int64(t), 10)
	case string:
		return t
	default:
		return fmt.Sprintf("%v", t)
	}
}

func pageQuery(cfg connectors.RuntimeConfig, page int) url.Values {
	q := url.Values{}
	q.Set("page", strconv.Itoa(page))
	n, err := strconv.Atoi(strings.TrimSpace(cfg.Config["page_size"]))
	if err != nil || n < 1 {
		n = 100
	}
	q.Set("pageSize", strconv.Itoa(n))
	return q
}

func accessToken(cfg connectors.RuntimeConfig) string {
	for _, key := range []string{"access_token", "credentials.access_token"} {
		if v := strings.TrimSpace(cfg.Secrets[key]); v != "" {
			return v
		}
		if v := strings.TrimSpace(cfg.Config[key]); v != "" {
			return v
		}
	}
	return ""
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	raw := strings.TrimSpace(cfg.Config["base_url"])
	if raw == "" {
		raw = defaultBaseURL
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return "", fmt.Errorf("smartsheets config base_url must be an absolute http(s) URL")
	}
	return strings.TrimRight(raw, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
