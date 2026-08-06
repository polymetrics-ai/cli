package dbtest

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type scriptedRunner struct {
	imagePresent  bool
	containerLive bool
	commands      [][]string
	contexts      []string
}

func (r *scriptedRunner) Run(_ context.Context, dockerContext string, args ...string) (string, error) {
	r.contexts = append(r.contexts, dockerContext)
	r.commands = append(r.commands, append([]string(nil), args...))
	if reflect.DeepEqual(args[:2], []string{"image", "inspect"}) {
		if r.imagePresent {
			return "", nil
		}
		return "", errors.New("not present")
	}
	if reflect.DeepEqual(args[:2], []string{"container", "inspect"}) && !r.containerLive {
		return "", errors.New("not present")
	}
	if args[0] == "run" {
		r.containerLive = true
	}
	if reflect.DeepEqual(args[:2], []string{"container", "rm"}) {
		r.containerLive = false
	}
	if args[0] == "port" {
		return "127.0.0.1:43123\n", nil
	}
	return "", nil
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

func TestNewRejectsLatestImageAndConflictingColimaReset(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*Config)
	}{
		{name: "latest image", mutate: func(c *Config) { c.Image = "example.invalid/mysql:latest" }},
		{name: "unversioned image", mutate: func(c *Config) { c.Image = "example.invalid/mysql" }},
		{name: "keep image while reset", mutate: func(c *Config) { c.KeepImage, c.ResetColima = true, true }},
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
	h, err := New(testConfig(runner))
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
	if err := h.Close(context.Background()); err != nil {
		t.Fatalf("Close(): %v", err)
	}

	var gotCleanup []string
	for _, command := range runner.commands {
		if reflect.DeepEqual(command[:2], []string{"container", "rm"}) || (len(command) > 1 && command[0] == "volume" && command[1] == "rm") || (len(command) > 1 && command[0] == "image" && command[1] == "rm") {
			gotCleanup = append(gotCleanup, command[0])
		}
	}
	if want := []string{"container", "volume", "image"}; !reflect.DeepEqual(gotCleanup, want) {
		t.Fatalf("cleanup order = %v, want %v", gotCleanup, want)
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
	for _, command := range runner.commands {
		if command[0] != "run" {
			continue
		}
		imageIndex, engineArgumentIndex := -1, -1
		for index, arg := range command {
			if arg == config.Image {
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

func TestCleanupPreservesPreexistingImageAndKeepImageOptIn(t *testing.T) {
	for _, tc := range []struct {
		name         string
		imagePresent bool
		keepImage    bool
		wantImageRM  bool
	}{
		{name: "preexisting image", imagePresent: true},
		{name: "keep image opt-in", keepImage: true},
		{name: "run-owned image", wantImageRM: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runner := &scriptedRunner{imagePresent: tc.imagePresent}
			config := testConfig(runner)
			config.KeepImage = tc.keepImage
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
			gotImageRM := false
			for _, command := range runner.commands {
				gotImageRM = gotImageRM || (len(command) > 1 && command[0] == "image" && command[1] == "rm")
			}
			if gotImageRM != tc.wantImageRM {
				t.Fatalf("image removal = %t, want %t", gotImageRM, tc.wantImageRM)
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
