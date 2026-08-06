package cli_test

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/cli"
	"polymetrics.ai/internal/connectors"
)

func TestETLRunRateLimitOutputIsStructuredHumanReadableAndSecretFree(t *testing.T) {
	root := t.TempDir()
	const token = "etl-rate-limit-token-must-not-escape"
	t.Setenv("PM_ETL_RATE_LIMIT_TOKEN", token)

	for _, args := range [][]string{
		{"init", "--root", root, "--json"},
		{"credentials", "add", "sample-rate-limit", "--connector", "sample", "--from-env", "token=PM_ETL_RATE_LIMIT_TOKEN", "--root", root, "--json"},
		{"credentials", "add", "warehouse-rate-limit", "--connector", "warehouse", "--config", "path=" + filepath.Join(root, ".polymetrics", "warehouse"), "--root", root, "--json"},
		{"connections", "create", "rate_limit_to_warehouse", "--source", "sample:sample-rate-limit", "--destination", "warehouse:warehouse-rate-limit", "--stream", "customers", "--primary-key", "id", "--cursor", "updated_at", "--table", "rate_limit_customers", "--root", root, "--json"},
	} {
		var stdout, stderr bytes.Buffer
		if code := cli.Run(args, &stdout, &stderr); code != 0 {
			t.Fatalf("Run(%v) code = %d stderr=%s stdout=%s", args, code, stderr.String(), stdout.String())
		}
	}

	human, humanErr := runETLRateLimitCLI(t, root, false)
	if !strings.Contains(human, "Rate limits: connector=sample declaration=undeclared") {
		t.Fatalf("human ETL output did not honestly report undeclared rate limits:\n%s", human)
	}
	if strings.Contains(human+humanErr, token) {
		t.Fatal("human ETL output leaked a credential value")
	}

	structured, structuredErr := runETLRateLimitCLI(t, root, true)
	if strings.Contains(structured+structuredErr, token) {
		t.Fatal("JSON ETL output leaked a credential value")
	}
	var result struct {
		Kind string `json:"kind"`
		Run  struct {
			RateLimit connectors.RateLimitSummary `json:"rate_limit"`
		} `json:"run"`
	}
	if err := json.Unmarshal([]byte(structured), &result); err != nil {
		t.Fatalf("unmarshal ETL JSON: %v\n%s", err, structured)
	}
	if result.Kind != "ETLRun" || len(result.Run.RateLimit.Connectors) != 2 {
		t.Fatalf("ETL JSON rate-limit summary = %+v", result)
	}
	for _, connector := range result.Run.RateLimit.Connectors {
		if connector.Declaration != connectors.RateLimitDeclarationUndeclared {
			t.Fatalf("connector %q declaration = %q, want undeclared", connector.Connector, connector.Declaration)
		}
		if len(connector.Policies) != 0 {
			t.Fatalf("undeclared connector %q reported policies: %+v", connector.Connector, connector.Policies)
		}
	}
}

func runETLRateLimitCLI(t *testing.T, root string, jsonOut bool) (stdout, stderr string) {
	t.Helper()
	args := []string{"etl", "run", "--connection", "rate_limit_to_warehouse", "--stream", "customers", "--root", root}
	if jsonOut {
		args = append(args, "--json")
	}
	var outBuf, errBuf bytes.Buffer
	if code := cli.Run(args, &outBuf, &errBuf); code != 0 {
		t.Fatalf("Run(%v) code = %d stderr=%s stdout=%s", args, code, errBuf.String(), outBuf.String())
	}
	return outBuf.String(), errBuf.String()
}
