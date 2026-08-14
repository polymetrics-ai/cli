package dbtest

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"
)

const testEndpoint = "unix:///tmp/dbtest.sock"

const (
	scriptedCapacityProbeID     = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	scriptedDatabaseContainerID = "1123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	scriptedSourceImageID       = "2123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	scriptedForeignImageID      = "3123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
)

type scriptedRunner struct {
	anonymousVolumePresent bool
	containerLive          bool
	capacityProbeOwners    map[string]string
	containerOwners        map[string]string
	raceForeignProbe       bool
	raceForeignContainer   bool
	raceForeignRunImage    bool
	retagAfterInspect      string
	pullErr                error
	tagErr                 error
	containerInspectErr    error
	runErr                 error
	infoErr                error
	infoOutput             string
	infoOutputs            []string
	infoCalls              int
	capacityOutput         string
	images                 map[string]bool
	imageIDs               map[string]string
	commands               [][]string
	endpoints              []string
}

type contextAwareRunner struct {
	scriptedRunner
}

func (r *contextAwareRunner) Run(ctx context.Context, endpoint string, args ...string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return r.scriptedRunner.Run(ctx, endpoint, args...)
}

func (r *scriptedRunner) Run(_ context.Context, endpoint string, args ...string) (string, error) {
	r.endpoints = append(r.endpoints, endpoint)
	r.commands = append(r.commands, append([]string(nil), args...))
	switch {
	case len(args) > 0 && args[0] == "pull":
		if len(args) > 1 {
			r.setImage(args[1], true)
		}
		return "", r.pullErr
	case commandHasPrefix(args, "image", "tag"):
		if len(args) > 3 {
			imageID := r.imageID(args[2])
			if r.raceForeignRunImage {
				imageID = scriptedForeignImageID
			}
			r.setImageWithID(args[3], imageID, true)
		}
		return "", r.tagErr
	case commandHasPrefix(args, "image", "rm"):
		if len(args) > 2 {
			r.setImage(args[2], false)
		}
	case commandHasPrefix(args, "image", "inspect"):
		reference := args[len(args)-1]
		if r.imagePresent(reference) {
			imageID := r.imageID(reference)
			if reference == r.retagAfterInspect {
				r.retagAfterInspect = ""
				r.setImageWithID(reference, scriptedForeignImageID, true)
			}
			if _, formatted := commandFlagValue(args, "--format"); formatted {
				return imageID + "\n", nil
			}
			return "", nil
		}
		return "", errContainerResourceNotFound
	case len(args) > 0 && args[0] == "volume":
		return "", errors.New("direct volume commands are not supported by the scripted harness")
	case commandHasPrefix(args, "container", "inspect"):
		if r.containerInspectErr != nil {
			return "", r.containerInspectErr
		}
		name := args[len(args)-1]
		if owner, present := r.capacityProbeOwners[name]; present {
			if _, formatted := commandFlagValue(args, "--format"); formatted {
				return scriptedCapacityProbeID + "\t" + owner + "\n", nil
			}
			return "", nil
		}
		if owner, present := r.containerOwners[name]; present {
			if _, formatted := commandFlagValue(args, "--format"); formatted {
				return scriptedDatabaseContainerID + "\t" + owner + "\n", nil
			}
			return "", nil
		}
		if r.containerLive {
			return "", nil
		}
		return "", errContainerResourceNotFound
	case len(args) > 0 && args[0] == "run":
		if strings.Contains(strings.Join(args, "\x00"), "\x00--entrypoint\x00/bin/df\x00") {
			if r.raceForeignProbe {
				if name, present := commandFlagValue(args, "--name"); present {
					r.setCapacityProbeOwner(name, "foreign")
				}
			}
			return r.capacityOutput, r.runErr
		}
		if name, present := commandFlagValue(args, "--name"); present {
			if r.raceForeignContainer {
				r.setContainerOwner(name, "foreign")
			} else if owner, present := commandLabelValue(args, databaseContainerOwnerLabel); present {
				r.setContainerOwner(name, owner)
				if _, present := commandFlagValue(args, "--volume"); present {
					r.anonymousVolumePresent = true
				}
			}
		}
		r.containerLive = true
		return "", r.runErr
	case commandHasPrefix(args, "container", "rm"):
		if len(args) > 2 {
			switch args[len(args)-1] {
			case scriptedCapacityProbeID:
				for name := range r.capacityProbeOwners {
					delete(r.capacityProbeOwners, name)
				}
			case scriptedDatabaseContainerID:
				for name := range r.containerOwners {
					delete(r.containerOwners, name)
				}
				r.containerLive = false
				if commandHasPrefix(args, "container", "rm", "--force", "--volumes", scriptedDatabaseContainerID) {
					r.anonymousVolumePresent = false
				}
			}
		}
	case len(args) > 0 && args[0] == "port":
		return "127.0.0.1:43123\n", nil
	case len(args) > 0 && args[0] == "info":
		if r.infoErr != nil {
			return "", r.infoErr
		}
		output := r.infoOutput
		if len(r.infoOutputs) > 0 {
			index := r.infoCalls
			if index >= len(r.infoOutputs) {
				index = len(r.infoOutputs) - 1
			}
			output = r.infoOutputs[index]
		}
		r.infoCalls++
		if output == "" {
			output = "/tmp/dbtest.sock\t/var/lib/containers/storage\t0\t0\n"
		}
		return output, nil
	}
	return "", nil
}

func (r *scriptedRunner) imagePresent(image string) bool {
	return r.images != nil && r.images[image]
}

func (r *scriptedRunner) setImage(image string, present bool) {
	r.setImageWithID(image, scriptedSourceImageID, present)
}

func (r *scriptedRunner) setImageWithID(image, imageID string, present bool) {
	if r.images == nil {
		r.images = make(map[string]bool)
	}
	if r.imageIDs == nil {
		r.imageIDs = make(map[string]string)
	}
	if present {
		r.images[image] = true
		r.imageIDs[image] = imageID
		return
	}
	delete(r.images, image)
	delete(r.imageIDs, image)
}

func (r *scriptedRunner) imageID(image string) string {
	if imageID, present := r.imageIDs[image]; present {
		return imageID
	}
	if safeImageID(image) {
		return image
	}
	return scriptedSourceImageID
}

func (r *scriptedRunner) setCapacityProbeOwner(name, owner string) {
	if r.capacityProbeOwners == nil {
		r.capacityProbeOwners = make(map[string]string)
	}
	r.capacityProbeOwners[name] = owner
}

func (r *scriptedRunner) setContainerOwner(name, owner string) {
	if r.containerOwners == nil {
		r.containerOwners = make(map[string]string)
	}
	r.containerOwners[name] = owner
}

func commandFlagValue(args []string, flag string) (string, bool) {
	for index := 0; index+1 < len(args); index++ {
		if args[index] == flag {
			return args[index+1], true
		}
	}
	return "", false
}

func commandLabelValue(args []string, label string) (string, bool) {
	prefix := label + "="
	for index := 0; index+1 < len(args); index++ {
		if args[index] == "--label" && strings.HasPrefix(args[index+1], prefix) {
			return strings.TrimPrefix(args[index+1], prefix), true
		}
	}
	return "", false
}

func commandHasPrefix(args []string, prefix ...string) bool {
	return len(args) >= len(prefix) && reflect.DeepEqual(args[:len(prefix)], prefix)
}

func commandsContain(commands [][]string, prefix ...string) bool {
	for _, command := range commands {
		if commandHasPrefix(command, prefix...) {
			return true
		}
	}
	return false
}

func commandIndex(commands [][]string, prefix ...string) int {
	for index, command := range commands {
		if commandHasPrefix(command, prefix...) {
			return index
		}
	}
	return -1
}

func testConfig(runner CommandRunner) Config {
	return Config{
		Engine:             "mysql",
		ContainerRuntime:   RuntimePodman,
		Image:              "example.invalid/mysql:8.4.11",
		ContainerPort:      3306,
		DataVolumePath:     "/var/lib/mysql",
		ContainerEndpoint:  testEndpoint,
		ExpectedImageBytes: 32,
		Run:                runner,
		DiskFreeAt: func(string) (uint64, error) {
			return 123, nil
		},
	}
}

