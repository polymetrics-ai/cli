// Package dbtest provides an opt-in, Podman-backed database integration-test
// harness. It is test support only: callers supply one engine's pinned image
// and server arguments, while this package owns Podman connection scoping,
// ephemeral resource names, loopback port allocation, disk accounting, and
// unconditional cleanup.
//
// A second engine is added by supplying another Config, not by copying this
// file. See mysql_integration_test.go for the reference caller.
package dbtest

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const defaultPodmanBinary = "podman"

var errPodmanResourceNotFound = errors.New("podman resource not found")

// engineSlots serialises engines by default. Parallel database containers
// multiply peak disk and memory on a host where disk is the binding
// constraint, so concurrency is opt-in and bounded rather than implicit.
var engineSlots = make(chan struct{}, 1)

// SetMaxConcurrentEngines bounds how many harnesses may hold a container at
// once. It is the opt-in parallel mode; the default of one keeps engines
// sequential. It must be called before any Start and panics on a
// non-positive or unbounded value.
func SetMaxConcurrentEngines(n int) {
	if n < 1 || n > 8 {
		panic("dbtest: concurrent engine limit must be between 1 and 8")
	}
	engineSlots = make(chan struct{}, n)
}

// Endpoint is the loopback-only address allocated for an ephemeral engine.
// It deliberately represents host and port separately so callers never need
// to persist or log a connection string.
type Endpoint struct {
	Host string
	Port int
}

// CommandRunner runs a Podman subcommand against one explicit connection. It
// exists solely to make resource ownership and cleanup testable without a
// running container engine.
type CommandRunner interface {
	Run(context.Context, string, ...string) (string, error)
}

// MachineRunner runs a Podman machine lifecycle command. It is separate from
// the container runner so a unit test can prove the optional host-disk
// reclaim happens after container cleanup without requiring a VM.
type MachineRunner interface {
	Run(context.Context, ...string) error
}

// Config describes one supported database-engine container. It is internal
// test support, not a generic command surface. Adding an engine means adding
// a Config, not new code in this package.
type Config struct {
	// Engine names the database engine, used only for resource names and
	// error text.
	Engine string
	// Image is a pinned image reference. A ":latest" or untagged reference is
	// refused so a surprise upstream bump cannot change a run's result.
	Image string
	// ContainerPort is the port the engine listens on inside the container.
	// The host port is always dynamically allocated, never this one.
	ContainerPort int
	// DataVolumePath is the engine's data directory inside the container.
	DataVolumePath string
	// Connection names the Podman connection to target. It is mandatory:
	// the default connection on a shared host belongs to another lane, so
	// falling back to it is refused before any Podman call.
	Connection string
	// Machine names the Podman machine backing Connection. It is required
	// only for ReclaimHostDisk and defaults to Connection.
	Machine string
	// ContainerArgs are passed to "podman run" before the image reference.
	ContainerArgs []string
	// EngineArgs are passed to the engine process after the image reference.
	EngineArgs []string
	// KeepImage retains the pulled source image for a subsequent run. The
	// default removes it, because on a VM-backed host a retained image holds
	// its bytes in the machine's disk file.
	KeepImage bool
	// ReclaimHostDisk trims the backing machine after cleanup so freed guest
	// blocks are punched out of the host's sparse disk file. Without it,
	// removing an image frees space inside the VM but not on the host.
	ReclaimHostDisk bool

	Run        CommandRunner
	RunMachine MachineRunner
	DiskFree   func() (uint64, error)
}

// Report records disk free before startup and after all teardown work,
// including an optional host-disk reclaim. A caller can print only these
// aggregate byte values to make leaks visible without logging connection
// information.
type Report struct {
	DiskFreeBefore    uint64
	DiskFreeAfter     uint64
	HostDiskReclaimed bool
}

// Harness owns exactly one generated container and one generated named
// volume. It is not safe to copy after first use.
type Harness struct {
	config Config

	containerName string
	volumeName    string
	runImage      string

	mu               sync.Mutex
	endpoint         Endpoint
	runImageKnown    bool
	sourceImageKnown bool
	containerKnown   bool
	volumeKnown      bool
	slotHeld         bool
	report           Report

	closeOnce  sync.Once
	closeErr   error
	stopSignal chan struct{}
}

