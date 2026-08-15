package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

// #4081 must not treat a typed writes.json action as self-approval. The
// pre-run App plan is closed over the connection's configured source predicate
// and fixed destination record; every apply below receives only a workset that
// was actually staged and independently reopened from the connection warehouse.
func TestIssueLabelDestinationRejectsUnapprovedOrMismatchedOrExpiredOrReplayedPlanBeforeProviderWrite(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T, issueLabelTransportApprovalFixture)
	}{
		{
			name: "missing pre-run approval",
			run: func(t *testing.T, fixture issueLabelTransportApprovalFixture) {
				receipt, workset := fixture.stageAndReopen(t, fixture.sourceIssue)
				if err := fixture.apply(t, receipt, workset, synctransport.DestinationApproval{}); err == nil {
					t.Fatal("ApplyDestination() accepted a missing pre-run approval")
				}
				fixture.assertProviderWrites(t, 0)
			},
		},
		{
			name: "same target and configuration with different reopened source issue",
			run: func(t *testing.T, fixture issueLabelTransportApprovalFixture) {
				_, approval := fixture.preRunApproval(t)
				receipt, workset := fixture.stageAndReopen(t, fixture.sourceIssue+1)
				if err := fixture.apply(t, receipt, workset, approval); err == nil {
					t.Fatal("ApplyDestination() accepted an approval for a different reopened source issue")
				}
				fixture.assertProviderWrites(t, 0)
			},
		},
		{
			name: "expired pre-run approval",
			run: func(t *testing.T, fixture issueLabelTransportApprovalFixture) {
				plan, approval := fixture.preRunApproval(t)
				fixture.expirePlanSeal(t, plan.ID)
				receipt, workset := fixture.stageAndReopen(t, fixture.sourceIssue)
				err := fixture.apply(t, receipt, workset, approval)
				if err == nil || !strings.Contains(err.Error(), "expired") {
					t.Fatalf("ApplyDestination() expired approval error = %v, want expired rejection", err)
				}
				fixture.assertProviderWrites(t, 0)
			},
		},
		{
			name: "replayed pre-run approval",
			run: func(t *testing.T, fixture issueLabelTransportApprovalFixture) {
				_, approval := fixture.preRunApproval(t)
				receipt, workset := fixture.stageAndReopen(t, fixture.sourceIssue)
				if err := fixture.apply(t, receipt, workset, approval); err != nil {
					t.Fatalf("first ApplyDestination() = %v", err)
				}
				fixture.assertProviderWrites(t, 1)
				err := fixture.apply(t, receipt, workset, approval)
				var replay *AuthorizationTokenReplayError
				if !errors.As(err, &replay) {
					t.Fatalf("ApplyDestination() replay error = %T %v, want AuthorizationTokenReplayError", err, err)
				}
				fixture.assertProviderWrites(t, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t, newIssueLabelTransportApprovalFixture(t))
		})
	}
}

func TestIssueLabelTransportNonAdditiveModesRequireExplicitConnectionConsent(t *testing.T) {
	tests := []struct {
		name       string
		mode       synccontract.Mode
		consentKey string
	}{
		{name: "set replace", mode: synccontract.ModeFullOverwrite, consentKey: "transport_allow_set_replace"},
		{name: "keyed", mode: synccontract.ModeIncrementalUpsert, consentKey: "transport_allow_keyed"},
	}

	for _, tt := range tests {
		t.Run(tt.name+" is disabled by default", func(t *testing.T) {
			fixture := newIssueLabelTransportApprovalFixture(t)
			fixture.configureMode(t, tt.mode, tt.consentKey, false)
			if _, err := fixture.app.PlanIssueLabelTransport(context.Background(), fixture.connection.ID); err == nil {
				t.Fatal("PlanIssueLabelTransport() accepted a non-additive mode without per-connection consent")
			}
			fixture.assertProviderWrites(t, 0)
			fixture.assertProviderSets(t, 0)
		})

		t.Run(tt.name+" persists a definition-owned replacement plan when enabled", func(t *testing.T) {
			fixture := newIssueLabelTransportApprovalFixture(t)
			fixture.configureMode(t, tt.mode, tt.consentKey, true)
			plan, err := fixture.app.PlanIssueLabelTransport(context.Background(), fixture.connection.ID)
			if err != nil {
				t.Fatalf("PlanIssueLabelTransport() = %v", err)
			}
			if plan.Action != "set_issue_labels" || plan.TransportBindingSHA256 == "" {
				t.Fatalf("non-additive plan = %+v, want the definition-owned set_issue_labels action and binding", plan)
			}
			fixture.assertProviderWrites(t, 0)
			fixture.assertProviderSets(t, 0)
		})
	}
}

