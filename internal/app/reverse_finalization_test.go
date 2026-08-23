package app

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
)

var errReverseFinalizationPersist = errors.New("reverse finalization state persistence failed")

func TestReverseFinalization_DoesNotPublishUnpersistedRun(t *testing.T) {
	tests := []struct {
		name       string
		configure  func(*App)
		wantStored bool
	}{
		{
			name: "definite pre-rename failure",
			configure: func(a *App) {
				a.store.Locker = &reverseFinalizationFailAtLock{failAt: 1, err: errReverseFinalizationPersist}
			},
		},
		{
			name: "post-rename directory sync uncertainty",
			configure: func(a *App) {
				a.store.SyncDirectory = func(string) error { return errReverseFinalizationPersist }
			},
			wantStored: true,
		},
	}
	for _, tt := range tests {
		for _, direct := range []bool{false, true} {
			name := "regular"
			if direct {
				name = "direct-write"
			}
			t.Run(tt.name+"/"+name, func(t *testing.T) {
				a := newReverseFinalizationTestApp(t)
				tt.configure(a)
				run := ReverseRun{ID: "rrun-terminal", PlanID: "rplan-terminal", Status: "running", RecordsStaged: 2, StartedAt: time.Unix(100, 0).UTC()}
				providerErr := errors.New("provider reported partial write")
				result := connectors.WriteResult{RecordsWritten: 1, RecordsFailed: 1}
				var got ReverseRun
				var err error
				if direct {
					got, err = a.finishOperationDirectWrite(run.PlanID, run, result, connectors.RuntimeConfig{}, run.RecordsStaged, providerErr)
				} else {
					got, err = a.finishReverseWrite(run.PlanID, run, result, connectors.RuntimeConfig{}, run.RecordsStaged, providerErr)
				}
				if !errors.Is(err, providerErr) {
					t.Fatalf("finalizer error = %v, want original provider error", err)
				}

				loaded, loadErr := a.store.LoadReadOnly()
				if loadErr != nil {
					t.Fatalf("LoadReadOnly() error = %v", loadErr)
				}
				storedRun, storedRunFound := reverseRunFromState(loaded, run.ID)
				storedPlan, planErr := reversePlanFromState(loaded, run.PlanID)
				if planErr != nil {
					t.Fatalf("durable reverse plan %q not found: %v", run.PlanID, planErr)
				}
				if !tt.wantStored {
					if got.ID != "" {
						t.Fatalf("finalizer returned %#v after definite non-commit, want zero run", got)
					}
					if !errors.Is(err, errReverseFinalizationPersist) {
						t.Fatalf("finalizer error = %v, want joined definite persistence error", err)
					}
					if storedRunFound || storedPlan.Status != "executing" {
						t.Fatalf("definite non-commit durable state has run=%#v present=%t plan=%#v, want no run and executing recovery-held plan", storedRun, storedRunFound, storedPlan)
					}
					return
				}

				if !storedRunFound || storedRun.Status != "failed" || storedRun.CompletedAt.IsZero() || storedPlan.Status != "failed" {
					t.Fatalf("durable terminal state run=%#v present=%t plan=%#v, want failed terminal run and plan", storedRun, storedRunFound, storedPlan)
				}
				if !errors.Is(err, errReverseFinalizationPersist) {
					t.Fatalf("finalizer error = %v, want joined post-rename persistence outcome", err)
				}
				if !reflect.DeepEqual(got, storedRun) {
					t.Fatalf("finalizer run = %#v, want exact reloaded durable run %#v", got, storedRun)
				}
			})
		}
	}
}

