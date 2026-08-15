package app_test

import (
	"context"
	"errors"
	"path/filepath"
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
		ManifestDigest:         "fmd_fixture_1",
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
		ManifestDigest:         "fmd_fixture_2",
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
		ManifestDigest:         "fmd_fixture_3",
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

func TestAuthorizedFlowActionPreparedIdentityBindsPayloadAndReachesReceipt(t *testing.T) {
	ctx := context.Background()
	a, target, req := newPreparedFlowActionFixture(t, ctx, "prepared-flow-1")

	prepared, err := a.PrepareAuthorizedFlowAction(ctx, req)
	if err != nil {
		t.Fatalf("PrepareAuthorizedFlowAction() error = %v", err)
	}
	if prepared.Identity == "" || prepared.FiringID != req.RunID {
		t.Fatalf("prepared action = %+v, want safe identity and firing identifiers", prepared)
	}

	drifted := req
	drifted.Records = append([]connectors.Record(nil), req.Records...)
	drifted.Records[0] = cloneTestRecord(drifted.Records[0])
	drifted.Records[0]["email"] = "changed@example.test"
	driftedPrepared, err := a.PrepareAuthorizedFlowAction(ctx, drifted)
	if err != nil {
		t.Fatalf("PrepareAuthorizedFlowAction(payload drift) error = %v", err)
	}
	if driftedPrepared.Identity == prepared.Identity {
		t.Fatalf("prepared identity = %q for two payloads, want payload-bound identities", prepared.Identity)
	}
	manifestDrift := req
	manifestDrift.ManifestDigest = "fmd_fixture_changed_query"
	manifestPrepared, err := a.PrepareAuthorizedFlowAction(ctx, manifestDrift)
	if err != nil {
		t.Fatalf("PrepareAuthorizedFlowAction(manifest drift) error = %v", err)
	}
	if manifestPrepared.Identity == prepared.Identity {
		t.Fatalf("prepared identity = %q for two manifests, want manifest-bound identities", prepared.Identity)
	}

	result, err := a.ExecutePreparedFlowAction(ctx, prepared)
	if err != nil {
		t.Fatalf("ExecutePreparedFlowAction() error = %v", err)
	}
	if result.PreparedExecutionIdentity != prepared.Identity || result.FiringID != prepared.FiringID {
		t.Fatalf("execution result = %+v, want prepared identity/firing propagated", result)
	}
	receipts := a.ListFlowActionReceipts()
	if len(receipts) != 1 || receipts[0].PreparedExecutionIdentity != prepared.Identity || receipts[0].FiringID != prepared.FiringID {
		t.Fatalf("flow action receipts = %+v, want safe prepared execution evidence", receipts)
	}
	if target.writeCalls() != 2 { // one write created the standing authorization; one is this firing
		t.Fatalf("provider writes = %d, want standing approval plus one prepared firing", target.writeCalls())
	}
}

func TestAuthorizedFlowActionCancellationReleasesPreparedLeaseAndReplayWritesOnce(t *testing.T) {
	ctx := context.Background()
	a, target, req := newPreparedFlowActionFixture(t, ctx, "prepared-flow-cancel")
	prepared, err := a.PrepareAuthorizedFlowAction(ctx, req)
	if err != nil {
		t.Fatalf("PrepareAuthorizedFlowAction() error = %v", err)
	}
	beforeEvents := len(target.events())

	canceled, cancel := context.WithCancel(ctx)
	cancel()
	_, err = a.ExecutePreparedFlowAction(canceled, prepared)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ExecutePreparedFlowAction(cancelled) error = %T %v, want context.Canceled", err, err)
	}
	if len(target.events()) != beforeEvents {
		t.Fatalf("cancelled firing changed provider events: events=%d/%d", len(target.events()), beforeEvents)
	}

	if _, err := a.ExecutePreparedFlowAction(ctx, prepared); err != nil {
		t.Fatalf("ExecutePreparedFlowAction(after cancellation) error = %v; prepared execution stayed parked", err)
	}
	writesAfterSuccess := target.writeCalls()
	_, err = a.ExecutePreparedFlowAction(ctx, prepared)
	var replay *app.PreparedExecutionReplayError
	if !errors.As(err, &replay) {
		t.Fatalf("ExecutePreparedFlowAction(replay) error = %T %v, want PreparedExecutionReplayError", err, err)
	}
	if target.writeCalls() != writesAfterSuccess || len(a.ListFlowActionReceipts()) != 1 {
		t.Fatalf("prepared replay changed provider/receipt state: writes=%d receipts=%d", target.writeCalls(), len(a.ListFlowActionReceipts()))
	}
}