func TestNewRequiresAnExplicitContainerRuntime(t *testing.T) {
	config := testConfig(&scriptedRunner{})
	config.ContainerRuntime = ""
	if _, err := New(config); err == nil {
		t.Fatal("New() accepted a database test configuration without an explicit Docker or Podman runtime")
	}
}

func TestNewAcceptsExplicitDockerRuntime(t *testing.T) {
	config := testConfig(&scriptedRunner{})
	config.ContainerRuntime = RuntimeDocker
	if _, err := New(config); err != nil {
		t.Fatalf("New() rejected an explicit Docker runtime: %v", err)
	}
}

func TestNewRejectsUnknownContainerRuntime(t *testing.T) {
	config := testConfig(&scriptedRunner{})
	config.ContainerRuntime = "colima"
	if _, err := New(config); err == nil {
		t.Fatal("New() accepted an unknown container runtime")
	}
}

func TestRuntimeCommandPinsTheConfiguredEndpoint(t *testing.T) {
	for _, tc := range []struct {
		runtime Runtime
		binary  string
		flag    string
	}{
		{runtime: RuntimeDocker, binary: defaultDockerBinary, flag: "--host"},
		{runtime: RuntimePodman, binary: defaultPodmanBinary, flag: "--url"},
	} {
		t.Run(string(tc.runtime), func(t *testing.T) {
			binary, args := tc.runtime.command(testEndpoint, "info")
			if binary != tc.binary {
				t.Fatalf("%s binary = %q, want %q", tc.runtime, binary, tc.binary)
			}
			want := []string{tc.flag, testEndpoint, "info"}
			if !reflect.DeepEqual(args, want) {
				t.Fatalf("%s arguments = %v, want %v", tc.runtime, args, want)
			}
		})
	}
}

func TestNewRejectsUnsafeEndpoints(t *testing.T) {
	for _, tc := range []struct {
		name     string
		endpoint string
	}{
		{name: "empty", endpoint: ""},
		{name: "named connection", endpoint: "lane-machine"},
		{name: "remote ssh endpoint", endpoint: "ssh://core@127.0.0.1/run/podman.sock"},
		{name: "remote TCP endpoint", endpoint: "tcp://127.0.0.1:1234"},
		{name: "hosted Unix endpoint", endpoint: "unix://localhost/tmp/dbtest.sock"},
		{name: "newline", endpoint: "unix:///tmp/dbtest.sock\n"},
	} {
		for _, runtime := range []Runtime{RuntimeDocker, RuntimePodman} {
			t.Run(string(runtime)+"/"+tc.name, func(t *testing.T) {
				config := testConfig(&scriptedRunner{})
				config.ContainerRuntime = runtime
				config.ContainerEndpoint = tc.endpoint
				if _, err := New(config); err == nil {
					t.Fatal("New() accepted an unsafe or non-local container endpoint")
				}
			})
		}
	}
}

func TestNewRejectsUnpinnedImages(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*Config)
	}{
		{name: "latest image", mutate: func(c *Config) { c.Image = "example.invalid/mysql:latest" }},
		{name: "unversioned image", mutate: func(c *Config) { c.Image = "example.invalid/mysql" }},
		{name: "digest form", mutate: func(c *Config) { c.Image = "example.invalid/mysql@sha256:abc" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config := testConfig(&scriptedRunner{})
			tc.mutate(&config)
			if _, err := New(config); err == nil {
				t.Fatal("New() accepted unsafe configuration")
			}
		})
	}
}

func TestNewRejectsUnsafeDockerCapacityProbeImages(t *testing.T) {
	for _, image := range []string{
		"example.invalid/probe:latest",
		"example.invalid/probe",
		"example.invalid/probe@sha256:abc",
		" example.invalid/probe:1.0",
	} {
		config := testConfig(&scriptedRunner{})
		config.DockerCapacityProbeImage = image
		if _, err := New(config); err == nil {
			t.Fatalf("New() accepted unsafe Docker capacity probe image %q", image)
		}
	}
}

func TestCleanupRemovesOnlyRunOwnedResources(t *testing.T) {
	runner := &scriptedRunner{}
	config := testConfig(runner)
	h, err := New(config)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	endpoint, err := h.Start(context.Background())
	if err != nil {
		t.Fatalf("Start(): %v", err)
	}
	if endpoint.Host != "127.0.0.1" || endpoint.Port == 3306 {
		t.Fatalf("endpoint = %+v, want loopback non-default port", endpoint)
	}
	if !runner.imagePresent(config.Image) || !runner.imagePresent(h.runImage) {
		t.Fatalf("images after start = %#v, want source and generated references", runner.images)
	}
	if h.containerID != scriptedDatabaseContainerID {
		t.Fatalf("container ID = %q, want the verified immutable ID", h.containerID)
	}
	if h.runImageID != scriptedSourceImageID {
		t.Fatalf("run image ID = %q, want the verified immutable source ID", h.runImageID)
	}
	if !commandsContain(runner.commands, "image", "tag", scriptedSourceImageID, h.runImage) {
		t.Fatalf("commands = %v, want an immutable source image tag", runner.commands)
	}
	if !commandsContain(runner.commands, "image", "inspect", "--format", imageIDFormat, h.runImage) {
		t.Fatalf("commands = %v, want generated-image identity verification", runner.commands)
	}
	if !commandsContain(runner.commands, "run", "--detach", "--name", h.containerName, "--volume", config.DataVolumePath) {
		t.Fatalf("commands = %v, want a container-bound anonymous data volume", runner.commands)
	}
	if commandsContain(runner.commands, "volume") {
		t.Fatalf("commands = %v, want no standalone volume command", runner.commands)
	}
	if !commandsContain(runner.commands, "container", "inspect", "--format", databaseContainerOwnerFormat, h.containerName) {
		t.Fatalf("commands = %v, want normal-container ownership inspection", runner.commands)
	}
	if !commandsContain(runner.commands, "port", scriptedDatabaseContainerID, "3306/tcp") {
		t.Fatalf("commands = %v, want port lookup by immutable container ID", runner.commands)
	}
	if err := h.Close(context.Background()); err != nil {
		t.Fatalf("Close(): %v", err)
	}
	if runner.containerLive {
		t.Fatal("cleanup retained the generated database container")
	}
	if runner.anonymousVolumePresent {
		t.Fatal("cleanup retained the generated anonymous database volume")
	}
	if !commandsContain(runner.commands, "container", "rm", "--force", "--volumes", scriptedDatabaseContainerID) {
		t.Fatalf("commands = %v, want database-container and anonymous-volume removal by immutable ID", runner.commands)
	}
	if commandsContain(runner.commands, "volume") {
		t.Fatalf("commands = %v, want no standalone volume cleanup", runner.commands)
	}
	if !runner.imagePresent(h.runImage) {
		t.Fatalf("images after cleanup = %#v, want the generated reference retained", runner.images)
	}
	if !runner.imagePresent(config.Image) {
		t.Fatalf("images after cleanup = %#v, want the source image retained", runner.images)
	}
	if commandsContain(runner.commands, "image", "rm") {
		t.Fatalf("commands = %v, want no mutable image-reference cleanup", runner.commands)
	}
	var gotCleanup [][]string
	for _, command := range runner.commands {
		if commandHasPrefix(command, "container", "rm") {
			gotCleanup = append(gotCleanup, command[:2])
		}
	}
	want := [][]string{{"container", "rm"}}
	if !reflect.DeepEqual(gotCleanup, want) {
		t.Fatalf("cleanup order = %v, want %v", gotCleanup, want)
	}
	for _, endpoint := range runner.endpoints {
		if endpoint != config.ContainerEndpoint {
			t.Fatalf("Podman command used endpoint %q, want the configured endpoint", endpoint)
		}
	}
	report := h.Report()
	if report.DiskFreeBefore != 123 || report.DiskFreeAfter != 123 {
		t.Fatalf("disk report = %+v, want target measurements", report)
	}
}

