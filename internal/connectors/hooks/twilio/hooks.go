// Package twilio implements the Tier-2 StreamHook for the twilio bundle
// (docs/migration/quarantine.json's OTHER/no-reason entry; investigation
// found the real blocker is pagination, not auth — see defs/twilio/docs.md's
// Overview). Twilio's real wire pagination convention (harvest,
// internal/connectors/twilio/twilio.go:161-208) reads a "next_page_uri"
// field from the response body whose value is a HOST-RELATIVE URL (e.g.
// "/2010-04-01/Accounts/AC_test/Messages.json?Page=1&PageSize=2"), never an
// absolute one. The engine's only declarative pagination type that reads a
// next-page URL from a response body, "next_url" (engine/paginate.go's
// nextURL/checkOrigin), enforces a same-origin SSRF guard that fail-closed
// REJECTS any next-page URL with an empty Host — correct guard behavior for
// a genuine cross-host redirect, but it also rejects Twilio's own
// legitimate host-relative convention, with no dialect escape hatch. This
// is the identical structural gap docs/migration/quarantine.json's
// rootly/safetyculture entries hit (both still quarantined ENGINE_GAP).
//
// This hook ports legacy's harvest/absoluteURL verbatim: the host-relative
// next_page_uri is resolved against the SAME requester's own resolved base
// origin (rt.Requester.BaseURL) — never a caller-controlled host — so the
// SSRF surface next_url's guard protects against does not reopen here; the
// hook only ever follows a path Twilio itself returned, scoped to the
// connection's own configured host.
//
// Auth is NOT hooked: Twilio's HTTP Basic account_sid/auth_token pair is
// fully declarative (mode: basic, streams.json's base.auth), so this hook
// set implements StreamHook only, well under the 2-interface Tier-2 cap.
package twilio

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

const (
	defaultPageSize = 50
	maxPageSize     = 1000
)

func init() {
	engine.RegisterHooks("twilio", func() engine.Hooks { return New() })
}

// Hooks is the twilio hook set: StreamHook only. It has no state of its
// own; every method is a pure function of its arguments.
type Hooks struct{}

// New returns a fresh twilio Hooks value (StreamHook implementation).
func New() engine.Hooks { return Hooks{} }

func (Hooks) ConnectorName() string { return "twilio" }

// streamEndpoint maps a stream name to the JSON object key holding its
// records array in the list response and the record mapper that flattens a
// raw item into a connectors.Record — byte-for-byte port of legacy's
// twilioStreamEndpoints routing table (twilio/streams.go:20-26) minus the
// resource path (already fully expressed by streams.json's stream
// path/RecordsSpec).
type streamEndpoint struct {
	recordsKey string
	mapRecord  func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{
	"messages":      {recordsKey: "messages", mapRecord: messageRecord},
	"calls":         {recordsKey: "calls", mapRecord: callRecord},
	"recordings":    {recordsKey: "recordings", mapRecord: recordingRecord},
	"conferences":   {recordsKey: "conferences", mapRecord: conferenceRecord},
	"usage_records": {recordsKey: "usage_records", mapRecord: usageRecordRecord},
}

// ReadStream drives Twilio's next_page_uri pagination for every stream this
// bundle declares (all 5 recognized stream names always return
// handled=true; an unrecognized stream name returns handled=false, letting
// the declarative fallback surface an honest "stream not found" error
// rather than this hook panicking).
func (h Hooks) ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (bool, error) {
	if err := ctx.Err(); err != nil {
		return true, err
	}

	name := stream.Name
	if name == "" {
		name = "messages"
	}
	endpoint, ok := streamEndpoints[name]
	if !ok {
		return false, nil
	}

	pageSize, err := pageSizeFor(req.Config)
	if err != nil {
		return true, err
	}
	maxPages, err := maxPagesFor(req.Config)
	if err != nil {
		return true, err
	}

	// stream.Path carries the templated "{{ secrets.account_sid }}"
	// account-scoping segment (streams.json); the declarative read path
	// would normally resolve this via engine.InterpolatePath before every
	// request (F1/SECURITY-REVIEW.md m3) — this hook fully replaces that
	// dispatch, so it must resolve the path itself, once, up front (the
	// resolved path never changes across pages; only next_page_uri does).
	path, err := engine.InterpolatePath(stream.Path, engine.Vars{Config: req.Config.Config, Secrets: req.Config.Secrets})
	if err != nil {
		return true, fmt.Errorf("twilio: resolve stream path: %w", err)
	}

	return true, h.harvest(ctx, rt.Requester, path, endpoint, pageSize, maxPages, emit)
}

// harvest ports legacy's harvest (twilio.go:161-208) verbatim: the first
// request carries a PageSize query param against the stream's own
// account-scoped path; each subsequent page follows next_page_uri
// (resolved against the requester's own base origin via absoluteURL) until
// next_page_uri is null/empty, or maxPages is reached (maxPages<=0 means
// unbounded, matching legacy's harvest loop condition exactly).
func (h Hooks) harvest(ctx context.Context, r *connsdk.Requester, path string, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	nextURL, err := absoluteURL(r.BaseURL, path)
	if err != nil {
		return err
	}
	firstQuery := url.Values{"PageSize": []string{strconv.Itoa(pageSize)}}

	for page := 0; maxPages <= 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var query url.Values
		if page == 0 {
			query = firstQuery
		}
		resp, err := r.Do(ctx, http.MethodGet, nextURL, query, nil)
		if err != nil {
			return fmt.Errorf("twilio: read %s: %w", endpoint.recordsKey, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsKey)
		if err != nil {
			return fmt.Errorf("twilio: decode %s page: %w", endpoint.recordsKey, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "next_page_uri")
		if err != nil {
			return fmt.Errorf("twilio: decode %s next_page_uri: %w", endpoint.recordsKey, err)
		}
		next = strings.TrimSpace(next)
		if next == "" || next == "null" {
			return nil
		}
		nextURL, err = absoluteURL(r.BaseURL, next)
		if err != nil {
			return err
		}
	}
	return nil
}

// Per-stream record mappers — byte-for-byte port of legacy's
// twilio/streams.go mapRecord functions: every schema-declared field is
// written explicitly (item[key] resolves to nil when the raw item omits
// that key, matching legacy's exact map-literal construction), never a
// generic "copy only present keys" passthrough. Twilio's wire shape is
// already snake_case, so no camelCase rename/nested-field hoist is needed.

func messageRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"sid":                   item["sid"],
		"account_sid":           item["account_sid"],
		"messaging_service_sid": item["messaging_service_sid"],
		"date_created":          item["date_created"],
		"date_sent":             item["date_sent"],
		"date_updated":          item["date_updated"],
		"from":                  item["from"],
		"to":                    item["to"],
		"body":                  item["body"],
		"status":                item["status"],
		"direction":             item["direction"],
		"num_segments":          item["num_segments"],
		"num_media":             item["num_media"],
		"price":                 item["price"],
		"price_unit":            item["price_unit"],
		"error_code":            item["error_code"],
		"error_message":         item["error_message"],
	}
}

func callRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"sid":          item["sid"],
		"account_sid":  item["account_sid"],
		"date_created": item["date_created"],
		"date_updated": item["date_updated"],
		"start_time":   item["start_time"],
		"end_time":     item["end_time"],
		"from":         item["from"],
		"to":           item["to"],
		"status":       item["status"],
		"direction":    item["direction"],
		"duration":     item["duration"],
		"price":        item["price"],
		"price_unit":   item["price_unit"],
	}
}

func recordingRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"sid":          item["sid"],
		"account_sid":  item["account_sid"],
		"call_sid":     item["call_sid"],
		"date_created": item["date_created"],
		"date_updated": item["date_updated"],
		"start_time":   item["start_time"],
		"duration":     item["duration"],
		"status":       item["status"],
		"channels":     item["channels"],
		"source":       item["source"],
		"price":        item["price"],
		"price_unit":   item["price_unit"],
	}
}

func conferenceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"sid":           item["sid"],
		"account_sid":   item["account_sid"],
		"friendly_name": item["friendly_name"],
		"date_created":  item["date_created"],
		"date_updated":  item["date_updated"],
		"status":        item["status"],
		"region":        item["region"],
	}
}

func usageRecordRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"account_sid": item["account_sid"],
		"category":    item["category"],
		"description": item["description"],
		"start_date":  item["start_date"],
		"end_date":    item["end_date"],
		"count":       item["count"],
		"count_unit":  item["count_unit"],
		"usage":       item["usage"],
		"usage_unit":  item["usage_unit"],
		"price":       item["price"],
		"price_unit":  item["price_unit"],
	}
}

// absoluteURL resolves a possibly-relative reference (Twilio's
// host-relative next_page_uri, or the stream's own account-scoped path)
// against base's host and scheme. Twilio's next_page_uri already includes
// the /2010-04-01 prefix, so a "/"-prefixed reference is resolved against
// the host root rather than appended to base's path — byte-for-byte port
// of legacy's absoluteURL (twilio.go:288-313).
func absoluteURL(base, ref string) (string, error) {
	b, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("twilio: base_url is invalid: %w", err)
	}
	r, err := url.Parse(ref)
	if err != nil {
		return "", fmt.Errorf("twilio: path is invalid: %w", err)
	}
	if r.IsAbs() {
		return r.String(), nil
	}
	if strings.HasPrefix(ref, "/") {
		resolved := *b
		resolved.Path = r.Path
		resolved.RawQuery = r.RawQuery
		return resolved.String(), nil
	}
	basePath := strings.TrimRight(b.Path, "/")
	resolved := *b
	resolved.Path = basePath + "/" + ref
	resolved.RawQuery = r.RawQuery
	return resolved.String(), nil
}

// pageSizeFor mirrors legacy's twilioPageSize (twilio.go:336-349).
func pageSizeFor(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("twilio config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("twilio config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

// maxPagesFor mirrors legacy's twilioMaxPages (twilio.go:351-364).
func maxPagesFor(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("twilio config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, fmt.Errorf("twilio config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}
