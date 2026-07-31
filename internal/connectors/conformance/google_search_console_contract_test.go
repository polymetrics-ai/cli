package conformance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestGoogleSearchConsoleOfficialParityContracts(t *testing.T) {
	b := loadTestBundle(t, "../defs", "google-search-console")

	expectedOfficial := map[string]string{
		"GET /webmasters/v3/sites":                                  "webmasters.sites.list",
		"GET /webmasters/v3/sites/{siteUrl}":                        "webmasters.sites.get",
		"GET /webmasters/v3/sites/{siteUrl}/sitemaps":               "webmasters.sitemaps.list",
		"GET /webmasters/v3/sites/{siteUrl}/sitemaps/{feedpath}":    "webmasters.sitemaps.get",
		"PUT /webmasters/v3/sites/{siteUrl}":                        "webmasters.sites.add",
		"DELETE /webmasters/v3/sites/{siteUrl}":                     "webmasters.sites.delete",
		"PUT /webmasters/v3/sites/{siteUrl}/sitemaps/{feedpath}":    "webmasters.sitemaps.submit",
		"DELETE /webmasters/v3/sites/{siteUrl}/sitemaps/{feedpath}": "webmasters.sitemaps.delete",
		"POST /webmasters/v3/sites/{siteUrl}/searchAnalytics/query": "webmasters.searchanalytics.query",
		"POST /v1/urlInspection/index:inspect":                      "searchconsole.urlInspection.index.inspect",
		"POST /v1/urlTestingTools/mobileFriendlyTest:run":           "searchconsole.urlTestingTools.mobileFriendlyTest.run",
	}

	if b.Surface == nil {
		t.Fatal("api_surface.json missing")
	}
	seen := map[string]bool{}
	for _, ep := range b.Surface.Endpoints {
		if ep.Excluded != nil {
			t.Fatalf("google-search-console must not exclude official operations after completed inventory gate: %s %s", ep.Method, ep.Path)
		}
		key := strings.ToUpper(strings.TrimSpace(ep.Method)) + " " + ep.Path
		if _, ok := expectedOfficial[key]; ok {
			seen[key] = true
		}
	}
	for key, opID := range expectedOfficial {
		if !seen[key] {
			t.Fatalf("official operation %s (%s) missing from api_surface.json", opID, key)
		}
	}

	for _, spec := range []struct {
		id     string
		method string
		path   string
	}{
		{"google-search-console.searchanalytics_query", "POST", "/webmasters/v3/sites/{siteUrl}/searchAnalytics/query"},
		{"google-search-console.urlinspection_index_inspect", "POST", "/v1/urlInspection/index:inspect"},
		{"google-search-console.mobile_friendly_test_run", "POST", "/v1/urlTestingTools/mobileFriendlyTest:run"},
	} {
		op, ok := googleSearchConsoleOperation(b, spec.id)
		if !ok {
			t.Fatalf("operation %q missing from operations.json", spec.id)
		}
		if op.Kind != "rest_read" || op.REST == nil || op.REST.Method != spec.method || op.REST.Path != spec.path {
			t.Fatalf("operation %q REST = kind:%s rest:%+v, want %s %s", spec.id, op.Kind, op.REST, spec.method, spec.path)
		}
		if op.REST.MaxBytes <= 0 || op.REST.MaxBytes > 1<<20 {
			t.Fatalf("operation %q max_bytes = %d, want bounded <= 1MiB", spec.id, op.REST.MaxBytes)
		}
		if !strings.EqualFold(op.REST.ContentType, "application/json") || len(op.REST.BodySchema) == 0 {
			t.Fatalf("operation %q must be a schema-gated application/json POST read", spec.id)
		}
	}

	for _, spec := range []struct {
		path string
		op   string
	}{
		{"direct search-analytics query", "google-search-console.searchanalytics_query"},
		{"direct url-inspection inspect", "google-search-console.urlinspection_index_inspect"},
		{"direct mobile-friendly-test run", "google-search-console.mobile_friendly_test_run"},
	} {
		cmd, ok := googleSearchConsoleCommand(b, spec.path)
		if !ok {
			t.Fatalf("command %q missing from cli_surface.json", spec.path)
		}
		if cmd.Intent != "direct_read" || cmd.Availability != "implemented" || cmd.Operation != spec.op || cmd.OutputPolicy != "json_redacted" {
			t.Fatalf("command %q = intent:%s availability:%s operation:%s output:%s", spec.path, cmd.Intent, cmd.Availability, cmd.Operation, cmd.OutputPolicy)
		}
		if len(cmd.APISurface) != 1 || strings.ToUpper(cmd.APISurface[0].Method) != "POST" {
			t.Fatalf("command %q api_surface = %+v, want one POST endpoint", spec.path, cmd.APISurface)
		}
	}

	for _, actionName := range []string{"delete_site", "delete_sitemap"} {
		action, ok := googleSearchConsoleWrite(b, actionName)
		if !ok {
			t.Fatalf("write %q missing", actionName)
		}
		if action.Confirm != "destructive" || action.Delete == nil || !action.Delete.Idempotent || !containsInt(action.Delete.MissingOkStatus, 404) {
			t.Fatalf("write %q destructive/idempotent policy incomplete: confirm=%q delete=%+v", actionName, action.Confirm, action.Delete)
		}
		if !containsStringSlice(action.RedactFields, "site_url") {
			t.Fatalf("write %q redact_fields = %v, want site_url", actionName, action.RedactFields)
		}
		if actionName == "delete_sitemap" && !containsStringSlice(action.RedactFields, "feedpath") {
			t.Fatalf("write %q redact_fields = %v, want feedpath", actionName, action.RedactFields)
		}
	}

	for _, stream := range []string{"sites", "site_details", "sitemaps", "sitemap_details", "search_analytics_by_date"} {
		fixture := filepath.Join("..", "defs", "google-search-console", "fixtures", "streams", stream, "page_1.json")
		if _, err := os.Stat(fixture); err != nil {
			t.Fatalf("fixture for stream %q missing at %s: %v", stream, fixture, err)
		}
	}

	if b.Certification == nil {
		t.Fatal("certification.json missing; fixture-only certification metadata must exist without claiming live certification")
	}
	if b.Certification.Source.DefaultStream != "sites" {
		t.Fatalf("source.default_stream = %q, want sites", b.Certification.Source.DefaultStream)
	}
	if len(b.Certification.DirectReadCandidates) != 3 {
		t.Fatalf("direct_read_candidates = %d, want 3", len(b.Certification.DirectReadCandidates))
	}
}

