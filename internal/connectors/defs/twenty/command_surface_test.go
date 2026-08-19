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
	for _, topic := range surface.HelpTopics {
		if topic.Name == "direct-read" && strings.Contains(strings.ToLower(topic.Summary), "planned") {
			t.Fatalf("direct-read help topic = %q, want current executable-contract wording", topic.Summary)
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

// TestTypedDestinationDeclaration pins the one reversible CRM projection that
// the closed declarative typed-destination adapter can execute. All Twenty
// operations remain CLI-reachable; deletes deliberately stay outside this
// no-tombstone transport and retain their typed destructive CLI contract.
func TestTypedDestinationDeclaration(t *testing.T) {
	bundle, err := engine.Load(os.DirFS(".."), twentyBundleName)
	if err != nil {
		t.Fatalf("load %s bundle: %v", twentyBundleName, err)
	}
	if bundle.SyncTransport == nil || bundle.SyncTransport.Source == nil || bundle.SyncTransport.Destination == nil {
		t.Fatalf("SyncTransport = %#v, want declared source and typed destination", bundle.SyncTransport)
	}

	source := bundle.SyncTransport.Source
	if source.Executor.Family != "declarative_api" || source.Executor.ID != "declarative_stream_source" {
		t.Fatalf("source executor = %#v, want declarative stream source", source.Executor)
	}
	if got, want := len(source.EligibleStreams), len(bundle.Streams); got != want {
		t.Fatalf("source eligible streams = %d, want every %d Twenty stream", got, want)
	}
	seenStreams := make(map[string]bool, len(source.EligibleStreams))
	for _, stream := range source.EligibleStreams {
		if seenStreams[stream] {
			t.Fatalf("source eligible streams repeats %q", stream)
		}
		seenStreams[stream] = true
	}
	for _, stream := range bundle.Streams {
		if !seenStreams[stream.Name] {
			t.Errorf("source transport omits executable stream %q", stream.Name)
		}
	}

	destination := bundle.SyncTransport.Destination
	if destination.Executor.Family != "declarative_api" || destination.Executor.ID != "declarative_typed_destination" {
		t.Fatalf("destination executor = %#v, want declarative typed destination", destination.Executor)
	}
	if destination.Acknowledgement != connectors.TransportAcknowledgementDurableWarehouse {
		t.Fatalf("destination acknowledgement = %q, want durable warehouse", destination.Acknowledgement)
	}
	if destination.Delivery.Idempotency != connectors.DeliveryIdempotencyKeyed || destination.Delivery.Deletes != connectors.DeliveryDeletesUnavailable {
		t.Fatalf("destination delivery = %#v, want keyed no-tombstone delivery", destination.Delivery)
	}
	if got := destination.EligibleActions; len(got) != 1 || got[0] != "create_companies" {
		t.Fatalf("destination eligible actions = %#v, want only reversible create_companies", got)
	}
	if got := destination.ApplyStrategies; len(got) != 1 || got[0].Action != "create_companies" || string(got[0].Mode) != "full_append" || string(got[0].Strategy) != "append" {
		t.Fatalf("destination apply strategies = %#v, want full_append create_companies", got)
	}
	if got := destination.SourceBindings; len(got) != 1 || len(got[0].EligibleStreams) != 1 || got[0].EligibleStreams[0] != "companies" {
		t.Fatalf("destination source bindings = %#v, want only companies", got)
	} else if got := got[0].RecordMapping; got.Kind != connectors.SourceRecordMappingKindInputFields || len(got.Inputs) != 1 || got.Inputs[0].Input != "name" || got.Inputs[0].Field != "name" {
		t.Fatalf("destination record mapping = %#v, want closed name-to-name input mapping", got)
	}

	for _, action := range bundle.Writes {
		if action.Name == "delete_companies" && action.Confirmation == nil {
			t.Fatal("delete_companies must retain its destructive CLI confirmation outside the no-tombstone destination")
		}
	}
}
