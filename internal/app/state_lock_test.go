package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	statestore "polymetrics.ai/internal/state"
)

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