func TestAuthorizedFlowActionRevokedAuthorizationRefusesAndReleasesPreparedLease(t *testing.T) {
	ctx := context.Background()
	a, target, req := newPreparedFlowActionFixture(t, ctx, "prepared-flow-revoked")
	prepared, err := a.PrepareAuthorizedFlowAction(ctx, req)
	if err != nil {
		t.Fatalf("PrepareAuthorizedFlowAction() error = %v", err)
	}
	beforeWrites := target.writeCalls()
	if err := a.RevokeAuthorization(req.AuthorizationReference); err != nil {
		t.Fatalf("RevokeAuthorization() error = %v", err)
	}

	_, err = a.ExecutePreparedFlowAction(ctx, prepared)
	var revoked *app.AuthorizationRevokedError
	if !errors.As(err, &revoked) {
		t.Fatalf("ExecutePreparedFlowAction(revoked) error = %T %v, want AuthorizationRevokedError", err, err)
	}
	if target.writeCalls() != beforeWrites || len(a.ListFlowActionReceipts()) != 0 {
		t.Fatalf("revoked authorization changed state: writes=%d receipts=%d", target.writeCalls()-beforeWrites, len(a.ListFlowActionReceipts()))
	}
}

func TestAuthorizedFlowActionTamperedPreparedIdentityRefusesBeforeDispatch(t *testing.T) {
	ctx := context.Background()
	a, target, req := newPreparedFlowActionFixture(t, ctx, "prepared-flow-tampered")
	prepared, err := a.PrepareAuthorizedFlowAction(ctx, req)
	if err != nil {
		t.Fatalf("PrepareAuthorizedFlowAction() error = %v", err)
	}
	prepared.Identity = "pex_tampered"
	beforeWrites := target.writeCalls()

	_, err = a.ExecutePreparedFlowAction(ctx, prepared)
	var refused *app.PreparedExecutionRefusedError
	if !errors.As(err, &refused) {
		t.Fatalf("ExecutePreparedFlowAction(tampered) error = %T %v, want PreparedExecutionRefusedError", err, err)
	}
	if target.writeCalls() != beforeWrites || len(a.ListFlowActionReceipts()) != 0 {
		t.Fatalf("tampered prepared execution changed state: writes=%d receipts=%d", target.writeCalls()-beforeWrites, len(a.ListFlowActionReceipts()))
	}
}

func TestAuthorizedFlowActionConcurrentPreparedExecutionHasOneWinner(t *testing.T) {
	ctx := context.Background()
	a, target, req := newPreparedFlowActionFixture(t, ctx, "prepared-flow-race")
	prepared, err := a.PrepareAuthorizedFlowAction(ctx, req)
	if err != nil {
		t.Fatalf("PrepareAuthorizedFlowAction() error = %v", err)
	}
	beforeWrites := target.writeCalls()

	start := make(chan struct{})
	errs := make(chan error, 2)
	for range 2 {
		go func() {
			<-start
			_, executeErr := a.ExecutePreparedFlowAction(ctx, prepared)
			errs <- executeErr
		}()
	}
	close(start)
	first, second := <-errs, <-errs
	winners, replayCount := 0, 0
	for _, executeErr := range []error{first, second} {
		if executeErr == nil {
			winners++
			continue
		}
		var replay *app.PreparedExecutionReplayError
		if errors.As(executeErr, &replay) {
			replayCount++
			continue
		}
		t.Fatalf("concurrent execution error = %T %v, want nil or prepared replay", executeErr, executeErr)
	}
	if winners != 1 || replayCount != 1 || target.writeCalls() != beforeWrites+1 || len(a.ListFlowActionReceipts()) != 1 {
		t.Fatalf("concurrent prepared result winners=%d replay=%d writes=%d receipts=%d", winners, replayCount, target.writeCalls()-beforeWrites, len(a.ListFlowActionReceipts()))
	}
}

