package dockerhub_test

import (
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	connectorDefs "polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

// spec.json declares both docker_username and namespace required, and every
// stream path plus the connection check interpolates namespace. Without
// add-time admission an incomplete credential is accepted and saved, and the
// operator only discovers the gap later, at read/check time, from a
// connector-internal template error.
func TestDockerhubCredentialAdmissionRejectsIncompleteConfig(t *testing.T) {
	bundle, err := engine.Load(connectorDefs.FS, "dockerhub")
	if err != nil {
		t.Fatalf("load Docker Hub bundle: %v", err)
	}
	connector := engine.New(bundle, nil)

	if !connector.HasConfigurationConstraints() {
		t.Fatal("HasConfigurationConstraints() = false, want true (spec.json declares required non-secret config)")
	}

	cases := []struct {
		name    string
		config  map[string]string
		wantKey string
	}{
		{
			name:    "no config at all",
			config:  map[string]string{},
			wantKey: "docker_username",
		},
		{
			name:    "namespace omitted",
			config:  map[string]string{"docker_username": "auth-identity"},
			wantKey: "namespace",
		},
		{
			name:    "namespace blank",
			config:  map[string]string{"docker_username": "auth-identity", "namespace": "   "},
			wantKey: "namespace",
		},
		{
			name:    "docker_username omitted",
			config:  map[string]string{"namespace": "target-namespace"},
			wantKey: "docker_username",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := connectors.ValidateConfiguration(connector, tc.config)
			if err == nil {
				t.Fatalf("ValidateConfiguration(%v) = nil, want an incomplete-credential rejection before the credential can be saved", tc.config)
			}
			if !strings.Contains(err.Error(), tc.wantKey) {
				t.Fatalf("error = %q, want it to name %q", err.Error(), tc.wantKey)
			}
		})
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
