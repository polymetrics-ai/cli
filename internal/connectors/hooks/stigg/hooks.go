// Package stigg implements the stigg bundle's StreamHook
// (docs/migration/quarantine.json's original ENGINE_GAP finding,
// conventions.md SS1's Tier-2 table: "whole-stream override"): Stigg's real
// API is a GraphQL-over-HTTP endpoint (POST /graphql, body carrying a
// query string; records extracted from data.<field>) -- engine/read.go's
// declarative read path always issues its request with a nil body
// (engine/bundle.go's StreamSpec.Body is declared but never read by
// read.go), so this cannot be expressed in streams.json alone. This ports
// internal/connectors/stigg/stigg.go's GraphQL query construction and
// data.<field> record extraction verbatim via StreamHook.ReadStream (all 4
// streams: products, plans, customers, subscriptions), reusing rt.Requester
// (the engine's already-built *connsdk.Requester: base URL and the bearer
// Authorization header are already resolved declaratively -- this hook adds
// ONLY the GraphQL query body, never touches auth).
package stigg

import (
	"context"
	"fmt"
	"net/http"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

func init() {
	engine.RegisterHooks("stigg", func() engine.Hooks { return Hooks{} })
}

// Hooks is stigg's Tier-2 hook set: StreamHook only (auth is fully
// declarative via streams.json's bearer mode -- see docs.md "Auth setup").
type Hooks struct{}

func (Hooks) ConnectorName() string { return "stigg" }

// streamQuery describes one GraphQL root query field (ported verbatim from
// stigg.go's streamEndpoints: query text + records.path).
type streamQuery struct {
	query       string
	recordsPath string
}

// streamQueries is the routing table for stigg's 4 read streams, ported
// verbatim from stigg.go's streamEndpoints map (query text and
// recordsPath only; field selection/mapping is identical to the raw
// GraphQL response shape, so no separate mapRecord is needed here --
// engine.readOneSequence's own schema-projection step selects exactly the
// same id/refId/displayName/status (or id/refId/customerId/status) fields
// legacy's copyRecord selected, via each stream's schemas/<stream>.json).
var streamQueries = map[string]streamQuery{
	"products":      {query: "query PolymetricsProducts { products { id refId displayName status } }", recordsPath: "data.products"},
	"plans":         {query: "query PolymetricsPlans { plans { id refId displayName status } }", recordsPath: "data.plans"},
	"customers":     {query: "query PolymetricsCustomers { customers { id refId displayName status } }", recordsPath: "data.customers"},
	"subscriptions": {query: "query PolymetricsSubscriptions { subscriptions { id refId customerId status } }", recordsPath: "data.subscriptions"},
}

// ReadStream implements engine.StreamHook, handling every declared stream
// (products, plans, customers, subscriptions) with handled=true always --
// the declarative streams.json fallback is a structural "shadow" path never
// exercised by production traffic (conformance's dynamic checks run with
// Hooks=nil, so it never reaches this hook at all).
func (h Hooks) ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (bool, error) {
	name := stream.Name
	if name == "" {
		name = "products"
	}
	sq, ok := streamQueries[name]
	if !ok {
		return false, nil
	}
	if err := ctx.Err(); err != nil {
		return true, err
	}

	resp, err := rt.Requester.Do(ctx, http.MethodPost, "/graphql", nil, map[string]any{"query": sq.query})
	if err != nil {
		return true, fmt.Errorf("read stigg %s: %w", name, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, sq.recordsPath)
	if err != nil {
		return true, fmt.Errorf("decode stigg %s: %w", name, err)
	}
	for _, rec := range records {
		if err := ctx.Err(); err != nil {
			return true, err
		}
		if err := emit(connectors.Record(rec)); err != nil {
			return true, err
		}
	}
	return true, nil
}
