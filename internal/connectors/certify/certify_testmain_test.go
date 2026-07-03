package certify_test

import (
	"os"
	"testing"

	"polymetrics.ai/internal/cli"
	"polymetrics.ai/internal/connectors/certify"
)

// TestMain wires the real internal/cli.Run entrypoint into this package's
// in-process CLI driver exactly once for the whole test binary
// (certify.SetCLIRunFunc), mirroring what cmd/pm/main.go does in
// production. certify cannot import internal/cli directly (internal/cli's
// own `pm connectors certify` dispatch imports certify, and Go forbids the
// resulting cycle), so every stage/harness/runner test in this package
// depends on this registration having already happened.
func TestMain(m *testing.M) {
	certify.SetCLIRunFunc(cli.Run)
	os.Exit(m.Run())
}
