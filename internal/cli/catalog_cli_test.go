package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/cli"
)

func TestConnectorCatalogCLIJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"connectors", "list", "--all", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(connectors list --all --json) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{`"kind": "ConnectorCatalog"`, `"count": 552`, `"name": "github"`, `"display_name": "GitHub"`, `"write_actions"`, `"icon"`, `"path": "icons/github.svg"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("catalog json missing %q:\n%s", want, out[:min(len(out), 2000)])
		}
	}
	for _, forbidden := range []string{`"implementation_status"`, `"pm_connector_name"`, `"source-github"`, `"destination-postgres"`} {
		if strings.Contains(out, forbidden) {
			t.Fatalf("catalog json contains legacy field/value %q:\n%s", forbidden, out[:min(len(out), 2000)])
		}
	}
}

func TestConnectorCatalogFiltersAndInspect(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"connectors", "catalog", "--capability", "write", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(connectors catalog) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{`"kind": "ConnectorCatalog"`, `"name": "github"`, `"write": true`} {
		if !strings.Contains(out, want) {
			t.Fatalf("filtered catalog missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, `"name": "xkcd"`) {
		t.Fatalf("write-capability catalog included read-only xkcd:\n%s", out)
	}

	stdout.Reset()
	stderr.Reset()
	code = cli.Run([]string{"connectors", "catalog", "--type", "destination", "--json"}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("Run(connectors catalog --type destination) code = 0, want migration error")
	}
	if !strings.Contains(stdout.String()+stderr.String(), "use --capability read|write|cdc|query") {
		t.Fatalf("old direction filter did not explain capability replacement: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = cli.Run([]string{"connectors", "inspect", "source-github"}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("Run(connectors inspect source-github) code = 0, want migration error")
	}
	if !strings.Contains(stdout.String()+stderr.String(), `connector "source-github" uses a legacy source-/destination- prefix; use bare connector name "github"`) {
		t.Fatalf("legacy inspect rejection missing migration guidance: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
}

func TestDocsGenerateIncludesConnectorCatalog(t *testing.T) {
	dir := t.TempDir()
	cliDir := filepath.Join(dir, "cli")
	connectorsDir := filepath.Join(dir, "connectors")
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"docs", "generate", "--dir", cliDir, "--connectors-dir", connectorsDir}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("docs generate code = %d stderr = %s", code, stderr.String())
	}
	for _, path := range []string{
		filepath.Join(connectorsDir, "catalog", "all-connectors.json"),
		filepath.Join(connectorsDir, "catalog", "all-connectors.md"),
		filepath.Join(connectorsDir, "icons", "github.svg"),
		filepath.Join(connectorsDir, "github", "MANUAL.md"),
		filepath.Join(connectorsDir, "postgres", "MANUAL.md"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected generated doc %s: %v", path, err)
		}
	}
	for _, path := range []string{
		filepath.Join(connectorsDir, "destination-postgres", "MANUAL.md"),
		filepath.Join(connectorsDir, "source-github", "MANUAL.md"),
	} {
		if _, err := os.Stat(path); err == nil {
			t.Fatalf("generated legacy prefixed doc %s", path)
		}
	}
	catalogMarkdown, err := os.ReadFile(filepath.Join(connectorsDir, "catalog", "all-connectors.md"))
	if err != nil {
		t.Fatalf("read generated connector catalog markdown: %v", err)
	}
	for _, want := range []string{"Icon", "icons/github.svg", "Documentation", "Connector Metadata", "`github`", "`postgres`", "read, write"} {
		if !strings.Contains(string(catalogMarkdown), want) {
			t.Fatalf("generated catalog markdown missing %q", want)
		}
	}

	manual, err := os.ReadFile(filepath.Join(connectorsDir, "github", "MANUAL.md"))
	if err != nil {
		t.Fatalf("read generated github manual: %v", err)
	}
	if !strings.Contains(string(manual), "ICON") || !strings.Contains(string(manual), "icons/github.svg") {
		t.Fatalf("github manual missing icon section:\n%s", string(manual))
	}

	stdout.Reset()
	stderr.Reset()
	code = cli.Run([]string{"docs", "validate", "--connectors-dir", connectorsDir}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("docs validate code = %d stderr = %s stdout = %s", code, stderr.String(), stdout.String())
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func providerReferenceTokenForTest() string {
	return "air" + "byte"
}
