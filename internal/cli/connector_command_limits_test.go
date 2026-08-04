package cli

import (
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/commandrunner"
)

// TestConnectorCommandMaxBytesDoesNotImposeAnIntentsCeiling pins the rule that
// the CLI resolves --max-bytes and nothing else.
//
// It used to default and clamp to the direct-read ceiling and hand that value
// to every intent, which capped binary downloads at 16 MiB against operations
// declaring 100 MiB -- and made the flag unable to reach the declared cap even
// when typed explicitly, contradicting the help text's promise that it only
// ever lowers.
func TestConnectorCommandMaxBytesDoesNotImposeAnIntentsCeiling(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want int
	}{
		{
			// Zero means "unset", which is how each runner is told to apply
			// its own default rather than the CLI's.
			name: "unset",
			args: []string{"artifact", "download"},
			want: 0,
		},
		{
			name: "above the direct read ceiling is passed through",
			args: []string{"artifact", "download", "--max-bytes", "104857600"},
			want: 104857600,
		},
		{
			name: "below the direct read ceiling is passed through",
			args: []string{"issues", "list", "--max-bytes", "2048"},
			want: 2048,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := connectorCommandMaxBytes(parseFlags(tt.args))
			if err != nil {
				t.Fatalf("connectorCommandMaxBytes: %v", err)
			}
			if got != tt.want {
				t.Fatalf("connectorCommandMaxBytes = %d, want %d", got, tt.want)
			}
		})
	}

	if got := commandrunner.MaxOperationDirectReadBytes; got >= 104857600 {
		t.Fatalf("direct read ceiling = %d, no longer below the declared binary cap this test relies on", got)
	}
}

func TestConnectorCommandMaxBytesRejectsNegative(t *testing.T) {
	if _, err := connectorCommandMaxBytes(parseFlags([]string{"x", "--max-bytes", "-1"})); err == nil {
		t.Fatal("connectorCommandMaxBytes error = nil, want a rejection of a negative limit")
	}
}

// TestConnectorDownloadFlagsMatchTheSharedDeclaration guards the parity the
// shared declaration exists for: runtime help must document exactly the flags
// the generated manual, skill and website docs document.
func TestConnectorDownloadFlagsMatchTheSharedDeclaration(t *testing.T) {
	var help strings.Builder
	writeConnectorDownloadFlags(&help, connectors.CommandSurfaceCommand{Intent: "binary_download"})
	rendered := help.String()
	for _, flag := range connectors.BinaryDownloadFlags() {
		if !strings.Contains(rendered, "--"+flag.Name) {
			t.Fatalf("runtime download help does not document --%s:\n%s", flag.Name, rendered)
		}
	}
	if !strings.Contains(rendered, "--dest-root (string) required") {
		t.Fatalf("runtime download help does not mark --dest-root required:\n%s", rendered)
	}

	var passive strings.Builder
	writeConnectorDownloadFlags(&passive, connectors.CommandSurfaceCommand{Intent: "direct_read"})
	if passive.String() != "" {
		t.Fatalf("download flags rendered for a non-download intent: %q", passive.String())
	}
}
