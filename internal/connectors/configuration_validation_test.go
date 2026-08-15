package connectors

import (
	"errors"
	"testing"

	"polymetrics.ai/internal/failures"
)

type configurationFailureConnector struct {
	Connector
	err error
}

func (c configurationFailureConnector) HasConfigurationConstraints() bool { return true }

func (c configurationFailureConnector) ValidateConfiguration(map[string]string) error { return c.err }

func TestValidateConfigurationPreservesTypedNonRetryableFailure(t *testing.T) {
	cause := errors.New("database declaration schema rejected sslmode")
	want, err := failures.New(failures.Input{
		Domain:    failures.DomainConfiguration,
		Code:      "invalid_sslmode",
		Message:   "sslmode must name a supported transport mode",
		FieldPath: "/connection/sslmode",
		References: []failures.Reference{
			{Kind: failures.ReferenceKindConnector, Value: "postgres"},
		},
	}, cause)
	if err != nil {
		t.Fatalf("failures.New() error = %v", err)
	}

	err = ValidateConfiguration(configurationFailureConnector{err: want}, map[string]string{"sslmode": "invalid"})
	var got *failures.Classification
	if !errors.As(err, &got) {
		t.Fatalf("ValidateConfiguration() error type = %T, want shared classification", err)
	}
	if got.Domain() != failures.DomainConfiguration || got.Retryable() {
		t.Fatalf("classification = domain=%q retryable=%v, want non-retryable configuration", got.Domain(), got.Retryable())
	}
	if got.FieldPath() != "/connection/sslmode" {
		t.Fatalf("FieldPath() = %q", got.FieldPath())
	}
	if !errors.Is(got, cause) {
		t.Fatalf("classification lost internal cause %v", cause)
	}
}
