package flow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// CheckpointStore persists step completion status across runs.
type CheckpointStore interface {
	Get(flowName, stepID string) (string, error)
	Set(flowName, stepID, status string) error
	Clear(flowName string) error
}

// FileCheckpointStore is a JSON-file-backed checkpoint store.
// The backing file lives at <Dir>/flow-checkpoints.json.
// The JSON structure is: { "<flowName>/<stepID>": "<status>" }.
type FileCheckpointStore struct {
	Dir string
	mu  sync.Mutex
}

func (s *FileCheckpointStore) path() string {
	return filepath.Join(s.Dir, "flow-checkpoints.json")
}

func (s *FileCheckpointStore) load() (map[string]string, error) {
	data, err := os.ReadFile(s.path())
	if os.IsNotExist(err) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, err
	}
	m := map[string]string{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *FileCheckpointStore) save(m map[string]string) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path(), data, 0o600)
}

func key(flowName, stepID string) string { return flowName + "/" + stepID }

func (s *FileCheckpointStore) Get(flowName, stepID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, err := s.load()
	if err != nil {
		return "", err
	}
	return m[key(flowName, stepID)], nil
}

func (s *FileCheckpointStore) Set(flowName, stepID, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, err := s.load()
	if err != nil {
		return err
	}
	m[key(flowName, stepID)] = status
	return s.save(m)
}

func (s *FileCheckpointStore) Clear(flowName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, err := s.load()
	if err != nil {
		return err
	}
	prefix := flowName + "/"
	for k := range m {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(m, k)
		}
	}
	return s.save(m)
}
