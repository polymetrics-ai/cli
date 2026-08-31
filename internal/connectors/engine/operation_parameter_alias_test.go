package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestProviderQueryParameterCLIName(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		want     string
		ok       bool
	}{
		{name: "plain", provider: "per_page", want: "per-page", ok: true},
		{name: "bracketed", provider: "filter[state]", want: "filter-state", ok: true},
		{name: "bracketed underscore segment", provider: "not[assignee_id]", want: "not-assignee-id", ok: true},
		{name: "nested bracket syntax remains unprojected", provider: "filter[state][name]", ok: false},
		// This foundation is deliberately narrow. A non-ASCII key remains
		// source-mapped/nonimplemented until its own policy is approved.
		{name: "unicode remains unprojected", provider: "filtré", ok: false},
		{name: "unsupported provider sigil remains unprojected", provider: "$filter", ok: false},
		{name: "empty bracket segment remains unprojected", provider: "filter[]", ok: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := ProviderQueryParameterCLIName(test.provider)
			if got != test.want || ok != test.ok {
				t.Fatalf("ProviderQueryParameterCLIName(%q) = (%q, %t), want (%q, %t)", test.provider, got, ok, test.want, test.ok)
			}
		})
	}
}

func TestProviderQueryParameterCLINamesRejectNormalizationCollision(t *testing.T) {
	err := ValidateProviderQueryParameterCLINames([]string{"filter[state]", "filter-state"})
	if err == nil || !strings.Contains(err.Error(), "filter-state") || !strings.Contains(err.Error(), "collision") {
		t.Fatalf("ValidateProviderQueryParameterCLINames collision error = %v, want named alias collision", err)
	}
	if err := ValidateProviderQueryParameterCLINames([]string{"filter[state]", "state"}); err != nil {
		t.Fatalf("ValidateProviderQueryParameterCLINames distinct aliases: %v", err)
	}
}

func TestOperationDirectReadProviderQueryAliasPreservesExactWireKey(t *testing.T) {
	var gotRawQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRawQuery = r.URL.RawQuery
		if got := r.URL.Query().Get("filter[state]"); got != "opened" {
			t.Fatalf("provider filter[state] = %q, want opened", got)
		}
		if got := r.URL.Query().Get("filter-state"); got != "" {
			t.Fatalf("CLI alias leaked onto wire as filter-state=%q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	op := OperationSpec{
		ID: "gitlab.projects.list", Kind: "rest_read", Summary: "List GitLab projects", Risk: "low", Approval: "none", OutputPolicy: "json_redacted",
		REST: &RESTOperationSpec{
			Method: http.MethodGet, Path: "/api/v4/projects", MaxBytes: 1024,
			Parameters: []OperationParameter{{Name: "filter[state]", In: "query", Type: "string"}},
		},
	}
	bundle := operationBindingTestBundle(srv.URL, op)
	bundle.CLISurface = &CLISurface{Commands: []CLICommand{{
		Path: "projects list", Intent: "direct_read", Availability: "implemented", Operation: op.ID,
		Flags: []CLIFlag{{Name: "filter-state", Type: "string", MapsTo: "query.filter[state]"}},
	}}}

	if _, err := OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{
		Operation: op.ID,
		Query:     map[string]string{"filter[state]": "opened"},
		CommandBindings: &connectors.OperationDirectReadBindings{
			Query: []string{"filter[state]"},
		},
		MaxBytes: 1024,
	}, nil); err != nil {
		t.Fatalf("OperationDirectRead: %v", err)
	}
	if gotRawQuery != "filter%5Bstate%5D=opened" {
		t.Fatalf("wire raw query = %q, want exact provider key encoding", gotRawQuery)
	}
}

func TestOperationParameterAliasRemainsUnavailableForMutations(t *testing.T) {
	err := validateOperationHeaderParameters(OperationSpec{
		ID: "gitlab.projects.update", Kind: "rest_write",
		REST: &RESTOperationSpec{Parameters: []OperationParameter{{Name: "filter[state]", In: "query", Type: "string"}}},
	})
	if err == nil {
		t.Fatal("rest_write accepted a provider-query alias; only typed direct reads may use this foundation")
	}
}
