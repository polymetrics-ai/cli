package database_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/synccontract"
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

func TestMappingContractV1MapsOnlyDeclaredTombstoneKeys(t *testing.T) {
	stringType, err := database.NewString(64, "")
	if err != nil {
		t.Fatal(err)
	}
	integerType, err := database.NewSignedInteger(64)
	if err != nil {
		t.Fatal(err)
	}
	stringPlan, err := database.CompileTypePlan(stringType, stringType)
	if err != nil {
		t.Fatal(err)
	}
	integerPlan, err := database.CompileTypePlan(integerType, integerType)
	if err != nil {
		t.Fatal(err)
	}
	mapping, err := database.NewMappingContractV1([]database.MappingColumnV1{
		{Source: "source_tenant", Target: "tenant", Type: stringPlan},
		{Source: "source_id", Target: "id", Type: integerPlan},
		{Source: "source_value", Target: "value", Type: stringPlan},
	})
	if err != nil {
		t.Fatal(err)
	}
	source := synccontract.Tombstone{
		Operation:   synccontract.OperationDelete,
		EventID:     synccontract.OpaqueToken("cdc-delete-tenant-9"),
		Key:         json.RawMessage(`{"source_tenant":"retain","source_id":9}`),
		DeleteImage: synccontract.DeleteImageKeyOnly,
		Position: synccontract.CheckpointPosition{
			Primary:    synccontract.OpaqueToken("0/16B6C50"),
			TieBreaker: synccontract.OpaqueToken("cdc-delete-tenant-9"),
		},
	}
	mapped, err := mapping.MapTombstone(source, []string{"source_tenant", "source_id"})
	if err != nil {
		t.Fatalf("MapTombstone() error = %v", err)
	}
	var key map[string]json.RawMessage
	if err := json.Unmarshal(mapped.Key, &key); err != nil {
		t.Fatalf("mapped tombstone key = %q, want JSON: %v", mapped.Key, err)
	}
	if got, want := string(key["tenant"]), `"retain"`; got != want {
		t.Fatalf("mapped tenant key = %s, want %s", got, want)
	}
	if got, want := string(key["id"]), `9`; got != want {
		t.Fatalf("mapped id key = %s, want %s", got, want)
	}
	if len(key) != 2 || key["source_tenant"] != nil || key["source_id"] != nil || key["value"] != nil {
		t.Fatalf("mapped tombstone keys = %#v, want only tenant and id", key)
	}
	if string(mapped.EventID) != string(source.EventID) || !reflect.DeepEqual(mapped.Position, source.Position) {
		t.Fatalf("mapped tombstone metadata = %#v, want source identity and position preserved", mapped)
	}

	for _, invalid := range []synccontract.Tombstone{
		func() synccontract.Tombstone {
			copy := source.Clone()
			copy.Key = json.RawMessage(`{"source_tenant":"retain"}`)
			return copy
		}(),
		func() synccontract.Tombstone {
			copy := source.Clone()
			copy.Key = json.RawMessage(`{"source_tenant":"retain","source_id":9,"unexpected":true}`)
			return copy
		}(),
	} {
		if got, err := mapping.MapTombstone(invalid, []string{"source_tenant", "source_id"}); err == nil || got.Key != nil {
			t.Fatalf("MapTombstone(%s) = (%#v, %v), want empty projection and refusal", invalid.Key, got, err)
		}
	}
}