func googleSearchConsoleOperation(b engine.Bundle, id string) (engine.OperationSpec, bool) {
	for _, op := range b.Operations {
		if op.ID == id {
			return op, true
		}
	}
	return engine.OperationSpec{}, false
}

func googleSearchConsoleCommand(b engine.Bundle, path string) (engine.CLICommand, bool) {
	if b.CLISurface == nil {
		return engine.CLICommand{}, false
	}
	for _, cmd := range b.CLISurface.Commands {
		if cmd.Path == path {
			return cmd, true
		}
	}
	return engine.CLICommand{}, false
}

func googleSearchConsoleWrite(b engine.Bundle, name string) (engine.WriteAction, bool) {
	for _, action := range b.Writes {
		if action.Name == name {
			return action, true
		}
	}
	return engine.WriteAction{}, false
}

func containsInt(values []int, want int) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestGoogleSearchConsoleDirectReadBodySchemasAreClosed(t *testing.T) {
	b := loadTestBundle(t, "../defs", "google-search-console")
	for _, id := range []string{
		"google-search-console.searchanalytics_query",
		"google-search-console.urlinspection_index_inspect",
		"google-search-console.mobile_friendly_test_run",
	} {
		op, ok := googleSearchConsoleOperation(b, id)
		if !ok || op.REST == nil {
			t.Fatalf("operation %q missing", id)
		}
		var schema struct {
			AdditionalProperties *bool `json:"additionalProperties"`
		}
		if err := json.Unmarshal(op.REST.BodySchema, &schema); err != nil {
			t.Fatalf("decode %q body_schema: %v", id, err)
		}
		if schema.AdditionalProperties == nil || *schema.AdditionalProperties {
			t.Fatalf("operation %q body_schema must declare additionalProperties=false", id)
		}
	}
}
