package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"regexp"
	"strings"
)

// This is the original terminal capture dialect, not a pluggable command or
// receipt executor. Cwd and snapshot strings are provenance and never opened.
type sourceFoundationCapture struct {
	Name                string   `json:"name"`
	Command             []string `json:"command"`
	Cwd                 string   `json:"cwd"`
	Head                string   `json:"head"`
	Tree                string   `json:"tree"`
	TrackedStatusBefore string   `json:"tracked_status_before"`
	GoEnvironment       struct {
		Version    string `json:"GOVERSION"`
		OS         string `json:"GOOS"`
		Arch       string `json:"GOARCH"`
		CGO        string `json:"CGO_ENABLED"`
		Flags      string `json:"GOFLAGS"`
		Experiment string `json:"GOEXPERIMENT"`
		Module     string `json:"GOMOD"`
		Work       string `json:"GOWORK"`
	} `json:"go_environment"`
	Inputs map[string]struct {
		SHA256   string `json:"sha256"`
		Bytes    int64  `json:"bytes"`
		Snapshot string `json:"snapshot"`
	} `json:"inputs"`
	Started        float64  `json:"started_epoch"`
	Completed      float64  `json:"completed_epoch"`
	Wall           float64  `json:"wall_seconds"`
	ExitCode       int      `json:"exit_code"`
	OutputSHA256   string   `json:"output_sha256"`
	OutputBytes    int64    `json:"output_bytes"`
	SelectedEvents int      `json:"selected_test_events"`
	PassingEvents  int      `json:"passing_test_events"`
	ChangedInputs  []string `json:"changed_inputs_after"`
}

type sourceFoundationExecutions struct {
	cache    *sourceProofFileCache
	module   string
	captures map[string]sourceFoundationCapture
}

func (s *sourceFoundationExecutions) validate(r sourceFoundationProofRecord) error {
	if s.module == "" {
		raw, file := s.cache.get("go.mod", 4<<20, false)
		if file.code != "" {
			return fmt.Errorf("foundation module input unavailable")
		}
		for _, line := range strings.Split(string(raw), "\n") {
			fields := strings.Fields(line)
			if len(fields) == 2 && fields[0] == "module" {
				s.module = fields[1]
				break
			}
		}
		if !validSourceID(s.module) {
			return fmt.Errorf("foundation module declaration invalid")
		}
	}
	capture, err := s.capture(r.Capture)
	if err != nil {
		return err
	}
	if !sourceFoundationCaptureCommand(capture.Command, r.Test) {
		return fmt.Errorf("foundation original command invalid")
	}
	if capture.OutputSHA256 != r.Output.SHA256 || capture.OutputBytes != r.Output.Bytes {
		return fmt.Errorf("foundation original output identity mismatch")
	}
	for _, input := range r.Inputs {
		original, exists := capture.Inputs[input.Path]
		if !exists || original.SHA256 != input.SHA256 || original.Bytes != input.Bytes || original.Snapshot == "" {
			return fmt.Errorf("foundation captured input mismatch")
		}
		_, file := s.cache.get(input.Path, 4<<20, false)
		if file.code != "" || file.hash != input.SHA256 || file.size != input.Bytes {
			return fmt.Errorf("foundation current input unavailable or changed")
		}
	}
	_, output := s.cache.get(r.Output.Path, 4<<20, true)
	if output.code != "" || output.hash != r.Output.SHA256 || output.size != r.Output.Bytes || !output.parsed {
		return fmt.Errorf("foundation output unavailable or changed")
	}
	result := output.result
	pkg := s.module + "/" + strings.TrimPrefix(r.Test.Package, "./")
	if !result.successfulSelection(pkg, r.Test.Selected) || !result.successfulSelection(pkg, r.Test.Symbol) {
		return fmt.Errorf("foundation selected result invalid")
	}
	runs, passes := 0, 0
	for _, test := range result.tests {
		runs, passes = runs+test.runs, passes+test.passes
	}
	if capture.SelectedEvents != runs || capture.PassingEvents != passes {
		return fmt.Errorf("foundation original event counts mismatch")
	}
	return nil
}