func TestStartFailsClosedWhenCachedTargetIdentityCannotBeProven(t *testing.T) {
	runner := &scriptedRunner{infoOutputs: []string{
		"/tmp/dbtest.sock\t/var/lib/containers/storage\t0\t0\n",
		"/tmp/other.sock\t/var/lib/containers/storage\t0\t0\n",
	}}
	config := testConfig(runner)
	runner.setImage(config.Image, true)
	h, err := New(config)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if _, err := h.Start(context.Background()); err == nil {
		t.Fatal("Start() accepted a cached image after the target identity changed")
	}
	for _, prefix := range [][]string{{"pull"}, {"image", "tag"}, {"run"}} {
		if commandsContain(runner.commands, prefix...) {
			t.Fatalf("commands = %v, want no mutation after identity proof failed", runner.commands)
		}
	}
}

func TestStartFailsClosedWhenCachedTargetCapacityCannotBeProven(t *testing.T) {
	runner := &scriptedRunner{}
	config := testConfig(runner)
	runner.setImage(config.Image, true)
	config.DiskFreeAt = func(string) (uint64, error) {
		return 0, errors.New("unavailable")
	}
	h, err := New(config)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	startErr := func() error {
		_, err := h.Start(context.Background())
		return err
	}()
	if startErr == nil || !strings.Contains(startErr.Error(), "capacity") {
		t.Fatalf("Start() error = %v, want a target capacity failure", startErr)
	}
	for _, prefix := range [][]string{{"pull"}, {"image", "tag"}, {"run"}} {
		if commandsContain(runner.commands, prefix...) {
			t.Fatalf("commands = %v, want no mutation before target capacity is proven", runner.commands)
		}
	}
}

func TestEveryTargetCommandRechecksTargetIdentity(t *testing.T) {
	runner := &scriptedRunner{}
	h, err := New(testConfig(runner))
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if _, err := h.Start(context.Background()); err != nil {
		t.Fatalf("Start(): %v", err)
	}
	if err := h.Close(context.Background()); err != nil {
		t.Fatalf("Close(): %v", err)
	}
	for index, command := range runner.commands {
		if commandHasPrefix(command, "info") {
			continue
		}
		if index == 0 || !commandHasPrefix(runner.commands[index-1], "info") {
			t.Fatalf("target command %v was not immediately preceded by an identity check: %v", command, runner.commands)
		}
	}
}

func TestCleanupFailsClosedWhenTargetIdentityChanges(t *testing.T) {
	for _, changedIdentity := range []string{
		"/tmp/other.sock\t/var/lib/containers/storage\t0\t0\n",
		"/tmp/dbtest.sock\t/var/lib/other-storage\t0\t0\n",
	} {
		t.Run(changedIdentity, func(t *testing.T) {
			runner := &scriptedRunner{}
			h, err := New(testConfig(runner))
			if err != nil {
				t.Fatalf("New(): %v", err)
			}
			if _, err := h.Start(context.Background()); err != nil {
				t.Fatalf("Start(): %v", err)
			}
			runner.infoOutput = changedIdentity
			if err := h.Close(context.Background()); err == nil {
				t.Fatal("Close() removed resources after target identity changed")
			}
			for _, prefix := range [][]string{{"container", "rm"}} {
				if commandsContain(runner.commands, prefix...) {
					t.Fatalf("commands = %v, want no cleanup mutation on an unproven endpoint", runner.commands)
				}
			}
		})
	}
}

func TestStartPlacesEngineArgumentsAfterImage(t *testing.T) {
	runner := &scriptedRunner{}
	config := testConfig(runner)
	config.ContainerArgs = []string{"--env", "SAFE_OPTION=value"}
	config.EngineArgs = []string{"--server-id=731000"}
	h, err := New(config)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if _, err := h.Start(context.Background()); err != nil {
		t.Fatalf("Start(): %v", err)
	}
	defer func() { _ = h.Close(context.Background()) }()
	for _, command := range runner.commands {
		if len(command) == 0 || command[0] != "run" {
			continue
		}
		imageIndex, engineArgumentIndex := -1, -1
		for index, arg := range command {
			if arg == h.runImageID {
				imageIndex = index
			}
			if arg == "--server-id=731000" {
				engineArgumentIndex = index
			}
		}
		if imageIndex < 0 || engineArgumentIndex != imageIndex+1 {
			t.Fatalf("run arguments = %v, want engine command after image", command)
		}
		return
	}
	t.Fatal("Start did not issue a Podman run command")
}

func TestCopyFileFromContainerUsesTheConfiguredEndpoint(t *testing.T) {
	runner := &scriptedRunner{}
	config := testConfig(runner)
	h, err := New(config)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if _, err := h.Start(context.Background()); err != nil {
		t.Fatalf("Start(): %v", err)
	}
	defer func() { _ = h.Close(context.Background()) }()
	destination := filepath.Join(t.TempDir(), "ca.pem")
	if err := h.CopyFileFromContainer(context.Background(), "/var/lib/mysql/ca.pem", destination); err != nil {
		t.Fatalf("CopyFileFromContainer(): %v", err)
	}
	if !commandsContain(runner.commands, "cp", h.containerID+":"+"/var/lib/mysql/ca.pem", destination) {
		t.Fatalf("commands = %v, want scoped container certificate copy", runner.commands)
	}
	for _, endpoint := range runner.endpoints {
		if endpoint != config.ContainerEndpoint {
			t.Fatalf("endpoint = %q, want %q", endpoint, config.ContainerEndpoint)
		}
	}
}

func TestStartPublishesOnlyLoopback(t *testing.T) {
	runner := &scriptedRunner{}
	h, err := New(testConfig(runner))
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if _, err := h.Start(context.Background()); err != nil {
		t.Fatalf("Start(): %v", err)
	}
	defer func() { _ = h.Close(context.Background()) }()
	for _, command := range runner.commands {
		if len(command) == 0 || command[0] != "run" {
			continue
		}
		for index, arg := range command {
			if arg != "--publish" {
				continue
			}
			if want := "127.0.0.1::3306"; command[index+1] != want {
				t.Fatalf("publish argument = %q, want %q", command[index+1], want)
			}
			return
		}
	}
	t.Fatal("Start did not publish the engine port on loopback")
}

func TestStartCleansResourcesAfterIndeterminateEngineOutcomes(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		runner               *scriptedRunner
		wantContainerCleanup bool
	}{
		{name: "pull returns an error before any generated reference is owned", runner: &scriptedRunner{pullErr: context.Canceled}},
		{name: "tag returns an error after the pull succeeded", runner: &scriptedRunner{tagErr: context.Canceled}},
		{name: "container run returns an error after creating the container", runner: &scriptedRunner{runErr: context.Canceled}, wantContainerCleanup: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config := testConfig(tc.runner)
			h, err := New(config)
			if err != nil {
				t.Fatalf("New(): %v", err)
			}
			if _, err := h.Start(context.Background()); err == nil {
				t.Fatal("Start() succeeded after an indeterminate engine outcome")
			}
			containerCleanup := commandsContain(tc.runner.commands, "container", "rm", "--force", "--volumes", scriptedDatabaseContainerID)
			if containerCleanup != tc.wantContainerCleanup {
				t.Fatalf("container cleanup = %t, want %t (commands = %v)", containerCleanup, tc.wantContainerCleanup, tc.runner.commands)
			}
			if commandsContain(tc.runner.commands, "volume") {
				t.Fatalf("commands = %v, want no standalone volume command", tc.runner.commands)
			}
			if commandsContain(tc.runner.commands, "image", "rm") {
				t.Fatalf("commands = %v, want no mutable image-reference cleanup", tc.runner.commands)
			}
			if tc.wantContainerCleanup && tc.runner.anonymousVolumePresent {
				t.Fatal("container-bound anonymous volume remained after container cleanup")
			}
		})
	}
}

func TestStartPreservesARunImageRetaggedAfterCreation(t *testing.T) {
	runner := &scriptedRunner{raceForeignRunImage: true}
	h, err := New(testConfig(runner))
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if _, err := h.Start(context.Background()); err == nil || !strings.Contains(err.Error(), "does not refer to the configured source image") {
		t.Fatalf("Start() error = %v, want a retagged run-image refusal", err)
	}
	if imageID := runner.imageID(h.runImage); imageID != scriptedForeignImageID {
		t.Fatalf("run image ID = %q, want foreign", imageID)
	}
	if !runner.imagePresent(h.runImage) {
		t.Fatal("Start() removed the retagged foreign run image")
	}
	if commandsContain(runner.commands, "image", "rm", h.runImage) {
		t.Fatalf("commands = %v, must not remove a retagged foreign run image", runner.commands)
	}
	if commandsContain(runner.commands, "run") {
		t.Fatalf("commands = %v, must not start a container from an unproven run image", runner.commands)
	}
}

