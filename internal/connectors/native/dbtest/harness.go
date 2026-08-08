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
	"path"
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
// sequential. It panics on a non-positive or unbounded value, and on a change
// made while a harness still holds a slot: that harness would return its token
// to the replaced channel and block Close forever after cleanup had run.
func SetMaxConcurrentEngines(n int) {
	if n < 1 || n > 8 {
		panic("dbtest: concurrent engine limit must be between 1 and 8")
	}
	if len(engineSlots) != 0 {
		panic("dbtest: concurrent engine limit must be set before any harness starts")
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

// MachineRunner runs a Podman machine lifecycle command and returns its
// standard output. It is separate from the container runner so a unit test can
// prove machine ownership and the host-disk reclaim without requiring a VM.
type MachineRunner interface {
	Run(context.Context, ...string) (string, error)
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
	// Machine names the Podman machine backing Connection and defaults to
	// Connection. The host-disk reclaim runs only against a machine this
	// process created through NewMachine; a caller-supplied name is used for
	// nothing but reporting. See reclaimHostDisk.
	Machine string
	// ContainerArgs are passed to "podman run" before the image reference.
	ContainerArgs []string
	// EngineArgs are passed to the engine process after the image reference.
	EngineArgs []string
	// ExpectedImageBytes is the engine image's approximate on-disk footprint.
	// It is mandatory because it is what the pre-pull headroom check is
	// measured against; an engine that declares nothing would silently skip
	// that check. See imagePullFreeSpaceFactor.
	ExpectedImageBytes uint64
	// RemoveSharedSourceImage opts in to removing the pulled source image from
	// a machine this run did not create. It is off by default: the source
	// reference is shared with every other lane on that machine, and a pull
	// against an already-cached image is a no-op, so a run cannot tell whether
	// it or someone else put the image there. The run-generated tag is always
	// removed, and on a machine this run created through NewMachine the source
	// image is removed too, because nothing else can be using it.
	RemoveSharedSourceImage bool

	Run        CommandRunner
	RunMachine MachineRunner
	DiskFree   func() (uint64, error)
}

// imagePullFreeSpaceFactor is the multiple of Config.ExpectedImageBytes that
// must be free on the host before an absent source image is pulled. A pull
// writes the compressed layers and then extracts them, so it transiently costs
// roughly twice what the image finally occupies, and the container's writable
// layer and the engine's data volume have to fit alongside it. Three times the
// declared footprint is the documented safety threshold: below it the pull is
// refused before it starts rather than filling the host and failing partway
// through with a disk the test then has to clean up.
const imagePullFreeSpaceFactor = 3

// maxExpectedImageBytes bounds the declared footprint, so the headroom
// multiplication above cannot overflow.
const maxExpectedImageBytes = uint64(1) << 40

// Report records disk free before startup and after all teardown work,
// including the host-disk reclaim. A caller can print only these aggregate
// byte values to make leaks visible without logging connection information.
type Report struct {
	DiskFreeBefore    uint64
	DiskFreeAfter     uint64
	HostDiskReclaimed bool
	// HostDiskReclaimSkipped names why the backing machine was left untrimmed
	// and is empty when the trim ran. A machine this run cannot prove it owns
	// is reported, never trimmed.
	HostDiskReclaimSkipped string
	// HostDiskReleasedBytesEstimate is the host free-space delta across the
	// whole run (after minus before). It is an estimate of what this run gave
	// back and nothing more: no image's own size is measured here, and any
	// unrelated host activity during the run lands in the same number. A
	// negative value is space the run still holds, which is what a skipped trim
	// leaves inside a machine's sparse disk file.
	HostDiskReleasedBytesEstimate int64
	// SourceImageRemoved reports whether the shared source image reference was
	// removed by this run.
	SourceImageRemoved bool
	// SourceImageRetainedReason names why the shared source image was left in
	// place, and is empty when it was removed or was never pulled.
	SourceImageRetainedReason string
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
	closed           bool
	report           Report

	// opMu serialises the create sequence against cleanup so an interrupt
	// can never remove a resource while the command that creates it is still
	// in flight, which would leave the created resource behind.
	opMu      sync.Mutex
	runCancel context.CancelFunc
	closeOnce sync.Once
	closeErr  error
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
	if config.ExpectedImageBytes == 0 || config.ExpectedImageBytes > maxExpectedImageBytes {
		return nil, errors.New("database test harness requires the engine image's approximate size in bytes")
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
	}, nil
}

// Start pulls the configured source image, tags it under one generated image
// reference, creates one named volume, starts one loopback-published
// container, and returns its dynamically assigned, non-default port. A pull
// that would have to download the image is refused when the target image store
// lacks the documented headroom for it. Call Close in a defer immediately after a
// successful New; Close is also safe after an unsuccessful Start.
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
	// Registered before the opMu unlock below, so on return the unlock runs
	// first and this cleanup never calls Close while still holding opMu. It is
	// also registered before the first failure point, and releases the slot
	// itself: Close returns the slot only on its first call, so a Start against
	// an already-closed harness would otherwise hold the token forever and
	// block every later engine on this process.
	defer func() {
		if startErr == nil {
			return
		}
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		if cleanupErr := h.Close(cleanupCtx); cleanupErr != nil {
			startErr = errors.Join(startErr, cleanupErr)
		}
		h.releaseSlot()
	}()

	before, err := h.config.DiskFree()
	if err != nil {
		return Endpoint{}, fmt.Errorf("measure disk free before %s test: %w", h.config.Engine, err)
	}
	h.mu.Lock()
	h.report.DiskFreeBefore = before
	h.mu.Unlock()
	h.installInterruptCleanup()

	// Every create runs under a context Close can cancel, so an interrupt
	// during a long pull aborts it instead of waiting for it.
	ctx, cancelRun := context.WithCancel(ctx)
	h.mu.Lock()
	h.runCancel = cancelRun
	h.mu.Unlock()
	defer cancelRun()

	h.opMu.Lock()
	defer h.opMu.Unlock()
	if err := h.abortIfClosed(); err != nil {
		return Endpoint{}, err
	}

	// This inspect sizes the pull; it is never ownership evidence. A source
	// reference that is already present may have been cached by any other lane
	// on a shared machine, which is exactly why cleanup gates its removal on
	// machine ownership rather than on presence.
	sourceAbsent, err := h.resourceAbsent(ctx, "image", "inspect", h.config.Image)
	if err != nil {
		return Endpoint{}, fmt.Errorf("inspect %s test source image: %w", h.config.Engine, err)
	}
	// A cached image costs no disk to "pull", so the headroom gate applies only
	// to a pull that has to download and extract one.
	if sourceAbsent {
		free, err := h.targetImageStoreFree(ctx)
		if err != nil {
			return Endpoint{}, fmt.Errorf("refusing to pull the %s test image because target image-store capacity cannot be measured: %w", h.config.Engine, err)
		}
		if err := h.assertPullHeadroom(free); err != nil {
			return Endpoint{}, err
		}
	}
	if _, err := h.config.Run.Run(ctx, h.config.Connection, "pull", h.config.Image); err != nil {
		return Endpoint{}, fmt.Errorf("pull %s test image: %w", h.config.Engine, err)
	}
	// The source reference now exists because this run's pull returned success.
	// Claim both references before issuing the tag. A tag that returns an
	// error may still have created the reference, so ownership is claimed
	// ahead of the command and cleanup tolerates absence.
	if err := h.abortIfClosed(); err != nil {
		return Endpoint{}, err
	}
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
	if err := h.abortIfClosed(); err != nil {
		return Endpoint{}, err
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
	if err := h.abortIfClosed(); err != nil {
		return Endpoint{}, err
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
	// Cleanup may have begun after the last create. Returning an endpoint now
	// would hand the caller a container that is already being destroyed, so
	// report failure instead and let the deferred cleanup finish the job.
	if err := h.abortIfClosed(); err != nil {
		return Endpoint{}, err
	}
	endpoint = Endpoint{Host: "127.0.0.1", Port: port}
	h.mu.Lock()
	h.endpoint = endpoint
	h.mu.Unlock()
	return endpoint, nil
}

// Close unconditionally attempts container, volume, generated image, source
// image, and host-disk reclaim cleanup in that order. Each later action still
// runs if an earlier one fails. A tagged test must call this in a defer before
// assertions, so a failing assertion cannot leak an image.
//
// The last two actions are gated on this run having created the backing
// machine: the source image is shared with every other lane on a machine this
// run did not create, and `fstrim -av` reaches every filesystem on it.
func (h *Harness) Close(ctx context.Context) error {
	h.closeOnce.Do(func() {
		// The interrupt registration outlives every removal below. Dropping it
		// first restored default signal handling for the whole teardown, so a
		// Ctrl-C between here and the last removal killed the process with this
		// run's container, volume, and images still on the machine.
		defer unregisterInterruptCleanup(h)
		defer h.releaseSlot()
		// Refuse further creates, abort any in flight, then wait for the
		// create sequence to settle. Without this wait, cleanup could run
		// between an ownership claim and the command that acts on it and
		// leave the resource behind.
		h.mu.Lock()
		h.closed = true
		cancelRun := h.runCancel
		h.mu.Unlock()
		if cancelRun != nil {
			cancelRun()
		}
		h.opMu.Lock()
		defer h.opMu.Unlock()

		// Ownership is proven before the first image removal, not after it: the
		// same answer decides whether the shared source image may be removed and
		// whether the machine may be trimmed, so it has to be established once,
		// ahead of both.
		machineOwned, ownershipReason := h.machineOwnership(ctx)

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
		sourceRemoved, sourceRetained := false, ""
		switch {
		case !h.known(func(h *Harness) bool { return h.sourceImageKnown }):
		case machineOwned || h.config.RemoveSharedSourceImage:
			if _, err := h.config.Run.Run(ctx, h.config.Connection, "image", "rm", h.config.Image); err != nil && !errors.Is(err, errPodmanResourceNotFound) {
				errs = append(errs, fmt.Errorf("remove %s test source image: %w", h.config.Engine, err))
			} else {
				sourceRemoved = true
			}
		default:
			sourceRetained = ownershipReason
		}

		reclaimed, skipped := false, ownershipReason
		if machineOwned {
			trimmed, reclaimErr := h.reclaimHostDisk(ctx)
			if reclaimErr != nil {
				errs = append(errs, reclaimErr)
				skipped = "podman machine trim failed"
			}
			reclaimed = trimmed
		}
		after, err := h.config.DiskFree()
		if err != nil {
			errs = append(errs, fmt.Errorf("measure disk free after %s test: %w", h.config.Engine, err))
		}
		h.mu.Lock()
		h.report.HostDiskReclaimed = reclaimed
		h.report.HostDiskReclaimSkipped = skipped
		h.report.SourceImageRemoved = sourceRemoved
		h.report.SourceImageRetainedReason = sourceRetained
		if err == nil {
			h.report.DiskFreeAfter = after
			h.report.HostDiskReleasedBytesEstimate = releasedBytesEstimate(h.report.DiskFreeBefore, after)
		}
		h.mu.Unlock()
		h.closeErr = errors.Join(errs...)
	})
	return h.closeErr
}

// reclaimHostDisk discards freed guest blocks so the machine's sparse disk
// file shrinks on the host, because removing an image frees space inside the
// VM but leaves the host file inflated. It runs on every cleanup rather than
// on an opt-in, but its caller invokes it only once machineOwnership has
// proven this process created the machine; an unowned machine is reported
// through the Report instead, never trimmed, because a pre-existing, shared,
// or remote endpoint belongs to another lane.
//
// Two passes are required: measured on Podman 5.3 with applehv, a single
// fstrim immediately after an image removal returns only part of the space,
// because the guest has not finished committing the deletion when the first
// discard is issued.
func (h *Harness) reclaimHostDisk(ctx context.Context) (bool, error) {
	for range 2 {
		if _, err := h.config.RunMachine.Run(ctx, "machine", "ssh", h.config.Machine, "sudo fstrim -av"); err != nil {
			return false, fmt.Errorf("reclaim host disk after %s test: %w", h.config.Engine, err)
		}
	}
	return true, nil
}

// assertPullHeadroom refuses a pull the target image store cannot afford. Reporting the
// shortfall before the download starts is the whole point: a pull that fills
// the disk partway through leaves both a partial image and a container the
// test then has to clean up on a host with nothing left to work with.
func (h *Harness) assertPullHeadroom(free uint64) error {
	required := h.config.ExpectedImageBytes * imagePullFreeSpaceFactor
	if free >= required {
		return nil
	}
	return fmt.Errorf("refusing to pull the %s test image: %d bytes free is below the %d bytes required for an image of about %d bytes",
		h.config.Engine, free, required, h.config.ExpectedImageBytes)
}

func (h *Harness) targetImageStoreFree(ctx context.Context) (uint64, error) {
	if h.config.Connection != h.config.Machine && h.config.Connection != h.config.Machine+"-root" {
		return 0, errors.New("Podman connection cannot be measured through the configured local machine")
	}
	name, err := h.config.RunMachine.Run(ctx, "machine", "inspect", "--format", "{{.Name}}", h.config.Machine)
	if err != nil {
		return 0, errors.New("configured Podman machine is not locally inspectable")
	}
	if strings.TrimSpace(name) != h.config.Machine {
		return 0, errors.New("configured Podman connection does not identify the local machine")
	}
	graphRoot, err := h.config.Run.Run(ctx, h.config.Connection, "info", "--format", "{{.Store.GraphRoot}}")
	if err != nil {
		return 0, errors.New("target Podman image store could not be located")
	}
	graphRoot, err = safePodmanStorePath(graphRoot)
	if err != nil {
		return 0, err
	}
	df, err := h.config.RunMachine.Run(ctx, "machine", "ssh", h.config.Machine, "df", "-Pk", graphRoot)
	if err != nil {
		return 0, errors.New("target Podman image store capacity query failed")
	}
	return parseAvailableDiskBytes(df)
}

func safePodmanStorePath(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if !path.IsAbs(value) || path.Clean(value) != value || strings.ContainsAny(value, "\r\n\x00") {
		return "", errors.New("target Podman image store path is invalid")
	}
	for _, r := range value {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '/' && r != '.' && r != '-' && r != '_' {
			return "", errors.New("target Podman image store path is unsafe")
		}
	}
	return value, nil
}

func parseAvailableDiskBytes(raw string) (uint64, error) {
	for _, line := range strings.Split(raw, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 || fields[0] == "Filesystem" {
			continue
		}
		available, err := strconv.ParseUint(fields[3], 10, 64)
		if err != nil || available > ^uint64(0)/1024 {
			return 0, errors.New("target Podman image store capacity is invalid")
		}
		return available * 1024, nil
	}
	return 0, errors.New("target Podman image store capacity was not returned")
}

// releasedBytesEstimate reports the host free-space delta across a run.
// Negative means the run ended holding space it had not released, which is
// what a skipped trim leaves behind.
func releasedBytesEstimate(before, after uint64) int64 {
	const maxInt64 = uint64(1)<<63 - 1
	if after >= before {
		return int64(min(after-before, maxInt64))
	}
	return -int64(min(before-after, maxInt64))
}

// machineOwnership proves, or refuses to prove, that trimming the configured
// machine can only affect this run. The decisive fact is that this process
// created the machine and still holds its ownership record: `fstrim -av`
// reaches every filesystem on a machine, so a name that merely matches
// caller-supplied configuration proves nothing about who else is using it.
// Two weaker checks follow as defence in depth — the connection every
// container command was scoped to must address that machine, which is how
// Podman names a machine's own connections, and the machine must still be
// locally defined. The returned reason names the failed check without echoing
// configuration.
func (h *Harness) machineOwnership(ctx context.Context) (bool, string) {
	if !machineIsOwned(h.config.Machine) {
		return false, "podman machine was not created by this run"
	}
	if h.config.Connection != h.config.Machine && h.config.Connection != h.config.Machine+"-root" {
		return false, "podman connection does not address the machine this run created"
	}
	name, err := h.config.RunMachine.Run(ctx, "machine", "inspect", "--format", "{{.Name}}", h.config.Machine)
	if err != nil {
		return false, "podman machine ownership could not be verified"
	}
	if strings.TrimSpace(name) != h.config.Machine {
		return false, "podman machine is not the local machine this run created"
	}
	return true, ""
}

// Report returns the aggregate disk measurement for this run.
func (h *Harness) Report() Report {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.report
}

func (h *Harness) CopyFileFromContainer(ctx context.Context, source, destination string) error {
	if !safeContainerFilePath(source) {
		return errors.New("database test harness requires a safe absolute container file path")
	}
	if !filepath.IsAbs(destination) || filepath.Clean(destination) != destination || strings.ContainsAny(destination, "\r\n\x00") {
		return errors.New("database test harness requires a safe absolute destination path")
	}
	h.mu.Lock()
	known, closed, container := h.containerKnown, h.closed, h.containerName
	h.mu.Unlock()
	if !known || closed {
		return errors.New("database test harness container is not available")
	}
	if _, err := h.config.Run.Run(ctx, h.config.Connection, "cp", container+":"+source, destination); err != nil {
		return fmt.Errorf("copy database test container file: %w", err)
	}
	return nil
}

func safeContainerFilePath(value string) bool {
	if !path.IsAbs(value) || path.Clean(value) != value || strings.ContainsAny(value, "\r\n\x00:") {
		return false
	}
	return true
}

// abortIfClosed stops the create sequence once cleanup has begun.
func (h *Harness) abortIfClosed() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return errors.New("database test harness was closed during startup")
	}
	return nil
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

// interruptCleanup is the single process-level signal handler. signal.Notify
// delivers to every registered channel, so a handler per harness let the first
// finished cleanup call os.Exit while a sibling's container, volume, or image
// removal was still in flight, leaving those resources behind under the
// bounded parallel mode. One registry closes every live harness before the
// process exits.
var interruptCleanup = struct {
	mu       sync.Mutex
	live     map[*Harness]struct{}
	machines map[*Machine]struct{}
	signals  chan os.Signal
	idle     chan struct{}
}{live: make(map[*Harness]struct{}), machines: make(map[*Machine]struct{})}

func (h *Harness) installInterruptCleanup() {
	interruptCleanup.mu.Lock()
	defer interruptCleanup.mu.Unlock()
	interruptCleanup.live[h] = struct{}{}
	armInterruptWatchLocked()
}

// registerInterruptMachine adds a machine this process created, so an
// interrupt removes a whole VM rather than leaking its disk image. Only owned
// machines are ever registered here.
func registerInterruptMachine(machine *Machine) {
	interruptCleanup.mu.Lock()
	defer interruptCleanup.mu.Unlock()
	interruptCleanup.machines[machine] = struct{}{}
	armInterruptWatchLocked()
}

func armInterruptWatchLocked() {
	if interruptCleanup.signals != nil {
		return
	}
	signals := make(chan os.Signal, 1)
	idle := make(chan struct{})
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	interruptCleanup.signals = signals
	interruptCleanup.idle = idle
	go awaitInterrupt(signals, idle)
}

// unregisterInterruptCleanup drops a closed harness and stops watching once no
// harness owns a resource, so the process keeps default signal handling after
// the last engine has been torn down.
func unregisterInterruptCleanup(h *Harness) {
	interruptCleanup.mu.Lock()
	defer interruptCleanup.mu.Unlock()
	delete(interruptCleanup.live, h)
	disarmInterruptWatchLocked()
}

func unregisterInterruptMachine(machine *Machine) {
	interruptCleanup.mu.Lock()
	defer interruptCleanup.mu.Unlock()
	delete(interruptCleanup.machines, machine)
	disarmInterruptWatchLocked()
}

func disarmInterruptWatchLocked() {
	if len(interruptCleanup.live) != 0 || len(interruptCleanup.machines) != 0 || interruptCleanup.signals == nil {
		return
	}
	signal.Stop(interruptCleanup.signals)
	close(interruptCleanup.idle)
	interruptCleanup.signals = nil
	interruptCleanup.idle = nil
}

func awaitInterrupt(signals chan os.Signal, idle chan struct{}) {
	select {
	case <-idle:
		return
	case <-signals:
	}
	signal.Stop(signals)
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	// Containers first: they live inside the machines removed next, and a
	// machine torn down underneath an in-flight container removal would report
	// a cleanup failure for work that had already become moot.
	closeLiveHarnesses(cleanupCtx)
	removeOwnedMachines(cleanupCtx)
	cancel()
	os.Exit(130)
}

// closeLiveHarnesses tears every registered harness down concurrently and
// returns only once the last one has finished, so nothing can exit while a
// sibling's Podman removal is still running.
func closeLiveHarnesses(ctx context.Context) {
	interruptCleanup.mu.Lock()
	live := make([]*Harness, 0, len(interruptCleanup.live))
	for harness := range interruptCleanup.live {
		live = append(live, harness)
	}
	interruptCleanup.mu.Unlock()

	var wg sync.WaitGroup
	for _, harness := range live {
		wg.Add(1)
		go func(harness *Harness) {
			defer wg.Done()
			_ = harness.Close(ctx)
		}(harness)
	}
	wg.Wait()
}

// removeOwnedMachines deletes every machine this process created. It removes
// nothing else: a machine that was already on the host has no record here.
func removeOwnedMachines(ctx context.Context) {
	interruptCleanup.mu.Lock()
	machines := make([]*Machine, 0, len(interruptCleanup.machines))
	for machine := range interruptCleanup.machines {
		machines = append(machines, machine)
	}
	interruptCleanup.mu.Unlock()

	var wg sync.WaitGroup
	for _, machine := range machines {
		wg.Add(1)
		go func(machine *Machine) {
			defer wg.Done()
			_ = machine.Remove(ctx)
		}(machine)
	}
	wg.Wait()
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
		// Podman's wording for an absent machine, which cleanup after a failed
		// `machine init` has to read as success rather than as a leak.
		"vm does not exist",
		"no such machine",
	} {
		if strings.Contains(stderr, absent) {
			return true
		}
	}
	return false
}

type podmanMachineRunner struct{}

func (podmanMachineRunner) Run(ctx context.Context, args ...string) (string, error) {
	output, err := exec.CommandContext(ctx, defaultPodmanBinary, args...).Output()
	if err != nil {
		// As for the container runner, keep command output and configuration
		// values out of a test error, but classify absence so removing a
		// machine that was never created is success rather than a leak report.
		if podmanResourceNotFound(err) {
			return "", fmt.Errorf("%w: podman machine command failed", errPodmanResourceNotFound)
		}
		return "", errors.New("podman machine command failed")
	}
	return string(output), nil
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
