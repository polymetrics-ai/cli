package conformance

import (
	"encoding/base64"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func TestGitHubExhaustiveProviderDouble(t *testing.T) {
	report, err := runGitHubExhaustiveProviderDouble(t)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("load GitHub bundle: %v", err)
	}
	if report.Streams != len(bundle.Streams) || report.WriteActions != len(bundle.Writes) || report.Operations != len(bundle.Operations) {
		t.Fatalf("provider-double totals = streams=%d writes=%d operations=%d, want source bundle %d/%d/%d",
			report.Streams, report.WriteActions, report.Operations, len(bundle.Streams), len(bundle.Writes), len(bundle.Operations))
	}
	if report.GenericStreams != 23 || report.GenericWrites != 41 {
		t.Fatalf("generic routes = streams=%d writes=%d, want 23/41", report.GenericStreams, report.GenericWrites)
	}
	if report.Failed != 0 {
		t.Fatalf("provider-double report has %d failed rows: %v", report.Failed, report.Failures)
	}
	for _, row := range report.Rows {
		if !strings.HasPrefix(row.Name, "github.graphql.") {
			continue
		}
		if row.State != "exercised" {
			t.Errorf("fixed GraphQL operation %q state = %q, want exercised (%s)", row.Name, row.State, row.Reason)
		}
	}
}

func TestSyntheticGitHubSecretSetRecordUsesSchemaValidCiphertext(t *testing.T) {
	bundle, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("load GitHub bundle: %v", err)
	}
	var action *engine.WriteAction
	for index := range bundle.Writes {
		if bundle.Writes[index].Name == "actions_secrets_secret_name3" {
			action = &bundle.Writes[index]
			break
		}
	}
	if action == nil {
		t.Fatal("GitHub actions secret-set write action is absent")
	}
	record := syntheticGitHubRecord(*action)
	value, ok := record["encrypted_value"].(string)
	if !ok || value == "" {
		t.Fatalf("synthetic encrypted_value = %#v, want base64 ciphertext", record["encrypted_value"])
	}
	if value == "provider-double" {
		t.Fatal("synthetic secret-set value is plaintext rather than ciphertext-shaped input")
	}
	if _, err := base64.StdEncoding.DecodeString(value); err != nil {
		t.Fatalf("synthetic encrypted_value is not base64: %v", err)
	}
}
