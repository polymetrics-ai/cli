package state_test

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"polymetrics.ai/internal/state"
)

type testConfig struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func TestJSONStoreSaveLoadAndKeepsPreviousFileOnFailedSave(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	store := state.JSONStore[map[string]any]{
		Path: path,
		Initial: func() map[string]any {
			return map[string]any{"name": "initial"}
		},
	}

	initial, err := store.Load()
	if err != nil {
		t.Fatalf("Load() initial error = %v", err)
	}
	if initial["name"] != "initial" {
		t.Fatalf("Load() initial name = %v, want initial", initial["name"])
	}

	if err := store.Save(map[string]any{"name": "saved"}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load() saved error = %v", err)
	}
	if got["name"] != "saved" {
		t.Fatalf("Load() saved name = %v, want saved", got["name"])
	}

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() before failed save error = %v", err)
	}
	if err := store.Save(map[string]any{"bad": func() {}}); err == nil {
		t.Fatal("Save() with unmarshalable value succeeded, want error")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() after failed save error = %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("failed Save() changed persisted file\nbefore: %s\nafter: %s", before, after)
	}

	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != filepath.Base(path) {
		t.Fatalf("state directory entries after failed save = %v, want only %s", entryNames(entries), filepath.Base(path))
	}
}

func TestJSONStoreUpdateUsesLockerToSerializeUpdates(t *testing.T) {
	locker := &fakeLocker{}
	store := state.JSONStore[testConfig]{
		Path:    filepath.Join(t.TempDir(), "state.json"),
		Initial: func() testConfig { return testConfig{} },
		Locker:  locker,
	}

	const updates = 32
	start := make(chan struct{})
	errs := make(chan error, updates)
	var wg sync.WaitGroup
	for i := 0; i < updates; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := store.Update(func(current testConfig) (testConfig, error) {
				current.Count++
				return current, nil
			})
			errs <- err
		}()
	}

	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Count != updates {
		t.Fatalf("Count = %d, want %d", got.Count, updates)
	}
	if locker.maxActive != 1 {
		t.Fatalf("max active locks = %d, want 1", locker.maxActive)
	}
	if locker.calls != updates+1 {
		t.Fatalf("lock calls = %d, want %d", locker.calls, updates+1)
	}
}

func TestJSONStoreUpdateUnlocksWhenCallbackReturnsError(t *testing.T) {
	locker := &fakeLocker{}
	store := state.JSONStore[testConfig]{
		Path:   filepath.Join(t.TempDir(), "state.json"),
		Locker: locker,
	}
	wantErr := errors.New("stop")

	_, err := store.Update(func(current testConfig) (testConfig, error) {
		current.Count++
		return current, wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Update() error = %v, want %v", err, wantErr)
	}
	if locker.active != 0 {
		t.Fatalf("active locks = %d, want 0", locker.active)
	}
}

func TestJSONStoreRedactedSnapshot(t *testing.T) {
	type credentials struct {
		Name   string         `json:"name"`
		Secret string         `json:"secret"`
		Nested map[string]any `json:"nested"`
	}
	store := state.JSONStore[credentials]{
		Path: filepath.Join(t.TempDir(), "state.json"),
		Redact: func(path []string, value any) any {
			if len(path) == 0 {
				return value
			}
			switch path[len(path)-1] {
			case "secret", "token":
				return "***"
			default:
				return value
			}
		},
	}
	want := credentials{
		Name:   "service",
		Secret: "top-secret",
		Nested: map[string]any{"token": "abc123", "visible": "yes"},
	}
	if err := store.Save(want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	snapshot, err := store.RedactedSnapshot()
	if err != nil {
		t.Fatalf("RedactedSnapshot() error = %v", err)
	}
	root, ok := snapshot.(map[string]any)
	if !ok {
		t.Fatalf("RedactedSnapshot() type = %T, want map[string]any", snapshot)
	}
	if root["name"] != "service" {
		t.Fatalf("name = %v, want service", root["name"])
	}
	if root["secret"] != "***" {
		t.Fatalf("secret = %v, want ***", root["secret"])
	}
	nested, ok := root["nested"].(map[string]any)
	if !ok {
		t.Fatalf("nested type = %T, want map[string]any", root["nested"])
	}
	if nested["token"] != "***" {
		t.Fatalf("nested token = %v, want ***", nested["token"])
	}
	if nested["visible"] != "yes" {
		t.Fatalf("nested visible = %v, want yes", nested["visible"])
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() after snapshot error = %v", err)
	}
	if loaded.Secret != want.Secret || loaded.Nested["token"] != want.Nested["token"] {
		t.Fatalf("RedactedSnapshot() mutated stored state: %#v", loaded)
	}
}

func TestFileLockUsesExclusiveLockFile(t *testing.T) {
	lock := state.FileLock{Path: filepath.Join(t.TempDir(), "state.lock")}
	unlock, err := lock.Lock()
	if err != nil {
		t.Fatalf("Lock() error = %v", err)
	}
	if _, err := lock.Lock(); err == nil {
		t.Fatal("second Lock() succeeded, want error")
	}
	if err := unlock(); err != nil {
		t.Fatalf("unlock() error = %v", err)
	}
	if err := unlock(); err != nil {
		t.Fatalf("second call to unlock() error = %v", err)
	}
	unlock, err = lock.Lock()
	if err != nil {
		t.Fatalf("Lock() after unlock error = %v", err)
	}
	if err := unlock(); err != nil {
		t.Fatalf("second unlock() error = %v", err)
	}
}

type fakeLocker struct {
	mu        sync.Mutex
	active    int
	maxActive int
	calls     int
}

func (l *fakeLocker) Lock() (func() error, error) {
	l.mu.Lock()
	l.active++
	l.calls++
	if l.active > l.maxActive {
		l.maxActive = l.active
	}
	return func() error {
		l.active--
		l.mu.Unlock()
		return nil
	}, nil
}

func entryNames(entries []os.DirEntry) []string {
	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name()
	}
	return names
}
