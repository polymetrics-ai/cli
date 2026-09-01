package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestPMBinaryExecutesInstalledApprovedJobFlow is the fresh executable proof
// for the connector-action authorization path:
//
//	cmd/pm -> cli.Run -> flow create -> approved reverse job resolution
//	cmd/pm -> cli.Run -> schedule install -> direct flow run payload
//	cmd/pm -> cli.Run -> flow run -> App.ExecuteAuthorizedFlowAction
//
// The installed firing must finish with a prepared-execution identity while
// every durable and rendered surface remains free of approval carriers.
func TestPMBinaryExecutesInstalledApprovedJobFlow(t *testing.T) {
	binary := buildTransportPM(t)
	sha, size := transportBinaryIdentity(t, binary)
	t.Logf("fresh connector-action pm binary sha256=%s size_bytes=%d", sha, size)
	root := filepath.Join(t.TempDir(), "project")
	sourceWarehouse := filepath.Join(root, ".polymetrics", "warehouse")
	targetWarehouse := filepath.Join(root, "target-warehouse")
	crontab := filepath.Join(root, "crontab")
	t.Setenv("PM_SAMPLE_TOKEN", "synthetic-sample-token")
	t.Setenv("PM_CRONTAB_FILE", crontab)

	mustRunBinary := func(stdin string, args ...string) string {
		t.Helper()
		output, err := runTransportPM(binary, stdin, args...)
		if err != nil {
			t.Fatalf("fresh pm command %s failed: %v", transportCommandName(args), err)
		}
		return output
	}

	mustRunBinary("", "init", "--root", root, "--json")
	mustRunBinary("",
		"credentials", "add", "sample-source", "--connector", "sample",
		"--from-env", "token=PM_SAMPLE_TOKEN", "--root", root, "--json",
	)
	mustRunBinary("",
		"credentials", "add", "source-warehouse", "--connector", "warehouse",
		"--config", "path="+sourceWarehouse, "--root", root, "--json",
	)
	mustRunBinary("",
		"credentials", "add", "target-warehouse", "--connector", "warehouse",
		"--config", "path="+targetWarehouse, "--root", root, "--json",
	)
	mustRunBinary("",
		"connections", "create", "sample-to-source",
		"--source", "sample:sample-source", "--destination", "warehouse:source-warehouse",
		"--stream", "customers", "--primary-key", "id", "--cursor", "updated_at",
		"--table", "sample_customers", "--root", root, "--json",
	)
	mustRunBinary("", "etl", "run", "--connection", "sample-to-source", "--stream", "customers", "--root", root, "--json")

	planOutput := mustRunBinary("",
		"reverse", "plan", "scheduled-customer-copy",
		"--source-table", "sample_customers", "--destination", "warehouse:target-warehouse",
		"--map", "id:id", "--map", "name:name", "--map", "email:email",
		"--root", root,
	)
	planIDMatch := regexp.MustCompile(`Created reverse plan (\S+)`).FindStringSubmatch(planOutput)
	if len(planIDMatch) != 2 {
		t.Fatal("fresh pm reverse plan did not return a plan identity")
	}
	planID := planIDMatch[1]
	approvalToken := binaryApprovalToken(planOutput)
	if approvalToken == "" {
		t.Fatal("fresh pm reverse plan did not issue an approval token")
	}
	if _, err := runTransportPM(binary, "", "reverse", "preview", planID, "--root", root, "--json"); err != nil {
		t.Fatal("fresh pm reverse preview failed")
	}
	if _, err := runTransportPM(binary, approvalToken+"\n",
		"reverse", "run", planID, "--approval-token-stdin", "--confirm", "destructive", "--root", root, "--json",
	); err != nil {
		t.Fatal("fresh pm reverse run did not establish standing job approval")
	}

	manifestPath := filepath.Join(t.TempDir(), "scheduled-flow.json")
	manifest := fmt.Sprintf(`{
		"version": 1,
		"name": "scheduled-customer-flow",
		"steps": [{
			"id": "copy-approved-customers",
			"kind": "action",
			"job": %q,
			"action_cfg": {"read_back_stream": "scheduled-customer-copy"}
		}]
	}`, planID)
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	mustRunBinary("", "flow", "create", "--file", manifestPath, "--root", root, "--json")
	mustRunBinary("", "schedule", "create", "--name", "nightly-customers", "--cron", "0 2 * * *", "--flow", "scheduled-customer-flow", "--root", root, "--json")
	mustRunBinary("", "schedule", "install", "nightly-customers", "--crontab", "--root", root, "--json")

	installed, err := os.ReadFile(crontab)
	if err != nil {
		t.Fatal(err)
	}
	wantPayload := binary + " --root " + root + " flow run scheduled-customer-flow --json"
	if !strings.Contains(string(installed), wantPayload) {
		t.Fatalf("installed crontab does not contain the direct flow payload: %q", installed)
	}
	for _, forbidden := range []string{"--authorization", "approval_token", "authorization_reference", approvalToken} {
		if strings.Contains(string(installed), forbidden) {
			t.Fatal("installed crontab retained forbidden approval material")
		}
	}

	// Execute the exact argv rendered into crontab; this is an installed firing,
	// not a fixture-only backend render assertion.
	firingOutput := mustRunBinary("", "--root", root, "flow", "run", "scheduled-customer-flow", "--json")
	var result struct {
		FlowName string `json:"flow_name"`
		Status   string `json:"status"`
		Steps    []struct {
			Status                    string `json:"status"`
			PreparedExecutionIdentity string `json:"prepared_execution_identity"`
		} `json:"steps"`
	}
	if err := json.Unmarshal([]byte(firingOutput), &result); err != nil {
		t.Fatalf("decode installed firing result: %v", err)
	}
	if result.FlowName != "scheduled-customer-flow" || result.Status != "ok" || len(result.Steps) != 1 || result.Steps[0].Status != "ok" || !regexp.MustCompile(`^pex_[0-9a-f]{64}$`).MatchString(result.Steps[0].PreparedExecutionIdentity) {
		t.Fatalf("installed firing did not return terminal prepared identity: %+v", result)
	}

	statusOutput := mustRunBinary("", "schedule", "inspect", "nightly-customers", "--root", root, "--json")
	var status struct {
		Status struct {
			Status   string `json:"status"`
			LastFire struct {
				FlowStatus                  string   `json:"flow_status"`
				PreparedExecutionIdentities []string `json:"prepared_execution_identities"`
			} `json:"last_fire"`
		} `json:"status"`
	}
	if err := json.Unmarshal([]byte(statusOutput), &status); err != nil {
		t.Fatal(err)
	}
	if status.Status.Status != "succeeded" || status.Status.LastFire.FlowStatus != "ok" || len(status.Status.LastFire.PreparedExecutionIdentities) != 1 || status.Status.LastFire.PreparedExecutionIdentities[0] != result.Steps[0].PreparedExecutionIdentity {
		t.Fatalf("installed firing terminal schedule state = %+v", status)
	}

	assertNoApprovalCarrierInTree(t, root, approvalToken)
}

// TestPMBinaryRefusesRequiredSharedRateBudgetBeforeSend proves the fresh
// executable path returns the stable refusal code before transport dispatch:
//
//	cmd/pm -> cli.Run -> connector command -> engine.Runtime.RequesterFor
//	-> RateBudgetRefusalError(shared_coordinator_unavailable)
func binaryApprovalToken(output string) string {
	match := regexp.MustCompile(`Approval token:\s+(\S+)`).FindStringSubmatch(output)
	if len(match) != 2 {
		return ""
	}
	return match[1]
}

func assertNoApprovalCarrierInTree(t *testing.T, root, approvalToken string) {
	t.Helper()
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, forbidden := range []string{approvalToken, "approval_token", "--authorization"} {
			if strings.Contains(string(contents), forbidden) {
				return fmt.Errorf("%s retained forbidden approval material", filepath.Base(path))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
