// Package dbtest provides an opt-in, Docker- or Podman-backed database
// integration-test harness. It is test support only: callers supply one
// engine's pinned image and server arguments, while this package owns runtime
// selection, endpoint scoping, ephemeral resource names, loopback port
// allocation, disk accounting, and unconditional cleanup.
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
	"net/url"
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

const (
	defaultPodmanBinary = "podman"
	defaultDockerBinary = "docker"
	cleanupTimeout      = 3 * time.Minute
)

var errContainerResourceNotFound = errors.New("container resource not found")

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

// Runtime selects the explicit container client used by a harness.
type Runtime string

const (
	// RuntimeDocker runs each command through Docker's explicit --host flag.
	RuntimeDocker Runtime = "docker"
	// RuntimePodman runs each command through Podman's explicit --url flag.
	RuntimePodman Runtime = "podman"
)

// CommandRunner runs a container-runtime subcommand against one explicit
// endpoint. It exists solely to make resource ownership and cleanup testable
// without a running container engine.
type CommandRunner interface {
	Run(context.Context, string, ...string) (string, error)
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
	// ContainerRuntime is the explicit Docker or Podman client for this run.
	ContainerRuntime Runtime
	// ContainerEndpoint is the direct local Unix API endpoint for
	// ContainerRuntime.
	ContainerEndpoint string
	// ContainerArgs are passed to the runtime's "run" command before the image
	// reference.
	ContainerArgs []string
	// EngineArgs are passed to the engine process after the image reference.
	EngineArgs []string
	// ExpectedImageBytes is the engine image's approximate on-disk footprint.
	// It is mandatory because it is what the pre-pull headroom check is
	// measured against; an engine that declares nothing would silently skip
	// that check. See imagePullFreeSpaceFactor.
	ExpectedImageBytes uint64
	// DockerCapacityProbeImage is a pre-cached pinned image used only when a
	// selected Docker daemon reports an image-store path that the client host
	// cannot measure directly (for example, a local Colima VM). dbtest never
	// pulls this image: it starts an ephemeral locked-down df probe through the
	// same configured endpoint. Leave it empty to refuse an unmeasurable Docker
	// store.
	DockerCapacityProbeImage string
	Run                      CommandRunner
	DiskFreeAt               func(string) (uint64, error)
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

// Report records free bytes on the verified target image store before startup
// and after all teardown work.
type Report struct {
	DiskFreeBefore uint64
	DiskFreeAfter  uint64
}

type targetIdentity struct {
	runtime    Runtime
	targetID   string
	graphRoot  string
	forwarded  bool
	daemonFree uint64
}

// Harness owns exactly one generated container with one anonymous data volume.
// It is not safe to copy after first use.
type Harness struct {
	config Config

	containerName      string
	containerOwner     string
	containerID        string
	runImage           string
	runImageID         string
	capacityProbeName  string
	capacityProbeOwner string

	mu                 sync.Mutex
	endpoint           Endpoint
	containerKnown     bool
	capacityProbeKnown bool
	slotHeld           bool
	closed             bool
	report             Report
	target             targetIdentity
	targetKnown        bool

	// opMu serialises the create sequence against cleanup so an interrupt
	// can never remove a resource while the command that creates it is still
	// in flight, which would leave the created resource behind.
	opMu      sync.Mutex
	runCancel context.CancelFunc
	closeOnce sync.Once
	closeErr  error
}

// New validates the test-only engine configuration.
func New(config Config) (*Harness, error) {
	if strings.TrimSpace(config.Engine) == "" {
		return nil, errors.New("database test harness requires an engine name")
	}
	if !config.ContainerRuntime.valid() {
		return nil, errors.New("database test harness requires an explicit container runtime: docker or podman")
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
	if _, err := containerEndpointPath(config.ContainerEndpoint); err != nil {
		return nil, errors.New("database test harness requires an explicit safe local Docker or Podman endpoint")
	}
	if config.ExpectedImageBytes == 0 || config.ExpectedImageBytes > maxExpectedImageBytes {
		return nil, errors.New("database test harness requires the engine image's approximate size in bytes")
	}
	if config.DockerCapacityProbeImage != "" && (!pinnedImage(config.DockerCapacityProbeImage) || strings.TrimSpace(config.DockerCapacityProbeImage) != config.DockerCapacityProbeImage) {
		return nil, errors.New("database test harness requires a safe pinned Docker capacity probe image")
	}
	if config.Run == nil {
		config.Run = containerRunner{runtime: config.ContainerRuntime}
	}
	if config.DiskFreeAt == nil {
		config.DiskFreeAt = diskFreeAt
	}

	suffix, err := randomSuffix()
	if err != nil {
		return nil, fmt.Errorf("generate database test resource name: %w", err)
	}
	containerOwner, err := randomSuffix()
	if err != nil {
		return nil, fmt.Errorf("generate database test container owner: %w", err)
	}
	capacityProbeOwner, err := randomSuffix()
	if err != nil {
		return nil, fmt.Errorf("generate database test capacity probe owner: %w", err)
	}
	prefix := "polymetrics-" + engineSlug(config.Engine) + "-it-" + suffix
	return &Harness{
		config:             config,
		containerName:      prefix,
		containerOwner:     containerOwner,
		runImage:           "localhost/" + prefix + ":run",
		capacityProbeName:  prefix + "-capacity",
		capacityProbeOwner: capacityProbeOwner,
	}, nil
}

// Start pulls the configured source image, tags it under one generated image
// reference, starts one loopback-published container with an anonymous data
// volume, and returns its dynamically assigned, non-default port. A pull
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
		cleanupCtx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
		defer cancel()
		if cleanupErr := h.Close(cleanupCtx); cleanupErr != nil {
			startErr = errors.Join(startErr, cleanupErr)
		}
		h.releaseSlot()
	}()

	if !h.installInterruptCleanup() {
		return Endpoint{}, errors.New("database test harness was closed during startup")
	}

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
	free, err := h.targetImageStoreFree(ctx)
	if err != nil {
		return Endpoint{}, fmt.Errorf("refusing to use the %s test target because its identity or image-store capacity cannot be proven: %w", h.config.Engine, err)
	}
	h.mu.Lock()
	h.report.DiskFreeBefore = free
	h.mu.Unlock()

	sourceAbsent, err := h.resourceAbsent(ctx, "image", "inspect", h.config.Image)
	if err != nil {
		return Endpoint{}, fmt.Errorf("inspect %s test source image: %w", h.config.Engine, err)
	}
	if sourceAbsent {
		if err := h.assertPullHeadroom(free); err != nil {
			return Endpoint{}, err
		}
	}
	if _, err := h.runOnVerifiedTarget(ctx, "pull", h.config.Image); err != nil {
		return Endpoint{}, fmt.Errorf("pull %s test image: %w", h.config.Engine, err)
	}
	sourceImageID, err := h.imageID(ctx, h.config.Image)
	if err != nil {
		return Endpoint{}, fmt.Errorf("inspect %s test image identity: %w", h.config.Engine, err)
	}
	if err := h.abortIfClosed(); err != nil {
		return Endpoint{}, err
	}
	h.mu.Lock()
	h.runImageID = sourceImageID
	h.mu.Unlock()
	if _, err := h.runOnVerifiedTarget(ctx, "image", "tag", sourceImageID, h.runImage); err != nil {
		return Endpoint{}, fmt.Errorf("tag %s test image: %w", h.config.Engine, err)
	}
	taggedImageID, err := h.imageID(ctx, h.runImage)
	if err != nil {
		return Endpoint{}, fmt.Errorf("inspect generated %s test image identity: %w", h.config.Engine, err)
	}
	if taggedImageID != sourceImageID {
		return Endpoint{}, errors.New("generated database test image does not refer to the configured source image")
	}

	args := []string{
		"run", "--detach",
		"--name", h.containerName,
		"--volume", h.config.DataVolumePath,
		"--publish", "127.0.0.1::" + strconv.Itoa(h.config.ContainerPort),
	}
	args = append(args, h.config.ContainerArgs...)
	args = append(args, "--label", databaseContainerOwnerLabel+"="+h.containerOwner)
	args = append(args, sourceImageID)
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
	if _, err := h.runOnVerifiedTarget(ctx, args...); err != nil {
		return Endpoint{}, fmt.Errorf("start %s test container: %w", h.config.Engine, err)
	}
	containerID, err := h.ownedContainerID(ctx, h.containerName, h.containerOwner, databaseContainerOwnerFormat)
	if err != nil {
		return Endpoint{}, fmt.Errorf("inspect generated %s test container ownership: %w", h.config.Engine, err)
	}
	h.mu.Lock()
	h.containerID = containerID
	h.mu.Unlock()

	mapping, err := h.runOnVerifiedTarget(ctx, "port", containerID, strconv.Itoa(h.config.ContainerPort)+"/tcp")
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

// Close unconditionally attempts harness cleanup. Each later action still runs
// if an earlier one fails.
// A tagged test must call this in a defer before assertions, so a failing
// assertion cannot leak a generated resource.
func (h *Harness) Close(_ context.Context) error {
	h.closeOnce.Do(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
		defer cancel()
		// The interrupt registration outlives every removal below. Dropping it
		// first restored default signal handling for the whole teardown, so a
		// Ctrl-C between here and the last removal killed the process with this
		// run's container and volume still on the machine.
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

		var errs []error
		if h.known(func(h *Harness) bool { return h.containerKnown }) {
			h.mu.Lock()
			containerID := h.containerID
			h.mu.Unlock()
			if err := h.removeOwnedContainer(cleanupCtx, h.containerName, h.containerOwner, containerID, databaseContainerOwnerFormat, true); err != nil {
				errs = append(errs, fmt.Errorf("remove %s test container: %w", h.config.Engine, err))
			}
		}
		if h.known(func(h *Harness) bool { return h.targetKnown }) {
			after, err := h.targetImageStoreFree(cleanupCtx)
			if err != nil {
				errs = append(errs, fmt.Errorf("measure target image-store free space after %s test: %w", h.config.Engine, err))
			} else {
				h.mu.Lock()
				h.report.DiskFreeAfter = after
				h.mu.Unlock()
			}
		}
		if h.known(func(h *Harness) bool { return h.capacityProbeKnown }) {
			if err := h.removeOwnedContainer(cleanupCtx, h.capacityProbeName, h.capacityProbeOwner, "", dockerCapacityProbeOwnerFormat, false); err != nil {
				errs = append(errs, fmt.Errorf("remove %s test capacity probe: %w", h.config.Engine, err))
			}
		}
		h.closeErr = errors.Join(errs...)
	})
	return h.closeErr
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
	target, err := h.assertTarget(ctx)
	if err != nil {
		return 0, err
	}
	if target.forwarded {
		return target.daemonFree, nil
	}
	free, err := h.config.DiskFreeAt(target.graphRoot)
	if err == nil {
		return free, nil
	}
	if target.runtime == RuntimeDocker && h.config.DockerCapacityProbeImage != "" {
		return h.dockerVMImageStoreFree(ctx, target)
	}
	return 0, errors.New("target container image-store capacity query failed")
}

