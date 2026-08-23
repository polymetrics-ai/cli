package certify

import (
	"crypto/sha256"
	"errors"
	"os"
	"testing"

	"polymetrics.ai/internal/credential"
)

func TestCertificationSourceSecretMaterialHonorsDeclaredFieldContracts(t *testing.T) {
	t.Run("required explicit empty", func(t *testing.T) {
		t.Setenv("PM_CERT_REQUIRED_EMPTY", "")
		_, _, err := certificationSourceSecretMaterial(Options{
			Connector: "stripe",
			SecretEnv: map[string]string{"client_secret": "PM_CERT_REQUIRED_EMPTY"},
		})
		var empty *credential.EmptySecretError
		if !errors.As(err, &empty) {
			t.Fatalf("certificationSourceSecretMaterial() error type = %T, want typed empty-secret classification", err)
		}
	})

	t.Run("optional explicit empty", func(t *testing.T) {
		t.Setenv("PM_CERT_OPTIONAL_EMPTY", "")
		fields, values, err := certificationSourceSecretMaterial(Options{
			Connector: "github",
			SecretEnv: map[string]string{"token": "PM_CERT_OPTIONAL_EMPTY"},
		})
		if err != nil {
			t.Fatalf("certificationSourceSecretMaterial() error = %v", err)
		}
		if len(fields) != 0 || len(values) != 0 {
			t.Fatal("optional explicit empty credential was retained")
		}
	})

	t.Run("optional absent", func(t *testing.T) {
		const envName = "PM_CERT_OPTIONAL_ABSENT"
		previous, wasSet := os.LookupEnv(envName)
		if err := os.Unsetenv(envName); err != nil {
			t.Fatalf("unset environment: %v", err)
		}
		t.Cleanup(func() {
			if wasSet {
				_ = os.Setenv(envName, previous)
				return
			}
			_ = os.Unsetenv(envName)
		})
		fields, values, err := certificationSourceSecretMaterial(Options{
			Connector: "github",
			SecretEnv: map[string]string{"token": envName},
		})
		if err != nil {
			t.Fatalf("certificationSourceSecretMaterial() error = %v", err)
		}
		if len(fields) != 0 || len(values) != 0 {
			t.Fatal("absent optional credential was retained")
		}
	})

	t.Run("required non-empty", func(t *testing.T) {
		const token = "certification-handoff-canary"
		t.Setenv("PM_CERT_REQUIRED_VALUE", token)
		fields, values, err := certificationSourceSecretMaterial(Options{
			Connector: "stripe",
			SecretEnv: map[string]string{"client_secret": "PM_CERT_REQUIRED_VALUE"},
		})
		if err != nil {
			t.Fatalf("certificationSourceSecretMaterial() error = %v", err)
		}
		if len(fields) != 1 || len(values) != 1 {
			t.Fatal("required non-empty credential was not retained")
		}
		wantHash := sha256.Sum256([]byte(token))
		fieldHash := sha256.Sum256([]byte(fields["client_secret"]))
		valueHash := sha256.Sum256([]byte(values[0]))
		if len(fields["client_secret"]) != len(token) || fieldHash != wantHash || valueHash != wantHash {
			t.Fatal("required non-empty credential bytes were not preserved")
		}
	})
}
