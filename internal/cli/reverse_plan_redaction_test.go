package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
)

// TestSafeReversePlanForOutputHonoursPlanRedactFields pins the CLI boundary
// that renders a reverse plan for `--json`.
//
// safeReversePlanForOutput redacts the sample using the plan's own
// RedactFields, so a plan constructor that leaves RedactFields empty makes the
// bundle's redact_fields declaration inert here. Both connector-command and
// source-table plans populate it.
func TestSafeReversePlanForOutputHonoursPlanRedactFields(t *testing.T) {
	const sentinel = "SEALED-SENTINEL-VALUE"

	plan := app.ReversePlan{
		ID:                     "rplan_fixture",
		Mode:                   "connector_command",
		Action:                 "set_secret",
		ConnectorCommandRecord: connectors.Record{"encrypted_value": sentinel},
		ConnectorCommandHeaders: map[string]string{
			"X-Request-Mode": sentinel,
		},
		ConnectorCommandHeaderValues: map[string][]string{
			"X-Request-Mode": {sentinel},
		},
		RedactFields: []string{"encrypted_value"},
		Sample: []connectors.Record{{
			"secret_name":     "DEPLOY_KEY",
			"encrypted_value": sentinel,
			"key_id":          "568250167242549743",
		}},
		ApprovalToken: "token-should-not-survive",
	}

	safe := safeReversePlanForOutput(plan)

	if safe.ConnectorCommandRecord != nil {
		t.Fatalf("ConnectorCommandRecord = %#v, want nil at the output boundary", safe.ConnectorCommandRecord)
	}
	if safe.ConnectorCommandHeaders != nil || safe.ConnectorCommandHeaderValues != nil {
		t.Fatalf("connector command headers = %#v / %#v, want nil at the output boundary", safe.ConnectorCommandHeaders, safe.ConnectorCommandHeaderValues)
	}
	if safe.ApprovalToken != "" {
		t.Fatalf("ApprovalToken = %q, want cleared at the output boundary", safe.ApprovalToken)
	}
	if got := safe.Sample[0]["encrypted_value"]; got != "redacted" {
		t.Fatalf("sample encrypted_value = %#v, want %q", got, "redacted")
	}
	if got := safe.Sample[0]["secret_name"]; got != "DEPLOY_KEY" {
		t.Fatalf("sample secret_name = %#v, want preserved", got)
	}

	emitted, err := json.Marshal(safe)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(emitted), sentinel) {
		t.Fatalf("emitted JSON contains the sealed sentinel: %s", emitted)
	}

	// Negative control: a plan declaring no redact fields must round-trip
	// its record fields unchanged, so this cannot pass by blanket redaction.
	plain := app.ReversePlan{
		Sample: []connectors.Record{{"id": "t3_abc", "dir": "1"}},
	}
	safePlain := safeReversePlanForOutput(plain)
	if got := safePlain.Sample[0]["id"]; got != "t3_abc" {
		t.Fatalf("sample id = %#v, want t3_abc unredacted when no redact fields are declared", got)
	}
}
