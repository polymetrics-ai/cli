// Package microsoftentraid implements the Tier-2 StreamHook for the
// microsoft-entra-id quarantine-repair migration (docs/migration/
// quarantine.json's microsoft-entra-id ENGINE_GAP entry, conventions.md §1's
// Tier-2 hooks table).
//
// Microsoft Graph's list pagination is "@odata.nextLink" — a JSON key
// containing a literal dot — carrying the next page's FULL ABSOLUTE URL
// (with its own $skiptoken cursor already embedded). Legacy hand-rolls this
// exactly (internal/connectors/microsoft-entra-id/microsoft-entra-id.go's
// harvest/nextLink, read-only reference): GET the resource, extract
// value[], decode the top-level "@odata.nextLink" string directly (not via
// a dotted-path helper — the key's own literal dot makes dotted-path
// addressing ambiguous), and if non-empty, re-request that exact URL
// verbatim. The engine's declarative next_url pagination type reads its
// cursor via connsdk.StringAt's dotted-path parser, which splits on "."
// and therefore cannot address a literal key containing a dot — this is
// the confirmed ENGINE_GAP this hook resolves without an engine change (see
// defs/microsoft-entra-id/docs.md's Streams notes for the full reasoning).
//
// Only one hook interface is implemented (StreamHook), well under the
// 2-interface Tier-2 cap; auth stays fully declarative
// (oauth2_client_credentials in streams.json, dual when-gated candidates
// mirroring sharepoint-lists-enterprise/microsoft-teams).
package microsoftentraid

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

const (
	defaultPageSize = 100
	maxPageSize     = 999
)

// graphResource maps a stream name to the Microsoft Graph collection path
// (relative to base_url) it reads from — the exact routing table legacy's
// streamEndpoints declares (microsoft-entra-id/streams.go).
var graphResource = map[string]string{
	"users":             "/users",
	"groups":            "/groups",
	"applications":      "/applications",
	"serviceprincipals": "/servicePrincipals",
	"directoryroles":    "/directoryRoles",
}

func init() {
	engine.RegisterHooks("microsoft-entra-id", func() engine.Hooks { return New() })
}

// Hooks implements engine.StreamHook for the microsoft-entra-id bundle. It
// has no state of its own; every method is a pure function of its
// arguments.
type Hooks struct{}

// New returns a fresh microsoft-entra-id Hooks value (StreamHook
// implementation).
func New() *Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "microsoft-entra-id" }

// ReadStream drives Microsoft Graph's @odata.nextLink pagination for every
// stream this bundle declares. handled is false only for a stream name this
// hook does not recognize (should not happen for a correctly-authored
// bundle; returning handled=false rather than panicking keeps the
// declarative path as an honest fallback per the Hooks interface contract).
func (h *Hooks) ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (bool, error) {
	if err := ctx.Err(); err != nil {
		return true, err
	}

	resource, ok := graphResource[stream.Name]
	if !ok {
		return false, nil
	}
	schema := rt.Bundle.Schemas[stream.Name]
	if schema == nil {
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

	props := schema.Properties()
	path := resource
	query := url.Values{"$top": []string{strconv.Itoa(pageSize)}}

	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return true, err
		}

		resp, err := rt.Requester.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return true, fmt.Errorf("microsoft-entra-id: read %s: %w", stream.Name, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "value")
		if err != nil {
			return true, fmt.Errorf("microsoft-entra-id: decode %s page: %w", stream.Name, err)
		}
		for _, raw := range records {
			if err := ctx.Err(); err != nil {
				return true, err
			}
			mapped := mapRecord(stream.Name, raw)
			if err := emit(connectors.Record(projectBySchema(mapped, props))); err != nil {
				return true, err
			}
		}

		next, err := nextLink(resp.Body)
		if err != nil {
			return true, fmt.Errorf("microsoft-entra-id: decode %s nextLink: %w", stream.Name, err)
		}
		if strings.TrimSpace(next) == "" {
			return true, nil
		}
		// nextLink is an absolute URL that already carries the cursor and any
		// page-size hints; subsequent pages must not re-merge $top.
		path = next
		query = nil
	}
	return true, nil
}

// nextLink extracts the Microsoft Graph "@odata.nextLink" absolute URL from
// a collection response body. The key contains a literal dot, so the
// engine's dotted-path helpers cannot select it; decode the top-level
// object directly, exactly like legacy's own nextLink helper.
func nextLink(body []byte) (string, error) {
	var envelope struct {
		NextLink string `json:"@odata.nextLink"`
	}
	if len(body) == 0 {
		return "", nil
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return "", fmt.Errorf("decode graph envelope: %w", err)
	}
	return strings.TrimSpace(envelope.NextLink), nil
}

// pageSizeFor mirrors legacy's pageSize (microsoft-entra-id.go).
func pageSizeFor(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("microsoft-entra-id config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("microsoft-entra-id config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

// maxPagesFor mirrors legacy's maxPages (microsoft-entra-id.go).
func maxPagesFor(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("microsoft-entra-id config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, fmt.Errorf("microsoft-entra-id config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// mapRecord renames a raw Graph directory object's camelCase fields to the
// bundle's snake_case schema fields, byte-for-byte matching legacy's
// per-stream mapRecord functions (microsoft-entra-id/streams.go).
func mapRecord(stream string, raw map[string]any) map[string]any {
	switch stream {
	case "users":
		return map[string]any{
			"id":                  raw["id"],
			"display_name":        raw["displayName"],
			"user_principal_name": raw["userPrincipalName"],
			"given_name":          raw["givenName"],
			"surname":             raw["surname"],
			"mail":                raw["mail"],
			"job_title":           raw["jobTitle"],
			"department":          raw["department"],
			"office_location":     raw["officeLocation"],
			"mobile_phone":        raw["mobilePhone"],
			"account_enabled":     raw["accountEnabled"],
		}
	case "groups":
		return map[string]any{
			"id":                raw["id"],
			"display_name":      raw["displayName"],
			"description":       raw["description"],
			"mail":              raw["mail"],
			"mail_nickname":     raw["mailNickname"],
			"mail_enabled":      raw["mailEnabled"],
			"security_enabled":  raw["securityEnabled"],
			"visibility":        raw["visibility"],
			"created_date_time": raw["createdDateTime"],
		}
	case "applications":
		return map[string]any{
			"id":                raw["id"],
			"app_id":            raw["appId"],
			"display_name":      raw["displayName"],
			"description":       raw["description"],
			"sign_in_audience":  raw["signInAudience"],
			"publisher_domain":  raw["publisherDomain"],
			"created_date_time": raw["createdDateTime"],
		}
	case "serviceprincipals":
		return map[string]any{
			"id":                        raw["id"],
			"app_id":                    raw["appId"],
			"display_name":              raw["displayName"],
			"service_principal_type":    raw["servicePrincipalType"],
			"account_enabled":           raw["accountEnabled"],
			"app_owner_organization_id": raw["appOwnerOrganizationId"],
			"sign_in_audience":          raw["signInAudience"],
		}
	case "directoryroles":
		return map[string]any{
			"id":               raw["id"],
			"display_name":     raw["displayName"],
			"description":      raw["description"],
			"role_template_id": raw["roleTemplateId"],
		}
	default:
		return raw
	}
}

// projectBySchema keeps only the schema-declared properties from mapped,
// matching conventions.md §2's schema-as-projection rule.
func projectBySchema(mapped map[string]any, props []string) map[string]any {
	out := make(map[string]any, len(props))
	for _, name := range props {
		if v, ok := mapped[name]; ok {
			out[name] = v
		}
	}
	return out
}
