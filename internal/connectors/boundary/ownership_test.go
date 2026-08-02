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

func TestOwnershipChangedPathCollectionIncludesDeletions(t *testing.T) {
	modes := []struct {
		name    string
		baseRef string
	}{
		{name: "worktree"},
		{name: "base", baseRef: "HEAD"},
	}
	for _, mode := range modes {
		t.Run(mode.name+"/unrelated_connector_deletion_rejected", func(t *testing.T) {
			root := newOwnershipDeletionRepo(t)
			runGit(t, root, "rm", "internal/connectors/defs/gong/streams.json")

			report, err := ValidateOwnership(root, OwnershipOptions{
				ScopeFile: "connector-scope.json",
				BaseRef:   mode.baseRef,
				Now:       fixedNow,
			})
			if err != nil {
				t.Fatalf("ValidateOwnership: %v", err)
			}
			requireOwnershipFinding(t, report, RuleOwnershipUnrelatedConnector, "gong", "internal/connectors/defs/gong/streams.json")
			requireOwnershipPath(t, report, "internal/connectors/defs/gong/streams.json", ownershipClassConnectorDefs, "gong", ownershipDecisionRejected)
		})

		t.Run(mode.name+"/target_connector_deletion_allowed", func(t *testing.T) {
			root := newOwnershipDeletionRepo(t)
			runGit(t, root, "rm", "internal/connectors/defs/github/streams.json")

			report, err := ValidateOwnership(root, OwnershipOptions{
				ScopeFile: "connector-scope.json",
				BaseRef:   mode.baseRef,
				Now:       fixedNow,
			})
			if err != nil {
				t.Fatalf("ValidateOwnership: %v", err)
			}
			if len(report.Findings) != 0 {
				t.Fatalf("expected clean ownership report, got %+v", report.Findings)
			}
			requireOwnershipPath(t, report, "internal/connectors/defs/github/streams.json", ownershipClassConnectorDefs, "github", ownershipDecisionAllowed)
		})
	}
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

func TestOwnershipAllowsCompactGeneratedIconAlias(t *testing.T) {
	root := newFixtureRepo(t, map[string]string{
		"connector-scope.json":                              `{"api_version":"polymetrics.ai/v1","kind":"ConnectorImplementationScope","connectors":["amazon-ads"]}`,
		"internal/connectors/defs/amazon-ads/metadata.json": `{"name":"amazon-ads","display_name":"Amazon Ads","integration_type":"api"}`,
	})

	report, err := ValidateOwnership(root, OwnershipOptions{
		ScopeFile: "connector-scope.json",
		ChangedPaths: []string{
			"docs/connectors/icons/amazonads.svg",
			"website/public/connectors/icons/amazonads.svg",
		},
		Now: fixedNow,
	})
	if err != nil {
		t.Fatalf("ValidateOwnership: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected clean ownership report, got %+v", report.Findings)
	}
	requireOwnershipPath(t, report, "docs/connectors/icons/amazonads.svg", ownershipClassConnectorIcon, "amazon-ads", ownershipDecisionAllowed)
	requireOwnershipPath(t, report, "website/public/connectors/icons/amazonads.svg", ownershipClassConnectorWebsiteIcon, "amazon-ads", ownershipDecisionAllowed)
}

func TestOwnershipDoesNotResolveCollidingCompactIconAliases(t *testing.T) {
	root := newFixtureRepo(t, map[string]string{
		"connector-scope.json":                        `{"api_version":"polymetrics.ai/v1","kind":"ConnectorImplementationScope","connectors":["a-bc"]}`,
		"internal/connectors/defs/a-bc/metadata.json": `{"name":"a-bc","display_name":"A BC","integration_type":"api"}`,
		"internal/connectors/defs/ab-c/metadata.json": `{"name":"ab-c","display_name":"AB C","integration_type":"api"}`,
	})

	report, err := ValidateOwnership(root, OwnershipOptions{
		ScopeFile:    "connector-scope.json",
		ChangedPaths: []string{"docs/connectors/icons/abc.svg"},
		Now:          fixedNow,
	})
	if err != nil {
		t.Fatalf("ValidateOwnership: %v", err)
	}
	requireOwnershipFinding(t, report, RuleOwnershipSharedPath, "a-bc", "docs/connectors/icons/abc.svg")
	requireOwnershipPath(t, report, "docs/connectors/icons/abc.svg", ownershipClassSharedDocs, "", ownershipDecisionRejected)
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

func TestOwnershipRejectsTopLevelProjectConfigEdits(t *testing.T) {
	root := newFixtureRepo(t, map[string]string{
		"connector-scope.json": `{"api_version":"polymetrics.ai/v1","kind":"ConnectorImplementationScope","connectors":["github"]}`,
	})
	configPaths := []string{
		".golangci.yml",
		".gitlab-ci.yml",
		".goreleaser.yaml",
		".gitignore",
	}

	report, err := ValidateOwnership(root, OwnershipOptions{
		ScopeFile:    "connector-scope.json",
		ChangedPaths: configPaths,
		Now:          fixedNow,
	})
	if err != nil {
		t.Fatalf("ValidateOwnership: %v", err)
	}
	for _, path := range configPaths {
		requireOwnershipFinding(t, report, RuleOwnershipSharedPath, "github", path)
		requireOwnershipPath(t, report, path, ownershipClassSharedRepo, "", ownershipDecisionRejected)
	}
}

func newOwnershipDeletionRepo(t *testing.T) string {
	t.Helper()
	root := newFixtureRepo(t, map[string]string{
		"connector-scope.json":                         `{"api_version":"polymetrics.ai/v1","kind":"ConnectorImplementationScope","connectors":["github"]}`,
		"internal/connectors/defs/github/streams.json": `{"streams":[]}`,
		"internal/connectors/defs/gong/streams.json":   `{"streams":[]}`,
	})
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "Test User")
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-m", "base")
	return root
}

func requireOwnershipPath(t *testing.T, report OwnershipReport, path, class, connector, decision string) OwnershipPath {
	t.Helper()
	for _, changed := range report.ChangedPaths {
		if changed.Path == path && changed.Class == class && changed.Connector == connector && changed.Decision == decision {
			return changed
		}
	}
	t.Fatalf("missing ownership path path=%s class=%s connector=%s decision=%s in %+v", path, class, connector, decision, report.ChangedPaths)
	return OwnershipPath{}
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