// New validates the test-only engine configuration. An explicit Podman
// connection is mandatory: the default connection on this host points at
// another lane's machine, so falling back to it is refused before any Podman
// call rather than failing later with an opaque socket error.
func New(config Config) (*Harness, error) {
	if strings.TrimSpace(config.Engine) == "" {
		return nil, errors.New("database test harness requires an engine name")
	}
	if !pinnedImage(config.Image) {
		return nil, errors.New("database test harness requires a safe pinned image tag")
	}
	if config.ContainerPort < 1 || config.ContainerPort > 65535 {
		return nil, errors.New("database test harness requires a valid container port")
	}
	if !filepath.IsAbs(config.DataVolumePath) || strings.ContainsAny(config.DataVolumePath, "\r\n:") {
		return nil, errors.New("database test harness requires a safe absolute engine data-volume path")
	}
	if !safeName(config.Connection) {
		return nil, errors.New("database test harness requires an explicit safe Podman connection")
	}
	if config.Machine == "" {
		config.Machine = config.Connection
	}
	if !safeName(config.Machine) {
		return nil, errors.New("database test harness requires a safe Podman machine name")
	}
	if config.Run == nil {
		config.Run = podmanRunner{}
	}
	if config.RunMachine == nil {
		config.RunMachine = podmanMachineRunner{}
	}
	if config.DiskFree == nil {
		config.DiskFree = diskFreeBytes
	}

	suffix, err := randomSuffix()
	if err != nil {
		return nil, fmt.Errorf("generate database test resource name: %w", err)
	}
	prefix := "polymetrics-" + engineSlug(config.Engine) + "-it-" + suffix
	return &Harness{
		config:        config,
		containerName: prefix,
		volumeName:    prefix + "-data",
		runImage:      "localhost/" + prefix + ":run",
		stopSignal:    make(chan struct{}),
	}, nil
}

// Start pulls the configured source image, tags it under one generated image
// reference, creates one named volume, starts one loopback-published
// container, and returns its dynamically assigned, non-default port. Call
// Close in a defer immediately after a successful New; Close is also safe
// after an unsuccessful Start.
func (h *Harness) Start(ctx context.Context) (endpoint Endpoint, startErr error) {
	if err := ctx.Err(); err != nil {
		return Endpoint{}, err
	}
	// Take the sequencing slot before the interrupt handler is armed and
	// before any resource exists, so a queued engine holds no container and
	// an interrupt while queued cannot leave one behind.
	select {
	case engineSlots <- struct{}{}:
	case <-ctx.Done():
		return Endpoint{}, ctx.Err()
	}
	h.mu.Lock()
	h.slotHeld = true
	h.mu.Unlock()

	before, err := h.config.DiskFree()
	if err != nil {
		return Endpoint{}, fmt.Errorf("measure disk free before %s test: %w", h.config.Engine, err)
	}
	h.mu.Lock()
	h.report.DiskFreeBefore = before
	h.mu.Unlock()
	h.installInterruptCleanup()
	defer func() {
		if startErr == nil {
			return
		}
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		if cleanupErr := h.Close(cleanupCtx); cleanupErr != nil {
			startErr = errors.Join(startErr, cleanupErr)
		}
	}()

	if _, err := h.config.Run.Run(ctx, h.config.Connection, "pull", h.config.Image); err != nil {
		return Endpoint{}, fmt.Errorf("pull %s test image: %w", h.config.Engine, err)
	}
	// The source reference now exists because this run's pull returned
	// success. Ownership is not inferred from an earlier inspect.
	// Claim both references before issuing the tag. A tag that returns an
	// error may still have created the reference, so ownership is claimed
	// ahead of the command and cleanup tolerates absence.
	h.mu.Lock()
	h.sourceImageKnown = true
	h.runImageKnown = true
	h.mu.Unlock()
	if _, err := h.config.Run.Run(ctx, h.config.Connection, "image", "tag", h.config.Image, h.runImage); err != nil {
		return Endpoint{}, fmt.Errorf("tag %s test image: %w", h.config.Engine, err)
	}

	volumeAbsent, err := h.resourceAbsent(ctx, "volume", "inspect", h.volumeName)
	if err != nil {
		return Endpoint{}, fmt.Errorf("inspect %s test volume: %w", h.config.Engine, err)
	}
	if !volumeAbsent {
		return Endpoint{}, errors.New("generated database test volume already exists")
	}
	h.mu.Lock()
	h.volumeKnown = true
	h.mu.Unlock()
	if _, err := h.config.Run.Run(ctx, h.config.Connection, "volume", "create", h.volumeName); err != nil {
		return Endpoint{}, fmt.Errorf("create %s test volume: %w", h.config.Engine, err)
	}

	args := []string{
		"run", "--detach",
		"--name", h.containerName,
		"--volume", h.volumeName + ":" + h.config.DataVolumePath,
		"--publish", "127.0.0.1::" + strconv.Itoa(h.config.ContainerPort),
	}
	args = append(args, h.config.ContainerArgs...)
	args = append(args, h.runImage)
	args = append(args, h.config.EngineArgs...)
	containerAbsent, err := h.resourceAbsent(ctx, "container", "inspect", h.containerName)
	if err != nil {
		return Endpoint{}, fmt.Errorf("inspect %s test container: %w", h.config.Engine, err)
	}
	if !containerAbsent {
		return Endpoint{}, errors.New("generated database test container already exists")
	}
	h.mu.Lock()
	h.containerKnown = true
	h.mu.Unlock()
	if _, err := h.config.Run.Run(ctx, h.config.Connection, args...); err != nil {
		return Endpoint{}, fmt.Errorf("start %s test container: %w", h.config.Engine, err)
	}

	mapping, err := h.config.Run.Run(ctx, h.config.Connection, "port", h.containerName, strconv.Itoa(h.config.ContainerPort)+"/tcp")
	if err != nil {
		return Endpoint{}, fmt.Errorf("discover %s test port: %w", h.config.Engine, err)
	}
	port, err := parseMappedPort(mapping, h.config.ContainerPort)
	if err != nil {
		return Endpoint{}, fmt.Errorf("parse %s test port: %w", h.config.Engine, err)
	}
	endpoint = Endpoint{Host: "127.0.0.1", Port: port}
	h.mu.Lock()
	h.endpoint = endpoint
	h.mu.Unlock()
	return endpoint, nil
}

