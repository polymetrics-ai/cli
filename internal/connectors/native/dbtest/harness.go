// Package dbtest provides an opt-in, Docker/Colima-backed database
// integration-test harness. It is test support only: callers supply one
// engine's pinned image and server arguments, while this package owns Docker
// context scoping, ephemeral names, loopback port allocation, disk accounting,
// and unconditional cleanup.
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

const (
	defaultDockerBinary  = "docker"
	defaultColimaBinary  = "colima"
	defaultColimaProfile = "default"
)

// Endpoint is the loopback-only address allocated for an ephemeral engine.
// It deliberately represents host and port separately so callers never need
// to persist or log a connection string.
type Endpoint struct {
	Host string
	Port int
}

// CommandRunner runs a Docker subcommand against one explicit context. It
// exists solely to make resource ownership and cleanup testable without a
// running container engine.
type CommandRunner interface {
	Run(context.Context, string, ...string) (string, error)
}

// ColimaRunner runs a Colima lifecycle command. It is separate from Docker so
// a unit test can prove the optional host-disk reset happens after Docker
// cleanup without requiring a VM.
type ColimaRunner interface {
	Run(context.Context, ...string) error
}

// Config describes one supported database-engine container. It is internal
// test support, not a generic command surface.
type Config struct {
	Engine         string
	Image          string
	ContainerPort  int
	DataVolumePath string
	DockerContext  string
	ContainerArgs  []string
	EngineArgs     []string
	KeepImage      bool
	ResetColima    bool
	ColimaProfile  string
	Run            CommandRunner
	RunColima      ColimaRunner
	DiskFree       func() (uint64, error)
}

// Report records disk free before startup and after all teardown work,
// including an optional Colima reset. A caller can print only these aggregate
// byte values to make leaks visible without logging connection information.
type Report struct {
	DiskFreeBefore uint64
	DiskFreeAfter  uint64
	ColimaReset    bool
}

// Harness owns exactly one generated container and one generated named volume.
// It is not safe to copy after first use.
type Harness struct {
	config Config

	containerName string
	volumeName    string

	mu             sync.Mutex
	endpoint       Endpoint
	pulledByRun    bool
	containerKnown bool
	volumeKnown    bool
	report         Report

	closeOnce  sync.Once
	closeErr   error
	stopSignal chan struct{}
}

// New validates the test-only engine configuration. An explicit Docker
// context is mandatory: falling back to a global default could address a
// different lane's runtime and is therefore refused before any Docker call.
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
	if strings.TrimSpace(config.DockerContext) == "" || strings.ContainsAny(config.DockerContext, "\r\n") {
		return nil, errors.New("database test harness requires explicit DOCKER_CONTEXT")
	}
	if config.ResetColima && config.KeepImage {
		return nil, errors.New("database test harness cannot keep an image while resetting Colima")
	}
	if config.ColimaProfile == "" {
		config.ColimaProfile = defaultColimaProfile
	}
	if !safeProfile(config.ColimaProfile) {
		return nil, errors.New("database test harness requires a safe Colima profile")
	}
	if config.Run == nil {
		config.Run = dockerRunner{}
	}
	if config.RunColima == nil {
		config.RunColima = colimaRunner{}
	}
	if config.DiskFree == nil {
		config.DiskFree = diskFreeBytes
	}

	suffix, err := randomSuffix()
	if err != nil {
		return nil, fmt.Errorf("generate database test resource name: %w", err)
	}
	prefix := "polymetrics-" + safeName(config.Engine) + "-it-" + suffix
	return &Harness{
		config:        config,
		containerName: prefix,
		volumeName:    prefix + "-data",
		stopSignal:    make(chan struct{}),
	}, nil
}

