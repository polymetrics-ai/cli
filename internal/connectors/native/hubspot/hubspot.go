// Package hubspot implements the narrow Tier-3 extension required for
// HubSpot's tenant-defined CRM objects. The bundle remains the source of
// connector identity, auth and operation ledger data; this package owns only
// provider discovery and fixed-path object reads that cannot be static.
package hubspot

import (
	"context"
	"errors"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/discovery"
	"polymetrics.ai/internal/connectors/engine"
)

const connectorName = "hubspot"

// Connector keeps the bundle runtime and discovery declaration together. The
// driver has no provider-specific branches; hubspotProvider is the thin
// provider declaration that supplies List, Describe and one field converter.
type Connector struct {
	engine.Base
	bundle engine.Bundle
	driver *discovery.Driver
}

var _ connectors.Connector = (*Connector)(nil)
var _ connectors.ManifestProvider = (*Connector)(nil)
var _ connectors.DefinitionProvider = (*Connector)(nil)

// New loads the embedded HubSpot bundle and configures the shared discovery
// driver. A broken declaration is a build-time invariant, so construction
// mirrors other native connector constructors and panics rather than exposing
// an impossible partially configured runtime.
func New() *Connector {
	bundle, err := engine.Load(defs.FS, connectorName)
	if err != nil {
		panic("native/hubspot: failed to load defs/hubspot bundle: " + err.Error())
	}
	driver, err := discovery.New(discovery.Spec{
		Connector:          connectorName,
		Fallback:           standardObjects(),
		FallbackPrimaryKey: []string{"hs_object_id"},
		FallbackCursor:     "hs_lastmodifieddate",
		Converter:          hubspotFieldSchema,
	})
	if err != nil {
		panic("native/hubspot: invalid discovery declaration: " + err.Error())
	}
	return &Connector{Base: engine.NewBase(bundle), bundle: bundle, driver: driver}
}

// Check uses the bundle's common authenticated runtime. metadata.json does
// not advertise Check until HubSpot supplies a purpose-built bounded endpoint;
// callers may still use Catalog as the actual authenticated discovery check.
func (c *Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	return engine.Check(ctx, c.bundle, cfg, nil)
}

func (c *Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	runtime, err := engine.NewRuntime(ctx, c.bundle, cfg, nil)
	if err != nil {
		return connectors.Catalog{}, err
	}
	return c.driver.Catalog(ctx, cfg, hubspotProvider{runtime: runtime})
}

// Manifest retains HubSpot's bundle-declared credential fields and risk
// contract in its manual despite the streams themselves being discovered.
func (c *Connector) Manifest() connectors.Manifest {
	return c.BundleManifest()
}

// Write is intentionally unsupported. Dynamic discovery enables source reads
// only and does not make any reverse-ETL HubSpot operation executable.
func (c *Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

var errUnknownStream = errors.New("HubSpot stream is not in the discovered catalog")
