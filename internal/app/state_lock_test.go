package app

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	statestore "polymetrics.ai/internal/state"
)

func TestOpenReadsCurrentStateWhileLegacyStateLockIsHeld(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	opened, err := Open(root)
	if err != nil {
		t.Fatalf("initial Open() error = %v", err)
	}

	legacyLock := statestore.FileLock{Path: filepath.Join(root, ".polymetrics", "state", "state.json.lock")}
	unlock, err := legacyLock.Lock()
	if err != nil {
		t.Fatalf("legacy Lock() error = %v", err)
	}
	t.Cleanup(func() {
		if err := unlock(); err != nil {
			t.Errorf("legacy unlock() error = %v", err)
		}
	})

	reopened, err := Open(root)
	if err != nil {
		t.Fatalf("Open() while a concurrent writer holds the state lock error = %v", err)
	}
	if reopened.state.Revision != opened.state.Revision {
		t.Fatalf("reopened state revision = %d, want snapshot revision %d", reopened.state.Revision, opened.state.Revision)
	}
}

func TestOpenRejectsMalformedStateBeforeCreatingDurableParkingStore(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	projectDir := filepath.Join(root, ".polymetrics")
	if err := os.WriteFile(filepath.Join(projectDir, "state", "state.json"), []byte(`{"revision":`), 0o600); err != nil {
		t.Fatalf("write malformed state: %v", err)
	}

	_, err := Open(root)
	var syntaxErr *json.SyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Fatalf("Open() error = %v, want wrapped *json.SyntaxError", err)
	}
	if _, statErr := os.Stat(filepath.Join(projectDir, "state", "rate-parking.json")); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("malformed state created durable parking store: %v", statErr)
	}
}

func TestNewStateStoreMutationsHonorLegacyStateLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	store := newStateStore(path)
	if err := store.Save(state{}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	legacyLock := statestore.FileLock{Path: path + ".lock"}
	unlock, err := legacyLock.Lock()
	if err != nil {
		t.Fatalf("legacy Lock() error = %v", err)
	}
	t.Cleanup(func() {
		if err := unlock(); err != nil {
			t.Errorf("legacy unlock() error = %v", err)
		}
	})

	for _, tc := range []struct {
		name   string
		mutate func() error
	}{
		{
			name: "save",
			mutate: func() error {
				return store.Save(state{})
			},
		},
		{
			name: "update",
			mutate: func() error {
				_, err := store.Update(func(current state) (state, error) {
					return current, nil
				})
				return err
			},
		},
		{
			name: "update after preflight",
			mutate: func() error {
				_, err := store.UpdateAfterPreflight(func(state) error {
					return nil
				}, func(current state) (state, error) {
					return current, nil
				})
				return err
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.mutate(); !errors.Is(err, os.ErrExist) {
				t.Fatalf("mutation error = %v, want legacy state lock conflict", err)
			}
		})
	}
}
