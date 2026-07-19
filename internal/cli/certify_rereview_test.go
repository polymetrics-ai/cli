package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors/certify"
)

func TestRereviewSweepCredentialConstraintsRejectBeforeProjectEffects(t *testing.T) {
	cases := []struct {
		name     string
		defaults string
		entry    string
	}{
		{name: "default rate", defaults: "    rate_limit_rps: 2\n"},
		{name: "default budget", defaults: "    budget_calls: 5\n"},
		{name: "default limit", defaults: "    limit: 10\n"},
		{name: "connector rate", entry: "    rate_limit_rps: 2\n"},
		{name: "connector budget", entry: "    budget_calls: 5\n"},
		{name: "connector limit", entry: "    limit: 10\n"},
		{name: "write without sandbox", entry: "    write: true\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "project")
			credsPath := filepath.Join(t.TempDir(), "creds.yaml")
			raw := "version: 1\n"
			if tc.defaults != "" {
				raw += "defaults:\n" + tc.defaults
			}
			raw += "connectors:\n  sample:\n" + tc.entry
			if err := os.WriteFile(credsPath, []byte(raw), 0o600); err != nil {
				t.Fatal(err)
			}
			_, err := (defaultCertifyCommandRuntime{}).Sweep(context.Background(), root, credsPath, time.Hour)
			if err == nil || !strings.Contains(err.Error(), "not supported") {
				t.Fatalf("sweep constraint error=%v, want fail-closed unsupported error", err)
			}
			if _, statErr := os.Stat(filepath.Join(root, ".polymetrics")); !os.IsNotExist(statErr) {
				t.Fatalf("invalid sweep created project effects: %v", statErr)
			}
		})
	}
}

func TestRereviewSweepConstraintsHaveNoTelemetryEffects(t *testing.T) {
	root := t.TempDir()
	credsPath := filepath.Join(t.TempDir(), "creds.yaml")
	raw := "version: 1\ndefaults:\n  rate_limit_rps: 2\nconnectors:\n  sample: {}\n"
	if err := os.WriteFile(credsPath, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PM_TELEMETRY", "file")
	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"connectors", "certify", "--sweep", "--credentials-file", credsPath,
		"--root", root, "--json",
	}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("invalid sweep constraint exit=%d, want usage 2", code)
	}
	if entries, err := os.ReadDir(filepath.Join(root, ".polymetrics", "telemetry")); err == nil && len(entries) != 0 {
		t.Fatal("invalid sweep constraint initialized telemetry")
	} else if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
}

func TestRereviewSingleReportPersistenceFailureIsNotSuccess(t *testing.T) {
	for _, tc := range []struct {
		name  string
		seed  func(t *testing.T, root string)
		leaks bool
	}{
		{
			name: "current symlink",
			seed: func(t *testing.T, root string) {
				t.Helper()
				certDir := filepath.Join(root, ".polymetrics", "certifications")
				if err := os.MkdirAll(certDir, 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Join(t.TempDir(), "outside.json"), filepath.Join(certDir, "sample.json")); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "certifications path unwritable shape",
			seed: func(t *testing.T, root string) {
				t.Helper()
				if err := os.MkdirAll(filepath.Join(root, ".polymetrics"), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(root, ".polymetrics", "certifications"), nil, 0o600); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "history symlink",
			seed: func(t *testing.T, root string) {
				t.Helper()
				certDir := filepath.Join(root, ".polymetrics", "certifications")
				if err := os.MkdirAll(certDir, 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(t.TempDir(), filepath.Join(certDir, "history")); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name:  "leak remains dominant",
			leaks: true,
			seed: func(t *testing.T, root string) {
				t.Helper()
				certDir := filepath.Join(root, ".polymetrics", "certifications")
				if err := os.MkdirAll(certDir, 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Join(t.TempDir(), "outside.json"), filepath.Join(certDir, "sample.json")); err != nil {
					t.Fatal(err)
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			tc.seed(t, root)
			rep := passingCLIReport("sample")
			rep.StartedAt = time.Now().Add(-time.Second).UTC()
			rep.CompletedAt = time.Now().UTC()
			if tc.leaks {
				rep.Passed = false
				rep.Leaks = []certify.Leak{{Tag: "pm-cert-sample-12345678-1700000000", Connector: "sample", Reason: "cleanup verification failed"}}
			}
			runtime := &fakeCertifyCommandRuntime{singleReport: rep}
			stdout, stderr, code := executeNativeConnectors(context.Background(), testRouterConfig(root, true), runtime, "connectors", "certify", "sample")
			if tc.leaks {
				if code != 3 {
					t.Fatalf("leaked report persistence exit=%d, want leak-dominant 3; stdout=%q stderr=%q", code, stdout, stderr)
				}
				if !strings.Contains(stdout, "report_persistence") {
					t.Fatal("leak-dominant result did not record report persistence failure")
				}
				return
			}
			if code == 0 {
				t.Fatalf("report persistence failure returned success; stdout=%q stderr=%q", stdout, stderr)
			}
			if !strings.Contains(stdout+stderr, "certif") {
				t.Fatal("report persistence failure was not surfaced")
			}
		})
	}
}

func TestRereviewInvalidCertifyArgsHaveNoLoggerOrTelemetryEffects(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		wantCode int
	}{
		{name: "unknown assigned", args: []string{"connectors", "certify", "sample", "--writ=false"}, wantCode: 2},
		{name: "unknown space value", args: []string{"connectors", "certify", "sample", "--writ", "false"}, wantCode: 2},
		{name: "malformed boolean space value", args: []string{"connectors", "certify", "sample", "--write", "maybe"}, wantCode: 2},
		{name: "malformed parallel space value", args: []string{"connectors", "certify", "--all", "--credentials-file", "unused.yaml", "--parallel", "nope"}, wantCode: 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("PM_TELEMETRY", "file")
			args := append(append([]string{}, tc.args...), "--root", root, "--json")
			var stdout, stderr bytes.Buffer
			if code := Run(args, &stdout, &stderr); code != tc.wantCode {
				t.Fatalf("invalid args exit=%d, want %d; stdout=%q stderr=%q", code, tc.wantCode, stdout.String(), stderr.String())
			}
			for _, effectPath := range []string{
				filepath.Join(root, ".polymetrics", "telemetry"),
				filepath.Join(root, ".polymetrics", "logs"),
			} {
				if entries, err := os.ReadDir(effectPath); err == nil && len(entries) != 0 {
					t.Fatalf("invalid certify args created files under %s", effectPath)
				} else if err != nil && !os.IsNotExist(err) {
					t.Fatalf("inspect %s: %v", effectPath, err)
				}
			}
		})
	}
}
