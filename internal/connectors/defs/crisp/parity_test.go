package crisp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
)

const crispProviderArtifact = "https://docs.crisp.chat/static/data/collections/rest-api-v1.postman"

type waveOneCommand struct {
	command string
	stream  string
	method  string
	path    string
}

var waveOneCommands = []waveOneCommand{
	{"conversations list", "list_conversations", "GET", "/v1/website/{website_id}/conversations/{page_number}"},
	{"conversations suggested-segments", "suggested_conversation_segments", "GET", "/v1/website/{website_id}/conversations/suggest/segments/{page_number}"},
	{"conversations suggested-data", "suggested_conversation_data", "GET", "/v1/website/{website_id}/conversations/suggest/data/{page_number}"},
	{"conversations spam-list", "spam_conversations", "GET", "/v1/website/{website_id}/conversations/spams/{page_number}"},
	{"conversations spam-content", "spam_conversation_content", "GET", "/v1/website/{website_id}/conversations/spam/{spam_id}/content"},
	{"conversation get", "conversation", "GET", "/v1/website/{website_id}/conversation/{session_id}"},
	{"conversation messages", "conversation_messages", "GET", "/v1/website/{website_id}/conversation/{session_id}/messages"},
	{"conversation message", "conversation_message", "GET", "/v1/website/{website_id}/conversation/{session_id}/message/{fingerprint}"},
	{"conversation routing", "conversation_routing", "GET", "/v1/website/{website_id}/conversation/{session_id}/routing"},
	{"conversation meta", "conversation_meta", "GET", "/v1/website/{website_id}/conversation/{session_id}/meta"},
	{"conversation original-message", "conversation_original_message", "GET", "/v1/website/{website_id}/conversation/{session_id}/original/{original_id}"},
	{"conversation pages", "conversation_pages", "GET", "/v1/website/{website_id}/conversation/{session_id}/pages/{page_number}"},
	{"conversation events", "conversation_events", "GET", "/v1/website/{website_id}/conversation/{session_id}/events/{page_number}"},
	{"conversation files", "conversation_files", "GET", "/v1/website/{website_id}/conversation/{session_id}/files/{page_number}"},
	{"conversation state", "conversation_state", "GET", "/v1/website/{website_id}/conversation/{session_id}/state"},
	{"conversation relations", "conversation_relations", "GET", "/v1/website/{website_id}/conversation/{session_id}/relations"},
	{"conversation participants", "conversation_participants", "GET", "/v1/website/{website_id}/conversation/{session_id}/participants"},
	{"conversation block-status", "conversation_block_status", "GET", "/v1/website/{website_id}/conversation/{session_id}/block"},
	{"conversation verify-status", "conversation_verify_status", "GET", "/v1/website/{website_id}/conversation/{session_id}/verify"},
	{"conversation browsing", "conversation_browsing", "GET", "/v1/website/{website_id}/conversation/{session_id}/browsing"},
	{"conversation call", "conversation_call", "GET", "/v1/website/{website_id}/conversation/{session_id}/call"},
}