const (
	dockerCapacityMountPath        = "/polymetrics-image-store"
	dockerCapacityProbeOwnerLabel  = "polymetrics.dbtest.capacity-probe-owner"
	dockerCapacityProbeOwnerFormat = "{{.Id}}\t{{ index .Config.Labels \"polymetrics.dbtest.capacity-probe-owner\" }}"
	databaseContainerOwnerLabel    = "polymetrics.dbtest.container-owner"
	databaseContainerOwnerFormat   = "{{.Id}}\t{{ index .Config.Labels \"polymetrics.dbtest.container-owner\" }}"
	imageIDFormat                  = "{{.Id}}"
	dockerCapacityHeader           = "Filesystem\t1-blocks\tUsed\tAvailable\tCapacity\tMounted\ton"
)

func (h *Harness) dockerVMImageStoreFree(ctx context.Context, target targetIdentity) (uint64, error) {
	probeImageID, err := h.imageID(ctx, h.config.DockerCapacityProbeImage)
	if errors.Is(err, errContainerResourceNotFound) {
		return 0, errors.New("target Docker image-store capacity requires a pre-cached Docker capacity probe image")
	}
	if err != nil {
		return 0, errors.New("target Docker capacity probe image could not be inspected")
	}

	probeContainerAbsent, err := h.resourceAbsent(ctx, "container", "inspect", h.capacityProbeName)
	if err != nil {
		return 0, errors.New("target Docker capacity probe name could not be inspected")
	}
	if !probeContainerAbsent {
		return 0, errors.New("generated Docker capacity probe already exists")
	}

	// Claim the transient name before starting the probe. Although --rm removes
	// a successful probe, Close also attempts its idempotent removal if Docker
	// reports an error after creating it or the test is interrupted.
	h.mu.Lock()
	h.capacityProbeKnown = true
	h.mu.Unlock()
	output, err := h.runOnVerifiedTarget(ctx,
		"run", "--pull=never", "--rm", "--name", h.capacityProbeName,
		"--label", dockerCapacityProbeOwnerLabel+"="+h.capacityProbeOwner,
		"--network", "none", "--read-only", "--cap-drop", "ALL", "--pids-limit", "16",
		"--security-opt", "no-new-privileges", "--env", "LC_ALL=C",
		"--mount", "type=bind,src="+target.graphRoot+",dst="+dockerCapacityMountPath+",readonly",
		"--entrypoint", "/bin/df", probeImageID,
		"-P", "-B1", dockerCapacityMountPath,
	)
	if err != nil {
		return 0, errors.New("target Docker image-store capacity probe failed")
	}
	free, err := parseDockerStoreFree(output)
	if err != nil {
		return 0, errors.New("target Docker endpoint returned an invalid Docker image-store capacity")
	}
	return free, nil
}

