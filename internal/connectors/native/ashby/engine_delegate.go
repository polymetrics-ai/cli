package ashby

import (
	"context"
	"sync"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

var (
	ashbyBundleOnce  sync.Once
	ashbyBundleValue engine.Bundle
	ashbyBundleErr   error
)

func ashbyBundle() engine.Bundle {
	ashbyBundleOnce.Do(func() {
		ashbyBundleValue, ashbyBundleErr = engine.Load(defs.FS, "ashby")
	})
	if ashbyBundleErr != nil {
		panic("native/ashby: failed to load defs/ashby bundle: " + ashbyBundleErr.Error())
	}
	return ashbyBundleValue
}

func ashbyEngineConnector() *engine.Connector {
	return engine.New(ashbyBundle(), nil)
}

// Definition exposes the generated Ashby bundle definition for inspect/docs
// while this native package still owns Ashby's custom POST cursor read loop.
func (c Connector) Definition() connectors.Definition {
	return ashbyEngineConnector().Definition()
}

// Manifest exposes the generated Ashby bundle manifest for inspect/docs.
func (c Connector) Manifest() connectors.Manifest {
	return ashbyEngineConnector().Manifest()
}

// CommandSurface exposes the generated, typed CLI surface for Ashby commands.
func (c Connector) CommandSurface() *connectors.CommandSurface {
	return ashbyEngineConnector().CommandSurface()
}

// ValidateWrite delegates typed Ashby reverse-ETL validation to the generated
// bundle. The bundle contains closed top-level JSON schemas and fixed endpoint
// paths; no generic HTTP passthrough is exposed by the native connector.
func (c Connector) ValidateWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) error {
	return engine.ValidateWrite(ctx, ashbyBundle(), req, records)
}

// DryRunWrite stages Ashby reverse-ETL records without network I/O. The actual
// mutation remains gated by the CLI's plan -> preview -> approval -> execute
// lifecycle.
func (c Connector) DryRunWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WritePreview, error) {
	return engine.DryRunWrite(ctx, ashbyBundle(), req, records, nil)
}

// Write executes only named Ashby write actions from writes.json. It does not
// accept raw methods, raw paths, or arbitrary request bodies.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return engine.Write(ctx, ashbyBundle(), req, records, nil)
}

// OperationDirectRead executes only named, bounded direct-read operations from
// operations.json. Search/file-metadata paths are fixed and schema-validated.
func (c Connector) OperationDirectRead(ctx context.Context, req connectors.OperationDirectReadRequest) (connectors.DirectReadResult, error) {
	return engine.OperationDirectRead(ctx, ashbyBundle(), req, nil)
}
