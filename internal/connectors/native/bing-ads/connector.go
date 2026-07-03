// Package bingads implements the Tier-3 native Bing Ads (Microsoft
// Advertising) source connector (conventions.md §1/§6: escalated from a
// Tier-2 hook-set attempt that needed 3 hook interfaces and ~570 lines, both
// past the Tier-2 caps — see defs/bing-ads/docs.md's "Why Tier 3, not Tier
// 2" section for the full reasoning). It follows the mandated Tier-3
// component split: connector.go (entry, wiring), connection.go
// (config resolution, requester/auth building), reader.go (Read/
// InitialState/StreamHook-equivalent routing), cataloger.go (Catalog +
// fixture generation). There is no cdc.go: Bing Ads has no change-data-
// capture concept at all (unlike postgres, which documents a genuine
// pglogrepl-gated CDC stub), so this file is simply omitted rather than
// padded with an inapplicable stub.
//
// Unlike a Tier-1/Tier-2 declarative bundle, this package implements
// connectors.Connector directly: Check/Catalog/Read are hand-written Go, not
// derived from streams.json (streams.json exists purely as
// documentation/schema-reference — see defs/bing-ads/docs.md). It still
// ships a defs bundle (internal/connectors/defs/bing-ads/{metadata.json,
// spec.json,streams.json,schemas/*.json,api_surface.json,docs.md}) so
// identity/spec/docs/schema stay uniform with every other connector, and it
// embeds engine.Base — built from that bundle at construction — purely to
// serve Name()/Metadata()/Definition() (engine.Base does NOT provide
// Check/Catalog/Read/Write, which remain this package's own
// implementation).
//
// Capabilities:
//   - Check:   a bounded AccountsInfo/Query POST confirms the OAuth
//     exchange, DeveloperToken, and connectivity (connection.go/reader.go).
//   - Catalog: a static 5-stream catalog (accounts/users/campaigns/
//     ad_groups/ads) — Bing Ads' schema is known ahead of time, not
//     discovered at runtime, so capabilities.dynamic_schema is false.
//   - Read:    POST-body query endpoints across two REST services (Customer
//     Management, Campaign Management), with per-account-id fan-out for
//     campaigns (reader.go).
//   - Write:   not implemented; read-only source for parity with the legacy
//     package. Capabilities.Write is false and Write returns
//     ErrUnsupportedOperation.
//
// A mode=fixture config (cfg.Config["mode"]=="fixture") short-circuits all
// network access so the conformance harness and unit tests can run with no
// live Microsoft Advertising credentials: in fixture mode Check succeeds
// and Read emits canned per-stream records (cataloger.go).
//
// NO init()/RegisterFactory/RegisterNativeLive call exists in this package
// in wave0 (enforced by a grep-guard test, bing_ads_test.go
// TestNoInitRegistration, mirroring native/postgres's identical guard) — the
// registration flip that wires native/bing-ads into the production registry
// is a wave6 change; wave0 only builds and tests the package standalone.
package bingads

import (
	"context"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

// connectorName is the registry key / catalog alias target, matching legacy
// bing_ads.go's connectorName constant.
const connectorName = "bing-ads"

// Connector is the Tier-3 native pm Bing Ads source connector. It embeds
// engine.Base for Name()/Metadata()/Definition(), synthesized from the
// embedded defs/bing-ads bundle loaded once at construction (New), and
// implements Check/Catalog/Read itself (connection.go/reader.go/
// cataloger.go).
type Connector struct {
	engine.Base
	// Client overrides the HTTP client used by every requester this
	// connector builds (token exchange and both REST services). Left nil in
	// production; injectable for tests.
	Client HTTPDoer
}

// New returns the Bing Ads connector as a connectors.Connector, loading its
// Definition()/Metadata() from the embedded defs/bing-ads bundle. New
// panics if the bundle fails to load, mirroring native/postgres's identical
// "build-time guaranteed" invariant (a bundle that fails to load here
// indicates a broken build, not a runtime/user error).
func New() Connector {
	b, err := engine.Load(defs.FS, connectorName)
	if err != nil {
		panic("native/bing-ads: failed to load defs/bing-ads bundle: " + err.Error())
	}
	return Connector{Base: engine.NewBase(b)}
}

// Metadata overrides engine.Base's bundle-synthesized Metadata with the
// legacy-shaped description text, matching the pre-migration
// connectors.Metadata field-for-field (parity target); Capabilities are
// still whatever the bundle's metadata.json declares (single source of
// truth for capability flags), so this override only refines
// Description/DisplayName wording, never capability semantics.
func (c Connector) Metadata() connectors.Metadata {
	m := c.Base.Metadata()
	m.Description = "Reads Microsoft Advertising (Bing Ads) accounts, users, campaigns, ad groups, and ads through the v13 Customer Management and Campaign Management REST APIs. Read-only."
	return m
}

// Write is unsupported: this is a read-only source connector (parity with
// the legacy package; capabilities.write is false).
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
