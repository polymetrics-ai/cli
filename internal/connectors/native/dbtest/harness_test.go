package dbtest

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"reflect"
	"strings"
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

// scriptedMachineRunner tracks the machines a test created through NewMachine
// and answers "machine inspect" for exactly those, so ownership can be
// exercised without a VM. inspectName and inspectErr override that answer to
// model a host that disagrees with the record.
type scriptedMachineRunner struct {
	inspectName string
	inspectErr  error
	initErr     error
	// initWritesBeforeFailing models the real command: podman writes the VM
	// config and its disk image before init can fail, and a cancelled context
	// kills it mid-write, so the machine exists even though init reported none.
	initWritesBeforeFailing bool
	created                 map[string]bool
	commands                [][]string
}

func (r *scriptedMachineRunner) Run(ctx context.Context, args ...string) (string, error) {
	r.commands = append(r.commands, append([]string(nil), args...))
	name := args[len(args)-1]
	switch {
	case commandHasPrefix(args, "machine", "init"):
		if r.initWritesBeforeFailing {
			r.setCreated(name, true)
		}
		if err := ctx.Err(); err != nil {
			return "", err
		}
		if r.initErr != nil {
			return "", r.initErr
		}
		r.setCreated(name, true)
	case commandHasPrefix(args, "machine", "stop"), commandHasPrefix(args, "machine", "rm"):
		if !r.created[name] {
			return "", fmt.Errorf("%w: no such machine", errPodmanResourceNotFound)
		}
		if commandHasPrefix(args, "machine", "rm") {
			r.setCreated(name, false)
		}
	case commandHasPrefix(args, "machine", "inspect"):
		if r.inspectErr != nil {
			return "", r.inspectErr
		}
		if r.inspectName != "" {
			return r.inspectName + "\n", nil
		}
		if !r.created[name] {
			return "", errors.New("no such machine")
		}
		return name + "\n", nil
	}
	return "", nil
}

func (r *scriptedMachineRunner) setCreated(name string, present bool) {
	if r.created == nil {
		r.created = make(map[string]bool)
	}
	if present {
		r.created[name] = true
		return
	}
	delete(r.created, name)
}

// newOwnedTestMachine creates a machine through the package's own API, which
// is the only thing that establishes an ownership record.
func newOwnedTestMachine(t *testing.T, runner *scriptedMachineRunner) *Machine {
	t.Helper()
	machine, err := NewMachine(context.Background(), MachineConfig{Engine: "mysql", RunMachine: runner})
	if err != nil {
		t.Fatalf("NewMachine(): %v", err)
	}
	t.Cleanup(func() { _ = machine.Remove(context.Background()) })
	return machine
}

