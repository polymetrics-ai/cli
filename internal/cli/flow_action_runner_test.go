package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/flow"
)

func TestConnectorFlowActionRunnerWritesReadsBackAndThenCheckpoints(t *testing.T) {
	ctx := testCtx(t)
	a := newFlowScopedWarehouseApp(t, ctx)
	target := &cliFlowActionTarget{}
	a.Registry().Register(target)
	_, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name: "flow-target", Connector: target.Name(), Secrets: map[string]string{"token": "fixture-token"},
	})
	require.NoError(t, err)

	plan, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
		Name:                  "flow-action-target",
		SourceTable:           "records",
		SourceConnection:      "acme",
		DestinationConnector:  target.Name(),
		DestinationCredential: "flow-target",
		Action:                "create",
		Mappings:              map[string]string{"id": "id"},
	})
	require.NoError(t, err)
	plan, _, err = a.PreviewReversePlan(ctx, plan.ID, nil)
	require.NoError(t, err)
	_, err = a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken, Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}})
	require.NoError(t, err)
	plan, err = a.GetReversePlan(plan.ID)
	require.NoError(t, err)
	require.NotEmpty(t, plan.AuthorizationReference)
	beforeWrites := target.writeCalls()
	sourceRecords, err := a.ReadActionSource(ctx, app.ActionSourceReadRequest{Table: "records", Connection: "acme"})
	require.NoError(t, err)
	require.NotEmpty(t, sourceRecords)

	manifestPath := writeManifestFile(t, fmt.Sprintf(`{
  "version": 1,
  "name": "approved-flow",
  "steps": [{
    "id": "create-targets",
    "kind": "action",
	"job": %q,
    "action_cfg": {
      "read_back_stream": "targets"
    }
  }]
}`, plan.ID))
	var stdout bytes.Buffer
	err = flowRun(ctx, config.Config{}, a, []string{"--file", manifestPath}, &stdout, true)
	require.NoError(t, err)
	assert.NotEmpty(t, stdout.String())
	assert.Equal(t, beforeWrites+1, target.writeCalls(), "the production runner must call the connector, not HTTPActionRunner")
	assert.Equal(t, []string{"validate", "preview", "write", "read"}, target.tailEvents(4))
	assert.True(t, target.lastWriteHadApproval(), "the production route must pass durable authorization evidence to the typed write")
	receipts := a.ListFlowActionReceipts()
	require.Len(t, receipts, 1)
	assert.Equal(t, "approved-flow", receipts[0].FlowName)
	assert.NotEmpty(t, receipts[0].PreparedExecutionIdentity)
	var result flow.RunResult
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &result))
	require.Len(t, result.Steps, 1)
	assert.Equal(t, receipts[0].PreparedExecutionIdentity, result.Steps[0].PreparedExecutionIdentity)
	checkpoints := &flow.FileCheckpointStore{Dir: a.ProjectDir()}
	status, err := checkpoints.Get("approved-flow", "create-targets")
	require.NoError(t, err)
	assert.Equal(t, "success", status, "checkpoint is written only after receipt-producing execution returns")
}

func TestConnectorFlowActionRunnerScopeDriftStopsBeforeTargetRequest(t *testing.T) {
	ctx := testCtx(t)
	a := newFlowScopedWarehouseApp(t, ctx)
	target := &cliFlowActionTarget{}
	a.Registry().Register(target)
	_, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name: "flow-target", Connector: target.Name(), Secrets: map[string]string{"token": "fixture-token"},
	})
	require.NoError(t, err)

	plan, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
		Name:                  "flow-action-target",
		SourceTable:           "records",
		SourceConnection:      "acme",
		DestinationConnector:  target.Name(),
		DestinationCredential: "flow-target",
		Action:                "create",
		Mappings:              map[string]string{"id": "id"},
	})
	require.NoError(t, err)
	plan, _, err = a.PreviewReversePlan(ctx, plan.ID, nil)
	require.NoError(t, err)
	_, err = a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken, Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}})
	require.NoError(t, err)
	plan, err = a.GetReversePlan(plan.ID)
	require.NoError(t, err)
	require.NotEmpty(t, plan.AuthorizationReference)
	driftReversePlanMapping(t, filepath.Dir(a.ProjectDir()), plan.ID)
	beforeWrites := target.writeCalls()
	beforeEvents := target.eventCount()

	manifestPath := writeManifestFile(t, fmt.Sprintf(`{
  "version": 1,
  "name": "scope-drift-flow",
  "steps": [{
    "id": "create-targets",
    "kind": "action",
	"job": %q,
    "action_cfg": {
      "read_back_stream": "targets"
    }
  }]
}`, plan.ID))
	var stdout bytes.Buffer
	err = flowRun(ctx, config.Config{}, a, []string{"--file", manifestPath}, &stdout, true)
	require.Error(t, err)
	assert.ErrorContains(t, err, "field_mappings")
	assert.Equal(t, beforeWrites, target.writeCalls(), "scope drift must produce zero connector writes")
	assert.Equal(t, beforeEvents, target.eventCount(), "scope drift must stop before validation, write, or read-back reaches the connector")
	assert.Empty(t, a.ListFlowActionReceipts(), "scope drift must produce no receipt")
}

