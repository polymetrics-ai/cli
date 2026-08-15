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
	if !strings.Contains(string(installed), "schedule fire "+manifest.Name+" --authorization "+manifest.AuthorizationReference) {
		t.Fatalf("installed payload = %q, want authorized schedule fire", installed)
	}

	beforeWrites := target.writeCalls()
	var stdout bytes.Buffer
	err = runScheduleFireWithApp(ctx, config.Config{}, root, a, []string{manifest.Name, "--authorization", manifest.AuthorizationReference}, &stdout, true)
	require.NoError(t, err)
	if target.writeCalls() != beforeWrites+1 {
		t.Fatalf("scheduled fire writes = %d, want %d", target.writeCalls(), beforeWrites+1)
	}
	if got := target.tailEvents(4); !equalStrings(got, []string{"validate", "preview", "write", "read"}) {
		t.Fatalf("scheduled target events = %v, want typed validation/write/read-back", got)
	}
	state, err := schedule.LoadFireState(root, manifest.Name)
	require.NoError(t, err)
	if state.Status != schedule.FireStatusSucceeded || state.LastFire.FlowStatus != "ok" || len(state.LastFire.ReceiptIDs) != 1 || state.LastFire.AuthorizationReference != manifest.AuthorizationReference {
		t.Fatalf("terminal schedule state = %#v, want successful opaque receipt", state)
	}
	if len(a.ListFlowActionReceipts()) != 1 {
		t.Fatalf("flow receipts = %#v, want one post-read-back receipt", a.ListFlowActionReceipts())
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
	if statusResult.Schedule.AuthorizationReference != manifest.AuthorizationReference || statusResult.Status.Status != schedule.FireStatusSucceeded || !equalStrings(statusResult.Status.LastFire.ReceiptIDs, state.LastFire.ReceiptIDs) {
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
	ctx, root, a, target, manifest, _ := newAuthorizedScheduleFixture(t)
	writeScheduledFlowManifest(t, a, manifest.Flow, target.Name(), "changed-id")
	beforeEvents := target.eventCount()

	err := runScheduleFireWithApp(ctx, config.Config{}, root, a, []string{manifest.Name, "--authorization", manifest.AuthorizationReference}, &bytes.Buffer{}, true)
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
	err = runScheduleFireWithApp(ctx, config.Config{}, root, a, []string{manifest.Name, "--authorization", manifest.AuthorizationReference}, &bytes.Buffer{}, true)
	if !errors.Is(err, schedule.ErrFireParked) {
		t.Fatalf("parked schedule replay error = %v, want ErrFireParked", err)
	}
}

func TestScheduleFireRevocationAndExpiryStopBeforeProviderRequest(t *testing.T) {
	t.Run("revoked", func(t *testing.T) {
		ctx, root, a, target, manifest, _ := newAuthorizedScheduleFixture(t)
		require.NoError(t, a.RevokeAuthorization(manifest.AuthorizationReference))
		beforeEvents := target.eventCount()
		err := runScheduleFireWithApp(ctx, config.Config{}, root, a, []string{manifest.Name, "--authorization", manifest.AuthorizationReference}, &bytes.Buffer{}, true)
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
		ctx, root, _, target, manifest, _ := newAuthorizedScheduleFixture(t)
		expireScheduleAuthorization(t, root, manifest.AuthorizationReference)
		reopened, err := app.Open(root)
		require.NoError(t, err)
		reopened.Registry().Register(target)
		beforeEvents := target.eventCount()
		err = runScheduleFireWithApp(ctx, config.Config{}, root, reopened, []string{manifest.Name, "--authorization", manifest.AuthorizationReference}, &bytes.Buffer{}, true)
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
	manifest := schedule.Manifest{
		Name: "rate-limited-fire", Cron: "0 2 * * *", Flow: "flow", AuthorizationReference: "auth_0123456789abcdef",
	}
	require.NoError(t, schedule.Save(root, manifest, false))
	lease, err := schedule.BeginFire(root, manifest.Name)
	require.NoError(t, err)
	if reason := scheduleFireStopReason(&connsdk.RateLimitError{}); reason != schedule.FireStopRateLimit {
		t.Fatalf("rate-limit fire reason = %q, want %q", reason, schedule.FireStopRateLimit)
	}
	require.NoError(t, lease.Park(schedule.FireStopRateLimit))
	if _, err := schedule.BeginFire(root, manifest.Name); !errors.Is(err, schedule.ErrFireParked) {
		t.Fatalf("rate-limited fire replay error = %v, want ErrFireParked", err)
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
	_, err = a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID: plan.ID, ApprovalToken: plan.ApprovalToken, Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	})
	require.NoError(t, err)
	authorizations := a.ListAuthorizations()
	require.Len(t, authorizations, 1)
	root := filepath.Dir(a.ProjectDir())
	manifest := schedule.Manifest{
		Name:                   "nightly-authorized-flow",
		Cron:                   "0 2 * * *",
		Flow:                   "scheduled-flow",
		AuthorizationReference: authorizations[0].Reference,
		CreatedAt:              time.Now().UTC(),
		UpdatedAt:              time.Now().UTC(),
	}
	writeScheduledFlowManifest(t, a, manifest.Flow, target.Name(), "id")
	require.NoError(t, schedule.Save(root, manifest, false))
	return ctx, root, a, target, manifest, plan
}

func writeScheduledFlowManifest(t *testing.T, a *app.App, flowName, targetName, mapping string) {
	t.Helper()
	flowsDir := filepath.Join(a.ProjectDir(), "flows")
	require.NoError(t, os.MkdirAll(flowsDir, 0o755))
	data := fmt.Sprintf(`{
  "version": 1,
  "name": %q,
  "steps": [{
    "id": "create-targets",
    "kind": "action",
    "action_cfg": {
      "source_table": "records",
      "source_connection": "acme",
      "destination_connector": %q,
      "destination_credential": "flow-target",
      "destination_table": "flow-action-target",
      "action": "create",
      "mappings": {"id": %q},
      "read_back_stream": "targets"
    }
  }]
}`, flowName, targetName, mapping)
	require.NoError(t, os.WriteFile(filepath.Join(flowsDir, flowName+".json"), []byte(data), 0o600))
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
