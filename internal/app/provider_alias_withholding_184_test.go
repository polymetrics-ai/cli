package app_test

import (
	"context"
	"strings"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
)

// A fixture uses the ordinary admitted alias shape; it does not grant a new
// runtime executor or replace the source projection/admission tests.
type providerAliasWithholding184 struct{ withholdingConnector }

func (c *providerAliasWithholding184) CommandSurface() *connectors.CommandSurface {
	s := c.withholdingConnector.CommandSurface()
	for i := range s.Commands {
		for j := range s.Commands[i].Flags {
			f := &s.Commands[i].Flags[j]
			if f.Name == "access-token" {
				f.Name = "provider-config"
				f.EnvOnly = true
			}
		}
	}
	return s
}

func TestProviderAliasWithheldReplay184(t *testing.T) {
	ctx := context.Background()
	connector := &providerAliasWithholding184{}
	a, root := withholdingApp(t, ctx, connector)
	flags := map[string][]string{"client-id": {"Iv1.fixture"}, "provider-config": {tokenSentinel}}
	plan, _, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{Name: "pm-display", Connector: connector.Name(), Credential: "withholding-local", Path: []string{"oauth", "revoke"}, Flags: flags})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stateBytes(t, root), tokenSentinel) {
		t.Fatal("alias value persisted")
	}
	if _, _, err = a.PreviewConnectorCommandPlan(ctx, plan.ID, nil); err == nil || !strings.Contains(err.Error(), "--provider-config") {
		t.Fatal("missing resupply did not name alias")
	}
	wrong := map[string][]string{"client-id": {"Iv1.fixture"}, "provider-config": {"different-fixture-value"}}
	if _, _, err = a.PreviewConnectorCommandPlan(ctx, plan.ID, wrong); err == nil {
		t.Fatal("wrong resupply accepted")
	}
	_, preview, err := a.PreviewConnectorCommandPlan(ctx, plan.ID, flags)
	if err != nil {
		t.Fatal(err)
	}
	if preview.RecordsStaged != 1 || strings.Contains(stateBytes(t, root), tokenSentinel) {
		t.Fatal("preview record count or withholding failed")
	}
	run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken, WithheldFlags: flags})
	if err != nil {
		t.Fatal(err)
	}
	if run.RecordsSucceeded != 1 || connector.lastWritten["access_token"] != tokenSentinel {
		t.Fatal("actual reconstituted provider record differs")
	}
	if strings.Contains(stateBytes(t, root), tokenSentinel) {
		t.Fatal("execute persisted withheld value")
	}
}
