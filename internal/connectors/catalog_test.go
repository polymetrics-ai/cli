package connectors

import (
	"strings"
	"testing"
)

func TestConnectorCatalogGeneratedBaseline(t *testing.T) {
	catalog := ConnectorCatalog()
	if got, want := len(catalog), 647; got != want {
		t.Fatalf("catalog len = %d, want %d", got, want)
	}
	counts := ConnectorCatalogCounts(catalog)
	if counts.Sources != 591 || counts.Destinations != 56 {
		t.Fatalf("counts = %+v, want 591 sources and 56 destinations", counts)
	}
	if counts.DocsPresent != 647 {
		t.Fatalf("docs_present = %d, want 647", counts.DocsPresent)
	}
	if counts.Enabled != 2 || counts.PlannedNativePort != 645 || counts.UnsupportedDeprecated != 0 {
		t.Fatalf("native enablement counts = %+v, want 2 enabled and 645 planned native ports", counts)
	}
	for _, entry := range catalog {
		if entry.Slug == "" || entry.Name == "" || entry.Type == "" || entry.DocumentationURL == "" || entry.ReleaseStage == "" || entry.SupportLevel == "" {
			t.Fatalf("incomplete entry: %+v", entry)
		}
		if entry.ImplementationStatus == "" || entry.RuntimeKind == "" {
			t.Fatalf("entry missing native support metadata: %+v", entry)
		}
		caps := entry.RuntimeCapabilities
		if !caps.Metadata {
			t.Fatalf("entry missing metadata capability: %+v", entry)
		}
		if entry.ImplementationStatus == ImplementationEnabled {
			if !caps.Check || !caps.Catalog || !caps.ETL {
				t.Fatalf("enabled entry %s missing runtime capabilities: %+v", entry.Slug, caps)
			}
			if caps.UnsupportedReason != "" {
				t.Fatalf("enabled entry has unsupported reason: %+v", entry)
			}
			continue
		}
		if caps.Check || caps.Catalog || caps.Read || caps.Write || caps.Query || caps.ETL || caps.ReverseETL {
			t.Fatalf("planned entry %s exposes executable runtime capabilities: %+v", entry.Slug, caps)
		}
		if caps.UnsupportedReason == "" {
			t.Fatalf("planned entry %s missing unsupported reason: %+v", entry.Slug, caps)
		}
	}
}

func TestConnectorCatalogLookupAndFilters(t *testing.T) {
	github, ok := ConnectorDefinitionBySlug("source-github")
	if !ok {
		t.Fatal("source-github not found")
	}
	if github.ImplementationStatus != ImplementationEnabled || github.RuntimeKind != RuntimeNativeGo || github.PMConnectorName != "github" {
		t.Fatalf("github native metadata = %+v", github)
	}
	githubCaps := github.RuntimeCapabilities
	if !githubCaps.Metadata || !githubCaps.Check || !githubCaps.Catalog || !githubCaps.Read || !githubCaps.Write || !githubCaps.ETL || !githubCaps.ReverseETL {
		t.Fatalf("github runtime capabilities = %+v", githubCaps)
	}
	if githubCaps.Query {
		t.Fatalf("github query capability = true, want false")
	}
	if len(github.SecretFields) == 0 {
		t.Fatalf("github secret fields were not extracted: %+v", github)
	}
	if github.DocumentationURL == "" || !strings.Contains(github.DocumentationURL, "docs.airbyte.com") {
		t.Fatalf("github documentation_url = %q, want Airbyte connector docs", github.DocumentationURL)
	}
	if github.ApplicationDocumentationURL == "" || !strings.Contains(github.ApplicationDocumentationURL, "docs.github.com") {
		t.Fatalf("github application_documentation_url = %q, want GitHub official docs", github.ApplicationDocumentationURL)
	}
	if len(github.OfficialApplicationDocs) == 0 {
		t.Fatalf("github official application docs were not extracted: %+v", github)
	}
	postgres, ok := ConnectorDefinitionBySlug("destination-postgres")
	if !ok {
		t.Fatal("destination-postgres not found")
	}
	postgresCaps := postgres.RuntimeCapabilities
	if !postgresCaps.Metadata || postgresCaps.Check || postgresCaps.Catalog || postgresCaps.Read || postgresCaps.Write || postgresCaps.Query || postgresCaps.ETL || postgresCaps.ReverseETL {
		t.Fatalf("destination-postgres runtime capabilities = %+v", postgresCaps)
	}
	if postgres.ImplementationStatus != ImplementationPlannedNativePort || postgresCaps.UnsupportedReason == "" {
		t.Fatalf("destination-postgres native status = %+v caps=%+v", postgres, postgresCaps)
	}
	filtered := FilterConnectorCatalog(ConnectorCatalog(), ConnectorCatalogFilter{Type: ConnectorTypeDestination, Stage: "generally_available"})
	if len(filtered) != 9 {
		t.Fatalf("GA destinations len = %d, want 9", len(filtered))
	}
	for _, entry := range filtered {
		if entry.Type != ConnectorTypeDestination || entry.ReleaseStage != "generally_available" {
			t.Fatalf("unexpected filtered entry: %+v", entry)
		}
	}
}