func (h *Harness) ownedContainerID(ctx context.Context, name, owner, ownerFormat string) (string, error) {
	output, err := h.runOnVerifiedTarget(ctx,
		"container", "inspect", "--format", ownerFormat, name,
	)
	if err != nil {
		return "", err
	}
	containerID, reportedOwner, valid := parseContainerOwnership(output)
	if !valid || reportedOwner != owner {
		return "", errors.New("generated container ownership could not be proven")
	}
	return containerID, nil
}

func (h *Harness) removeOwnedContainer(ctx context.Context, name, owner, containerID, ownerFormat string, removeAnonymousVolumes bool) error {
	if containerID == "" {
		var err error
		containerID, err = h.ownedContainerID(ctx, name, owner, ownerFormat)
		if errors.Is(err, errContainerResourceNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
	}
	args := []string{"container", "rm", "--force"}
	if removeAnonymousVolumes {
		args = append(args, "--volumes")
	}
	args = append(args, containerID)
	if _, err := h.runOnVerifiedTarget(ctx, args...); err != nil && !errors.Is(err, errContainerResourceNotFound) {
		return err
	}
	return nil
}

func (h *Harness) imageID(ctx context.Context, reference string) (string, error) {
	output, err := h.runOnVerifiedTarget(ctx, "image", "inspect", "--format", imageIDFormat, reference)
	if err != nil {
		return "", err
	}
	imageID, valid := parseImageID(output)
	if !valid {
		return "", errors.New("container image identity could not be proven")
	}
	return imageID, nil
}

func (h *Harness) runOnVerifiedTarget(ctx context.Context, args ...string) (string, error) {
	if _, err := h.assertTarget(ctx); err != nil {
		return "", err
	}
	return h.config.Run.Run(ctx, h.config.ContainerEndpoint, args...)
}

func (h *Harness) assertTarget(ctx context.Context) (targetIdentity, error) {
	endpointPath, err := containerEndpointPath(h.config.ContainerEndpoint)
	if err != nil {
		return targetIdentity{}, err
	}
	raw, err := h.config.Run.Run(ctx, h.config.ContainerEndpoint, "info", "--format", h.config.ContainerRuntime.infoFormat())
	if err != nil {
		return targetIdentity{}, fmt.Errorf("target %s endpoint could not be inspected", h.config.ContainerRuntime)
	}
	target, err := h.config.ContainerRuntime.parseTargetIdentity(raw, endpointPath)
	if err != nil {
		return targetIdentity{}, err
	}
	target.runtime = h.config.ContainerRuntime
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.targetKnown && !sameTargetIdentity(h.target, target) {
		return targetIdentity{}, errors.New("target container endpoint identity changed during the test")
	}
	h.target = target
	h.targetKnown = true
	return target, nil
}

func containerEndpointPath(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" || value != raw || strings.ContainsAny(value, "\r\n\x00") {
		return "", errors.New("target container endpoint is invalid")
	}
	endpoint, err := url.Parse(value)
	if err != nil || endpoint.Scheme != "unix" || endpoint.Host != "" || endpoint.Opaque != "" || endpoint.RawQuery != "" || endpoint.Fragment != "" || endpoint.User != nil || endpoint.RawPath != "" {
		return "", errors.New("target container endpoint must be a direct local Unix socket")
	}
	return safeContainerStorePath(endpoint.Path)
}

