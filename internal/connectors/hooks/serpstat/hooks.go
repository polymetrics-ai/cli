// Package serpstat implements the serpstat bundle's StreamHook
// (docs/migration/quarantine.json's original ENGINE_GAP finding,
// conventions.md SS1's Tier-2 table: "sub-resource fan-out reads" /
// "whole-stream override"): Serpstat's real API is a JSON-RPC-over-HTTP
// endpoint (POST /v4, body `{"id":<page>,"method":"<procedure>","params":
// {...}}`) whose pagination state (the page number) lives INSIDE the
// request body -- engine/read.go's declarative read path always issues its
// request with a nil body (engine/bundle.go's StreamSpec.Body is declared
// but never read by read.go), so this cannot be expressed in streams.json
// alone. This ports internal/connectors/serpstat/serpstat.go's JSON-RPC
// body construction, in-body page-number pagination, and result.data
// record extraction verbatim via StreamHook.ReadStream, reusing rt.Requester
// (the engine's already-built *connsdk.Requester: base URL, the
// api_key_query "token" param, and headers are already resolved
// declaratively -- this hook adds ONLY the JSON-RPC body, never touches
// auth).
package serpstat

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

// mirror legacy serpstat.go's identically named constants.
const (
	defaultDomain   = "serpstat.com"
	defaultRegion   = "g_us"
	defaultPageSize = 10
	defaultMaxPages = 1
)

func init() {
	engine.RegisterHooks("serpstat", func() engine.Hooks { return Hooks{} })
}

// Hooks is serpstat's Tier-2 hook set: StreamHook only (auth is fully
// declarative via streams.json's api_key_query "token" param -- see docs.md
// "Auth setup").
type Hooks struct{}

func (Hooks) ConnectorName() string { return "serpstat" }

// jsonRPCMethod maps a stream name to its Serpstat JSON-RPC procedure
// (ported verbatim from serpstat.go's streamEndpoints).
var jsonRPCMethod = map[string]string{
	"domain_keywords":    "SerpstatDomainProcedure.getKeywords",
	"domain_competitors": "SerpstatDomainProcedure.getCompetitors",
	"domain_urls":        "SerpstatDomainProcedure.getDomainUrls",
}

// ReadStream implements engine.StreamHook, handling both declared streams
// (domain_keywords, domain_competitors) with handled=true always -- the
// declarative streams.json fallback is a structural "shadow" path never
// exercised by production traffic (conformance's dynamic checks run with
// Hooks=nil, so it never reaches this hook at all).
func (h Hooks) ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (bool, error) {
	name := stream.Name
	if name == "" {
		name = "domain_keywords"
	}
	method, ok := jsonRPCMethod[name]
	if !ok {
		return false, nil
	}

	pageSize, err := positiveInt(req.Config.Config["page_size"], defaultPageSize, 1, 1000, "page_size")
	if err != nil {
		return true, err
	}
	pages, err := parsePages(req.Config.Config["pages_to_fetch"])
	if err != nil {
		return true, err
	}
	domain := strings.TrimSpace(req.Config.Config["domain"])
	if domain == "" {
		domain = defaultDomain
	}
	region := strings.TrimSpace(req.Config.Config["region_id"])
	if region == "" {
		region = defaultRegion
	}

	for page := 1; pages == 0 || page <= pages; page++ {
		if err := ctx.Err(); err != nil {
			return true, err
		}
		body := map[string]any{
			"id":     page,
			"method": method,
			"params": map[string]any{"domain": domain, "se": region, "page": page, "size": pageSize},
		}
		resp, err := rt.Requester.Do(ctx, http.MethodPost, "", nil, body)
		if err != nil {
			return true, fmt.Errorf("read serpstat %s page %d: %w", name, page, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "result.data")
		if err != nil {
			return true, fmt.Errorf("decode serpstat %s page %d: %w", name, page, err)
		}
		for _, rec := range records {
			if err := ctx.Err(); err != nil {
				return true, err
			}
			if err := emit(connectors.Record(rec)); err != nil {
				return true, err
			}
		}
		if len(records) < pageSize {
			return true, nil
		}
	}
	return true, nil
}

// positiveInt mirrors legacy serpstat.go's positiveInt helper exactly:
// an empty raw value returns def; otherwise raw must parse as an integer in
// [min,max].
func positiveInt(raw string, def, min, max int, name string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return def, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < min || n > max {
		return 0, fmt.Errorf("%s must be between %d and %d", name, min, max)
	}
	return n, nil
}

// parsePages mirrors legacy serpstat.go's parsePages helper exactly: an
// empty raw value returns defaultMaxPages (1); otherwise raw must parse as a
// non-negative integer (0 means unbounded).
func parsePages(raw string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return defaultMaxPages, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, errors.New("pages_to_fetch must be a non-negative integer")
	}
	return n, nil
}