func TestDatabaseWritePlanSealsMappingBeforeSessionMutation(t *testing.T) {
	definition := testDatabaseWriteDefinition(t)
	control := testDatabaseWriteControl(t, "orders", "stream-orders", 1)
	baseMapping := testDatabaseWriteMapping(t, "source_id", "target_id")
	changedMapping := testDatabaseWriteMapping(t, "alternate_id", "target_id")
	basePlan, err := database.NewDatabaseWritePlan(context.Background(), database.DatabaseWritePlanRequest{
		Definition:  definition,
		Control:     control,
		Mode:        synccontract.ModeFullAppend,
		Strategy:    connectors.ApplyStrategyAppend,
		Mapping:     baseMapping,
		RecordCount: 1,
		BatchSize:   1,
	})
	if err != nil {
		t.Fatalf("NewDatabaseWritePlan(base) error = %v", err)
	}
	changedPlan, err := database.NewDatabaseWritePlan(context.Background(), database.DatabaseWritePlanRequest{
		Definition:  definition,
		Control:     control,
		Mode:        synccontract.ModeFullAppend,
		Strategy:    connectors.ApplyStrategyAppend,
		Mapping:     changedMapping,
		RecordCount: 1,
		BatchSize:   1,
	})
	if err != nil {
		t.Fatalf("NewDatabaseWritePlan(changed mapping) error = %v", err)
	}
	if got := basePlan.Mapping().Columns(); len(got) != 1 || got[0].Target != "target_id" || !got[0].Type.Target().Equal(baseMapping.Columns()[0].Type.Target()) {
		t.Fatalf("plan mapping = %#v, want target column/type projection", got)
	}

	driver := &databaseWriteDriverFake{atomicOverwrite: true}
	executor := testDatabaseWriteExecutor(t, driver)
	approval := testDatabaseWriteApproval(t, executor, basePlan)
	if _, err := executor.Execute(context.Background(), changedPlan, approval, []connectors.Record{{"alternate_id": int64(42)}}); !errors.Is(err, database.ErrDatabaseWriteApprovalInvalid) {
		t.Fatalf("Execute(changed mapping) error = %v, want approval mismatch", err)
	}
	if driver.beginCalls != 0 || driver.batchCalls != 0 || driver.commitCalls != 0 || driver.rollbackCalls != 0 {
		t.Fatalf("changed mapping mutated session: begin/batch/commit/rollback = %d/%d/%d/%d", driver.beginCalls, driver.batchCalls, driver.commitCalls, driver.rollbackCalls)
	}
}

func TestDatabaseWriteExecutorDeletesOnlyExplicitTombstones(t *testing.T) {
	definition := testDatabaseWriteDefinition(t)
	control := testDatabaseWriteControl(t, "orders", "stream-orders", 1)
	mapping := testDatabaseWriteMapping(t, "source_id", "target_id")
	driver := &databaseWriteDriverFake{targetRows: map[string]bool{"delete-me": true}}
	executor := testDatabaseWriteExecutor(t, driver)

	ordinaryPlan, err := database.NewDatabaseWritePlan(context.Background(), database.DatabaseWritePlanRequest{
		Definition:     definition,
		Control:        control,
		Mode:           synccontract.ModeIncrementalUpsert,
		Strategy:       connectors.ApplyStrategyMerge,
		Mapping:        mapping,
		Keys:           []string{"target_id"},
		RecordCount:    1,
		TombstoneCount: 0,
		BatchSize:      2,
		Destructive:    false,
	})
	if err != nil {
		t.Fatalf("NewDatabaseWritePlan(ordinary) error = %v", err)
	}
	ordinaryInput, err := database.NewDatabaseWriteInput([]connectors.Record{{"source_id": int64(1)}}, database.TombstoneEnvelope{})
	if err != nil {
		t.Fatalf("NewDatabaseWriteInput(ordinary) error = %v", err)
	}
	if _, err := executor.ExecuteInput(context.Background(), ordinaryPlan, testDatabaseWriteApproval(t, executor, ordinaryPlan), ordinaryInput); err != nil {
		t.Fatalf("ExecuteInput(ordinary) error = %v", err)
	}
	if !driver.targetRows["delete-me"] {
		t.Fatal("row absent from an ordinary batch was deleted")
	}

	tombstone := synccontract.Tombstone{
		Operation:   synccontract.OperationDelete,
		EventID:     synccontract.OpaqueToken("delete-me-event"),
		Key:         json.RawMessage(`{"target_id":"delete-me"}`),
		DeleteImage: synccontract.DeleteImageKeyOnly,
		Position: synccontract.CheckpointPosition{
			Primary:    synccontract.OpaqueToken("00000001"),
			TieBreaker: synccontract.OpaqueToken("00000001"),
		},
	}
	envelope, err := database.NewTombstoneEnvelope([]synccontract.Tombstone{tombstone})
	if err != nil {
		t.Fatalf("NewTombstoneEnvelope() error = %v", err)
	}
	tombstonePlan, err := database.NewDatabaseWritePlan(context.Background(), database.DatabaseWritePlanRequest{
		Definition:     definition,
		Control:        control,
		Mode:           synccontract.ModeIncrementalUpsert,
		Strategy:       connectors.ApplyStrategyMerge,
		Mapping:        mapping,
		Keys:           []string{"target_id"},
		RecordCount:    0,
		TombstoneCount: 1,
		BatchSize:      2,
		Destructive:    false,
	})
	if err != nil {
		t.Fatalf("NewDatabaseWritePlan(tombstone) error = %v", err)
	}
	tombstoneInput, err := database.NewDatabaseWriteInput(nil, envelope)
	if err != nil {
		t.Fatalf("NewDatabaseWriteInput(tombstone) error = %v", err)
	}
	if _, err := executor.ExecuteInput(context.Background(), tombstonePlan, testDatabaseWriteApproval(t, executor, tombstonePlan), tombstoneInput); err != nil {
		t.Fatalf("ExecuteInput(tombstone) error = %v", err)
	}
	if driver.targetRows["delete-me"] {
		t.Fatal("explicit tombstone did not delete its seeded fake row")
	}
}

