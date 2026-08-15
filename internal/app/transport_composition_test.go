package app

import (
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

func TestOpenRegistersDefinitionOwnedProductionTransports(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	postgres, ok := a.registry.Get("postgres")
	if !ok {
		t.Fatal("PostgreSQL connector is not registered")
	}
	github, ok := a.registry.Get("github")
	if !ok {
		t.Fatal("GitHub connector is not registered")
	}
	resolved, err := a.transports.Preflight(synctransport.PreflightRequest{
		Source:      postgres,
		Destination: github,
		Stream:      "snapshot",
		Mode:        synccontract.ModeFullAppend,
	})
	if err != nil {
		t.Fatalf("definition-owned PostgreSQL-to-GitHub preflight = %v", err)
	}
	if got, want := resolved.Source.TransportExecutorReference(), (connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "postgres_bounded_snapshot"}); got != want {
		t.Fatalf("registered source reference = %+v, want %+v", got, want)
	}
	if got, want := resolved.Destination.TransportExecutorReference(), (connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyDeclarativeAPI, ID: "issue_label_destination"}); got != want {
		t.Fatalf("registered destination reference = %+v, want %+v", got, want)
	}
	githubResolved, err := a.transports.Preflight(synctransport.PreflightRequest{
		Source:      github,
		Destination: github,
		Stream:      "issues",
		Mode:        synccontract.ModeFullAppend,
	})
	if err != nil {
		t.Fatalf("definition-owned GitHub-to-GitHub preflight = %v", err)
	}
	if got, want := githubResolved.Source.TransportExecutorReference(), (connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyDeclarativeAPI, ID: "issue_label_source"}); got != want {
		t.Fatalf("registered GitHub source reference = %+v, want %+v", got, want)
	}
}
