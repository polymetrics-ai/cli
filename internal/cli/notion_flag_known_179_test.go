//go:build cp16_known_defects

package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

// CP16-FLAG-NOTION-179 is an explicitly unresolved CP23 authoring correction.
// This separately selected regression must remain failing until that correction;
// a green required-only fleet sweep does not certify this optional parameter.
func TestNotionSuppliedLimitKnown179(t *testing.T) {
	t.Run("omission_credential_boundary", func(t *testing.T) {
		root := t.TempDir()
		var out, diag bytes.Buffer
		if Run([]string{"init", "--root", root, "--json"}, &out, &diag) != 0 {
			t.Fatal("fixture init")
		}
		out.Reset()
		diag.Reset()
		code := runWithPreflightRegistry([]string{"notion", "meeting-note", "query", "--root", root, "--json"}, &out, &diag, bundleregistry.New())
		var result struct{ Error struct{ Message string } }
		if err := json.Unmarshal(out.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if code == 0 || result.Error.Message != "missing --credential" {
			t.Fatalf("omission boundary: %s", out.String())
		}
	})
	for _, prepare := range []bool{false, true} {
		name := "runner_binding_control"
		if prepare {
			name = "supplied_value_after_real_cli_preparation"
		}
		t.Run(name, func(t *testing.T) {
			requests := make(chan map[string]any, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" || r.URL.Path != "/v1/blocks/meeting_notes/query" {
					http.Error(w, "wrong fixed route", 400)
					return
				}
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					http.Error(w, "invalid JSON", 400)
					return
				}
				requests <- body
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"results":[{"id":"fixture-note"}],"has_more":false}`))
			}))
			defer server.Close()
			bundle, err := engine.Load(defs.FS, "notion")
			if err != nil {
				t.Fatal(err)
			}
			flags := parseFlags([]string{"--limit", "5"}).values
			if prepare {
				flags = connectorCommandFlags(flags)
			}
			_, err = commandrunner.Run(t.Context(), engine.New(bundle, nil), commandrunner.Request{
				Path: []string{"meeting-note", "query"}, Flags: flags, Limit: 5,
				Config: connectors.RuntimeConfig{Config: map[string]string{"base_url": server.URL}, Secrets: map[string]string{"token": "synthetic-local-fixture"}},
			}, func(connectors.Record) error { t.Fatal("operation read must not emit stream rows"); return nil })
			if err != nil {
				t.Fatal(err)
			}
			select {
			case body := <-requests:
				if body["limit"] != float64(5) {
					t.Fatalf("CP16-FLAG-NOTION-179: supplied provider body.limit omitted: got%#v want limit=5", body)
				}
			default:
				t.Fatal("no actual local request")
			}
		})
	}
}
