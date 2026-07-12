package main

import (
	"os"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestMondayFullSurfaceAllOpsCovered(t *testing.T) {
	findings, _ := validateBundleDir(os.DirFS("../../internal/connectors/defs"), "monday")
	if len(findings) != 0 {
		t.Fatalf("monday validate findings = %+v, want none", findings)
	}

	b, err := engine.Load(os.DirFS("../../internal/connectors/defs"), "monday")
	if err != nil {
		t.Fatalf("Load(monday): %v", err)
	}
	if b.Surface == nil {
		t.Fatal("monday api_surface.json is not loaded")
	}

	methods := map[string]int{}
	coveredStreams := map[string]bool{}
	coveredDirectReads := map[string]bool{}
	coveredWrites := map[string]bool{}
	for i, ep := range b.Surface.Endpoints {
		methods[strings.ToUpper(ep.Method)]++
		if ep.Operation != nil || ep.Excluded != nil {
			t.Fatalf("endpoint %d (%s %s) is not fully modeled: operation=%+v excluded=%+v", i, ep.Method, ep.Path, ep.Operation, ep.Excluded)
		}
		if ep.CoveredBy == nil {
			t.Fatalf("endpoint %d (%s %s) has no covered_by mapping", i, ep.Method, ep.Path)
		}
		if ep.CoveredBy.Stream != "" {
			coveredStreams[ep.CoveredBy.Stream] = true
		}
		if ep.CoveredBy.DirectRead != "" {
			coveredDirectReads[ep.CoveredBy.DirectRead] = true
		}
		for _, direct := range ep.CoveredBy.DirectReads {
			coveredDirectReads[direct] = true
		}
		if ep.CoveredBy.Write != "" {
			coveredWrites[ep.CoveredBy.Write] = true
		}
	}
	if len(b.Surface.Endpoints) != 367 || methods["GET"] != 87 || methods["POST"] != 280 {
		t.Fatalf("surface methods = %+v total %d, want 87 GET + 280 POST", methods, len(b.Surface.Endpoints))
	}
	if len(coveredStreams) != 5 {
		t.Fatalf("covered streams = %+v, want existing 5 stream implementations", coveredStreams)
	}
	if len(coveredDirectReads) != 82 {
		t.Fatalf("covered direct reads = %d, want 82 non-stream query operations", len(coveredDirectReads))
	}
	if len(coveredWrites) != 280 {
		t.Fatalf("covered writes = %d, want 280 mutation write actions", len(coveredWrites))
	}
	if len(b.Writes) != 280 {
		t.Fatalf("writes.json actions = %d, want 280", len(b.Writes))
	}
	if !b.Metadata.Capabilities.Write {
		t.Fatal("metadata capabilities.write = false, want true for modeled reverse ETL actions")
	}

	implementedDirect := map[string]bool{}
	implementedWrites := map[string]bool{}
	for _, cmd := range b.CLISurface.Commands {
		if cmd.Intent == "raw_api" || cmd.Intent == "direct_write" {
			t.Fatalf("command %q exposes forbidden intent %q", cmd.Path, cmd.Intent)
		}
		if cmd.Intent == "direct_read" && cmd.Availability == "implemented" {
			implementedDirect[cmd.Path] = true
		}
		if cmd.Intent == "reverse_etl" && cmd.Write != "" {
			implementedWrites[cmd.Write] = true
			if !strings.Contains(cmd.Approval, "plan, preview, approval, execute") {
				t.Fatalf("reverse ETL command %q approval = %q, want approval flow", cmd.Path, cmd.Approval)
			}
		}
	}
	for direct := range coveredDirectReads {
		if !implementedDirect[direct] {
			t.Fatalf("direct read %q is covered but has no implemented command", direct)
		}
	}
	for write := range coveredWrites {
		if !implementedWrites[write] {
			t.Fatalf("write %q is covered but has no reverse_etl command metadata", write)
		}
	}
}