func TestIssueLabelTransportNonAdditiveModesRequirePerConnectionAuthorizationBeforeProviderWrite(t *testing.T) {
	tests := []struct {
		name       string
		mode       synccontract.Mode
		consentKey string
	}{
		{name: "set replace", mode: synccontract.ModeFullOverwrite, consentKey: issueLabelTransportSetReplaceConsentConfig},
		{name: "keyed", mode: synccontract.ModeIncrementalUpsert, consentKey: issueLabelTransportKeyedConsentConfig},
	}

	for _, tt := range tests {
		t.Run(tt.name+" authorizes once, replays never, and allows identical scope unattended", func(t *testing.T) {
			fixture := newIssueLabelTransportApprovalFixture(t)
			fixture.configureMode(t, tt.mode, tt.consentKey, true)
			plan, approval := fixture.preRunApproval(t)
			before, err := fixture.app.AuthorizationScopeForReversePlan(context.Background(), plan.ID)
			if err != nil {
				t.Fatalf("AuthorizationScopeForReversePlan(before apply) = %v", err)
			}
			receipt, workset := fixture.stageAndReopenForMode(t, fixture.sourceIssue, tt.mode)
			if err := fixture.applyForMode(t, tt.mode, receipt, workset, approval); err != nil {
				t.Fatalf("first non-additive ApplyDestination() = %v", err)
			}
			fixture.assertProviderWrites(t, 0)
			fixture.assertProviderSets(t, 1)
			if err := fixture.readBackForMode(t, tt.mode, workset); err != nil {
				t.Fatalf("ReadBackDestination() after non-additive apply = %v", err)
			}

			stored, err := fixture.app.GetReversePlan(plan.ID)
			if err != nil {
				t.Fatalf("GetReversePlan() = %v", err)
			}
			if stored.AuthorizationReference == "" {
				t.Fatal("first non-additive apply did not persist a durable authorization reference")
			}
			after, err := fixture.app.AuthorizationScopeForReversePlan(context.Background(), plan.ID)
			if err != nil {
				t.Fatalf("AuthorizationScopeForReversePlan(after apply) = %v", err)
			}
			beforeID, err := AuthorizationScopeIdentity(before)
			if err != nil {
				t.Fatal(err)
			}
			afterID, err := AuthorizationScopeIdentity(after)
			if err != nil {
				t.Fatal(err)
			}
			if beforeID != afterID {
				t.Fatalf("durable authorization scope changed across an identical run: before=%q after=%q", beforeID, afterID)
			}

			err = fixture.applyForMode(t, tt.mode, receipt, workset, approval)
			var replay *AuthorizationTokenReplayError
			if !errors.As(err, &replay) {
				t.Fatalf("ApplyDestination() replay error = %T %v, want AuthorizationTokenReplayError", err, err)
			}
			fixture.assertProviderSets(t, 1)

			unattended := approval
			unattended.ApprovalToken = ""
			if err := fixture.applyForMode(t, tt.mode, receipt, workset, unattended); err != nil {
				t.Fatalf("identical-scope unattended ApplyDestination() = %v", err)
			}
			fixture.assertProviderSets(t, 2)
			if err := fixture.readBackForMode(t, tt.mode, workset); err != nil {
				t.Fatalf("ReadBackDestination() after unattended non-additive apply = %v", err)
			}

			if err := fixture.app.RevokeAuthorization(stored.AuthorizationReference); err != nil {
				t.Fatalf("RevokeAuthorization() = %v", err)
			}
			err = fixture.applyForMode(t, tt.mode, receipt, workset, unattended)
			var revoked *AuthorizationRevokedError
			if !errors.As(err, &revoked) {
				t.Fatalf("ApplyDestination() after authorization revoke = %v, want AuthorizationRevokedError", err)
			}
			fixture.assertProviderSets(t, 2)
		})

		t.Run(tt.name+" changed scope and disabled switch stop before PUT", func(t *testing.T) {
			fixture := newIssueLabelTransportApprovalFixture(t)
			fixture.configureMode(t, tt.mode, tt.consentKey, true)
			_, approval := fixture.preRunApproval(t)
			receipt, workset := fixture.stageAndReopenForMode(t, fixture.sourceIssue, tt.mode)
			if err := fixture.applyForMode(t, tt.mode, receipt, workset, approval); err != nil {
				t.Fatalf("first non-additive ApplyDestination() = %v", err)
			}
			fixture.assertProviderSets(t, 1)
			unattended := approval
			unattended.ApprovalToken = ""
			approvedRuntime := fixture.destinationRuntime(t)

			fixture.setDestinationConfig(t, issueLabelTransportLabelConfig, "changed-label")
			if err := fixture.applyForModeWithRuntime(t, tt.mode, receipt, workset, unattended, approvedRuntime); err == nil {
				t.Fatal("ApplyDestination() accepted a changed non-additive authorization scope")
			}
			fixture.assertProviderSets(t, 1)

			fixture.setDestinationConfig(t, issueLabelTransportLabelConfig, fixture.label)
			fixture.configureMode(t, tt.mode, tt.consentKey, false)
			if err := fixture.applyForModeWithRuntime(t, tt.mode, receipt, workset, unattended, approvedRuntime); err == nil {
				t.Fatal("ApplyDestination() accepted a non-additive mode after its per-connection switch was disabled")
			}
			fixture.assertProviderSets(t, 1)
			fixture.assertProviderWrites(t, 0)
		})
	}
}

func TestIssueLabelTransportCleanupUsesItsOwnDerivedApproval(t *testing.T) {
	fixture := newIssueLabelTransportApprovalFixture(t)
	forwardPlan, forwardApproval := fixture.preRunApproval(t)
	receipt, workset := fixture.stageAndReopen(t, fixture.sourceIssue)
	if err := fixture.apply(t, receipt, workset, forwardApproval); err != nil {
		t.Fatalf("forward ApplyDestination() = %v", err)
	}
	fixture.assertProviderWrites(t, 1)

	cleanupPlan, cleanupApproval := fixture.preRunCleanupApproval(t, forwardPlan.ID)
	if cleanupPlan.TransportForwardPlanID != forwardPlan.ID || cleanupPlan.Action != fixture.cleanupAction {
		t.Fatalf("cleanup plan = %+v, want inverse derived from forward plan %q", cleanupPlan, forwardPlan.ID)
	}
	if _, err := fixture.app.ApplyIssueLabelTransportCleanup(context.Background(), fixture.connection.ID, forwardApproval); err == nil {
		t.Fatal("cleanup accepted the forward approval token")
	}
	fixture.assertProviderDeletes(t, 0)
	if _, err := fixture.app.ApplyIssueLabelTransportCleanup(context.Background(), fixture.connection.ID, cleanupApproval); err != nil {
		t.Fatalf("ApplyIssueLabelTransportCleanup() = %v", err)
	}
	fixture.assertProviderDeletes(t, 1)
	if _, err := fixture.app.ApplyIssueLabelTransportCleanup(context.Background(), fixture.connection.ID, cleanupApproval); err == nil {
		t.Fatal("cleanup accepted a replayed cleanup approval")
	}
	fixture.assertProviderDeletes(t, 1)
}

