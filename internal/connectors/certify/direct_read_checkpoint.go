package certify

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const directReadCheckpointSchemaVersion = 1

// directReadCheckpoint persists only declaration-owned names and pass state.
// It deliberately contains no credential, provider response, request URL, or
// rate-limit scope. A changed candidate set or non-secret configuration has a
// different fingerprint and cannot silently reuse a previous run's passes.
type directReadCheckpoint struct {
	SchemaVersion int             `json:"schema_version"`
	Fingerprint   string          `json:"fingerprint"`
	Completed     map[string]bool `json:"completed"`
}

type directReadCheckpointInput struct {
	Connector  string                `json:"connector"`
	Candidates []directReadCandidate `json:"candidates"`
	Config     map[string]string     `json:"config"`
}

func newDirectReadCheckpoint(connector string, candidates []directReadCandidate, config map[string]string) (directReadCheckpoint, error) {
	payload, err := json.Marshal(directReadCheckpointInput{
		Connector:  connector,
		Candidates: candidates,
		Config:     config,
	})
	if err != nil {
		return directReadCheckpoint{}, fmt.Errorf("marshal direct-read checkpoint fingerprint: %w", err)
	}
	sum := sha256.Sum256(payload)
	return directReadCheckpoint{
		SchemaVersion: directReadCheckpointSchemaVersion,
		Fingerprint:   hex.EncodeToString(sum[:]),
		Completed:     make(map[string]bool, len(candidates)),
	}, nil
}

func loadDirectReadCheckpoint(path, connector string, candidates []directReadCandidate, config map[string]string) (directReadCheckpoint, error) {
	expected, err := newDirectReadCheckpoint(connector, candidates, config)
	if err != nil {
		return directReadCheckpoint{}, err
	}
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return expected, nil
	}
	if err != nil {
		return directReadCheckpoint{}, fmt.Errorf("read direct-read checkpoint: %w", err)
	}
	var checkpoint directReadCheckpoint
	if err := json.Unmarshal(raw, &checkpoint); err != nil {
		return directReadCheckpoint{}, fmt.Errorf("decode direct-read checkpoint: %w", err)
	}
	if checkpoint.SchemaVersion != directReadCheckpointSchemaVersion || checkpoint.Fingerprint != expected.Fingerprint {
		return directReadCheckpoint{}, errors.New("direct-read resume checkpoint does not match this declaration or configuration")
	}
	if checkpoint.Completed == nil {
		checkpoint.Completed = make(map[string]bool)
	}
	return checkpoint, nil
}

func saveDirectReadCheckpoint(path string, checkpoint directReadCheckpoint) error {
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create direct-read checkpoint directory: %w", err)
	}
	raw, err := json.Marshal(checkpoint)
	if err != nil {
		return fmt.Errorf("encode direct-read checkpoint: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".direct-read-checkpoint-*")
	if err != nil {
		return fmt.Errorf("create direct-read checkpoint: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("protect direct-read checkpoint: %w", err)
	}
	if _, err := temporary.Write(raw); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write direct-read checkpoint: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close direct-read checkpoint: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace direct-read checkpoint: %w", err)
	}
	return nil
}
