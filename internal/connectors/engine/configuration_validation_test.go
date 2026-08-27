package engine

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/failures"
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
		{name: "invalid uri", config: map[string]string{"base_url": "not-a-uri"}, wantError: `/base_url: configuration value must use "uri" format`},
		{name: "invalid date", config: map[string]string{"start_date": "2026-02-30"}, wantError: `/start_date: configuration value must use "date" format`},
		{name: "invalid date-time", config: map[string]string{"since": "not-a-date-time"}, wantError: `/since: configuration value must use "date-time" format`},
		{name: "invalid pattern", config: map[string]string{"domain": "not.allowed"}, wantError: "/domain: configuration value does not match the declared pattern"},
		{name: "invalid enum", config: map[string]string{"environment": "preview"}, wantError: "/environment: configuration value must be one of the declared values"},
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

func TestSchemaValidateConfigurationRejectsCredentialBearingBaseURL(t *testing.T) {
	sch, err := CompileSchema(json.RawMessage(`{
		"type": "object",
		"properties": {
			"base_url": {"type": "string", "format": "uri"}
		}
	}`))
	if err != nil {
		t.Fatalf("CompileSchema() error = %v", err)
	}

	tests := []struct {
		name      string
		baseURL   string
		wantError string
	}{
		{
			name:      "rejects URL user info",
			baseURL:   "https://user:pass@api.provider.example/v2",
			wantError: `/base_url: configuration value must use "uri" format`,
		},
		{
			name:      "rejects query component",
			baseURL:   "https://api.provider.example/v2?token=placeholder",
			wantError: `/base_url: configuration value must use "uri" format`,
		},
		{
			name:    "accepts ordinary endpoint",
			baseURL: "https://api.provider.example/v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sch.ValidateConfiguration(map[string]string{"base_url": tt.baseURL})
			if tt.wantError == "" {
				if err != nil {
					t.Fatalf("ValidateConfiguration() error = %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("ValidateConfiguration() error = nil, want credential-bearing URL rejection")
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("ValidateConfiguration() error = %q, want %q", err, tt.wantError)
			}
			if strings.Contains(err.Error(), tt.baseURL) {
				t.Fatalf("ValidateConfiguration() error leaked rejected URL")
			}
		})
	}
}

func TestSchemaValidateConfigurationRejectsCredentialBearingBaseURLWithoutDeclaredURIFormat(t *testing.T) {
	sch, err := CompileSchema(json.RawMessage(`{
		"type": "object",
		"properties": {
			"base_url": {"type": "string"}
		}
	}`))
	if err != nil {
		t.Fatalf("CompileSchema() error = %v", err)
	}
	if !sch.HasConfigurationConstraints() {
		t.Fatal("HasConfigurationConstraints() = false, want base_url safety constraint")
	}

	for _, tt := range []struct {
		name      string
		baseURL   string
		wantError bool
	}{
		{name: "user info", baseURL: "https://user:pass@api.provider.example/v2", wantError: true},
		{name: "query component", baseURL: "https://api.provider.example/v2?token=placeholder", wantError: true},
		{name: "ordinary endpoint", baseURL: "https://api.provider.example/v2"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := sch.ValidateConfiguration(map[string]string{"base_url": tt.baseURL})
			if !tt.wantError {
				if err != nil {
					t.Fatalf("ValidateConfiguration() error = %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("ValidateConfiguration() error = nil, want credential-bearing URL rejection")
			}
			if strings.Contains(err.Error(), tt.baseURL) {
				t.Fatal("ValidateConfiguration() error leaked rejected URL")
			}
		})
	}
}

func TestSchemaWithoutConfigurationConstraintsIsNotAdvertised(t *testing.T) {
	sch, err := CompileSchema(json.RawMessage(`{
		"type": "object",
		"required": ["count"],
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

func TestSchemaValidateConfigurationReturnsSharedClassification(t *testing.T) {
	sch, err := CompileSchema(json.RawMessage(`{
		"type": "object",
		"properties": {"base_url": {"type": "string", "format": "uri"}}
	}`))
	if err != nil {
		t.Fatalf("CompileSchema() error = %v", err)
	}
	err = sch.ValidateConfiguration(map[string]string{"base_url": "not-a-uri"})
	var classification *failures.Classification
	if !errors.As(err, &classification) {
		t.Fatalf("ValidateConfiguration() error type = %T, want shared classification", err)
	}
	if got, want := classification.Domain(), failures.DomainConfiguration; got != want {
		t.Fatalf("Domain() = %q, want %q", got, want)
	}
	if got, want := classification.Code(), "format_mismatch"; got != want {
		t.Fatalf("Code() = %q, want %q", got, want)
	}
	if got, want := classification.FieldPath(), "/base_url"; got != want {
		t.Fatalf("FieldPath() = %q, want %q", got, want)
	}
	if classification.Retryable() {
		t.Fatal("configuration failure is retryable, want false")
	}
	if classification.Cause() == nil {
		t.Fatal("configuration failure omitted internal cause")
	}
}

func TestJSONPointerForKeyEscapesRFC6901Segments(t *testing.T) {
	if got, want := jsonPointerForKey("client/id~primary"), "/client~1id~0primary"; got != want {
		t.Fatalf("jsonPointerForKey() = %q, want %q", got, want)
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
