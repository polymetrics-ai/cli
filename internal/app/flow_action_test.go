package app_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
)

func TestExecuteAuthorizedFlowActionAcknowledgesReadsBackThenRecordsReceipt(t *testing.T) {
	ctx := context.Background()
	a, sourcePlan := setupApprovedReversePlan(t, ctx)
	target := &flowActionTarget{}
	a.Registry().Register(target)
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name: "flow-target", Connector: target.Name(), Secrets: map[string]string{"token": "fixture-token"},
	}); err != nil {
		t.Fatalf("AddCredential(flow target) error = %v", err)
	}

	plan, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
		Name:                  "flow-action-target",
		SourceTable:           sourcePlan.SourceTable,
		SourceConnection:      sourcePlan.SourceConnection,
		DestinationConnector:  target.Name(),
		DestinationCredential: "flow-target",
		Action:                "create",
		Mappings:              map[string]string{"id": "id", "email": "email"},
	})
	if err != nil {
		t.Fatalf("PlanReverseETL(flow authorization) error = %v", err)
	}
	plan, _, err = a.PreviewReversePlan(ctx, plan.ID, nil)
	if err != nil {
		t.Fatalf("PreviewReversePlan(flow authorization) error = %v", err)
	}
	if _, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken, Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}}); err != nil {
		t.Fatalf("RunReverseETL(flow authorization) error = %v", err)
	}
	authorizations := a.ListAuthorizations()
	if len(authorizations) != 1 {
		t.Fatalf("ListAuthorizations() = %#v, want one durable authorization", authorizations)
	}
	target.receiptCount = func() int { return len(a.ListFlowActionReceipts()) }
	beforeWrites := target.writeCalls()

	records, err := a.ReadActionSource(ctx, app.ActionSourceReadRequest{Table: sourcePlan.SourceTable, Connection: sourcePlan.SourceConnection})
	if err != nil {
		t.Fatalf("ReadActionSource() error = %v", err)
	}
	result, err := a.ExecuteAuthorizedFlowAction(ctx, app.FlowActionExecutionRequest{
		FlowName:               "close-stale-issues",
		StepID:                 "close",
		RunID:                  "flow-run-1",
		SourceTable:            sourcePlan.SourceTable,
		SourceConnection:       sourcePlan.SourceConnection,
		DestinationTable:       "flow-action-target",
		DestinationConnector:   target.Name(),
		DestinationCredential:  "flow-target",
		Action:                 "create",
		Mappings:               map[string]string{"id": "id", "email": "email"},
		AuthorizationReference: authorizations[0].Reference,
		ReadBackStream:         "targets",
		Records:                records,
	})
	if err != nil {
		t.Fatalf("ExecuteAuthorizedFlowAction() error = %v", err)
	}
	if result.RecordsSucceeded != len(records) || result.ReceiptID == "" {
		t.Fatalf("ExecuteAuthorizedFlowAction() = %+v, want acknowledged records and receipt", result)
	}
	if got := target.writeCalls(); got != beforeWrites+1 {
		t.Fatalf("provider writes = %d, want %d; typed connector write was not dispatched exactly once", got, beforeWrites+1)
	}
	if got := target.events(); len(got) < 4 || got[len(got)-4] != "validate" || got[len(got)-3] != "preview" || got[len(got)-2] != "write" || got[len(got)-1] != "read" {
		t.Fatalf("connector event order = %#v, want validate -> preview -> write -> read before receipt", got)
	}
	if !target.lastWriteHadApproval() {
		t.Fatal("connector write did not receive durable authorization evidence")
	}
	if target.sawReceiptBeforeReadBack() {
		t.Fatal("receipt was observable by the connector before write acknowledgement and read-back completed")
	}
	receipts := a.ListFlowActionReceipts()
	if len(receipts) != 1 || receipts[0].ID != result.ReceiptID || receipts[0].RunID != "flow-run-1" {
		t.Fatalf("flow receipts = %#v, want one persisted acknowledged receipt", receipts)
	}
}

