package schedule

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var (
	// ErrFireInProgress rejects a concurrent or interrupted fire. Both cases
	// remain halted until the existing owner records a terminal result; a
	// scheduler must never replay a potentially non-idempotent write.
	ErrFireInProgress = errors.New("schedule fire is already in progress")
	// ErrFireParked rejects another automatic firing after a terminal unsafe
	// outcome. An operator must create a fresh approved schedule instead.
	ErrFireParked = errors.New("schedule fire is parked")
)

// FireStatus is the safe state of the most recent schedule firing.
type FireStatus string

const (
	FireStatusReady     FireStatus = "ready"
	FireStatusRunning   FireStatus = "running"
	FireStatusSucceeded FireStatus = "succeeded"
	FireStatusParked    FireStatus = "parked"
)

// FireStopReason is a closed category, deliberately not a raw provider error.
type FireStopReason string

const (
	FireStopNone      FireStopReason = ""
	FireStopAmbiguous FireStopReason = "ambiguous_delivery"
	FireStopRateLimit FireStopReason = "rate_limited"
	FireStopCleanup   FireStopReason = "cleanup_failed"
	FireStopFailed    FireStopReason = "failed"
	FireStopScope     FireStopReason = "authorization_scope_changed"
	FireStopRevoked   FireStopReason = "authorization_revoked"
	FireStopExpired   FireStopReason = "authorization_expired"
)

// FireReceipt retains the terminal proof a scheduler may safely surface.
// It contains only opaque identifiers and result counts; payloads, provider
// messages, credentials, tokens, and raw configuration never cross this
// storage boundary.
type FireReceipt struct {
	FlowName               string         `json:"flow_name"`
	FlowStatus             string         `json:"flow_status"`
	AuthorizationReference string         `json:"authorization_reference"`
	ReceiptIDs             []string       `json:"receipt_ids,omitempty"`
	StopReason             FireStopReason `json:"stop_reason,omitempty"`
	StartedAt              time.Time      `json:"started_at"`
	CompletedAt            time.Time      `json:"completed_at,omitempty"`
}

// FireState is the persisted, schedule-owned status. Running and parked are
// deliberate non-replay states.
type FireState struct {
	Status   FireStatus  `json:"status"`
	LastFire FireReceipt `json:"last_fire,omitempty"`
}

// FireLease owns one scheduled execution. Completion and parking both release
// its lock only after durable terminal state has been written.
type FireLease struct {
	root     string
	manifest Manifest
}

// BeginFire acquires an exclusive schedule fire lease and records the running
// state before any caller can dispatch a flow.
func BeginFire(root, name string) (*FireLease, error) {
	manifest, err := Load(root, name)
	if err != nil {
		return nil, err
	}
	if manifest.AuthorizationReference == "" {
		return nil, errors.New("schedule authorization reference is required")
	}
	state, err := LoadFireState(root, name)
	if err != nil {
		return nil, err
	}
	if state.Status == FireStatusParked {
		return nil, ErrFireParked
	}
	// The persisted running state is the crash-recovery guard, independent of
	// the advisory lock's lifetime. A fire that died after any potentially
	// non-idempotent connector call must remain halted even if an operator or
	// filesystem cleanup removes its lock file.
	if state.Status == FireStatusRunning {
		return nil, ErrFireInProgress
	}

	lock, err := os.OpenFile(fireLockPath(root, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil, ErrFireInProgress
		}
		return nil, err
	}
	if _, err := fmt.Fprintf(lock, "%d\n", os.Getpid()); err != nil {
		_ = lock.Close()
		_ = os.Remove(fireLockPath(root, name))
		return nil, err
	}
	if err := lock.Close(); err != nil {
		_ = os.Remove(fireLockPath(root, name))
		return nil, err
	}

	state = FireState{Status: FireStatusRunning, LastFire: FireReceipt{
		FlowName: manifest.Flow, AuthorizationReference: manifest.AuthorizationReference, StartedAt: time.Now().UTC(),
	}}
	if err := saveFireState(root, name, state); err != nil {
		_ = os.Remove(fireLockPath(root, name))
		return nil, err
	}
	return &FireLease{root: root, manifest: manifest}, nil
}

// Complete records a successfully read-back flow result and then releases the
// lease. A cleanup failure parks the fire by retaining its lock.
func (l *FireLease) Complete(receipt FireReceipt) error {
	if l == nil {
		return errors.New("schedule fire lease is required")
	}
	receipt = l.normalizeReceipt(receipt)
	if receipt.FlowStatus == "" {
		receipt.FlowStatus = "success"
	}
	receipt.StopReason = FireStopNone
	state := FireState{Status: FireStatusSucceeded, LastFire: receipt}
	if err := saveFireState(l.root, l.manifest.Name, state); err != nil {
		return err
	}
	if err := os.Remove(fireLockPath(l.root, l.manifest.Name)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		state.Status = FireStatusParked
		state.LastFire.StopReason = FireStopCleanup
		_ = saveFireState(l.root, l.manifest.Name, state)
		return fmt.Errorf("release completed schedule fire: %w", err)
	}
	return nil
}

// Park records a fail-closed terminal result before releasing the lease. A
// failed cleanup leaves the lock in place, which is stricter than replay.
func (l *FireLease) Park(reason FireStopReason) error {
	if l == nil {
		return errors.New("schedule fire lease is required")
	}
	if reason == FireStopNone {
		return errors.New("schedule fire park reason is required")
	}
	state, err := LoadFireState(l.root, l.manifest.Name)
	if err != nil {
		return err
	}
	receipt := l.normalizeReceipt(state.LastFire)
	receipt.FlowStatus = "failed"
	receipt.StopReason = reason
	state = FireState{Status: FireStatusParked, LastFire: receipt}
	if err := saveFireState(l.root, l.manifest.Name, state); err != nil {
		return err
	}
	if err := os.Remove(fireLockPath(l.root, l.manifest.Name)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("release parked schedule fire: %w", err)
	}
	return nil
}

func (l *FireLease) normalizeReceipt(receipt FireReceipt) FireReceipt {
	if receipt.FlowName == "" {
		receipt.FlowName = l.manifest.Flow
	}
	receipt.AuthorizationReference = l.manifest.AuthorizationReference
	if receipt.StartedAt.IsZero() {
		receipt.StartedAt = time.Now().UTC()
	}
	receipt.StartedAt = receipt.StartedAt.UTC()
	if receipt.CompletedAt.IsZero() {
		receipt.CompletedAt = time.Now().UTC()
	}
	receipt.CompletedAt = receipt.CompletedAt.UTC()
	receipt.ReceiptIDs = append([]string(nil), receipt.ReceiptIDs...)
	return receipt
}

// LoadFireState reads the safe terminal state. A newly created schedule is
// ready; it has no prior receipt.
func LoadFireState(root, name string) (FireState, error) {
	data, err := os.ReadFile(fireStatePath(root, name))
	if errors.Is(err, os.ErrNotExist) {
		return FireState{Status: FireStatusReady}, nil
	}
	if err != nil {
		return FireState{}, err
	}
	var state FireState
	if err := json.Unmarshal(data, &state); err != nil {
		return FireState{}, fmt.Errorf("schedule: unmarshal fire state: %w", err)
	}
	if state.Status == "" {
		state.Status = FireStatusReady
	}
	state.LastFire.ReceiptIDs = append([]string(nil), state.LastFire.ReceiptIDs...)
	return state, nil
}

func saveFireState(root, name string, state FireState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("schedule: marshal fire state: %w", err)
	}
	return writeFileAtomic(fireStatePath(root, name), data, 0o600)
}

func writeFileAtomic(path string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
