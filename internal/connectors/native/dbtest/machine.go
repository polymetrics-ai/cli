package dbtest

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// ownedMachines records the Podman machines this process created. It is the
// only ownership evidence the host-disk reclaim accepts.
//
// A name match against caller-supplied configuration is not that evidence: it
// says nothing about who created the VM or what else runs on it, and
// `fstrim -av` reaches every filesystem on the machine, not just this run's
// containers. Creating the machine here is the one thing that does establish
// exclusive use, so a caller-supplied, pre-existing, shared, or remote machine
// is never trimmed — its unreclaimed bytes are reported instead.
var ownedMachines = struct {
	mu    sync.Mutex
	names map[string]struct{}
}{names: make(map[string]struct{})}

func registerOwnedMachine(name string) {
	ownedMachines.mu.Lock()
	defer ownedMachines.mu.Unlock()
	ownedMachines.names[name] = struct{}{}
}

func unregisterOwnedMachine(name string) {
	ownedMachines.mu.Lock()
	defer ownedMachines.mu.Unlock()
	delete(ownedMachines.names, name)
}

func machineIsOwned(name string) bool {
	ownedMachines.mu.Lock()
	defer ownedMachines.mu.Unlock()
	_, owned := ownedMachines.names[name]
	return owned
}

const (
	// machineNamePrefix marks a machine as this package's own, and is short
	// because Podman bounds the whole name.
	machineNamePrefix = "pmdb-"
	// maxPodmanMachineNameLength is Podman's own limit. Exceeding it fails at
	// `machine init` with an error this package would have to discard, so the
	// name is checked here instead.
	maxPodmanMachineNameLength = 30
)

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}

// MachineConfig describes the one ephemeral Podman machine a run may create
// for itself. Sizing is explicit because a database container's peak disk is
// the binding constraint on a VM-backed host.
type MachineConfig struct {
	// Engine names the database engine the machine is created for, used only
	// in the generated machine name.
	Engine string
	// CPUs, MemoryMiB, and DiskGiB size the machine. Each has a small default
	// sufficient for one database engine.
	CPUs      int
	MemoryMiB int
	DiskGiB   int

	RunMachine MachineRunner
}

// Machine is one Podman machine this process created and therefore may trim
// and remove. There is no way to construct one around a machine that already
// existed, which is what keeps the ownership record honest.
type Machine struct {
	name string
	run  MachineRunner

	removeOnce sync.Once
	removeErr  error
}

// NewMachine creates and starts one uniquely named Podman machine, recording
// this process as its owner. The caller must call Remove in a defer: the
// machine is this run's to delete and nothing else will collect it.
//
// A failed start still returns the machine, because `machine init` may have
// created the VM before the start failed; removing it is then the caller's
// only way to avoid leaking a whole disk image.
func NewMachine(ctx context.Context, config MachineConfig) (*Machine, error) {
	if strings.TrimSpace(config.Engine) == "" {
		return nil, errors.New("database test machine requires an engine name")
	}
	if config.RunMachine == nil {
		config.RunMachine = podmanMachineRunner{}
	}
	if config.CPUs < 1 || config.CPUs > 8 {
		config.CPUs = 2
	}
	if config.MemoryMiB < 1024 || config.MemoryMiB > 16384 {
		config.MemoryMiB = 2048
	}
	if config.DiskGiB < 8 || config.DiskGiB > 128 {
		config.DiskGiB = 16
	}
	suffix, err := randomSuffix()
	if err != nil {
		return nil, fmt.Errorf("generate database test machine name: %w", err)
	}
	// Podman refuses a machine name over 30 characters, so this name is built
	// to a shorter budget than the container names: a truncated engine slug
	// and half the random suffix, which still leaves 48 bits of uniqueness.
	name := machineNamePrefix + truncate(engineSlug(config.Engine), 8) + "-" + suffix[:12]
	if !safeName(name) || len(name) > maxPodmanMachineNameLength {
		return nil, errors.New("database test machine name is unsafe or too long for podman")
	}

	machine := &Machine{name: name, run: config.RunMachine}
	if _, err := config.RunMachine.Run(ctx,
		"machine", "init",
		"--cpus", strconv.Itoa(config.CPUs),
		"--memory", strconv.Itoa(config.MemoryMiB),
		"--disk-size", strconv.Itoa(config.DiskGiB),
		name,
	); err != nil {
		return nil, fmt.Errorf("create database test machine: %w", err)
	}
	// Claimed only after init reported success, and before start, so an
	// interrupt during the start still finds a machine it is allowed to remove.
	registerOwnedMachine(name)
	registerInterruptMachine(machine)
	if _, err := config.RunMachine.Run(ctx, "machine", "start", name); err != nil {
		return machine, fmt.Errorf("start database test machine: %w", err)
	}
	return machine, nil
}

// Connection is the Podman connection name that addresses this machine. Podman
// names a machine's rootless connection after the machine itself.
func (m *Machine) Connection() string { return m.name }

// Name is the generated Podman machine name.
func (m *Machine) Name() string { return m.name }

// Remove stops and deletes only this machine, then drops its ownership record
// so nothing can trim a name this process no longer owns. It is idempotent.
func (m *Machine) Remove(ctx context.Context) error {
	m.removeOnce.Do(func() {
		// A machine that is already stopped makes this fail; the removal below
		// is what has to succeed, and it stops a running machine itself.
		_, _ = m.run.Run(ctx, "machine", "stop", m.name)
		if _, err := m.run.Run(ctx, "machine", "rm", "--force", m.name); err != nil {
			m.removeErr = fmt.Errorf("remove database test machine: %w", err)
		}
		unregisterOwnedMachine(m.name)
		unregisterInterruptMachine(m)
	})
	return m.removeErr
}