func TestCloseRetainsRunImageReference(t *testing.T) {
	runner := &scriptedRunner{}
	h, err := New(testConfig(runner))
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if _, err := h.Start(context.Background()); err != nil {
		t.Fatalf("Start(): %v", err)
	}
	runner.setImageWithID(h.runImage, scriptedForeignImageID, true)
	runner.commands = nil

	if err := h.Close(context.Background()); err != nil {
		t.Fatalf("Close(): %v", err)
	}
	if imageID := runner.imageID(h.runImage); imageID != scriptedForeignImageID {
		t.Fatalf("run image ID = %q, want foreign", imageID)
	}
	if commandsContain(runner.commands, "image") {
		t.Fatalf("commands = %v, must retain the mutable run-image reference", runner.commands)
	}
}

func TestStartPreservesAContainerCreatedAfterPrecheck(t *testing.T) {
	runner := &scriptedRunner{
		raceForeignContainer: true,
		runErr:               errors.New("name already in use"),
	}
	h, err := New(testConfig(runner))
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if _, err := h.Start(context.Background()); err == nil || !strings.Contains(err.Error(), "start mysql test container") {
		t.Fatalf("Start() error = %v, want a raced container creation failure", err)
	}
	if owner := runner.containerOwners[h.containerName]; owner != "foreign" {
		t.Fatalf("container owner = %q, want foreign", owner)
	}
	if !runner.containerLive {
		t.Fatal("Start() removed the foreign container created after its name precheck")
	}
	if !commandsContain(runner.commands, "container", "inspect", "--format", databaseContainerOwnerFormat, h.containerName) {
		t.Fatalf("commands = %v, want normal-container ownership inspection", runner.commands)
	}
	if commandsContain(runner.commands, "container", "rm", "--force") {
		t.Fatalf("commands = %v, must not remove a raced foreign container", runner.commands)
	}
}

func TestStartReleasesTheEngineSlotOnEveryExitPath(t *testing.T) {
	for _, tc := range []struct {
		name   string
		runner *scriptedRunner
	}{
		{name: "successful run", runner: &scriptedRunner{}},
		{name: "failed start", runner: &scriptedRunner{runErr: context.Canceled}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, err := New(testConfig(tc.runner))
			if err != nil {
				t.Fatalf("New(): %v", err)
			}
			if _, err := h.Start(context.Background()); err == nil {
				_ = h.Close(context.Background())
			}
			select {
			case engineSlots <- struct{}{}:
				<-engineSlots
			default:
				t.Fatal("engine slot was not released after the run finished")
			}
		})
	}
}

func TestStartUsesAnImmutableSourceImageID(t *testing.T) {
	runner := &scriptedRunner{}
	config := testConfig(runner)
	config.ExpectedImageBytes = 830 << 20
	config.DiskFreeAt = func(string) (uint64, error) { return 8 << 30, nil }
	runner.setImage(config.Image, true)
	h, err := New(config)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if _, err := h.Start(context.Background()); err != nil {
		t.Fatalf("Start(): %v", err)
	}
	if err := h.Close(context.Background()); err != nil {
		t.Fatalf("Close(): %v", err)
	}
	if !commandsContain(runner.commands, "pull", config.Image) || !commandsContain(runner.commands, "image", "tag", scriptedSourceImageID, h.runImage) {
		t.Fatalf("commands = %v, want a generated image reference from the immutable source", runner.commands)
	}
	if commandsContain(runner.commands, "image", "rm", config.Image) {
		t.Fatalf("commands = %v, want no source-image removal", runner.commands)
	}
}

func TestContainerResourceNotFoundClassifiesOnlyKnownAbsence(t *testing.T) {
	for _, stderr := range []string{
		`Error: no such container "x"`,
		`Error: no such volume "x"`,
		"Error: x: image not known",
		"Error response from daemon: No such image",
	} {
		if !containerResourceNotFound(&exec.ExitError{Stderr: []byte(stderr)}) {
			t.Fatalf("containerResourceNotFound(%q) = false, want recognized absence", stderr)
		}
	}
	for _, stderr := range []string{"Error response from daemon: access denied", "Error: unable to connect to Podman socket"} {
		if containerResourceNotFound(&exec.ExitError{Stderr: []byte(stderr)}) {
			t.Fatalf("containerResourceNotFound(%q) = true, want indeterminate error", stderr)
		}
	}
}

func TestStartUsesContainerBoundAnonymousVolume(t *testing.T) {
	runner := &scriptedRunner{}
	config := testConfig(runner)
	h, err := New(config)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if _, err := h.Start(context.Background()); err != nil {
		t.Fatalf("Start(): %v", err)
	}
	if !runner.anonymousVolumePresent {
		t.Fatal("Start() did not create the container-bound anonymous volume")
	}
	if !commandsContain(runner.commands, "run", "--detach", "--name", h.containerName, "--volume", config.DataVolumePath) {
		t.Fatalf("commands = %v, want an anonymous data-volume mount", runner.commands)
	}
	if commandsContain(runner.commands, "volume") {
		t.Fatalf("commands = %v, want no standalone volume command", runner.commands)
	}
	if err := h.Close(context.Background()); err != nil {
		t.Fatalf("Close(): %v", err)
	}
	if runner.anonymousVolumePresent {
		t.Fatal("Close() did not remove the container-bound anonymous volume")
	}
	if !commandsContain(runner.commands, "container", "rm", "--force", "--volumes", scriptedDatabaseContainerID) {
		t.Fatalf("commands = %v, want anonymous-volume cleanup through the immutable container ID", runner.commands)
	}
}

func TestCloseRemovesAnonymousVolumeWithOwnedContainer(t *testing.T) {
	runner := &scriptedRunner{}
	h, err := New(testConfig(runner))
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	runner.setContainerOwner(h.containerName, h.containerOwner)
	runner.containerLive = true
	runner.anonymousVolumePresent = true
	h.mu.Lock()
	h.containerKnown = true
	h.mu.Unlock()

	if err := h.Close(context.Background()); err != nil {
		t.Fatalf("Close(): %v", err)
	}
	if runner.anonymousVolumePresent {
		t.Fatal("Close() did not remove the container-bound anonymous volume")
	}
	if !commandsContain(runner.commands, "container", "rm", "--force", "--volumes", scriptedDatabaseContainerID) {
		t.Fatalf("commands = %v, want ownership-proven container cleanup with volumes", runner.commands)
	}
	if commandsContain(runner.commands, "volume") {
		t.Fatalf("commands = %v, must not remove a volume by name", runner.commands)
	}
}

func TestInterruptCleanupClosesEveryLiveHarnessBeforeExiting(t *testing.T) {
	SetMaxConcurrentEngines(2)
	defer SetMaxConcurrentEngines(1)
	runners := []*scriptedRunner{{}, {}}
	harnesses := make([]*Harness, 0, len(runners))
	for _, runner := range runners {
		h, err := New(testConfig(runner))
		if err != nil {
			t.Fatalf("New(): %v", err)
		}
		if _, err := h.Start(context.Background()); err != nil {
			t.Fatalf("Start(): %v", err)
		}
		harnesses = append(harnesses, h)
	}
	closeLiveHarnesses(context.Background())
	for index, runner := range runners {
		if runner.containerLive || runner.anonymousVolumePresent {
			t.Fatalf("harness %d retained a container-bound resource through interrupt cleanup", index)
		}
		if !runner.imagePresent(harnesses[index].runImage) {
			t.Fatalf("harness %d did not retain its generated image reference", index)
		}
	}
	for _, h := range harnesses {
		if err := h.Close(context.Background()); err != nil {
			t.Fatalf("Close(): %v", err)
		}
	}
}