func TestIssueLabelTransportCleanupRejectsUnauthenticatedForwardPlanBeforeProviderDelete(t *testing.T) {
	for _, tt := range issueLabelTransportForwardPlanTamperCases() {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newIssueLabelTransportApprovalFixture(t)
			forwardPlan, forwardApproval := fixture.preRunApproval(t)
			receipt, workset := fixture.stageAndReopen(t, fixture.sourceIssue)
			if err := fixture.apply(t, receipt, workset, forwardApproval); err != nil {
				t.Fatalf("forward ApplyDestination() = %v", err)
			}
			fixture.assertProviderWrites(t, 1)
			tt.mutate(t, fixture, forwardPlan.ID)
			if _, err := fixture.app.PlanIssueLabelTransportCleanup(context.Background(), fixture.connection.ID, forwardPlan.ID); err == nil {
				t.Fatal("PlanIssueLabelTransportCleanup() accepted a tampered forward plan")
			}
			fixture.assertProviderDeletes(t, 0)
		})
	}
}

func TestIssueLabelTransportCleanupApplyRejectsTamperedForwardPlanBeforeProviderDelete(t *testing.T) {
	for _, tt := range issueLabelTransportForwardPlanTamperCases() {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newIssueLabelTransportApprovalFixture(t)
			forwardPlan, forwardApproval := fixture.preRunApproval(t)
			receipt, workset := fixture.stageAndReopen(t, fixture.sourceIssue)
			if err := fixture.apply(t, receipt, workset, forwardApproval); err != nil {
				t.Fatalf("forward ApplyDestination() = %v", err)
			}
			_, cleanupApproval := fixture.preRunCleanupApproval(t, forwardPlan.ID)
			tt.mutate(t, fixture, forwardPlan.ID)
			if _, err := fixture.app.ApplyIssueLabelTransportCleanup(context.Background(), fixture.connection.ID, cleanupApproval); err == nil {
				t.Fatal("ApplyIssueLabelTransportCleanup() accepted a tampered forward plan")
			}
			fixture.assertProviderDeletes(t, 0)
		})
	}
}

func TestIssueLabelTransportCleanupTreatsMissingLabelAsSuccessfulInverse(t *testing.T) {
	fixture := newIssueLabelTransportApprovalFixtureWithMissingCleanupLabel(t)
	forwardPlan, forwardApproval := fixture.preRunApproval(t)
	receipt, workset := fixture.stageAndReopen(t, fixture.sourceIssue)
	if err := fixture.apply(t, receipt, workset, forwardApproval); err != nil {
		t.Fatalf("forward ApplyDestination() = %v", err)
	}
	cleanupPlan, cleanupApproval := fixture.preRunCleanupApproval(t, forwardPlan.ID)
	if _, err := fixture.app.ApplyIssueLabelTransportCleanup(context.Background(), fixture.connection.ID, cleanupApproval); err != nil {
		t.Fatalf("ApplyIssueLabelTransportCleanup() = %v, want missing-label success for %q", err, cleanupPlan.ID)
	}
	fixture.assertProviderDeletes(t, 1)
}

func TestIssueLabelTransportContractUsesDefinitionOwnedActionBindings(t *testing.T) {
	definition := connectors.Definition{
		Streams: []connectors.StreamSummary{{Name: "issues"}},
		WriteActions: []connectors.WriteActionInfo{
			{
				Name:   "attach_ticket_tag",
				Method: "POST",
				Path:   "/tickets/{{ record.ticket_id }}/tags",
				TransportBinding: &connectors.TransportActionBinding{
					Capability: connectors.TransportCapabilityIssueLabel,
					Role:       connectors.TransportActionRoleApply,
					Modes:      []synccontract.Mode{synccontract.ModeFullAppend},
					Inputs: []connectors.TransportInputBinding{
						{Input: connectors.TransportInputTargetIssue, Field: "ticket_id", Shape: connectors.TransportInputShapeScalar},
						{Input: connectors.TransportInputLabel, Field: "tag_values", Shape: connectors.TransportInputShapeList},
					},
				},
			},
			{
				Name:   "replace_ticket_tags",
				Method: "PUT",
				Path:   "/tickets/{{ record.ticket_id }}/tags",
				TransportBinding: &connectors.TransportActionBinding{
					Capability: connectors.TransportCapabilityIssueLabel,
					Role:       connectors.TransportActionRoleReplace,
					Modes:      []synccontract.Mode{synccontract.ModeFullOverwrite, synccontract.ModeIncrementalUpsert},
					Inputs: []connectors.TransportInputBinding{
						{Input: connectors.TransportInputTargetIssue, Field: "ticket_id", Shape: connectors.TransportInputShapeScalar},
						{Input: connectors.TransportInputLabel, Field: "tag_values", Shape: connectors.TransportInputShapeList},
					},
				},
			},
			{
				Name:   "detach_ticket_tag",
				Method: "DELETE",
				Path:   "/tickets/{{ record.ticket_id }}/tags/{{ record.tag }}",
				TransportBinding: &connectors.TransportActionBinding{
					Capability: connectors.TransportCapabilityIssueLabel,
					Role:       connectors.TransportActionRoleCleanup,
					Modes:      []synccontract.Mode{synccontract.ModeFullAppend},
					Inputs: []connectors.TransportInputBinding{
						{Input: connectors.TransportInputTargetIssue, Field: "ticket_id", Shape: connectors.TransportInputShapeScalar},
						{Input: connectors.TransportInputLabel, Field: "tag", Shape: connectors.TransportInputShapeScalar},
					},
				},
			},
		},
	}

	contract, err := issueLabelTransportContractForDefinition(definition)
	if err != nil {
		t.Fatalf("issueLabelTransportContractForDefinition() = %v", err)
	}
	if contract.apply.name != "attach_ticket_tag" || contract.replace.name != "replace_ticket_tags" || contract.cleanup.name != "detach_ticket_tag" {
		t.Fatalf("contract actions = apply %q replace %q cleanup %q", contract.apply.name, contract.replace.name, contract.cleanup.name)
	}
	applyRecord, err := contract.apply.record(200, "transport-demo")
	if err != nil {
		t.Fatalf("apply record = %v", err)
	}
	if applyRecord["ticket_id"] != 200 {
		t.Fatalf("apply record ticket_id = %#v, want 200", applyRecord["ticket_id"])
	}
	replaceAction, err := contract.actionForSyncMode(synccontract.ModeFullOverwrite)
	if err != nil {
		t.Fatalf("replacement action = %v", err)
	}
	if replaceAction.name != "replace_ticket_tags" {
		t.Fatalf("replacement action = %q, want definition-owned replacement", replaceAction.name)
	}
	if got, ok := applyRecord["tag_values"].([]string); !ok || len(got) != 1 || got[0] != "transport-demo" {
		t.Fatalf("apply record tag_values = %#v, want singleton transport-demo", applyRecord["tag_values"])
	}
	cleanupRecord, err := contract.cleanup.record(200, "transport-demo")
	if err != nil {
		t.Fatalf("cleanup record = %v", err)
	}
	if cleanupRecord["ticket_id"] != 200 || cleanupRecord["tag"] != "transport-demo" {
		t.Fatalf("cleanup record = %#v, want definition-owned scalar fields", cleanupRecord)
	}
}

