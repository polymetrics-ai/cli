// Package postgres implements the Tier-3 native PostgreSQL source connector
// (architecture v2 design §B.7 Tier 3, PLAN.md T-17/B-17 — the golden
// migration reference for every future database/file/native connector). It
// is a database connector (family: db) built on github.com/jackc/pgx/v5 and
// the PostgreSQL logical-replication protocol library github.com/jackc/pglogrepl,
// following the mandated Tier-3
// component split: connector.go (entry, wiring), connection.go
// (config/DSN/identifier safety), reader.go (Read/InitialState), cataloger.go
// (Catalog/discovery + fixtures), cdc.go (fail-closed logical-replication
// foundation). Each file is well under the design's <400-line cap.
//
// Unlike a Tier-1/Tier-2 declarative bundle, this package implements
// connectors.Connector directly: Check/Catalog/Read/Write are hand-written
// Go, not derived from streams.json (there is none — this is a
// capabilities.dynamic_schema bundle, since a database's tables are
// discovered at runtime from PostgreSQL's system catalogs, not declared ahead of
// time). It still ships a defs bundle
// (internal/connectors/defs/postgres/{metadata.json,spec.json,
// api_surface.json,database.json,docs.md}) so identity/spec/docs stay uniform
// with every other connector. database.json is a typed policy declaration
// only: it does not register a driver or promote write/CDC capability. The
// connector embeds engine.Base — built from that bundle at construction — to
// serve its bundle-derived identity and definition base. Definition adds this
// package's dynamic-catalog snapshot transport declaration; Base does NOT
// provide Check/Catalog/Read/Write, which remain this package's own
// implementation.
//
// Capabilities:
//   - Check:   pgxpool connect + ping using host/port/database/username/
//     sslmode and the password secret (connection.go/cataloger.go).
//   - Catalog: discover configured-schema base tables, ordered columns, keys,
//     and supported native/logical types from pg_catalog (typed_catalog.go),
//     then derive the legacy Field projection only for compatibility.
//   - Read:    snapshot SELECT over a stream, with optional
//     cursor-incremental filtering on a configurable cursor column
//     (reader.go; see StatefulReader below).
//   - Write:   not implemented; this is a read-only source for wave0 parity
//     with the legacy package. Capabilities.Write is false and Write
//     returns ErrUnsupportedOperation.
//
// CDC uses PostgreSQL 14+ pgoutput v2 streamed transaction staging (cdc.go).
//
// A mode=fixture config (cfg.Config["mode"]=="fixture") short-circuits all
// network access so the conformance harness and unit tests can run with no
// live DB: in fixture mode Check succeeds, Catalog returns canned streams,
// and Read emits canned rows.
//
// NO init()/RegisterFactory call exists in this package
// in wave0 (enforced by capability_surface_test.go's
// TestNoInitRegistration grep guard) — the registration flip that wires
// native/postgres into the production registry is a wave6 change; wave0 only builds and
// tests the package standalone, exactly as instructed.
package postgres

import (
	"context"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

// Connector is the Tier-3 native pm PostgreSQL source connector. It embeds
// engine.Base for Name()/Metadata() and the bundle-derived portion of
// Definition(), synthesized from the defs/postgres bundle loaded once at
// construction (New), and implements Check/Catalog/Read/Write itself
// (connection.go/cataloger.go/reader.go/cdc.go) since a database connector's
// tables are discovered dynamically, not declared in a streams.json.
type Connector struct {
	engine.Base

	// databaseDefinition is the immutable database.json projection loaded with
	// the same embedded bundle that supplies Base. Catalog discovery uses it
	// directly so native/logical typing cannot become a second, disconnected
	// PostgreSQL model.
	databaseDefinition database.Definition
}

// postgresCapabilityOverride is a declarative capability row. Future native
// capability lanes add a row rather than interleaving another override in
// Metadata's composition control flow.
type postgresCapabilityOverride struct {
	name   string
	value  bool
	target func(*connectors.Capabilities) *bool
}

var postgresCapabilityOverrides = []postgresCapabilityOverride{
	{
		name:  "cdc",
		value: true,
		target: func(capabilities *connectors.Capabilities) *bool {
			return &capabilities.CDC
		},
	},
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
	if b.Database == nil {
		panic("native/postgres: defs/postgres bundle is missing database.json")
	}
	return Connector{Base: engine.NewBase(b), databaseDefinition: *b.Database}
}

// Metadata overrides engine.Base's bundle-synthesized Metadata with the
// legacy-shaped description text, matching the pre-migration
// connectors.Metadata field-for-field (parity target). Capabilities remain
// owned by metadata.json; the current table repeats the proven native CDC
// capability so this connector's legacy metadata projection stays aligned.
func (c Connector) Metadata() connectors.Metadata {
	m := c.Base.Metadata()
	m.Description = "Reads PostgreSQL tables: dynamically discovers schemas/columns from PostgreSQL system catalogs, snapshots tables, supports cursor-incremental reads, and supports PostgreSQL 14+ logical-replication CDC. Read-only source."
	for _, override := range postgresCapabilityOverrides {
		*override.target(&m.Capabilities) = override.value
	}
	return m
}

// Definition adds the PostgreSQL-owned bounded snapshot transport declaration
// to the bundle-derived connector definition. The dynamic catalog means there
// is no static table list to place in a JSON bundle; the one logical snapshot
// stream resolves its relation from the checkpoint source scope at execution.
// Registration remains explicit through RegisterSnapshotTransportSource, so a
// declaration cannot become executable merely by constructing a connector.
func (c Connector) Definition() connectors.Definition {
	definition := c.Base.Definition()
	definition.SyncTransport = &connectors.SyncTransportDescriptor{
		Source: postgresSnapshotTransportDescriptor(),
	}
	return definition
}

func (c Connector) Manifest() connectors.Manifest {
	manifest := c.BundleManifest()
	for i := range manifest.SecretFields {
		if manifest.SecretFields[i].Name == "password" {
			manifest.SecretFields[i].RequiredWhen = "mode is not fixture"
			manifest.SecretFields[i].Description = "Fixture mode does not open a source connection."
		}
	}
	manifest.AuthModes = []connectors.AuthModeSpec{{
		Name:         "password",
		Description:  "Live connections require password authentication; peer/socket and client-certificate modes, including ambient certificates, are unsupported.",
		ConfigFields: []string{"host", "database", "username"},
		SecretFields: []string{"password"},
		Read:         true,
	}}
	return manifest
}

// Write is unsupported: this is a read-only source connector (wave0 parity
// with the legacy package; capabilities.write is false).
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
