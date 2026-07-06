// Package postgres implements the Tier-3 native PostgreSQL source connector
// (architecture v2 design §B.7 Tier 3, PLAN.md T-17/B-17 — the golden
// migration reference for every future database/file/native connector). It
// is a database connector (family: db) built on github.com/jackc/pgx/v5
// (already in go.mod — no new dependency), following the mandated Tier-3
// component split: connector.go (entry, wiring), connection.go
// (config/DSN/identifier safety), reader.go (Read/InitialState), cataloger.go
// (Catalog/discovery + fixtures), cdc.go (documented CDC stub). Each file is
// well under the design's <400-line cap.
//
// Unlike a Tier-1/Tier-2 declarative bundle, this package implements
// connectors.Connector directly: Check/Catalog/Read/Write are hand-written
// Go, not derived from streams.json (there is none — this is a
// capabilities.dynamic_schema bundle, since a database's tables are
// discovered at runtime from information_schema, not declared ahead of
// time). It still ships a defs bundle
// (internal/connectors/defs/postgres/{metadata.json,spec.json,
// api_surface.json,docs.md}) so identity/spec/docs stay uniform with every
// other connector, and it embeds engine.Base — built from that bundle at
// construction — purely to serve Name()/Metadata()/Definition() (design
// §B.7: "they embed engine.Base which serves Definition() ... from the
// bundle"; Base does NOT provide Check/Catalog/Read/Write, which remain
// this package's own implementation).
//
// Capabilities:
//   - Check:   pgxpool connect + ping using host/port/database/username/
//     sslmode and the password secret (connection.go/cataloger.go).
//   - Catalog: discover tables and columns from information_schema and map
//     each PostgreSQL data_type to a coarse Field type (cataloger.go).
//   - Read:    snapshot SELECT over a stream, with optional
//     cursor-incremental filtering on a configurable cursor column
//     (reader.go; see StatefulReader below).
//   - Write:   not implemented; this is a read-only source for wave0 parity
//     with the legacy package. Capabilities.Write is false and Write
//     returns ErrUnsupportedOperation.
//
// CDC (change data capture) is a documented STUB (cdc.go): ReadCDC returns
// ErrUnsupportedOperation because full logical-replication CDC requires the
// pglogrepl dependency, a gated add not present in go.mod.
//
// A mode=fixture config (cfg.Config["mode"]=="fixture") short-circuits all
// network access so the conformance harness and unit tests can run with no
// live DB: in fixture mode Check succeeds, Catalog returns canned streams,
// and Read emits canned rows.
//
// NO init()/RegisterFactory call exists in this package
// in wave0 (enforced by a grep-guard test, postgres_test.go
// TestNoInitRegistration) — the registration flip that wires native/postgres
// into the production registry is a wave6 change; wave0 only builds and
// tests the package standalone, exactly as instructed.
package postgres

import (
	"context"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

// Connector is the Tier-3 native pm PostgreSQL source connector. It embeds
// engine.Base for Name()/Metadata()/Definition(), synthesized from the
// defs/postgres bundle loaded once at construction (New), and implements
// Check/Catalog/Read/Write itself (connection.go/cataloger.go/reader.go/
// cdc.go) since a database connector's tables are discovered dynamically,
// not declared in a streams.json.
type Connector struct {
	engine.Base
}

// New returns the PostgreSQL connector as a connectors.Connector, loading
// its Definition()/Metadata() from the embedded defs/postgres bundle. New
// panics if the bundle fails to load — the same "build-time guaranteed by
// connectorgen validate + tests" invariant engine.NewRegistry documents for
// its own bundle loading (design §C.2), since a bundle that fails to load
// here indicates a broken build, not a runtime/user error.
func New() Connector {
	b, err := engine.Load(defs.FS, "postgres")
	if err != nil {
		panic("native/postgres: failed to load defs/postgres bundle: " + err.Error())
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
	m.Description = "Reads PostgreSQL tables: discovers schemas/columns from information_schema, snapshots tables, and supports cursor-incremental reads on a configurable cursor column. Read-only source; CDC is a documented stub pending the gated pglogrepl dependency."
	return m
}

// Write is unsupported: this is a read-only source connector (wave0 parity
// with the legacy package; capabilities.write is false).
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
