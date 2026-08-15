package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"polymetrics.ai/internal/durability"
)

type Locker interface {
	Lock() (func() error, error)
}

// CommitOutcome describes whether a state update definitely did not commit,
// committed, or may have committed despite a durability error.
type CommitOutcome uint8

const (
	CommitOutcomeNotCommitted CommitOutcome = iota
	CommitOutcomeCommitted
	CommitOutcomeIndeterminate
)

// MayHaveCommitted reports whether callers must preserve in-memory state and
// avoid replaying an operation that may already be durable.
func (o CommitOutcome) MayHaveCommitted() bool {
	return o == CommitOutcomeCommitted || o == CommitOutcomeIndeterminate
}

// String returns the stable diagnostic name for o.
func (o CommitOutcome) String() string {
	switch o {
	case CommitOutcomeCommitted:
		return "committed"
	case CommitOutcomeIndeterminate:
		return "indeterminate"
	default:
		return "not_committed"
	}
}

// CommitOutcomeError preserves the commit outcome when a post-rename
// durability or unlock error makes the final on-disk state uncertain.
type CommitOutcomeError struct {
	Outcome CommitOutcome
	Err     error
}

func (e *CommitOutcomeError) Error() string {
	if e == nil || e.Err == nil {
		return "state commit outcome is unavailable"
	}
	return fmt.Sprintf("state commit %s: %v", e.Outcome, e.Err)
}

func (e *CommitOutcomeError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// CommitOutcomeForError extracts a wrapped commit outcome, or returns
// CommitOutcomeNotCommitted when err carries no such outcome.
func CommitOutcomeForError(err error) CommitOutcome {
	var outcome *CommitOutcomeError
	if errors.As(err, &outcome) {
		return outcome.Outcome
	}
	return CommitOutcomeNotCommitted
}

// JSONStore persists a single JSON value at Path.
type JSONStore[T any] struct {
	Path          string
	Initial       func() T
	Locker        Locker
	Redact        func(path []string, value any) any
	SyncDirectory func(string) error
}

func (s JSONStore[T]) Load() (out T, err error) {
	unlock, err := s.lock()
	if err != nil {
		return out, err
	}
	defer func() { finishUnlock(unlock, &err, CommitOutcomeNotCommitted) }()

	return s.loadNoLock()
}

func (s JSONStore[T]) LoadReadOnly() (out T, err error) {
	return s.loadNoLock()
}

func (s JSONStore[T]) Save(value T) (err error) {
	unlock, err := s.lock()
	if err != nil {
		return err
	}
	outcome := CommitOutcomeNotCommitted
	defer func() { finishUnlock(unlock, &err, outcome) }()

	if err := s.saveNoLock(value); err != nil {
		return err
	}
	outcome = CommitOutcomeCommitted
	return nil
}

func (s JSONStore[T]) Update(update func(T) (T, error)) (out T, err error) {
	if update == nil {
		return out, errors.New("state update function is required")
	}
	unlock, err := s.lock()
	if err != nil {
		return out, err
	}
	outcome := CommitOutcomeNotCommitted
	defer func() { finishUnlock(unlock, &err, outcome) }()

	current, err := s.loadNoLock()
	if err != nil {
		return current, err
	}
	next, err := update(current)
	if err != nil {
		return current, err
	}
	if err := s.saveNoLock(next); err != nil {
		return next, err
	}
	outcome = CommitOutcomeCommitted
	return next, nil
}

func (s JSONStore[T]) UpdateAfterPreflight(preflight func(T) error, update func(T) (T, error)) (out T, err error) {
	if preflight == nil {
		return out, errors.New("state preflight function is required")
	}
	if update == nil {
		return out, errors.New("state update function is required")
	}

	current, err := s.LoadReadOnly()
	if err != nil {
		return current, err
	}
	if err := preflight(current); err != nil {
		return current, err
	}

	unlock, err := s.lock()
	if err != nil {
		return out, err
	}
	outcome := CommitOutcomeNotCommitted
	defer func() { finishUnlock(unlock, &err, outcome) }()

	current, err = s.loadNoLock()
	if err != nil {
		return current, err
	}
	if err := preflight(current); err != nil {
		return current, err
	}
	next, err := update(current)
	if err != nil {
		return current, err
	}
	if err := s.saveNoLock(next); err != nil {
		return next, err
	}
	outcome = CommitOutcomeCommitted
	return next, nil
}

func (s JSONStore[T]) RedactedSnapshot() (any, error) {
	value, err := s.Load()
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal state snapshot: %w", err)
	}
	var snapshot any
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("decode state snapshot: %w", err)
	}
	if s.Redact == nil {
		return snapshot, nil
	}
	return redactValue(nil, snapshot, s.Redact), nil
}

func (s JSONStore[T]) loadNoLock() (out T, err error) {
	if s.Path == "" {
		return out, errors.New("state path is required")
	}
	data, err := os.ReadFile(s.Path)
	if errors.Is(err, os.ErrNotExist) {
		if s.Initial != nil {
			return s.Initial(), nil
		}
		return out, nil
	}
	if err != nil {
		return out, fmt.Errorf("read state: %w", err)
	}
	if len(data) == 0 {
		if s.Initial != nil {
			return s.Initial(), nil
		}
		return out, nil
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return out, fmt.Errorf("decode state: %w", err)
	}
	return out, nil
}

func (s JSONStore[T]) saveNoLock(value T) (err error) {
	if s.Path == "" {
		return errors.New("state path is required")
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(s.Path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(s.Path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temporary state file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() {
		if tmp != nil {
			_ = tmp.Close()
		}
		if err != nil {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("write temporary state file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("sync temporary state file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temporary state file: %w", err)
	}
	tmp = nil
	if err := os.Rename(tmpPath, s.Path); err != nil {
		return fmt.Errorf("replace state file: %w", err)
	}
	syncDirectory := s.SyncDirectory
	if syncDirectory == nil {
		syncDirectory = durability.SyncDirectory
	}
	if err := syncDirectory(dir); err != nil {
		return &CommitOutcomeError{
			Outcome: CommitOutcomeIndeterminate,
			Err:     fmt.Errorf("sync state directory: %w", err),
		}
	}
	return nil
}

func (s JSONStore[T]) lock() (func() error, error) {
	if s.Locker == nil {
		return nil, nil
	}
	unlock, err := s.Locker.Lock()
	if err != nil {
		return nil, fmt.Errorf("lock state: %w", err)
	}
	return unlock, nil
}

func finishUnlock(unlock func() error, err *error, outcome CommitOutcome) {
	if unlock == nil {
		return
	}
	if unlockErr := unlock(); *err == nil && unlockErr != nil {
		unlockErr = fmt.Errorf("unlock state: %w", unlockErr)
		if outcome.MayHaveCommitted() {
			*err = &CommitOutcomeError{Outcome: outcome, Err: unlockErr}
			return
		}
		*err = unlockErr
	}
}

func redactValue(path []string, value any, redact func([]string, any) any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, child := range typed {
			out[key] = redactValue(appendPath(path, key), child, redact)
		}
		return redact(copyPath(path), out)
	case []any:
		out := make([]any, len(typed))
		for i, child := range typed {
			out[i] = redactValue(appendPath(path, strconv.Itoa(i)), child, redact)
		}
		return redact(copyPath(path), out)
	default:
		return redact(copyPath(path), value)
	}
}

func appendPath(path []string, next string) []string {
	out := make([]string, len(path)+1)
	copy(out, path)
	out[len(path)] = next
	return out
}

func copyPath(path []string) []string {
	return append([]string(nil), path...)
}
