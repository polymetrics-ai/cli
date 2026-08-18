package app

import (
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func TestDeclaredTransportCertificationFailsWhenDeclarationIsMissing(t *testing.T) {
	bundle, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatal(err)
	}
	bundle.SyncTransport = nil

	applicable, err := declaredTransportCertificationApplicability(engine.New(bundle, nil))
	if err == nil || applicable {
		t.Fatalf("missing transport declaration returned applicable=%t error=%v, want terminal failure", applicable, err)
	}
	if !strings.Contains(err.Error(), "absent or incomplete") {
		t.Fatalf("missing transport declaration error = %q", err)
	}
}
