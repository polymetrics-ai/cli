package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

var migratedAPIDocs = []string{
	"apify-dataset", "ashby", "aws-cloudtrail", "babelforce", "basecamp", "bunny-inc", "canny", "copper", "dixa", "fastbill", "feishu", "free-agent-connector", "freightview", "google-analytics-data-api", "google-classroom", "google-pagespeed-insights", "less-annoying-crm", "lokalise", "mendeley", "mercado-ads", "metabase", "mode", "my-hours", "pocket", "prestashop", "rootly", "safetyculture", "yahoo-finance-price",
}

var documentedBodyField = regexp.MustCompile("`([^`]+)`=")

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
				streamDocs, ok := streamDocsBlock(content, stream.Name)
				if !ok {
					t.Fatalf("%s docs omit rendered stream %q", connector, stream.Name)
				}
				want := fmt.Sprintf("- `%s`: `%s %s`; records `%s`", stream.Name, stream.Method, stream.Path, stream.Records.Path)
				if !strings.Contains(streamDocs, want) {
					t.Fatalf("%s docs omit rendered stream %q", connector, want)
				}
				if err := assertDocumentedStreamBody(streamDocs, stream); err != nil {
					t.Fatalf("%s stream %q: %v", connector, stream.Name, err)
				}
				pagination := stream.Pagination
				paginationDocs := streamDocs
				paginationLabel := "  - Pagination:"
				if pagination == nil {
					pagination = bundle.HTTP.Pagination
					paginationDocs = content
					paginationLabel = "Default stream pagination:"
				}
				if pagination != nil {
					wantPagination := fmt.Sprintf("%s `%s`.", paginationLabel, pagination.Type)
					if !strings.Contains(paginationDocs, wantPagination) {
						t.Fatalf("%s docs omit rendered pagination %q for stream %q", connector, wantPagination, stream.Name)
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

func streamDocsBlock(document, streamName string) (string, bool) {
	start := strings.Index(document, "- `"+streamName+"`:")
	if start < 0 {
		return "", false
	}
	end := len(document)
	if next := strings.Index(document[start+1:], "\n- `"); next >= 0 {
		end = start + 1 + next
	}
	return document[start:end], true
}

func assertDocumentedStreamBody(document string, stream engine.StreamSpec) error {
	bodyPrefix := "  - JSON body:"
	if stream.BodyType == "form" {
		bodyPrefix = "  - Form body:"
	}
	var bodyLine string
	for _, line := range strings.Split(document, "\n") {
		if strings.HasPrefix(line, bodyPrefix) {
			bodyLine = line
			break
		}
	}
	if len(stream.Body) == 0 {
		if bodyLine != "" {
			return fmt.Errorf("docs describe an undeclared %s", strings.TrimSuffix(strings.TrimPrefix(bodyPrefix, "  - "), ":"))
		}
		return nil
	}
	if bodyLine == "" {
		return fmt.Errorf("docs omit declared %s", strings.TrimSuffix(strings.TrimPrefix(bodyPrefix, "  - "), ":"))
	}
	documented := documentedBodyField.FindAllStringSubmatch(bodyLine, -1)
	if len(documented) != len(stream.Body) {
		return fmt.Errorf("docs body fields = %d, want %d", len(documented), len(stream.Body))
	}
	seen := make(map[string]struct{}, len(documented))
	for _, match := range documented {
		name := match[1]
		if _, declared := stream.Body[name]; !declared {
			return fmt.Errorf("docs body includes undeclared field %q", name)
		}
		if _, duplicate := seen[name]; duplicate {
			return fmt.Errorf("docs body repeats field %q", name)
		}
		seen[name] = struct{}{}
	}
	for name, value := range stream.Body {
		if _, ok := seen[name]; !ok {
			return fmt.Errorf("docs body omits field %q", name)
		}
		if expected, exact := documentedScalarBodyValue(value); exact && !strings.Contains(bodyLine, "`"+name+"`="+expected) {
			return fmt.Errorf("docs body field %q does not preserve declared value %s", name, expected)
		}
	}
	return nil
}

func documentedScalarBodyValue(value any) (string, bool) {
	switch value := value.(type) {
	case string:
		return value, true
	case bool, float64, json.Number:
		return fmt.Sprint(value), true
	case map[string]any:
		template, ok := value["template"].(string)
		return template, ok
	default:
		return "", false
	}
}
