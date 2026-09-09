package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

type observedCredentialVault248 struct {
	credentialVault
	gets    int
	failure error
}

func (v *observedCredentialVault248) Get(ctx context.Context, id string) (map[string]string, error) {
	v.gets++
	if v.failure != nil {
		return nil, v.failure
	}
	return v.credentialVault.Get(ctx, id)
}

type observedDestination248 struct {
	batchDestination
	ids []string
}

func (d *observedDestination248) Write(ctx context.Context, req connectors.WriteRequest, rows []connectors.Record) (connectors.WriteResult, error) {
	for _, row := range rows {
		d.ids = append(d.ids, fmt.Sprint(row["id"]))
	}
	return d.batchDestination.Write(ctx, req, rows)
}

type bypassReadInputGuard248 struct{ *engine.Connector }

func (*bypassReadInputGuard248) ValidateReadInputs(context.Context, connectors.ReadInputValidationRequest) error {
	return nil
}

func TestRunETLRealInputVaultFrontier248(t *testing.T) {
	for _, mode := range []string{"invalid", "healthy_overlay", "bypassed_guard_fault", "structured_invalid", "structured_healthy_overlay", "structured_bypassed_guard_fault"} {
		t.Run(mode, func(t *testing.T) {
			structured := strings.HasPrefix(mode, "structured_")
			mode = strings.TrimPrefix(mode, "structured_")
			inputName, healthyInput, expectedURI := "page_size", "3", "/records?page_size=3"
			if structured {
				inputName = "filter"
				healthyInput = `["a","b","a"]`
				expectedURI = "/records?filter%5B%5D=a&filter%5B%5D=b&filter%5B%5D=a"
			}
			var sends atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" || r.URL.RequestURI() != expectedURI {
					t.Errorf("unexpected input wire %s %s", r.Method, r.URL.RequestURI())
				}
				sends.Add(1)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"data":[{"id":101},{"id":202},{"id":303}]}`))
			}))
			defer server.Close()
			files := fstest.MapFS{}
			for name, raw := range map[string]string{
				"metadata.json":                        `{"name":"inputprobe","display_name":"Input probe","description":"Input contract fixture","integration_type":"api","release_stage":"ga","capabilities":{"check":true,"read":true,"write":false,"query":false,"cdc":false,"dynamic_schema":false}}`,
				"spec.json":                            `{"type":"object","required":["base_url"],"properties":{"base_url":{"type":"string"}}}`,
				"streams.json":                         `{"base":{"url":"{{ config.base_url }}","auth":[],"headers":{},"pagination":{"type":"none"},"check":{"method":"GET","path":"/ping"}},"streams":[{"name":"records","path":"/records","records":{"path":"data"},"schema":"schemas/records.json","query":{"page_size":{"template":"{{ config.page_size }}","omit_when_absent":true}},"request_inputs":{"version":1,"schema":"schemas/inputs.json","bindings":[{"config_key":"page_size","in":"query","name":"page_size"}]}}]}`,
				"schemas/records.json":                 `{"type":"object","x-primary-key":["id"],"properties":{"id":{"type":"integer"}}}`,
				"schemas/inputs.json":                  `{"type":"object","properties":{"path":{"type":"object","properties":{},"additionalProperties":false},"query":{"type":"object","properties":{"page_size":{"type":"integer","minimum":2,"maximum":5}},"additionalProperties":false},"header":{"type":"object","properties":{},"additionalProperties":false}},"required":["path","query","header"],"additionalProperties":false}`,
				"fixtures/streams/records/page_1.json": `{"request":{"method":"GET","path":"/records","query":{}},"response":{"status":200,"body":{"data":[]}}}`,
				"docs.md":                              "# Overview\n\nTest.\n\n## Auth setup\n\nNone.\n\n## Streams notes\n\nTest.\n\n## Write actions & risks\n\nNone.\n",
			} {
				files["inputprobe/"+name] = &fstest.MapFile{Data: []byte(raw)}
			}
			if structured {
				for _, name := range []string{"streams.json", "schemas/inputs.json"} {
					files["inputprobe/"+name].Data = []byte(strings.ReplaceAll(string(files["inputprobe/"+name].Data), "page_size", "filter"))
				}
				var inputs, streams map[string]any
				if err := json.Unmarshal(files["inputprobe/schemas/inputs.json"].Data, &inputs); err != nil {
					t.Fatal(err)
				}
				query := inputs["properties"].(map[string]any)["query"].(map[string]any)
				query["properties"] = map[string]any{"filter": map[string]any{"type": "array", "minItems": 1, "maxItems": 3, "items": map[string]any{"type": "string"}}}
				raw, err := json.Marshal(inputs)
				if err != nil {
					t.Fatal(err)
				}
				files["inputprobe/schemas/inputs.json"].Data = raw
				if err := json.Unmarshal(files["inputprobe/streams.json"].Data, &streams); err != nil {
					t.Fatal(err)
				}
				contract := streams["streams"].([]any)[0].(map[string]any)["request_inputs"].(map[string]any)
				contract["query_encoding"] = map[string]any{"version": 1, "max_depth": 8, "max_members": 32, "max_items": 32, "max_pairs": 32, "max_bytes": 1024, "fields": map[string]any{"filter": map[string]any{"mode": "bracket_repeated", "empty_array": "omit", "null": "reject"}}}
				raw, err = json.Marshal(streams)
				if err != nil {
					t.Fatal(err)
				}
				files["inputprobe/streams.json"].Data = raw
			}
			bundle, err := engine.Load(files, "inputprobe")
			if err != nil {
				t.Fatal(err)
			}
			source := engine.New(bundle, nil)
			registry := connectors.NewEmptyRegistry()
			if mode == "bypassed_guard_fault" {
				err = registry.Register(&bypassReadInputGuard248{Connector: source})
			} else {
				err = registry.Register(source)
			}
			if err != nil {
				t.Fatal(err)
			}
			destination := &observedDestination248{}
			if err := registry.Register(destination); err != nil {
				t.Fatal(err)
			}
			root := t.TempDir()
			if err := InitProject(root); err != nil {
				t.Fatal(err)
			}
			a, err := OpenWithRegistry(root, registry)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = a.Close() })
			if _, err = a.AddCredential(t.Context(), AddCredentialRequest{Name: "source", Connector: "inputprobe", Config: map[string]string{"base_url": server.URL, inputName: "invalid"}}); err != nil {
				t.Fatal(err)
			}
			if _, err = a.AddCredential(t.Context(), AddCredentialRequest{Name: "destination", Connector: destination.Name(), Config: map[string]string{"path": filepath.Join(root, "out")}}); err != nil {
				t.Fatal(err)
			}
			overlay := map[string]string{}
			if mode == "healthy_overlay" {
				overlay[inputName] = healthyInput
			}
			if _, err = a.CreateConnection(t.Context(), CreateConnectionRequest{Name: "input_job", Source: EndpointConfig{Connector: "inputprobe", Credential: "source", Config: overlay}, Destination: EndpointConfig{Connector: destination.Name(), Credential: "destination"}, Streams: map[string]StreamConfig{"records": {SyncMode: "full_refresh_overwrite", PrimaryKey: []string{"id"}, DestinationTable: "records"}}}); err != nil {
				t.Fatal(err)
			}
			vaultFault := errors.New("observed vault read frontier")
			observer := &observedCredentialVault248{credentialVault: a.vault}
			if mode != "healthy_overlay" {
				observer.failure = vaultFault
			}
			a.vault = observer
			beforeRuns := len(a.state.Runs)
			run, err := a.RunETL(t.Context(), RunETLRequest{Connection: "input_job", Stream: "records", BatchSize: 2})
			switch mode {
			case "invalid":
				if err == nil || observer.gets != 0 || sends.Load() != 0 || len(a.state.Runs) != beforeRuns || len(destination.batches) != 0 {
					t.Fatalf("input frontier err=%v vault=%d sends=%d runs=%d batches=%d", err, observer.gets, sends.Load(), len(a.state.Runs), len(destination.batches))
				}
			case "bypassed_guard_fault":
				if !errors.Is(err, vaultFault) || observer.gets == 0 || sends.Load() != 0 {
					t.Fatalf("fault control did not reach actual vault: err=%v gets=%d sends=%d", err, observer.gets, sends.Load())
				}
			case "healthy_overlay":
				if err != nil || observer.gets == 0 || sends.Load() != 1 || run.RecordsRead != 3 || run.RecordsLoaded != 3 || len(destination.batches) != 2 || !reflect.DeepEqual(destination.ids, []string{"101", "202", "303"}) {
					t.Fatalf("healthy frontier err=%v vault=%d sends=%d run=%+v batches=%v", err, observer.gets, sends.Load(), run, destination.batches)
				}
			}
		})
	}
}
