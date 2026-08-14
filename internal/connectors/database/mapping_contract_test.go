package database_test

import (
	"reflect"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
)

func TestMappingContractV1ProjectsLosslessValuesAndRoundTrips(t *testing.T) {
	sourceType, err := database.NewSignedInteger(32)
	if err != nil {
		t.Fatal(err)
	}
	targetType, err := database.NewSignedInteger(64)
	if err != nil {
		t.Fatal(err)
	}
	typePlan, err := database.CompileTypePlan(sourceType, targetType)
	if err != nil {
		t.Fatal(err)
	}
	mapping, err := database.NewMappingContractV1([]database.MappingColumnV1{{
		Source: "source_id",
		Target: "target_id",
		Type:   typePlan,
	}})
	if err != nil {
		t.Fatalf("NewMappingContractV1() error = %v", err)
	}

	if got := mapping.Version(); got != database.MappingContractVersionV1 {
		t.Fatalf("Version() = %d, want %d", got, database.MappingContractVersionV1)
	}
	mapped, err := mapping.MapRecord(connectors.Record{"source_id": int32(42)})
	if err != nil {
		t.Fatalf("MapRecord() error = %v", err)
	}
	if got, want := mapped, (connectors.Record{"target_id": int64(42)}); !reflect.DeepEqual(got, want) {
		t.Fatalf("MapRecord() = %#v, want %#v", got, want)
	}
	roundTripped, err := mapping.UnmapRecord(mapped)
	if err != nil {
		t.Fatalf("UnmapRecord() error = %v", err)
	}
	if got, want := roundTripped, (connectors.Record{"source_id": int32(42)}); !reflect.DeepEqual(got, want) {
		t.Fatalf("UnmapRecord() = %#v, want %#v", got, want)
	}
}

func TestMappingContractV1RefusesUnrepresentableMappingsAndValues(t *testing.T) {
	int32Type, err := database.NewSignedInteger(32)
	if err != nil {
		t.Fatal(err)
	}
	int64Type, err := database.NewSignedInteger(64)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.CompileTypePlan(int64Type, int32Type); err == nil {
		t.Fatal("CompileTypePlan(int64, int32) error = nil, want narrowing refusal")
	}
	typePlan, err := database.CompileTypePlan(int32Type, int64Type)
	if err != nil {
		t.Fatal(err)
	}
	mapping, err := database.NewMappingContractV1([]database.MappingColumnV1{{
		Source: "source_id",
		Target: "target_id",
		Type:   typePlan,
	}})
	if err != nil {
		t.Fatal(err)
	}
	if mapped, err := mapping.MapRecord(connectors.Record{"source_id": int64(42)}); err == nil || mapped != nil {
		t.Fatalf("MapRecord(unrepresentable value) = (%#v, %v), want nil target projection and refusal", mapped, err)
	}
}