func TestExecuteAuthorizedFlowActionScopeDriftRefusesBeforeProviderWrite(t *testing.T) {
	ctx := context.Background()
	a, sourcePlan := setupApprovedReversePlan(t, ctx)
	target := &flowActionTarget{}
	a.Registry().Register(target)
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{Name: "flow-target", Connector: target.Name(), Secrets: map[string]string{"token": "fixture-token"}}); err != nil {
		t.Fatalf("AddCredential(flow target) error = %v", err)
	}
	plan, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
		Name:                  "flow-action-target",
		SourceTable:           sourcePlan.SourceTable,
		SourceConnection:      sourcePlan.SourceConnection,
		DestinationConnector:  target.Name(),
		DestinationCredential: "flow-target",
		Action:                "create",
		Mappings:              map[string]string{"id": "id", "email": "email"},
	})
	if err != nil {
		t.Fatalf("PlanReverseETL(flow authorization) error = %v", err)
	}
	plan, _, err = a.PreviewReversePlan(ctx, plan.ID, nil)
	if err != nil {
		t.Fatalf("PreviewReversePlan(flow authorization) error = %v", err)
	}
	if _, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken, Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}}); err != nil {
		t.Fatalf("RunReverseETL(flow authorization) error = %v", err)
	}
	authorization := a.ListAuthorizations()[0]
	records, err := a.ReadActionSource(ctx, app.ActionSourceReadRequest{Table: sourcePlan.SourceTable, Connection: sourcePlan.SourceConnection})
	if err != nil {
		t.Fatalf("ReadActionSource() error = %v", err)
	}
	beforeWrites := target.writeCalls()
	_, err = a.ExecuteAuthorizedFlowAction(ctx, app.FlowActionExecutionRequest{
		FlowName:               "close-stale-issues",
		StepID:                 "close",
		RunID:                  "flow-run-2",
		SourceTable:            sourcePlan.SourceTable,
		SourceConnection:       sourcePlan.SourceConnection,
		DestinationTable:       "flow-action-target",
		DestinationConnector:   target.Name(),
		DestinationCredential:  "flow-target",
		Action:                 "create",
		Mappings:               map[string]string{"id": "changed-id", "email": "email"},
		AuthorizationReference: authorization.Reference,
		ReadBackStream:         "targets",
		Records:                records,
	})
	var changed *app.AuthorizationScopeChangedError
	if !errors.As(err, &changed) || changed.Property != "field_mappings" {
		t.Fatalf("ExecuteAuthorizedFlowAction(scope drift) error = %T %v, want field_mappings scope refusal", err, err)
	}
	if got := target.writeCalls(); got != beforeWrites {
		t.Fatalf("provider writes after scope-drift refusal = %d, want %d", got, beforeWrites)
	}
	if got := len(a.ListFlowActionReceipts()); got != 0 {
		t.Fatalf("receipts after scope-drift refusal = %d, want zero", got)
	}
}

func TestExecuteAuthorizedFlowActionReadBackFailurePersistsNoReceipt(t *testing.T) {
	ctx := context.Background()
	a, sourcePlan := setupApprovedReversePlan(t, ctx)
	target := &flowActionTarget{readErr: errors.New("target read-back failed")}
	a.Registry().Register(target)
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{Name: "flow-target", Connector: target.Name(), Secrets: map[string]string{"token": "fixture-token"}}); err != nil {
		t.Fatalf("AddCredential(flow target) error = %v", err)
	}
	plan, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
		Name:                  "flow-action-target",
		SourceTable:           sourcePlan.SourceTable,
		SourceConnection:      sourcePlan.SourceConnection,
		DestinationConnector:  target.Name(),
		DestinationCredential: "flow-target",
		Action:                "create",
		Mappings:              map[string]string{"id": "id", "email": "email"},
	})
	if err != nil {
		t.Fatalf("PlanReverseETL(flow authorization) error = %v", err)
	}
	plan, _, err = a.PreviewReversePlan(ctx, plan.ID, nil)
	if err != nil {
		t.Fatalf("PreviewReversePlan(flow authorization) error = %v", err)
	}
	if _, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken, Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}}); err != nil {
		t.Fatalf("RunReverseETL(flow authorization) error = %v", err)
	}
	records, err := a.ReadActionSource(ctx, app.ActionSourceReadRequest{Table: sourcePlan.SourceTable, Connection: sourcePlan.SourceConnection})
	if err != nil {
		t.Fatalf("ReadActionSource() error = %v", err)
	}
	beforeWrites := target.writeCalls()
	_, err = a.ExecuteAuthorizedFlowAction(ctx, app.FlowActionExecutionRequest{
		FlowName:               "close-stale-issues",
		StepID:                 "close",
		RunID:                  "flow-run-3",
		SourceTable:            sourcePlan.SourceTable,
		SourceConnection:       sourcePlan.SourceConnection,
		DestinationTable:       "flow-action-target",
		DestinationConnector:   target.Name(),
		DestinationCredential:  "flow-target",
		Action:                 "create",
		Mappings:               map[string]string{"id": "id", "email": "email"},
		AuthorizationReference: a.ListAuthorizations()[0].Reference,
		ReadBackStream:         "targets",
		Records:                records,
	})
	if err == nil || err.Error() != "target read-back failed" {
		t.Fatalf("ExecuteAuthorizedFlowAction(read-back failure) error = %v, want read-back refusal", err)
	}
	if got := target.writeCalls(); got != beforeWrites+1 {
		t.Fatalf("provider writes after read-back failure = %d, want %d", got, beforeWrites+1)
	}
	if got := len(a.ListFlowActionReceipts()); got != 0 {
		t.Fatalf("receipts before failed read-back = %d, want zero", got)
	}
}

