package cli_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"polymetrics.ai/internal/cli"
)

func TestConnectorInspectProjectsDeclaredSyncTransport(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"connectors", "inspect", "github", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(connectors inspect github --json) code = %d stderr = %s", code, stderr.String())
	}

	var response struct {
		SyncTransport struct {
			Source struct {
				Status string `json:"status"`
			} `json:"source"`
			Destination struct {
				Status string `json:"status"`
			} `json:"destination"`
		} `json:"sync_transport"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal inspect JSON: %v\n%s", err, stdout.String())
	}
	if response.SyncTransport.Source.Status != "declared" || response.SyncTransport.Destination.Status != "declared" {
		t.Fatalf("sync transport eligibility = %#v, want declared source and destination from the GitHub definition", response.SyncTransport)
	}
}

func TestConnectorsHelpExplainsDeclaredNoneInspectionPolicy(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"help", "connectors"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(help connectors) code = %d stderr = %s", code, stderr.String())
	}
	for _, want := range []string{
		"acknowledgement=none remains declared",
		"only durable_warehouse can execute",
		"POLLING-WATERMARK ELIGIBILITY",
		"not CDC or change capture",
		"hard deletes after a",
		"cursor-advancing soft-delete",
		"source identity mismatch",
		"explicit rebootstrap",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("connector help missing %q:\n%s", want, stdout.String())
		}
	}
}
