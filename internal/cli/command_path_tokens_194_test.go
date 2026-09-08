package cli

import (
	"polymetrics.ai/internal/connectors/commandrunner"
	"reflect"
	"strings"
	"testing"
)

func TestCommandPathPositionalTokens194(t *testing.T) {
	for _, path := range []string{"widgets fixed", "widgets --help", "widgets --root", "widgets --json", "--widgets fixed", "widgets child --provider-field", "--help widgets", "widgets dotted.name", "widgets widget-id", "widgets widget_id", "widgets -literal"} {
		t.Run(path, func(t *testing.T) {
			expected := strings.Split(path, " ")
			root, jsonOut, clean := parseGlobal(expected)
			parsed := parseFlags(clean)
			same := reflect.DeepEqual(parsed.values["_"], expected)
			t.Logf("path=%q root=%q json=%v positional=%q expected=%q equal=%v", path, root, jsonOut, parsed.values["_"], expected, same)
			_, err := commandrunner.CommandPathSegments(path)
			if !strings.Contains(path, "--") {
				if !same || err != nil {
					t.Fatalf("healthy path mismatch: %v", err)
				}
			} else {
				if same {
					t.Error("flag-like path unexpectedly retained all positional segments")
				}
				if err == nil {
					t.Error("flag-like command path admitted by CommandPathSegments")
				}
			}
		})
	}
}