func (runtime Runtime) valid() bool {
	return runtime == RuntimeDocker || runtime == RuntimePodman
}

func (runtime Runtime) infoFormat() string {
	switch runtime {
	case RuntimeDocker:
		return "{{.ID}}\t{{.DockerRootDir}}"
	case RuntimePodman:
		return "{{.Host.RemoteSocket.Path}}\t{{.Store.GraphRoot}}\t{{.Store.GraphRootAllocated}}\t{{.Store.GraphRootUsed}}"
	default:
		return ""
	}
}

func (runtime Runtime) parseTargetIdentity(raw, endpointPath string) (targetIdentity, error) {
	switch runtime {
	case RuntimeDocker:
		return parseDockerTargetIdentity(raw)
	case RuntimePodman:
		return parseTargetIdentity(raw, endpointPath)
	default:
		return targetIdentity{}, errors.New("target container runtime is invalid")
	}
}

func parseDockerTargetIdentity(raw string) (targetIdentity, error) {
	line, valid := singleCommandRecord(raw)
	if !valid {
		return targetIdentity{}, errors.New("target Docker endpoint did not report a daemon identity and locally measurable image-store path")
	}
	fields := strings.Split(line, "\t")
	if len(fields) != 2 || !safeContainerIdentity(fields[0]) {
		return targetIdentity{}, errors.New("target Docker endpoint did not report a daemon identity and locally measurable image-store path")
	}
	graphRoot, err := safeContainerStorePath(fields[1])
	if err != nil {
		return targetIdentity{}, errors.New("target Docker image store path is invalid")
	}
	return targetIdentity{targetID: fields[0], graphRoot: graphRoot}, nil
}

