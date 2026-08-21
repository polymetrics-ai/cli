package ashby

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
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

// PreflightOperationDirectRead delegates the exact command binding to Ashby's
// declarative operation contract. Ashby owns its stream loop natively, but its
// fixed direct reads execute through the engine and must use the same
// no-network admission check as every engine-backed connector.
func (c Connector) PreflightOperationDirectRead(operation, method, path string, maxBytes int, outputPolicy string) error {
	return engine.PreflightOperationDirectRead(ashbyBundle(), operation, method, path, maxBytes, outputPolicy)
}

// PreflightOperationDirectReadBindings keeps Ashby's native adapter on the
// same closed command-binding contract as fully declarative connectors.
func (c Connector) PreflightOperationDirectReadBindings(operation string, pathFields, queryFields, bodyFields []string, rawBody bool) error {
	return engine.PreflightOperationDirectReadBindings(ashbyBundle(), operation, pathFields, queryFields, bodyFields, rawBody)
}

// ValidateWrite delegates typed Ashby reverse-ETL validation to the generated
// bundle. The bundle contains closed top-level JSON schemas and fixed endpoint
// paths; no generic HTTP passthrough is exposed by the native connector.
func (c Connector) ValidateWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) error {
	return engine.ValidateWrite(ctx, ashbyBundle(), req, records)
}

type ashbyEngineHooks struct{}

func (ashbyEngineHooks) ConnectorName() string { return "ashby" }

var (
	_ engine.PreparedWriteHook              = ashbyEngineHooks{}
	_ engine.PreparedWriteResponseValidator = ashbyEngineHooks{}
)

// PrepareWrite keeps Ashby's native success-envelope rule while moving its
// physical POST selection into the engine's declaration-owned approval plan.
// Each action names itself; the engine resolves its fixed method, path,
// headers, bounded body, and receipt projection from writes.json before any
// approval is minted.
func (ashbyEngineHooks) PrepareWrite(action engine.WriteAction, records []connectors.Record) (engine.PreparedWriteHookPlan, bool, error) {
	if len(action.PathFields) != 0 || strings.Contains(action.Path, "{{") || (action.BodyType != "" && action.BodyType != "json") || len(action.BodyFields) != 0 || len(action.BodySchema) != 0 || action.GraphQL != nil || action.Multipart != nil {
		return engine.PreparedWriteHookPlan{}, true, fmt.Errorf("ashby write action %q uses an unsupported request shape", action.Name)
	}
	plan := engine.PreparedWriteHookPlan{Records: make([]engine.PreparedWriteHookRecord, len(records))}
	for index, record := range records {
		sealed := make(connectors.Record, len(record))
		for key, value := range record {
			sealed[key] = value
		}
		plan.Records[index].Steps = []engine.PreparedWriteHookStep{{
			Action: action.Name,
			Record: sealed,
		}}
	}
	return plan, true, nil
}

func (ashbyEngineHooks) ValidatePreparedWriteResponse(_ engine.WriteAction, _ connectors.Record, response *connsdk.Response) error {
	if response == nil {
		return fmt.Errorf("ashby write did not return a provider response")
	}
	return ashbyValidateSuccessEnvelope(response.Body)
}

func (ashbyEngineHooks) ExecuteWrite(ctx context.Context, action engine.WriteAction, rec connectors.Record, rt *engine.Runtime) (bool, []*connsdk.Response, error) {
	if rt == nil || rt.Requester == nil {
		return true, nil, fmt.Errorf("ashby write runtime is unavailable")
	}
	if len(action.PathFields) != 0 || strings.Contains(action.Path, "{{") || (action.BodyType != "" && action.BodyType != "json") || len(action.BodyFields) != 0 || len(action.BodySchema) != 0 || action.GraphQL != nil || action.Multipart != nil {
		return true, nil, fmt.Errorf("ashby write action %q uses an unsupported request shape", action.Name)
	}
	var payload any
	if len(rec) > 0 {
		payload = map[string]any(rec)
	}
	resp, err := rt.Requester.Do(ctx, action.Method, action.Path, nil, payload)
	if err != nil {
		return true, ashbyResponseSlice(resp), err
	}
	if err := ashbyValidateSuccessEnvelope(resp.Body); err != nil {
		return true, ashbyResponseSlice(resp), err
	}
	return true, ashbyResponseSlice(resp), nil
}

func ashbyResponseSlice(response *connsdk.Response) []*connsdk.Response {
	if response == nil {
		return nil
	}
	return []*connsdk.Response{response}
}

// DryRunWrite stages Ashby reverse-ETL records without network I/O. The actual
// mutation remains gated by the CLI's plan -> preview -> approval -> execute
// lifecycle.
func (c Connector) DryRunWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WritePreview, error) {
	return engine.DryRunWrite(ctx, ashbyBundle(), req, records, ashbyEngineHooks{})
}

// Write executes only named Ashby write actions from writes.json. It does not
// accept raw methods, raw paths, or arbitrary request bodies.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return engine.Write(ctx, ashbyBundle(), req, records, ashbyEngineHooks{})
}

// OperationDirectRead executes only named, bounded direct-read operations from
// operations.json. Search/file-metadata paths are fixed and schema-validated.
func (c Connector) OperationDirectRead(ctx context.Context, req connectors.OperationDirectReadRequest) (connectors.DirectReadResult, error) {
	result, err := engine.OperationDirectRead(ctx, ashbyBundle(), req, nil)
	if err != nil {
		return result, err
	}
	if err := ashbyValidateSuccessEnvelopeValue(result.Body); err != nil {
		return result, fmt.Errorf("ashby direct read: %w", err)
	}
	return result, nil
}
