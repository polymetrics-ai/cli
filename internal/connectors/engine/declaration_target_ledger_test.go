package engine

import (
	"encoding/json"
	"strings"
	"testing"
	"testing/fstest"
)

func TestDeclarationTargetLedgerRejectsNoncanonicalProviderCitations(t *testing.T) {
	tests := []struct {
		name      string
		sourceURL string
		want      string
	}{
		{name: "uppercase DNS host", sourceURL: "https://PROVIDER.EXAMPLE.TEST/reference", want: "canonical provider citation URL"},
		{name: "explicit default HTTPS port", sourceURL: "https://provider.example.test:443/reference", want: "canonical provider citation URL"},
		{name: "unstable query order", sourceURL: "https://provider.example.test/reference?b=2&a=1", want: "canonical provider citation URL"},
		{name: "non-normalized escaped path", sourceURL: "https://provider.example.test/%72eference", want: "canonical provider citation URL"},
		{name: "trailing-dot DNS host", sourceURL: "https://provider.example.test./reference", want: "provider citation URL"},
		{name: "ambiguous empty DNS label", sourceURL: "https://provider..example.test/reference", want: "provider citation URL"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			source := declarationAdmissionSourceOperation{
				Connector: "acme", ID: "acme.widgets.list", Protocol: "rest",
				SourceURL: testCase.sourceURL, Location: "Widgets > list", ProviderOperationID: "widgets/list",
				Method: "GET", Path: "/widgets", Binding: CommandBindingIdentity{Kind: "stream", ID: "widgets"},
				DestructiveKind: "none",
			}
			catalog := declarationAdmissionSourceCatalog{
				SchemaVersion: 2, Cohort: "test",
				SourceOperations: []declarationAdmissionSourceOperation{source},
			}
			raw, err := json.Marshal(catalog)
			if err != nil {
				t.Fatalf("marshal compact ledger: %v", err)
			}
			_, err = loadDeclarationTargetLedgers(fstest.MapFS{DeclarationAdmissionSourcesFile: &fstest.MapFile{Data: raw}})
			if err == nil || !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("load noncanonical compact ledger = %v, want %q refusal", err, testCase.want)
			}
		})
	}
}

func TestDeclarationTargetLedgerRejectsDuplicateProviderProvenanceAcrossBindings(t *testing.T) {
	raw := []byte(`{
  "schema_version": 2,
  "cohort": "test",
  "source_operations": [
    {
      "connector": "acme",
      "id": "acme.widgets.list",
      "protocol": "rest",
      "source_url": "https://provider.example.test/reference",
      "location": "Widgets > list",
      "operation_id": "widgets/list",
      "method": "GET",
      "base_path": "/v1",
      "path": "/widgets",
      "binding": {"kind": "stream", "id": "widgets"},
      "destructive_kind": "none"
    },
    {
      "connector": "acme",
      "id": "acme.widgets.list-alias",
      "protocol": "rest",
      "source_url": "https://provider.example.test/reference",
      "location": "Widgets > list",
      "operation_id": "widgets/list",
      "method": "GET",
      "base_path": "/v1",
      "path": "/widgets",
      "binding": {"kind": "stream", "id": "widgets_alias"},
      "destructive_kind": "none"
    }
  ]
}`)
	_, err := loadDeclarationTargetLedgers(fstest.MapFS{DeclarationAdmissionSourcesFile: &fstest.MapFile{Data: raw}})
	if err == nil {
		t.Fatal("compact declaration target ledger accepted duplicate provider provenance under another binding")
	}
}

func TestDeclarationTargetLedgerRejectsDuplicateJSONMembers(t *testing.T) {
	raw := []byte(`{
  "schema_version": 2,
  "cohort": "test",
  "source_operations": [{
    "connector": "acme",
    "id": "acme.widgets.list",
    "protocol": "rest",
    "source_url": "https://provider.example.test/reference",
    "source_url": "https://provider.example.test/reference",
    "location": "Widgets > list",
    "operation_id": "widgets/list",
    "method": "GET",
    "path": "/widgets",
    "binding": {"kind": "stream", "id": "widgets"},
    "destructive_kind": "none"
  }]
}`)

	_, err := loadDeclarationTargetLedgers(fstest.MapFS{DeclarationAdmissionSourcesFile: &fstest.MapFile{Data: raw}})
	if err == nil || !strings.Contains(err.Error(), "duplicate JSON object member") {
		t.Fatalf("load compact declaration ledger error = %v, want duplicate JSON-member rejection", err)
	}
}

func TestDeclarationTargetLedgerAcceptsGraphQLOperationIdentity(t *testing.T) {
	raw := []byte(`{
  "schema_version": 2,
  "cohort": "test",
  "source_operations": [{
    "connector": "acme",
    "id": "acme.graphql.list-widgets",
    "protocol": "graphql",
    "source_url": "https://provider.example.test/graphql/reference",
    "location": "Query.listWidgets",
    "operation_id": "ListWidgets",
    "method": "GRAPHQL",
    "path": "ListWidgets",
    "binding": {"kind": "stream", "id": "widgets"},
    "destructive_kind": "none"
  }]
}`)
	ledgers, err := loadDeclarationTargetLedgers(fstest.MapFS{DeclarationAdmissionSourcesFile: &fstest.MapFile{Data: raw}})
	if err != nil {
		t.Fatalf("load GraphQL declaration target ledger: %v", err)
	}
	target, ok := ledgers["acme"].target("acme.graphql.list-widgets")
	if !ok || target.Method != "GRAPHQL" || target.Path != "ListWidgets" {
		t.Fatalf("GraphQL target = %+v (found=%v), want operation identity", target, ok)
	}
}