type cliFlowActionTarget struct {
	mu            sync.Mutex
	writes        int
	events_       []string
	records       []connectors.Record
	writeApproval bool
	writeErr      error
}

func (*cliFlowActionTarget) Name() string { return "flow-cli-target" }

func (*cliFlowActionTarget) Metadata() connectors.Metadata {
	return connectors.Metadata{Capabilities: connectors.Capabilities{Write: true}}
}

func (c *cliFlowActionTarget) Manifest() connectors.Manifest {
	return connectors.Manifest{Metadata: c.Metadata(), WriteActions: []connectors.WriteActionSpec{{Name: "create", Method: "POST", Confirm: "destructive"}}}
}

func (*cliFlowActionTarget) Check(context.Context, connectors.RuntimeConfig) error { return nil }

func (*cliFlowActionTarget) Catalog(context.Context, connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{}, nil
}

func (c *cliFlowActionTarget) Read(_ context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	c.mu.Lock()
	c.events_ = append(c.events_, "read")
	records := append([]connectors.Record(nil), c.records...)
	c.mu.Unlock()
	if req.Stream != "targets" {
		return errors.New("unexpected read-back stream")
	}
	for _, record := range records {
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

func (c *cliFlowActionTarget) ValidateWrite(_ context.Context, _ connectors.WriteRequest, _ []connectors.Record) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events_ = append(c.events_, "validate")
	return nil
}

func (c *cliFlowActionTarget) DryRunWrite(_ context.Context, req connectors.WriteRequest, _ []connectors.Record) (connectors.WritePreview, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events_ = append(c.events_, "preview")
	return connectors.WritePreview{Digest: "flow-cli-target-preview", ApprovalTarget: cliFlowActionApprovalTarget(c.Name(), req)}, nil
}

func (c *cliFlowActionTarget) Write(_ context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events_ = append(c.events_, "write")
	c.writeApproval = req.Approval != nil
	if !c.writeApproval {
		return connectors.WriteResult{}, errors.New("durable approval evidence is required")
	}
	if err := req.Approval.Authorize(cliFlowActionApprovalTarget(c.Name(), req), "flow-cli-target-preview", time.Now().UTC()); err != nil {
		return connectors.WriteResult{}, err
	}
	if c.writeErr != nil {
		return connectors.WriteResult{}, c.writeErr
	}
	c.writes++
	c.records = append(c.records, records...)
	return connectors.WriteResult{RecordsWritten: len(records)}, nil
}

func (c *cliFlowActionTarget) setWriteError(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.writeErr = err
}

func (c *cliFlowActionTarget) writeCalls() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.writes
}

func (c *cliFlowActionTarget) tailEvents(count int) []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.events_) < count {
		return append([]string(nil), c.events_...)
	}
	return append([]string(nil), c.events_[len(c.events_)-count:]...)
}

func (c *cliFlowActionTarget) eventCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.events_)
}

func (c *cliFlowActionTarget) lastWriteHadApproval() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.writeApproval
}

func cliFlowActionApprovalTarget(name string, req connectors.WriteRequest) connectors.WriteApprovalTarget {
	return connectors.WriteApprovalTarget{
		Connector: name, Operation: req.Action, Method: "POST", TargetDigest: "flow-cli-target", CredentialRevision: req.Config.CredentialRevision,
		ConfigurationDigest: req.Config.ConfigurationDigest, Batchable: true, Scope: connectors.WriteApprovalScopeProject,
		Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	}
}