func safeContainerIdentity(value string) bool {
	if value == "" || strings.TrimSpace(value) != value || strings.ContainsAny(value, "\t\r\n\x00") {
		return false
	}
	for _, r := range value {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != ':' && r != '-' && r != '_' {
			return false
		}
	}
	return true
}

func parseTargetIdentity(raw, endpointPath string) (targetIdentity, error) {
	line, valid := singleCommandRecord(raw)
	if !valid {
		return targetIdentity{}, errors.New("target Podman endpoint returned an invalid identity")
	}
	fields := strings.Split(line, "\t")
	if len(fields) != 4 {
		return targetIdentity{}, errors.New("target Podman endpoint returned an invalid identity")
	}
	socketPath, forwarded, err := podmanReportedSocketPath(fields[0])
	if err != nil {
		return targetIdentity{}, errors.New("target Podman endpoint returned an invalid socket path")
	}
	graphRoot, err := safeContainerStorePath(fields[1])
	if err != nil {
		return targetIdentity{}, errors.New("target Podman image store path is invalid")
	}
	if socketPath == endpointPath {
		return targetIdentity{targetID: socketPath, graphRoot: graphRoot}, nil
	}
	if !forwarded {
		return targetIdentity{}, errors.New("target Podman endpoint identity does not match the configured socket")
	}
	free, err := podmanStoreFree(fields[2], fields[3])
	if err != nil {
		return targetIdentity{}, errors.New("target Podman endpoint returned an invalid image-store capacity")
	}
	return targetIdentity{targetID: socketPath, graphRoot: graphRoot, forwarded: true, daemonFree: free}, nil
}

func sameTargetIdentity(left, right targetIdentity) bool {
	return left.runtime == right.runtime && left.targetID == right.targetID && left.graphRoot == right.graphRoot && left.forwarded == right.forwarded
}

func podmanReportedSocketPath(raw string) (string, bool, error) {
	value := strings.TrimSpace(raw)
	if value == "" || value != raw {
		return "", false, errors.New("invalid Podman socket path")
	}
	if strings.HasPrefix(value, "unix://") {
		path, err := containerEndpointPath(value)
		return path, true, err
	}
	path, err := safeContainerStorePath(value)
	return path, false, err
}

func podmanStoreFree(allocatedRaw, usedRaw string) (uint64, error) {
	allocated, err := parsePodmanBytes(allocatedRaw)
	if err != nil {
		return 0, err
	}
	used, err := parsePodmanBytes(usedRaw)
	if err != nil || used > allocated {
		return 0, errors.New("invalid Podman image-store capacity")
	}
	return allocated - used, nil
}

