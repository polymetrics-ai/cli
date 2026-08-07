package dbtest

import (
	"context"
	"errors"
	"os/exec"
	"reflect"
	"testing"
	"time"
)

type scriptedRunner struct {
	volumePresent       bool
	containerLive       bool
	pullErr             error
	tagErr              error
	volumeInspectErr    error
	volumeCreateErr     error
	containerInspectErr error
	runErr              error
	images              map[string]bool
	commands            [][]string
	connections         []string
}

func (r *scriptedRunner) Run(_ context.Context, connection string, args ...string) (string, error) {
	r.connections = append(r.connections, connection)
	r.commands = append(r.commands, append([]string(nil), args...))
	switch {
	case len(args) > 0 && args[0] == "pull":
		if len(args) > 1 {
			r.setImage(args[1], true)
		}
		return "", r.pullErr
	case commandHasPrefix(args, "image", "tag"):
		if len(args) > 3 {
			r.setImage(args[3], true)
		}
		return "", r.tagErr
	case commandHasPrefix(args, "image", "rm"):
		if len(args) > 2 {
			r.setImage(args[2], false)
		}
	case commandHasPrefix(args, "volume", "inspect"):
		if r.volumeInspectErr != nil {
			return "", r.volumeInspectErr
		}
		if r.volumePresent {
			return "", nil
		}
		return "", errPodmanResourceNotFound
	case commandHasPrefix(args, "volume", "create"):
		r.volumePresent = true
		return "", r.volumeCreateErr
	case commandHasPrefix(args, "volume", "rm"):
		r.volumePresent = false
	case commandHasPrefix(args, "container", "inspect"):
		if r.containerInspectErr != nil {
			return "", r.containerInspectErr
		}
		if r.containerLive {
			return "", nil
		}
		return "", errPodmanResourceNotFound
	case len(args) > 0 && args[0] == "run":
		r.containerLive = true
		return "", r.runErr
	case commandHasPrefix(args, "container", "rm"):
		r.containerLive = false
	case len(args) > 0 && args[0] == "port":
		return "127.0.0.1:43123\n", nil
	}
	return "", nil
}

func (r *scriptedRunner) imagePresent(image string) bool {
	return r.images != nil && r.images[image]
}

