package database_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/warehouse"
)

func TestDriverAdmissionRequiresRegisteredCompatibleNativeAdmission(t *testing.T) {
	definition := loadTestDefinition(t, definitionWithAdmittedMode(validDefinitionJSON))
	// This fixture is admission-only; it is not a published PostgreSQL command.
	contract := nativeContract("postgres-wire", "fixture-postgres-admission", "fixture-postgres-admission-v1")
	inbound := testInboundCommand(t, contract)

	registry, err := database.NewDriverRegistry(declaredDriver{descriptor: postgresDriverDescriptor()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := registry.Admit(context.Background(), definition, inbound); err == nil {
		t.Fatal("driver declaration alone passed admission without a native admission object")
	}

	inboundAdmission := testInboundNativeAdmission(t, contract)
	admitted := admittedDriver{
		declaredDriver: declaredDriver{descriptor: postgresDriverDescriptor()},
		admissions:     []database.DatabaseNativeAdmission{inboundAdmission},
	}
	registry, err = database.NewDriverRegistry(admitted)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := registry.Admit(context.Background(), definition, inbound); err != nil {
		t.Fatalf("registered compatible native admission rejected: %v", err)
	}

	wrongVersion := loadTestDefinition(t, strings.Replace(definitionWithAdmittedMode(validDefinitionJSON), `"api_version": 1`, `"api_version": 2`, 1))
	if _, err := registry.Admit(context.Background(), wrongVersion, inbound); err == nil {
		t.Fatal("incompatible driver API version passed admission")
	}

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := registry.Admit(cancelled, definition, inbound); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled admission error = %v, want context.Canceled", err)
	}
}

func TestDatabaseNativeAdmissionIsBoundToOneWarehouseLeg(t *testing.T) {
	definition := loadTestDefinition(t, definitionWithAdmittedMode(validDefinitionJSON))
	contract := nativeContract("postgres-wire", "fixture-postgres-warehouse-leg", "fixture-postgres-warehouse-leg-v1")
	inbound := testInboundCommand(t, contract)
	outbound := testOutboundCommand(t, contract)
	registry, err := database.NewDriverRegistry(admittedDriver{
		declaredDriver: declaredDriver{descriptor: postgresDriverDescriptor()},
		admissions:     []database.DatabaseNativeAdmission{testInboundNativeAdmission(t, contract)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := registry.Admit(context.Background(), definition, inbound); err != nil {
		t.Fatalf("inbound admission error = %v", err)
	}
	if _, err := registry.Admit(context.Background(), definition, outbound); !errors.Is(err, database.ErrNativeDriverAdmissionMismatch) {
		t.Fatalf("outbound admission through an inbound record error = %v, want ErrNativeDriverAdmissionMismatch", err)
	}
}

func TestDriverRegistryRejectsSharedNativeContractAcrossWarehouseLegs(t *testing.T) {
	definition := loadTestDefinition(t, definitionWithAdmittedMode(validDefinitionJSON))
	contract := nativeContract("postgres-wire", "fixture-shared-warehouse-leg", "fixture-shared-warehouse-leg-v1")
	shared := nativeAdmissionFor(contract)
	inboundAdmission, err := database.NewDatabaseInboundAdmission(shared)
	if err != nil {
		t.Fatal(err)
	}
	outboundAdmission, err := database.NewDatabaseOutboundAdmission(shared)
	if err != nil {
		t.Fatal(err)
	}
	registry, err := database.NewDriverRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.Register(admittedDriver{
		declaredDriver: declaredDriver{descriptor: postgresDriverDescriptor()},
		admissions:     []database.DatabaseNativeAdmission{inboundAdmission, outboundAdmission},
	}); !errors.Is(err, database.ErrNativeDriverAdmissionLegConflict) {
		t.Fatalf("Register(shared cross-leg admission) error = %v, want ErrNativeDriverAdmissionLegConflict", err)
	}

	inbound := testInboundCommand(t, contract)
	if _, err := registry.Admit(context.Background(), definition, inbound); !errors.Is(err, database.ErrDriverNotRegistered) {
		t.Fatalf("inbound admission after rejected registration error = %v, want ErrDriverNotRegistered", err)
	}
	outbound := testOutboundCommand(t, contract)
	if _, err := registry.Admit(context.Background(), definition, outbound); !errors.Is(err, database.ErrDriverNotRegistered) {
		t.Fatalf("outbound admission after rejected registration error = %v, want ErrDriverNotRegistered", err)
	}
}

func TestDriverRegistryRejectsCrossDriverNativeContractReuse(t *testing.T) {
	definition := loadTestDefinition(t, definitionWithAdmittedMode(validDefinitionJSON))
	secondDefinition := loadTestDefinition(t, strings.Replace(
		definitionWithAdmittedMode(validDefinitionJSON),
		`"id": "postgres"`,
		`"id": "second-driver"`,
		1,
	))
	contract := nativeContract("postgres-wire", "fixture-cross-driver-warehouse-leg", "fixture-cross-driver-warehouse-leg-v1")
	inbound := testInboundCommand(t, contract)
	outbound := testOutboundCommand(t, contract)
	registry, err := database.NewDriverRegistry(admittedDriver{
		declaredDriver: declaredDriver{descriptor: postgresDriverDescriptor()},
		admissions:     []database.DatabaseNativeAdmission{testInboundNativeAdmission(t, contract)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := registry.Admit(context.Background(), definition, inbound); err != nil {
		t.Fatalf("first-driver inbound admission error = %v", err)
	}
	if err := registry.Register(admittedDriver{
		declaredDriver: declaredDriver{descriptor: database.DriverDescriptor{ID: "second-driver", Protocol: "postgres-wire", APIVersion: 1}},
		admissions:     []database.DatabaseNativeAdmission{testOutboundNativeAdmission(t, contract)},
	}); !errors.Is(err, database.ErrNativeDriverAdmissionLegConflict) {
		t.Fatalf("Register(cross-driver shared admission) error = %v, want ErrNativeDriverAdmissionLegConflict", err)
	}
	if _, err := registry.Admit(context.Background(), secondDefinition, outbound); !errors.Is(err, database.ErrDriverNotRegistered) {
		t.Fatalf("second-driver outbound admission after rejected registration error = %v, want ErrDriverNotRegistered", err)
	}

	distinctContract := nativeContract("postgres-wire", "fixture-distinct-warehouse-leg", "fixture-distinct-warehouse-leg-v1")
	if err := registry.Register(admittedDriver{
		declaredDriver: declaredDriver{descriptor: database.DriverDescriptor{ID: "distinct-driver", Protocol: "postgres-wire", APIVersion: 1}},
		admissions:     []database.DatabaseNativeAdmission{testOutboundNativeAdmission(t, distinctContract)},
	}); err != nil {
		t.Fatalf("Register(distinct cross-driver admission) error = %v", err)
	}
}

func TestWarehouseMediationUsesSharedArtifactAndSeparateDatabaseLegs(t *testing.T) {
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
	source, err := database.NewSourceRef(sourceIdentity, relation)
	if err != nil {
		t.Fatal(err)
	}
	target, err := database.NewTargetRef(targetIdentity, relation)
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := warehouse.NewArtifactRef(sourceIdentity, "widgets")
	if err != nil {
		t.Fatal(err)
	}
	inboundRef, err := database.NewWarehouseInboundRef(source, artifact)
	if err != nil {
		t.Fatal(err)
	}
	outboundRef, err := database.NewWarehouseOutboundRef(artifact, target)
	if err != nil {
		t.Fatal(err)
	}
	if got := inboundRef.Warehouse().Identity(); !got.SameIdentity(sourceIdentity) {
		t.Fatalf("warehouse landing owner = %#v, want source identity %#v", got, sourceIdentity)
	}
	if got := outboundRef.Warehouse().Identity(); !got.SameIdentity(sourceIdentity) {
		t.Fatalf("warehouse delivery owner = %#v, want source identity %#v", got, sourceIdentity)
	}
	if got := outboundRef.Target().Identity(); !got.SameIdentity(targetIdentity) {
		t.Fatalf("warehouse delivery target = %#v, want target identity %#v", got, targetIdentity)
	}

	wrongArtifact, err := warehouse.NewArtifactRef(targetIdentity, "widgets")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.NewWarehouseInboundRef(source, wrongArtifact); err == nil {
		t.Fatal("NewWarehouseInboundRef() accepted a warehouse owned by another source connection")
	}
	if _, err := warehouse.NewArtifactRef(sourceIdentity, ".."); err == nil {
		t.Fatal("NewArtifactRef() accepted an unsafe warehouse table component")
	}

	inboundContract := nativeContract("postgres-wire", "fixture-postgres-source-to-warehouse", "fixture-postgres-source-to-warehouse-v1")
	inbound, err := database.NewDatabaseInboundCommand(inboundRef, inboundContract)
	if err != nil {
		t.Fatal(err)
	}
	outboundContract := nativeContract("postgres-wire", "fixture-postgres-warehouse-to-target", "fixture-postgres-warehouse-to-target-v1")
	outbound, err := database.NewDatabaseOutboundCommand(outboundRef, outboundContract)
	if err != nil {
		t.Fatal(err)
	}

	definition := loadTestDefinition(t, strings.Replace(
		validDefinitionJSON,
		`"admitted_modes": []`,
		`"admitted_modes": ["full_append"]`,
		1,
	))
	inboundAdmission := testInboundNativeAdmission(t, inboundContract)
	outboundAdmission := testOutboundNativeAdmission(t, outboundContract)
	admitted := admittedDriver{
		declaredDriver: declaredDriver{descriptor: postgresDriverDescriptor()},
		admissions: []database.DatabaseNativeAdmission{
			inboundAdmission,
			outboundAdmission,
		},
	}
	registry, err := database.NewDriverRegistry(admitted)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := registry.Admit(context.Background(), definition, inbound); err != nil {
		t.Fatalf("warehouse inbound admission error = %v", err)
	}
	if _, err := registry.Admit(context.Background(), definition, outbound); err != nil {
		t.Fatalf("warehouse outbound admission error = %v", err)
	}
	inboundOnly, err := database.NewDriverRegistry(admittedDriver{
		declaredDriver: declaredDriver{descriptor: postgresDriverDescriptor()},
		admissions:     []database.DatabaseNativeAdmission{inboundAdmission},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := inboundOnly.Admit(context.Background(), definition, outbound); !errors.Is(err, database.ErrNativeDriverAdmissionMismatch) {
		t.Fatalf("outbound admission through an inbound descriptor error = %v, want ErrNativeDriverAdmissionMismatch", err)
	}
	mutatedContract := inbound.Contract()
	mutatedContract.Modes[0] = synccontract.ModeFullOverwrite
	if got := inbound.Contract().Modes; len(got) != 1 || got[0] != synccontract.ModeFullAppend {
		t.Fatalf("DatabaseInboundCommand.Contract() = %v after caller mutation, want independent full_append projection", got)
	}
	if _, err := registry.Admit(context.Background(), loadTestDefinition(t, validDefinitionJSON), inbound); !errors.Is(err, database.ErrDriverModeNotDeclared) {
		t.Fatalf("empty declared modes admission error = %v, want ErrDriverModeNotDeclared", err)
	}
}

func TestMySQLLayerTwoReferenceCompilesAgainstSharedWarehouseArtifact(t *testing.T) {
	var _ database.Driver = mysqlLayerTwo{}
	var _ database.NativeAdmittedDriver = mysqlLayerTwo{}

	sourceIdentity := database.ConnectionIdentity{
		WorkspaceID:  "workspace-1",
		ConnectorID:  "mysql",
		ConnectionID: "mysql-source-connection",
	}
	targetIdentity := database.ConnectionIdentity{
		WorkspaceID:  "workspace-1",
		ConnectorID:  "mysql",
		ConnectionID: "mysql-target-connection",
	}
	relation := database.RelationRef{
		Schema: database.SchemaRef{
			Catalog: database.CatalogRef{Name: "analytics"},
			Name:    "public",
		},
		Name: "widgets",
	}
	source, err := database.NewSourceRef(sourceIdentity, relation)
	if err != nil {
		t.Fatal(err)
	}
	target, err := database.NewTargetRef(targetIdentity, relation)
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := warehouse.NewArtifactRef(sourceIdentity, "widgets")
	if err != nil {
		t.Fatal(err)
	}
	inboundRef, err := database.NewWarehouseInboundRef(source, artifact)
	if err != nil {
		t.Fatal(err)
	}
	outboundRef, err := database.NewWarehouseOutboundRef(artifact, target)
	if err != nil {
		t.Fatal(err)
	}
	if got := inboundRef.Warehouse().Identity(); !got.SameIdentity(sourceIdentity) {
		t.Fatalf("MySQL inbound warehouse identity = %#v, want %#v", got, sourceIdentity)
	}
	if got := outboundRef.Warehouse().Identity(); !got.SameIdentity(sourceIdentity) {
		t.Fatalf("MySQL outbound warehouse identity = %#v, want %#v", got, sourceIdentity)
	}
	if got := outboundRef.Target().Identity(); !got.SameIdentity(targetIdentity) {
		t.Fatalf("MySQL outbound target identity = %#v, want %#v", got, targetIdentity)
	}

	inboundContract := nativeContract("mysql-wire", "fixture-mysql-source-to-warehouse", "fixture-mysql-source-to-warehouse-v1")
	inbound, err := database.NewDatabaseInboundCommand(inboundRef, inboundContract)
	if err != nil {
		t.Fatal(err)
	}
	outboundContract := nativeContract("mysql-wire", "fixture-mysql-warehouse-to-target", "fixture-mysql-warehouse-to-target-v1")
	outbound, err := database.NewDatabaseOutboundCommand(outboundRef, outboundContract)
	if err != nil {
		t.Fatal(err)
	}
	mysql := mysqlLayerTwo{
		admissions: []database.DatabaseNativeAdmission{
			testInboundNativeAdmission(t, inboundContract),
			testOutboundNativeAdmission(t, outboundContract),
		},
	}
	registry, err := database.NewDriverRegistry(mysql)
	if err != nil {
		t.Fatal(err)
	}
	definition := loadTestDefinition(t, definitionWithAdmittedMode(mysqlDefinitionJSON))
	if _, err := registry.Admit(context.Background(), definition, inbound); err != nil {
		t.Fatalf("MySQL inbound admission error = %v", err)
	}
	if _, err := registry.Admit(context.Background(), definition, outbound); err != nil {
		t.Fatalf("MySQL outbound admission error = %v", err)
	}
}
