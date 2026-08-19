package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/flow"
	"polymetrics.ai/internal/schedule"
)

func TestInstalledScheduleFireRunsAuthorizedRoundTripAndRestoresBackend(t *testing.T) {
	ctx, root, a, target, manifest, plan := newAuthorizedScheduleFixture(t)
	backendFile := filepath.Join(t.TempDir(), "crontab")
	baseline := []byte("# preserved fixture backend\n")
	require.NoError(t, os.WriteFile(backendFile, baseline, 0o600))
	backend := schedule.CrontabBackend{File: backendFile}
	require.NoError(t, backend.Install(ctx, manifest, "/fixture/pm"))

	installed, err := os.ReadFile(backendFile)
	require.NoError(t, err)
	if !strings.Contains(string(installed), "flow run "+manifest.Flow+" --json") || strings.Contains(string(installed), "authorization") {
		t.Fatalf("installed payload = %q, want direct existing-flow run with no authorization carrier", installed)
	}

	beforeWrites := target.writeCalls()
	var stdout bytes.Buffer
	err = flowRun(ctx, config.Config{}, a, []string{manifest.Flow}, &stdout, true)
	require.NoError(t, err)
	if target.writeCalls() != beforeWrites+1 {
		t.Fatalf("scheduled fire writes = %d, want %d", target.writeCalls(), beforeWrites+1)
	}
	if got := target.tailEvents(4); !equalStrings(got, []string{"validate", "preview", "write", "read"}) {
		t.Fatalf("scheduled target events = %v, want typed validation/write/read-back", got)
	}
	state, err := schedule.LoadFireState(root, manifest.Name)
	require.NoError(t, err)
	if state.Status != schedule.FireStatusSucceeded || state.LastFire.FlowStatus != "ok" || len(state.LastFire.ReceiptIDs) != 1 {
		t.Fatalf("terminal schedule state = %#v, want successful opaque receipt", state)
	}
	if len(a.ListFlowActionReceipts()) != 1 {
		t.Fatalf("flow receipts = %#v, want one post-read-back receipt", a.ListFlowActionReceipts())
	}
	if got, want := state.LastFire.PreparedExecutionIdentities, []string{a.ListFlowActionReceipts()[0].PreparedExecutionIdentity}; !equalStrings(got, want) {
		t.Fatalf("schedule prepared execution identities = %v, want %v", got, want)
	}

	var inspect bytes.Buffer
	require.NoError(t, runScheduleInspect(root, []string{manifest.Name}, &inspect, false))
	if !strings.Contains(inspect.String(), "Status: succeeded") || !strings.Contains(inspect.String(), state.LastFire.ReceiptIDs[0]) {
		t.Fatalf("schedule inspect = %q, want terminal status and opaque receipt", inspect.String())
	}
	statusJSON, statusErr, statusCode := scheduleRun(t, root, "schedule", "status", manifest.Name, "--json")
	if statusCode != 0 {
		t.Fatalf("schedule status: exit %d stderr=%q", statusCode, statusErr)
	}
	var statusResult struct {
		Schedule schedule.Manifest  `json:"schedule"`
		Status   schedule.FireState `json:"status"`
	}
	require.NoError(t, json.Unmarshal([]byte(statusJSON), &statusResult))
	if statusResult.Status.Status != schedule.FireStatusSucceeded || !equalStrings(statusResult.Status.LastFire.ReceiptIDs, state.LastFire.ReceiptIDs) {
		t.Fatalf("schedule status = %#v, want safe reference and terminal receipt", statusResult)
	}
	for _, forbidden := range []string{"fixture-token", plan.ApprovalToken} {
		if strings.Contains(stdout.String()+inspect.String()+statusJSON+string(installed), forbidden) {
			t.Fatalf("schedule output leaked %q", forbidden)
		}
	}
	for _, path := range []string{
		filepath.Join(root, "schedules", manifest.Name+".json"),
		filepath.Join(root, "schedules", manifest.Name+".fire.json"),
	} {
		data, readErr := os.ReadFile(path)
		require.NoError(t, readErr)
		for _, forbidden := range []string{"fixture-token", plan.ApprovalToken} {
			if strings.Contains(string(data), forbidden) {
				t.Fatalf("%s leaked %q: %s", path, forbidden, data)
			}
		}
	}

	require.NoError(t, backend.Remove(ctx, manifest.Name))
	restored, err := os.ReadFile(backendFile)
	require.NoError(t, err)
	if !bytes.Equal(restored, baseline) {
		t.Fatalf("backend cleanup = %q, want byte-for-byte restoration %q", restored, baseline)
	}
}