func TestInterruptDrainKeepsWatchingAndRejectsNewStarts(t *testing.T) {
	firstRunner := &scriptedRunner{}
	first, err := New(testConfig(firstRunner))
	if err != nil {
		t.Fatalf("New(first): %v", err)
	}
	if _, err := first.Start(context.Background()); err != nil {
		t.Fatalf("Start(first): %v", err)
	}

	secondRunner := &scriptedRunner{}
	second, err := New(testConfig(secondRunner))
	if err != nil {
		t.Fatalf("New(second): %v", err)
	}
	interruptCleanup.mu.Lock()
	signals := interruptCleanup.signals
	interruptCleanup.mu.Unlock()
	if signals == nil {
		t.Fatal("Start(first) did not arm interrupt cleanup")
	}
	live := beginInterruptDrain()
	defer func() {
		closeHarnesses(context.Background(), live)
		_ = second.Close(context.Background())
		finishInterruptDrain(signals)
	}()
	if len(live) != 1 || live[0] != first {
		t.Fatalf("interrupt drain snapshot = %#v, want first harness", live)
	}

	closeHarnesses(context.Background(), live)
	interruptCleanup.mu.Lock()
	stillWatching := interruptCleanup.signals == signals
	draining := interruptCleanup.draining
	interruptCleanup.mu.Unlock()
	if !stillWatching || !draining {
		t.Fatalf("interrupt cleanup state after teardown: watching=%t draining=%t", stillWatching, draining)
	}
	if _, err := second.Start(context.Background()); err == nil {
		t.Fatal("Start(second) succeeded after interrupt draining began")
	}
	if len(secondRunner.commands) != 0 {
		t.Fatalf("Start(second) issued Podman commands after interrupt draining: %v", secondRunner.commands)
	}
}

func TestWaitForInterruptCleanupAbsorbsRepeatedSignals(t *testing.T) {
	signals := make(chan os.Signal)
	done := make(chan struct{})
	returned := make(chan struct{})
	go func() {
		waitForInterruptCleanup(signals, done)
		close(returned)
	}()
	signals <- os.Interrupt
	signals <- syscall.SIGTERM
	select {
	case <-returned:
		t.Fatal("interrupt cleanup stopped before teardown completed")
	default:
	}
	close(done)
	select {
	case <-returned:
	case <-time.After(time.Second):
		t.Fatal("interrupt cleanup did not finish after teardown")
	}
}

func TestSetMaxConcurrentEnginesRefusesAChangeWhileASlotIsHeld(t *testing.T) {
	h, err := New(testConfig(&scriptedRunner{}))
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if _, err := h.Start(context.Background()); err != nil {
		t.Fatalf("Start(): %v", err)
	}
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Error("SetMaxConcurrentEngines() replaced the slot channel while a harness held a token")
		}
		if err := h.Close(context.Background()); err != nil {
			t.Errorf("Close(): %v", err)
		}
	}()
	SetMaxConcurrentEngines(2)
}

func TestParseMappedPortRefusesTheEngineDefaultPort(t *testing.T) {
	if _, err := parseMappedPort("127.0.0.1:3306\n", 3306); err == nil {
		t.Fatal("parseMappedPort() accepted the engine default host port")
	}
	port, err := parseMappedPort("127.0.0.1:43123\n", 3306)
	if err != nil || port != 43123 {
		t.Fatalf("parseMappedPort() = %d, %v, want 43123", port, err)
	}
}

type interruptingRunner struct {
	scriptedRunner
	harness  *Harness
	on       string
	fired    bool
	done     chan struct{}
	closeErr error
}

func (r *interruptingRunner) Run(ctx context.Context, endpoint string, args ...string) (string, error) {
	if !r.fired && len(args) > 0 && args[0] == r.on {
		r.fired = true
		r.done = make(chan struct{})
		go func() {
			defer close(r.done)
			r.closeErr = r.harness.Close(context.Background())
		}()
		select {
		case <-r.done:
			r.closeErr = errors.New("Close finished while a create command was still running")
		case <-time.After(25 * time.Millisecond):
		}
	}
	return r.scriptedRunner.Run(ctx, endpoint, args...)
}

func TestCloseDuringCreateStillRemovesTheCreatedResource(t *testing.T) {
	for _, step := range []string{"run"} {
		t.Run("interrupted during "+step, func(t *testing.T) {
			runner := &interruptingRunner{on: step}
			h, err := New(testConfig(runner))
			if err != nil {
				t.Fatalf("New(): %v", err)
			}
			runner.harness = h
			if _, err := h.Start(context.Background()); err == nil {
				t.Fatal("Start() succeeded after the harness was closed mid-startup")
			}
			if runner.done == nil {
				t.Fatalf("the %q step never ran, so the race was not exercised", step)
			}
			select {
			case <-runner.done:
			case <-time.After(30 * time.Second):
				t.Fatal("Close did not return after Start finished")
			}
			if runner.closeErr != nil {
				t.Fatalf("Close() during startup: %v", runner.closeErr)
			}
			if runner.containerLive || runner.anonymousVolumePresent {
				t.Fatal("interrupt left a container-bound resource behind")
			}
			if !runner.imagePresent(h.runImage) {
				t.Fatal("interrupt did not retain the generated image reference")
			}
			select {
			case engineSlots <- struct{}{}:
				<-engineSlots
			default:
				t.Fatal("interrupt leaked the engine slot")
			}
		})
	}
}

type teardownProbeRunner struct {
	scriptedRunner
	harness *Harness
	probes  []teardownProbe
}

type teardownProbe struct {
	command    string
	registered bool
	armed      bool
}

func (r *teardownProbeRunner) Run(ctx context.Context, endpoint string, args ...string) (string, error) {
	if len(args) > 1 && args[1] == "rm" {
		interruptCleanup.mu.Lock()
		_, registered := interruptCleanup.live[r.harness]
		armed := interruptCleanup.signals != nil
		interruptCleanup.mu.Unlock()
		r.probes = append(r.probes, teardownProbe{command: args[0] + " " + args[1], registered: registered, armed: armed})
	}
	return r.scriptedRunner.Run(ctx, endpoint, args...)
}

func TestCloseKeepsInterruptCleanupArmedUntilTeardownFinishes(t *testing.T) {
	runner := &teardownProbeRunner{}
	h, err := New(testConfig(runner))
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	runner.harness = h
	if _, err := h.Start(context.Background()); err != nil {
		t.Fatalf("Start(): %v", err)
	}
	if err := h.Close(context.Background()); err != nil {
		t.Fatalf("Close(): %v", err)
	}
	if len(runner.probes) != 1 {
		t.Fatalf("teardown probes = %+v, want only immutable-container cleanup", runner.probes)
	}
	for _, probe := range runner.probes {
		if !probe.registered || !probe.armed {
			t.Fatalf("%q ran outside the interrupt cleanup: registered=%t armed=%t", probe.command, probe.registered, probe.armed)
		}
	}
	interruptCleanup.mu.Lock()
	_, registered := interruptCleanup.live[h]
	interruptCleanup.mu.Unlock()
	if registered {
		t.Fatal("Close left the finished harness registered for interrupt cleanup")
	}
}

