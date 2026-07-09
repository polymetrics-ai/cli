package commandrunner

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
)

func TestRunMondayBoardListCommand(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		var body struct {
			Query string `json:"query"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if !strings.Contains(body.Query, "boards") || !strings.Contains(body.Query, "page: 1") {
			t.Fatalf("query = %q, want boards page 1 query", body.Query)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"boards":[{"id":"b1","name":"Roadmap","state":"active"},{"id":"b2","name":"Launch","state":"active"}]}}`))
	}))
	defer srv.Close()

	connector, ok := bundleregistry.New().Get("monday")
	if !ok {
		t.Fatal("monday connector not found")
	}

	var rows []connectors.Record
	result, err := Run(context.Background(), connector, Request{
		Path: []string{"board", "list"},
		Config: connectors.RuntimeConfig{Config: map[string]string{
			"base_url":  srv.URL,
			"page_size": "2",
			"max_pages": "1",
		}},
		Limit: 2,
	}, func(record connectors.Record) error {
		rows = append(rows, record)
		return nil
	})
	if err != nil {
		t.Fatalf("Run board list: %v", err)
	}
	if result.Connector != "monday" || result.Command != "board list" || result.Stream != "boards" || result.Count != 2 {
		t.Fatalf("result = %+v, want monday board list boards count 2", result)
	}
	if len(rows) != 2 || rows[0]["id"] != "b1" || rows[0]["name"] != "Roadmap" {
		t.Fatalf("rows = %+v, want sanitized board records", rows)
	}
}

func TestRunMondayDirectReadCommands(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		var body struct {
			Query         string `json:"query"`
			OperationName string `json:"operationName"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		switch body.OperationName {
		case "MondayMeGetMe":
			if !strings.Contains(body.Query, "me") || strings.Contains(body.Query, "__typename") {
				t.Fatalf("me query = %q, want fixed me query", body.Query)
			}
			_, _ = w.Write([]byte(`{"data":{"me":{"id":"u1","name":"Ada"}}}`))
		case "MondayAccountGetAccount":
			if !strings.Contains(body.Query, "account") || strings.Contains(body.Query, "__typename") {
				t.Fatalf("account query = %q, want fixed account query", body.Query)
			}
			_, _ = w.Write([]byte(`{"data":{"account":{"id":"a1","name":"Example"}}}`))
		default:
			t.Fatalf("operationName = %q, want Monday me/account direct read", body.OperationName)
		}
	}))
	defer srv.Close()

	connector, ok := bundleregistry.New().Get("monday")
	if !ok {
		t.Fatal("monday connector not found")
	}

	tests := []struct {
		name string
		path []string
		want string
	}{
		{name: "me", path: []string{"me", "view"}, want: "me"},
		{name: "account", path: []string{"account", "view"}, want: "account"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Run(context.Background(), connector, Request{
				Path: tt.path,
				Config: connectors.RuntimeConfig{Config: map[string]string{
					"base_url": srv.URL,
				}},
			}, func(connectors.Record) error {
				t.Fatal("emit called for direct-read command")
				return nil
			})
			if err != nil {
				t.Fatalf("Run %v: %v", tt.path, err)
			}
			if result.DirectRead == nil {
				t.Fatalf("DirectRead = nil, want result")
			}
			data, ok := result.DirectRead.Body.(map[string]any)
			if !ok || data[tt.want] == nil {
				t.Fatalf("direct read body = %+v, want %s payload", result.DirectRead.Body, tt.want)
			}
		})
	}
}
