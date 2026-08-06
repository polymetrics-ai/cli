package dbtest

import (
	"context"
	"os/exec"
	"reflect"
	"testing"
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
	contexts            []string
}

func (r *scriptedRunner) Run(_ context.Context, dockerContext string, args ...string) (string, error) {
	r.contexts = append(r.contexts, dockerContext)
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
		return "", errDockerResourceNotFound
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
		return "", errDockerResourceNotFound
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

type scriptedColimaRunner struct{ commands [][]string }

func (r *scriptedColimaRunner) Run(_ context.Context, args ...string) error {
	r.commands = append(r.commands, append([]string(nil), args...))
	return nil
}

func testConfig(runner CommandRunner) Config {
	return Config{
		Engine:         "mysql",
		Image:          "example.invalid/mysql:8.4.11",
		ContainerPort:  3306,
		DataVolumePath: "/var/lib/mysql",
		DockerContext:  "colima",
		Run:            runner,
		DiskFree: func() (uint64, error) {
			return 123, nil
		},
	}
}

func TestNewRejectsUnscopedDockerContext(t *testing.T) {
	config := testConfig(&scriptedRunner{})
	config.DockerContext = ""
	_, err := New(config)
	if err == nil {
		t.Fatal("New() succeeded without an explicit Docker context")
	}
}

func TestNewRejectsUnpinnedImages(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*Config)
	}{
		{name: "latest image", mutate: func(c *Config) { c.Image = "example.invalid/mysql:latest" }},
		{name: "unversioned image", mutate: func(c *Config) { c.Image = "example.invalid/mysql" }},
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
	if !runner.imagePresent(config.Image) || runner.imagePresent(h.runImage) {
		t.Fatalf("images after cleanup = %#v, want shared source only", runner.images)
	}

	var gotCleanup []string
	for _, command := range runner.commands {
		if commandHasPrefix(command, "container", "rm") || commandHasPrefix(command, "volume", "rm") || commandHasPrefix(command, "image", "rm") {
			gotCleanup = append(gotCleanup, command[0])
		}
	}
	if want := []string{"container", "volume", "image"}; !reflect.DeepEqual(gotCleanup, want) {
		t.Fatalf("cleanup order = %v, want %v", gotCleanup, want)
	}
	if commandsContain(runner.commands, "image", "rm", config.Image) || !commandsContain(runner.commands, "image", "rm", h.runImage) {
		t.Fatalf("image cleanup commands = %v, want generated reference only", runner.commands)
	}
	for _, dockerContext := range runner.contexts {
		if dockerContext != "colima" {
			t.Fatalf("Docker command used context %q, want explicit context", dockerContext)
		}
	}
	report := h.Report()
	if report.DiskFreeBefore != 123 || report.DiskFreeAfter != 123 || report.ColimaReset {
		t.Fatalf("disk report = %+v, want before/after values without reset", report)
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
	t.Fatal("Start did not issue a Docker run command")
}

func TestStartCleansResourcesAfterIndeterminateDaemonOutcomes(t *testing.T) {
	for _, tc := range []struct {
		name             string
		runner           *scriptedRunner
		wantClean        [][]string
		wantNoImageClean bool
	}{
		{
			name:             "pull returns an error before a generated reference exists",
			runner:           &scriptedRunner{pullErr: context.Canceled},
			wantNoImageClean: true,
		},
		{
			name:   "tag returns an error after creating the generated reference",
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
				t.Fatal("Start() succeeded after an indeterminate daemon outcome")
			}
			for _, prefix := range tc.wantClean {
				if !commandsContain(tc.runner.commands, prefix...) {
					t.Fatalf("cleanup commands = %v, missing %v", tc.runner.commands, prefix)
				}
			}
			if tc.wantNoImageClean && commandsContain(tc.runner.commands, "image", "rm") {
				t.Fatalf("cleanup commands = %v, want no generated image removal", tc.runner.commands)
			}
			if tc.wantNoImageClean && !tc.runner.imagePresent(config.Image) {
				t.Fatalf("images after pull error = %#v, want retained shared source cache", tc.runner.images)
			}
			if !tc.wantNoImageClean && tc.runner.imagePresent(h.runImage) {
				t.Fatalf("images after cleanup = %#v, want generated reference removed", tc.runner.images)
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

func TestDockerResourceNotFoundClassifiesOnlyKnownAbsence(t *testing.T) {
	if !dockerResourceNotFound(&exec.ExitError{Stderr: []byte("Error response from daemon: No such image")}) {
		t.Fatal("dockerResourceNotFound() did not recognize a missing image")
	}
	if dockerResourceNotFound(&exec.ExitError{Stderr: []byte("Error response from daemon: access denied")}) {
		t.Fatal("dockerResourceNotFound() classified an indeterminate inspect error as absence")
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

func TestCleanupRemovesOnlyGeneratedImageReference(t *testing.T) {
	for _, tc := range []struct {
		name          string
		sourcePresent bool
	}{
		{name: "preexisting source image", sourcePresent: true},
		{name: "source image pulled for the run"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runner := &scriptedRunner{}
			config := testConfig(runner)
			if tc.sourcePresent {
				runner.setImage(config.Image, true)
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
			if !runner.imagePresent(config.Image) || runner.imagePresent(h.runImage) {
				t.Fatalf("images after cleanup = %#v, want shared source only", runner.images)
			}
			if commandsContain(runner.commands, "image", "rm", config.Image) || !commandsContain(runner.commands, "image", "rm", h.runImage) {
				t.Fatalf("image cleanup commands = %v, want generated reference only", runner.commands)
			}
		})
	}
}

func TestCleanupResetsColimaOnlyAfterDockerCleanup(t *testing.T) {
	runner := &scriptedRunner{}
	colima := &scriptedColimaRunner{}
	config := testConfig(runner)
	config.ResetColima = true
	config.ColimaProfile = "default"
	config.RunColima = colima
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
	if want := [][]string{{"delete", "--force", "--profile", "default"}, {"start", "--profile", "default", "--runtime", "docker"}}; !reflect.DeepEqual(colima.commands, want) {
		t.Fatalf("Colima lifecycle commands = %v, want %v", colima.commands, want)
	}
	if !h.Report().ColimaReset {
		t.Fatal("report did not record the completed Colima reset")
	}
	lastDockerCleanup := -1
	for index, command := range runner.commands {
		if len(command) > 1 && ((command[0] == "container" && command[1] == "rm") || (command[0] == "volume" && command[1] == "rm") || (command[0] == "image" && command[1] == "rm")) {
			lastDockerCleanup = index
		}
	}
	if lastDockerCleanup < 0 {
		t.Fatal("Docker cleanup was not attempted before Colima reset")
	}
}