type flowActionTarget struct {
	mu            sync.Mutex
	writes        int
	events_       []string
	records       []connectors.Record
	readErr       error
	receiptCount  func() int
	earlyReceipt  bool
	writeApproval bool
}

func (*flowActionTarget) Name() string { return "flow-action-target" }

func (*flowActionTarget) Metadata() connectors.Metadata {
	return connectors.Metadata{Capabilities: connectors.Capabilities{Write: true}}
}

func (c *flowActionTarget) Manifest() connectors.Manifest {
	return connectors.Manifest{Metadata: c.Metadata(), WriteActions: []connectors.WriteActionSpec{{Name: "create", Method: "POST", Confirm: "destructive"}}}
}

func (*flowActionTarget) Check(context.Context, connectors.RuntimeConfig) error { return nil }

func (*flowActionTarget) Catalog(context.Context, connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{}, nil
}

func (c *flowActionTarget) Read(_ context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	c.mu.Lock()
	c.events_ = append(c.events_, "read")
	if c.receiptCount != nil && c.receiptCount() != 0 {
		c.earlyReceipt = true
	}
	readErr := c.readErr
	records := append([]connectors.Record(nil), c.records...)
	c.mu.Unlock()
	if readErr != nil {
		return readErr
	}
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

func (c *flowActionTarget) ValidateWrite(_ context.Context, _ connectors.WriteRequest, _ []connectors.Record) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events_ = append(c.events_, "validate")
	return nil
}

func (c *flowActionTarget) DryRunWrite(_ context.Context, req connectors.WriteRequest, _ []connectors.Record) (connectors.WritePreview, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events_ = append(c.events_, "preview")
	return connectors.WritePreview{Digest: "flow-action-target-preview", ApprovalTarget: flowActionApprovalTarget(c.Name(), req)}, nil
}

func (c *flowActionTarget) Write(_ context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events_ = append(c.events_, "write")
	c.writeApproval = req.Approval != nil
	if !c.writeApproval {
		return connectors.WriteResult{}, errors.New("durable approval evidence is required")
	}
	if err := req.Approval.Authorize(flowActionApprovalTarget(c.Name(), req), "flow-action-target-preview", time.Now().UTC()); err != nil {
		return connectors.WriteResult{}, err
	}
	if c.receiptCount != nil && c.receiptCount() != 0 {
		c.earlyReceipt = true
	}
	c.writes++
	c.records = append(c.records, records...)
	return connectors.WriteResult{RecordsWritten: len(records)}, nil
}

func (c *flowActionTarget) writeCalls() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.writes
}

func (c *flowActionTarget) events() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]string(nil), c.events_...)
}

func (c *flowActionTarget) sawReceiptBeforeReadBack() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.earlyReceipt
}

func (c *flowActionTarget) lastWriteHadApproval() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.writeApproval
}

func flowActionApprovalTarget(name string, req connectors.WriteRequest) connectors.WriteApprovalTarget {
	return connectors.WriteApprovalTarget{
		Connector: name, Operation: req.Action, Method: "POST", TargetDigest: "flow-action-target", CredentialRevision: req.Config.CredentialRevision,
		ConfigurationDigest: req.Config.ConfigurationDigest, Batchable: true, Scope: connectors.WriteApprovalScopeProject,
		Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	}
}
