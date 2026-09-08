package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func TestProviderPlanNameWire184(t *testing.T) {
	requests := make(chan string, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r.Method + " " + r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"plan": "opensource", "ci_pipeline_size": 5})
	}))
	defer server.Close()
	bundle, err := engine.Load(defs.FS, "gitlab")
	if err != nil {
		t.Fatal(err)
	}
	path := []string{"api", "op-474554202f6170692f76342f6170706c69636174696f6e2f706c616e5f6c696d697473"}
	flags := connectorCommandFlags(parseFlags([]string{"--provider-plan-name", "opensource", "--plan-name", "pm-display"}).values)
	result, err := commandrunner.Run(t.Context(), engine.New(bundle, nil), commandrunner.Request{
		Path: path, Flags: flags, Config: connectors.RuntimeConfig{Config: map[string]string{"base_url": server.URL + "/api/v4"}, Secrets: map[string]string{"access_token": "synthetic-local-fixture"}},
	}, func(connectors.Record) error { t.Fatal("operation read emitted stream row"); return nil })
	if err != nil {
		t.Fatal(err)
	}
	if result.DirectRead == nil || result.DirectRead.Operation != "source_read_474554202f6170706c69636174696f6e2f706c616e5f6c696d697473" || !reflect.DeepEqual(result.DirectRead.Body, map[string]any{"plan": "opensource", "ci_pipeline_size": json.Number("5")}) {
		t.Fatalf("returned operation/body differs: %#v", result.DirectRead)
	}
	select {
	case got := <-requests:
		if got != "GET /api/v4/application/plan_limits?plan_name=opensource" {
			t.Fatalf("wrong provider wire: %s", got)
		}
	default:
		t.Fatal("no request")
	}
	select {
	case <-requests:
		t.Fatal("extra physical request")
	default:
	}
}
