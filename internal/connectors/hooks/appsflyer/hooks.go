// Package appsflyer implements the appsflyer bundle's StreamHook
// (conventions.md §1's Tier-2 table: "response decompression/CSV parsing"):
// AppsFlyer's Pull API raw-data export endpoints
// (GET /api/raw-data/export/app/{app_id}/<report>/v5) return a text/csv
// body, not JSON — engine/read.go's declarative record-extraction path only
// ever decodes JSON (connsdk.RecordsAt), so it cannot express this wire
// shape at all. This hook ports internal/connectors/appsflyer/
// appsflyer.go's emitCSV/snake/reportPath/reportQuery logic verbatim,
// reusing rt.Requester (the engine's already-built *connsdk.Requester: base
// URL/auth/headers already resolved from streams.json's base block).
//
// Only one hook interface is implemented (StreamHook), well under the
// Tier-2 2-interface cap; this file is well under the ~300-line soft
// target.
package appsflyer

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"net/url"
	"strings"
	"unicode"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func init() {
	engine.RegisterHooks("appsflyer", func() engine.Hooks { return Hooks{} })
}

// Hooks is appsflyer's Tier-2 hook set: StreamHook only.
type Hooks struct{}

func (Hooks) ConnectorName() string { return "appsflyer" }

// streamReportPaths mirrors legacy appsflyer.go's streamPaths map: the
// report-name path segment for each stream (identical to the stream name
// here, kept as an explicit table for parity with legacy's shape and to
// guard against a future stream/report-name divergence).
var streamReportPaths = map[string]string{
	"installs_report":      "installs_report",
	"in_app_events_report": "in_app_events_report",
}

// ReadStream implements engine.StreamHook: builds the report path/query
// exactly like legacy's reportPath/reportQuery, issues one GET via
// rt.Requester, and decodes the text/csv response into records via
// emitCSV. handled=false is returned only for an unknown stream name (never
// reached in practice since engine.findStream already validates the name
// against streams.json before ReadStream is called), letting the
// declarative path produce its own "stream not found" error.
func (h Hooks) ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (bool, error) {
	report, ok := streamReportPaths[stream.Name]
	if !ok {
		return false, nil
	}

	path := reportPath(req.Config, report)
	query := reportQuery(req.Config)

	resp, err := rt.Requester.Do(ctx, "GET", path, query, nil)
	if err != nil {
		return true, fmt.Errorf("read appsflyer %s: %w", stream.Name, err)
	}
	if err := emitCSV(ctx, resp.Body, emit); err != nil {
		return true, err
	}
	return true, nil
}

// reportPath mirrors legacy's reportPath (appsflyer.go:110-112) field-for-
// field, including url.PathEscape on app_id (identical safety to legacy's
// own call).
func reportPath(cfg connectors.RuntimeConfig, report string) string {
	appID := strings.TrimSpace(cfg.Config["app_id"])
	return "/api/raw-data/export/app/" + url.PathEscape(appID) + "/" + report + "/v5"
}

// reportQuery mirrors legacy's reportQuery (appsflyer.go:114-128): from/to
// are the date-only portion (truncated at the first space, matching
// legacy's firstDate) of start_date/end_date, with end_date defaulting to
// start_date when unset (legacy's first(end_date, start_date)); timezone is
// sent verbatim when configured.
func reportQuery(cfg connectors.RuntimeConfig) url.Values {
	q := url.Values{}
	from := strings.TrimSpace(cfg.Config["start_date"])
	to := strings.TrimSpace(firstNonEmpty(cfg.Config["end_date"], cfg.Config["start_date"]))
	if from != "" {
		q.Set("from", firstDate(from))
	}
	if to != "" {
		q.Set("to", firstDate(to))
	}
	if tz := strings.TrimSpace(cfg.Config["timezone"]); tz != "" {
		q.Set("timezone", tz)
	}
	return q
}

// emitCSV mirrors legacy's emitCSV (appsflyer.go:130-159) field-for-field:
// decode every row via encoding/csv (FieldsPerRecord: -1, tolerating
// ragged rows exactly like legacy), snake-case the header row, and emit one
// connectors.Record per data row with every column present (verbatim
// passthrough of whatever columns the live report returns).
func emitCSV(ctx context.Context, body []byte, emit func(connectors.Record) error) error {
	r := csv.NewReader(bytes.NewReader(body))
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		return fmt.Errorf("decode appsflyer csv: %w", err)
	}
	if len(rows) == 0 {
		return nil
	}
	headers := make([]string, len(rows[0]))
	for i, h := range rows[0] {
		headers[i] = snake(h)
	}
	for _, row := range rows[1:] {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{}
		for i, v := range row {
			if i < len(headers) && headers[i] != "" {
				rec[headers[i]] = v
			}
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

// snake mirrors legacy's snake (appsflyer.go:173-188) verbatim, including
// the apps_flyer -> appsflyer correction so "AppsFlyer ID" becomes
// appsflyer_id, not apps_flyer_id.
func snake(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	var b strings.Builder
	lastUnderscore := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastUnderscore = false
		} else if !lastUnderscore {
			b.WriteByte('_')
			lastUnderscore = true
		}
	}
	out := strings.Trim(b.String(), "_")
	return strings.ReplaceAll(out, "apps_flyer", "appsflyer")
}

// firstDate mirrors legacy's firstDate (appsflyer.go:190-195): truncates at
// the first space, so a "YYYY-MM-DD HH:MM:SS"-shaped config value is
// reduced to its date-only portion; a bare date passes through unchanged.
func firstDate(raw string) string {
	if i := strings.IndexByte(raw, ' '); i > 0 {
		return raw[:i]
	}
	return raw
}

// firstNonEmpty mirrors legacy's first (appsflyer.go:197-204): returns the
// first non-blank (after TrimSpace) value.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