func (s *sourceFoundationExecutions) capture(pin sourceArtifactPin) (sourceFoundationCapture, error) {
	raw, file := s.cache.get(pin.Path, 64<<20, false)
	if file.code != "" || file.hash != pin.SHA256 || file.size != pin.Bytes {
		return sourceFoundationCapture{}, fmt.Errorf("foundation original capture unavailable or changed")
	}
	if capture, ok := s.captures[pin.Path]; ok {
		return capture, nil
	}
	var capture sourceFoundationCapture
	if decodeSourceJSON(raw, &capture) != nil || decodeStrictJSON(raw, &capture) != nil {
		return capture, fmt.Errorf("foundation original capture invalid")
	}
	// Explicit presence matters for zero-valued terminal fields such as exit0
	// and empty GOFLAGS; omission cannot manufacture a completed original run.
	var members map[string]json.RawMessage
	if json.Unmarshal(raw, &members) != nil {
		return capture, fmt.Errorf("foundation original capture incomplete")
	}
	for _, name := range []string{"name", "command", "cwd", "head", "tree", "tracked_status_before", "go_environment", "inputs", "started_epoch", "completed_epoch", "wall_seconds", "exit_code", "output_sha256", "output_bytes", "selected_test_events", "passing_test_events", "changed_inputs_after"} {
		if value, present := members[name]; !present || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return capture, fmt.Errorf("foundation original capture incomplete")
		}
	}
	var environment map[string]json.RawMessage
	if json.Unmarshal(members["go_environment"], &environment) != nil || len(environment) != 8 {
		return capture, fmt.Errorf("foundation original environment incomplete")
	}
	for _, name := range []string{"GOVERSION", "GOOS", "GOARCH", "CGO_ENABLED", "GOFLAGS", "GOEXPERIMENT", "GOMOD", "GOWORK"} {
		if value, present := environment[name]; !present || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return capture, fmt.Errorf("foundation original environment incomplete")
		}
	}
	finite := func(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= 0 }
	if !validSourceID(capture.Name) || capture.Inputs == nil || capture.ChangedInputs == nil || len(capture.ChangedInputs) != 0 || capture.ExitCode != 0 || !finite(capture.Started) || !finite(capture.Completed) || !finite(capture.Wall) || capture.Completed < capture.Started || capture.SelectedEvents < 1 || capture.PassingEvents < 1 || capture.OutputBytes < 1 || !sourceProofDigest(capture.OutputSHA256) {
		return capture, fmt.Errorf("foundation original capture not successful terminal evidence")
	}
	if !filepath.IsAbs(capture.Cwd) || filepath.Clean(capture.Cwd) != capture.Cwd || len(capture.Head) != 40 || len(capture.Tree) != 40 {
		return capture, fmt.Errorf("foundation original source provenance invalid")
	}
	env := capture.GoEnvironment
	if !strings.HasPrefix(env.Version, "go1.") || (env.OS != "darwin" && env.OS != "linux") || (env.Arch != "arm64" && env.Arch != "amd64") || (env.CGO != "0" && env.CGO != "1") || env.Flags != "" || env.Experiment != "" || env.Module != filepath.Join(capture.Cwd, "go.mod") || (env.Work != "" && env.Work != "off") {
		return capture, fmt.Errorf("foundation original environment not supported")
	}
	s.captures[pin.Path] = capture
	return capture, nil
}

func sourceFoundationCaptureCommand(command []string, test sourceFoundationProofTest) bool {
	if len(command) < 2 || command[0] != "go" || command[1] != "test" {
		return false
	}
	parts := strings.Split(test.Selected, "/")
	for i, part := range parts {
		parts[i] = "^" + regexp.QuoteMeta(part) + "$"
	}
	selected := strings.Join(parts, "/")
	seen := map[string]bool{}
	for i := 2; i < len(command); i++ {
		arg := command[i]
		if seen[arg] {
			return false
		}
		seen[arg] = true
		switch arg {
		case "-json", "-count=1", "-race":
		case "-timeout":
			i++
			if i == len(command) || command[i] != "20m" {
				return false
			}
		case "-run":
			i++
			if i == len(command) || (command[i] != selected && command[i] != "^"+regexp.QuoteMeta(test.Symbol)+"$") {
				return false
			}
		default:
			if arg != test.Package {
				return false
			}
		}
	}
	return seen["-json"] && seen["-count=1"] && seen["-timeout"] && seen["-run"] && seen[test.Package]
}
