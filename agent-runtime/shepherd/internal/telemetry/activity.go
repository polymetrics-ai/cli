package telemetry

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const segmentName = "activity-000001.jsonl"

var allowedKinds = map[string]struct{}{
	"run.started": {}, "run.terminal": {}, "unit.started": {}, "unit.terminal": {},
	"tool.started": {}, "tool.terminal": {}, "validation": {}, "transition": {},
	"heartbeat": {}, "cost": {}, "model.activity": {}, "human.decision": {},
	"decision": {},
}

type Activity struct {
	ID           string    `json:"event_id"`
	RunID        string    `json:"run_id"`
	UnitID       string    `json:"unit_id,omitempty"`
	Kind         string    `json:"kind"`
	Status       string    `json:"status,omitempty"`
	Model        string    `json:"model,omitempty"`
	Tool         string    `json:"tool,omitempty"`
	At           time.Time `json:"at"`
	DurationMS   int64     `json:"duration_ms,omitempty"`
	InputTokens  int64     `json:"input_tokens,omitempty"`
	OutputTokens int64     `json:"output_tokens,omitempty"`
	CostMicros   int64     `json:"cost_micros,omitempty"`
	Detail       string    `json:"-"`
}

type Store struct {
	mu   sync.Mutex
	file *os.File
	seen map[string]struct{}
}

func Open(_ context.Context, directory string) (*Store, error) {
	if directory == "" {
		return nil, errors.New("activity directory is required")
	}
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, fmt.Errorf("create activity directory: %w", err)
	}
	if err := os.Chmod(directory, 0o700); err != nil {
		return nil, fmt.Errorf("secure activity directory: %w", err)
	}
	path := filepath.Join(directory, segmentName)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open activity segment: %w", err)
	}
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		return nil, err
	}
	store := &Store{file: file, seen: make(map[string]struct{})}
	if err := store.recoverAndIndex(); err != nil {
		_ = file.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.file.Close()
}

func (s *Store) Append(_ context.Context, activity Activity) (bool, error) {
	if err := validate(activity); err != nil {
		return false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.seen[activity.ID]; exists {
		return false, nil
	}
	raw, err := json.Marshal(activity)
	if err != nil {
		return false, err
	}
	if _, err := s.file.Write(append(raw, '\n')); err != nil {
		return false, fmt.Errorf("append activity: %w", err)
	}
	if err := s.file.Sync(); err != nil {
		return false, fmt.Errorf("sync activity: %w", err)
	}
	s.seen[activity.ID] = struct{}{}
	return true, nil
}

func (s *Store) recoverAndIndex() error {
	raw, err := io.ReadAll(s.file)
	if err != nil {
		return err
	}
	if len(raw) > 0 && raw[len(raw)-1] != '\n' {
		last := bytes.LastIndexByte(raw, '\n')
		if last < 0 {
			raw = nil
		} else {
			raw = raw[:last+1]
		}
		if err := s.file.Truncate(int64(len(raw))); err != nil {
			return fmt.Errorf("truncate torn activity tail: %w", err)
		}
	}
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 4096), 256*1024)
	for scanner.Scan() {
		var activity Activity
		if err := json.Unmarshal(scanner.Bytes(), &activity); err != nil {
			return fmt.Errorf("decode durable activity: %w", err)
		}
		if err := validate(activity); err != nil {
			return fmt.Errorf("invalid durable activity: %w", err)
		}
		s.seen[activity.ID] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	_, err = s.file.Seek(0, io.SeekEnd)
	return err
}

func validate(activity Activity) error {
	if activity.ID == "" || activity.RunID == "" || activity.At.IsZero() {
		return errors.New("event, run, and timestamp are required")
	}
	if activity.Detail != "" {
		return errors.New("free-form detail is forbidden")
	}
	if _, ok := allowedKinds[activity.Kind]; !ok {
		return fmt.Errorf("activity kind %q is not allowlisted", activity.Kind)
	}
	for label, value := range map[string]string{"id": activity.ID, "run": activity.RunID, "unit": activity.UnitID, "status": activity.Status, "model": activity.Model, "tool": activity.Tool} {
		if len(value) > 256 || strings.ContainsAny(value, "\r\n\x00") {
			return fmt.Errorf("%s field is unsafe", label)
		}
	}
	if activity.DurationMS < 0 || activity.InputTokens < 0 || activity.OutputTokens < 0 || activity.CostMicros < 0 {
		return errors.New("numeric activity fields must be non-negative")
	}
	return nil
}
