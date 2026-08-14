package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestConnectorsInspectLabelsProcessLocalRateLimitProtection(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"connectors", "inspect", "github", "--json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("pm connectors inspect github --json exit = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"rate_limit_coordination":{"mode":"process_local"`) {
		t.Fatal("inspect output did not label process-local rate-limit protection")
	}
	if strings.Contains(stdout.String(), "opaque-projection") || strings.Contains(stdout.String(), "binding") {
		t.Fatal("inspect output exposed protected coordination material")
	}
}
