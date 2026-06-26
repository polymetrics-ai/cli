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
	for _, want := range []string{`"kind": "ConnectorCatalog"`, `"count": 647`, `"enabled": 2`, `"planned_native_port": 645`, `"slug": "source-github"`, `"implementation_status": "enabled"`, `"slug": "source-stripe"`, `"pm_connector_name": "stripe"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("catalog json missing %q:\n%s", want, out[:min(len(out), 2000)])
		}
	}
}

func TestConnectorCatalogFiltersAndInspect(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"connectors", "catalog", "--type", "destination", "--stage", "generally_available", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(connectors catalog) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{`"kind": "ConnectorCatalog"`, `"count": 9`, `"slug": "destination-postgres"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("filtered catalog missing %q:\n%s", want, out)
		}
	}

	stdout.Reset()
	stderr.Reset()
	code = cli.Run([]string{"connectors", "inspect", "destination-postgres"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(connectors inspect destination-postgres) code = %d stderr = %s", code, stderr.String())
	}
	manual := stdout.String()
	for _, want := range []string{"NAME", "IMPLEMENTATION STATUS", "RUNTIME CAPABILITIES", "metadata=true", "read=false", "write=false", "query=false", "planned_native_port", "destination_go", "https://docs.airbyte.com/integrations/destinations/postgres"} {
		if !strings.Contains(manual, want) {
			t.Fatalf("catalog manual missing %q:\n%s", want, manual)
		}
	}
}

func TestConnectorPortPlanCLI(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"connectors", "port-plan", "--all", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(connectors port-plan --all --json) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{`"kind": "NativePortPlanList"`, `"count": 647`, `"slug": "source-postgres"`, `"family": "database_cdc_source"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("port plan json missing %q:\n%s", want, out[:min(len(out), 2000)])
		}
	}

	stdout.Reset()
	stderr.Reset()
	code = cli.Run([]string{"connectors", "port-plan", "source-postgres"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(connectors port-plan source-postgres) code = %d stderr = %s", code, stderr.String())
	}
	manual := stdout.String()
	for _, want := range []string{"NAME", "NATIVE PORT PLAN", "CDC", "postgres_logical_replication", "wal_level=logical", "CONFORMANCE"} {
		if !strings.Contains(manual, want) {
			t.Fatalf("port plan manual missing %q:\n%s", want, manual)
		}
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
		filepath.Join(connectorsDir, "destination-postgres", "MANUAL.md"),
		filepath.Join(connectorsDir, "source-github", "MANUAL.md"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected generated doc %s: %v", path, err)
		}
	}
	catalogMarkdown, err := os.ReadFile(filepath.Join(connectorsDir, "catalog", "all-connectors.md"))
	if err != nil {
		t.Fatalf("read generated connector catalog markdown: %v", err)
	}
	for _, want := range []string{"Official Application Documentation", "Airbyte Connector Documentation", "docs.github.com", "docs.airbyte.com/integrations/sources/github"} {
		if !strings.Contains(string(catalogMarkdown), want) {
			t.Fatalf("generated catalog markdown missing %q", want)
		}
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
