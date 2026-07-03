// Package smartsheets is the Tier-2 StreamHook for the smartsheets defs
// bundle: Smartsheet's row-listing endpoint returns each row's cell values
// as an array keyed by a numeric columnId, with the column TITLE only
// resolvable via a sibling columns[] array carried in the SAME page body
// (not inside the row itself). Legacy (smartsheets.go's rowRecord/
// columnsByID) flattens each row's cells into dynamically-named top-level
// fields (one per sheet column, e.g. "Name"/"Status") — the engine's
// declarative records/computed_fields dialect has no primitive that can
// look up a sibling array by a per-record foreign key and use its values as
// OUTPUT FIELD NAMES (computed_fields only ever produces a fixed,
// statically-declared set of named fields; a stream's real column set is
// only known at read time, per sheet). This is a genuine ENGINE_GAP for the
// column-flatten step specifically — see docs.md's Known limits — expressed
// here via the sanctioned StreamHook seam (conventions.md Tier-2 table:
// "whole-stream override"), reusing rt.Requester (the engine's already-built
// HTTP client/auth/base-URL plumbing) exactly as the declarative path itself
// would.
package smartsheets

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

const defaultPageSize = 100

func init() {
	engine.RegisterHooks("smartsheets", func() engine.Hooks { return Hooks{} })
}

// Hooks is smartsheets' Tier-2 hook set: a single StreamHook handling only
// the sheet_rows stream (handled=false for every other stream falls back to
// the declarative path — the sheets stream is fully declarative and never
// reaches this hook in practice, since engine.HooksFor gates dispatch, but
// returning false here keeps the contract honest: hooks are additive, not a
// blanket override, per conventions.md §1).
type Hooks struct{}

func (Hooks) ConnectorName() string { return "smartsheets" }

// ReadStream implements engine.StreamHook. Only "sheet_rows" is handled;
// every other stream name returns handled=false so the engine's declarative
// read path (streams.json's "sheets" entry) runs unmodified.
func (h Hooks) ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (bool, error) {
	if stream.Name != "sheet_rows" {
		return false, nil
	}
	spreadsheetID := strings.TrimSpace(req.Config.Config["spreadsheet_id"])
	if spreadsheetID == "" {
		return true, fmt.Errorf("smartsheets connector requires config spreadsheet_id")
	}
	return true, h.readRows(ctx, rt.Requester, spreadsheetID, pageSize(req.Config), emit)
}

// readRows ports smartsheets.go's readRows verbatim: paginate /sheets/{id}
// with include=rows, decode the page body once, build a columnId->title map
// from the page's sibling columns[] array, and flatten each row's cells
// into dynamically-named fields keyed by column title (falling back to
// "cell_<columnId>" when a column has no title) alongside the fixed
// sheet_id/sheet_name/row_id/row_number/modified_at/cells fields the
// declarative schema declares.
func (h Hooks) readRows(ctx context.Context, r *connsdk.Requester, spreadsheetID string, size int, emit func(connectors.Record) error) error {
	path := "/sheets/" + url.PathEscape(spreadsheetID)
	for page := 1; ; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		q := url.Values{}
		q.Set("page", strconv.Itoa(page))
		q.Set("pageSize", strconv.Itoa(size))
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

// Check implements engine.CheckHook is NOT implemented — the declarative
// base.check (GET /sheets?pageSize=1) already matches legacy's Check exactly
// (smartsheets.go:39-55), so no hook is needed for it.

func pageSize(cfg connectors.RuntimeConfig) int {
	n, err := strconv.Atoi(strings.TrimSpace(cfg.Config["page_size"]))
	if err != nil || n < 1 {
		return defaultPageSize
	}
	return n
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

func decode(body []byte) (map[string]any, error) {
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode smartsheets response: %w", err)
	}
	return out, nil
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