func TestCrispWaveOneParityContract(t *testing.T) {
	bundle, err := engine.Load(os.DirFS(".."), "crisp")
	if err != nil {
		t.Fatalf("load Crisp bundle: %v", err)
	}
	if bundle.Surface == nil {
		t.Fatal("Crisp api_surface.json did not load")
	}
	if bundle.CLISurface == nil {
		t.Fatal("Crisp cli_surface.json did not load")
	}
	if !bundle.Metadata.Capabilities.Check || !bundle.Metadata.Capabilities.Read || bundle.Metadata.Capabilities.Write {
		t.Fatalf("Crisp capabilities = %+v, want check/read only for Wave 1", bundle.Metadata.Capabilities)
	}
	if got := len(bundle.Surface.Endpoints); got != 234 {
		t.Fatalf("Crisp provider ledger rows = %d, want 234", got)
	}
	if got := len(bundle.Streams); got != len(waveOneCommands) {
		t.Fatalf("Crisp stream count = %d, want %d Wave 1 streams", got, len(waveOneCommands))
	}
	if len(bundle.HTTP.Auth) != 1 || bundle.HTTP.Auth[0].Mode != "basic" {
		t.Fatalf("Crisp auth = %+v, want one HTTP Basic auth declaration", bundle.HTTP.Auth)
	}
	if got := bundle.HTTP.Headers["X-Crisp-Tier"]; got != "{{ config.token_tier }}" {
		t.Fatalf("X-Crisp-Tier header = %q, want config token tier", got)
	}

	streams := map[string]bool{}
	for _, stream := range bundle.Streams {
		streams[stream.Name] = true
	}
	endpoints := map[string]engine.SurfaceEndpoint{}
	blocked := 0
	for _, endpoint := range bundle.Surface.Endpoints {
		key := endpoint.Method + " " + endpoint.Path
		endpoints[key] = endpoint
		if endpoint.Operation != nil {
			blocked++
			if endpoint.Operation.SourceURL != crispProviderArtifact || endpoint.Operation.Reason == "" {
				t.Fatalf("blocked ledger row %s lost its cited named reason: %+v", key, endpoint.Operation)
			}
		}
	}
	if blocked != 213 {
		t.Fatalf("blocked Crisp provider operations = %d, want 213 after Wave 1", blocked)
	}

	commands := map[string]engine.CLICommand{}
	for _, command := range bundle.CLISurface.Commands {
		commands[command.Path] = command
	}
	connector := engine.New(bundle, nil)
	for _, want := range waveOneCommands {
		t.Run(want.command, func(t *testing.T) {
			if !streams[want.stream] {
				t.Fatalf("missing stream %q", want.stream)
			}
			endpoint, ok := endpoints[want.method+" "+want.path]
			if !ok || endpoint.CoveredBy == nil || endpoint.CoveredBy.Stream != want.stream || endpoint.Operation != nil {
				t.Fatalf("ledger endpoint %s %s = %+v, want covered_by.stream=%q only", want.method, want.path, endpoint, want.stream)
			}
			command, ok := commands[want.command]
			if !ok {
				t.Fatalf("missing CLI command %q", want.command)
			}
			if command.Intent != "etl" || command.Availability != "implemented" || command.Stream != want.stream {
				t.Fatalf("CLI command %q = %+v, want implemented ETL stream %q", want.command, command, want.stream)
			}
			if command.SourceURL != crispProviderArtifact {
				t.Fatalf("CLI command %q source_url = %q, want provider artifact citation", want.command, command.SourceURL)
			}
			if len(command.RedactFields) != 0 {
				t.Fatalf("CLI command %q declares forbidden redact_fields: %v", want.command, command.RedactFields)
			}
			if len(command.APISurface) != 1 || command.APISurface[0].Method != want.method || command.APISurface[0].Path != want.path {
				t.Fatalf("CLI command %q api_surface = %+v, want %s %s", want.command, command.APISurface, want.method, want.path)
			}
			if err := commandrunner.Preflight(connector, strings.Fields(want.command)); err != nil {
				t.Fatalf("Preflight(%q): %v", want.command, err)
			}
		})
	}
	manual := connectors.RenderConnectorManual(connector)
	for _, want := range []string{
		"Crisp conversation-list read commands\n    conversations list",
		"Crisp conversation-scoped read commands\n    conversation get",
	} {
		if !strings.Contains(manual, want) {
			t.Fatalf("Crisp manual missing grouped commands %q:\n%s", want, manual)
		}
	}
	if strings.Contains(manual, "Other Commands") {
		t.Fatalf("Crisp manual has unexpected Other Commands section:\n%s", manual)
	}
}

func TestCrispPageNumberMinimumRejectsCommandOverride(t *testing.T) {
	bundle, err := engine.Load(os.DirFS(".."), "crisp")
	if err != nil {
		t.Fatalf("load Crisp bundle: %v", err)
	}
	pageNumberCommands := []string{
		"conversations list",
		"conversations suggested-segments",
		"conversations suggested-data",
		"conversations spam-list",
		"conversation pages",
		"conversation events",
		"conversation files",
	}
	for _, path := range pageNumberCommands {
		t.Run(path, func(t *testing.T) {
			var command *engine.CLICommand
			for i := range bundle.CLISurface.Commands {
				candidate := &bundle.CLISurface.Commands[i]
				if candidate.Path == path {
					command = candidate
					break
				}
			}
			if command == nil {
				t.Fatalf("missing command %q", path)
			}
			for _, flag := range command.Flags {
				if flag.Name == "page-number" {
					if flag.Type != "integer" || flag.Minimum == nil || *flag.Minimum != 1 {
						t.Fatalf("page-number flag = %+v, want integer minimum 1", flag)
					}
					return
				}
			}
			t.Fatalf("command %q missing page-number flag", path)
		})
	}

	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		requests.Add(1)
	}))
	defer server.Close()

	_, err = commandrunner.Run(context.Background(), engine.New(bundle, nil), commandrunner.Request{
		Path:  []string{"conversations", "list"},
		Flags: map[string][]string{"page-number": {"0"}},
		Config: connectors.RuntimeConfig{Config: map[string]string{
			"base_url":    server.URL,
			"website_id":  "fixture-website",
			"page_number": "1",
			"token_tier":  "website",
		}},
	}, func(connectors.Record) error { return nil })
	var minimumErr *commandrunner.MinimumFlagError
	if !errors.As(err, &minimumErr) {
		t.Fatalf("Crisp page-number 0 error = %T %v, want MinimumFlagError", err, err)
	}
	if minimumErr.Parameter != "page-number" || minimumErr.Minimum != 1 {
		t.Fatalf("MinimumFlagError = %+v, want page-number minimum 1", minimumErr)
	}
	if requests.Load() != 0 {
		t.Fatalf("rejected page number sent %d requests, want none", requests.Load())
	}
}