func TestCloseCleansUpWithCanceledCallerContext(t *testing.T) {
	runner := &contextAwareRunner{}
	h, err := New(testConfig(runner))
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if _, err := h.Start(context.Background()); err != nil {
		t.Fatalf("Start(): %v", err)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := h.Close(canceled); err != nil {
		t.Fatalf("Close() with canceled context: %v", err)
	}
	if runner.containerLive || runner.anonymousVolumePresent {
		t.Fatal("Close() with canceled context left a container-bound resource behind")
	}
	if !runner.imagePresent(h.runImage) {
		t.Fatal("Close() with canceled context did not retain the generated image reference")
	}
	interruptCleanup.mu.Lock()
	_, registered := interruptCleanup.live[h]
	interruptCleanup.mu.Unlock()
	if registered {
		t.Fatal("Close() with canceled context left interrupt cleanup registered")
	}
}

func TestNewRequiresTheImageFootprint(t *testing.T) {
	for _, bytes := range []uint64{0, maxExpectedImageBytes + 1} {
		config := testConfig(&scriptedRunner{})
		config.ExpectedImageBytes = bytes
		if _, err := New(config); err == nil {
			t.Fatalf("New() accepted an image footprint of %d, which cannot gate a pull", bytes)
		}
	}
}

func TestStartRefusesToPullWithoutHeadroomForTheImage(t *testing.T) {
	const imageBytes = 830 << 20
	for _, tc := range []struct {
		name     string
		free     uint64
		cached   bool
		wantPull bool
	}{
		{name: "free space barely exceeds the image itself", free: 854 << 20},
		{name: "free space clears the documented headroom", free: imagePullFreeSpaceFactor * imageBytes, wantPull: true},
		{name: "cached image needs no headroom after capacity proof", free: 854 << 20, cached: true, wantPull: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runner := &scriptedRunner{}
			config := testConfig(runner)
			config.ExpectedImageBytes = imageBytes
			config.DiskFreeAt = func(string) (uint64, error) { return tc.free, nil }
			if tc.cached {
				runner.setImage(config.Image, true)
			}
			h, err := New(config)
			if err != nil {
				t.Fatalf("New(): %v", err)
			}
			_, startErr := h.Start(context.Background())
			if err := h.Close(context.Background()); err != nil {
				t.Fatalf("Close(): %v", err)
			}
			if tc.wantPull && startErr != nil {
				t.Fatalf("Start(): %v", startErr)
			}
			if !tc.wantPull && startErr == nil {
				t.Fatal("Start() pulled an image the target had no room for")
			}
			if pulled := commandsContain(runner.commands, "pull", config.Image); pulled != tc.wantPull {
				t.Fatalf("pull issued = %t, want %t (commands = %v)", pulled, tc.wantPull, runner.commands)
			}
			if !tc.wantPull && (runner.containerLive || runner.anonymousVolumePresent) {
				t.Fatal("a refused start left a container or volume behind")
			}
		})
	}
}

func TestTargetCapacityUsesTheProvenStorePath(t *testing.T) {
	runner := &scriptedRunner{}
	config := testConfig(runner)
	var paths []string
	config.DiskFreeAt = func(path string) (uint64, error) {
		paths = append(paths, path)
		return 1 << 30, nil
	}
	h, err := New(config)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if _, err := h.Start(context.Background()); err != nil {
		t.Fatalf("Start(): %v", err)
	}
	if err := h.Close(context.Background()); err != nil {
		t.Fatalf("Close(): %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("target capacity was never measured")
	}
	for _, path := range paths {
		if path != "/var/lib/containers/storage" {
			t.Fatalf("capacity path = %q, want the proven image-store path", path)
		}
	}
}

func TestStartUsesDockerTargetIdentityAndCapacity(t *testing.T) {
	runner := &scriptedRunner{infoOutput: "DAEMON:12345\t/var/lib/docker\n"}
	config := testConfig(runner)
	config.ContainerRuntime = RuntimeDocker
	var paths []string
	config.DiskFreeAt = func(path string) (uint64, error) {
		paths = append(paths, path)
		return 1 << 30, nil
	}
	h, err := New(config)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if _, err := h.Start(context.Background()); err != nil {
		t.Fatalf("Start(): %v", err)
	}
	if err := h.Close(context.Background()); err != nil {
		t.Fatalf("Close(): %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("Docker target capacity was never measured")
	}
	for _, path := range paths {
		if path != "/var/lib/docker" {
			t.Fatalf("Docker capacity path = %q, want the proven image-store path", path)
		}
	}
	if !commandsContain(runner.commands, "info", "--format", RuntimeDocker.infoFormat()) {
		t.Fatalf("commands = %v, want Docker identity inspection", runner.commands)
	}
}

func TestDockerVMCapacityUsesOnlyAPreCachedLockedDownProbe(t *testing.T) {
	runner := &scriptedRunner{
		infoOutput:     "DAEMON:12345\t/var/lib/docker\n",
		capacityOutput: "Filesystem 1-blocks Used Available Capacity Mounted on\n/dev/root 1000 800 123 80% /polymetrics-image-store\n",
	}
	config := testConfig(runner)
	config.ContainerRuntime = RuntimeDocker
	config.DockerCapacityProbeImage = "example.invalid/capacity-probe:1.0"
	runner.setImage(config.DockerCapacityProbeImage, true)
	config.DiskFreeAt = func(string) (uint64, error) {
		return 0, errors.New("Docker store belongs to a VM")
	}
	h, err := New(config)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}

	free, err := h.targetImageStoreFree(context.Background())
	if err != nil {
		t.Fatalf("targetImageStoreFree(): %v", err)
	}
	if free != 123 {
		t.Fatalf("targetImageStoreFree() = %d, want daemon-side 123", free)
	}
	if commandsContain(runner.commands, "pull", config.DockerCapacityProbeImage) {
		t.Fatalf("commands = %v, must never pull the capacity probe", runner.commands)
	}
	if !commandsContain(runner.commands, "image", "inspect", "--format", imageIDFormat, config.DockerCapacityProbeImage) {
		t.Fatalf("commands = %v, want immutable capacity-probe image inspection", runner.commands)
	}
	if !commandsContain(runner.commands,
		"run", "--pull=never", "--rm", "--name", h.capacityProbeName,
		"--label", dockerCapacityProbeOwnerLabel+"="+h.capacityProbeOwner,
		"--network", "none", "--read-only", "--cap-drop", "ALL", "--pids-limit", "16",
		"--security-opt", "no-new-privileges", "--env", "LC_ALL=C",
		"--mount", "type=bind,src=/var/lib/docker,dst=/polymetrics-image-store,readonly",
		"--entrypoint", "/bin/df", scriptedSourceImageID,
		"-P", "-B1", "/polymetrics-image-store",
	) {
		t.Fatalf("commands = %v, want the locked-down daemon capacity probe", runner.commands)
	}
	if err := h.Close(context.Background()); err != nil {
		t.Fatalf("Close(): %v", err)
	}
	if !commandsContain(runner.commands, "container", "inspect", "--format", dockerCapacityProbeOwnerFormat, h.capacityProbeName) {
		t.Fatalf("commands = %v, want capacity probe ownership cleanup check", runner.commands)
	}
	if commandsContain(runner.commands, "container", "rm", "--force") {
		t.Fatalf("commands = %v, want no removal after the probe self-cleans", runner.commands)
	}
}

func TestDockerVMCapacityUsesProbeImageIDAfterTagChanges(t *testing.T) {
	runner := &scriptedRunner{
		infoOutput:     "DAEMON:12345\t/var/lib/docker\n",
		capacityOutput: "Filesystem 1-blocks Used Available Capacity Mounted on\n/dev/root 1000 800 123 80% /polymetrics-image-store\n",
	}
	config := testConfig(runner)
	config.ContainerRuntime = RuntimeDocker
	config.DockerCapacityProbeImage = "example.invalid/capacity-probe:1.0"
	runner.setImage(config.DockerCapacityProbeImage, true)
	runner.retagAfterInspect = config.DockerCapacityProbeImage
	config.DiskFreeAt = func(string) (uint64, error) {
		return 0, errors.New("Docker store belongs to a VM")
	}
	h, err := New(config)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}

	free, err := h.targetImageStoreFree(context.Background())
	if err != nil {
		t.Fatalf("targetImageStoreFree(): %v", err)
	}
	if free != 123 {
		t.Fatalf("targetImageStoreFree() = %d, want daemon-side 123", free)
	}
	if imageID := runner.imageID(config.DockerCapacityProbeImage); imageID != scriptedForeignImageID {
		t.Fatalf("capacity probe tag image ID = %q, want a raced foreign tag", imageID)
	}
	if !commandsContain(runner.commands,
		"run", "--pull=never", "--rm", "--name", h.capacityProbeName,
		"--label", dockerCapacityProbeOwnerLabel+"="+h.capacityProbeOwner,
		"--network", "none", "--read-only", "--cap-drop", "ALL", "--pids-limit", "16",
		"--security-opt", "no-new-privileges", "--env", "LC_ALL=C",
		"--mount", "type=bind,src=/var/lib/docker,dst=/polymetrics-image-store,readonly",
		"--entrypoint", "/bin/df", scriptedSourceImageID,
		"-P", "-B1", "/polymetrics-image-store",
	) {
		t.Fatalf("commands = %v, want the cached immutable probe image ID", runner.commands)
	}
}

