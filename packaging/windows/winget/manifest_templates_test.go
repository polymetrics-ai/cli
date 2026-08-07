package winget_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const packageIdentifier = "PolymetricsAI.PolymetricsCLI"

func TestManifestTemplatesUseApprovedPackageIdentifier(t *testing.T) {
	t.Parallel()

	for _, name := range []string{
		"PolymetricsAI.PolymetricsCLI.yaml.tmpl",
		"PolymetricsAI.PolymetricsCLI.locale.en-US.yaml.tmpl",
		"PolymetricsAI.PolymetricsCLI.installer.yaml.tmpl",
	} {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			content := readTemplate(t, name)
			if !strings.Contains(content, "PackageIdentifier: "+packageIdentifier) {
				t.Fatalf("%s does not contain approved PackageIdentifier %q", name, packageIdentifier)
			}
		})
	}
}

func TestInstallerTemplateUsesPlaceholdersForUnpublishedSignedHashes(t *testing.T) {
	t.Parallel()

	content := readTemplate(t, "PolymetricsAI.PolymetricsCLI.installer.yaml.tmpl")

	for _, want := range []string{
		"Architecture: x64",
		"InstallerType: wix",
		"Scope: machine",
		"InstallerSha256: <SHA256_OF_FINAL_SIGNED_X64_MSI>",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("installer template missing %q", want)
		}
	}
	// arm64 must be absent, not merely unasserted. pm embeds DuckDB and
	// go-duckdb ships no windows/arm64 library, so no arm64 MSI exists to
	// point at; a manifest advertising one would send WinGet users to a
	// download that was never published.
	for _, unwanted := range []string{
		"Architecture: arm64",
		"windows_arm64.msi",
		"<SHA256_OF_FINAL_SIGNED_ARM64_MSI>",
		"<MSI_PRODUCT_CODE_ARM64>",
	} {
		if strings.Contains(content, unwanted) {
			t.Fatalf("installer template still advertises an unpublished arm64 installer: %q", unwanted)
		}
	}

	realHash := regexp.MustCompile(`(?m)^\s*InstallerSha256:\s*[A-Fa-f0-9]{64}\s*$`)
	if realHash.MatchString(content) {
		t.Fatalf("installer template contains a real InstallerSha256; unpublished signed MSI hashes must stay placeholders")
	}
}

func readTemplate(t *testing.T, name string) string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(".", name))
	if err != nil {
		t.Fatalf("read template %s: %v", name, err)
	}
	return string(data)
}
