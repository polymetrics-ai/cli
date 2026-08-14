package database_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/synccontract"
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