func TestCrispPageNumberMinimumRejectsConfigOverride(t *testing.T) {
	bundle, err := engine.Load(os.DirFS(".."), "crisp")
	if err != nil {
		t.Fatalf("load Crisp bundle: %v", err)
	}

	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		requests.Add(1)
	}))
	defer server.Close()

	_, err = commandrunner.Run(context.Background(), engine.New(bundle, nil), commandrunner.Request{
		Path: []string{"conversations", "list"},
		Config: connectors.RuntimeConfig{Config: map[string]string{
			"base_url":    server.URL,
			"website_id":  "fixture-website",
			"page_number": "0",
			"token_tier":  "website",
		}},
	}, func(connectors.Record) error { return nil })
	var minimumErr *commandrunner.MinimumFlagError
	if !errors.As(err, &minimumErr) {
		t.Fatalf("Crisp config page_number 0 error = %T %v, want MinimumFlagError", err, err)
	}
	if minimumErr.Parameter != "page-number" || minimumErr.Minimum != 1 {
		t.Fatalf("MinimumFlagError = %+v, want page-number minimum 1", minimumErr)
	}
	if requests.Load() != 0 {
		t.Fatalf("rejected config page number sent %d requests, want none", requests.Load())
	}
}

func TestCrispListCommandPreservesFixtureContent(t *testing.T) {
	fixture, err := os.ReadFile("fixtures/streams/list_conversations/page_1.json")
	if err != nil {
		t.Fatalf("read list-conversations fixture: %v", err)
	}
	var replay struct {
		Response struct {
			Status int             `json:"status"`
			Body   json.RawMessage `json:"body"`
		} `json:"response"`
	}
	if err := json.Unmarshal(fixture, &replay); err != nil {
		t.Fatalf("decode list-conversations fixture: %v", err)
	}

	var gotMethod, gotPath, gotTier, gotUser, gotKey string
	var gotQuery string
	requestDone := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer close(requestDone)
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotTier = r.Header.Get("X-Crisp-Tier")
		gotUser, gotKey, _ = r.BasicAuth()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(replay.Response.Status)
		_, _ = w.Write(replay.Response.Body)
	}))
	defer server.Close()

	bundle, err := engine.Load(os.DirFS(".."), "crisp")
	if err != nil {
		t.Fatalf("load Crisp bundle: %v", err)
	}
	connector := engine.New(bundle, nil)
	var records []connectors.Record
	result, err := commandrunner.Run(context.Background(), connector, commandrunner.Request{
		Path:  strings.Fields("conversations list"),
		Flags: map[string][]string{"per-page": {"37"}},
		Config: connectors.RuntimeConfig{
			Config: map[string]string{
				"base_url":    server.URL,
				"website_id":  "fixture-website",
				"page_number": "1",
				"token_tier":  "website",
			},
			Secrets: map[string]string{
				"identifier": "fixture-identifier",
				"key":        "fixture-key",
			},
		},
		Limit: 1,
	}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	})
	if err != nil {
		t.Fatalf("run Crisp conversations list fixture: %v", err)
	}
	<-requestDone
	if result.Count != 1 || len(records) != 1 {
		t.Fatalf("run result/records = %+v/%#v, want one emitted fixture record", result, records)
	}
	if got := records[0]["content"]; got != "fixture-visible-crisp-content" {
		t.Fatalf("emitted fixture content = %#v, want complete Crisp content", got)
	}
	if gotMethod != http.MethodGet || gotPath != "/v1/website/fixture-website/conversations/1" || gotQuery != "per_page=37" {
		t.Fatalf("fixture request = %s %s?%s, want caller-selected GET /v1/website/fixture-website/conversations/1?per_page=37", gotMethod, gotPath, gotQuery)
	}
	if gotTier != "website" || gotUser != "fixture-identifier" || gotKey != "fixture-key" {
		t.Fatalf("fixture auth = tier=%q user=%q key=%q, want complete declared auth", gotTier, gotUser, gotKey)
	}
}
