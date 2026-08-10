package database_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/warehouse"
)

func TestDatabaseDefinitionStrictLoadAndDefensiveProjection(t *testing.T) {
	tests := []struct {
		name      string
		document  string
		wantError bool
		secret    string
	}{
		{name: "valid", document: validDefinitionJSON},
		{
			name:      "unknown field is rejected without echoing its value",
			document:  strings.Replace(validDefinitionJSON, `"schema_version": 1,`, `"schema_version": 1, "unknown": "secret-do-not-echo",`, 1),
			wantError: true,
			secret:    "secret-do-not-echo",
		},
		{
			name:      "unknown logical type is rejected",
			document:  strings.Replace(validDefinitionJSON, `"signed_integer"`, `"untrusted_type"`, 1),
			wantError: true,
		},
		{
			name:      "unsupported schema version is rejected",
			document:  strings.Replace(validDefinitionJSON, `"schema_version": 1`, `"schema_version": 2`, 1),
			wantError: true,
		},
		{
			name:      "nested unknown field is rejected",
			document:  strings.Replace(validDefinitionJSON, `"id": "postgres",`, `"id": "postgres", "unexpected": true,`, 1),
			wantError: true,
		},
		{
			name:      "explicit null is rejected",
			document:  strings.Replace(validDefinitionJSON, `"logical": {"kind": "signed_integer", "bits": 32}`, `"logical": {"kind": "boolean", "bits": null}`, 1),
			wantError: true,
		},
		{
			name: "nested opaque logical mapping is rejected",
			document: strings.Replace(
				validDefinitionJSON,
				`"logical": {"kind": "signed_integer", "bits": 32}`,
				`"logical": {"kind": "array", "element": {"kind": "opaque_native", "opaque_engine": "test-engine", "opaque_name": "unmapped"}}`,
				1,
			),
			wantError: true,
		},
		{
			name:      "unbounded page policy is rejected",
			document:  strings.Replace(validDefinitionJSON, `"default": 100`, `"default": 0`, 1),
			wantError: true,
		},
		{
			name:      "overflowing connect timeout is rejected",
			document:  strings.Replace(validDefinitionJSON, `"connect_timeout_ms": 1000`, `"connect_timeout_ms": 18446744073710`, 1),
			wantError: true,
		},
		{
			name:      "overflowing operation timeout is rejected",
			document:  strings.Replace(validDefinitionJSON, `"operation_timeout_ms": 5000`, `"operation_timeout_ms": 18446744073710`, 1),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			definition, err := database.Load(context.Background(), fstest.MapFS{
				"database.json": &fstest.MapFile{Data: []byte(tt.document)},
			})
			if tt.wantError {
				if err == nil {
					t.Fatal("Load() error = nil, want rejection")
				}
				if tt.secret != "" && strings.Contains(err.Error(), tt.secret) {
					t.Fatalf("Load() error exposed supplied value: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			schema := database.DefinitionSchema()
			schema[0] = 'x'
			if !json.Valid(database.DefinitionSchema()) {
				t.Fatal("DefinitionSchema() exposed mutable embedded schema state")
			}

			mappings := definition.TypeMappings()
			if len(mappings) == 0 {
				t.Fatal("TypeMappings() = empty, want PostgreSQL mapping")
			}
			mappings[0].Native.Name = "mutated"
			if got := definition.TypeMappings()[0].Native.Name; got == "mutated" {
				t.Fatal("TypeMappings() returned mutable internal state")
			}

			modes := definition.AdmittedModes()
			if len(modes) != 0 {
				t.Fatalf("AdmittedModes() = %v, want empty declaration", modes)
			}

			policy := definition.Resources()
			if got, err := policy.EffectivePageSize(0); err != nil || got != 100 {
				t.Fatalf("EffectivePageSize(0) = (%d, %v), want (100, nil)", got, err)
			}
			if _, err := policy.EffectivePageSize(1001); err == nil {
				t.Fatal("EffectivePageSize(max+1) = nil, want bounded refusal")
			}
		})
	}

	modeDefinition := loadTestDefinition(t, strings.Replace(
		validDefinitionJSON,
		`"admitted_modes": []`,
		`"admitted_modes": ["full_append"]`,
		1,
	))
	modes := modeDefinition.AdmittedModes()
	modes[0] = synccontract.ModeFullOverwrite
	if got := modeDefinition.AdmittedModes(); len(got) != 1 || got[0] != synccontract.ModeFullAppend {
		t.Fatalf("AdmittedModes() = %v after caller mutation, want independent full_append projection", got)
	}
}

func TestDatabaseDefinitionRejectsAmbiguousMembers(t *testing.T) {
	tests := []struct {
		name     string
		document string
		path     string
	}{
		{
			name:     "repeated strict member",
			document: strings.Replace(validDefinitionJSON, `"bits": 32`, `"bits": 32, "bits": 64`, 1),
			path:     `$.type_mappings[0].logical.bits`,
		},
		{
			name:     "case-aliased member alongside canonical member",
			document: strings.Replace(validDefinitionJSON, `"bits": 32`, `"Bits": 32, "bits": 64`, 1),
			path:     `$.type_mappings[0].logical.bits`,
		},
		{
			name:     "case-aliased root strict member",
			document: strings.Replace(validDefinitionJSON, `"schema_version": 1`, `"Schema_Version": 1`, 1),
			path:     `$.schema_version`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := database.Load(context.Background(), fstest.MapFS{
				"database.json": &fstest.MapFile{Data: []byte(tt.document)},
			})
			if err == nil {
				t.Fatal("Load() error = nil, want ambiguous member rejection")
			}
			if !errors.Is(err, database.ErrInvalidDefinition) {
				t.Fatalf("Load() error = %v, want ErrInvalidDefinition", err)
			}
			if !strings.Contains(err.Error(), tt.path) {
				t.Fatalf("Load() error = %v, want field path %q", err, tt.path)
			}
		})
	}

	_, err := database.Load(context.Background(), fstest.MapFS{
		"database.json": &fstest.MapFile{Data: []byte(strings.Replace(validDefinitionJSON, `"max_bytes": 63`, `"max_bytes": "not-an-integer"`, 1))},
	})
	if err == nil {
		t.Fatal("Load() error = nil, want typed configuration rejection")
	}
	var typeError *json.UnmarshalTypeError
	if !errors.As(err, &typeError) {
		t.Fatalf("Load() error = %v, want retained json.UnmarshalTypeError", err)
	}
	if !strings.Contains(err.Error(), `$.identifiers.max_bytes`) {
		t.Fatalf("Load() error = %v, want exact typed field path", err)
	}

	_, err = database.Load(context.Background(), fstest.MapFS{
		"database.json": &fstest.MapFile{Data: []byte(strings.Replace(validDefinitionJSON, `"bits": 32`, `"bits": {}`, 1))},
	})
	if err == nil {
		t.Fatal("Load() error = nil, want compound typed configuration rejection")
	}
	if !errors.As(err, &typeError) {
		t.Fatalf("Load() error = %v, want retained json.UnmarshalTypeError", err)
	}
	if !strings.Contains(err.Error(), `$.type_mappings[0].logical.bits`) {
		t.Fatalf("Load() error = %v, want exact indexed typed field path", err)
	}
}

func TestDatabaseDefinitionEnforcesSchemaNumericConstraints(t *testing.T) {
	tests := []struct {
		name       string
		document   string
		path       string
		value      string
		constraint string
	}{
		{
			name:       "logical minimum",
			document:   strings.Replace(validDefinitionJSON, `"bits": 32`, `"bits": 0`, 1),
			path:       `$.type_mappings[0].logical.bits`,
			value:      "0",
			constraint: "minimum 8",
		},
		{
			name:       "identifier maximum",
			document:   strings.Replace(validDefinitionJSON, `"max_bytes": 63`, `"max_bytes": 257`, 1),
			path:       `$.identifiers.max_bytes`,
			value:      "257",
			constraint: "maximum 256",
		},
		{
			name:       "schema version enum",
			document:   strings.Replace(validDefinitionJSON, `"schema_version": 1`, `"schema_version": 2`, 1),
			path:       `$.schema_version`,
			value:      "2",
			constraint: "enum [1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := database.Load(context.Background(), fstest.MapFS{
				"database.json": &fstest.MapFile{Data: []byte(tt.document)},
			})
			if err == nil {
				t.Fatal("Load() error = nil, want schema constraint rejection")
			}
			if !errors.Is(err, database.ErrInvalidDefinition) {
				t.Fatalf("Load() error = %v, want ErrInvalidDefinition", err)
			}
			var definitionError *database.DefinitionError
			if !errors.As(err, &definitionError) {
				t.Fatalf("Load() error = %v, want retained DefinitionError", err)
			}
			for _, expected := range []string{tt.path, "value " + tt.value, tt.constraint} {
				if !strings.Contains(err.Error(), expected) {
					t.Fatalf("Load() error = %v, want %q", err, expected)
				}
			}
		})
	}
}

