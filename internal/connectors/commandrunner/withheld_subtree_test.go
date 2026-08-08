package commandrunner

import (
	"encoding/json"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func subtreeJSON(t *testing.T, record connectors.Record, path ...string) string {
	t.Helper()
	var current any = map[string]any(record)
	for _, part := range path {
		object, ok := current.(map[string]any)
		if !ok {
			t.Fatalf("record %#v has no object at %q", record, part)
		}
		current, ok = object[part]
		if !ok {
			t.Fatalf("record %#v has no value at %q", record, part)
		}
	}
	raw, err := json.Marshal(current)
	if err != nil {
		t.Fatalf("marshal subtree: %v", err)
	}
	return string(raw)
}

// TestReconstituteWithheldFieldsRebuildsSubtreeFromDescendantFlags covers the
// shape where a declared sensitive field is an ancestor of several flag targets
// rather than a flag target itself. Withholding removes the whole subtree, which
// is the correct reading of the declaration, so reconstitution has to rebuild it
// from the flags beneath it or the plan can never be previewed or run.
func TestReconstituteWithheldFieldsRebuildsSubtreeFromDescendantFlags(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{
		Commands: []connectors.CommandSurfaceCommand{
			{
				Path:         "invoices retries create",
				Intent:       "reverse_etl",
				Availability: "implemented",
				Write:        "create_invoice_retry",
				Flags: []connectors.CommandSurfaceFlag{
					{Name: "account-code", Type: "string", MapsTo: "record.account.code", Required: true},
					{Name: "account-billing-infos-0-gateway-code", Type: "string", MapsTo: "record.account.billing_infos.0.gateway_code", Required: true},
					{Name: "account-billing-infos-0-transactions-0-gateway-error-code", Type: "string", MapsTo: "record.account.billing_infos.0.transactions.0.gateway_error_code", Required: true},
					{Name: "currency", Type: "string", MapsTo: "record.currency", Required: true},
				},
			},
		},
	}}
	path := []string{"invoices", "retries", "create"}
	flags := map[string][]string{
		"account-code":                         {"acct_fixture"},
		"account-billing-infos-0-gateway-code": {"stripe"},
		"account-billing-infos-0-transactions-0-gateway-error-code": {"card_declined"},
		"currency": {"USD"},
	}

	planned, err := recordOverrides(connector.surface.Commands[0], flags)
	if err != nil {
		t.Fatalf("recordOverrides: %v", err)
	}

	record, missing, err := ReconstituteWithheldFields(connector, path, []string{"account.billing_infos"}, flags)
	if err != nil {
		t.Fatalf("ReconstituteWithheldFields: %v", err)
	}
	if len(missing) != 0 {
		t.Fatalf("missing = %#v, want the subtree rebuilt from its descendant flags", missing)
	}
	want := subtreeJSON(t, planned, "account", "billing_infos")
	got := subtreeJSON(t, record, "account", "billing_infos")
	if got != want {
		t.Fatalf("rebuilt subtree = %s, want the planned subtree %s", got, want)
	}
	if _, present := record["currency"]; present {
		t.Fatalf("record = %#v, want only the withheld subtree rebuilt", record)
	}

	// Negative control: with the descendant flags absent the caller is told
	// which flags supply the subtree, never the bare ancestor path no flag can
	// re-supply.
	_, missing, err = ReconstituteWithheldFields(connector, path, []string{"account.billing_infos"}, map[string][]string{"account-code": {"acct_fixture"}})
	if err != nil {
		t.Fatalf("ReconstituteWithheldFields with no descendant flags: %v", err)
	}
	if len(missing) != 2 {
		t.Fatalf("missing = %#v, want both descendant flags named", missing)
	}
	for _, name := range missing {
		if name == "account.billing_infos" {
			t.Fatalf("missing = %#v, names a path no flag can supply", missing)
		}
	}
}

