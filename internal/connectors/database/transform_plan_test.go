package database_test

import (
	"errors"
	"reflect"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
)

func TestTransformPlanV1NormalizesAndValidatesTypedProjectedFilter(t *testing.T) {
	plan, err := database.ParseTransformPlanV1([]byte(`{
  "where": {"not_equal": [{"mod": ["id", 10]}, 0]},
  "select": [
    {"source":"id","target":"event_id","type":"int64"},
    {"expr":{"date":"updated_at"},"target":"event_date","type":"date"},
    {"expr":{"cast":{"multiply":["amount",100]}},"target":"amount_cents","type":"int64","rounding":"exact"},
    {"expr":{"upper":"status"},"target":"status","type":"string"}
  ], "version": 1
}`))
	if err != nil {
		t.Fatalf("ParseTransformPlanV1() error = %v", err)
	}
	if got := string(plan.NormalizedJSON()); got != `{"version":1,"select":[{"source":"id","target":"event_id","type":"int64"},{"target":"event_date","type":"date","expr":{"date":"updated_at"}},{"target":"amount_cents","type":"int64","rounding":"exact","expr":{"cast":{"multiply":["amount",100]}}},{"target":"status","type":"string","expr":{"upper":"status"}}],"where":{"not_equal":[{"mod":["id",10]},0]}}` {
		t.Fatalf("NormalizedJSON() = %s", got)
	}
	if len(plan.Hash()) != 64 {
		t.Fatalf("Hash() = %q, want SHA-256 hex", plan.Hash())
	}
	relation := transformPlanRelation(t)
	if err := plan.ValidateAgainstRelation(relation); err != nil {
		t.Fatalf("ValidateAgainstRelation() error = %v", err)
	}
	mapping, err := plan.OutputMapping()
	if err != nil {
		t.Fatalf("OutputMapping() error = %v", err)
	}
	if got, want := len(mapping.Columns()), 4; got != want {
		t.Fatalf("OutputMapping columns = %d, want %d", got, want)
	}
	fields, err := plan.SourceFields()
	if err != nil {
		t.Fatalf("SourceFields() error = %v", err)
	}
	if want := []string{"amount", "id", "status", "updated_at"}; !reflect.DeepEqual(fields, want) {
		t.Fatalf("SourceFields() = %v, want %v", fields, want)
	}
}

func TestTransformPlanV1RefusesUnknownOperationBeforeConnectionIO(t *testing.T) {
	_, err := database.ParseTransformPlanV1([]byte(`{"version":1,"select":[{"expr":{"sql":"DROP TABLE target"},"target":"event_id","type":"int64"}]}`))
	if !errors.Is(err, database.ErrTransformPlanInvalid) {
		t.Fatalf("ParseTransformPlanV1(unsafe expression) error = %T %v, want ErrTransformPlanInvalid", err, err)
	}
}

// TestTransformPlanV1PreservesTypedTimestampDirectly is the identity-mapping
// half of the throughput tax proof. A timestamp is a closed logical type, not
// a request for source SQL or format-driven coercion, and must be preservable
// without turning a control mapping into a date transform.
func TestTransformPlanV1PreservesTypedTimestampDirectly(t *testing.T) {
	plan, err := database.ParseTransformPlanV1([]byte(`{"version":1,"select":[{"source":"updated_at","target":"updated_at","type":"timestamp"}]}`))
	if err != nil {
		t.Fatalf("ParseTransformPlanV1(timestamp direct) error = %v", err)
	}
	if err := plan.ValidateAgainstRelation(transformPlanRelation(t)); err != nil {
		t.Fatalf("ValidateAgainstRelation(timestamp direct) error = %v", err)
	}
	output, err := plan.OutputRelation(transformPlanRelation(t))
	if err != nil || len(output.Columns) != 1 || output.Columns[0].Type.Kind() != database.LogicalTimestamp {
		t.Fatalf("OutputRelation(timestamp direct) = %#v, %v; want one typed timestamp", output, err)
	}
}

func TestTransformPlanV1RejectsDuplicateTargetAndSchemaDrift(t *testing.T) {
	t.Run("duplicate target", func(t *testing.T) {
		_, err := database.ParseTransformPlanV1([]byte(`{"version":1,"select":[{"source":"id","target":"event_id","type":"int64"},{"source":"id","target":"event_id","type":"int64"}]}`))
		if !errors.Is(err, database.ErrTransformPlanInvalid) {
			t.Fatalf("ParseTransformPlanV1(duplicate target) error = %T %v, want ErrTransformPlanInvalid", err, err)
		}
	})
	t.Run("missing source field", func(t *testing.T) {
		plan, err := database.ParseTransformPlanV1([]byte(`{"version":1,"select":[{"source":"missing","target":"event_id","type":"int64"}]}`))
		if err != nil {
			t.Fatal(err)
		}
		if err := plan.ValidateAgainstRelation(transformPlanRelation(t)); !errors.Is(err, database.ErrTransformPlanInvalid) {
			t.Fatalf("ValidateAgainstRelation(missing field) error = %T %v, want ErrTransformPlanInvalid", err, err)
		}
	})
}

func transformPlanRelation(t *testing.T) database.Relation {
	t.Helper()
	catalog, err := database.CatalogFromConnectorCatalog(connectors.Catalog{
		Connector: "transform_fixture",
		Streams: []connectors.Stream{{
			Name:   "events",
			Schema: []byte(`{"type":"object","required":["id","updated_at","amount","status"],"properties":{"id":{"type":"integer"},"updated_at":{"type":"string","format":"date-time"},"amount":{"type":"integer"},"status":{"type":"string"}}}`),
		}},
	}, "events")
	if err != nil {
		t.Fatal(err)
	}
	return catalog.Relations()[0]
}
