package cli

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
	"polymetrics.ai/internal/connectors/engine"
)

func TestProviderConfigAndPMPlan185(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls++ }))
	defer server.Close()
	root := t.TempDir()
	registry := bundleregistry.New()
	run := func(args []string) (int, string) {
		t.Helper()
		var out, diag bytes.Buffer
		code := runWithPreflightRegistry(args, &out, &diag, registry)
		return code, out.String() + diag.String()
	}
	if code, _ := run([]string{"init", "--root", root, "--json"}); code != 0 {
		t.Fatal("init")
	}
	if code, _ := run([]string{"credentials", "add", "fixture", "--connector", "github", "--config", "public_access=true", "--config", "base_url=" + server.URL, "--config", "owner=acme", "--config", "repo=widgets", "--root", root, "--json"}); code != 0 {
		t.Fatal("credential")
	}
	connector, _ := registry.Get("github")
	var decl connectors.CommandSurfaceCommand
	for _, c := range connector.(connectors.CommandSurfaceProvider).CommandSurface().Commands {
		if c.Path == "api repos create-webhook" {
			decl = c
		}
	}
	args := productionFixtureArgs178(t, "github", decl, map[string]engine.Bundle{})
	args = append([]string{"github", "api", "repos", "create-webhook"}, args...)
	original := `{"url":"https://example.invalid/original","content_type":"json"}`
	args = append(args, "--provider-config", original, "--config", "rate_limit_account=cp16-plan", "--plan-name", "pm-display-name", "--credential", "fixture", "--root", root)
	code, out := run(args)
	if code != 0 {
		t.Fatalf("plan failed: %s", out)
	}
	match := regexp.MustCompile(`Created connector command plan (\S+)`).FindStringSubmatch(out)
	if len(match) != 2 {
		t.Fatal("missing plan")
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	before, err := a.GetReversePlan(match[1])
	if err != nil {
		t.Fatal(err)
	}
	if before.Name != "pm-display-name" || before.DestinationConfig["rate_limit_account"] != "cp16-plan" {
		t.Fatal("PM display/config not separately persisted")
	}
	want := map[string]any{"url": "https://example.invalid/original", "content_type": "json"}
	if !reflect.DeepEqual(before.ConnectorCommandRecord["config"], want) {
		t.Fatal("provider config was replaced by PM config")
	}
	_, _ = run([]string{"github", "api", "repos", "create-webhook", "--plan", match[1], "--preview", "--provider-config", `{"url":"https://example.invalid/replacement"}`, "--root", root})
	after, err := a.GetReversePlan(match[1])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(after.ConnectorCommandRecord, before.ConnectorCommandRecord) || after.PlanHash != before.PlanHash {
		t.Fatal("continuation changed sealed provider record")
	}
	if calls != 0 {
		t.Fatal("plan/preview reached provider")
	}
	if strings.Contains(before.DestinationConfig["rate_limit_account"], "example.invalid") {
		t.Fatal("provider config became rate cohort")
	}
}
