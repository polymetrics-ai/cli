// Package plaid implements the plaid bundle's StreamHook + CheckHook
// (docs/migration/quarantine.json: AUTH_COMPLEX, "Plaid's entire API is
// POST-only with ALL credentials (client_id/secret) and ALL pagination/
// filter state (count/offset/country_codes) carried in the JSON request
// body, never in headers or query params"): engine/read.go's declarative
// read path always calls Requester.Do with a nil body, and StreamSpec.Body
// (bundle.go), though declared on the struct, is never read back out
// anywhere in read.go — there is no way to express a body-carried
// credential or pagination state in streams.json alone. This ports legacy
// internal/connectors/plaid/plaid.go's authBody/readPage/Check verbatim,
// reusing rt.Requester (the engine's already-built *connsdk.Requester: base
// URL/retry/rate-limit plumbing already resolved; base.auth is mode:none
// since there is no header/query credential to declare).
package plaid

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

const (
	defaultPageSize = 100
	defaultMaxPages = 3
)

func init() {
	engine.RegisterHooks("plaid", func() engine.Hooks { return Hooks{} })
}

// Hooks is plaid's Tier-2 hook set: StreamHook (both reads) + CheckHook.
type Hooks struct{}

func (Hooks) ConnectorName() string { return "plaid" }

type streamEndpoint struct {
	path        string
	recordsPath string
	mapRecord   func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{
	"institutions": {path: "institutions/get", recordsPath: "institutions", mapRecord: institutionRecord},
	"categories":   {path: "categories/get", recordsPath: "categories", mapRecord: categoryRecord},
}

// ReadStream implements engine.StreamHook, handling both declared streams
// (institutions, categories) with handled=true always — the declarative
// streams.json fallback is a structural "shadow" path exercised only by
// conformance's dynamic checks (Hooks=nil there), never here, matching
// monday's documented precedent (docs.md "Streams notes").
func (h Hooks) ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (bool, error) {
	name := stream.Name
	if name == "" {
		name = "institutions"
	}
	endpoint, ok := streamEndpoints[name]
	if !ok {
		return false, nil
	}

	body, err := authBody(req.Config)
	if err != nil {
		return true, err
	}

	if name == "categories" {
		_, err := h.readPage(ctx, rt.Requester, endpoint, body, emit)
		return true, err
	}

	pageSize, err := pageSize(req.Config)
	if err != nil {
		return true, err
	}
	maxPages, err := maxPages(req.Config)
	if err != nil {
		return true, err
	}

	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return true, err
		}
		body["count"] = pageSize
		body["offset"] = offset
		body["country_codes"] = countryCodes(req.Config)
		count, err := h.readPage(ctx, rt.Requester, endpoint, body, emit)
		if err != nil {
			return true, err
		}
		if count < pageSize {
			return true, nil
		}
		offset += pageSize
	}
	return true, nil
}

// Check implements engine.CheckHook: a bounded categories/get call (with the
// auth body) confirms credentials and connectivity, matching legacy's
// Check (plaid.go).
func (h Hooks) Check(ctx context.Context, cfg connectors.RuntimeConfig, rt *engine.Runtime) (bool, error) {
	body, err := authBody(cfg)
	if err != nil {
		return true, err
	}
	if _, err := rt.Requester.Do(ctx, http.MethodPost, "categories/get", nil, body); err != nil {
		return true, fmt.Errorf("check plaid: %w", err)
	}
	return true, nil
}

func (h Hooks) readPage(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, body map[string]any, emit func(connectors.Record) error) (int, error) {
	resp, err := r.Do(ctx, http.MethodPost, endpoint.path, nil, body)
	if err != nil {
		return 0, fmt.Errorf("read plaid %s: %w", endpoint.path, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return 0, fmt.Errorf("decode plaid %s: %w", endpoint.path, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return 0, err
		}
	}
	return len(records), nil
}

// authBody builds the {client_id, secret} body fields every Plaid request
// carries (plaid.go's authBody, ported verbatim). Secret values flow only
// into this outgoing JSON body; never logged, never in an error string.
func authBody(cfg connectors.RuntimeConfig) (map[string]any, error) {
	clientID := strings.TrimSpace(cfg.Secrets["client_id"])
	secret := strings.TrimSpace(cfg.Secrets["secret"])
	if clientID == "" || secret == "" {
		return nil, errors.New("plaid connector requires secrets client_id and secret")
	}
	return map[string]any{"client_id": clientID, "secret": secret}, nil
}

func countryCodes(cfg connectors.RuntimeConfig) []string {
	values := splitCSV(cfg.Config["country_codes"])
	if len(values) == 0 {
		return []string{"US"}
	}
	return values
}

func splitCSV(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func institutionRecord(item map[string]any) connectors.Record {
	return connectors.Record{"institution_id": item["institution_id"], "name": item["name"], "country_codes": joinAny(item["country_codes"])}
}

func categoryRecord(item map[string]any) connectors.Record {
	return connectors.Record{"category_id": item["category_id"], "group": item["group"], "hierarchy": joinAny(item["hierarchy"])}
}

func joinAny(v any) string {
	list, ok := v.([]any)
	if !ok {
		return ""
	}
	parts := make([]string, 0, len(list))
	for _, item := range list {
		parts = append(parts, fmt.Sprintf("%v", item))
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return intConfig(cfg, "page_size", defaultPageSize, 1, 500)
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	return intConfig(cfg, "max_pages", defaultMaxPages, 0, 10000)
}

func intConfig(cfg connectors.RuntimeConfig, key string, def, min, max int) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config[key]))
	if raw == "" {
		return def, nil
	}
	if key == "max_pages" && (raw == "all" || raw == "unlimited") {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < min {
		return 0, fmt.Errorf("plaid config %s must be an integer >= %d", key, min)
	}
	if max > 0 && value > max {
		return max, nil
	}
	return value, nil
}
