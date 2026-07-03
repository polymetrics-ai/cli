// Package tallyprime implements the Tier-3 native TallyPrime source
// connector (architecture v2 design §B.7 Tier 3 — see
// internal/connectors/native/postgres for the golden migration reference).
// It is a Tier-3 build, not a declarative Tier-1/Tier-2 bundle, because
// docs/migration/quarantine.json classifies it NON_REST: "TallyPrime = local
// TDL envelope gateway (XML and JSON modes both envelope/RPC, POST to
// localhost:9000, no REST resources/pagination) — Tier-3 native build
// queued". There is no legacy internal/connectors/tally-prime package —
// unlike postgres/amazon-sqs this is a new build, not a migration, so there
// is no parity target.
//
// TallyPrime exposes exactly one wire endpoint (its local Gateway Server,
// default http://localhost:9000): every logical object (companies, ledgers,
// groups, stock items, vouchers) is a different TDL Export/Collection
// ENVELOPE/HEADER/BODY document POSTed to that same endpoint, never a
// distinct REST resource/path — the same "no REST resources/pagination"
// shape that makes amazon-sqs's SigV4/XML envelope Tier-2-ineligible, here
// for a local RPC/TDL protocol instead of a signed cloud API.
//
// Component split (mirrors postgres's connector.go/connection.go/reader.go/
// cataloger.go; no cdc.go — TallyPrime's Gateway Server is polled per Read
// with no subscription/webhook mechanism, so there is no CDC path to stub,
// unlike postgres's documented pglogrepl stub):
//   - connector.go (this file) — entry/wiring, Metadata, Write stub.
//   - connection.go — gateway_url/company config resolution, the envelope
//     HTTP POST helper (JSON and XML wire encodings), Check.
//   - reader.go — Read: builds one Export/Collection envelope per stream,
//     decodes the response, InitialState/incremental cursor for vouchers.
//   - cataloger.go — Catalog: the five core-object stream definitions
//     (companies/ledgers/groups/stock_items/vouchers) plus fixture-mode
//     canned catalog/rows.
//
// Still ships a defs bundle (internal/connectors/defs/tally-prime/
// {metadata.json,spec.json,api_surface.json,docs.md}) so identity/spec/docs
// stay uniform with every other connector; metadata.json sets
// capabilities.dynamic_schema: true and the bundle ships no streams.json
// (bundle.go's loadStreams only tolerates a missing streams.json when
// dynamic_schema is true) — a structural requirement of the Tier-3 loader,
// matching native/postgres's and native/amazon-sqs's identical precedent.
//
// A mode=fixture config (cfg.Config["mode"]=="fixture") short-circuits all
// network access so the conformance harness and unit tests can run with no
// live TallyPrime instance: in fixture mode Check succeeds, Catalog returns
// canned streams, and Read emits canned rows.
//
// NO init()/RegisterFactory/RegisterNativeLive call exists in this package
// (enforced by a grep-guard test, tally_prime_test.go TestNoInitRegistration)
// — the registration flip that wires native/tally-prime into the production
// registry is a later-wave change; this wave only builds and tests the
// package standalone.
package tallyprime

import (
	"context"
	"net/http"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

// Connector is the Tier-3 native TallyPrime source connector. It embeds
// engine.Base for Name()/Metadata()/Definition(), synthesized from the
// embedded defs/tally-prime bundle loaded once at construction (New), and
// implements Check/Catalog/Read itself since TallyPrime's TDL envelope
// protocol has no declarative equivalent.
//
// Client is an optional test seam (mirrors native/amazon-sqs's identical
// Client field): a nil Client means "build a fresh *http.Client with the
// configured http_timeout_seconds per request" (connection.go's do).
type Connector struct {
	engine.Base

	Client *http.Client
}

// New returns the TallyPrime connector as a connectors.Connector, loading
// its Definition()/Metadata() from the embedded defs/tally-prime bundle. New
// panics if the bundle fails to load — a broken build, not a runtime error
// (mirrors native/postgres.New's and native/amazon-sqs.New's identical
// contract).
func New() Connector {
	b, err := engine.Load(defs.FS, "tally-prime")
	if err != nil {
		panic("native/tally-prime: failed to load defs/tally-prime bundle: " + err.Error())
	}
	return Connector{Base: engine.NewBase(b)}
}

// Metadata overrides engine.Base's bundle-synthesized Metadata with a
// refined description; Capabilities are still whatever the bundle's
// metadata.json declares (single source of truth for capability flags), so
// this override only refines Description/DisplayName wording, never
// capability semantics (mirrors native/postgres's and native/amazon-sqs's
// identical Metadata override pattern).
func (c Connector) Metadata() connectors.Metadata {
	m := c.Base.Metadata()
	m.Description = "Reads TallyPrime accounting data (companies, ledgers, groups, stock items, vouchers) via TDL Export/Collection envelope requests POSTed to a locally-running TallyPrime Gateway Server. Read-only source; schema is discovered dynamically since TallyPrime has no static REST resource surface."
	return m
}

// Write is unsupported: this is a read-only source connector. TallyPrime's
// Import-mode envelopes (TALLYREQUEST=Import) can mutate company data, but
// this connector never builds one — capabilities.write is false and Write
// always returns ErrUnsupportedOperation.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// httpClient returns c.Client if set, else a fresh client using the
// configured http_timeout_seconds (default defaultHTTPTimeout).
func (c Connector) httpClient(cfg connectors.RuntimeConfig) *http.Client {
	if c.Client != nil {
		return c.Client
	}
	return &http.Client{Timeout: httpTimeout(cfg)}
}

// defaultHTTPTimeout is used when config http_timeout_seconds is absent or
// invalid.
const defaultHTTPTimeout = 30 * time.Second
