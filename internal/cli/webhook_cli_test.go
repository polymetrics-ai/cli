package cli_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"polymetrics.ai/internal/cli"
)

func TestWebhookCommandsDeclareModesWithoutLeakingCallbacks(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	if code := cli.Run([]string{"init", "--root", root, "--json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("init exit=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}

	tests := []struct {
		name string
		args []string
		mode string
	}{
		{
			name: "operator endpoint",
			args: []string{"webhooks", "configure", "operator", "--mode", "operator_endpoint", "--callback-url", "https://operator.example.test/receiver", "--receipt-capacity", "2", "--root", root, "--json"},
			mode: "operator_endpoint",
		},
		{
			name: "external tailscale funnel",
			args: []string{"webhooks", "configure", "funnel", "--mode", "external_tunnel", "--tunnel-tool", "tailscale_funnel", "--callback-url", "https://node.tailnet.ts.net/receiver", "--heartbeat-ttl", "1m", "--receipt-capacity", "2", "--root", root, "--json"},
			mode: "external_tunnel",
		},
		{
			name: "provider pull or stream",
			args: []string{"webhooks", "configure", "pull", "--mode", "provider_pull_or_stream", "--adapter", "provider-event-stream-v1", "--receipt-capacity", "2", "--root", root, "--json"},
			mode: "provider_pull_or_stream",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout.Reset()
			stderr.Reset()
			if code := cli.Run(tt.args, &stdout, &stderr); code != 0 {
				t.Fatalf("exit=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
			}
			if strings.Contains(stdout.String(), "operator.example.test") || strings.Contains(stdout.String(), "node.tailnet.ts.net") {
				t.Fatalf("status leaked callback URL: %s", stdout.String())
			}
			var response map[string]any
			if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if tt.mode == "provider_pull_or_stream" {
				if response["kind"] != "IngressAdapter" || response["receiver"] != nil || strings.Contains(stdout.String(), "WebhookReceiver") {
					t.Fatalf("pull/stream response is webhook-framed: %#v", response)
				}
				adapter, ok := response["ingress_adapter"].(map[string]any)
				if !ok || adapter["mode"] != tt.mode || adapter["adapter_reference"] != "provider-event-stream-v1" {
					t.Fatalf("ingress adapter response = %#v", response)
				}
				return
			}
			status, ok := response["receiver"].(map[string]any)
			if !ok || status["mode"] != tt.mode {
				t.Fatalf("receiver response = %#v", response)
			}
		})
	}

	stdout.Reset()
	stderr.Reset()
	if code := cli.Run([]string{"webhooks", "configure", "funnel", "--mode", "external_tunnel", "--tunnel-tool", "tailscale_funnel", "--callback-url", "https://node.tailnet.ts.net/rotated", "--heartbeat-ttl", "1m", "--receipt-capacity", "2", "--root", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("rotation configure exit=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "recovery_outcome=source_generation_changed") || strings.Contains(stdout.String(), "node.tailnet.ts.net") {
		t.Fatalf("rotation status was not safe and typed: %s", stdout.String())
	}

	for _, args := range [][]string{{"webhooks", "--root", root}, {"help", "webhooks", "--root", root}, {"webhooks", "--help", "--root", root}} {
		stdout.Reset()
		stderr.Reset()
		if code := cli.Run(args, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "pm webhooks") {
			t.Fatalf("help args=%v exit=%d stdout=%s stderr=%s", args, code, stdout.String(), stderr.String())
		}
	}
}
