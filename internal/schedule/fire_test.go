package schedule

import (
	"errors"
	"os"
	"testing"
)

func TestFireLeasePersistsTerminalReceiptAndRefusesOverlapOrCrashReplay(t *testing.T) {
	root := t.TempDir()
	manifest := makeManifest("authorized-fire")
	if err := Save(root, manifest, false); err != nil {
		t.Fatalf("Save: %v", err)
	}

	lease, err := BeginFire(root, manifest.Name)
	if err != nil {
		t.Fatalf("BeginFire: %v", err)
	}
	if _, err := BeginFire(root, manifest.Name); !errors.Is(err, ErrFireInProgress) {
		t.Fatalf("overlapping BeginFire error = %v, want ErrFireInProgress", err)
	}

	if err := lease.Complete(FireReceipt{FlowName: manifest.Flow, FlowStatus: "ok", ReceiptIDs: []string{"fact_safe_receipt"}}); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	state, err := LoadFireState(root, manifest.Name)
	if err != nil {
		t.Fatalf("LoadFireState: %v", err)
	}
	if state.Status != FireStatusSucceeded || state.LastFire.FlowStatus != "ok" || state.LastFire.AuthorizationReference != manifest.AuthorizationReference || len(state.LastFire.ReceiptIDs) != 1 {
		t.Fatalf("terminal fire state = %#v, want successful safe receipt state", state)
	}
}

func TestInterruptedFireHaltsWithoutReplay(t *testing.T) {
	root := t.TempDir()
	manifest := makeManifest("crashed-fire")
	if err := Save(root, manifest, false); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := BeginFire(root, manifest.Name); err != nil {
		t.Fatalf("BeginFire: %v", err)
	}
	state, err := LoadFireState(root, manifest.Name)
	if err != nil {
		t.Fatalf("LoadFireState: %v", err)
	}
	if state.Status != FireStatusRunning {
		t.Fatalf("interrupted fire state = %#v, want running halt state", state)
	}
	if err := os.Remove(fireLockPath(root, manifest.Name)); err != nil {
		t.Fatalf("remove stale fire lock: %v", err)
	}
	if _, err := BeginFire(root, manifest.Name); !errors.Is(err, ErrFireInProgress) {
		t.Fatalf("crashed fire replay error after lock loss = %v, want ErrFireInProgress", err)
	}
}

func TestDeleteRefusesAnActiveFire(t *testing.T) {
	root := t.TempDir()
	manifest := makeManifest("active-delete")
	if err := Save(root, manifest, false); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := BeginFire(root, manifest.Name); err != nil {
		t.Fatalf("BeginFire: %v", err)
	}
	if err := Delete(root, manifest.Name); !errors.Is(err, ErrFireInProgress) {
		t.Fatalf("Delete active fire error = %v, want ErrFireInProgress", err)
	}
	if _, err := Load(root, manifest.Name); err != nil {
		t.Fatalf("active schedule manifest removed: %v", err)
	}
}

func TestParkedFireNeverReplays(t *testing.T) {
	root := t.TempDir()
	manifest := makeManifest("parked-fire")
	if err := Save(root, manifest, false); err != nil {
		t.Fatalf("Save: %v", err)
	}
	lease, err := BeginFire(root, manifest.Name)
	if err != nil {
		t.Fatalf("BeginFire: %v", err)
	}
	if err := lease.Park(FireStopAmbiguous); err != nil {
		t.Fatalf("Park: %v", err)
	}
	if _, err := BeginFire(root, manifest.Name); !errors.Is(err, ErrFireParked) {
		t.Fatalf("replay BeginFire error = %v, want ErrFireParked", err)
	}
}
