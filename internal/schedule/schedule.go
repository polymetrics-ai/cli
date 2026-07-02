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

var validName = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)

func validateName(name string) error {
	if name == "" {
		return errors.New("schedule name must not be empty")
	}
	if !validName.MatchString(name) {
		return fmt.Errorf("invalid schedule name %q: must match [a-z0-9][a-z0-9-]*, max 64 chars", name)
	}
	return nil
}

func schedulesDir(root string) string {
	return filepath.Join(root, "schedules")
}

func manifestPath(root, name string) string {
	return filepath.Join(schedulesDir(root), name+".json")
}

// Save writes a manifest to <root>/schedules/<name>.json.
func Save(root string, m Manifest, allowOverwrite bool) error {
	if err := validateName(m.Name); err != nil {
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
	return os.WriteFile(path, data, 0o644)
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
	return m, nil
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
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
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
	err := os.Remove(manifestPath(root, name))
	if err != nil {
		return err
	}
	return nil
}