func TestIssueLabelTransportSourceStopsAfterMatchedFullPage(t *testing.T) {
	fixture := newIssueLabelTransportApprovalFixture(t)
	resume := streamResumeExpectation(fixture.sourceConnector, fixture.sourceCredential, fixture.sourceRuntime, "issues")
	var pages []synctransport.SourcePage
	err := fixture.sourceExecutor.ReadTransport(context.Background(), synctransport.SourceRequest{
		Connector: fixture.sourceConnector,
		Runtime:   fixture.sourceRuntime,
		Stream:    "issues",
		Mode:      synccontract.ModeFullAppend,
		BatchSize: 1,
		Resume:    resume,
	}, func(page synctransport.SourcePage) error {
		pages = append(pages, page)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadTransport() = %v", err)
	}
	fixture.assertProviderReads(t, 1)
	if len(pages) != 1 || len(pages[0].Records) != 1 {
		t.Fatalf("source pages = %#v, want one singleton page", pages)
	}
	if number, err := issueNumberFromRecord(pages[0].Records[0]); err != nil || number != fixture.sourceIssue {
		t.Fatalf("source record number = %d, %v; want %d", number, err, fixture.sourceIssue)
	}
}

func TestIssueLabelTransportSourceCollectsBoundedBatchForManagedTarget(t *testing.T) {
	fixture := newIssueLabelTransportApprovalFixture(t)
	fixture.sourceRuntime.Config = cloneStringMap(fixture.sourceRuntime.Config)
	delete(fixture.sourceRuntime.Config, issueLabelTransportSourceIssueConfig)
	resume := streamResumeExpectation(fixture.sourceConnector, fixture.sourceCredential, fixture.sourceRuntime, "issues")
	var pages []synctransport.SourcePage
	err := fixture.sourceExecutor.ReadTransport(context.Background(), synctransport.SourceRequest{
		Connector: fixture.sourceConnector,
		Runtime:   fixture.sourceRuntime,
		Stream:    "issues",
		Mode:      synccontract.ModeIncrementalUpsert,
		BatchSize: 3,
		Resume:    resume,
	}, func(page synctransport.SourcePage) error {
		pages = append(pages, page)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadTransport() = %v", err)
	}
	fixture.assertProviderReads(t, 1)
	if len(pages) != 1 || len(pages[0].Records) != 3 {
		t.Fatalf("collection source pages = %#v, want one three-record bounded page", pages)
	}
	for index, want := range []int{fixture.sourceIssue, fixture.targetIssue, 1} {
		if number, err := issueNumberFromRecord(pages[0].Records[index]); err != nil || number != want {
			t.Fatalf("collection source record %d number = %d, %v; want %d", index, number, err, want)
		}
	}
}

func TestIssueLabelTransportReadBackStopsAfterMatchedFullPage(t *testing.T) {
	fixture := newIssueLabelTransportApprovalFixture(t)
	err := fixture.executor.ReadBackDestination(context.Background(), synctransport.DestinationReadBackRequest{
		Plan: synctransport.DestinationPlan{ApplyStrategy: connectors.DestinationApplyStrategy{
			Mode:     synccontract.ModeFullAppend,
			Strategy: connectors.ApplyStrategyAppend,
			Action:   fixture.applyAction,
		}},
		Workset: synctransport.WarehouseWorkset{ID: "reopened-workset"},
		Runtime: fixture.runtime,
	})
	if err != nil {
		t.Fatalf("ReadBackDestination() = %v", err)
	}
	fixture.assertProviderReads(t, 1)
}

func TestIssueLabelTransportSourceDoesNotReadBeyondFirstPageWhenIssueMissing(t *testing.T) {
	fixture := newIssueLabelTransportApprovalFixtureWithFirstPageMissingIssues(t)
	resume := streamResumeExpectation(fixture.sourceConnector, fixture.sourceCredential, fixture.sourceRuntime, "issues")
	emitted := false
	err := fixture.sourceExecutor.ReadTransport(context.Background(), synctransport.SourceRequest{
		Connector: fixture.sourceConnector,
		Runtime:   fixture.sourceRuntime,
		Stream:    "issues",
		Mode:      synccontract.ModeFullAppend,
		BatchSize: 1,
		Resume:    resume,
	}, func(synctransport.SourcePage) error {
		emitted = true
		return nil
	})
	if err == nil {
		t.Error("ReadTransport() succeeded after the configured issue was absent from the first page")
	}
	if emitted {
		t.Error("ReadTransport() emitted a source page after the configured issue was absent from the first page")
	}
	fixture.assertProviderReads(t, 1)
}

func TestIssueLabelTransportReadBackDoesNotReadBeyondFirstPageWhenIssueMissing(t *testing.T) {
	fixture := newIssueLabelTransportApprovalFixtureWithFirstPageMissingIssues(t)
	err := fixture.executor.ReadBackDestination(context.Background(), synctransport.DestinationReadBackRequest{
		Plan: synctransport.DestinationPlan{ApplyStrategy: connectors.DestinationApplyStrategy{
			Mode:     synccontract.ModeFullAppend,
			Strategy: connectors.ApplyStrategyAppend,
			Action:   fixture.applyAction,
		}},
		Workset: synctransport.WarehouseWorkset{ID: "reopened-workset"},
		Runtime: fixture.runtime,
	})
	if err == nil {
		t.Error("ReadBackDestination() succeeded after the configured issue was absent from the first page")
	}
	fixture.assertProviderReads(t, 1)
}

type issueLabelTransportApprovalFixture struct {
	app              *App
	connection       Connection
	executor         *issueLabelDestinationExecutor
	sourceExecutor   *issueLabelSourceExecutor
	sourceConnector  connectors.Connector
	sourceCredential CredentialMeta
	sourceRuntime    connectors.RuntimeConfig
	runtime          connectors.RuntimeConfig
	sourceIssue      int
	targetIssue      int
	label            string
	applyAction      string
	replaceAction    string
	cleanupAction    string
	reads            *int
	writes           *int
	sets             *int
	deletes          *int
}

func newIssueLabelTransportApprovalFixture(t *testing.T) issueLabelTransportApprovalFixture {
	return newIssueLabelTransportApprovalFixtureWithCleanupStatus(t, http.StatusNoContent)
}

func newIssueLabelTransportApprovalFixtureWithMissingCleanupLabel(t *testing.T) issueLabelTransportApprovalFixture {
	return newIssueLabelTransportApprovalFixtureWithCleanupStatus(t, http.StatusNotFound)
}

func newIssueLabelTransportApprovalFixtureWithCleanupStatus(t *testing.T, cleanupStatus int) issueLabelTransportApprovalFixture {
	return newIssueLabelTransportApprovalFixtureWithIssuePages(t, cleanupStatus, func(page string) []map[string]any {
		if page != "" && page != "1" {
			return []map[string]any{}
		}
		return issueLabelTransportFullPage(100, 200, "transport-demo")
	})
}

func newIssueLabelTransportApprovalFixtureWithFirstPageMissingIssues(t *testing.T) issueLabelTransportApprovalFixture {
	return newIssueLabelTransportApprovalFixtureWithIssuePages(t, http.StatusNoContent, func(page string) []map[string]any {
		if page == "2" {
			return issueLabelTransportFullPage(100, 200, "transport-demo")
		}
		return issueLabelTransportFullPage(101, 102, "")
	})
}

func newIssueLabelTransportApprovalFixtureWithIssuePages(t *testing.T, cleanupStatus int, issuePage func(string) []map[string]any) issueLabelTransportApprovalFixture {
	t.Helper()
	ctx := context.Background()
	reads := 0
	writes := 0
	sets := 0
	deletes := 0
	targetLabels := []string{"transport-demo", "legacy"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			reads++
			if request.URL.Path != "/repos/acme/widgets/issues" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			records := issuePage(request.URL.Query().Get("page"))
			for _, record := range records {
				if number, ok := record["number"].(int); ok && number == 200 {
					record["labels"] = issueLabelTransportLabelResponse(targetLabels)
				}
			}
			writeIssueLabelTransportJSON(t, w, records)
		case http.MethodPost:
			if request.URL.Path != "/repos/acme/widgets/issues/200/labels" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			var body map[string]any
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			writes++
			for _, label := range issueLabelTransportRequestLabels(body) {
				if !issueLabelTransportContainsLabel(targetLabels, label) {
					targetLabels = append(targetLabels, label)
				}
			}
			writeIssueLabelTransportJSON(t, w, issueLabelTransportLabelResponse(targetLabels))
		case http.MethodPut:
			if request.URL.Path != "/repos/acme/widgets/issues/200/labels" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			var body map[string]any
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			sets++
			targetLabels = issueLabelTransportRequestLabels(body)
			writeIssueLabelTransportJSON(t, w, issueLabelTransportLabelResponse(targetLabels))
		case http.MethodDelete:
			if request.URL.Path != "/repos/acme/widgets/issues/200/labels/transport-demo" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			deletes++
			w.WriteHeader(cleanupStatus)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.AddCredential(ctx, AddCredentialRequest{
		Name:      "github-transport-local",
		Connector: "github",
		Config: map[string]string{
			"owner":         "acme",
			"repo":          "widgets",
			"public_access": "true",
			"base_url":      server.URL,
		},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.AddCredential(ctx, AddCredentialRequest{
		Name:      "other-github-credential",
		Connector: "github",
		Config: map[string]string{
			"owner":         "acme",
			"repo":          "widgets",
			"public_access": "true",
			"base_url":      server.URL,
		},
	}); err != nil {
		t.Fatal(err)
	}
	connection, err := a.CreateConnection(ctx, CreateConnectionRequest{
		Name: "github_transport_approval",
		Source: EndpointConfig{Connector: "github", Credential: "github-transport-local", Config: map[string]string{
			issueLabelTransportSourceIssueConfig: "100",
		}},
		Destination: EndpointConfig{Connector: "github", Credential: "github-transport-local", Config: map[string]string{
			issueLabelTransportTargetIssueConfig: "200",
			issueLabelTransportLabelConfig:       "transport-demo",
		}},
		Streams: map[string]StreamConfig{
			"issues": {SyncMode: string(synccontract.ModeFullAppend), DestinationTable: "issues"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	sourceConnector, sourceCredential, sourceRuntime, err := a.resolveEndpointWithCredential(ctx, connection.Source)
	if err != nil {
		t.Fatal(err)
	}
	_, runtime, err := a.resolveEndpoint(ctx, connection.Destination)
	if err != nil {
		t.Fatal(err)
	}
	registered, ok := a.registry.Get("github")
	if !ok {
		t.Fatal("GitHub connector is not registered")
	}
	github, contract, err := issueLabelTransportConnectorContract(registered)
	if err != nil || github == nil {
		t.Fatalf("GitHub transport connector = %T, want declarative issue-label connector: %v", registered, err)
	}
	return issueLabelTransportApprovalFixture{
		app:              a,
		connection:       connection,
		executor:         &issueLabelDestinationExecutor{app: a, connector: github, contract: contract},
		sourceExecutor:   &issueLabelSourceExecutor{connector: github, contract: contract},
		sourceConnector:  sourceConnector,
		sourceCredential: sourceCredential,
		sourceRuntime:    sourceRuntime,
		runtime:          runtime,
		sourceIssue:      100,
		targetIssue:      200,
		label:            "transport-demo",
		applyAction:      contract.apply.name,
		replaceAction:    contract.replace.name,
		cleanupAction:    contract.cleanup.name,
		reads:            &reads,
		writes:           &writes,
		sets:             &sets,
		deletes:          &deletes,
	}
}

func writeIssueLabelTransportJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Errorf("encode GitHub transport fixture response: %v", err)
	}
}

func issueLabelTransportRequestLabels(body map[string]any) []string {
	values, ok := body["labels"].([]any)
	if !ok {
		return nil
	}
	labels := make([]string, 0, len(values))
	for _, value := range values {
		label, ok := value.(string)
		if ok && strings.TrimSpace(label) != "" {
			labels = append(labels, label)
		}
	}
	return labels
}

func issueLabelTransportLabelResponse(labels []string) []map[string]any {
	response := make([]map[string]any, 0, len(labels))
	for _, label := range labels {
		response = append(response, map[string]any{"name": label})
	}
	return response
}

func issueLabelTransportContainsLabel(labels []string, want string) bool {
	for _, label := range labels {
		if label == want {
			return true
		}
	}
	return false
}

func (f issueLabelTransportApprovalFixture) preRunApproval(t *testing.T) (ReversePlan, synctransport.DestinationApproval) {
	t.Helper()
	plan, err := f.app.PlanIssueLabelTransport(context.Background(), f.connection.ID)
	if err != nil {
		t.Fatalf("PlanIssueLabelTransport() = %v", err)
	}
	plan, preview, err := f.app.PreviewIssueLabelTransport(context.Background(), plan.ID)
	if err != nil {
		t.Fatalf("PreviewIssueLabelTransport() = %v", err)
	}
	if preview.Digest == "" || plan.ApprovalToken == "" || plan.ConfirmationPolicy.Kind != connectors.ConfirmationKindDestructive {
		t.Fatalf("pre-run GitHub transport approval = plan=%+v preview=%+v, want persisted destructive grant", plan, preview)
	}
	return plan, synctransport.DestinationApproval{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	}
}

func (f issueLabelTransportApprovalFixture) preRunCleanupApproval(t *testing.T, forwardPlanID string) (ReversePlan, synctransport.DestinationApproval) {
	t.Helper()
	plan, err := f.app.PlanIssueLabelTransportCleanup(context.Background(), f.connection.ID, forwardPlanID)
	if err != nil {
		t.Fatalf("PlanIssueLabelTransportCleanup() = %v", err)
	}
	plan, preview, err := f.app.PreviewIssueLabelTransport(context.Background(), plan.ID)
	if err != nil {
		t.Fatalf("PreviewIssueLabelTransport(cleanup) = %v", err)
	}
	if preview.Digest == "" || plan.ApprovalToken == "" || plan.ConfirmationPolicy.Kind != connectors.ConfirmationKindDestructive {
		t.Fatalf("pre-run GitHub transport cleanup approval = plan=%+v preview=%+v, want persisted destructive grant", plan, preview)
	}
	return plan, synctransport.DestinationApproval{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	}
}

func (f issueLabelTransportApprovalFixture) stageAndReopen(t *testing.T, sourceIssue int) (synctransport.WarehouseReceipt, synctransport.WarehouseWorkset) {
	return f.stageAndReopenForMode(t, sourceIssue, synccontract.ModeFullAppend)
}

func (f issueLabelTransportApprovalFixture) stageAndReopenForMode(t *testing.T, sourceIssue int, mode synccontract.Mode) (synctransport.WarehouseReceipt, synctransport.WarehouseWorkset) {
	t.Helper()
	page := synctransport.SourcePage{Records: []connectors.Record{{
		"id":     "source-issue",
		"number": sourceIssue,
		"title":  "warehouse-owned transport source",
	}}}
	receipt, err := f.app.transportStage.Stage(context.Background(), synctransport.WarehouseStageRequest{
		ConnectionID:    f.connection.ID,
		Generation:      1,
		SourceName:      "github",
		DestinationName: "github",
		Stream:          "issues",
		Mode:            mode,
		Page:            page,
	})
	if err != nil {
		t.Fatalf("Stage() = %v", err)
	}
	// Destroy the source-owned alias before Reopen; the apply receives only the
	// persisted WAL/DuckDB/Parquet representation and immutable receipt.
	page.Records[0]["number"] = -1
	page.Records = nil
	page = synctransport.SourcePage{}
	workset, err := f.app.transportStage.Reopen(context.Background(), receipt)
	if err != nil {
		t.Fatalf("Reopen() = %v", err)
	}
	return receipt, workset
}

func (f issueLabelTransportApprovalFixture) expirePlanSeal(t *testing.T, planID string) {
	t.Helper()
	found := false
	for i := range f.app.state.ReversePlans {
		plan := &f.app.state.ReversePlans[i]
		if plan.ID != planID {
			continue
		}
		if plan.PlanSeal == nil {
			t.Fatalf("plan %q has no seal", planID)
		}
		seal := *plan.PlanSeal
		seal.ExpiresAt = time.Now().UTC().Add(-time.Minute)
		var err error
		seal.MAC, err = f.app.approval.planSealMAC(seal)
		if err != nil {
			t.Fatal(err)
		}
		plan.PlanSeal = &seal
		plan.ExpiresAt = seal.ExpiresAt
		found = true
		break
	}
	if !found {
		t.Fatalf("plan %q was not stored", planID)
	}
	if err := f.app.save(); err != nil {
		t.Fatal(err)
	}
}

func (f issueLabelTransportApprovalFixture) mutateForwardPlan(t *testing.T, planID string, mutate func(*ReversePlan)) {
	t.Helper()
	found := false
	for i := range f.app.state.ReversePlans {
		if f.app.state.ReversePlans[i].ID != planID {
			continue
		}
		mutate(&f.app.state.ReversePlans[i])
		found = true
		break
	}
	if !found {
		t.Fatalf("forward plan %q was not stored", planID)
	}
	if err := f.app.save(); err != nil {
		t.Fatal(err)
	}
}

func (f issueLabelTransportApprovalFixture) apply(t *testing.T, receipt synctransport.WarehouseReceipt, workset synctransport.WarehouseWorkset, approval synctransport.DestinationApproval) error {
	return f.applyForMode(t, synccontract.ModeFullAppend, receipt, workset, approval)
}

func (f issueLabelTransportApprovalFixture) applyForMode(t *testing.T, mode synccontract.Mode, receipt synctransport.WarehouseReceipt, workset synctransport.WarehouseWorkset, approval synctransport.DestinationApproval) error {
	return f.applyForModeWithRuntime(t, mode, receipt, workset, approval, f.destinationRuntime(t))
}

func (f issueLabelTransportApprovalFixture) applyForModeWithRuntime(t *testing.T, mode synccontract.Mode, receipt synctransport.WarehouseReceipt, workset synctransport.WarehouseWorkset, approval synctransport.DestinationApproval, runtime connectors.RuntimeConfig) error {
	t.Helper()
	action, err := f.executor.contract.actionForSyncMode(mode)
	if err != nil {
		t.Fatalf("definition action for mode %q = %v", mode, err)
	}
	strategy, err := issueLabelTransportApplyStrategy(mode)
	if err != nil {
		t.Fatalf("definition strategy for mode %q = %v", mode, err)
	}
	acknowledgement, err := f.executor.ApplyDestination(context.Background(), synctransport.DestinationApplyRequest{
		ConnectionID: f.connection.ID,
		Plan: synctransport.DestinationPlan{ApplyStrategy: connectors.DestinationApplyStrategy{
			Mode:     mode,
			Strategy: strategy,
			Action:   action.name,
		}},
		Receipt:  receipt,
		Workset:  workset,
		Runtime:  runtime,
		Approval: approval,
	})
	if err == nil && acknowledgement.Sink != "github" {
		t.Fatalf("ApplyDestination() acknowledgement sink = %q, want github", acknowledgement.Sink)
	}
	return err
}

func (f issueLabelTransportApprovalFixture) readBackForMode(t *testing.T, mode synccontract.Mode, workset synctransport.WarehouseWorkset) error {
	t.Helper()
	action, err := f.executor.contract.actionForSyncMode(mode)
	if err != nil {
		t.Fatalf("definition read-back action for mode %q = %v", mode, err)
	}
	strategy, err := issueLabelTransportApplyStrategy(mode)
	if err != nil {
		t.Fatalf("definition read-back strategy for mode %q = %v", mode, err)
	}
	return f.executor.ReadBackDestination(context.Background(), synctransport.DestinationReadBackRequest{
		Plan: synctransport.DestinationPlan{ApplyStrategy: connectors.DestinationApplyStrategy{
			Mode: mode, Strategy: strategy, Action: action.name,
		}},
		Workset: workset,
		Runtime: f.destinationRuntime(t),
	})
}

func (f issueLabelTransportApprovalFixture) destinationRuntime(t *testing.T) connectors.RuntimeConfig {
	t.Helper()
	conn, err := f.app.issueLabelTransportConnection(f.connection.ID)
	if err != nil {
		t.Fatalf("issueLabelTransportConnection() = %v", err)
	}
	_, runtime, err := f.app.resolveEndpoint(context.Background(), conn.Destination)
	if err != nil {
		t.Fatalf("resolve destination runtime = %v", err)
	}
	return runtime
}

func (f issueLabelTransportApprovalFixture) assertProviderWrites(t *testing.T, want int) {
	t.Helper()
	if got := *f.writes; got != want {
		t.Fatalf("GitHub label POST calls = %d, want %d", got, want)
	}
}

func (f issueLabelTransportApprovalFixture) assertProviderSets(t *testing.T, want int) {
	t.Helper()
	if got := *f.sets; got != want {
		t.Fatalf("GitHub label PUT calls = %d, want %d", got, want)
	}
}

func (f issueLabelTransportApprovalFixture) configureMode(t *testing.T, mode synccontract.Mode, consentKey string, enabled bool) {
	t.Helper()
	updated := false
	for index := range f.app.state.Connections {
		connection := &f.app.state.Connections[index]
		if connection.ID != f.connection.ID {
			continue
		}
		connection.Streams = cloneStreamConfigs(connection.Streams)
		stream := connection.Streams["issues"]
		stream.SyncMode = string(mode)
		connection.Streams["issues"] = stream
		connection.Destination.Config = cloneStringMap(connection.Destination.Config)
		if enabled {
			connection.Destination.Config[consentKey] = "true"
		} else {
			delete(connection.Destination.Config, consentKey)
		}
		updated = true
		break
	}
	if !updated {
		t.Fatalf("connection %q was not stored", f.connection.ID)
	}
	if err := f.app.save(); err != nil {
		t.Fatal(err)
	}
}

func (f issueLabelTransportApprovalFixture) setDestinationConfig(t *testing.T, key, value string) {
	t.Helper()
	updated := false
	for index := range f.app.state.Connections {
		connection := &f.app.state.Connections[index]
		if connection.ID != f.connection.ID {
			continue
		}
		connection.Destination.Config = cloneStringMap(connection.Destination.Config)
		connection.Destination.Config[key] = value
		updated = true
		break
	}
	if !updated {
		t.Fatalf("connection %q was not stored", f.connection.ID)
	}
	if err := f.app.save(); err != nil {
		t.Fatal(err)
	}
}

func (f issueLabelTransportApprovalFixture) assertProviderReads(t *testing.T, want int) {
	t.Helper()
	if got := *f.reads; got != want {
		t.Fatalf("GitHub issue GET calls = %d, want %d", got, want)
	}
}

func (f issueLabelTransportApprovalFixture) assertProviderDeletes(t *testing.T, want int) {
	t.Helper()
	if got := *f.deletes; got != want {
		t.Fatalf("GitHub label DELETE calls = %d, want %d", got, want)
	}
}

func issueLabelTransportFullPage(sourceIssue, targetIssue int, label string) []map[string]any {
	records := []map[string]any{
		issueLabelTransportTestIssue(sourceIssue, ""),
		issueLabelTransportTestIssue(targetIssue, label),
	}
	for number := 1; len(records) < 100; number++ {
		if number == sourceIssue || number == targetIssue {
			continue
		}
		records = append(records, issueLabelTransportTestIssue(number, ""))
	}
	return records
}

func issueLabelTransportTestIssue(number int, label string) map[string]any {
	labels := []map[string]any{}
	if label != "" {
		labels = append(labels, map[string]any{"name": label})
	}
	return map[string]any{
		"id":      number,
		"node_id": fmt.Sprintf("I_%d", number),
		"number":  number,
		"title":   "transport test issue",
		"state":   "open",
		"labels":  labels,
	}
}

type issueLabelTransportForwardPlanTamperCase struct {
	name   string
	mutate func(*testing.T, issueLabelTransportApprovalFixture, string)
}

func issueLabelTransportForwardPlanTamperCases() []issueLabelTransportForwardPlanTamperCase {
	return []issueLabelTransportForwardPlanTamperCase{
		{
			name: "destination connector",
			mutate: func(t *testing.T, fixture issueLabelTransportApprovalFixture, planID string) {
				fixture.mutateForwardPlan(t, planID, func(plan *ReversePlan) { plan.DestinationConnector = "warehouse" })
			},
		},
		{
			name: "destination credential",
			mutate: func(t *testing.T, fixture issueLabelTransportApprovalFixture, planID string) {
				fixture.mutateForwardPlan(t, planID, func(plan *ReversePlan) { plan.DestinationCredential = "other-github-credential" })
			},
		},
		{
			name: "destination config",
			mutate: func(t *testing.T, fixture issueLabelTransportApprovalFixture, planID string) {
				fixture.mutateForwardPlan(t, planID, func(plan *ReversePlan) {
					plan.DestinationConfig = cloneStringMap(plan.DestinationConfig)
					plan.DestinationConfig[issueLabelTransportTargetIssueConfig] = "201"
				})
			},
		},
		{
			name: "binding",
			mutate: func(t *testing.T, fixture issueLabelTransportApprovalFixture, planID string) {
				fixture.mutateForwardPlan(t, planID, func(plan *ReversePlan) { plan.TransportBindingSHA256 = "tampered-binding" })
			},
		},
		{
			name: "plan hash",
			mutate: func(t *testing.T, fixture issueLabelTransportApprovalFixture, planID string) {
				fixture.mutateForwardPlan(t, planID, func(plan *ReversePlan) { plan.PlanHash = "tampered-plan-hash" })
			},
		},
		{
			name: "plan seal",
			mutate: func(t *testing.T, fixture issueLabelTransportApprovalFixture, planID string) {
				fixture.mutateForwardPlan(t, planID, func(plan *ReversePlan) {
					if plan.PlanSeal == nil {
						t.Fatal("forward plan has no seal")
					}
					plan.PlanSeal.MAC = "tampered-seal"
				})
			},
		},
	}
}
