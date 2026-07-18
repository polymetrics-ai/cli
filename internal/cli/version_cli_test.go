package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestVersionDeterministicPlainAndJSONOutput(t *testing.T) {
	const wantPlain = "pm dev\ncommit: none\nbuilt: unknown\n"

	var firstPlain, firstPlainErr bytes.Buffer
	if code := Run([]string{"version"}, &firstPlain, &firstPlainErr); code != 0 {
		t.Fatalf("version exit = %d, want 0; stdout=%s stderr=%s", code, firstPlain.String(), firstPlainErr.String())
	}
	if firstPlain.String() != wantPlain || firstPlainErr.Len() != 0 {
		t.Fatalf("version output mismatch:\nstdout=%q\nstderr=%q", firstPlain.String(), firstPlainErr.String())
	}

	var secondPlain, secondPlainErr bytes.Buffer
	if code := Run([]string{"version"}, &secondPlain, &secondPlainErr); code != 0 {
		t.Fatalf("second version exit = %d, want 0; stdout=%s stderr=%s", code, secondPlain.String(), secondPlainErr.String())
	}
	if secondPlain.String() != firstPlain.String() || secondPlainErr.String() != firstPlainErr.String() {
		t.Fatalf("plain version output is not deterministic:\nfirst=%q/%q\nsecond=%q/%q", firstPlain.String(), firstPlainErr.String(), secondPlain.String(), secondPlainErr.String())
	}

	var firstJSON, firstJSONErr bytes.Buffer
	if code := Run([]string{"version", "--json"}, &firstJSON, &firstJSONErr); code != 0 {
		t.Fatalf("version --json exit = %d, want 0; stdout=%s stderr=%s", code, firstJSON.String(), firstJSONErr.String())
	}
	if firstJSONErr.Len() != 0 {
		t.Fatalf("version --json stderr = %q, want empty", firstJSONErr.String())
	}
	var got struct {
		APIVersion string `json:"api_version"`
		Kind       string `json:"kind"`
		Version    string `json:"version"`
		Commit     string `json:"commit"`
		Date       string `json:"date"`
	}
	if err := json.Unmarshal(firstJSON.Bytes(), &got); err != nil {
		t.Fatalf("decode version JSON: %v\n%s", err, firstJSON.String())
	}
	if got.APIVersion != apiVersion || got.Kind != "Version" || got.Version != "dev" || got.Commit != "none" || got.Date != "unknown" {
		t.Fatalf("version JSON = %+v, want deterministic development metadata", got)
	}

	var secondJSON, secondJSONErr bytes.Buffer
	if code := Run([]string{"--json", "version"}, &secondJSON, &secondJSONErr); code != 0 {
		t.Fatalf("second version --json exit = %d, want 0; stdout=%s stderr=%s", code, secondJSON.String(), secondJSONErr.String())
	}
	if secondJSON.String() != firstJSON.String() || secondJSONErr.String() != firstJSONErr.String() {
		t.Fatalf("JSON version output is not deterministic:\nfirst=%q/%q\nsecond=%q/%q", firstJSON.String(), firstJSONErr.String(), secondJSON.String(), secondJSONErr.String())
	}
}

func TestVersionJSONBooleanAssignmentsStayConnected(t *testing.T) {
	const wantPlain = "pm dev\ncommit: none\nbuilt: unknown\n"

	t.Run("true selects JSON", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		if code := Run([]string{"version", "--json=true"}, &stdout, &stderr); code != 0 {
			t.Fatalf("version --json=true exit = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
		}
		if stderr.Len() != 0 {
			t.Fatalf("version --json=true stderr = %q, want empty", stderr.String())
		}
		var env struct {
			Kind string `json:"kind"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
			t.Fatalf("version --json=true stdout not JSON: %v\n%s", err, stdout.String())
		}
		if env.Kind != "Version" {
			t.Fatalf("version --json=true kind = %q, want Version", env.Kind)
		}
	})

	t.Run("false overrides configured JSON", func(t *testing.T) {
		t.Setenv("POLYMETRICS_JSON", "true")
		var stdout, stderr bytes.Buffer
		if code := Run([]string{"version", "--json=false"}, &stdout, &stderr); code != 0 {
			t.Fatalf("version --json=false exit = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
		}
		if stdout.String() != wantPlain || stderr.Len() != 0 {
			t.Fatalf("version --json=false output mismatch:\nstdout=%q\nstderr=%q", stdout.String(), stderr.String())
		}
	})

	t.Run("true selects JSON help", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		if code := Run([]string{"version", "--help", "--json=true"}, &stdout, &stderr); code != 0 {
			t.Fatalf("version --help --json=true exit = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
		}
		if stderr.Len() != 0 {
			t.Fatalf("version --help --json=true stderr = %q, want empty", stderr.String())
		}
		var manual struct {
			Kind    string `json:"kind"`
			Command string `json:"command"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &manual); err != nil {
			t.Fatalf("version --help --json=true stdout not JSON: %v\n%s", err, stdout.String())
		}
		if manual.Kind != "CommandManual" || manual.Command != "version" {
			t.Fatalf("version --help --json=true manual = %+v, want version CommandManual", manual)
		}
	})
}