func TestAuthorizedFlowActionWriteFailureParksPreparedExecutionAndPersistsNoReceipt(t *testing.T) {
	ctx := context.Background()
	a, target, req := newPreparedFlowActionFixture(t, ctx, "prepared-flow-partial")
	prepared, err := a.PrepareAuthorizedFlowAction(ctx, req)
	if err != nil {
		t.Fatalf("PrepareAuthorizedFlowAction() error = %v", err)
	}
	target.setWriteError(errors.New("write failed after possible dispatch"))
	beforeWrites := target.writeCalls()
	_, err = a.ExecutePreparedFlowAction(ctx, prepared)
	if err == nil || err.Error() != "write failed after possible dispatch" {
		t.Fatalf("ExecutePreparedFlowAction(write failure) error = %v", err)
	}
	if target.writeCalls() != beforeWrites || len(a.ListFlowActionReceipts()) != 0 {
		t.Fatalf("failed write recorded success: writes=%d receipts=%d", target.writeCalls()-beforeWrites, len(a.ListFlowActionReceipts()))
	}

	reopened, err := app.Open(filepath.Dir(a.ProjectDir()))
	if err != nil {
		t.Fatalf("Open(replay project) error = %v", err)
	}
	target.setWriteError(nil)
	reopened.Registry().Register(target)
	_, err = reopened.ExecutePreparedFlowAction(ctx, prepared)
	var replay *app.PreparedExecutionReplayError
	if !errors.As(err, &replay) {
		t.Fatalf("ExecutePreparedFlowAction(reopen replay) error = %T %v, want PreparedExecutionReplayError", err, err)
	}
	if target.writeCalls() != beforeWrites || len(reopened.ListFlowActionReceipts()) != 0 {
		t.Fatalf("reopened replay changed provider/receipt state: writes=%d receipts=%d", target.writeCalls()-beforeWrites, len(reopened.ListFlowActionReceipts()))
	}
}

func newPreparedFlowActionFixture(t *testing.T, ctx context.Context, runID string) (*app.App, *flowActionTarget, app.FlowActionExecutionRequest) {
	t.Helper()
	a, sourcePlan := setupApprovedReversePlan(t, ctx)
	target := &flowActionTarget{}
	a.Registry().Register(target)
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{Name: "flow-target", Connector: target.Name(), Secrets: map[string]string{"token": "fixture-token"}}); err != nil {
		t.Fatalf("AddCredential(flow target) error = %v", err)
	}
	plan, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
		Name: "flow-action-target", SourceTable: sourcePlan.SourceTable, SourceConnection: sourcePlan.SourceConnection,
		DestinationConnector: target.Name(), DestinationCredential: "flow-target", Action: "create",
		Mappings: map[string]string{"id": "id", "email": "email"},
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
	return a, target, app.FlowActionExecutionRequest{
		FlowName: "prepared-flow", StepID: "create", RunID: runID, ManifestDigest: "fmd_fixture_prepared",
		SourceTable: sourcePlan.SourceTable, SourceConnection: sourcePlan.SourceConnection,
		DestinationTable: "flow-action-target", DestinationConnector: target.Name(), DestinationCredential: "flow-target",
		Action: "create", Mappings: map[string]string{"id": "id", "email": "email"},
		AuthorizationReference: a.ListAuthorizations()[0].Reference, ReadBackStream: "targets", Records: records,
	}
}

func cloneTestRecord(record connectors.Record) connectors.Record {
	clone := make(connectors.Record, len(record))
	for key, value := range record {
		clone[key] = value
	}
	return clone
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
	writeErr      error
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
	if c.writeErr != nil {
		return connectors.WriteResult{}, c.writeErr
	}
	if c.receiptCount != nil && c.receiptCount() != 0 {
		c.earlyReceipt = true
	}
	c.writes++
	c.records = append(c.records, records...)
	return connectors.WriteResult{RecordsWritten: len(records)}, nil
}

func (c *flowActionTarget) setWriteError(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.writeErr = err
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
