// Package schedule manages pm schedule manifests and backend installation.
package schedule

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Manifest is a persisted schedule definition.
type Manifest struct {
	Name      string    `json:"name"`
	Cron      string    `json:"cron"`
	Flow      string    `json:"flow"`
	Root      string    `json:"root,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BackendKind identifies which scheduler backend is active.
type BackendKind string

const (
	KindLaunchd  BackendKind = "launchd"
	KindSystemd  BackendKind = "systemd"
	KindCrontab  BackendKind = "crontab"
	KindTemporal BackendKind = "temporal"
)

// Backend is the scheduler backend interface.
type Backend interface {
	Install(ctx context.Context, m Manifest, pmBin string) error
	Remove(ctx context.Context, name string) error
	Kind() BackendKind
}

var (
	validName          = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)
	validFlowReference = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`)
)

type FlowReferenceReason string

const (
	FlowReferenceMissing   FlowReferenceReason = "missing"
	FlowReferenceMalformed FlowReferenceReason = "malformed"
	FlowReferenceAmbiguous FlowReferenceReason = "ambiguous"
	FlowReferenceInvalid   FlowReferenceReason = "invalid"
)

// FlowReferenceError names the exact flow a schedule could not positively
// resolve before any manifest or scheduler backend write.
type FlowReferenceError struct {
	Flow   string
	Reason FlowReferenceReason
	Err    error
}

func (e *FlowReferenceError) Error() string {
	if e == nil {
		return "schedule flow reference refused"
	}
	message := fmt.Sprintf("schedule flow %q is %s", e.Flow, e.Reason)
	if e.Err != nil {
		message += ": " + e.Err.Error()
	}
	return message
}

func (e *FlowReferenceError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func validateName(name string) error {
	if name == "" {
		return errors.New("schedule name must not be empty")
	}
	if !validName.MatchString(name) {
		return fmt.Errorf("invalid schedule name %q: must match [a-z0-9][a-z0-9-]*, max 64 chars", name)
	}
	return nil
}

func validateFlowReference(reference string) error {
	if !validFlowReference.MatchString(reference) {
		return &FlowReferenceError{Flow: reference, Reason: FlowReferenceMalformed}
	}
	return nil
}

func schedulesDir(root string) string {
	return filepath.Join(root, "schedules")
}

func manifestPath(root, name string) string {
	return filepath.Join(schedulesDir(root), name+".json")
}

func fireStatePath(root, name string) string {
	return filepath.Join(schedulesDir(root), name+".fire.json")
}

func fireLockPath(root, name string) string {
	return filepath.Join(schedulesDir(root), name+".fire.lock")
}

// Save writes a manifest to <root>/schedules/<name>.json.
func Save(root string, m Manifest, allowOverwrite bool) error {
	if err := validateName(m.Name); err != nil {
		return err
	}
	if err := validateFlowReference(m.Flow); err != nil {
		return err
	}
	dir := schedulesDir(root)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("schedule: mkdir: %w", err)
	}
	path := manifestPath(root, m.Name)
	if !allowOverwrite {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("schedule %q already exists", m.Name)
		}
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("schedule: marshal: %w", err)
	}
	return writeFileAtomic(path, data, 0o600)
}

// Load reads a manifest from <root>/schedules/<name>.json.
func Load(root, name string) (Manifest, error) {
	data, err := os.ReadFile(manifestPath(root, name))
	if err != nil {
		return Manifest{}, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return Manifest{}, fmt.Errorf("schedule: unmarshal: %w", err)
	}
	if err := validateName(m.Name); err != nil {
		return Manifest{}, fmt.Errorf("schedule: invalid name: %w", err)
	}
	if err := validateFlowReference(m.Flow); err != nil {
		return Manifest{}, err
	}
	return m, nil
}

// FindByFlow returns the sole schedule that owns a named flow. A direct
// installed `flow run` cannot safely guess between multiple schedules, so an
// ambiguous manually-authored inventory is refused before execution.
func FindByFlow(root, flow string) (Manifest, bool, error) {
	if err := validateFlowReference(flow); err != nil {
		return Manifest{}, false, err
	}
	manifests, err := List(root)
	if err != nil {
		return Manifest{}, false, err
	}
	var found Manifest
	count := 0
	for _, manifest := range manifests {
		if manifest.Flow != flow {
			continue
		}
		found = manifest
		count++
	}
	if count > 1 {
		return Manifest{}, false, &FlowReferenceError{Flow: flow, Reason: FlowReferenceAmbiguous}
	}
	return found, count == 1, nil
}

// List returns all manifests under <root>/schedules/.
func List(root string) ([]Manifest, error) {
	dir := schedulesDir(root)
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("schedule: readdir: %w", err)
	}
	var manifests []Manifest
	for _, e := range entries {
		// Fire state is stored next to the manifest. It is intentionally not a
		// schedule definition and must never appear as a malformed second entry
		// after the first firing.
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" || strings.HasSuffix(e.Name(), ".fire.json") {
			continue
		}
		name := e.Name()[:len(e.Name())-5]
		m, err := Load(root, name)
		if err != nil {
			return nil, err
		}
		manifests = append(manifests, m)
	}
	return manifests, nil
}

// Delete removes the manifest file for name.
func Delete(root, name string) error {
	if _, err := os.Stat(fireLockPath(root, name)); err == nil {
		return ErrFireInProgress
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	err := os.Remove(manifestPath(root, name))
	if err != nil {
		return err
	}
	if err := os.Remove(fireStatePath(root, name)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
