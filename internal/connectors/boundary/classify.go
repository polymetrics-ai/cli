package boundary

import (
	"os"
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
	Certify     bool
	Description string
}

func classifyPath(root, rel string) pathClass {
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
	if strings.HasPrefix(rel, "cmd/iconregistrygen/") || rel == "internal/connectors/icons.go" || rel == "internal/cli/connector_docs.go" {
		return pathClass{Class: pathClassAllowedGenerated, Allowed: true}
	}
	if strings.HasPrefix(rel, "internal/connectors/defs/") {
		return pathClass{Class: pathClassAllowedDefs, Allowed: true}
	}
	if isAllowedHookImplementation(rel) {
		return pathClass{Class: pathClassAllowedHook, Allowed: true}
	}
	if isAllowedNativeImplementation(rel) {
		return pathClass{Class: pathClassAllowedNative, Allowed: true}
	}
	if isNativeWiring(rel) {
		return pathClass{Class: pathClassNativeWiring, Allowed: true}
	}
	if strings.HasSuffix(rel, ".go") && isGeneratedGo(root, rel) {
		return pathClass{Class: pathClassAllowedGenerated, Allowed: true}
	}
	if strings.HasPrefix(rel, "internal/") && !strings.HasPrefix(rel, "internal/connectors/") && !strings.HasPrefix(rel, "internal/cli/") && !strings.HasPrefix(rel, "internal/app/") {
		return pathClass{Class: pathClassIgnored, Allowed: true}
	}
	if isLegacyConnectorPackage(rel) {
		return pathClass{Class: pathClassAllowedNative, Allowed: true}
	}
	if !strings.HasSuffix(rel, ".go") {
		return pathClass{Class: pathClassIgnored, Allowed: true}
	}
	if strings.HasPrefix(rel, "cmd/") || strings.HasPrefix(rel, "internal/") {
		pc := pathClass{Class: pathClassSharedProduction, ScanGo: true}
		pc.Certify = strings.HasPrefix(rel, "internal/connectors/certify/")
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

func isAllowedHookImplementation(rel string) bool {
	const prefix = "internal/connectors/hooks/"
	if !strings.HasPrefix(rel, prefix) {
		return false
	}
	rest := strings.TrimPrefix(rel, prefix)
	name, _, ok := strings.Cut(rest, "/")
	return ok && name != "hookset" && name != ""
}

func isAllowedNativeImplementation(rel string) bool {
	const prefix = "internal/connectors/native/"
	if !strings.HasPrefix(rel, prefix) {
		return false
	}
	rest := strings.TrimPrefix(rel, prefix)
	name, _, ok := strings.Cut(rest, "/")
	return ok && name != "nativeset" && name != ""
}

func isNativeWiring(rel string) bool {
	return rel == "internal/connectors/native/nativeset/factories.go" ||
		rel == "internal/connectors/native/nativeset/promoted.go"
}

func isLegacyConnectorPackage(rel string) bool {
	const prefix = "internal/connectors/"
	if !strings.HasPrefix(rel, prefix) {
		return false
	}
	rest := strings.TrimPrefix(rel, prefix)
	name, _, ok := strings.Cut(rest, "/")
	if !ok || name == "" {
		return false
	}
	switch name {
	case "boundary", "certify", "connsdk", "commandrunner", "defs", "engine", "hooks", "native", "bundleregistry", "conformance":
		return false
	default:
		return strings.HasSuffix(rel, ".go")
	}
}

func isGeneratedGo(root, rel string) bool {
	if strings.HasSuffix(rel, "_gen.go") {
		return true
	}
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return false
	}
	if len(b) > 2048 {
		b = b[:2048]
	}
	return strings.Contains(string(b), "Code generated")
}
