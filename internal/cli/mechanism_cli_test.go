package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"polymetrics.ai/internal/cli"
)

// TestConnectorInspectSurfacesMechanism proves the mechanism field is wired
// end-to-end through the real registry, not just at the engine/connectors
// package boundary: every bundle that declares no metadata.json "mechanism"
// block (every bundle today — this PR adds no connectors) still gets the
// conservative official_api default at load time, and that default is
// visible in `pm connectors inspect --json` output.
func TestConnectorInspectSurfacesMechanism(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"connectors", "inspect", "github", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(connectors inspect github --json) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{`"mechanism"`, `"kind": "official_api"`, `"sanctioned_by_provider": true`} {
		if !strings.Contains(out, want) {
			t.Fatalf("inspect json missing %q:\n%s", want, out[:min(len(out), 2000)])
		}
	}
}

// TestConnectorListTextModeShowsNoMarkerForOfficialConnectors is the
// negative half of the [UNOFFICIAL] marker check: nothing shipped today is
// a web_session mechanism, so the text-mode connector list must never print
// the marker for any of them.
func TestConnectorListTextModeShowsNoMarkerForOfficialConnectors(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"connectors", "list"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(connectors list) code = %d stderr = %s", code, stderr.String())
	}
	if strings.Contains(stdout.String(), "[UNOFFICIAL]") {
		t.Fatalf("connectors list printed [UNOFFICIAL] for a shipped official-only connector set:\n%s", stdout.String())
	}
}