func TestScheduleFireScopeDriftStopsBeforeProviderRequestAndParks(t *testing.T) {
	ctx, root, a, target, manifest, plan := newAuthorizedScheduleFixture(t)
	driftReversePlanMapping(t, root, plan.ID)
	beforeEvents := target.eventCount()

	err := flowRun(ctx, config.Config{}, a, []string{manifest.Flow}, &bytes.Buffer{}, true)
	var changed *app.AuthorizationScopeChangedError
	if !errors.As(err, &changed) {
		t.Fatalf("schedule scope drift error = %T %v, want AuthorizationScopeChangedError", err, err)
	}
	if target.eventCount() != beforeEvents {
		t.Fatalf("scope drift reached the provider: events before=%d after=%d", beforeEvents, target.eventCount())
	}
	state, err := schedule.LoadFireState(root, manifest.Name)
	require.NoError(t, err)
	if state.Status != schedule.FireStatusParked || state.LastFire.StopReason != schedule.FireStopScope {
		t.Fatalf("scope drift state = %#v, want parked authorization scope", state)
	}
	err = flowRun(ctx, config.Config{}, a, []string{manifest.Flow}, &bytes.Buffer{}, true)
	if !errors.Is(err, schedule.ErrFireParked) {
		t.Fatalf("parked schedule replay error = %v, want ErrFireParked", err)
	}
}

func TestScheduleFireRevocationAndExpiryStopBeforeProviderRequest(t *testing.T) {
	t.Run("revoked", func(t *testing.T) {
		ctx, root, a, target, manifest, plan := newAuthorizedScheduleFixture(t)
		require.NoError(t, a.RevokeAuthorization(plan.AuthorizationReference))
		beforeEvents := target.eventCount()
		err := flowRun(ctx, config.Config{}, a, []string{manifest.Flow}, &bytes.Buffer{}, true)
		var revoked *app.AuthorizationRevokedError
		if !errors.As(err, &revoked) {
			t.Fatalf("revoked schedule error = %T %v, want AuthorizationRevokedError", err, err)
		}
		if target.eventCount() != beforeEvents {
			t.Fatalf("revoked schedule reached provider: before=%d after=%d", beforeEvents, target.eventCount())
		}
		state, stateErr := schedule.LoadFireState(root, manifest.Name)
		require.NoError(t, stateErr)
		if state.Status != schedule.FireStatusParked || state.LastFire.StopReason != schedule.FireStopRevoked {
			t.Fatalf("revoked fire state = %#v, want parked revoked", state)
		}
	})

	t.Run("expired", func(t *testing.T) {
		ctx, root, _, target, manifest, plan := newAuthorizedScheduleFixture(t)
		expireScheduleAuthorization(t, root, plan.AuthorizationReference)
		reopened, err := app.Open(root)
		require.NoError(t, err)
		reopened.Registry().Register(target)
		beforeEvents := target.eventCount()
		err = flowRun(ctx, config.Config{}, reopened, []string{manifest.Flow}, &bytes.Buffer{}, true)
		var expired *app.AuthorizationExpiredError
		if !errors.As(err, &expired) {
			t.Fatalf("expired schedule error = %T %v, want AuthorizationExpiredError", err, err)
		}
		if target.eventCount() != beforeEvents {
			t.Fatalf("expired schedule reached provider: before=%d after=%d", beforeEvents, target.eventCount())
		}
		state, stateErr := schedule.LoadFireState(root, manifest.Name)
		require.NoError(t, stateErr)
		if state.Status != schedule.FireStatusParked || state.LastFire.StopReason != schedule.FireStopExpired {
			t.Fatalf("expired fire state = %#v, want parked expired", state)
		}
	})
}

