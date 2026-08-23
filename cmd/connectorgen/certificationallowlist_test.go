package main

import "testing"

func TestCertificationConnectorAllowlistIncludesBatch1LiveTargets(t *testing.T) {
	for _, connector := range []string{"jira", "asana", "notion"} {
		t.Run(connector, func(t *testing.T) {
			if !certificationConnectorAllowed(connector) {
				t.Fatalf("certificationConnectorAllowed(%q) = false, want true", connector)
			}
		})
	}
}
