package engine

import (
	"errors"
	"testing"
	"testing/fstest"
)

// The metadata policy scanner does not load CLI citation artifacts. The real
// execution loader still rejects malformed cli_surface before constructing a
// connector or obtaining credentials, so scanner independence loses no gate.
func TestExecutionCLISurfaceStillRejectsMalformedJSON210(t *testing.T) {
	for _, tc := range []struct {
		name, content string
		valid         bool
	}{
		{"healthy execution surface", `{"usage":"pm acme <command>","tagline":"Acme","commands":[]}`, true},
		{"malformed execution JSON", `{`, false},
		{"malformed historical citation", `{"source_cli":`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := fullValidBundleFS("acme")
			fixture["acme/cli_surface.json"] = &fstest.MapFile{Data: []byte(tc.content)}
			bundle, err := Load(fixture, "acme")
			if tc.valid {
				if err != nil || bundle.CLISurface == nil || bundle.CLISurface.Usage != "pm acme <command>" || len(bundle.Streams) != 1 {
					t.Fatalf("healthy execution surface was not loaded: %v", err)
				}
				return
			}
			var diagnostic *BundleDiagnosticError
			if !errors.As(err, &diagnostic) || diagnostic.File != "cli_surface.json" || bundle.Name != "" {
				t.Fatalf("malformed execution surface passed loader boundary: %v", err)
			}
		})
	}
}
