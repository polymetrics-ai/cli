// Package mysql implements the Tier-3 native MySQL source connector. It uses
// the MySQL wire protocol for connection checks, dynamic catalog discovery,
// bounded snapshot and incremental reads.
package mysql

import (
	"context"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

// Connector is a dynamic-schema, read-only MySQL source. engine.Base supplies
// bundle-backed identity/definition metadata; all database operations remain
// explicit native code because SQL and binlog replication are not HTTP streams.
type Connector struct {
	engine.Base
}

// New creates a connector from the embedded MySQL definition. A malformed
// embedded bundle is a build invariant violation, so it panics rather than
// presenting a runtime configuration error.
func New() Connector {
	bundle, err := engine.Load(defs.FS, "mysql")
	if err != nil {
		panic("native/mysql: failed to load defs/mysql bundle: " + err.Error())
	}
	return NewFromBundle(bundle)
}

// NewFromBundle constructs MySQL from the manifest-selected execution bundle.
func NewFromBundle(bundle engine.Bundle) Connector {
	return Connector{Base: engine.NewBase(bundle)}
}

// Metadata keeps capability fields sourced exclusively from metadata.json.
func (c Connector) Metadata() connectors.Metadata {
	m := c.Base.Metadata()
	m.Description = "Native MySQL source connector for wire-protocol checks, dynamic schemas, and bounded reads. Read-only source."
	return m
}

func (c Connector) Manifest() connectors.Manifest {
	return c.BundleManifest()
}

// Write is unsupported because this connector is a read-only source.
func (c Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
