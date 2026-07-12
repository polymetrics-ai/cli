package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type Locker interface {
	Lock() (func() error, error)
}

// JSONStore persists a single JSON value at Path.
type JSONStore[T any] struct {
	Path    string
	Initial func() T
	Locker  Locker
	Redact  func(path []string, value any) any
}

func (s JSONStore[T]) Load() (out T, err error) {
	unlock, err := s.lock()
	if err != nil {
		return out, err
	}
	defer finishUnlock(unlock, &err)

	return s.loadNoLock()
}

func (s JSONStore[T]) Save(value T) (err error) {
	unlock, err := s.lock()
	if err != nil {
		return err
	}
	defer finishUnlock(unlock, &err)

	return s.saveNoLock(value)
}

func (s JSONStore[T]) Update(update func(T) (T, error)) (out T, err error) {
	if update == nil {
		return out, errors.New("state update function is required")
	}
	unlock, err := s.lock()
	if err != nil {
		return out, err
	}
	defer finishUnlock(unlock, &err)

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
	_ = syncDir(dir)
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

func finishUnlock(unlock func() error, err *error) {
	if unlock == nil {
		return
	}
	if unlockErr := unlock(); *err == nil && unlockErr != nil {
		*err = fmt.Errorf("unlock state: %w", unlockErr)
	}
}

func syncDir(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = dir.Close() }()
	return dir.Sync()
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
