// Package uscensus implements the us-census defs bundle's StreamHook
// (conventions.md Tier-2 table), resolving the recorded ENGINE_GAP
// quarantine blocker (docs/migration/quarantine.json): legacy's sole
// `query` stream returns a raw top-level JSON array-of-arrays -- a header
// row of caller-driven field names followed by data rows -- with field
// names derived entirely at read-time from that header row, itself a
// function of the caller-supplied query_path/query_params `get=`
// qualifier. There is no fixed records/schema shape the declarative dialect
// can project, so this hook ports legacy internal/connectors/us-census/
// us_census.go's censusRows header-mapping logic almost verbatim via
// StreamHook.ReadStream.
//
// Auth needs no hook at all: legacy sends the configured api_key secret as
// a `key` query-string parameter (connsdk.APIKeyQuery), which
// streams.json's base.auth already declares declaratively
// (mode: api_key_query). Only one hook interface (StreamHook) is
// implemented here, well under the 2-interface Tier-2 cap.
package uscensus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func init() {
	engine.RegisterHooks("us-census", func() engine.Hooks { return Hooks{} })
}

// Hooks is the us-census hook set. It implements engine.StreamHook only.
type Hooks struct{}

func (Hooks) ConnectorName() string { return "us-census" }

// ReadStream implements engine.StreamHook. It always returns handled=true
// for the (sole) "query" stream -- mirroring legacy's Read (us_census.go:
// 62-105) verbatim: build the request from config.query_path/
// config.query_params, GET it via rt.Requester (auth/base URL/headers
// already resolved by the engine), decode the response as a raw JSON
// array-of-arrays, and map the header row (row 0) to lower-cased field
// names for every following data row.
func (Hooks) ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (bool, error) {
	if err := ctx.Err(); err != nil {
		return true, err
	}

	path := strings.TrimSpace(req.Config.Config["query_path"])
	if path == "" {
		return true, fmt.Errorf("us-census connector requires config query_path")
	}
	query, err := url.ParseQuery(strings.TrimSpace(req.Config.Config["query_params"]))
	if err != nil {
		return true, fmt.Errorf("us-census config query_params is invalid: %w", err)
	}

	resp, err := rt.Requester.Do(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		return true, fmt.Errorf("read us-census query: %w", err)
	}

	records, err := censusRows(resp.Body)
	if err != nil {
		return true, err
	}
	for _, rec := range records {
		if err := ctx.Err(); err != nil {
			return true, err
		}
		if err := emit(rec); err != nil {
			return true, err
		}
	}
	return true, nil
}

// censusRows decodes body as a raw JSON array-of-arrays (header row + data
// rows) and maps each data row into a record keyed by the lower-cased
// header row values -- ported verbatim from legacy us_census.go's
// censusRows, including its exact tolerances: fewer than 2 rows (no header,
// or header-only) yields zero records, and a column whose header cell is
// empty (or a row shorter than the header) is silently skipped for that
// column only.
func censusRows(body []byte) ([]connectors.Record, error) {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var rows [][]any
	if err := dec.Decode(&rows); err != nil {
		return nil, fmt.Errorf("decode us-census query: %w", err)
	}
	if len(rows) < 2 {
		return nil, nil
	}
	headers := make([]string, len(rows[0]))
	for i, header := range rows[0] {
		headers[i] = strings.ToLower(stringValue(header))
	}
	out := make([]connectors.Record, 0, len(rows)-1)
	for _, row := range rows[1:] {
		rec := connectors.Record{}
		for i, value := range row {
			if i >= len(headers) || headers[i] == "" {
				continue
			}
			rec[headers[i]] = stringValue(value)
		}
		out = append(out, rec)
	}
	return out, nil
}

func stringValue(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case json.Number:
		return t.String()
	default:
		return fmt.Sprint(t)
	}
}
