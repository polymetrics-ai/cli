package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/flow"
)

// connectorFlowActionRunner is the sole production runner built by flowRun.
// It has no URL, HTTP client, SQL, or raw operation escape hatch.
type connectorFlowActionRunner struct {
	app            *app.App
	flowName       string
	manifestDigest string
}

var _ flow.StepActionRunner = (*connectorFlowActionRunner)(nil)
var _ flow.StepActionPreflight = (*connectorFlowActionRunner)(nil)

func (r *connectorFlowActionRunner) PreflightStep(ctx context.Context, step flow.FlowStep, _ string) error {
	if r == nil || r.app == nil {
		return errors.New("flow action runner requires an app")
	}
	if step.ActionCfg == nil {
		return errors.New("flow action configuration is required")
	}
	cfg := step.ActionCfg
	action := cfg.Action
	if action == "" {
		action = "upsert"
	}
	return r.app.ValidateAuthorizedFlowAction(ctx, app.FlowActionExecutionRequest{
		FlowName:               r.flowName,
		StepID:                 step.ID,
		ManifestDigest:         r.manifestDigest,
		SourceTable:            cfg.SourceTable,
		SourceConnection:       cfg.SourceConnection,
		DestinationTable:       cfg.DestinationTable,
		DestinationConnector:   cfg.DestinationConnector,
		DestinationCredential:  cfg.DestinationCredential,
		DestinationConfig:      cfg.DestinationConfig,
		Action:                 action,
		Mappings:               cfg.Mappings,
		AuthorizationReference: cfg.AuthorizationReference,
		ReadBackStream:         cfg.ReadBackStream,
	})
}

func (r *connectorFlowActionRunner) ExecuteStep(ctx context.Context, step flow.FlowStep, records []map[string]any, _ string, runID string) (flow.ActionResult, error) {
	if r == nil || r.app == nil {
		return flow.ActionResult{}, errors.New("flow action runner requires an app")
	}
	if step.ActionCfg == nil {
		return flow.ActionResult{}, errors.New("flow action configuration is required")
	}
	cfg := step.ActionCfg
	action := cfg.Action
	if action == "" {
		action = "upsert"
	}
	if cfg.DestinationTable == "" {
		return flow.ActionResult{}, fmt.Errorf("flow action step %q requires action_cfg.destination_table", step.ID)
	}
	if strings.TrimSpace(cfg.ReadBackStream) == "" {
		return flow.ActionResult{}, fmt.Errorf("flow action step %q requires action_cfg.read_back_stream", step.ID)
	}
	connectorRecords := make([]connectors.Record, len(records))
	for i, record := range records {
		connectorRecords[i] = connectors.Record(record)
	}
	result, err := r.app.ExecuteAuthorizedFlowAction(ctx, app.FlowActionExecutionRequest{
		FlowName:               r.flowName,
		StepID:                 step.ID,
		RunID:                  runID,
		ManifestDigest:         r.manifestDigest,
		SourceTable:            cfg.SourceTable,
		SourceConnection:       cfg.SourceConnection,
		DestinationTable:       cfg.DestinationTable,
		DestinationConnector:   cfg.DestinationConnector,
		DestinationCredential:  cfg.DestinationCredential,
		DestinationConfig:      cfg.DestinationConfig,
		Action:                 action,
		Mappings:               cfg.Mappings,
		AuthorizationReference: cfg.AuthorizationReference,
		ReadBackStream:         cfg.ReadBackStream,
		Records:                connectorRecords,
	})
	if err != nil {
		return flow.ActionResult{RecordsAttempted: result.RecordsAttempted, RecordsSucceeded: result.RecordsSucceeded, RecordsFailed: result.RecordsFailed, PreparedExecutionIdentity: result.PreparedExecutionIdentity}, err
	}
	return flow.ActionResult{RecordsAttempted: result.RecordsAttempted, RecordsSucceeded: result.RecordsSucceeded, RecordsFailed: result.RecordsFailed, ReceiptIDs: []string{result.ReceiptID}, PreparedExecutionIdentity: result.PreparedExecutionIdentity}, nil
}