func parsePodmanBytes(raw string) (uint64, error) {
	if raw == "" || strings.TrimSpace(raw) != raw {
		return 0, errors.New("invalid Podman byte count")
	}
	for _, r := range raw {
		if r < '0' || r > '9' {
			return 0, errors.New("invalid Podman byte count")
		}
	}
	return strconv.ParseUint(raw, 10, 64)
}

func parseDockerStoreFree(raw string) (uint64, error) {
	if strings.ContainsAny(raw, "\r\x00") {
		return 0, errors.New("invalid Docker image-store capacity")
	}
	lines := strings.Split(strings.TrimSuffix(raw, "\n"), "\n")
	if len(lines) != 2 || strings.Join(strings.Fields(lines[0]), "\t") != dockerCapacityHeader {
		return 0, errors.New("invalid Docker image-store capacity")
	}
	fields := strings.Fields(lines[1])
	if len(fields) != 6 || fields[5] != dockerCapacityMountPath {
		return 0, errors.New("invalid Docker image-store capacity")
	}
	total, err := parseDockerCapacityBytes(fields[1])
	if err != nil {
		return 0, err
	}
	used, err := parseDockerCapacityBytes(fields[2])
	if err != nil {
		return 0, err
	}
	free, err := parseDockerCapacityBytes(fields[3])
	if err != nil || used > total || free > total || used > total-free {
		return 0, errors.New("invalid Docker image-store capacity")
	}
	capacity := strings.TrimSuffix(fields[4], "%")
	percent, err := parseDockerCapacityBytes(capacity)
	if err != nil || !strings.HasSuffix(fields[4], "%") || percent > 100 {
		return 0, errors.New("invalid Docker image-store capacity")
	}
	return free, nil
}

func parseDockerCapacityBytes(value string) (uint64, error) {
	if value == "" {
		return 0, errors.New("invalid Docker image-store capacity")
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return 0, errors.New("invalid Docker image-store capacity")
		}
	}
	return strconv.ParseUint(value, 10, 64)
}

func singleCommandRecord(raw string) (string, bool) {
	line := strings.TrimSuffix(raw, "\n")
	return line, line != "" && !strings.ContainsAny(line, "\r\n")
}

func parseContainerOwnership(raw string) (string, string, bool) {
	line, valid := singleCommandRecord(raw)
	if !valid {
		return "", "", false
	}
	fields := strings.Split(line, "\t")
	if len(fields) != 2 || !safeContainerID(fields[0]) {
		return "", "", false
	}
	return fields[0], fields[1], true
}

func parseImageID(raw string) (string, bool) {
	line, valid := singleCommandRecord(raw)
	if !valid {
		return "", false
	}
	return line, safeImageID(line)
}

func safeImageID(value string) bool {
	value = strings.TrimPrefix(value, "sha256:")
	return safeContainerID(value)
}