func TestDockerVMCapacityRefusesAPreexistingProbeWithoutClaimingItsCleanup(t *testing.T) {
	runner := &scriptedRunner{
		infoOutput: "DAEMON:12345\t/var/lib/docker\n",
	}
	config := testConfig(runner)
	config.ContainerRuntime = RuntimeDocker
	config.DockerCapacityProbeImage = "example.invalid/capacity-probe:1.0"
	runner.setImage(config.DockerCapacityProbeImage, true)
	config.DiskFreeAt = func(string) (uint64, error) {
		return 0, errors.New("Docker store belongs to a VM")
	}
	h, err := New(config)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	runner.setCapacityProbeOwner(h.capacityProbeName, "foreign")

	if _, err := h.targetImageStoreFree(context.Background()); err == nil || !strings.Contains(err.Error(), "capacity probe already exists") {
		t.Fatalf("targetImageStoreFree() error = %v, want an existing probe refusal", err)
	}
	if h.known(func(h *Harness) bool { return h.capacityProbeKnown }) {
		t.Fatal("targetImageStoreFree() claimed cleanup ownership of a pre-existing probe")
	}
	if err := h.Close(context.Background()); err == nil {
		t.Fatal("Close() accepted an unproven capacity probe")
	}
	if commandsContain(runner.commands, "container", "rm", "--force") {
		t.Fatalf("commands = %v, want no removal of a pre-existing capacity probe", runner.commands)
	}
}

func TestDockerVMCapacityPreservesAProbeCreatedAfterPrecheck(t *testing.T) {
	runner := &scriptedRunner{
		infoOutput:       "DAEMON:12345\t/var/lib/docker\n",
		raceForeignProbe: true,
		runErr:           errors.New("name already in use"),
	}
	config := testConfig(runner)
	config.ContainerRuntime = RuntimeDocker
	config.DockerCapacityProbeImage = "example.invalid/capacity-probe:1.0"
	runner.setImage(config.DockerCapacityProbeImage, true)
	config.DiskFreeAt = func(string) (uint64, error) {
		return 0, errors.New("Docker store belongs to a VM")
	}
	h, err := New(config)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}

	if _, err := h.targetImageStoreFree(context.Background()); err == nil || !strings.Contains(err.Error(), "capacity probe failed") {
		t.Fatalf("targetImageStoreFree() error = %v, want a raced capacity probe failure", err)
	}
	if !h.known(func(h *Harness) bool { return h.capacityProbeKnown }) {
		t.Fatal("targetImageStoreFree() did not arm cleanup for an indeterminate probe creation")
	}
	if owner := runner.capacityProbeOwners[h.capacityProbeName]; owner != "foreign" {
		t.Fatalf("capacity probe owner = %q, want foreign", owner)
	}
	if err := h.Close(context.Background()); err == nil || !strings.Contains(err.Error(), "ownership could not be proven") {
		t.Fatalf("Close() error = %v, want an unproven capacity probe ownership error", err)
	}
	if !commandsContain(runner.commands, "container", "inspect", "--format", dockerCapacityProbeOwnerFormat, h.capacityProbeName) {
		t.Fatalf("commands = %v, want capacity probe ownership inspection", runner.commands)
	}
	if commandsContain(runner.commands, "container", "rm", "--force") {
		t.Fatalf("commands = %v, must not remove a raced foreign capacity probe", runner.commands)
	}
}

func TestCloseRemovesOnlyLabelOwnedCapacityProbe(t *testing.T) {
	runner := &scriptedRunner{infoOutput: "DAEMON:12345\t/var/lib/docker\n"}
	config := testConfig(runner)
	config.ContainerRuntime = RuntimeDocker
	h, err := New(config)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	runner.setCapacityProbeOwner(h.capacityProbeName, h.capacityProbeOwner)
	h.mu.Lock()
	h.capacityProbeKnown = true
	h.mu.Unlock()

	if err := h.Close(context.Background()); err != nil {
		t.Fatalf("Close(): %v", err)
	}
	if _, present := runner.capacityProbeOwners[h.capacityProbeName]; present {
		t.Fatal("Close() did not remove the harness-owned capacity probe")
	}
	inspect := commandIndex(runner.commands, "container", "inspect", "--format", dockerCapacityProbeOwnerFormat, h.capacityProbeName)
	remove := commandIndex(runner.commands, "container", "rm", "--force", scriptedCapacityProbeID)
	if inspect < 0 || remove < 0 || inspect > remove {
		t.Fatalf("commands = %v, want ownership inspection before capacity probe removal", runner.commands)
	}
}

func TestCloseRemovesOnlyLabelOwnedDatabaseContainer(t *testing.T) {
	runner := &scriptedRunner{}
	h, err := New(testConfig(runner))
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	runner.setContainerOwner(h.containerName, h.containerOwner)
	runner.containerLive = true
	h.mu.Lock()
	h.containerKnown = true
	h.mu.Unlock()

	if err := h.Close(context.Background()); err != nil {
		t.Fatalf("Close(): %v", err)
	}
	if _, present := runner.containerOwners[h.containerName]; present {
		t.Fatal("Close() did not remove the harness-owned database container")
	}
	inspect := commandIndex(runner.commands, "container", "inspect", "--format", databaseContainerOwnerFormat, h.containerName)
	remove := commandIndex(runner.commands, "container", "rm", "--force", "--volumes", scriptedDatabaseContainerID)
	if inspect < 0 || remove < 0 || inspect > remove {
		t.Fatalf("commands = %v, want ownership inspection before database-container removal", runner.commands)
	}
}

func TestDockerVMCapacityRefusesAnUncachedOrMalformedProbe(t *testing.T) {
	for _, tc := range []struct {
		name           string
		probeCached    bool
		capacityOutput string
		wantError      string
	}{
		{name: "uncached", wantError: "pre-cached Docker capacity probe image"},
		{name: "malformed output", probeCached: true, capacityOutput: "Filesystem 1-blocks Used Available Capacity Mounted on\n/dev/root 1000 800 not-a-number 80% /polymetrics-image-store\n", wantError: "invalid Docker image-store capacity"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runner := &scriptedRunner{
				infoOutput:     "DAEMON:12345\t/var/lib/docker\n",
				capacityOutput: tc.capacityOutput,
			}
			config := testConfig(runner)
			config.ContainerRuntime = RuntimeDocker
			config.DockerCapacityProbeImage = "example.invalid/capacity-probe:1.0"
			runner.setImage(config.DockerCapacityProbeImage, tc.probeCached)
			config.DiskFreeAt = func(string) (uint64, error) {
				return 0, errors.New("Docker store belongs to a VM")
			}
			h, err := New(config)
			if err != nil {
				t.Fatalf("New(): %v", err)
			}

			if _, err := h.targetImageStoreFree(context.Background()); err == nil || !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("targetImageStoreFree() error = %v, want %q", err, tc.wantError)
			}
			if commandsContain(runner.commands, "pull", config.DockerCapacityProbeImage) {
				t.Fatalf("commands = %v, must never pull the capacity probe", runner.commands)
			}
		})
	}
}

func TestStartRefusesDockerVMStoreWithoutAPreCachedProbeBeforeMutating(t *testing.T) {
	runner := &scriptedRunner{infoOutput: "DAEMON:12345\t/var/lib/docker\n"}
	config := testConfig(runner)
	config.ContainerRuntime = RuntimeDocker
	runner.setImage(config.Image, true)
	config.DiskFreeAt = func(string) (uint64, error) {
		return 0, errors.New("Docker store belongs to a VM")
	}
	h, err := New(config)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if _, err := h.Start(context.Background()); err == nil || !strings.Contains(err.Error(), "capacity cannot be proven") {
		t.Fatalf("Start() error = %v, want an unproven Docker VM capacity refusal", err)
	}
	for _, prefix := range [][]string{{"pull"}, {"image", "tag"}, {"run"}} {
		if commandsContain(runner.commands, prefix...) {
			t.Fatalf("commands = %v, want no mutation before Docker VM capacity is proven", runner.commands)
		}
	}
}

