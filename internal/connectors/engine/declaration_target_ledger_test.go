package engine

import (
	"testing"
	"testing/fstest"
)

func TestDeclarationTargetLedgerRejectsDuplicateProviderProvenanceAcrossBindings(t *testing.T) {
	raw := []byte(`{
  "schema_version": 1,
  "cohort": "test",
  "expected_connectors": 1,
  "expected_source_operations": 2,
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

func TestDeclarationTargetLedgerAcceptsGraphQLOperationIdentity(t *testing.T) {
	raw := []byte(`{
  "schema_version": 1,
  "cohort": "test",
  "expected_connectors": 1,
  "expected_source_operations": 1,
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