// TestReconstituteWithheldFieldsSkipsUnsuppliedOptionalDescendant pins the
// asymmetry that keeps the ancestor fix from reintroducing the unsatisfiable
// precondition it replaces: recordOverrides skips an optional flag with no
// value, so reconstitution must skip it too rather than demand it back.
func TestReconstituteWithheldFieldsSkipsUnsuppliedOptionalDescendant(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{
		Commands: []connectors.CommandSurfaceCommand{
			{
				Path:         "hook add",
				Intent:       "reverse_etl",
				Availability: "implemented",
				Write:        "add_hook",
				Flags: []connectors.CommandSurfaceFlag{
					{Name: "config-url", Type: "string", MapsTo: "record.config.url", Required: true},
					{Name: "config-secret", Type: "string", MapsTo: "record.config.secret"},
				},
			},
		},
	}}
	flags := map[string][]string{"config-url": {"https://hooks.example.test/x"}}

	planned, err := recordOverrides(connector.surface.Commands[0], flags)
	if err != nil {
		t.Fatalf("recordOverrides: %v", err)
	}
	record, missing, err := ReconstituteWithheldFields(connector, []string{"hook", "add"}, []string{"config"}, flags)
	if err != nil {
		t.Fatalf("ReconstituteWithheldFields: %v", err)
	}
	if len(missing) != 0 {
		t.Fatalf("missing = %#v, want an unsupplied optional descendant skipped", missing)
	}
	if got, want := subtreeJSON(t, record, "config"), subtreeJSON(t, planned, "config"); got != want {
		t.Fatalf("rebuilt subtree = %s, want the planned subtree %s", got, want)
	}
}

// TestReconstituteWithheldFieldsRebuildsRecurlyBillingInfos is the same
// assertion against the shipped bundle that surfaced the shape, so the fix
// cannot regress behind a fixture that drifts from it.
func TestReconstituteWithheldFieldsRebuildsRecurlyBillingInfos(t *testing.T) {
	bundle, err := engine.Load(defs.FS, "recurly")
	if err != nil {
		t.Fatalf("load recurly bundle: %v", err)
	}
	connector := engine.New(bundle, nil)
	path := []string{"invoices", "retries", "create"}
	cmd, _, err := resolvePreflightCommand(connector, path)
	if err != nil {
		t.Fatalf("resolvePreflightCommand: %v", err)
	}

	flags := map[string][]string{}
	for _, flag := range cmd.Flags {
		switch {
		case flag.Type == "boolean":
			flags[flag.Name] = []string{"true"}
		case flag.Type == "number" || flag.Type == "integer":
			flags[flag.Name] = []string{"1"}
		case flag.Type == "enum":
			if len(flag.Values) == 0 {
				t.Fatalf("enum flag --%s declares no values", flag.Name)
			}
			flags[flag.Name] = []string{flag.Values[0]}
		case flag.Format == "date-time":
			flags[flag.Name] = []string{"2026-08-08T00:00:00Z"}
		case flag.Format == "date":
			flags[flag.Name] = []string{"2026-08-08"}
		default:
			flags[flag.Name] = []string{"fixture-" + flag.Name}
		}
	}

	planned, err := recordOverrides(cmd, flags)
	if err != nil {
		t.Fatalf("recordOverrides: %v", err)
	}
	record, missing, err := ReconstituteWithheldFields(connector, path, []string{"account.billing_infos"}, flags)
	if err != nil {
		t.Fatalf("ReconstituteWithheldFields: %v", err)
	}
	if len(missing) != 0 {
		t.Fatalf("missing = %#v, want recurly's declared account.billing_infos rebuilt", missing)
	}
	if got, want := subtreeJSON(t, record, "account", "billing_infos"), subtreeJSON(t, planned, "account", "billing_infos"); got != want {
		t.Fatalf("rebuilt subtree = %s, want the planned subtree %s", got, want)
	}
}