func TestVersionFlagAndPositionalHelpCompatibility(t *testing.T) {
	var canonicalOut, canonicalErr bytes.Buffer
	if code := Run([]string{"help", "version"}, &canonicalOut, &canonicalErr); code != 0 {
		t.Fatalf("help version exit = %d, want 0; stdout=%s stderr=%s", code, canonicalOut.String(), canonicalErr.String())
	}
	if canonicalErr.Len() != 0 {
		t.Fatalf("help version stderr = %q, want empty", canonicalErr.String())
	}

	for _, args := range [][]string{{"version", "--help"}, {"version", "-h"}, {"version", "help"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := Run(args, &stdout, &stderr); code != 0 {
				t.Fatalf("Run(%v) exit = %d, want 0; stdout=%s stderr=%s", args, code, stdout.String(), stderr.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("Run(%v) stderr = %q, want empty", args, stderr.String())
			}
			if stdout.String() != canonicalOut.String() {
				t.Fatalf("Run(%v) help mismatch:\nwant:\n%s\ngot:\n%s", args, canonicalOut.String(), stdout.String())
			}
		})
	}

	for _, args := range [][]string{{"version", "--help", "--json"}, {"version", "help", "--json"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := Run(args, &stdout, &stderr); code != 0 {
				t.Fatalf("Run(%v) exit = %d, want 0; stdout=%s stderr=%s", args, code, stdout.String(), stderr.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("Run(%v) stderr = %q, want empty", args, stderr.String())
			}
			var manual struct {
				APIVersion string `json:"api_version"`
				Kind       string `json:"kind"`
				Command    string `json:"command"`
				Manual     string `json:"manual"`
			}
			if err := json.Unmarshal(stdout.Bytes(), &manual); err != nil {
				t.Fatalf("Run(%v) stdout not JSON: %v\n%s", args, err, stdout.String())
			}
			if manual.APIVersion != apiVersion || manual.Kind != "CommandManual" || manual.Command != "version" || manual.Manual != canonicalOut.String() {
				t.Fatalf("Run(%v) manual = %+v, want canonical version CommandManual", args, manual)
			}
		})
	}
}

func TestVersionUnknownFlagAndInvalidActionRemainUsageErrors(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
	}{
		{name: "unknown flag", args: []string{"version", "--definitely-unknown", "--json"}},
		{name: "invalid action", args: []string{"version", "bogus", "--json"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := Run(tt.args, &stdout, &stderr); code != 2 {
				t.Fatalf("Run(%v) exit = %d, want 2; stdout=%s stderr=%s", tt.args, code, stdout.String(), stderr.String())
			}
			var env struct {
				Kind  string `json:"kind"`
				Error struct {
					Category string `json:"category"`
				} `json:"error"`
			}
			if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
				t.Fatalf("Run(%v) stdout not JSON: %v\n%s", tt.args, err, stdout.String())
			}
			if env.Kind != "Error" || env.Error.Category != "usage" {
				t.Fatalf("Run(%v) envelope = %+v, want usage Error", tt.args, env)
			}
			if strings.Contains(stdout.String(), `"kind": "CommandManual"`) {
				t.Fatalf("Run(%v) rendered manual instead of usage error: %s", tt.args, stdout.String())
			}
			if stderr.Len() == 0 {
				t.Fatalf("Run(%v) stderr empty, want usage diagnostic", tt.args)
			}
		})
	}
}