func TestReverseFinalizationRejectsIncompleteNilErrorAcknowledgements(t *testing.T) {
	tests := []struct {
		name       string
		action     string
		result     connectors.WriteResult
		wantFailed bool
	}{
		{name: "partial write", action: "create_issue", result: connectors.WriteResult{RecordsWritten: 1}, wantFailed: true},
		{name: "zero acknowledgement", action: "create_issue", result: connectors.WriteResult{}, wantFailed: true},
		{name: "negative written", action: "create_issue", result: connectors.WriteResult{RecordsWritten: -1, RecordsFailed: 3}, wantFailed: true},
		{name: "over-counted write", action: "create_issue", result: connectors.WriteResult{RecordsWritten: 3}, wantFailed: true},
		{name: "unchanged forbidden for create", action: "create_issue", result: connectors.WriteResult{RecordsUnchanged: 2}, wantFailed: true},
		{name: "unchanged forbidden for delete without declared missing status", action: "delete_label", result: connectors.WriteResult{RecordsUnchanged: 2}, wantFailed: true},
		{name: "unchanged allowed for idempotent delete", action: "delete_issue_comment", result: connectors.WriteResult{RecordsUnchanged: 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newReverseFinalizationTestApp(t)
			plan := a.state.ReversePlans[0]
			plan.DestinationConnector = "github"
			plan.Action = tt.action
			a.state.ReversePlans[0] = plan
			if err := a.save(); err != nil {
				t.Fatalf("save action policy plan: %v", err)
			}
			run := ReverseRun{ID: "rrun-ack", PlanID: plan.ID, Status: "running", RecordsStaged: 2, StartedAt: time.Unix(100, 0).UTC()}
			got, err := a.finishReverseWrite(plan.ID, run, tt.result, connectors.RuntimeConfig{}, run.RecordsStaged, nil)
			stored, planErr := a.GetReversePlan(plan.ID)
			if planErr != nil {
				t.Fatalf("GetReversePlan(%q): %v", plan.ID, planErr)
			}
			if tt.wantFailed {
				if err == nil {
					t.Fatal("finalizer accepted incomplete nil-error acknowledgement")
				}
				if got.Status != "failed" || stored.Status != "failed" {
					t.Fatalf("run/plan status = %q/%q, want failed/failed", got.Status, stored.Status)
				}
				if got.RecordsSucceeded < 0 || got.RecordsFailed < 0 {
					t.Fatalf("failed acknowledgement persisted negative counters: %#v", got)
				}
				return
			}
			if err != nil || got.Status != "completed" || stored.Status != "executed" {
				t.Fatalf("allowed unchanged acknowledgement = run=%#v plan=%#v err=%v, want completed/executed", got, stored, err)
			}
		})
	}
}

func TestOrdinaryETLTerminalPersistenceReloadsExactDurableRun(t *testing.T) {
	tests := []struct {
		name       string
		configure  func(*App)
		wantStored bool
	}{
		{
			name: "definite pre-rename failure",
			configure: func(a *App) {
				a.store.Locker = &reverseFinalizationFailAtLock{failAt: 1, err: errReverseFinalizationPersist}
			},
		},
		{
			name: "post-rename directory sync uncertainty",
			configure: func(a *App) {
				a.store.SyncDirectory = func(string) error { return errReverseFinalizationPersist }
			},
			wantStored: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newReverseFinalizationTestApp(t)
			run, err := a.beginRun(Run{ID: "run-terminal", Type: "etl", Status: "running", StartedAt: time.Unix(100, 0).UTC()})
			if err != nil {
				t.Fatalf("beginRun() error = %v", err)
			}
			tt.configure(a)
			operationalErr := errors.New("ordinary source failed")
			got, err := a.failRun(run.ID, operationalErr)
			if !errors.Is(err, operationalErr) {
				t.Fatalf("failRun() error = %v, want original operational error", err)
			}

			loaded, loadErr := a.store.LoadReadOnly()
			if loadErr != nil {
				t.Fatalf("LoadReadOnly() error = %v", loadErr)
			}
			stored, found := terminalETLRunFromState(loaded, run.ID)
			if !tt.wantStored {
				if got.ID != "" {
					t.Fatalf("failRun() returned %#v after definite non-commit, want zero run", got)
				}
				if !errors.Is(err, errReverseFinalizationPersist) {
					t.Fatalf("failRun() error = %v, want joined persistence error", err)
				}
				if found == nil {
					t.Fatalf("definite non-commit stored terminal run %#v", stored)
				}
				return
			}
			if found != nil {
				t.Fatalf("durable terminal run error = %v", found)
			}
			if !errors.Is(err, errReverseFinalizationPersist) {
				t.Fatalf("failRun() error = %v, want joined persistence outcome", err)
			}
			if !reflect.DeepEqual(got, stored) {
				t.Fatalf("failRun() = %#v, want exact reloaded durable run %#v", got, stored)
			}
		})
	}
}

func newReverseFinalizationTestApp(t *testing.T) *App {
	t.Helper()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	a.state.ReversePlans = []ReversePlan{{ID: "rplan-terminal", Status: "executing"}}
	if err := a.save(); err != nil {
		t.Fatalf("save reverse finalization plan error = %v", err)
	}
	return a
}

func reverseRunFromState(loaded state, id string) (ReverseRun, bool) {
	for _, run := range loaded.ReverseRuns {
		if run.ID == id {
			return run, true
		}
	}
	return ReverseRun{}, false
}

type reverseFinalizationFailAtLock struct {
	calls  int
	failAt int
	err    error
}

func (l *reverseFinalizationFailAtLock) Lock() (func() error, error) {
	l.calls++
	if l.calls == l.failAt {
		return nil, l.err
	}
	return func() error { return nil }, nil
}