// Close unconditionally attempts container, volume, generated image, source
// image, and optional host-disk reclaim cleanup in that order. Each later
// action still runs if an earlier one fails. A tagged test must call this in
// a defer before assertions, so a failing assertion cannot leak an image.
func (h *Harness) Close(ctx context.Context) error {
	h.closeOnce.Do(func() {
		close(h.stopSignal)
		defer h.releaseSlot()
		var errs []error
		if h.known(func(h *Harness) bool { return h.containerKnown }) {
			if _, err := h.config.Run.Run(ctx, h.config.Connection, "container", "rm", "--force", h.containerName); err != nil && !errors.Is(err, errPodmanResourceNotFound) {
				errs = append(errs, fmt.Errorf("remove %s test container: %w", h.config.Engine, err))
			}
		}
		if h.known(func(h *Harness) bool { return h.volumeKnown }) {
			if _, err := h.config.Run.Run(ctx, h.config.Connection, "volume", "rm", "--force", h.volumeName); err != nil && !errors.Is(err, errPodmanResourceNotFound) {
				errs = append(errs, fmt.Errorf("remove %s test volume: %w", h.config.Engine, err))
			}
		}
		if h.known(func(h *Harness) bool { return h.runImageKnown }) {
			if _, err := h.config.Run.Run(ctx, h.config.Connection, "image", "rm", h.runImage); err != nil && !errors.Is(err, errPodmanResourceNotFound) {
				errs = append(errs, fmt.Errorf("remove %s test image: %w", h.config.Engine, err))
			}
		}
		if !h.config.KeepImage && h.known(func(h *Harness) bool { return h.sourceImageKnown }) {
			if _, err := h.config.Run.Run(ctx, h.config.Connection, "image", "rm", h.config.Image); err != nil && !errors.Is(err, errPodmanResourceNotFound) {
				errs = append(errs, fmt.Errorf("remove %s test source image: %w", h.config.Engine, err))
			}
		}
		if h.config.ReclaimHostDisk {
			if err := h.reclaimHostDisk(ctx); err != nil {
				errs = append(errs, err)
			} else {
				h.mu.Lock()
				h.report.HostDiskReclaimed = true
				h.mu.Unlock()
			}
		}
		after, err := h.config.DiskFree()
		if err != nil {
			errs = append(errs, fmt.Errorf("measure disk free after %s test: %w", h.config.Engine, err))
		} else {
			h.mu.Lock()
			h.report.DiskFreeAfter = after
			h.mu.Unlock()
		}
		h.closeErr = errors.Join(errs...)
	})
	return h.closeErr
}

// reclaimHostDisk discards freed guest blocks so the machine's sparse disk
// file shrinks on the host. Two passes are required: measured on Podman 5.3
// with applehv, a single fstrim immediately after an image removal returns
// only part of the space, because the guest has not finished committing the
// deletion when the first discard is issued.
func (h *Harness) reclaimHostDisk(ctx context.Context) error {
	for range 2 {
		if err := h.config.RunMachine.Run(ctx, "machine", "ssh", h.config.Machine, "sudo fstrim -av"); err != nil {
			return fmt.Errorf("reclaim host disk after %s test: %w", h.config.Engine, err)
		}
	}
	return nil
}

