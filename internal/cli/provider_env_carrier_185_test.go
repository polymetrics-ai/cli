package cli

import (
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestProviderAliasEnvironmentCarrier185(t *testing.T) {
	t.Setenv("PM_CP16_ALIAS_ENV_185", "synthetic-withheld-value")
	surface := &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{{Path: "provider configure", Flags: []connectors.CommandSurfaceFlag{{Name: "provider-config", Type: "string", MapsTo: "record.config", Required: true, EnvOnly: true}}}}}
	resolved, err := resolveConnectorCommandEnvironmentOnlyFlags(surface, []string{"provider", "configure"}, map[string][]string{"from-env": {"provider-config=PM_CP16_ALIAS_ENV_185"}, "config": {"rate_limit_account=pm-fixture"}})
	if err != nil {
		t.Fatal(err)
	}
	prepared := connectorCommandFlags(resolved)
	if len(prepared["provider-config"]) != 1 || prepared["provider-config"][0] != "synthetic-withheld-value" || len(prepared["config"]) != 0 || len(prepared["from-env"]) != 0 {
		t.Fatal("carrier mixed provider value and PM configuration")
	}
	_, err = resolveConnectorCommandEnvironmentOnlyFlags(surface, []string{"provider", "configure"}, map[string][]string{"provider-config": {"synthetic-withheld-value"}})
	if err == nil || !strings.Contains(err.Error(), "--from-env") || strings.Contains(err.Error(), "synthetic-withheld-value") {
		t.Fatal("unsafe direct carrier/refusal")
	}
}
