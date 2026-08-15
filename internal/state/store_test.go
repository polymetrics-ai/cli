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

func TestJSONStoreUpdateReportsCommittedOutcomeAfterUnlockFailure(t *testing.T) {
	store := state.JSONStore[testConfig]{
		Path:   filepath.Join(t.TempDir(), "state.json"),
		Locker: &failingUnlockLocker{},
	}

	updated, err := store.Update(func(current testConfig) (testConfig, error) {
		current.Count++
		return current, nil
	})
	var outcome *state.CommitOutcomeError
	if !errors.As(err, &outcome) {
		t.Fatalf("Update() error = %T %v, want CommitOutcomeError", err, err)
	}
	if outcome.Outcome != state.CommitOutcomeCommitted || !outcome.Outcome.MayHaveCommitted() {
		t.Fatalf("commit outcome = %q, want committed", outcome.Outcome)
	}
	if updated.Count != 1 {
		t.Fatalf("updated count = %d, want 1", updated.Count)
	}

	persisted, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if persisted.Count != updated.Count {
		t.Fatalf("persisted count = %d, want %d", persisted.Count, updated.Count)
	}
}

func TestJSONStoreUpdateReportsIndeterminateOutcomeAfterDirectorySyncFailure(t *testing.T) {
	wantErr := errors.New("directory sync failed")
	store := state.JSONStore[testConfig]{
		Path: filepath.Join(t.TempDir(), "state.json"),
		SyncDirectory: func(string) error {
			return wantErr
		},
	}

	updated, err := store.Update(func(current testConfig) (testConfig, error) {
		current.Count++
		return current, nil
	})
	var outcome *state.CommitOutcomeError
	if !errors.As(err, &outcome) {
		t.Fatalf("Update() error = %T %v, want CommitOutcomeError", err, err)
	}
	if outcome.Outcome != state.CommitOutcomeIndeterminate || !outcome.Outcome.MayHaveCommitted() {
		t.Fatalf("commit outcome = %q, want indeterminate", outcome.Outcome)
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("Update() error = %v, want wrapped %v", err, wantErr)
	}
	if updated.Count != 1 {
		t.Fatalf("updated count = %d, want 1", updated.Count)
	}

	persisted, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if persisted.Count != updated.Count {
		t.Fatalf("persisted count = %d, want %d", persisted.Count, updated.Count)
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

func TestJSONStoreUpdateAfterPreflightRejectsWithoutStateLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	store := state.JSONStore[testConfig]{
		Path:   path,
		Locker: state.FileLock{Path: path + ".lock"},
	}
	if err := store.Save(testConfig{Name: "active"}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(before) error = %v", err)
	}

	wantErr := errors.New("approval already consumed")
	_, err = store.UpdateAfterPreflight(func(testConfig) error {
		return wantErr
	}, func(current testConfig) (testConfig, error) {
		current.Count++
		return current, nil
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("UpdateAfterPreflight() error = %v, want %v", err, wantErr)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(after) error = %v", err)
	}
	if string(after) != string(before) {
		t.Fatal("preflight rejection rewrote state")
	}
	if _, err := os.Stat(path + ".lock"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("preflight rejection created a state lock: %v", err)
	}
}

func TestJSONStoreUpdateAfterPreflightHonorsLegacyStateLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	legacy := state.JSONStore[testConfig]{
		Path:   path,
		Locker: state.FileLock{Path: path + ".lock"},
	}
	if err := legacy.Save(testConfig{Name: "active"}); err != nil {
		t.Fatalf("legacy Save() error = %v", err)
	}

	entered := make(chan struct{})
	release := make(chan struct{})
	legacyResult := make(chan error, 1)
	go func() {
		_, err := legacy.Update(func(current testConfig) (testConfig, error) {
			close(entered)
			<-release
			current.Count++
			return current, nil
		})
		legacyResult <- err
	}()
	<-entered

	released := false
	defer func() {
		if !released {
			close(release)
		}
	}()
	store := state.JSONStore[testConfig]{
		Path:   path,
		Locker: state.FileLock{Path: path + ".lock"},
	}
	updated := false
	_, err := store.UpdateAfterPreflight(func(current testConfig) error {
		if current.Name != "active" {
			return errors.New("unexpected state")
		}
		return nil
	}, func(current testConfig) (testConfig, error) {
		updated = true
		current.Count++
		return current, nil
	})
	if !errors.Is(err, os.ErrExist) {
		t.Fatalf("UpdateAfterPreflight() error = %v, want legacy lock conflict", err)
	}
	if updated {
		t.Fatal("update ran while the legacy lock was held")
	}

	close(release)
	released = true
	if err := <-legacyResult; err != nil {
		t.Fatalf("legacy Update() error = %v", err)
	}
	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.Count != 1 {
		t.Fatalf("Count = %d, want only the legacy update", loaded.Count)
	}
}

type fakeLocker struct {
	mu        sync.Mutex
	active    int
	maxActive int
	calls     int
}

type failingUnlockLocker struct {
	failed bool
}

func (l *failingUnlockLocker) Lock() (func() error, error) {
	return func() error {
		if l.failed {
			return nil
		}
		l.failed = true
		return errors.New("unlock failed")
	}, nil
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
