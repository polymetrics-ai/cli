package connectors_test

import (
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestParseWriteConfirmationIsClosed(t *testing.T) {
	confirmation, err := connectors.ParseWriteConfirmation("destructive")
	if err != nil {
		t.Fatalf("ParseWriteConfirmation(destructive): %v", err)
	}
	if confirmation.Kind != connectors.ConfirmationKindDestructive {
		t.Fatalf("confirmation = %+v, want destructive", confirmation)
	}

	for _, input := range []string{"yes", "delete widgets", " destructive now "} {
		t.Run(input, func(t *testing.T) {
			if _, err := connectors.ParseWriteConfirmation(input); err == nil {
				t.Fatalf("ParseWriteConfirmation(%q) accepted free text", input)
			}
		})
	}
}
