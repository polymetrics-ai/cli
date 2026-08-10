package dockerhub_test

import (
	"testing"

	"polymetrics.ai/internal/connectors"
	connectorDefs "polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func TestDockerhubSchemaRequiredKeysDoNotBecomeGlobalCredentialAdmission(t *testing.T) {
	bundle, err := engine.Load(connectorDefs.FS, "dockerhub")
	if err != nil {
		t.Fatalf("load Docker Hub bundle: %v", err)
	}
	connector := engine.New(bundle, nil)

	if !connector.HasConfigurationConstraints() {
		t.Fatal("HasConfigurationConstraints() = false, want Docker Hub's declared format constraints")
	}
	for _, config := range []map[string]string{
		{},
		{"docker_username": "auth-identity"},
		{"namespace": "target-namespace"},
	} {
		if err := connectors.ValidateConfiguration(connector, config); err != nil {
			t.Fatalf("ValidateConfiguration(%v) = %v, want required-only schema keys to remain connector-local", config, err)
		}
	}
}

// The optional fields stay optional: a complete credential is admitted with
// no repository/tag/PAT/SCIM token, and base_url stays defaulted.
func TestDockerhubCredentialAdmissionAcceptsCompleteConfig(t *testing.T) {
	bundle, err := engine.Load(connectorDefs.FS, "dockerhub")
	if err != nil {
		t.Fatalf("load Docker Hub bundle: %v", err)
	}
	connector := engine.New(bundle, nil)

	config := map[string]string{"docker_username": "auth-identity", "namespace": "target-namespace"}
	if err := connectors.ValidateConfiguration(connector, config); err != nil {
		t.Fatalf("ValidateConfiguration(%v) = %v, want a complete public-read credential to be admitted", config, err)
	}
}
