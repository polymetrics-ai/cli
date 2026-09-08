package app_test

import (
	"errors"
	"os"
	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	"testing"
)

type savedPreflightProbe206 struct {
	*engine.Connector
	calls   []string
	failure error
}

func (p *savedPreflightProbe206) PreflightSavedWriteAction(name string) error {
	p.calls = append(p.calls, name)
	return p.failure
}

func TestPlanReverseETLChecksSelectedSavedPreflight206(t *testing.T) {
	bundle, err := engine.Load(os.DirFS("testdata/bundles"), humanProxyConnector)
	if err != nil {
		t.Fatal(err)
	}
	failure := errors.New("independent selected saved schema refusal")
	probe := &savedPreflightProbe206{Connector: engine.New(bundle, nil), failure: failure}
	registry := connectors.NewEmptyRegistry()
	if err := registry.Register(probe); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := app.OpenWithRegistry(root, registry)
	if err != nil {
		t.Fatal(err)
	}
	_, err = a.PlanReverseETL(t.Context(), app.PlanReverseETLRequest{Name: "preflight", SourceTable: "not_created", DestinationConnector: humanProxyConnector, DestinationCredential: "humanproxy-local", Action: "sync_profile", Mappings: map[string]string{"name": "name"}})
	if !errors.Is(err, failure) || len(probe.calls) != 1 || probe.calls[0] != "sync_profile" {
		t.Fatalf("selected preflight error=%v calls=%v; must refuse before source acquisition", err, probe.calls)
	}
}
