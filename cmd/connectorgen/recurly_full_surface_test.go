package main

import (
	"encoding/json"
	"os"
	"testing"
)

// TestRecurlyFullSurfaceCommandAndOperationCoverage locks the complete
// 197-command CLI surface invariant: every one of the 93 ETL streams and 96
// typed reverse-ETL write actions has its own `pm recurly <command>` entry,
// plus the 5 implemented direct-read previews and the 3 bounded
// binary/export operations. This prevents the connector from regressing to an
// ETL-only surface with no per-operation command entries.
func TestRecurlyFullSurfaceCommandAndOperationCoverage(t *testing.T) {
	api := loadRecurlyJSON[struct {
		Endpoints []struct {
			Method    string         `json:"method"`
			Path      string         `json:"path"`
			CoveredBy map[string]any `json:"covered_by"`
			Operation map[string]any `json:"operation"`
		} `json:"endpoints"`
	}](t, "../../internal/connectors/defs/recurly/api_surface.json")

	cli := loadRecurlyJSON[struct {
		Commands []struct {
			Path         string `json:"path"`
			Intent       string `json:"intent"`
			Availability string `json:"availability"`
			Stream       string `json:"stream"`
			Write        string `json:"write"`
			Operation    string `json:"operation"`
			OutputPolicy string `json:"output_policy"`
			Risk         string `json:"risk"`
			Approval     string `json:"approval"`
		} `json:"commands"`
	}](t, "../../internal/connectors/defs/recurly/cli_surface.json")

	streams := loadRecurlyJSON[struct {
		Streams []struct {
			Name string `json:"name"`
		} `json:"streams"`
	}](t, "../../internal/connectors/defs/recurly/streams.json")

	writes := loadRecurlyJSON[struct {
		Actions []struct {
			Name string `json:"name"`
		} `json:"actions"`
	}](t, "../../internal/connectors/defs/recurly/writes.json")

	if got := len(streams.Streams); got != 93 {
		t.Fatalf("streams = %d, want 93", got)
	}
	if got := len(writes.Actions); got != 96 {
		t.Fatalf("write actions = %d, want 96", got)
	}

	coverage := map[string]int{}
	for _, ep := range api.Endpoints {
		if ep.CoveredBy != nil {
			for key := range ep.CoveredBy {
				coverage[key]++
			}
		}
		if ep.Operation != nil {
			coverage["operation"]++
		}
	}
	if got, want := coverage["stream"], 93; got != want {
		t.Fatalf("api_surface coverage[stream] = %d, want %d", got, want)
	}
	if got, want := coverage["write"], 96; got != want {
		t.Fatalf("api_surface coverage[write] = %d, want %d", got, want)
	}
	if got, want := coverage["direct_read"], 5; got != want {
		t.Fatalf("api_surface coverage[direct_read] = %d, want %d", got, want)
	}
	if got, want := coverage["operation"], 3; got != want {
		t.Fatalf("api_surface coverage[operation] = %d, want %d", got, want)
	}

	if got, want := len(cli.Commands), 197; got != want {
		t.Fatalf("cli_surface commands = %d, want 197", got)
	}

	// Every stream and every write must have its own executable command.
	streamSet := map[string]bool{}
	for _, s := range streams.Streams {
		streamSet[s.Name] = false
	}
	writeSet := map[string]bool{}
	for _, w := range writes.Actions {
		writeSet[w.Name] = false
	}

	intent := map[string]int{}
	availability := map[string]int{}
	for _, cmd := range cli.Commands {
		intent[cmd.Intent]++
		availability[cmd.Availability]++
		switch cmd.Intent {
		case "etl":
			if cmd.Stream == "" {
				t.Fatalf("etl command %q has no stream", cmd.Path)
			}
			if _, ok := streamSet[cmd.Stream]; !ok {
				t.Fatalf("etl command %q references unknown stream %q", cmd.Path, cmd.Stream)
			}
			streamSet[cmd.Stream] = true
		case "reverse_etl":
			if cmd.Write == "" {
				t.Fatalf("reverse_etl command %q has no write action", cmd.Path)
			}
			if _, ok := writeSet[cmd.Write]; !ok {
				t.Fatalf("reverse_etl command %q references unknown write %q", cmd.Path, cmd.Write)
			}
			writeSet[cmd.Write] = true
			// Reverse-ETL commands must carry a truthful disposition with
			// risk and approval text for the plan/preview/approve/execute
			// lifecycle.
			if cmd.Risk == "" || cmd.Approval == "" {
				t.Fatalf("reverse_etl command %q must declare risk and approval", cmd.Path)
			}
		case "direct_read":
			if cmd.Availability == "implemented" && cmd.OutputPolicy == "" {
				t.Fatalf("implemented direct_read command %q must declare output_policy", cmd.Path)
			}
		}
	}

	for name, seen := range streamSet {
		if !seen {
			t.Fatalf("stream %q has no cli_surface command", name)
		}
	}
	for name, seen := range writeSet {
		if !seen {
			t.Fatalf("write action %q has no cli_surface command", name)
		}
	}

	if got, want := intent["etl"], 93; got != want {
		t.Fatalf("etl commands = %d, want 93", got)
	}
	if got, want := intent["reverse_etl"], 96; got != want {
		t.Fatalf("reverse_etl commands = %d, want 96", got)
	}
	if got, want := intent["direct_read"], 8; got != want {
		t.Fatalf("direct_read commands = %d, want 8", got)
	}
	if got, want := availability["implemented"], 194; got != want {
		t.Fatalf("implemented commands = %d, want 194", got)
	}
	// The 3 binary/export endpoints are recorded as bounded planned metadata.
	if got, want := availability["planned"], 3; got != want {
		t.Fatalf("planned commands = %d, want 3", got)
	}
}

func loadRecurlyJSON[T any](t *testing.T, path string) T {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
	return out
}

// TestRecurlyMetadataEnablesWriteCapability mirrors the Gong expectation: a
// connector shipping 96 typed reverse-ETL write actions must advertise the
// write capability so the command surface is not ETL-only.
func TestRecurlyMetadataEnablesWriteCapability(t *testing.T) {
	metadata := loadRecurlyJSON[struct {
		Capabilities struct {
			Read  bool `json:"read"`
			Write bool `json:"write"`
		} `json:"capabilities"`
	}](t, "../../internal/connectors/defs/recurly/metadata.json")
	if !metadata.Capabilities.Read || !metadata.Capabilities.Write {
		t.Fatalf("Recurly capabilities read/write = %t/%t, want true/true", metadata.Capabilities.Read, metadata.Capabilities.Write)
	}
}