// Report returns the aggregate disk measurement for this run.
func (h *Harness) Report() Report {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.report
}

func (h *Harness) releaseSlot() {
	h.mu.Lock()
	held := h.slotHeld
	h.slotHeld = false
	h.mu.Unlock()
	if held {
		<-engineSlots
	}
}

func (h *Harness) known(check func(*Harness) bool) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return check(h)
}

func (h *Harness) resourceAbsent(ctx context.Context, args ...string) (bool, error) {
	_, err := h.config.Run.Run(ctx, h.config.Connection, args...)
	if err == nil {
		return false, nil
	}
	if errors.Is(err, errPodmanResourceNotFound) {
		return true, nil
	}
	return false, err
}

func (h *Harness) installInterruptCleanup() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-h.stopSignal:
			signal.Stop(signals)
		case <-signals:
			signal.Stop(signals)
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			_ = h.Close(cleanupCtx)
			cancel()
			os.Exit(130)
		}
	}()
}

type podmanRunner struct{}

func (podmanRunner) Run(ctx context.Context, connection string, args ...string) (string, error) {
	command := append([]string{"--connection", connection}, args...)
	cmd := exec.CommandContext(ctx, defaultPodmanBinary, command...)
	output, err := cmd.Output()
	if err != nil {
		if podmanResourceNotFound(err) {
			return "", fmt.Errorf("%w: podman command failed: %w", errPodmanResourceNotFound, err)
		}
		// Do not attach command arguments or captured output. Future engine
		// configurations may contain authentication material and test support
		// must never turn it into an error/logging channel.
		return "", fmt.Errorf("podman command failed: %w", err)
	}
	return string(output), nil
}

// podmanResourceNotFound classifies only the absence phrasings Podman emits.
// "image not known" is Podman's wording where Docker says "no such image";
// both are accepted so a Docker-compatible endpoint behaves the same.
func podmanResourceNotFound(err error) bool {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return false
	}
	stderr := strings.ToLower(string(exitErr.Stderr))
	for _, absent := range []string{
		"no such container",
		"no such volume",
		"no such image",
		"no such object",
		"image not known",
		"no such network",
	} {
		if strings.Contains(stderr, absent) {
			return true
		}
	}
	return false
}

type podmanMachineRunner struct{}

func (podmanMachineRunner) Run(ctx context.Context, args ...string) error {
	if err := exec.CommandContext(ctx, defaultPodmanBinary, args...).Run(); err != nil {
		// As for the container runner, keep command output and configuration
		// values out of a test error.
		return errors.New("podman machine command failed")
	}
	return nil
}

func diskFreeBytes() (uint64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(filepath.Clean("."), &stat); err != nil {
		return 0, err
	}
	return uint64(stat.Bavail) * uint64(stat.Bsize), nil
}

func parseMappedPort(raw string, defaultPort int) (int, error) {
	line := strings.TrimSpace(strings.Split(strings.TrimSpace(raw), "\n")[0])
	idx := strings.LastIndexByte(line, ':')
	if idx < 0 || idx == len(line)-1 {
		return 0, errors.New("podman returned no host port")
	}
	port, err := strconv.Atoi(line[idx+1:])
	if err != nil || port < 1 || port > 65535 {
		return 0, errors.New("podman returned an invalid host port")
	}
	if port == defaultPort {
		return 0, errors.New("podman allocated the database default host port")
	}
	return port, nil
}

func pinnedImage(image string) bool {
	image = strings.TrimSpace(image)
	if image == "" || strings.ContainsAny(image, "\r\n@") || strings.HasSuffix(image, ":latest") {
		return false
	}
	lastSlash := strings.LastIndexByte(image, '/')
	return strings.LastIndexByte(image[lastSlash+1:], ':') > 0
}

// safeName accepts only the characters Podman connection and machine names
// use, so a configured value can never introduce an extra argument or shell
// metacharacter into a command line.
func safeName(value string) bool {
	if strings.TrimSpace(value) == "" || value != strings.TrimSpace(value) {
		return false
	}
	for _, r := range value {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '-' && r != '_' && r != '.' {
			return false
		}
	}
	return true
}

func randomSuffix() (string, error) {
	var raw [8]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw[:]), nil
}

func engineSlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var out strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			out.WriteRune(r)
		}
	}
	if out.Len() == 0 {
		return "db"
	}
	return out.String()
}
