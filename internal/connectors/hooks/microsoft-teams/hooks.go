// Package microsoftteams implements the microsoft-teams bundle's Tier-2
// StreamHook (wave2 quarantine repair, docs/migration/quarantine.json's
// "microsoft-teams" ENGINE_GAP entry): Microsoft Graph's @odata.nextLink
// pagination cursor is an absolute URL read from a response-body key that
// contains a literal "." ("@odata.nextLink"). The engine's declarative
// next_url pagination type reads its cursor via connsdk.StringAt's
// dotted-path traversal (engine/paginate.go), which necessarily treats any
// "." in a path as a nesting separator -- there is no way to address a
// literal dotted key with that parser. This ports legacy
// internal/connectors/microsoft-teams/microsoft-teams.go's harvest/nextLink
// loop verbatim (~150 lines, well under the 300-line Tier-2 soft target,
// docs/migration/conventions.md §1). Only one hook interface is implemented
// (StreamHook); auth is fully declarative (oauth2_client_credentials in
// streams.json) and needs no AuthHook at all.
package microsoftteams

import (
	"bytes"
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

// graphStreamPaths mirrors legacy's graphStreamEndpoints routing table
// (microsoft-teams/streams.go) -- the Graph collection path segment
// (relative to base_url) each stream reads from.
var graphStreamPaths = map[string]string{
	"users":                    "/users",
	"groups":                   "/groups",
	"channels":                 "/teams/getAllChannels",
	"team_device_usage_report": "/reports/getTeamsDeviceUsageUserDetail",
}

func init() {
	engine.RegisterHooks("microsoft-teams", func() engine.Hooks { return Hooks{} })
}

// Hooks is the microsoft-teams bundle's Tier-2 hook set: StreamHook only. It
// has no state of its own; every method is a pure function of its
// arguments.
type Hooks struct{}

func (Hooks) ConnectorName() string { return "microsoft-teams" }

// ReadStream drives Microsoft Graph's @odata.nextLink pagination for every
// stream this bundle declares. handled is always true for a recognized
// stream name; an unrecognized name returns handled=false so the
// declarative fallback stays an honest path per the Hooks interface
// contract (should not happen for a correctly authored bundle).
func (h Hooks) ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (bool, error) {
	if err := ctx.Err(); err != nil {
		return true, err
	}

	name := stream.Name
	if name == "" {
		name = "users"
	}
	path, ok := graphStreamPaths[name]
	if !ok {
		return false, nil
	}

	schema := rt.Bundle.Schemas[name]
	if schema == nil {
		return false, nil
	}
	props := schema.Properties()

	query := url.Values{}
	if name == "team_device_usage_report" {
		// The Teams device usage report requires the aggregation period
		// (legacy microsoft-teams.go:133-136).
		query.Set("period", devicePeriod(req.Config))
		query.Set("$format", "application/json")
	}

	maxPages := maxPagesFor(req.Config)
	return true, h.harvest(ctx, rt.Requester, path, query, maxPages, props, emit)
}

// harvest follows Microsoft Graph's @odata.nextLink pagination. Graph
// collections return {value:[...], "@odata.nextLink":"<absolute url>"}; the
// next page is the nextLink URL verbatim -- connsdk.Requester treats an
// absolute http(s) path as-is, so the nextLink is passed straight back in.
// Byte-for-byte port of legacy's harvest (microsoft-teams.go:148-187).
func (h Hooks) harvest(ctx context.Context, r *connsdk.Requester, firstPath string, firstQuery url.Values, maxPages int, props []string, emit func(connectors.Record) error) error {
	path := firstPath
	query := firstQuery
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("microsoft-teams: read %s: %w", firstPath, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "value")
		if err != nil {
			return fmt.Errorf("microsoft-teams: decode %s page: %w", firstPath, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(connectors.Record(projectBySchema(item, props))); err != nil {
				return err
			}
		}
		// The pagination key "@odata.nextLink" contains a literal dot, so it
		// cannot be read via connsdk.StringAt's dotted-path traversal;
		// decode the literal key directly.
		next, err := nextLink(resp.Body)
		if err != nil {
			return fmt.Errorf("microsoft-teams: decode %s nextLink: %w", firstPath, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		// nextLink is an absolute URL carrying its own query (skiptoken
		// etc.); pass it as the path with no extra query so it is used
		// verbatim.
		path = next
		query = nil
	}
	return nil
}

// projectBySchema keeps only the schema-declared properties from raw,
// matching conventions.md §2's schema-as-projection rule -- microsoft-teams'
// schemas are derived field-for-field from legacy's mapRecord functions,
// which rename camelCase Graph fields to snake_case (e.g. displayName ->
// display_name); the schema property names ARE the record output names, so
// this reads raw Graph fields by their SCHEMA (already-renamed) name. Graph
// itself never emits snake_case keys, so this hook renames explicitly
// rather than relying on schema projection's exact-key-match default.
func projectBySchema(raw map[string]any, props []string) map[string]any {
	out := make(map[string]any, len(props))
	for _, name := range props {
		if v, ok := graphFieldFor(raw, name); ok {
			out[name] = v
		}
	}
	return out
}

// graphFieldFor maps a schema (snake_case) property name back to its raw
// Microsoft Graph (camelCase) field name, mirroring legacy's per-stream
// graph*Record functions (microsoft-teams/streams.go) field-for-field.
func graphFieldFor(raw map[string]any, schemaField string) (any, bool) {
	graphKey, ok := graphFieldNames[schemaField]
	if !ok {
		return nil, false
	}
	v, ok := raw[graphKey]
	return v, ok
}

// graphFieldNames is the union of every stream's schema-field -> raw-Graph-
// field mapping (legacy's graphUserRecord/graphGroupRecord/
// graphChannelRecord/graphDeviceUsageRecord). Field names are unique enough
// across streams that a single flat map is unambiguous, since
// projectBySchema only ever looks up the current stream's own declared
// properties.
var graphFieldNames = map[string]string{
	"id":                  "id",
	"display_name":        "displayName",
	"user_principal_name": "userPrincipalName",
	"mail":                "mail",
	"job_title":           "jobTitle",
	"office_location":     "officeLocation",
	"mobile_phone":        "mobilePhone",
	"account_enabled":     "accountEnabled",
	"description":         "description",
	"mail_nickname":       "mailNickname",
	"visibility":          "visibility",
	"created_date_time":   "createdDateTime",
	"security_enabled":    "securityEnabled",
	"mail_enabled":        "mailEnabled",
	"email":               "email",
	"membership_type":     "membershipType",
	"web_url":             "webUrl",
	"last_activity_date":  "lastActivityDate",
	"is_deleted":          "isDeleted",
	"used_web":            "usedWeb",
	"used_windows_phone":  "usedWindowsPhone",
	"used_android_phone":  "usedAndroidPhone",
	"used_i_os":           "usedIOS",
	"used_mac":            "usedMac",
	"report_period":       "reportPeriod",
}

// nextLink reads the literal "@odata.nextLink" key from a Graph response
// body. It is an absolute URL (carrying its own $skiptoken) or empty on the
// last page. Byte-for-byte port of legacy's nextLink
// (microsoft-teams.go:381-391).
func nextLink(body []byte) (string, error) {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var envelope struct {
		NextLink string `json:"@odata.nextLink"`
	}
	if err := dec.Decode(&envelope); err != nil {
		return "", fmt.Errorf("decode graph envelope: %w", err)
	}
	return envelope.NextLink, nil
}

// devicePeriod resolves the Teams device usage report aggregation period.
// Graph accepts D7, D30, D90, D180; default D7. Byte-for-byte port of
// legacy's devicePeriod (microsoft-teams.go:357-365).
func devicePeriod(cfg connectors.RuntimeConfig) string {
	if cfg.Config != nil {
		switch strings.ToUpper(strings.TrimSpace(cfg.Config["period"])) {
		case "D7", "D30", "D90", "D180":
			return strings.ToUpper(strings.TrimSpace(cfg.Config["period"]))
		}
	}
	return "D7"
}

// maxPagesFor mirrors legacy's graphMaxPages (microsoft-teams.go:367-377):
// permissive parse, never errors -- an empty/"all"/"unlimited"/malformed/
// negative value means unbounded (0).
func maxPagesFor(cfg connectors.RuntimeConfig) int {
	raw := strings.ToLower(strings.TrimSpace(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0
	}
	return n
}
