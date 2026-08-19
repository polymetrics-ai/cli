package certifications

import (
	"encoding/json"
	"testing"
)

func TestGeneratedStatusMakesUncertifiedConnectorVisible(t *testing.T) {
	status, err := StatusFor("github")
	if err != nil {
		t.Fatalf("StatusFor(github) error = %v", err)
	}
	if status.Certified {
		t.Fatal("github unexpectedly reported certified without accepted live proof")
	}
	if status.Label != uncertifiedLabel || status.Warning != uncertifiedWarning {
		t.Fatalf("github status = %#v, want visible community-build warning", status)
	}
}

func TestGeneratedStatusIncludesEveryAllowlistedConnector(t *testing.T) {
	var artifact statusArtifact
	if err := decodeStatusJSON(embeddedStatusJSON, &artifact); err != nil {
		t.Fatalf("parse generated status artifact: %v", err)
	}
	expected := make(map[string]bool, len(artifact.CertificationScope))
	for _, connector := range artifact.CertificationScope {
		if expected[connector] {
			t.Fatalf("generated status artifact duplicates allowlisted connector %q", connector)
		}
		expected[connector] = true
	}

	statuses, err := AllStatuses()
	if err != nil {
		t.Fatalf("AllStatuses() error = %v", err)
	}
	if len(statuses) == 0 {
		t.Fatal("AllStatuses() returned no generated connector statuses")
	}
	found := map[string]bool{}
	for _, status := range statuses {
		found[status.Connector] = true
	}
	for connector := range expected {
		if !found[connector] {
			t.Fatalf("AllStatuses() omitted allowlisted connector %q", connector)
		}
	}
	for connector := range found {
		if !expected[connector] {
			t.Fatalf("AllStatuses() returned non-allowlisted connector %q", connector)
		}
	}
	if len(found) != len(expected) {
		t.Fatalf("AllStatuses() returned %d statuses, want exactly %d allowlisted connectors", len(found), len(expected))
	}
}

func TestGeneratedStatusRejectsMissingScopedConnector(t *testing.T) {
	artifact := statusArtifact{
		SchemaVersion:      statusSchemaVersion,
		GeneratedCommand:   statusGeneratedCommand,
		CertificationScope: []string{"github", "postgres"},
		Connectors: []Status{{
			Connector: "github",
			Label:     uncertifiedLabel,
			Warning:   uncertifiedWarning,
		}},
	}
	raw, err := json.Marshal(artifact)
	if err != nil {
		t.Fatalf("marshal status artifact: %v", err)
	}
	if _, _, err := loadStatusArtifact(raw); err == nil || err.Error() != `generated certification status omits connector "postgres"` {
		t.Fatalf("loadStatusArtifact() error = %v, want omitted postgres error", err)
	}
}

func TestUnallowlistedConnectorHasNoCertificationClaim(t *testing.T) {
	status, err := StatusForRegistered("mysql", true)
	if err != nil {
		t.Fatalf("StatusFor(mysql) error = %v", err)
	}
	if status.Certified || status.Label != uncertifiedLabel || status.Warning != uncertifiedWarning {
		t.Fatalf("mysql status = %#v, want unchanged visible uncertified warning", status)
	}
}

func TestUnknownConnectorStatusRemainsAnError(t *testing.T) {
	if _, err := StatusForRegistered("not-a-connector", false); err == nil {
		t.Fatal("StatusForRegistered(not-a-connector) error = nil, want omitted-status error")
	}
}
