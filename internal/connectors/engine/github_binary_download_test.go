package engine

import (
	"testing"

	"polymetrics.ai/internal/connectors/defs"
)

// GitHub's archive endpoints redirect from api.github.com to codeload.github.com.
// The binary executor refuses cross-host hops unless the individual operation
// declares an allowlist, so these production operations must retain the narrow
// codeload declaration rather than silently falling back to allow_cross_host.
func TestGitHubArchiveDownloadsAllowOnlyCodeloadRedirect(t *testing.T) {
	bundle, err := Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("Load(defs.FS, github): %v", err)
	}

	for _, operationID := range []string{"github.tarball_ref", "github.zipball_ref"} {
		operationID := operationID
		t.Run(operationID, func(t *testing.T) {
			op, err := findOperation(bundle, operationID)
			if err != nil {
				t.Fatalf("findOperation(%q): %v", operationID, err)
			}
			if op.Binary == nil {
				t.Fatal("archive operation has no binary declaration")
			}
			if op.Binary.AllowCrossHost {
				t.Fatal("archive download must not allow every cross-host redirect")
			}
			if got := op.Binary.AllowedHosts; len(got) != 1 || got[0] != "codeload.github.com" {
				t.Fatalf("allowed_hosts = %v, want [codeload.github.com]", got)
			}
		})
	}
}
