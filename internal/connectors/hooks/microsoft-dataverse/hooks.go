// Package microsoftdataverse implements the Tier-2 StreamHook for the
// microsoft-dataverse migration (docs/migration/conventions.md §1's Tier-2
// hooks table), porting internal/connectors/microsoft-dataverse (read-only
// reference, unchanged) almost verbatim.
//
// Microsoft Dataverse's Web API list pagination is "@odata.nextLink" — a
// JSON key containing a literal dot — carrying the next page's FULL
// ABSOLUTE URL. The engine's declarative next_url pagination type reads its
// cursor via connsdk.StringAt's dotted-path parser, which splits on "." and
// therefore cannot address a literal key containing a dot. This is the same
// confirmed ENGINE_GAP already resolved the identical way for
// microsoft-entra-id/microsoft-lists/microsoft-teams (see
// defs/microsoft-dataverse/docs.md's Streams notes).
//
// Only one hook interface is implemented (StreamHook), well under the
// 2-interface Tier-2 cap; auth stays fully declarative
// (oauth2_client_credentials in streams.json, dual when-gated candidates
// mirroring microsoft-entra-id/sharepoint-lists-enterprise/microsoft-teams).
package microsoftdataverse

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

const (
	defaultPageSize = 100
	maxPageSize     = 5000
)

// entityEndpoints maps a stream name to its Dataverse entity-set resource
// path and the raw-field routing legacy's per-stream mapRecord functions use
// (microsoft-dataverse/streams.go, read-only reference).
var entityEndpoints = map[string]struct {
	resource string
	idField  string
	nameKeys []string
}{
	"accounts":      {resource: "accounts", idField: "accountid", nameKeys: []string{"name"}},
	"contacts":      {resource: "contacts", idField: "contactid", nameKeys: []string{"fullname", "name"}},
	"leads":         {resource: "leads", idField: "leadid", nameKeys: []string{"fullname", "subject", "name"}},
	"opportunities": {resource: "opportunities", idField: "opportunityid", nameKeys: []string{"name"}},
	"systemusers":   {resource: "systemusers", idField: "systemuserid", nameKeys: []string{"fullname", "name"}},
}

func init() {
	engine.RegisterHooks("microsoft-dataverse", func() engine.Hooks { return New() })
}

// Hooks implements engine.StreamHook for the microsoft-dataverse bundle. It
// has no state of its own; every method is a pure function of its
// arguments.
type Hooks struct{}

// New returns a fresh microsoft-dataverse Hooks value (StreamHook
// implementation).
func New() *Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "microsoft-dataverse" }

// ReadStream drives Microsoft Dataverse's @odata.nextLink pagination for
// every stream this bundle declares, porting legacy's harvest/nextLink
// (microsoft-dataverse.go) exactly. handled is false only for a stream name
// this hook does not recognize.
func (h *Hooks) ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (bool, error) {
	if err := ctx.Err(); err != nil {
		return true, err
	}

	endpoint, ok := entityEndpoints[stream.Name]
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

	path := endpoint.resource
	query := url.Values{"$top": []string{strconv.Itoa(pageSize)}}

	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return true, err
		}

		resp, err := rt.Requester.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return true, fmt.Errorf("microsoft-dataverse: read %s: %w", stream.Name, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "value")
		if err != nil {
			return true, fmt.Errorf("microsoft-dataverse: decode %s page: %w", stream.Name, err)
		}
		for _, raw := range records {
			if err := ctx.Err(); err != nil {
				return true, err
			}
			if err := emit(mapRecord(endpoint, raw)); err != nil {
				return true, err
			}
		}

		next, err := nextLink(resp.Body)
		if err != nil {
			return true, fmt.Errorf("microsoft-dataverse: decode %s nextLink: %w", stream.Name, err)
		}
		if strings.TrimSpace(next) == "" {
			return true, nil
		}
		// nextLink is an absolute URL that already carries the $skiptoken
		// cursor; subsequent pages must not re-merge $top.
		path = next
		query = nil
	}
	return true, nil
}

// nextLink extracts the Dataverse "@odata.nextLink" absolute URL from an
// entity-set collection response body. The key contains a literal dot, so
// the engine's dotted-path helpers cannot select it; decode the top-level
// object directly, exactly like legacy's own nextLink helper.
func nextLink(body []byte) (string, error) {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var root map[string]any
	if err := dec.Decode(&root); err != nil {
		return "", err
	}
	next, _ := root["@odata.nextLink"].(string)
	return next, nil
}

// mapRecord renders the legacy baseRecord shape (id/name/email/created_on/
// modified_on) for raw, matching microsoft-dataverse/streams.go's
// accountRecord/contactRecord/leadRecord/opportunityRecord/systemUserRecord
// + baseRecord field-for-field, including each stream's name-field fallback
// chain and systemusers' internalemailaddress email fallback.
func mapRecord(endpoint struct {
	resource string
	idField  string
	nameKeys []string
}, raw map[string]any) connectors.Record {
	var name any
	for _, key := range endpoint.nameKeys {
		if v := raw[key]; v != nil {
			name = v
			break
		}
	}
	email := raw["emailaddress1"]
	if email == nil {
		email = raw["internalemailaddress"]
	}
	return connectors.Record{
		"id":          raw[endpoint.idField],
		"name":        name,
		"email":       email,
		"created_on":  raw["createdon"],
		"modified_on": raw["modifiedon"],
	}
}

// pageSizeFor mirrors legacy's pageSize (microsoft-dataverse.go).
func pageSizeFor(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("microsoft-dataverse config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("microsoft-dataverse config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

// maxPagesFor mirrors legacy's maxPages (microsoft-dataverse.go).
func maxPagesFor(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("microsoft-dataverse config max_pages must be a non-negative integer: %w", err)
	}
	if value < 0 {
		return 0, fmt.Errorf("microsoft-dataverse config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}
