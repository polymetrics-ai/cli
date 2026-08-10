package certifications

import "testing"

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

func TestGeneratedStatusIncludesEveryConnector(t *testing.T) {
	statuses, err := AllStatuses()
	if err != nil {
		t.Fatalf("AllStatuses() error = %v", err)
	}
	if len(statuses) == 0 {
		t.Fatal("AllStatuses() returned no generated connector statuses")
	}
	foundGitHub := false
	for _, status := range statuses {
		if status.Connector == "github" {
			foundGitHub = true
		}
	}
	if !foundGitHub {
		t.Fatal("AllStatuses() omitted github")
	}
}
