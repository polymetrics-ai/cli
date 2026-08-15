package database_test

import (
	"encoding/json"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
)

func TestCatalogFromConnectorCatalogProjectsSealedAPIStream(t *testing.T) {
	catalog := connectors.Catalog{
		Connector: "github",
		Streams: []connectors.Stream{{
			Name:       "issues",
			PrimaryKey: []string{"node_id"},
			Schema: json.RawMessage(`{
				"type":"object",
				"required":["node_id"],
				"properties":{
					"node_id":{"type":"string"},
					"number":{"type":"integer"},
					"locked":{"type":"boolean"},
					"labels":{"type":"array","items":{"type":"object"}},
					"body":{"type":["string","null"]}
				}
			}`),
		}},
	}

	typed, err := database.CatalogFromConnectorCatalog(catalog, "issues")
	if err != nil {
		t.Fatalf("CatalogFromConnectorCatalog() error = %v", err)
	}
	relations := typed.Relations()
	if len(relations) != 1 {
		t.Fatalf("relations = %d, want 1", len(relations))
	}
	relation := relations[0]
	if relation.Ref.Schema.Catalog.Name != "github" || relation.Ref.Schema.Name != "api" || relation.Ref.Name != "issues" {
		t.Fatalf("relation ref = %+v, want github.api.issues", relation.Ref)
	}
	if len(relation.Columns) != 5 {
		t.Fatalf("columns = %d, want 5", len(relation.Columns))
	}
	kinds := map[string]database.LogicalKind{}
	nullable := map[string]bool{}
	for _, column := range relation.Columns {
		kinds[column.Ref.Name] = column.Type.Kind()
		nullable[column.Ref.Name] = column.Nullable
	}
	if kinds["node_id"] != database.LogicalString || kinds["number"] != database.LogicalSignedInteger || kinds["locked"] != database.LogicalBoolean || kinds["labels"] != database.LogicalJSON || kinds["body"] != database.LogicalString {
		t.Fatalf("projected kinds = %+v", kinds)
	}
	if nullable["node_id"] || !nullable["body"] || !nullable["number"] {
		t.Fatalf("projected nullability = %+v", nullable)
	}
	if len(relation.Keys) != 1 || relation.Keys[0].Kind != database.KeyPrimary || len(relation.Keys[0].Columns) != 1 || relation.Keys[0].Columns[0].Name != "node_id" {
		t.Fatalf("projected keys = %+v", relation.Keys)
	}
}

func TestCatalogFromConnectorCatalogRefusesUnsealedOrAmbiguousStream(t *testing.T) {
	for name, catalog := range map[string]connectors.Catalog{
		"missing schema": {Connector: "github", Streams: []connectors.Stream{{Name: "issues"}}},
		"missing stream": {Connector: "github", Streams: []connectors.Stream{{Name: "repository", Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"integer"}}}`)}}},
		"duplicate stream": {Connector: "github", Streams: []connectors.Stream{
			{Name: "issues", Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"integer"}}}`)},
			{Name: "issues", Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"integer"}}}`)},
		}},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := database.CatalogFromConnectorCatalog(catalog, "issues"); err == nil {
				t.Fatal("CatalogFromConnectorCatalog() accepted an unsealed or ambiguous stream")
			}
		})
	}
}
