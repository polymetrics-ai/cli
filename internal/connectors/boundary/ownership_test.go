package boundary

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestOwnershipScopeRequiresExactlyOneConnectorSlug(t *testing.T) {
	root := newFixtureRepo(t, map[string]string{
		"connector-scope.json": `{"api_version":"polymetrics.ai/v1","kind":"ConnectorImplementationScope","connectors":["github","gong"]}`,
	})

	_, err := ValidateOwnership(root, OwnershipOptions{
		ScopeFile:    "connector-scope.json",
		ChangedPaths: []string{"internal/connectors/defs/github/metadata.json"},
		Now:          fixedNow,
	})
	var cfgErr *ConfigError
	if !errors.As(err, &cfgErr) {
		t.Fatalf("ValidateOwnership error = %v, want ConfigError", err)
	}
	if !strings.Contains(err.Error(), "exactly one connector") {
		t.Fatalf("scope error = %q, want exactly-one connector message", err)
	}
}

func TestOwnershipAutoDetectsSingleTargetWithoutScopeFile(t *testing.T) {
	root := newFixtureRepo(t, nil)

	report, err := ValidateOwnership(root, OwnershipOptions{
		ChangedPaths: []string{
			"internal/connectors/defs/github/metadata.json",
			"docs/connectors/github/MANUAL.md",
		},
		Now: fixedNow,
	})
	if err != nil {
		t.Fatalf("ValidateOwnership: %v", err)
	}
	if report.TargetConnector != "github" {
		t.Fatalf("TargetConnector = %q, want github", report.TargetConnector)
	}
	if !slices.Equal(report.InferredConnectors, []string{"github"}) {
		t.Fatalf("InferredConnectors = %v, want [github]", report.InferredConnectors)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected clean ownership report, got %+v", report.Findings)
	}
}

func TestOwnershipRejectsSharedRuntimeWhenConnectorScopeInferred(t *testing.T) {
	root := newFixtureRepo(t, nil)

	report, err := ValidateOwnership(root, OwnershipOptions{
		ChangedPaths: []string{
			"internal/connectors/defs/github/streams.json",
			"internal/connectors/engine/read.go",
		},
		Now: fixedNow,
	})
	if err != nil {
		t.Fatalf("ValidateOwnership: %v", err)
	}
	requireOwnershipFinding(t, report, RuleOwnershipSharedPath, "github", "internal/connectors/engine/read.go")
}

func TestOwnershipRejectsUnrelatedConnectorDefinitions(t *testing.T) {
	root := newFixtureRepo(t, map[string]string{
		"connector-scope.json": `{"api_version":"polymetrics.ai/v1","kind":"ConnectorImplementationScope","connectors":["github"]}`,
	})

	report, err := ValidateOwnership(root, OwnershipOptions{
		ScopeFile: "connector-scope.json",
		ChangedPaths: []string{
			"internal/connectors/defs/github/metadata.json",
			"internal/connectors/defs/gong/streams.json",
		},
		Now: fixedNow,
	})
	if err != nil {
		t.Fatalf("ValidateOwnership: %v", err)
	}
	requireOwnershipFinding(t, report, RuleOwnershipUnrelatedConnector, "gong", "internal/connectors/defs/gong/streams.json")
}

func TestOwnershipRejectsUnrelatedGeneratedConnectorDocsAndWebsite(t *testing.T) {
	root := newFixtureRepo(t, map[string]string{
		"connector-scope.json": `{"api_version":"polymetrics.ai/v1","kind":"ConnectorImplementationScope","connectors":["github"]}`,
	})

	report, err := ValidateOwnership(root, OwnershipOptions{
		ScopeFile: "connector-scope.json",
		ChangedPaths: []string{
			"internal/connectors/defs/github/metadata.json",
			"docs/connectors/gong/MANUAL.md",
			"website/public/connectors/icons/gong.svg",
		},
		Now: fixedNow,
	})
	if err != nil {
		t.Fatalf("ValidateOwnership: %v", err)
	}
	requireOwnershipFinding(t, report, RuleOwnershipUnrelatedGenerated, "gong", "docs/connectors/gong/MANUAL.md")
	requireOwnershipFinding(t, report, RuleOwnershipUnrelatedGenerated, "gong", "website/public/connectors/icons/gong.svg")
}

func TestOwnershipAllowsTargetDefinitionsFixturesDocsAndNarrowSharedOutputs(t *testing.T) {
	root := newFixtureRepo(t, map[string]string{
		"connector-scope.json": `{"api_version":"polymetrics.ai/v1","kind":"ConnectorImplementationScope","connectors":["github"]}`,
	})

	report, err := ValidateOwnership(root, OwnershipOptions{
		ScopeFile: "connector-scope.json",
		ChangedPaths: []string{
			"internal/connectors/defs/github/metadata.json",
			"internal/connectors/defs/github/fixtures/streams/issues/page_1.json",
			"internal/connectors/hooks/github/hooks.go",
			"internal/connectors/native/github/connector.go",
			"internal/connectors/github/legacy.go",
			"docs/connectors/github/MANUAL.md",
			"docs/connectors/icons/github.svg",
			"website/public/connectors/icons/github.svg",
			"internal/connectors/defs/defs.go",
			"internal/connectors/hooks/hookset/hookset_gen.go",
			"internal/connectors/native/nativeset/nativeset_gen.go",
			"docs/connectors/README.md",
			"docs/cli/connectors.md",
			"internal/cli/testdata/golden_transcripts.json",
			"website/data/connectors.generated.json",
			"website/lib/connectors.generated.ts",
			"website/lib/connectors.catalog.generated.ts",
			"website/lib/connectors.catalog.data.generated.json",
		},
		Now: fixedNow,
	})
	if err != nil {
		t.Fatalf("ValidateOwnership: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected clean ownership report, got %+v", report.Findings)
	}
}

func TestOwnershipRejectsGateConfigAndExceptionEdits(t *testing.T) {
	root := newFixtureRepo(t, map[string]string{
		"connector-scope.json": `{"api_version":"polymetrics.ai/v1","kind":"ConnectorImplementationScope","connectors":["github"]}`,
	})

	report, err := ValidateOwnership(root, OwnershipOptions{
		ScopeFile: "connector-scope.json",
		ChangedPaths: []string{
			"internal/connectors/defs/github/metadata.json",
			DefaultExceptionsPath,
			"cmd/connectorgen/ownership.go",
		},
		Now: fixedNow,
	})
	if err != nil {
		t.Fatalf("ValidateOwnership: %v", err)
	}
	requireOwnershipFinding(t, report, RuleOwnershipGateConfigEdit, "", DefaultExceptionsPath)
	requireOwnershipFinding(t, report, RuleOwnershipGateConfigEdit, "", "cmd/connectorgen/ownership.go")
}

func requireOwnershipFinding(t *testing.T, report OwnershipReport, rule, connector, path string) Finding {
	t.Helper()
	for _, f := range report.Findings {
		if f.Rule == rule && f.Connector == connector && f.Path == path {
			return f
		}
	}
	t.Fatalf("missing ownership finding rule=%s connector=%s path=%s in %+v", rule, connector, path, report.Findings)
	return Finding{}
}