func TestScheduleFireRateLimitParksWithoutReplay(t *testing.T) {
	root := t.TempDir()
	manifest := schedule.Manifest{Name: "rate-limited-fire", Cron: "0 2 * * *", Flow: "flow"}
	require.NoError(t, schedule.Save(root, manifest, false))
	lease, err := schedule.BeginFire(root, manifest.Name)
	require.NoError(t, err)
	if reason := scheduleFireStopReason(&connsdk.RateLimitError{}); reason != schedule.FireStopRateLimit {
		t.Fatalf("rate-limit fire reason = %q, want %q", reason, schedule.FireStopRateLimit)
	}
	if reason := scheduleFireStopReason(&connsdk.RateBudgetRefusalError{Code: connsdk.RateBudgetRefusalSharedCoordinatorUnavailable}); reason != schedule.FireStopRateLimit {
		t.Fatalf("rate-budget refusal fire reason = %q, want %q", reason, schedule.FireStopRateLimit)
	}
	require.NoError(t, lease.Park(schedule.FireStopRateLimit))
	if _, err := schedule.BeginFire(root, manifest.Name); !errors.Is(err, schedule.ErrFireParked) {
		t.Fatalf("rate-limited fire replay error = %v, want ErrFireParked", err)
	}
}

func TestScheduleFireCancellationParksBeforeProviderOrCheckpoint(t *testing.T) {
	ctx, root, a, target, manifest, _ := newAuthorizedScheduleFixture(t)
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	beforeEvents := target.eventCount()

	err := flowRun(canceled, config.Config{}, a, []string{manifest.Flow}, &bytes.Buffer{}, true)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled schedule fire error = %T %v, want context.Canceled", err, err)
	}
	if target.eventCount() != beforeEvents {
		t.Fatalf("cancelled schedule fire reached provider: before=%d after=%d", beforeEvents, target.eventCount())
	}
	checkpoint := &flow.FileCheckpointStore{Dir: a.ProjectDir()}
	status, checkpointErr := checkpoint.Get(manifest.Flow, "create-targets")
	require.NoError(t, checkpointErr)
	if status != "" {
		t.Fatalf("cancelled schedule checkpoint = %q, want empty", status)
	}
	state, stateErr := schedule.LoadFireState(root, manifest.Name)
	require.NoError(t, stateErr)
	if state.Status != schedule.FireStatusParked || state.LastFire.StopReason != schedule.FireStopFailed {
		t.Fatalf("cancelled schedule state = %#v, want parked", state)
	}
}

func TestScheduleFirePartialWriteParksPreparedExecutionAndNeverCheckpoints(t *testing.T) {
	ctx, root, a, target, manifest, _ := newAuthorizedScheduleFixture(t)
	target.setWriteError(errors.New("provider result unknown after dispatch"))
	beforeEvents := target.eventCount()
	beforeWrites := target.writeCalls()

	err := flowRun(ctx, config.Config{}, a, []string{manifest.Flow}, &bytes.Buffer{}, true)
	if err == nil || !strings.Contains(err.Error(), "provider result unknown") {
		t.Fatalf("partial schedule write error = %v", err)
	}
	if target.writeCalls() != beforeWrites || target.eventCount() <= beforeEvents {
		t.Fatalf("partial schedule write accounting writes=%d events=%d", target.writeCalls()-beforeWrites, target.eventCount()-beforeEvents)
	}
	checkpoint := &flow.FileCheckpointStore{Dir: a.ProjectDir()}
	status, checkpointErr := checkpoint.Get(manifest.Flow, "create-targets")
	require.NoError(t, checkpointErr)
	if status != "" {
		t.Fatalf("partial schedule checkpoint = %q, want empty", status)
	}
	state, stateErr := schedule.LoadFireState(root, manifest.Name)
	require.NoError(t, stateErr)
	if state.Status != schedule.FireStatusParked || state.LastFire.StopReason != schedule.FireStopFailed || len(a.ListFlowActionReceipts()) != 0 {
		t.Fatalf("partial schedule terminal state = %#v receipts=%v", state, a.ListFlowActionReceipts())
	}
	if _, statErr := os.Stat(filepath.Join(root, "schedules", manifest.Name+".fire.lock")); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("terminal state was not persisted before lock cleanup: stat error = %v", statErr)
	}
	afterEvents := target.eventCount()
	if replayErr := flowRun(ctx, config.Config{}, a, []string{manifest.Flow}, &bytes.Buffer{}, true); !errors.Is(replayErr, schedule.ErrFireParked) {
		t.Fatalf("partial schedule replay error = %v, want ErrFireParked", replayErr)
	}
	if target.eventCount() != afterEvents {
		t.Fatalf("parked schedule replay reached provider: before=%d after=%d", afterEvents, target.eventCount())
	}
}

