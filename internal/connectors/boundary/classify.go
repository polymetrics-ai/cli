package boundary

import (
	"path/filepath"
	"strings"
)

const (
	pathClassSharedProduction = "shared_production_go"
	pathClassAllowedDefs      = "defs"
	pathClassAllowedHook      = "hook"
	pathClassAllowedNative    = "native"
	pathClassAllowedGenerated = "generated"
	pathClassAllowedTest      = "test_fixture"
	pathClassDocs             = "documentation"
	pathClassIgnored          = "ignored"
	pathClassNativeWiring     = "native_wiring"
)

type pathClass struct {
	Class       string
	ScanGo      bool
	Allowed     bool
	DocsOutput  bool
	Description string
}

func classifyPath(rel string, lx lexicon) pathClass {
	rel = normalizeRelPath(rel)
	base := filepath.Base(rel)
	if rel == "" || strings.HasPrefix(rel, ".git/") || strings.Contains(rel, "/.git/") || strings.HasPrefix(rel, "vendor/") || strings.Contains(rel, "/node_modules/") {
		return pathClass{Class: pathClassIgnored, Allowed: true}
	}
	if strings.HasSuffix(base, "_test.go") || strings.Contains(rel, "/testdata/") || strings.HasPrefix(rel, "testdata/") {
		return pathClass{Class: pathClassAllowedTest, Allowed: true}
	}
	if strings.HasPrefix(rel, "docs/") || strings.HasPrefix(rel, "website/") || strings.HasPrefix(rel, ".planning/") || strings.HasPrefix(rel, ".agents/") || rel == "README.md" || rel == "CHANGELOG.md" {
		return pathClass{Class: pathClassDocs, Allowed: true}
	}
	if strings.HasPrefix(rel, "cmd/iconregistrygen/") || rel == "internal/connectors/icons.go" {
		return pathClass{Class: pathClassAllowedGenerated, Allowed: true}
	}
	if strings.HasPrefix(rel, "internal/connectors/defs/") {
		return pathClass{Class: pathClassAllowedDefs, Allowed: true}
	}
	if isAllowedHookImplementation(rel, lx) {
		return pathClass{Class: pathClassAllowedHook, Allowed: true}
	}
	if isAllowedNativeImplementation(rel, lx) {
		return pathClass{Class: pathClassAllowedNative, Allowed: true}
	}
	if isNativeWiring(rel) {
		return pathClass{Class: pathClassNativeWiring, Allowed: true}
	}
	if isKnownGeneratedGo(rel) {
		return pathClass{Class: pathClassAllowedGenerated, Allowed: true}
	}
	if isLegacyConnectorPackage(rel, lx) {
		return pathClass{Class: pathClassAllowedNative, Allowed: true}
	}
	if !strings.HasSuffix(rel, ".go") {
		return pathClass{Class: pathClassIgnored, Allowed: true}
	}
	if strings.HasPrefix(rel, "cmd/") || strings.HasPrefix(rel, "internal/") {
		pc := pathClass{Class: pathClassSharedProduction, ScanGo: true}
		pc.DocsOutput = rel == "internal/connectors/guide.go" || rel == "internal/cli/docs.go" || rel == "internal/cli/skills.go" || rel == "internal/cli/connector_docs.go"
		return pc
	}
	return pathClass{Class: pathClassIgnored, Allowed: true}
}

func normalizeRelPath(path string) string {
	path = filepath.ToSlash(path)
	path = strings.TrimPrefix(path, "./")
	return strings.Trim(path, "/")
}

func isAllowedHookImplementation(rel string, lx lexicon) bool {
	const prefix = "internal/connectors/hooks/"
	if !strings.HasPrefix(rel, prefix) {
		return false
	}
	rest := strings.TrimPrefix(rel, prefix)
	name, _, ok := strings.Cut(rest, "/")
	if !ok || name == "hookset" || name == "" {
		return false
	}
	_, ok = lx.byName[name]
	return ok
}

func isAllowedNativeImplementation(rel string, lx lexicon) bool {
	const prefix = "internal/connectors/native/"
	if !strings.HasPrefix(rel, prefix) {
		return false
	}
	rest := strings.TrimPrefix(rel, prefix)
	name, _, ok := strings.Cut(rest, "/")
	if !ok || name == "nativeset" || name == "" {
		return false
	}
	_, ok = lx.byName[name]
	return ok
}

func isNativeWiring(rel string) bool {
	return rel == "internal/connectors/native/nativeset/factories.go" ||
		rel == "internal/connectors/native/nativeset/promoted.go"
}

func isLegacyConnectorPackage(rel string, lx lexicon) bool {
	const prefix = "internal/connectors/"
	if !strings.HasPrefix(rel, prefix) {
		return false
	}
	rest := strings.TrimPrefix(rel, prefix)
	name, _, ok := strings.Cut(rest, "/")
	if !ok || name == "" {
		return false
	}
	_, ok = lx.byName[name]
	return ok && strings.HasSuffix(rel, ".go")
}

func isKnownGeneratedGo(rel string) bool {
	switch rel {
	case "internal/connectors/hooks/hookset/hookset_gen.go",
		"internal/connectors/manifestindex/index_gen.go",
		"internal/connectors/native/nativeset/nativeset_gen.go":
		return true
	default:
		return false
	}
}
