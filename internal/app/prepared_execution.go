package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var preparedExecutionIdentityPattern = regexp.MustCompile(`^pex_[0-9a-f]{64}$`)

// PreparedExecutionRefusedError reports that process-private prepared evidence
// changed before dispatch. The identity is safe opaque evidence, never an
// approval token or credential.
type PreparedExecutionRefusedError struct {
	Identity string
	Reason   string
}

func (e *PreparedExecutionRefusedError) Error() string {
	if e == nil {
		return "prepared execution refused"
	}
	return fmt.Sprintf("prepared execution %q refused: %s", e.Identity, e.Reason)
}

// PreparedExecutionReplayError refuses a concurrent, completed, or ambiguous
// replay of the same prepared payload identity.
type PreparedExecutionReplayError struct {
	Identity string
}

func (e *PreparedExecutionReplayError) Error() string {
	if e == nil || e.Identity == "" {
		return "prepared execution already started"
	}
	return fmt.Sprintf("prepared execution %q already started", e.Identity)
}

type preparedExecutionLease struct {
	path string
}

func (a *App) acquirePreparedExecutionLease(identity string) (*preparedExecutionLease, error) {
	if a == nil || a.projectDir == "" {
		return nil, errors.New("prepared execution project is required")
	}
	if !preparedExecutionIdentityPattern.MatchString(identity) {
		return nil, &PreparedExecutionRefusedError{Identity: identity, Reason: "invalid_identity"}
	}
	dir := filepath.Join(a.projectDir, "prepared-executions")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create prepared execution directory: %w", err)
	}
	path := filepath.Join(dir, identity+".lock")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil, &PreparedExecutionReplayError{Identity: identity}
		}
		return nil, fmt.Errorf("acquire prepared execution lease: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return nil, fmt.Errorf("sync prepared execution lease: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return nil, fmt.Errorf("close prepared execution lease: %w", err)
	}
	if err := syncPreparedExecutionDir(dir); err != nil {
		_ = os.Remove(path)
		return nil, err
	}
	return &preparedExecutionLease{path: path}, nil
}

func (l *preparedExecutionLease) Release() error {
	if l == nil || l.path == "" {
		return nil
	}
	if err := os.Remove(l.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return syncPreparedExecutionDir(filepath.Dir(l.path))
}

func syncPreparedExecutionDir(dir string) error {
	directory, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("open prepared execution directory: %w", err)
	}
	defer directory.Close()
	if err := directory.Sync(); err != nil {
		return fmt.Errorf("sync prepared execution directory: %w", err)
	}
	return nil
}