// Start pulls an image only when absent, creates one named volume, starts one
// loopback-published container, and returns its dynamically assigned,
// non-default port. Call Close in a defer immediately after a successful New;
// Close is also safe after an unsuccessful Start.
func (h *Harness) Start(ctx context.Context) (Endpoint, error) {
	if err := ctx.Err(); err != nil {
		return Endpoint{}, err
	}
	before, err := h.config.DiskFree()
	if err != nil {
		return Endpoint{}, fmt.Errorf("measure disk free before %s test: %w", h.config.Engine, err)
	}
	h.mu.Lock()
	h.report.DiskFreeBefore = before
	h.mu.Unlock()
	h.installInterruptCleanup()

	if _, err := h.config.Run.Run(ctx, h.config.DockerContext, "image", "inspect", h.config.Image); err != nil {
		if _, err := h.config.Run.Run(ctx, h.config.DockerContext, "pull", h.config.Image); err != nil {
			return Endpoint{}, fmt.Errorf("pull %s test image: %w", h.config.Engine, err)
		}
		h.mu.Lock()
		h.pulledByRun = true
		h.mu.Unlock()
	}

	if _, err := h.config.Run.Run(ctx, h.config.DockerContext, "volume", "create", h.volumeName); err != nil {
		return Endpoint{}, fmt.Errorf("create %s test volume: %w", h.config.Engine, err)
	}
	h.mu.Lock()
	h.volumeKnown = true
	h.mu.Unlock()

	args := []string{
		"run", "--detach",
		"--name", h.containerName,
		"--volume", h.volumeName + ":" + h.config.DataVolumePath,
		"--publish", "127.0.0.1::" + strconv.Itoa(h.config.ContainerPort),
	}
	args = append(args, h.config.ContainerArgs...)
	args = append(args, h.config.Image)
	args = append(args, h.config.EngineArgs...)
	h.mu.Lock()
	h.containerKnown = true
	h.mu.Unlock()
	if _, err := h.config.Run.Run(ctx, h.config.DockerContext, args...); err != nil {
		return Endpoint{}, fmt.Errorf("start %s test container: %w", h.config.Engine, err)
	}

	mapping, err := h.config.Run.Run(ctx, h.config.DockerContext, "port", h.containerName, strconv.Itoa(h.config.ContainerPort)+"/tcp")
	if err != nil {
		return Endpoint{}, fmt.Errorf("discover %s test port: %w", h.config.Engine, err)
	}
	port, err := parseMappedPort(mapping, h.config.ContainerPort)
	if err != nil {
		return Endpoint{}, fmt.Errorf("parse %s test port: %w", h.config.Engine, err)
	}
	endpoint := Endpoint{Host: "127.0.0.1", Port: port}
	h.mu.Lock()
	h.endpoint = endpoint
	h.mu.Unlock()
	return endpoint, nil
}

// Close unconditionally attempts container, volume, image, and optional
// Colima reset cleanup in that order. Each later action still runs if an
// earlier one fails. A tagged test must call this in a defer before assertions.
func (h *Harness) Close(ctx context.Context) error {
	h.closeOnce.Do(func() {
		close(h.stopSignal)
		var errs []error
		if h.containerIsKnown() {
			if _, err := h.config.Run.Run(ctx, h.config.DockerContext, "container", "inspect", h.containerName); err == nil {
				if _, err := h.config.Run.Run(ctx, h.config.DockerContext, "container", "rm", "--force", h.containerName); err != nil {
					errs = append(errs, fmt.Errorf("remove %s test container: %w", h.config.Engine, err))
				}
			}
		}
		if h.volumeIsKnown() {
			if _, err := h.config.Run.Run(ctx, h.config.DockerContext, "volume", "rm", "--force", h.volumeName); err != nil {
				errs = append(errs, fmt.Errorf("remove %s test volume: %w", h.config.Engine, err))
			}
		}
		if h.imageWasPulled() && !h.config.KeepImage {
			if _, err := h.config.Run.Run(ctx, h.config.DockerContext, "image", "rm", h.config.Image); err != nil {
				errs = append(errs, fmt.Errorf("remove %s test image: %w", h.config.Engine, err))
			}
		}
		if h.config.ResetColima {
			if err := h.config.RunColima.Run(ctx, "delete", "--force", "--profile", h.config.ColimaProfile); err != nil {
				errs = append(errs, fmt.Errorf("reset Colima after %s test: %w", h.config.Engine, err))
			} else if err := h.config.RunColima.Run(ctx, "start", "--profile", h.config.ColimaProfile, "--runtime", "docker"); err != nil {
				errs = append(errs, fmt.Errorf("restart Colima after %s test: %w", h.config.Engine, err))
			} else {
				h.mu.Lock()
				h.report.ColimaReset = true
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

// Report returns the immutable aggregate disk measurement for this run.
func (h *Harness) Report() Report {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.report
}

func (h *Harness) containerIsKnown() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.containerKnown
}

func (h *Harness) volumeIsKnown() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.volumeKnown
}

func (h *Harness) imageWasPulled() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.pulledByRun
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

type dockerRunner struct{}

func (dockerRunner) Run(ctx context.Context, dockerContext string, args ...string) (string, error) {
	command := append([]string{"--context", dockerContext}, args...)
	cmd := exec.CommandContext(ctx, defaultDockerBinary, command...)
	output, err := cmd.Output()
	if err != nil {
		// Do not attach command arguments or captured output. Future engine
		// configurations may contain authentication material and test support
		// must never turn it into an error/logging channel.
		return "", fmt.Errorf("docker command failed: %w", err)
	}
	return string(output), nil
}

type colimaRunner struct{}

func (colimaRunner) Run(ctx context.Context, args ...string) error {
	if err := exec.CommandContext(ctx, defaultColimaBinary, args...).Run(); err != nil {
		// As for Docker, keep command output and values out of a test error.
		return fmt.Errorf("colima command failed: %w", err)
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
		return 0, errors.New("docker returned no host port")
	}
	port, err := strconv.Atoi(line[idx+1:])
	if err != nil || port < 1 || port > 65535 {
		return 0, errors.New("docker returned an invalid host port")
	}
	if port == defaultPort {
		return 0, errors.New("docker allocated the database default host port")
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

func safeProfile(value string) bool {
	if strings.TrimSpace(value) == "" {
		return false
	}
	for _, r := range value {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '-' && r != '_' {
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

func safeName(value string) string {
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
