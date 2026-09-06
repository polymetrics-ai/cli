package database_test

import (
	"context"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/warehouse"
)

// mysqlLayerTwo is a test-only layer-two proof. It declares no production
// connector capability or executable MySQL operation.
type mysqlLayerTwo struct {
	admissions []database.DatabaseNativeAdmission
}

func (mysqlLayerTwo) DatabaseDriverDescriptor() database.DriverDescriptor {
	return database.DriverDescriptor{ID: "mysql", Protocol: "mysql-wire", APIVersion: 1}
}

func (d mysqlLayerTwo) DatabaseNativeAdmissions() []database.DatabaseNativeAdmission {
	return cloneDatabaseNativeAdmissions(d.admissions)
}

func loadTestDefinition(t *testing.T, document string) database.Definition {
	t.Helper()
	definition, err := database.Load(context.Background(), fstest.MapFS{
		"database.json": &fstest.MapFile{Data: []byte(document)},
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	return definition
}

func testReadPlanRequest(t *testing.T) database.ReadPlanRequest {
	t.Helper()
	definition := loadTestDefinition(t, validDefinitionJSON)
	identity := database.ConnectionIdentity{
		WorkspaceID:  "workspace-1",
		ConnectorID:  "postgres",
		ConnectionID: "connection-1",
	}
	relation := database.RelationRef{
		Schema: database.SchemaRef{
			Catalog: database.CatalogRef{Name: "analytics"},
			Name:    "public",
		},
		Name: "widgets",
	}
	idColumn := database.ColumnRef{Relation: relation, Name: "id"}
	integer, err := database.NewSignedInteger(32)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := database.NewCatalog(database.CatalogRef{Name: "analytics"}, []database.Relation{{
		Ref: relation,
		Columns: []database.Column{{
			Ref:     idColumn,
			Type:    integer,
			Ordinal: 1,
		}},
		Keys: []database.Key{{
			Name:    "widgets_pk",
			Kind:    database.KeyPrimary,
			Columns: []database.ColumnRef{idColumn},
		}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	source, err := database.NewSourceRef(identity, relation)
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := warehouse.NewArtifactRef(identity, "widgets")
	if err != nil {
		t.Fatal(err)
	}
	inbound, err := database.NewWarehouseInboundRef(source, artifact)
	if err != nil {
		t.Fatal(err)
	}
	return database.ReadPlanRequest{
		Inbound:    inbound,
		Definition: definition,
		Catalog:    catalog,
		Relation:   relation,
		Columns:    []database.ColumnRef{idColumn},
		Order:      []database.OrderTerm{{Column: idColumn, Direction: database.SortAscending}},
	}
}

func definitionWithAdmittedMode(document string) string {
	return strings.Replace(document, `"admitted_modes": []`, `"admitted_modes": ["full_append"]`, 1)
}

func testInboundCommand(t *testing.T, contract synccontract.NativeCommandContract) database.DatabaseInboundCommand {
	t.Helper()
	identity := database.ConnectionIdentity{
		WorkspaceID:  "workspace-1",
		ConnectorID:  "postgres",
		ConnectionID: "source-connection",
	}
	relation := database.RelationRef{
		Schema: database.SchemaRef{
			Catalog: database.CatalogRef{Name: "analytics"},
			Name:    "public",
		},
		Name: "widgets",
	}
	source, err := database.NewSourceRef(identity, relation)
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := warehouse.NewArtifactRef(identity, "widgets")
	if err != nil {
		t.Fatal(err)
	}
	inbound, err := database.NewWarehouseInboundRef(source, artifact)
	if err != nil {
		t.Fatal(err)
	}
	command, err := database.NewDatabaseInboundCommand(inbound, contract)
	if err != nil {
		t.Fatal(err)
	}
	return command
}

func testOutboundCommand(t *testing.T, contract synccontract.NativeCommandContract) database.DatabaseOutboundCommand {
	t.Helper()
	sourceIdentity := database.ConnectionIdentity{
		WorkspaceID:  "workspace-1",
		ConnectorID:  "postgres",
		ConnectionID: "source-connection",
	}
	targetIdentity := database.ConnectionIdentity{
		WorkspaceID:  "workspace-1",
		ConnectorID:  "postgres",
		ConnectionID: "target-connection",
	}
	relation := database.RelationRef{
		Schema: database.SchemaRef{
			Catalog: database.CatalogRef{Name: "analytics"},
			Name:    "public",
		},
		Name: "widgets",
	}
	artifact, err := warehouse.NewArtifactRef(sourceIdentity, "widgets")
	if err != nil {
		t.Fatal(err)
	}
	target, err := database.NewTargetRef(targetIdentity, relation)
	if err != nil {
		t.Fatal(err)
	}
	outbound, err := database.NewWarehouseOutboundRef(artifact, target)
	if err != nil {
		t.Fatal(err)
	}
	command, err := database.NewDatabaseOutboundCommand(outbound, contract)
	if err != nil {
		t.Fatal(err)
	}
	return command
}

type declaredDriver struct {
	descriptor database.DriverDescriptor
}

func (d declaredDriver) DatabaseDriverDescriptor() database.DriverDescriptor { return d.descriptor }

type admittedDriver struct {
	declaredDriver
	admissions []database.DatabaseNativeAdmission
}

func (d admittedDriver) DatabaseNativeAdmissions() []database.DatabaseNativeAdmission {
	return cloneDatabaseNativeAdmissions(d.admissions)
}

type nativeAdmission struct {
	native synccontract.NativeSyncExecutorDescriptor
}

func nativeContract(protocol, command, executorID string) synccontract.NativeCommandContract {
	return synccontract.NativeCommandContract{
		ContractVersion: synccontract.NativeCommandContractVersion,
		Protocol:        protocol,
		Command:         command,
		Executor:        synccontract.ExecutorReference{Kind: "native", ID: executorID},
		Modes:           []synccontract.Mode{synccontract.ModeFullAppend},
	}
}

func nativeAdmissionFor(contract synccontract.NativeCommandContract) nativeAdmission {
	return nativeAdmission{
		native: synccontract.NativeSyncExecutorDescriptor{
			Protocol: contract.Protocol,
			Command:  contract.Command,
			Executor: contract.Executor,
			Modes:    append([]synccontract.Mode(nil), contract.Modes...),
		},
	}
}

func testInboundNativeAdmission(t *testing.T, contract synccontract.NativeCommandContract) database.DatabaseNativeAdmission {
	t.Helper()
	admission, err := database.NewDatabaseInboundAdmission(nativeAdmissionFor(contract))
	if err != nil {
		t.Fatal(err)
	}
	return admission
}

func testOutboundNativeAdmission(t *testing.T, contract synccontract.NativeCommandContract) database.DatabaseNativeAdmission {
	t.Helper()
	admission, err := database.NewDatabaseOutboundAdmission(nativeAdmissionFor(contract))
	if err != nil {
		t.Fatal(err)
	}
	return admission
}

func (d nativeAdmission) NativeSyncExecutorDescriptor() synccontract.NativeSyncExecutorDescriptor {
	return d.native
}

func cloneDatabaseNativeAdmissions(admissions []database.DatabaseNativeAdmission) []database.DatabaseNativeAdmission {
	return append([]database.DatabaseNativeAdmission(nil), admissions...)
}

type cancelOnErrCallContext struct {
	cancelAt int
	errCalls int
	done     chan struct{}
}

func (c *cancelOnErrCallContext) Deadline() (time.Time, bool) { return time.Time{}, false }

func (c *cancelOnErrCallContext) Done() <-chan struct{} { return c.done }

func (c *cancelOnErrCallContext) Err() error {
	c.errCalls++
	if c.errCalls < c.cancelAt {
		return nil
	}
	select {
	case <-c.done:
	default:
		close(c.done)
	}
	return context.Canceled
}

func (*cancelOnErrCallContext) Value(any) any { return nil }

func postgresDriverDescriptor() database.DriverDescriptor {
	return database.DriverDescriptor{ID: "postgres", Protocol: "postgres-wire", APIVersion: 1}
}

// mysqlDefinitionJSON is a test-only strict-definition fixture. It is not a
// production MySQL manifest or capability declaration.
const mysqlDefinitionJSON = `{
  "schema_version": 1,
  "driver": {"id": "mysql", "protocol": "mysql-wire", "api_version": 1},
  "catalog": {"qualification_order": ["schema", "relation"]},
  "identifiers": {"quote_style": "double_quote", "case_fold": "lower", "max_bytes": 63},
  "resources": {
    "read_page": {"default": 100, "maximum": 1000},
    "write_batch": {"default": 25, "maximum": 250},
    "pool": {"default": 2, "maximum": 8},
    "connect_timeout_ms": 1000,
    "operation_timeout_ms": 5000,
    "max_parameters": 1000
  },
  "type_mappings": [
    {"native": {"name": "int"}, "logical": {"kind": "signed_integer", "bits": 32}}
  ],
  "admitted_modes": []
}`

const validDefinitionJSON = `{
  "schema_version": 1,
  "driver": {"id": "postgres", "protocol": "postgres-wire", "api_version": 1},
  "catalog": {"qualification_order": ["schema", "relation"]},
  "identifiers": {"quote_style": "double_quote", "case_fold": "lower", "max_bytes": 63},
  "resources": {
    "read_page": {"default": 100, "maximum": 1000},
    "write_batch": {"default": 25, "maximum": 250},
    "pool": {"default": 2, "maximum": 8},
    "connect_timeout_ms": 1000,
    "operation_timeout_ms": 5000,
    "max_parameters": 1000
  },
  "type_mappings": [
    {"native": {"name": "int4"}, "logical": {"kind": "signed_integer", "bits": 32}}
  ],
  "admitted_modes": []
}`
