// Package faker implements the Tier-3 native "Sample Data" source connector
// (architecture v2 design §B.7 Tier 3 — see internal/connectors/native/postgres
// for the golden migration reference). Unlike every HTTP-based Tier-1/Tier-2
// connector, this package speaks no wire protocol at all: it deterministically
// generates sample users/purchases/products records in-process from a
// count/seed formula, exactly as the legacy internal/connectors/faker package
// does. It migrates to Tier 3 (not a declarative bundle) because the engine's
// declarative dialect has no "generate N synthetic records from a
// counter/seed formula" primitive — every streams.json construct assumes an
// HTTP request/response page to extract records from, which does not exist
// here (docs/migration/conventions.md §6 item 3: "the protocol is not
// HTTP/REST at all").
//
// Component split (mirroring postgres's connector.go/cataloger.go/reader.go):
//   - connector.go (this file) — entry/wiring, Metadata, Write stub.
//   - cataloger.go — Catalog: the fixed 3-stream catalog.
//   - reader.go — Read: the deterministic per-stream generation logic.
//
// There is no connection.go: this connector opens no connection of any kind
// (no DSN, no host, no credentials) — config parsing (count/seed) lives in
// reader.go next to the generation logic that consumes it, since that is its
// only caller.
//
// Still ships a defs bundle (internal/connectors/defs/faker/{metadata.json,
// spec.json,api_surface.json,docs.md}) so identity/spec/docs stay uniform
// with every other connector; metadata.json sets capabilities.dynamic_schema:
// true and the bundle ships no streams.json (bundle.go's loadStreams only
// tolerates a missing streams.json when dynamic_schema is true) — a
// structural requirement of the Tier-3 loader, not a claim that this
// connector's fixed 3-stream catalog is actually schema-discovered at
// runtime from anywhere (see docs.md's Known limits).
//
// NO init()/RegisterFactory/RegisterNativeLive call exists in this package
// (enforced by a grep-guard test, faker_test.go TestNoInitRegistration) — the
// registration flip that wires native/faker into the production registry is
// a wave6 change; this wave only builds and tests the package standalone.
package faker

import (
	"context"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

// Connector is the Tier-3 native Sample Data connector. It embeds
// engine.Base for Name()/Metadata()/Definition(), synthesized from the
// embedded defs/faker bundle loaded once at construction (New), and
// implements Check/Catalog/Read/Write itself since generation has no
// declarative equivalent.
type Connector struct {
	engine.Base
}

// New returns the Sample Data connector as a connectors.Connector, loading
// its Definition()/Metadata() from the embedded defs/faker bundle. New
// panics if the bundle fails to load — a broken build, not a runtime error
// (mirrors native/postgres.New's identical contract).
func New() Connector {
	b, err := engine.Load(defs.FS, "faker")
	if err != nil {
		panic("native/faker: failed to load defs/faker bundle: " + err.Error())
	}
	return Connector{Base: engine.NewBase(b)}
}

// Check always succeeds (subject to context cancellation): this connector
// makes no network calls and holds no connection to verify, matching
// legacy's Check exactly (faker.go:34).
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	return ctx.Err()
}

// Write is unsupported: this is a read-only source connector, matching
// legacy's Write returning ErrUnsupportedOperation (faker.go:99-101).
// Capabilities.Write is false.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
