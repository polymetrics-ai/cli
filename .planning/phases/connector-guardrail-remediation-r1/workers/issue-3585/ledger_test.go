package issue3585

import (
	"os"
	"strings"
	"testing"
)

type ledgerRow struct {
	pr       string
	mergeSHA string
	path     string
	decision string
}

func TestDispositionLedgerCoversAuditedSharedPaths(t *testing.T) {
	content, err := os.ReadFile("DISPOSITION-LEDGER.md")
	if err != nil {
		t.Fatalf("read disposition ledger: %v", err)
	}

	text := string(content)
	if strings.Contains(strings.ToLower(text), "todo") || strings.Contains(strings.ToLower(text), "pending disposition") {
		t.Fatalf("ledger must not contain TODO or pending disposition placeholders")
	}

	required := []ledgerRow{
		{pr: "#3530", mergeSHA: "86d510927a05aa56b184bf5a8778b5444c69b9b1", path: "internal/connectors/engine/write.go", decision: "preserved_general_foundation"},
		{pr: "#3530", mergeSHA: "86d510927a05aa56b184bf5a8778b5444c69b9b1", path: "internal/connectors/engine/write_test.go", decision: "preserved_test_evidence"},
		{pr: "#3535", mergeSHA: "5d61794f76c42cc7a97a4b29f6ffcc09dd39dbae", path: "cmd/connectorgen/main.go", decision: "preserved_general_foundation"},
		{pr: "#3535", mergeSHA: "5d61794f76c42cc7a97a4b29f6ffcc09dd39dbae", path: "cmd/connectorgen/main_test.go", decision: "preserved_test_evidence"},
		{pr: "#3535", mergeSHA: "5d61794f76c42cc7a97a4b29f6ffcc09dd39dbae", path: "cmd/connectorgen/validate.go", decision: "preserved_general_foundation"},
		{pr: "#3535", mergeSHA: "5d61794f76c42cc7a97a4b29f6ffcc09dd39dbae", path: "internal/cli/cli.go", decision: "preserved_general_foundation"},
		{pr: "#3535", mergeSHA: "5d61794f76c42cc7a97a4b29f6ffcc09dd39dbae", path: "internal/connectors/command_surface.go", decision: "preserved_general_foundation"},
		{pr: "#3535", mergeSHA: "5d61794f76c42cc7a97a4b29f6ffcc09dd39dbae", path: "internal/connectors/commandrunner/runner.go", decision: "preserved_general_foundation"},
		{pr: "#3535", mergeSHA: "5d61794f76c42cc7a97a4b29f6ffcc09dd39dbae", path: "internal/connectors/commandrunner/runner_test.go", decision: "preserved_test_evidence"},
		{pr: "#3535", mergeSHA: "5d61794f76c42cc7a97a4b29f6ffcc09dd39dbae", path: "internal/connectors/engine/bundle.go", decision: "preserved_general_foundation"},
		{pr: "#3535", mergeSHA: "5d61794f76c42cc7a97a4b29f6ffcc09dd39dbae", path: "internal/connectors/engine/connector.go", decision: "preserved_general_foundation"},
		{pr: "#3535", mergeSHA: "5d61794f76c42cc7a97a4b29f6ffcc09dd39dbae", path: "internal/connectors/engine/schema/cli_surface.schema.json", decision: "preserved_general_foundation"},
		{pr: "#3535", mergeSHA: "5d61794f76c42cc7a97a4b29f6ffcc09dd39dbae", path: "internal/connectors/hooks/google-ads/hooks.go", decision: "connector_owned_allowed"},
		{pr: "#3535", mergeSHA: "5d61794f76c42cc7a97a4b29f6ffcc09dd39dbae", path: "internal/connectors/hooks/google-ads/hooks_test.go", decision: "connector_owned_allowed"},
		{pr: "#3535", mergeSHA: "5d61794f76c42cc7a97a4b29f6ffcc09dd39dbae", path: "internal/connectors/defs/gong/cli_surface.json", decision: "delegated_to_3586"},
		{pr: "#3536", mergeSHA: "b053dc4a3ad7f9895637a09560ea8a9a76bec507", path: "internal/connectors/commandrunner/runner.go", decision: "preserved_general_foundation"},
		{pr: "#3536", mergeSHA: "b053dc4a3ad7f9895637a09560ea8a9a76bec507", path: "internal/connectors/commandrunner/runner_test.go", decision: "preserved_test_evidence"},
	}

	for _, row := range required {
		if !hasLedgerRow(text, row) {
			t.Fatalf("missing ledger row for PR %s path %s decision %s", row.pr, row.path, row.decision)
		}
	}
}

func hasLedgerRow(text string, row ledgerRow) bool {
	for _, line := range strings.Split(text, "\n") {
		if !strings.Contains(line, "|") {
			continue
		}
		if strings.Contains(line, row.pr) &&
			strings.Contains(line, row.mergeSHA) &&
			strings.Contains(line, row.path) &&
			strings.Contains(line, row.decision) {
			return true
		}
	}
	return false
}