func TestResourcePolicyBoundsEveryDatabaseResource(t *testing.T) {
	policy := loadTestDefinition(t, validDefinitionJSON).Resources()

	if got, err := policy.EffectivePageSize(0); err != nil || got != 100 {
		t.Fatalf("EffectivePageSize(0) = (%d, %v), want (100, nil)", got, err)
	}
	if got, err := policy.EffectiveBatchSize(0); err != nil || got != 25 {
		t.Fatalf("EffectiveBatchSize(0) = (%d, %v), want (25, nil)", got, err)
	}
	if got, err := policy.EffectivePoolSize(0); err != nil || got != 2 {
		t.Fatalf("EffectivePoolSize(0) = (%d, %v), want (2, nil)", got, err)
	}
	if _, err := policy.EffectiveBatchSize(251); err == nil {
		t.Fatal("EffectiveBatchSize(max+1) = nil, want bounded refusal")
	}
	if _, err := policy.EffectivePoolSize(9); err == nil {
		t.Fatal("EffectivePoolSize(max+1) = nil, want bounded refusal")
	}

	operationContext, cancel, err := policy.WithOperationTimeout(context.Background())
	if err != nil {
		t.Fatalf("WithOperationTimeout() error = %v", err)
	}
	defer cancel()
	if _, hasDeadline := operationContext.Deadline(); !hasDeadline {
		t.Fatal("WithOperationTimeout() returned a context without a finite deadline")
	}

	for _, mutation := range []struct {
		name   string
		mutate func(*database.ResourcePolicy)
	}{
		{name: "zero max parameters", mutate: func(p *database.ResourcePolicy) { p.MaxParameters = 0 }},
		{name: "zero connect timeout", mutate: func(p *database.ResourcePolicy) { p.ConnectTimeout = 0 }},
		{name: "zero operation timeout", mutate: func(p *database.ResourcePolicy) { p.OperationTimeout = 0 }},
		{name: "unbounded page maximum", mutate: func(p *database.ResourcePolicy) { p.ReadPage.Maximum = 100_001 }},
		{name: "unbounded batch maximum", mutate: func(p *database.ResourcePolicy) { p.WriteBatch.Maximum = 10_001 }},
		{name: "unbounded pool maximum", mutate: func(p *database.ResourcePolicy) { p.Pool.Maximum = 129 }},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			invalid := policy
			mutation.mutate(&invalid)
			if _, _, err := invalid.WithOperationTimeout(context.Background()); err == nil {
				t.Fatal("WithOperationTimeout() error = nil, want invalid resource policy refusal")
			}
		})
	}
}