func TestDatabaseWriteExecutorRefusesTombstoneMismatchBeforeSessionMutation(t *testing.T) {
	definition := testDatabaseWriteDefinition(t)
	control := testDatabaseWriteControl(t, "orders", "stream-orders", 1)
	plan, err := database.NewDatabaseWritePlan(context.Background(), database.DatabaseWritePlanRequest{
		Definition:     definition,
		Control:        control,
		Mode:           synccontract.ModeIncrementalUpsert,
		Strategy:       connectors.ApplyStrategyMerge,
		Mapping:        testDatabaseWriteMapping(t, "source_id", "target_id"),
		Keys:           []string{"target_id"},
		RecordCount:    1,
		TombstoneCount: 1,
		BatchSize:      2,
		Destructive:    false,
	})
	if err != nil {
		t.Fatalf("NewDatabaseWritePlan() error = %v", err)
	}
	ordinaryInput, err := database.NewDatabaseWriteInput([]connectors.Record{{"source_id": int64(1)}}, database.TombstoneEnvelope{})
	if err != nil {
		t.Fatal(err)
	}
	driver := &databaseWriteDriverFake{atomicOverwrite: true, targetRows: map[string]bool{"retained": true}}
	executor := testDatabaseWriteExecutor(t, driver)
	approval := testDatabaseWriteApproval(t, executor, plan)
	if _, err := executor.ExecuteInput(context.Background(), plan, approval, ordinaryInput); !errors.Is(err, database.ErrDatabaseWritePlanInvalid) {
		t.Fatalf("ExecuteInput(mismatched tombstones) error = %v, want plan refusal", err)
	}
	if approval.Consumed() || driver.beginCalls != 0 || driver.batchCalls != 0 || driver.commitCalls != 0 || driver.rollbackCalls != 0 || !driver.targetRows["retained"] {
		t.Fatalf("mismatched tombstones mutated state: approval/begin/batch/commit/rollback/row = %t/%d/%d/%d/%d/%t", approval.Consumed(), driver.beginCalls, driver.batchCalls, driver.commitCalls, driver.rollbackCalls, driver.targetRows["retained"])
	}

	invalid := synccontract.Tombstone{Operation: synccontract.OperationTruncate, EventID: synccontract.OpaqueToken("not-a-row-delete")}
	if envelope, err := database.NewTombstoneEnvelope([]synccontract.Tombstone{invalid}); !errors.Is(err, database.ErrTombstoneEnvelopeInvalid) || envelope.Count() != 0 {
		t.Fatalf("NewTombstoneEnvelope(non-delete) = (%#v, %v), want empty envelope and explicit refusal", envelope, err)
	}
}