func (r *scriptedRunner) setImage(image string, present bool) {
	if r.images == nil {
		r.images = make(map[string]bool)
	}
	if present {
		r.images[image] = true
		return
	}
	delete(r.images, image)
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

type scriptedMachineRunner struct{ commands [][]string }

func (r *scriptedMachineRunner) Run(_ context.Context, args ...string) error {
	r.commands = append(r.commands, append([]string(nil), args...))
	return nil
}

func testConfig(runner CommandRunner) Config {
	return Config{
		Engine:         "mysql",
		Image:          "example.invalid/mysql:8.4.11",
		ContainerPort:  3306,
		DataVolumePath: "/var/lib/mysql",
		Connection:     "lane-machine",
		Run:            runner,
		DiskFree: func() (uint64, error) {
			return 123, nil
		},
	}
}

func TestNewRejectsUnscopedConnection(t *testing.T) {
	for _, tc := range []struct {
		name       string
		connection string
	}{
		{name: "empty falls back to the default connection", connection: ""},
		{name: "whitespace only", connection: "  "},
		{name: "argument injection", connection: "lane --remote"},
		{name: "newline injection", connection: "lane\nmachine"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config := testConfig(&scriptedRunner{})
			config.Connection = tc.connection
			if _, err := New(config); err == nil {
				t.Fatal("New() accepted an unsafe or absent Podman connection")
			}
		})
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
	if err := h.Close(context.Background()); err != nil {
		t.Fatalf("Close(): %v", err)
	}
	if runner.imagePresent(config.Image) || runner.imagePresent(h.runImage) {
		t.Fatalf("images after cleanup = %#v, want every run-owned reference removed", runner.images)
	}

	var gotCleanup [][]string
	for _, command := range runner.commands {
		if commandHasPrefix(command, "container", "rm") || commandHasPrefix(command, "volume", "rm") || commandHasPrefix(command, "image", "rm") {
			gotCleanup = append(gotCleanup, command[:2])
		}
	}
	want := [][]string{{"container", "rm"}, {"volume", "rm"}, {"image", "rm"}, {"image", "rm"}}
	if !reflect.DeepEqual(gotCleanup, want) {
		t.Fatalf("cleanup order = %v, want %v", gotCleanup, want)
	}
	for _, connection := range runner.connections {
		if connection != "lane-machine" {
			t.Fatalf("podman command used connection %q, want the explicit connection", connection)
		}
	}
	report := h.Report()
	if report.DiskFreeBefore != 123 || report.DiskFreeAfter != 123 || report.HostDiskReclaimed {
		t.Fatalf("disk report = %+v, want before/after values without reclaim", report)
	}
}

func TestCleanupKeepsSourceImageOnlyWhenOptedIn(t *testing.T) {
	runner := &scriptedRunner{}
	config := testConfig(runner)
	config.KeepImage = true
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
	if !runner.imagePresent(config.Image) {
		t.Fatalf("images after opted-in cleanup = %#v, want retained source image", runner.images)
	}
	if runner.imagePresent(h.runImage) {
		t.Fatalf("images after opted-in cleanup = %#v, want generated reference removed", runner.images)
	}
	if commandsContain(runner.commands, "image", "rm", config.Image) {
		t.Fatalf("commands = %v, want no source image removal when KeepImage is set", runner.commands)
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
		if command[0] != "run" {
			continue
		}
		imageIndex, engineArgumentIndex := -1, -1
		for index, arg := range command {
			if arg == h.runImage {
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
	t.Fatal("Start did not issue a podman run command")
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
		if command[0] != "run" {
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
		name             string
		runner           *scriptedRunner
		wantClean        [][]string
		wantNoImageClean bool
	}{
		{
			name:             "pull returns an error before any reference is owned",
			runner:           &scriptedRunner{pullErr: context.Canceled},
			wantNoImageClean: true,
		},
		{
			name:   "tag returns an error after the pull succeeded",
			runner: &scriptedRunner{tagErr: context.Canceled},
			wantClean: [][]string{
				{"image", "rm"},
			},
		},
		{
			name:   "volume create returns an error after creating the volume",
			runner: &scriptedRunner{volumeCreateErr: context.Canceled},
			wantClean: [][]string{
				{"volume", "rm"},
				{"image", "rm"},
			},
		},
		{
			name:   "container run returns an error after creating the container",
			runner: &scriptedRunner{runErr: context.Canceled},
			wantClean: [][]string{
				{"container", "rm"},
				{"volume", "rm"},
				{"image", "rm"},
			},
		},
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
			for _, prefix := range tc.wantClean {
				if !commandsContain(tc.runner.commands, prefix...) {
					t.Fatalf("cleanup commands = %v, missing %v", tc.runner.commands, prefix)
				}
			}
			if tc.wantNoImageClean && commandsContain(tc.runner.commands, "image", "rm") {
				t.Fatalf("cleanup commands = %v, want no image removal when the pull failed", tc.runner.commands)
			}
			if !tc.wantNoImageClean && tc.runner.imagePresent(h.runImage) {
				t.Fatalf("images after cleanup = %#v, want generated reference removed", tc.runner.images)
			}
		})
	}
}

func TestStartReleasesTheEngineSlotOnEveryExitPath(t *testing.T) {
	// A leaked slot would deadlock the next sequential engine rather than
	// fail it, so prove the slot returns after both outcomes.
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

func TestStartUsesGeneratedImageReferenceWithoutInspectingSource(t *testing.T) {
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
	if commandsContain(runner.commands, "image", "inspect") {
		t.Fatalf("commands = %v, want no shared image inspection", runner.commands)
	}
	if !commandsContain(runner.commands, "pull", config.Image) || !commandsContain(runner.commands, "image", "tag", config.Image, h.runImage) {
		t.Fatalf("commands = %v, want generated image reference", runner.commands)
	}
}

func TestPodmanResourceNotFoundClassifiesOnlyKnownAbsence(t *testing.T) {
	for _, stderr := range []string{
		`Error: no such container "x"`,
		`Error: no such volume "x"`,
		// Podman's wording where Docker says "no such image". Missing this
		// phrasing turns an ordinary absent image into a cleanup failure.
		"Error: x: image not known",
		"Error response from daemon: No such image",
	} {
		if !podmanResourceNotFound(&exec.ExitError{Stderr: []byte(stderr)}) {
			t.Fatalf("podmanResourceNotFound(%q) = false, want recognized absence", stderr)
		}
	}
	for _, stderr := range []string{
		"Error response from daemon: access denied",
		"Error: unable to connect to Podman socket",
	} {
		if podmanResourceNotFound(&exec.ExitError{Stderr: []byte(stderr)}) {
			t.Fatalf("podmanResourceNotFound(%q) = true, want indeterminate error", stderr)
		}
	}
}

func TestStartPreservesPreexistingGeneratedVolume(t *testing.T) {
	runner := &scriptedRunner{volumePresent: true}
	h, err := New(testConfig(runner))
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if _, err := h.Start(context.Background()); err == nil {
		t.Fatal("Start() accepted an existing generated volume")
	}
	if commandsContain(runner.commands, "volume", "rm") {
		t.Fatalf("commands = %v, want no removal of a pre-existing volume", runner.commands)
	}
}

func TestCleanupReclaimsHostDiskOnlyAfterContainerCleanup(t *testing.T) {
	runner := &scriptedRunner{}
	machine := &scriptedMachineRunner{}
	config := testConfig(runner)
	config.ReclaimHostDisk = true
	config.Machine = "lane-machine"
	config.RunMachine = machine
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
	// Two passes are required: one discard immediately after an image removal
	// returns only part of the space on Podman 5.3 with applehv.
	trim := []string{"machine", "ssh", "lane-machine", "sudo fstrim -av"}
	if want := [][]string{trim, trim}; !reflect.DeepEqual(machine.commands, want) {
		t.Fatalf("machine lifecycle commands = %v, want %v", machine.commands, want)
	}
	if !h.Report().HostDiskReclaimed {
		t.Fatal("report did not record the completed host-disk reclaim")
	}
	lastContainerCleanup := -1
	for index, command := range runner.commands {
		if len(command) > 1 && ((command[0] == "container" && command[1] == "rm") || (command[0] == "volume" && command[1] == "rm") || (command[0] == "image" && command[1] == "rm")) {
			lastContainerCleanup = index
		}
	}
	if lastContainerCleanup < 0 {
		t.Fatal("container cleanup was not attempted before the host-disk reclaim")
	}
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

// interruptingRunner fires Close from inside a create command, which is where
// a real signal is most damaging: cleanup that runs between an ownership
// claim and the command acting on it would leave the resource behind.
type interruptingRunner struct {
	scriptedRunner
	harness  *Harness
	on       string
	fired    bool
	done     chan struct{}
	closeErr error
}

func (r *interruptingRunner) Run(ctx context.Context, connection string, args ...string) (string, error) {
	if !r.fired && len(args) > 0 && args[0] == r.on {
		r.fired = true
		r.done = make(chan struct{})
		go func() {
			defer close(r.done)
			r.closeErr = r.harness.Close(context.Background())
		}()
		// Cleanup must not complete while this create is still in flight.
		// The caller waits for r.done only after Start has returned.
		select {
		case <-r.done:
			r.closeErr = errors.New("Close finished while a create command was still running")
		case <-time.After(25 * time.Millisecond):
		}
	}
	return r.scriptedRunner.Run(ctx, connection, args...)
}

func TestCloseDuringCreateStillRemovesTheCreatedResource(t *testing.T) {
	for _, step := range []string{"run", "volume"} {
		t.Run("interrupted during "+step, func(t *testing.T) {
			runner := &interruptingRunner{on: step}
			config := testConfig(runner)
			h, err := New(config)
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
			// Whatever the interrupted step created must have been removed.
			if runner.containerLive {
				t.Fatal("interrupt left the container behind")
			}
			if runner.volumePresent {
				t.Fatal("interrupt left the volume behind")
			}
			if runner.imagePresent(h.runImage) || runner.imagePresent(config.Image) {
				t.Fatalf("interrupt left images behind: %#v", runner.images)
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
	if commandsContain(runner.commands, "run") || commandsContain(runner.commands, "volume", "create") {
		t.Fatalf("commands = %v, want no creates after Close", runner.commands)
	}
}