func TestLogicalTypeCompatibilityIsLosslessOrRejected(t *testing.T) {
	int32Type, err := database.NewSignedInteger(32)
	if err != nil {
		t.Fatal(err)
	}
	int64Type, err := database.NewSignedInteger(64)
	if err != nil {
		t.Fatal(err)
	}
	if plan, err := database.CompileTypePlan(int32Type, int64Type); err != nil || plan.Classification() != database.CompatibilityLossless {
		t.Fatalf("CompileTypePlan(int32, int64) = (%v, %v), want lossless plan", plan, err)
	}

	textType, err := database.NewString(0, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.CompileTypePlan(int32Type, textType); err == nil {
		t.Fatal("CompileTypePlan(int32, text) = nil, want explicit refusal instead of string fallback")
	}

	withTimezone, err := database.NewTimestamp(6, true)
	if err != nil {
		t.Fatal(err)
	}
	withoutTimezone, err := database.NewTimestamp(6, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.CompileTypePlan(withTimezone, withoutTimezone); err == nil {
		t.Fatal("CompileTypePlan(timestamp with zone, without zone) = nil, want refusal")
	}

	opaque, err := database.NewOpaqueNative("postgres", "citext", nil)
	if err != nil {
		t.Fatal(err)
	}
	arrayOfOpaque, err := database.NewArray(opaque)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name      string
		typeValue database.LogicalType
	}{
		{name: "direct opaque native", typeValue: opaque},
		{name: "nested opaque native", typeValue: arrayOfOpaque},
	} {
		t.Run(tt.name, func(t *testing.T) {
			classification, err := database.ClassifyTypeCompatibility(tt.typeValue, tt.typeValue)
			if err != nil || classification != database.CompatibilityUnsupported {
				t.Fatalf("ClassifyTypeCompatibility() = (%q, %v), want unsupported", classification, err)
			}
			if _, err := database.CompileTypePlan(tt.typeValue, tt.typeValue); err == nil {
				t.Fatal("CompileTypePlan() error = nil, want unsupported type refusal")
			}
		})
	}

	decimal93, err := database.NewDecimal(9, 3)
	if err != nil {
		t.Fatal(err)
	}
	decimal124, err := database.NewDecimal(12, 4)
	if err != nil {
		t.Fatal(err)
	}
	if plan, err := database.CompileTypePlan(decimal93, decimal124); err != nil || plan.Classification() != database.CompatibilityLossless {
		t.Fatalf("CompileTypePlan(decimal(9,3), decimal(12,4)) = (%v, %v), want lossless plan", plan, err)
	}

	arrayOfInt32, err := database.NewArray(int32Type)
	if err != nil {
		t.Fatal(err)
	}
	arrayOfInt64, err := database.NewArray(int64Type)
	if err != nil {
		t.Fatal(err)
	}
	if plan, err := database.CompileTypePlan(arrayOfInt32, arrayOfInt64); err != nil || plan.Classification() != database.CompatibilityLossless {
		t.Fatalf("CompileTypePlan(array<int32>, array<int64>) = (%v, %v), want lossless plan", plan, err)
	}
}

func TestStructuredCatalogIdentityAndReadPlanAreStable(t *testing.T) {
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
	updatedAtColumn := database.ColumnRef{Relation: relation, Name: "updated_at"}
	int64Type, err := database.NewSignedInteger(64)
	if err != nil {
		t.Fatal(err)
	}
	timestamp, err := database.NewTimestamp(6, true)
	if err != nil {
		t.Fatal(err)
	}
	catalogInput := []database.Relation{{
		Ref: relation,
		Columns: []database.Column{
			{Ref: idColumn, Type: int64Type, Ordinal: 1},
			{Ref: updatedAtColumn, Type: timestamp, Ordinal: 2},
		},
		Keys: []database.Key{{
			Name:    "widgets_pk",
			Kind:    database.KeyPrimary,
			Columns: []database.ColumnRef{idColumn},
		}},
	}}
	catalog, err := database.NewCatalog(database.CatalogRef{Name: "analytics"}, catalogInput)
	if err != nil {
		t.Fatalf("NewCatalog() error = %v", err)
	}
	catalogInput[0].Columns[0].Ref.Name = "input_mutated"
	if got := catalog.Relations()[0].Columns[0].Ref.Name; got == "input_mutated" {
		t.Fatal("NewCatalog() retained caller-owned mutable relation state")
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
	target, err := database.NewTargetRef(identity, relation)
	if err != nil {
		t.Fatal(err)
	}
	if got := source.Identity(); got != identity {
		t.Fatalf("source identity = %#v, want %#v", got, identity)
	}
	if got := target.Identity(); got != identity {
		t.Fatalf("target identity = %#v, want %#v", got, identity)
	}
	if _, err := database.NewTargetRef(database.ConnectionIdentity{
		WorkspaceID: "workspace-1",
		ConnectorID: "postgres",
	}, relation); err == nil {
		t.Fatal("NewTargetRef() accepted an owner without a connection ID")
	}

	plan, err := database.NewReadPlan(context.Background(), database.ReadPlanRequest{
		Inbound:    inbound,
		Definition: definition,
		Catalog:    catalog,
		Relation:   relation,
		Columns:    []database.ColumnRef{idColumn, updatedAtColumn},
		Order: []database.OrderTerm{
			{Column: updatedAtColumn, Direction: database.SortAscending},
			{Column: idColumn, Direction: database.SortAscending},
		},
		PageSize: 0,
	})
	if err != nil {
		t.Fatalf("NewReadPlan() error = %v", err)
	}
	if plan.SchemaFingerprint().IsZero() {
		t.Fatal("read plan has no schema fingerprint")
	}
	if got := plan.PageSize(); got != definition.Resources().ReadPage.Default {
		t.Fatalf("read plan page size = %d, want definition default %d", got, definition.Resources().ReadPage.Default)
	}
	if got := plan.Warehouse(); got.Identity() != identity || got.Table() != "widgets" {
		t.Fatalf("read plan warehouse = %#v/%q, want source-owned widgets artifact", got.Identity(), got.Table())
	}
	relations := catalog.Relations()
	relations[0].Columns[0].Ref.Name = "mutated"
	if got := catalog.Relations()[0].Columns[0].Ref.Name; got == "mutated" {
		t.Fatal("Catalog.Relations() returned mutable internal state")
	}
	columns := plan.Columns()
	columns[0].Name = "mutated"
	if got := plan.Columns()[0].Name; got == "mutated" {
		t.Fatal("ReadPlan.Columns() returned mutable internal state")
	}
	order := plan.Order()
	order[0].Direction = database.SortDescending
	if got := plan.Order()[0].Direction; got == database.SortDescending {
		t.Fatal("ReadPlan.Order() returned mutable internal state")
	}

	permutedColumns := catalog.Relations()[0].Columns
	permutedColumns[0], permutedColumns[1] = permutedColumns[1], permutedColumns[0]
	permuted, err := database.NewCatalog(database.CatalogRef{Name: "analytics"}, []database.Relation{{
		Ref:     relation,
		Columns: permutedColumns,
		Keys: []database.Key{{
			Name:    "widgets_pk",
			Kind:    database.KeyPrimary,
			Columns: []database.ColumnRef{idColumn},
		}},
	}})
	if err != nil {
		t.Fatalf("NewCatalog(permuted columns) error = %v", err)
	}
	if got, want := permuted.Fingerprint(), catalog.Fingerprint(); got != want {
		t.Fatalf("schema fingerprint changed with column input order: got %s, want %s", got, want)
	}

	_, err = database.NewReadPlan(context.Background(), database.ReadPlanRequest{
		Inbound:    inbound,
		Definition: definition,
		Catalog:    catalog,
		Relation:   relation,
		Columns:    []database.ColumnRef{idColumn, updatedAtColumn},
		Order:      []database.OrderTerm{{Column: updatedAtColumn, Direction: database.SortAscending}},
		PageSize:   100,
	})
	if err == nil {
		t.Fatal("NewReadPlan() accepted an order without the unique-key suffix")
	}

	_, err = database.NewReadPlan(context.Background(), database.ReadPlanRequest{
		Inbound:    inbound,
		Definition: definition,
		Catalog:    catalog,
		Relation:   relation,
		Columns:    []database.ColumnRef{idColumn, updatedAtColumn},
		Order: []database.OrderTerm{
			{Column: updatedAtColumn, Direction: database.SortAscending},
			{Column: idColumn, Direction: database.SortAscending},
		},
		PageSize: definition.Resources().ReadPage.Maximum + 1,
	})
	if err == nil {
		t.Fatal("NewReadPlan() accepted a page size above the definition maximum")
	}

	emailColumn := database.ColumnRef{Relation: relation, Name: "email"}
	textType, err := database.NewString(0, "")
	if err != nil {
		t.Fatal(err)
	}
	nullableUniqueCatalog, err := database.NewCatalog(database.CatalogRef{Name: "analytics"}, []database.Relation{{
		Ref: relation,
		Columns: []database.Column{
			{Ref: idColumn, Type: int64Type, Ordinal: 1},
			{Ref: emailColumn, Type: textType, Nullable: true, Ordinal: 2},
		},
		Keys: []database.Key{{
			Name:    "widgets_email_key",
			Kind:    database.KeyUnique,
			Columns: []database.ColumnRef{emailColumn},
		}},
	}})
	if err != nil {
		t.Fatalf("NewCatalog(nullable unique key) error = %v", err)
	}
	if _, err := database.NewReadPlan(context.Background(), database.ReadPlanRequest{
		Inbound:    inbound,
		Definition: definition,
		Catalog:    nullableUniqueCatalog,
		Relation:   relation,
		Columns:    []database.ColumnRef{emailColumn},
		Order:      []database.OrderTerm{{Column: emailColumn, Direction: database.SortAscending}},
		PageSize:   100,
	}); err == nil {
		t.Fatal("NewReadPlan() accepted a nullable unique-key suffix")
	}

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := database.NewReadPlan(cancelled, database.ReadPlanRequest{
		Inbound:    inbound,
		Definition: definition,
		Catalog:    catalog,
		Relation:   relation,
		Columns:    []database.ColumnRef{idColumn, updatedAtColumn},
		Order: []database.OrderTerm{
			{Column: updatedAtColumn, Direction: database.SortAscending},
			{Column: idColumn, Direction: database.SortAscending},
		},
		PageSize: 100,
	}); !errors.Is(err, context.Canceled) {
		t.Fatalf("NewReadPlan(cancelled) error = %v, want context.Canceled", err)
	}
}

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

func TestDatabaseLoadAndReadPlanHonorCancellation(t *testing.T) {
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := database.Load(cancelled, fstest.MapFS{"database.json": &fstest.MapFile{Data: []byte(validDefinitionJSON)}}); !errors.Is(err, context.Canceled) {
		t.Fatalf("Load(cancelled) error = %v, want context.Canceled", err)
	}
}

func TestDatabaseLoadChecksCancellationBeforeReturningProjection(t *testing.T) {
	ctx := &cancelOnErrCallContext{cancelAt: 4, done: make(chan struct{})}
	definition, err := database.Load(ctx, fstest.MapFS{"database.json": &fstest.MapFile{Data: []byte(validDefinitionJSON)}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Load() error = %v, want context.Canceled", err)
	}
	if definition.SchemaVersion() != 0 {
		t.Fatalf("Load() definition schema version = %d, want zero projection after cancellation", definition.SchemaVersion())
	}
}

func TestReadPlanChecksCancellationBeforeReturningProjection(t *testing.T) {
	ctx := &cancelOnErrCallContext{cancelAt: 3, done: make(chan struct{})}
	plan, err := database.NewReadPlan(ctx, testReadPlanRequest(t))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("NewReadPlan() error = %v, want context.Canceled", err)
	}
	if plan.PageSize() != 0 {
		t.Fatalf("NewReadPlan() page size = %d, want zero plan after cancellation", plan.PageSize())
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
	native   synccontract.NativeSyncExecutorDescriptor
	evidence synccontract.ConformanceEvidence
}

func nativeContract(protocol, command, executorID string) synccontract.NativeCommandContract {
	return synccontract.NativeCommandContract{
		ContractVersion: synccontract.NativeCommandContractVersion,
		Protocol:        protocol,
		Command:         command,
		Executor:        synccontract.ExecutorReference{Kind: "native", ID: executorID},
		Modes:           []synccontract.Mode{synccontract.ModeFullAppend},
		Conformance:     synccontract.RequiredConformanceEvidence(),
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
		evidence: contract.Conformance,
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

func (d nativeAdmission) NativeSyncConformanceEvidence() synccontract.ConformanceEvidence {
	return d.evidence
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