func testConfig(runner CommandRunner) Config {
	return Config{
		Engine:         "mysql",
		Image:          "example.invalid/mysql:8.4.11",
		ContainerPort:  3306,
		DataVolumePath: "/var/lib/mysql",
		Connection:     "lane-machine",
		Run:            runner,
		RunMachine:     &scriptedMachineRunner{},
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
	if report.HostDiskReclaimSkipped == "" {
		t.Fatalf("disk report = %+v, want a named reason for the skipped reclaim", report)
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

func TestNewMachineCreatesAndRemovesOnlyItsOwnMachine(t *testing.T) {
	runner := &scriptedMachineRunner{}
	machine, err := NewMachine(context.Background(), MachineConfig{Engine: "mysql", RunMachine: runner})
	if err != nil {
		t.Fatalf("NewMachine(): %v", err)
	}
	if !strings.HasPrefix(machine.Name(), machineNamePrefix+"mysql-") || machine.Connection() != machine.Name() {
		t.Fatalf("machine = %q / %q, want a generated name addressed by its own connection", machine.Name(), machine.Connection())
	}
	// Podman refuses a longer name, and the runner discards its output, so an
	// over-long name would surface only as an opaque init failure.
	if len(machine.Name()) > maxPodmanMachineNameLength {
		t.Fatalf("machine name %q is %d characters, want at most %d", machine.Name(), len(machine.Name()), maxPodmanMachineNameLength)
	}
	if !machineIsOwned(machine.Name()) {
		t.Fatal("NewMachine() did not record this run as the machine's owner")
	}
	if !commandsContain(runner.commands, "machine", "init") || !commandsContain(runner.commands, "machine", "start") {
		t.Fatalf("machine commands = %v, want init and start", runner.commands)
	}

	if err := machine.Remove(context.Background()); err != nil {
		t.Fatalf("Remove(): %v", err)
	}
	if !commandsContain(runner.commands, "machine", "rm", "--force", machine.Name()) {
		t.Fatalf("machine commands = %v, want the created machine removed by name", runner.commands)
	}
	// Ownership must not outlive the machine: a later run reusing the name
	// would otherwise inherit a trim right over a VM this process no longer has.
	if machineIsOwned(machine.Name()) {
		t.Fatal("Remove() kept the ownership record for a deleted machine")
	}
	if err := machine.Remove(context.Background()); err != nil {
		t.Fatalf("second Remove(): %v", err)
	}
}

// `podman machine init` writes a multi-GiB disk image before it can fail, and
// a cancelled context kills it mid-write, so a claim taken only on success
// would leave that image on the host with nothing holding authority to delete
// it.
func TestNewMachineRemovesWhatAFailedInitLeftBehind(t *testing.T) {
	for _, tc := range []struct {
		name        string
		runner      *scriptedMachineRunner
		cancelled   bool
		wantWritten bool
	}{
		{
			name:        "init failed after writing the machine",
			runner:      &scriptedMachineRunner{initErr: errors.New("init failed"), initWritesBeforeFailing: true},
			wantWritten: true,
		},
		{
			name:        "init was cancelled mid-write",
			runner:      &scriptedMachineRunner{initWritesBeforeFailing: true},
			cancelled:   true,
			wantWritten: true,
		},
		{
			name:   "init failed before writing anything",
			runner: &scriptedMachineRunner{initErr: errors.New("init failed")},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tc.cancelled {
				cancel()
			}
			machine, err := NewMachine(ctx, MachineConfig{Engine: "mysql", RunMachine: tc.runner})
			if err == nil {
				t.Fatal("NewMachine() reported success after the init failed")
			}
			// Without a handle the caller cannot clean up at all: the
			// integration test only defers Remove on a non-nil machine.
			if machine == nil {
				t.Fatal("NewMachine() returned no handle for a machine init may have created")
			}
			if !commandsContain(tc.runner.commands, "machine", "rm", "--force", machine.Name()) {
				t.Fatalf("machine commands = %v, want the attempted machine removed by name", tc.runner.commands)
			}
			if tc.runner.created[machine.Name()] {
				t.Fatalf("machine %q survived the failed init", machine.Name())
			}
			if machineIsOwned(machine.Name()) {
				t.Fatalf("ownership record for %q outlived the failed init", machine.Name())
			}
			// A machine init never created is absent, not leaked, so cleanup
			// must not report a failure for it.
			if err := machine.Remove(context.Background()); err != nil {
				t.Fatalf("cleanup after the failed init reported: %v", err)
			}
			for _, command := range tc.runner.commands {
				if command[len(command)-1] != machine.Name() {
					t.Fatalf("machine command %v targeted something other than this run's machine", command)
				}
			}
		})
	}
}

func TestCleanupReclaimsHostDiskOnlyAfterContainerCleanup(t *testing.T) {
	runner := &scriptedRunner{}
	machineRunner := &scriptedMachineRunner{}
	machine := newOwnedTestMachine(t, machineRunner)
	config := testConfig(runner)
	config.Connection = machine.Connection()
	config.Machine = machine.Name()
	config.RunMachine = machineRunner
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
	var trims [][]string
	for _, command := range machineRunner.commands {
		if commandHasPrefix(command, "machine", "ssh") {
			trims = append(trims, command)
		}
	}
	trim := []string{"machine", "ssh", machine.Name(), "sudo fstrim -av"}
	if want := [][]string{trim, trim}; !reflect.DeepEqual(trims, want) {
		t.Fatalf("machine trim commands = %v, want %v", trims, want)
	}
	if !commandsContain(machineRunner.commands, "machine", "inspect", "--format", "{{.Name}}", machine.Name()) {
		t.Fatalf("machine commands = %v, want the machine confirmed before the trim", machineRunner.commands)
	}
	report := h.Report()
	if !report.HostDiskReclaimed || report.HostDiskReclaimSkipped != "" {
		t.Fatalf("report = %+v, want a completed host-disk reclaim", report)
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

// fstrim reaches every filesystem on a machine, so the gate has to be
// ownership rather than a name that happens to match what a caller configured.
func TestCleanupNeverTrimsAMachineThisRunDidNotCreate(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(*testing.T, *Config) *scriptedMachineRunner
	}{
		{
			// The decisive case: everything a name check could look at agrees,
			// and the machine is still not this run's to trim.
			name: "caller supplied a machine this run never created",
			setup: func(_ *testing.T, config *Config) *scriptedMachineRunner {
				runner := &scriptedMachineRunner{inspectName: "lane-machine"}
				config.Connection = "lane-machine"
				config.Machine = "lane-machine"
				return runner
			},
		},
		{
			name: "ownership was released before cleanup",
			setup: func(t *testing.T, config *Config) *scriptedMachineRunner {
				runner := &scriptedMachineRunner{}
				machine := newOwnedTestMachine(t, runner)
				config.Connection = machine.Connection()
				config.Machine = machine.Name()
				runner.inspectName = machine.Name()
				if err := machine.Remove(context.Background()); err != nil {
					t.Fatalf("Remove(): %v", err)
				}
				return runner
			},
		},
		{
			name: "connection addresses another machine",
			setup: func(t *testing.T, config *Config) *scriptedMachineRunner {
				runner := &scriptedMachineRunner{}
				machine := newOwnedTestMachine(t, runner)
				config.Connection = "lane-machine"
				config.Machine = machine.Name()
				return runner
			},
		},
		{
			name: "machine is no longer defined on this host",
			setup: func(t *testing.T, config *Config) *scriptedMachineRunner {
				runner := &scriptedMachineRunner{}
				machine := newOwnedTestMachine(t, runner)
				config.Connection = machine.Connection()
				config.Machine = machine.Name()
				runner.inspectErr = errors.New("machine inspect failed")
				return runner
			},
		},
		{
			name: "machine inspect names a different machine",
			setup: func(t *testing.T, config *Config) *scriptedMachineRunner {
				runner := &scriptedMachineRunner{}
				machine := newOwnedTestMachine(t, runner)
				config.Connection = machine.Connection()
				config.Machine = machine.Name()
				runner.inspectName = "someone-elses-machine"
				return runner
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runner := &scriptedRunner{}
			config := testConfig(runner)
			config.RunMachine = tc.setup(t, &config)
			machineRunner := config.RunMachine.(*scriptedMachineRunner)
			free := []uint64{4096, 1024}
			config.DiskFree = func() (uint64, error) {
				value := free[0]
				if len(free) > 1 {
					free = free[1:]
				}
				return value, nil
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
			if commandsContain(machineRunner.commands, "machine", "ssh") {
				t.Fatalf("machine commands = %v, want no trim of an unowned machine", machineRunner.commands)
			}
			report := h.Report()
			if report.HostDiskReclaimed || report.HostDiskReclaimSkipped == "" {
				t.Fatalf("report = %+v, want a skipped reclaim with a named reason", report)
			}
			if report.HostDiskReclaimableBytes != 3072 {
				t.Fatalf("report = %+v, want the still-held host bytes reported", report)
			}
		})
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

	// A per-harness signal handler let the first finished cleanup exit the
	// process while its siblings were still removing containers.
	closeLiveHarnesses(context.Background())
	for index, runner := range runners {
		if runner.containerLive || runner.volumePresent {
			t.Fatalf("harness %d kept container or volume through the interrupt", index)
		}
		if runner.imagePresent(harnesses[index].runImage) {
			t.Fatalf("harness %d kept its generated image through the interrupt", index)
		}
	}
	for _, h := range harnesses {
		if err := h.Close(context.Background()); err != nil {
			t.Fatalf("Close(): %v", err)
		}
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
	// Close returns the slot only once, so this refused Start has to return its
	// own token. A leak here would not fail the next engine, it would block it
	// until its context expired.
	select {
	case engineSlots <- struct{}{}:
		<-engineSlots
	default:
		t.Fatal("a Start refused after Close leaked the engine slot")
	}
}
