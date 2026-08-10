package engine

import (
	"encoding/json"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestSchemaValidateConfigurationAppliesDeclaredConstraintsOnly(t *testing.T) {
	sch, err := CompileSchema(json.RawMessage(`{
		"type": "object",
		"required": ["unrelated_required"],
		"properties": {
			"base_url": {"type": "string", "format": "uri"},
			"start_date": {"type": "string", "format": "date"},
			"since": {"type": "string", "format": "date-time"},
			"domain": {"type": "string", "pattern": "^[a-z0-9-]+$"},
			"environment": {"type": "string", "enum": ["production", "sandbox"]},
			"unconstrained": {"type": "string"}
		}
	}`))
	if err != nil {
		t.Fatalf("CompileSchema() error = %v", err)
	}
	if !sch.HasConfigurationConstraints() {
		t.Fatal("HasConfigurationConstraints() = false, want true")
	}

	tests := []struct {
		name      string
		config    map[string]string
		wantError string
	}{
		{
			name: "valid declared values",
			config: map[string]string{
				"base_url":    "https://api.example.test/v1",
				"start_date":  "2026-02-03",
				"since":       "2026-02-03T04:05:06Z",
				"domain":      "acme-42",
				"environment": "production",
			},
		},
		{name: "invalid uri", config: map[string]string{"base_url": "not-a-uri"}, wantError: `/base_url: value does not match format "uri"`},
		{name: "invalid date", config: map[string]string{"start_date": "2026-02-30"}, wantError: `/start_date: value does not match format "date"`},
		{name: "invalid date-time", config: map[string]string{"since": "not-a-date-time"}, wantError: `/since: value does not match format "date-time"`},
		{name: "invalid pattern", config: map[string]string{"domain": "not.allowed"}, wantError: `/domain: value does not match pattern "^[a-z0-9-]+$"`},
		{name: "invalid enum", config: map[string]string{"environment": "preview"}, wantError: "/environment: value not in enum"},
		{name: "does not enforce required or unconstrained field", config: map[string]string{"unconstrained": "any value"}},
		{name: "does not reject undeclared fields", config: map[string]string{"extra": "still accepted"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sch.ValidateConfiguration(tt.config)
			if tt.wantError == "" {
				if err != nil {
					t.Fatalf("ValidateConfiguration() error = %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("ValidateConfiguration() error = nil, want constraint rejection")
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("ValidateConfiguration() error = %q, want %q", err, tt.wantError)
			}
		})
	}
}

func TestSchemaWithoutConfigurationConstraintsIsNotAdvertised(t *testing.T) {
	sch, err := CompileSchema(json.RawMessage(`{
		"type": "object",
		"properties": {
			"count": {"type": "integer"},
			"label": {"type": "string"}
		}
	}`))
	if err != nil {
		t.Fatalf("CompileSchema() error = %v", err)
	}
	if sch.HasConfigurationConstraints() {
		t.Fatal("HasConfigurationConstraints() = true, want false for a spec with no declared configuration constraint")
	}
	if err := sch.ValidateConfiguration(map[string]string{"count": "not-an-integer", "extra": "still accepted"}); err != nil {
		t.Fatalf("ValidateConfiguration() error = %v, want no new constraints", err)
	}
}

func TestSchemaRequiredConfigurationKeyIsNotCredentialAdmission(t *testing.T) {
	sch, err := CompileSchema(json.RawMessage(`{
		"type": "object",
		"required": ["namespace"],
		"properties": {
			"namespace": {"type": "string"},
			"label": {"type": "string"}
		}
	}`))
	if err != nil {
		t.Fatalf("CompileSchema() error = %v", err)
	}
	if sch.HasConfigurationConstraints() {
		t.Fatal("HasConfigurationConstraints() = true, want false for required-only schema")
	}
	for _, config := range []map[string]string{nil, {}, {"label": "x"}, {"namespace": "  "}} {
		if err := sch.ValidateConfiguration(config); err != nil {
			t.Fatalf("ValidateConfiguration(%v) error = %v, want required-only admission", config, err)
		}
	}
	if err := sch.ValidateConfiguration(map[string]string{"namespace": "target"}); err != nil {
		t.Fatalf("ValidateConfiguration(complete) error = %v, want admission", err)
	}
}

// The two kinds of required property this boundary cannot see stay optional
// here: a secret is supplied through a separate map, and a defaulted property
// is materialized by the engine rather than the caller.
func TestSchemaRequiredSecretAndDefaultedKeysAreNotConfigurationConstraints(t *testing.T) {
	sch, err := CompileSchema(json.RawMessage(`{
		"type": "object",
		"required": ["token", "base_url"],
		"properties": {
			"token": {"type": "string", "x-secret": true},
			"base_url": {"type": "string", "default": "https://example.test"}
		}
	}`))
	if err != nil {
		t.Fatalf("CompileSchema() error = %v", err)
	}
	if sch.HasConfigurationConstraints() {
		t.Fatal("HasConfigurationConstraints() = true, want false: neither a required secret nor a defaulted key is checkable against the config map")
	}
	if err := sch.ValidateConfiguration(map[string]string{}); err != nil {
		t.Fatalf("ValidateConfiguration() error = %v, want admission", err)
	}
}

func TestConnectorConfigurationConstraintContractReflectsDeclaration(t *testing.T) {
	unconstrained, err := CompileSchema(json.RawMessage(`{
		"type": "object",
		"properties": {"count": {"type": "integer"}}
	}`))
	if err != nil {
		t.Fatalf("CompileSchema(unconstrained) error = %v", err)
	}
	unconstrainedConnector := New(Bundle{Spec: unconstrained}, nil)
	validator, ok := any(unconstrainedConnector).(connectors.ConfigurationConstraintValidator)
	if !ok {
		t.Fatal("engine connector does not expose ConfigurationConstraintValidator")
	}
	if validator.HasConfigurationConstraints() {
		t.Fatal("HasConfigurationConstraints() = true, want false without a declared constraint")
	}
	if err := connectors.ValidateConfiguration(unconstrainedConnector, map[string]string{"count": "not-an-integer"}); err != nil {
		t.Fatalf("ValidateConfiguration(unconstrained) error = %v, want no validation", err)
	}

	constrained, err := CompileSchema(json.RawMessage(`{
		"type": "object",
		"properties": {"environment": {"type": "string", "enum": ["production", "sandbox"]}}
	}`))
	if err != nil {
		t.Fatalf("CompileSchema(constrained) error = %v", err)
	}
	constrainedConnector := New(Bundle{Spec: constrained}, nil)
	validator, ok = any(constrainedConnector).(connectors.ConfigurationConstraintValidator)
	if !ok || !validator.HasConfigurationConstraints() {
		t.Fatal("constrained engine connector does not advertise its declared constraint")
	}
	if err := connectors.ValidateConfiguration(constrainedConnector, map[string]string{"environment": "preview"}); err == nil {
		t.Fatal("ValidateConfiguration(constrained) error = nil, want enum rejection")
	}
}
