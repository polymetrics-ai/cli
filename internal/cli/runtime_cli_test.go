package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRuntimeMalformedKnownGlobalFlagsAreUsageErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "runtime missing root value", args: []string{"--json", "runtime", "--root"}},
		{name: "runtime doctor missing root value", args: []string{"--json", "runtime", "doctor", "--root"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run(tt.args, &stdout, &stderr)
			if code != 2 {
				t.Fatalf("Run(%v) exit = %d, want 2; stdout=%s stderr=%s", tt.args, code, stdout.String(), stderr.String())
			}
			if !strings.Contains(stdout.String(), `"category": "usage"`) {
				t.Fatalf("Run(%v) stdout missing usage JSON:\n%s", tt.args, stdout.String())
			}
			if strings.Contains(stdout.String(), `"category": "internal"`) {
				t.Fatalf("Run(%v) misclassified Cobra parse error as internal:\n%s", tt.args, stdout.String())
			}
			if !strings.Contains(stderr.String(), "flag needs an argument") {
				t.Fatalf("Run(%v) stderr missing pflag parse diagnostic:\n%s", tt.args, stderr.String())
			}
		})
	}
}

func TestRuntimeBareHelpAndInvalidActionSemantics(t *testing.T) {
	var helpOut, helpErr bytes.Buffer
	if code := Run([]string{"help", "runtime"}, &helpOut, &helpErr); code != 0 {
		t.Fatalf("help runtime exit = %d, stderr = %s", code, helpErr.String())
	}
	if helpErr.Len() != 0 {
		t.Fatalf("help runtime stderr = %q, want empty", helpErr.String())
	}
	if !strings.Contains(helpOut.String(), "pm runtime - inspect external runtime dependencies") {
		t.Fatalf("help runtime output missing runtime manual:\n%s", helpOut.String())
	}

	for _, args := range [][]string{{"runtime"}, {"runtime", "--help"}, {"runtime", "help"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := Run(args, &stdout, &stderr); code != 0 {
				t.Fatalf("Run(%v) exit = %d, stderr = %s", args, code, stderr.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("Run(%v) stderr = %q, want empty", args, stderr.String())
			}
			if stdout.String() != helpOut.String() {
				t.Fatalf("Run(%v) help changed:\nwant:\n%s\ngot:\n%s", args, helpOut.String(), stdout.String())
			}
		})
	}

	for _, args := range [][]string{{"runtime", "--json"}, {"runtime", "help", "--json"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := Run(args, &stdout, &stderr); code != 0 {
				t.Fatalf("Run(%v) exit = %d, stderr = %s", args, code, stderr.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("Run(%v) stderr = %q, want empty", args, stderr.String())
			}
			var manual struct {
				Kind    string `json:"kind"`
				Command string `json:"command"`
			}
			if err := json.Unmarshal(stdout.Bytes(), &manual); err != nil {
				t.Fatalf("Run(%v) stdout not JSON: %v — %s", args, err, stdout.String())
			}
			if manual.Kind != "CommandManual" || manual.Command != "runtime" {
				t.Fatalf("Run(%v) manual envelope = %+v, want runtime CommandManual", args, manual)
			}
		})
	}

	var badOut, badErr bytes.Buffer
	if code := Run([]string{"runtime", "bogus", "--json"}, &badOut, &badErr); code != 2 {
		t.Fatalf("runtime bogus exit = %d, want 2; stdout=%s stderr=%s", code, badOut.String(), badErr.String())
	}
	if !strings.Contains(badOut.String(), `"category": "usage"`) {
		t.Fatalf("runtime bogus stdout missing usage JSON:\n%s", badOut.String())
	}
	if strings.Contains(badOut.String(), `"kind": "CommandManual"`) {
		t.Fatalf("runtime bogus rendered contextual help instead of usage error:\n%s", badOut.String())
	}
}

func TestRuntimeDoctorNativeCobraPreservesLegacySemantics(t *testing.T) {
	root := writeMigrationConfig(t, `runtime:
  postgres_url: postgres://user:secret@127.0.0.1:1/polymetrics?sslmode=disable
  dragonfly_addr: 127.0.0.1:2
  temporal_addr: 127.0.0.1:3
`)

	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"runtime", "doctor",
		"--unknown", "ignored",
		"extra-positional",
		"--root", root,
		"--json",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runtime doctor exit = %d, stderr = %s stdout = %s", code, stderr.String(), stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("runtime doctor stderr = %q, want empty", stderr.String())
	}
	if strings.Contains(stdout.String(), "secret") {
		t.Fatalf("runtime doctor leaked postgres password:\n%s", stdout.String())
	}

	var env struct {
		Kind   string         `json:"kind"`
		Config map[string]any `json:"config"`
		Report struct {
			Checks []struct {
				Name     string `json:"name"`
				Endpoint string `json:"endpoint"`
			} `json:"checks"`
		} `json:"report"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("runtime doctor stdout not JSON: %v — %s", err, stdout.String())
	}
	if env.Kind != "RuntimeDoctor" {
		t.Fatalf("kind = %q, want RuntimeDoctor", env.Kind)
	}
	if env.Config["postgres_url"] != "postgres://***@127.0.0.1:1/polymetrics" {
		t.Fatalf("postgres_url = %v, want redacted configured endpoint", env.Config["postgres_url"])
	}
	if env.Config["dragonfly_addr"] != "127.0.0.1:2" || env.Config["temporal_addr"] != "127.0.0.1:3" {
		t.Fatalf("runtime config = %v, want configured endpoints", env.Config)
	}
	endpoints := map[string]string{}
	for _, check := range env.Report.Checks {
		endpoints[check.Name] = check.Endpoint
	}
	if endpoints["postgres"] != "postgres://***@127.0.0.1:1/polymetrics" {
		t.Fatalf("postgres endpoint = %q, want redacted endpoint", endpoints["postgres"])
	}
	if endpoints["dragonfly"] != "127.0.0.1:2" {
		t.Fatalf("dragonfly endpoint = %q, want configured endpoint", endpoints["dragonfly"])
	}
	if endpoints["temporal"] != "127.0.0.1:3" {
		t.Fatalf("temporal endpoint = %q, want configured endpoint", endpoints["temporal"])
	}
}
