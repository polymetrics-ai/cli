package dbtest

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
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
	name           string
	run            MachineRunner
	defaultBefore  string
	defaultAfter   string
	restoreDefault bool

	removeOnce sync.Once
	removeErr  error
}

// NewMachine creates and starts one uniquely named Podman machine, recording
// this process as its owner. The caller must call Remove in a defer: the
// machine is this run's to delete and nothing else will collect it.
//
// Every failure after the name is generated still returns the machine, and the
// ownership record is taken before the first command that can create state.
// `podman machine init` writes the VM config and a multi-GiB disk image before
// it can fail, and a cancelled context SIGKILLs it mid-write, so a claim taken
// only on success would leave a whole disk image on the host with nothing
// holding the authority to delete it. This mirrors the container path, which
// claims each resource before issuing the command that creates it.
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

	defaultBefore, err := podmanDefaultConnection(ctx, config.RunMachine)
	if err != nil {
		return nil, fmt.Errorf("inspect global Podman default connection: %w", err)
	}
	machine := &Machine{name: name, run: config.RunMachine, defaultBefore: defaultBefore}
	registerOwnedMachine(name)
	registerInterruptMachine(machine)
	if _, err := config.RunMachine.Run(ctx,
		"machine", "init",
		"--update-connection=false",
		"--cpus", strconv.Itoa(config.CPUs),
		"--memory", strconv.Itoa(config.MemoryMiB),
		"--disk-size", strconv.Itoa(config.DiskGiB),
		name,
	); err != nil {
		return machine, errors.Join(fmt.Errorf("create database test machine: %w", err), machine.captureDefaultAndRemove())
	}
	if err := machine.captureDefaultAfter(ctx); err != nil {
		return machine, errors.Join(fmt.Errorf("inspect global Podman default connection after machine init: %w", err), machine.captureDefaultAndRemove())
	}
	if _, err := config.RunMachine.Run(ctx, "machine", "start", name); err != nil {
		return machine, errors.Join(fmt.Errorf("start database test machine: %w", err), machine.removeDetached())
	}
	return machine, nil
}

func podmanDefaultConnection(ctx context.Context, run MachineRunner) (string, error) {
	raw, err := run.Run(ctx, "system", "connection", "list", "--format", "{{.Name}}\t{{.Default}}")
	if err != nil {
		return "", err
	}
	return parsePodmanDefaultConnection(raw)
}

func parsePodmanDefaultConnection(raw string) (string, error) {
	defaultName := ""
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		name, value, ok := strings.Cut(line, "\t")
		if !ok || strings.TrimSpace(name) == "" {
			return "", errors.New("podman returned an invalid connection listing")
		}
		switch strings.TrimSpace(value) {
		case "false":
		case "true":
			if defaultName != "" {
				return "", errors.New("podman returned multiple default connections")
			}
			defaultName = strings.TrimSpace(name)
		default:
			return "", errors.New("podman returned an invalid connection default value")
		}
	}
	return defaultName, nil
}

func (m *Machine) captureDefaultAfter(ctx context.Context) error {
	after, err := podmanDefaultConnection(ctx, m.run)
	if err != nil {
		return err
	}
	m.defaultAfter = after
	m.restoreDefault = after != m.defaultBefore && (after == m.name || after == m.name+"-root")
	return nil
}

func (m *Machine) captureDefaultAndRemove() error {
	ctx, cancel := context.WithTimeout(context.Background(), machineRemoveTimeout)
	defer cancel()
	var errs []error
	if err := m.captureDefaultAfter(ctx); err != nil {
		errs = append(errs, fmt.Errorf("inspect global Podman default connection after machine init: %w", err))
	}
	if err := m.Remove(ctx); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// machineRemoveTimeout bounds the detached teardown. Removing a machine is a
// handful of local file operations once the VM is stopped.
const machineRemoveTimeout = 5 * time.Minute

// removeDetached tears the machine down on a context of its own, because the
// caller's context may be exactly what killed the command whose partial state
// needs removing. It targets only this run's generated name.
func (m *Machine) removeDetached() error {
	ctx, cancel := context.WithTimeout(context.Background(), machineRemoveTimeout)
	defer cancel()
	return m.Remove(ctx)
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
		var errs []error
		verifyDefaultAfterRemoval, err := m.restoreGlobalDefault(ctx)
		if err != nil {
			errs = append(errs, err)
		}
		// A machine that is already stopped makes this fail; the removal below
		// is what has to succeed, and it stops a running machine itself.
		_, _ = m.run.Run(ctx, "machine", "stop", m.name)
		// A machine `init` never got as far as creating is absent, not leaked,
		// so its absence is the outcome this wanted rather than a failure.
		if _, err := m.run.Run(ctx, "machine", "rm", "--force", m.name); err != nil && !errors.Is(err, errPodmanResourceNotFound) {
			errs = append(errs, fmt.Errorf("remove database test machine: %w", err))
		} else if verifyDefaultAfterRemoval {
			current, inspectErr := podmanDefaultConnection(ctx, m.run)
			if inspectErr != nil {
				errs = append(errs, fmt.Errorf("inspect global Podman default connection after machine removal: %w", inspectErr))
			} else if current != "" {
				errs = append(errs, errors.New("global Podman default connection changed from its pre-run unset state"))
			}
		}
		m.removeErr = errors.Join(errs...)
		unregisterOwnedMachine(m.name)
		unregisterInterruptMachine(m)
	})
	return m.removeErr
}

func (m *Machine) restoreGlobalDefault(ctx context.Context) (bool, error) {
	if !m.restoreDefault {
		return false, nil
	}
	current, err := podmanDefaultConnection(ctx, m.run)
	if err != nil {
		return false, fmt.Errorf("inspect global Podman default connection before machine removal: %w", err)
	}
	if current != m.defaultAfter {
		return false, nil
	}
	if m.defaultBefore == "" {
		return true, nil
	}
	if _, err := m.run.Run(ctx, "system", "connection", "default", m.defaultBefore); err != nil {
		return false, fmt.Errorf("restore global Podman default connection: %w", err)
	}
	return false, nil
}
