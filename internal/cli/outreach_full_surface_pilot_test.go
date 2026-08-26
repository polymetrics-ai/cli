package cli

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// TestPMBinaryOutreachRepresentativeCommandsReachCredentialBoundary proves
// public command reachability without configuring credentials or allowing
// provider I/O.  A valid path must reach this boundary before any mapping,
// certification, hash, or live-certification policy can reject it.
func TestPMBinaryOutreachRepresentativeCommandsReachCredentialBoundary(t *testing.T) {
	binary := buildTransportPM(t)
	root := filepath.Join(t.TempDir(), "project")
	mustRunFixtureTransportPM(t, binary, "outreach-pilot-fixture-token", "init", "--root", root, "--json")

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "etl read",
			args: []string{"outreach", "prospects", "list", "--root", root, "--json"},
		},
		{
			name: "reverse etl direct write",
			args: []string{"outreach", "create", "account", "note", "apply", "--data", `{"type":"accountNotes","attributes":{}}`, "--root", root, "--json"},
		},
		{
			name: "destructive delete",
			args: []string{"outreach", "delete", "account", "apply", "--id", "fixture-account", "--confirm", "destructive", "--root", root, "--json"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output, diagnostics, err := runFixtureTransportPMJSON(binary, "", tc.args...)
			if err == nil {
				t.Fatalf("pm %s unexpectedly succeeded\nstdout:\n%s\nstderr:\n%s", transportCommandName(tc.args), output, diagnostics)
			}
			combined := output + diagnostics
			if !strings.Contains(combined, "missing --credential") {
				t.Fatalf("pm %s = %v\nstdout:\n%s\nstderr:\n%s\nwant credential boundary", transportCommandName(tc.args), err, output, diagnostics)
			}
		})
	}
}

// TestPMBinaryOutreachFixtureBindsDeclaredMethodAndPath ensures the public
// command has no caller-controlled method/path/source identity override.  The
// only successful request in this test is a local fixture request whose route
// is the method and path declared by the Outreach stream; no provider is
// contacted and the fixture token is synthetic.
func TestPMBinaryOutreachFixtureBindsDeclaredMethodAndPath(t *testing.T) {
	const token = "outreach-pilot-fixture-token"
	const tokenEnv = "PM_OUTREACH_PILOT_FIXTURE_TOKEN"
	t.Setenv(tokenEnv, token)

	var observed struct {
		sync.Mutex
		requests []string
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if got, want := request.Header.Get("Authorization"), "Bearer "+token; got != want {
			http.Error(w, "fixture request did not carry declared bearer authentication", http.StatusUnauthorized)
			return
		}
		observed.Lock()
		observed.requests = append(observed.requests, request.Method+" "+request.URL.EscapedPath())
		observed.Unlock()
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{"data":[{"id":"fixture-prospect","type":"prospects","attributes":{"email":"fixture@example.test","name":"Fixture Prospect","createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-01T00:00:00Z"}}],"links":{"next":null}}`))
	}))
	t.Cleanup(server.Close)

	binary := buildTransportPM(t)
	root := filepath.Join(t.TempDir(), "project")
	mustRunFixtureTransportPM(t, binary, token, "init", "--root", root, "--json")
	mustRunFixtureTransportPM(t, binary, token,
		"credentials", "add", "outreach-fixture", "--connector", "outreach",
		"--config", "base_url="+server.URL+"/api/v2",
		"--from-env", "access_token="+tokenEnv,
		"--root", root, "--json",
	)

	for _, override := range []struct {
		name string
		args []string
	}{
		{name: "method", args: []string{"--method", "DELETE"}},
		{name: "path", args: []string{"--path", "/api/v2/accounts"}},
		{name: "source url", args: []string{"--source-url", "https://example.invalid/openapi.json"}},
	} {
		t.Run("rejects caller supplied "+override.name, func(t *testing.T) {
			args := append([]string{"outreach", "prospects", "list", "--credential", "outreach-fixture", "--root", root, "--json"}, override.args...)
			output, diagnostics, err := runFixtureTransportPMJSON(binary, "", args...)
			if err == nil {
				t.Fatalf("pm %s unexpectedly accepted override\nstdout:\n%s\nstderr:\n%s", transportCommandName(args), output, diagnostics)
			}
			if !strings.Contains(output+diagnostics, "unknown flag --") {
				t.Fatalf("pm %s = %v\nstdout:\n%s\nstderr:\n%s\nwant unknown-flag refusal", transportCommandName(args), err, output, diagnostics)
			}
			observed.Lock()
			got := len(observed.requests)
			observed.Unlock()
			if got != 0 {
				t.Fatalf("caller supplied %s reached fixture transport: %v", override.name, observed.requests)
			}
		})
	}

	output, diagnostics, err := runFixtureTransportPMJSON(binary, "",
		"outreach", "prospects", "list", "--credential", "outreach-fixture", "--limit", "1", "--root", root, "--json",
	)
	if err != nil {
		t.Fatalf("pm outreach prospects list failed: %v\nstdout:\n%s\nstderr:\n%s", err, output, diagnostics)
	}
	observed.Lock()
	got := append([]string(nil), observed.requests...)
	observed.Unlock()
	if want := []string{"GET /api/v2/prospects"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("fixture requests = %v, want %v", got, want)
	}
}
