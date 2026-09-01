package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestConnectorsInspectLabelsProcessLocalRateLimitProtection(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"connectors", "inspect", "github", "--json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("pm connectors inspect github --json exit = %d, stderr = %s", code, stderr.String())
	}
	var response struct {
		RateLimitCoordination struct {
			Mode    string `json:"mode"`
			Message string `json:"message"`
		} `json:"rate_limit_coordination"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("decode inspect JSON: %v", err)
	}
	if response.RateLimitCoordination.Mode != "process_local" {
		t.Fatal("inspect output did not label process-local rate-limit protection")
	}
	if strings.Contains(stdout.String(), "opaque-projection") || strings.Contains(stdout.String(), "binding") {
		t.Fatal("inspect output exposed protected coordination material")
	}

	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"connectors", "inspect", "github"}, &stdout, &stderr); code != 0 {
		t.Fatalf("pm connectors inspect github exit = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "RATE LIMIT COORDINATION\n  Process-local rate-limit protection coordinates this pm process only; it is not shared across processes.") {
		t.Fatal("human inspection did not state the process-local rate-limit boundary")
	}
}
