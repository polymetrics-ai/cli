package engine

import (
	"strings"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors/defs"
)

func TestOperationEndpointLedgerRuntimeProjectionFailsClosed(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	delete(fsys, "acme/api_surface.json")
	makeFile := func(raw string) *fstest.MapFile {
		return &fstest.MapFile{Data: []byte(raw)}
	}
	fsys["acme/operations.json"] = makeFile(`{
  "operations": [
    {
      "id": "acme.widgets.get",
      "kind": "rest_read",
      "summary": "Read one widget",
      "risk": "low",
      "approval": "none",
      "output_policy": "json_redacted",
      "rest": {
        "method": "GET",
        "path": "/widgets/{id}",
        "max_bytes": 1024
      }
    }
  ]
}`)

	preflight := func(b Bundle) error {
		return PreflightOperationDirectRead(b, "acme.widgets.get", "GET", "/widgets/{id}", maxOperationDirectReadBytes, "json_redacted")
	}

	bundle, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load without runtime endpoint ledger: %v", err)
	}
	if err := preflight(bundle); err == nil || !strings.Contains(err.Error(), "runtime operation endpoint ledger is unavailable") {
		t.Fatalf("preflight without runtime endpoint ledger = %v, want unavailable rejection", err)
	}

	fsys[RuntimeOperationEndpointLedgerFile] = makeFile(`{
  "other": []
}`)
	bundle, err = Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load with unresolved runtime endpoint ledger: %v", err)
	}
	if err := preflight(bundle); err == nil || !strings.Contains(err.Error(), "runtime operation endpoint ledger is unavailable") {
		t.Fatalf("preflight with unresolved runtime endpoint ledger = %v, want unavailable rejection", err)
	}

	fsys[RuntimeOperationEndpointLedgerFile] = makeFile(`{
  "acme": []
}`)
	bundle, err = Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load with incomplete runtime endpoint ledger: %v", err)
	}
	if err := preflight(bundle); err == nil || !strings.Contains(err.Error(), "runtime operation endpoint ledger does not contain") {
		t.Fatalf("preflight with incomplete runtime endpoint ledger = %v, want endpoint rejection", err)
	}

	fsys[RuntimeOperationEndpointLedgerFile] = makeFile(`{
  "acme": [
    {
      "method": "GET",
      "path": "/widgets/{id}",
      "kind": "rest_read",
      "max_bytes": 1024
    }
  ]
}`)
	bundle, err = Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load with matching runtime endpoint ledger: %v", err)
	}
	if err := preflight(bundle); err != nil {
		t.Fatalf("preflight with matching runtime endpoint ledger: %v", err)
	}

	fsys[RuntimeOperationEndpointLedgerFile] = makeFile(`{
  "acme": [{"unexpected": true}]
}`)
	if _, err := Load(fsys, "acme"); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("Load with malformed runtime endpoint ledger = %v, want strict decoder rejection", err)
	}
}

func TestOperationEndpointLedgerSingleBundleLoadValidatesWholeLedger(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys[RuntimeOperationEndpointLedgerFile] = &fstest.MapFile{Data: []byte(`{
  "acme": [],
  "other": [{"unexpected": true}]
}`)}

	if _, err := Load(fsys, "acme"); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("Load with malformed runtime endpoint ledger entry = %v, want strict decoder rejection", err)
	}
	if _, err := LoadAll(fsys); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("LoadAll with malformed runtime endpoint ledger entry = %v, want strict decoder rejection", err)
	}
}

func TestShippedOperationEndpointLedgerRejectsMissingProjection(t *testing.T) {
	bundles, err := LoadAll(defs.FS)
	if err != nil {
		t.Fatalf("LoadAll(defs.FS): %v", err)
	}

	checked := 0
	rejected := 0
	for _, bundle := range bundles {
		if bundle.Surface != nil {
			t.Fatalf("shipped bundle %q unexpectedly includes api_surface.json", bundle.Name)
		}
		if bundle.CLISurface == nil {
			continue
		}
		for _, command := range bundle.CLISurface.Commands {
			if command.Intent != "direct_read" || command.Availability != "implemented" || command.Operation == "" {
				continue
			}
			if len(command.APISurface) != 1 {
				t.Fatalf("%s %q has %d api_surface endpoints, want one", bundle.Name, command.Path, len(command.APISurface))
			}
			endpoint := command.APISurface[0]
			checked++
			if err := PreflightOperationDirectRead(bundle, command.Operation, endpoint.Method, endpoint.Path, maxOperationDirectReadBytes, command.OutputPolicy); err != nil {
				t.Fatalf("%s %q preflight with shipped endpoint ledger: %v", bundle.Name, command.Path, err)
			}

			withoutLedger := bundle
			withoutLedger.directReadLedger = nil
			err := PreflightOperationDirectRead(withoutLedger, command.Operation, endpoint.Method, endpoint.Path, maxOperationDirectReadBytes, command.OutputPolicy)
			if err == nil || !strings.Contains(err.Error(), "runtime operation endpoint ledger is unavailable") {
				t.Fatalf("%s %q preflight without endpoint ledger = %v, want fail-closed rejection", bundle.Name, command.Path, err)
			}
			rejected++
		}
	}
	if checked == 0 {
		t.Fatal("no implemented operation-backed direct-read commands were checked")
	}
	if rejected != checked {
		t.Fatalf("missing endpoint ledger rejected %d/%d commands", rejected, checked)
	}
	t.Logf("missing shipped runtime endpoint ledger rejects %d operation-backed implemented direct-read command(s) that the prior surface-optional check accepted", rejected)
}
