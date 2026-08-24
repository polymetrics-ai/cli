package bundleregistry

import (
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

// TestGitHubReleaseUploadGuidanceKeepsTheCompositeAliasHonest protects the
// user-facing distinction between the unsupported gh-style multi-file alias
// and the implemented one-file binary_upload command. A generated manual that
// says no bounded executor exists would make the actual safe command
// undiscoverable even though both rows describe the same provider endpoint.
func TestGitHubReleaseUploadGuidanceKeepsTheCompositeAliasHonest(t *testing.T) {
	registry := New()
	github, ok := registry.Get("github")
	if !ok {
		t.Fatal("github connector is not registered")
	}
	manual := connectors.RenderConnectorManual(github)
	for _, want := range []string{
		"release upload - Upload release assets [intent=direct_write availability=unsupported_local",
		"select multiple local files and clobber or label existing assets",
		"releases assets upload - Upload one release asset as exact binary bytes. [intent=binary_upload availability=implemented",
	} {
		if !strings.Contains(manual, want) {
			t.Fatalf("GitHub manual missing %q", want)
		}
	}
	if strings.Contains(manual, "no bounded binary-upload executor exists") {
		t.Fatalf("GitHub manual denies its implemented bounded upload executor")
	}
}
