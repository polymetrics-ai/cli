// Package twenty keeps the Twenty CRM bundle's declared provider inventory
// aligned with the command runner's executable surface.
package twenty

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
)

const twentyBundleName = "twenty"

// TestAllProviderOperationsHaveExecutableCommands is deliberately written
// before importing the recovered bundle. It pins the #277 REST inventory and
// proves every published command reaches the real runtime preflight rather
// than merely being listed in JSON.
func TestAllProviderOperationsHaveExecutableCommands(t *testing.T) {
	bundle, err := engine.Load(os.DirFS(".."), twentyBundleName)
	if err != nil {
		t.Fatalf("load %s bundle: %v", twentyBundleName, err)
	}
	if bundle.Surface == nil {
		t.Fatal("api_surface.json did not load")
	}
	if got := len(bundle.Surface.Endpoints); got != 168 {
		t.Fatalf("Twenty API-surface rows = %d, want 168", got)
	}

	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	surface := connector.CommandSurface()
	if surface == nil {
		t.Fatal("Twenty has no cli_surface.json")
	}
	if got := len(surface.Commands); got != 168 {
		t.Fatalf("Twenty CLI commands = %d, want 168", got)
	}

	wantIntents := map[string]int{
		"etl":         28,
		"direct_read": 28,
		"reverse_etl": 112,
	}
	gotIntents := make(map[string]int, len(wantIntents))
	for _, command := range surface.Commands {
		gotIntents[command.Intent]++
		if command.Availability != "implemented" {
			t.Errorf("%q availability = %q, want implemented", command.Path, command.Availability)
		}
		if err := commandrunner.Preflight(connector, strings.Fields(command.Path)); err != nil {
			t.Errorf("Preflight(%q) = %v, want executable command", command.Path, err)
		}
	}
	for intent, want := range wantIntents {
		if got := gotIntents[intent]; got != want {
			t.Errorf("%s commands = %d, want %d", intent, got, want)
		}
	}
}

func TestOperationReadAndStructuredWriteSafetyContracts(t *testing.T) {
	bundle, err := engine.Load(os.DirFS(".."), twentyBundleName)
	if err != nil {
		t.Fatalf("load %s bundle: %v", twentyBundleName, err)
	}
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))

	t.Run("happy direct read has an exact provider route", func(t *testing.T) {
		if err := connector.PreflightOperationDirectRead("twenty.calendar-events.get", "GET", "/rest/calendarEvents/{id}", 1<<20, "json_redacted"); err != nil {
			t.Fatalf("PreflightOperationDirectRead(calendar events): %v", err)
		}
	})

	t.Run("bad direct read rejects a stale hyphenated provider route", func(t *testing.T) {
		err := connector.PreflightOperationDirectRead("twenty.calendar-events.get", "GET", "/rest/calendar-events/{id}", 1<<20, "json_redacted")
		if err == nil || !strings.Contains(err.Error(), "does not match declared operation path") {
			t.Fatalf("PreflightOperationDirectRead(stale path) = %v, want declared-path rejection", err)
		}
	})

	t.Run("edge batch JSON is limited to the documented sixty records", func(t *testing.T) {
		if err := connector.PreflightStructuredJSONRecordField("batch_people", "records"); err != nil {
			t.Fatalf("PreflightStructuredJSONRecordField(records): %v", err)
		}
		if err := connector.PreflightStructuredJSONRecordField("batch_people", "raw_body"); err == nil {
			t.Fatal("PreflightStructuredJSONRecordField(raw_body) unexpectedly accepted a generic body escape hatch")
		}

		var action *engine.WriteAction
		for i := range bundle.Writes {
			if bundle.Writes[i].Name == "batch_people" {
				action = &bundle.Writes[i]
				break
			}
		}
		if action == nil {
			t.Fatal("batch_people write action is missing")
		}
		var schema struct {
			Properties map[string]struct {
				MaxItems int `json:"maxItems"`
			} `json:"properties"`
		}
		if err := json.Unmarshal(action.RecordSchema, &schema); err != nil {
			t.Fatalf("decode batch_people record schema: %v", err)
		}
		if got := schema.Properties["records"].MaxItems; got != 60 {
			t.Fatalf("batch_people records.maxItems = %d, want 60", got)
		}
	})

	for _, action := range bundle.Writes {
		if action.Kind != "delete" {
			continue
		}
		if action.Confirmation == nil || action.Confirmation.Kind != connectors.ConfirmationKindDestructive {
			t.Errorf("delete action %q confirmation = %#v, want typed destructive confirmation", action.Name, action.Confirmation)
		}
	}
}