func safeContainerID(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, r := range value {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}

func safeContainerStorePath(raw string) (string, error) {
	if raw == "" || strings.TrimSpace(raw) != raw || !path.IsAbs(raw) || path.Clean(raw) != raw || strings.ContainsAny(raw, "\r\n\x00") {
		return "", errors.New("target container image store path is invalid")
	}
	for _, r := range raw {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '/' && r != '.' && r != '-' && r != '_' {
			return "", errors.New("target container image store path is unsafe")
		}
	}
	return raw, nil
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
	known, closed, containerID := h.containerKnown, h.closed, h.containerID
	h.mu.Unlock()
	if !known || closed || containerID == "" {
		return errors.New("database test harness container is not available")
	}
	if _, err := h.runOnVerifiedTarget(ctx, "cp", containerID+":"+source, destination); err != nil {
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
	_, err := h.runOnVerifiedTarget(ctx, args...)
	if err == nil {
		return false, nil
	}
	if errors.Is(err, errContainerResourceNotFound) {
		return true, nil
	}
	return false, err
}

// interruptCleanup is the single process-level signal handler. signal.Notify
// delivers to every registered channel, so a handler per harness let the first
// finished cleanup call os.Exit while a sibling's container or volume removal
// was still in flight, leaving those resources behind under the
// bounded parallel mode. One registry closes every live harness before the
// process exits.
var interruptCleanup = struct {
	mu       sync.Mutex
	live     map[*Harness]struct{}
	signals  chan os.Signal
	idle     chan struct{}
	draining bool
}{live: make(map[*Harness]struct{})}

func (h *Harness) installInterruptCleanup() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return false
	}
	interruptCleanup.mu.Lock()
	defer interruptCleanup.mu.Unlock()
	if interruptCleanup.draining {
		return false
	}
	interruptCleanup.live[h] = struct{}{}
	armInterruptWatchLocked()
	return true
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

func disarmInterruptWatchLocked() {
	if len(interruptCleanup.live) != 0 || interruptCleanup.signals == nil || interruptCleanup.draining {
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
	live := beginInterruptDrain()
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	done := make(chan struct{})
	go func() {
		closeHarnesses(cleanupCtx, live)
		close(done)
	}()
	waitForInterruptCleanup(signals, done)
	cancel()
	finishInterruptDrain(signals)
	os.Exit(130)
}

func waitForInterruptCleanup(signals <-chan os.Signal, done <-chan struct{}) {
	for {
		select {
		case <-signals:
		case <-done:
			return
		}
	}
}

// closeLiveHarnesses tears every registered harness down concurrently and
// returns only once the last one has finished, so nothing can exit while a
// sibling's generated-resource cleanup is still running.
func closeLiveHarnesses(ctx context.Context) {
	closeHarnesses(ctx, snapshotLiveHarnesses())
}

func beginInterruptDrain() []*Harness {
	interruptCleanup.mu.Lock()
	defer interruptCleanup.mu.Unlock()
	if interruptCleanup.draining {
		return nil
	}
	interruptCleanup.draining = true
	return liveHarnessesLocked()
}

func finishInterruptDrain(signals chan os.Signal) {
	interruptCleanup.mu.Lock()
	defer interruptCleanup.mu.Unlock()
	if interruptCleanup.signals != signals {
		return
	}
	signal.Stop(signals)
	close(interruptCleanup.idle)
	interruptCleanup.signals = nil
	interruptCleanup.idle = nil
	interruptCleanup.draining = false
}

func snapshotLiveHarnesses() []*Harness {
	interruptCleanup.mu.Lock()
	defer interruptCleanup.mu.Unlock()
	return liveHarnessesLocked()
}

func liveHarnessesLocked() []*Harness {
	live := make([]*Harness, 0, len(interruptCleanup.live))
	for harness := range interruptCleanup.live {
		live = append(live, harness)
	}
	return live
}

func closeHarnesses(ctx context.Context, live []*Harness) {
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

type containerRunner struct {
	runtime Runtime
}

func (runner containerRunner) Run(ctx context.Context, endpoint string, args ...string) (string, error) {
	binary, command := runner.runtime.command(endpoint, args...)
	cmd := exec.CommandContext(ctx, binary, command...)
	output, err := cmd.Output()
	if err != nil {
		if containerResourceNotFound(err) {
			return "", fmt.Errorf("%w: %s command failed: %w", errContainerResourceNotFound, runner.runtime, err)
		}
		// Do not attach command arguments or captured output. Future engine
		// configurations may contain authentication material and test support
		// must never turn it into an error/logging channel.
		return "", fmt.Errorf("%s command failed: %w", runner.runtime, err)
	}
	return string(output), nil
}

func (runtime Runtime) command(endpoint string, args ...string) (string, []string) {
	switch runtime {
	case RuntimeDocker:
		return defaultDockerBinary, append([]string{"--host", endpoint}, args...)
	case RuntimePodman:
		return defaultPodmanBinary, append([]string{"--url", endpoint}, args...)
	default:
		return "", nil
	}
}

// containerResourceNotFound classifies only the absence phrasings Docker and
// Podman emit. It must not turn endpoint, permission, or daemon failures into
// a harmless absent-resource cleanup result.
func containerResourceNotFound(err error) bool {
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

func parseMappedPort(raw string, defaultPort int) (int, error) {
	line := strings.TrimSpace(strings.Split(strings.TrimSpace(raw), "\n")[0])
	idx := strings.LastIndexByte(line, ':')
	if idx < 0 || idx == len(line)-1 {
		return 0, errors.New("container runtime returned no host port")
	}
	port, err := strconv.Atoi(line[idx+1:])
	if err != nil || port < 1 || port > 65535 {
		return 0, errors.New("container runtime returned an invalid host port")
	}
	if port == defaultPort {
		return 0, errors.New("container runtime allocated the database default host port")
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