func newAuthorizedScheduleFixture(t *testing.T) (context.Context, string, *app.App, *cliFlowActionTarget, schedule.Manifest, app.ReversePlan) {
	t.Helper()
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
	approvalToken := plan.ApprovalToken
	_, err = a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID: plan.ID, ApprovalToken: plan.ApprovalToken, Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	})
	require.NoError(t, err)
	plan, err = a.GetReversePlan(plan.ID)
	require.NoError(t, err)
	plan.ApprovalToken = approvalToken
	require.NotEmpty(t, plan.AuthorizationReference)
	root := filepath.Dir(a.ProjectDir())
	manifest := schedule.Manifest{Name: "nightly-authorized-flow", Cron: "0 2 * * *", Flow: "scheduled-flow", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	writeScheduledFlowManifest(t, a, manifest.Flow, plan.ID)
	require.NoError(t, schedule.Save(root, manifest, false))
	return ctx, root, a, target, manifest, plan
}

func writeScheduledFlowManifest(t *testing.T, a *app.App, flowName, jobReference string) {
	t.Helper()
	flowsDir := filepath.Join(a.ProjectDir(), "flows")
	require.NoError(t, os.MkdirAll(flowsDir, 0o755))
	data := fmt.Sprintf(`{
  "version": 1,
  "name": %q,
	  "steps": [{
	    "id": "create-targets",
	    "kind": "action",
	    "job": %q,
	    "action_cfg": {
	      "read_back_stream": "targets"
	    }
	  }]
}`, flowName, jobReference)
	require.NoError(t, os.WriteFile(filepath.Join(flowsDir, flowName+".json"), []byte(data), 0o600))
}

func driftReversePlanMapping(t *testing.T, root, planID string) {
	t.Helper()
	path := filepath.Join(root, ".polymetrics", "state", "state.json")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var state map[string]any
	require.NoError(t, json.Unmarshal(data, &state))
	plans, ok := state["reverse_plans"].([]any)
	require.True(t, ok)
	changed := false
	for _, raw := range plans {
		plan, ok := raw.(map[string]any)
		if !ok || plan["id"] != planID {
			continue
		}
		plan["mappings"] = map[string]any{"id": "changed-id"}
		changed = true
	}
	require.True(t, changed)
	updated, err := json.MarshalIndent(state, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, updated, 0o600))
}

func expireScheduleAuthorization(t *testing.T, root, reference string) {
	t.Helper()
	path := filepath.Join(root, ".polymetrics", "state", "state.json")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &raw))
	var records []app.AuthorizationRecord
	require.NoError(t, json.Unmarshal(raw["authorizations"], &records))
	for i := range records {
		if records[i].Reference != reference {
			continue
		}
		records[i].Scope.ExpiresAt = time.Now().UTC().Add(-time.Minute)
		identity, identityErr := app.AuthorizationScopeIdentity(records[i].Scope)
		require.NoError(t, identityErr)
		records[i].ScopeIdentity = identity
	}
	raw["authorizations"], err = json.Marshal(records)
	require.NoError(t, err)
	updated, err := json.Marshal(raw)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, updated, 0o600))
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