func TestStartRefusesMalformedDockerVMProbeHeaderBeforeMutating(t *testing.T) {
	runner := &scriptedRunner{
		infoOutput:     "DAEMON:12345\t/var/lib/docker\n",
		capacityOutput: "garbage Available\n/dev/root 1000 800 123 80% /polymetrics-image-store\n",
	}
	config := testConfig(runner)
	config.ContainerRuntime = RuntimeDocker
	config.DockerCapacityProbeImage = "example.invalid/capacity-probe:1.0"
	runner.setImage(config.DockerCapacityProbeImage, true)
	config.DiskFreeAt = func(string) (uint64, error) {
		return 0, errors.New("Docker store belongs to a VM")
	}
	h, err := New(config)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}

	if _, err := h.Start(context.Background()); err == nil || !strings.Contains(err.Error(), "capacity cannot be proven") {
		t.Fatalf("Start() error = %v, want a malformed Docker VM capacity refusal", err)
	}
	for _, prefix := range [][]string{{"pull", config.Image}, {"image", "tag"}, {"run", "--detach"}} {
		if commandsContain(runner.commands, prefix...) {
			t.Fatalf("commands = %v, want no database-image mutation after malformed capacity evidence", runner.commands)
		}
	}
}

func TestParseDockerStoreFree(t *testing.T) {
	free, err := parseDockerStoreFree("Filesystem 1-blocks Used Available Capacity Mounted on\n/dev/root 102888095744 1486704640 101384613888 2% /polymetrics-image-store\n")
	if err != nil {
		t.Fatalf("parseDockerStoreFree(): %v", err)
	}
	if free != 101384613888 {
		t.Fatalf("parseDockerStoreFree() = %d, want 101384613888", free)
	}
	for _, raw := range []string{
		"",
		"Filesystem 1-blocks Used Available Capacity Mounted on\n",
		"Filesystem 1-blocks Used Available Capacity Mounted on\n/dev/root 1000 800 -1 80% /polymetrics-image-store\n",
		"Filesystem 1-blocks Used Available Capacity Mounted on\n/dev/root 1000 800 100 80% /other\n",
		"Other 1-blocks Used Free Capacity Mounted on\n/dev/root 1000 800 100 80% /polymetrics-image-store\n",
		"garbage Available\n/dev/root 1000 800 100 80% /polymetrics-image-store\n",
		"Filesystem 1-blocks Used Capacity Available Mounted on\n/dev/root 1000 800 100 80% /polymetrics-image-store\n",
		"Filesystem 1024-blocks Used Available Capacity Mounted on\n/dev/root 1000 800 100 80% /polymetrics-image-store\n",
	} {
		if _, err := parseDockerStoreFree(raw); err == nil {
			t.Fatalf("parseDockerStoreFree(%q) accepted invalid capacity", raw)
		}
	}
}

func TestParseImageID(t *testing.T) {
	for _, raw := range []string{
		scriptedSourceImageID + "\n",
		"sha256:" + scriptedSourceImageID + "\n",
	} {
		if imageID, valid := parseImageID(raw); !valid || imageID != strings.TrimSuffix(raw, "\n") {
			t.Fatalf("parseImageID(%q) = %q, %t, want a valid immutable image ID", raw, imageID, valid)
		}
	}
	for _, raw := range []string{"", "sha256:short\n", scriptedSourceImageID + "\nsecond", "sha512:" + scriptedSourceImageID + "\n"} {
		if _, valid := parseImageID(raw); valid {
			t.Fatalf("parseImageID(%q) accepted an invalid image ID", raw)
		}
	}
}

func TestStartUsesPodmanMachineDaemonCapacity(t *testing.T) {
	runner := &scriptedRunner{infoOutput: "unix:///run/user/1000/podman/podman.sock\t/var/lib/containers/storage\t300\t100\n"}
	config := testConfig(runner)
	var hostCapacityCalls int
	config.DiskFreeAt = func(string) (uint64, error) {
		hostCapacityCalls++
		return 0, errors.New("host store is not mounted")
	}
	h, err := New(config)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	defer func() {
		if err := h.Close(context.Background()); err != nil {
			t.Fatalf("Close(): %v", err)
		}
	}()
	if _, err := h.Start(context.Background()); err != nil {
		t.Fatalf("Start(): %v", err)
	}
	if hostCapacityCalls != 0 {
		t.Fatalf("host capacity checks = %d, want daemon capacity for a forwarded Podman machine", hostCapacityCalls)
	}
	if report := h.Report(); report.DiskFreeBefore != 200 {
		t.Fatalf("disk report before = %d, want daemon-reported free capacity", report.DiskFreeBefore)
	}
}

func TestStartRefusesForwardedPodmanMachineWithoutDaemonCapacity(t *testing.T) {
	runner := &scriptedRunner{infoOutput: "unix:///run/user/1000/podman/podman.sock\t/var/lib/containers/storage\tunknown\t100\n"}
	config := testConfig(runner)
	config.DiskFreeAt = func(string) (uint64, error) {
		return 1 << 30, nil
	}
	h, err := New(config)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if _, err := h.Start(context.Background()); err == nil || !strings.Contains(err.Error(), "invalid image-store capacity") {
		t.Fatalf("Start() error = %v, want a forwarded-target capacity refusal", err)
	}
	for _, prefix := range [][]string{{"pull"}, {"image", "tag"}, {"run"}} {
		if commandsContain(runner.commands, prefix...) {
			t.Fatalf("commands = %v, want no mutation before forwarded capacity is proven", runner.commands)
		}
	}
}

func TestStartRefusesToCreateAfterClose(t *testing.T) {
	runner := &scriptedRunner{}
	h, err := New(testConfig(runner))
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if err := h.Close(context.Background()); err != nil {
		t.Fatalf("Close(): %v", err)
	}
	if _, err := h.Start(context.Background()); err == nil {
		t.Fatal("Start() created resources after the harness was closed")
	}
	if commandsContain(runner.commands, "run") {
		t.Fatalf("commands = %v, want no creates after Close", runner.commands)
	}
	interruptCleanup.mu.Lock()
	_, registered := interruptCleanup.live[h]
	interruptCleanup.mu.Unlock()
	if registered {
		t.Fatal("Start() registered an already closed harness for interrupt cleanup")
	}
	select {
	case engineSlots <- struct{}{}:
		<-engineSlots
	default:
		t.Fatal("a Start refused after Close leaked the engine slot")
	}
}

func TestParseTargetIdentity(t *testing.T) {
	identity, err := parseTargetIdentity("/tmp/dbtest.sock\t/var/lib/containers/storage\t0\t0\n", "/tmp/dbtest.sock")
	if err != nil {
		t.Fatalf("parseTargetIdentity(): %v", err)
	}
	if identity.targetID != "/tmp/dbtest.sock" || identity.graphRoot != "/var/lib/containers/storage" {
		t.Fatalf("identity = %+v, want socket and image-store paths", identity)
	}
	for _, raw := range []string{"", "/tmp/socket", "/tmp/socket\t../store\t0\t0", "/tmp/socket\t/var/lib/containers/storage \t0\t0\n", "/tmp/socket\t/store\t0\t0\nsecond"} {
		if _, err := parseTargetIdentity(raw, "/tmp/socket"); err == nil {
			t.Fatalf("parseTargetIdentity(%q) accepted invalid target evidence", raw)
		}
	}
}

func TestParseDockerTargetIdentity(t *testing.T) {
	identity, err := parseDockerTargetIdentity("ABCDE:12345\t/var/lib/docker\n")
	if err != nil {
		t.Fatalf("parseDockerTargetIdentity(): %v", err)
	}
	if identity.targetID != "ABCDE:12345" || identity.graphRoot != "/var/lib/docker" {
		t.Fatalf("identity = %+v, want Docker daemon ID and image-store path", identity)
	}
	for _, raw := range []string{"", "daemon-only", "daemon\t../docker", "daemon\t/var/lib/docker \n", "daemon\t/var/lib/docker\nsecond", "bad id!\t/var/lib/docker"} {
		if _, err := parseDockerTargetIdentity(raw); err == nil {
			t.Fatalf("parseDockerTargetIdentity(%q) accepted invalid target evidence", raw)
		}
	}
}
