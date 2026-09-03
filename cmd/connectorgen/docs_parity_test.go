package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

var migratedAPIDocs = []string{
	"apify-dataset", "ashby", "aws-cloudtrail", "babelforce", "basecamp", "bunny-inc", "canny", "copper", "dixa", "fastbill", "feishu", "free-agent-connector", "freightview", "google-analytics-data-api", "google-classroom", "mercado-ads", "metabase", "mode", "pocket", "prestashop", "safetyculture", "yahoo-finance-price",
}

func TestMigratedAPIDocsDescribeRenderedStreams(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("repoRoot: %v", err)
	}
	for _, connector := range migratedAPIDocs {
		t.Run(connector, func(t *testing.T) {
			document, err := os.ReadFile(filepath.Join(root, "internal", "connectors", "defs", connector, "docs.md"))
			if err != nil {
				t.Fatal(err)
			}
			content := string(document)
			if strings.Contains(content, "connector-managed request") || strings.Contains(content, "connector-specific implementation") || strings.Contains(content, "native connector owns") {
				t.Fatalf("%s docs retain obsolete executor ownership", connector)
			}
			bundle, err := engine.Load(defs.FS, connector)
			if err != nil {
				t.Fatalf("load rendered bundle: %v", err)
			}
			if bundle.HTTP.Check != nil {
				want := fmt.Sprintf("Connection check: `%s %s`", bundle.HTTP.Check.Method, bundle.HTTP.Check.Path)
				if !strings.Contains(content, want) {
					t.Fatalf("%s docs omit rendered check %q", connector, want)
				}
			}
			for _, stream := range bundle.Streams {
				want := fmt.Sprintf("- `%s`: `%s %s`; records `%s`", stream.Name, stream.Method, stream.Path, stream.Records.Path)
				if !strings.Contains(content, want) {
					t.Fatalf("%s docs omit rendered stream %q", connector, want)
				}
				if len(stream.Body) != 0 && !strings.Contains(content, "  - JSON body:") {
					t.Fatalf("%s docs omit declared JSON body semantics for stream %q", connector, stream.Name)
				}
				if stream.Pagination != nil {
					wantPagination := fmt.Sprintf("  - Pagination: `%s`.", stream.Pagination.Type)
					if !strings.Contains(content, wantPagination) {
						t.Fatalf("%s docs omit rendered pagination %q", connector, wantPagination)
					}
				}
				if stream.ResponseError != nil {
					if stream.ResponseError.SuccessPath != "" {
						wantSuccess := fmt.Sprintf("`success_path`=`%s`", stream.ResponseError.SuccessPath)
						if !strings.Contains(content, wantSuccess) {
							t.Fatalf("%s docs omit declared success envelope %q", connector, wantSuccess)
						}
					}
					if stream.ResponseError.Path != "" {
						wantPath := fmt.Sprintf("`path`=`%s`", stream.ResponseError.Path)
						if !strings.Contains(content, wantPath) {
							t.Fatalf("%s docs omit declared response error %q", connector, wantPath)
						}
					}
				}
			}
			if connector == "ashby" && (strings.Contains(content, "`base_url`") || strings.Contains(content, "`mode`") || strings.Contains(strings.ToLower(content), "fixture-only") || strings.Contains(strings.ToLower(bundle.Metadata.Description), "fixture-only")) {
				t.Fatal("Ashby docs or metadata retain removed caller-origin or fixture configuration")
			}
		})
	}
}
