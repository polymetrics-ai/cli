package certify_test

import (
	"context"
	"io"
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
	certify.SetCLIRunContextFunc(func(ctx context.Context, args []string, stdout, stderr io.Writer, opts certify.CLIInvocationOptions) int {
		return cli.RunWithContext(ctx, args, stdout, stderr, cli.RunOptions{Mode: cli.ModePlain, ScheduleCrontabFile: opts.CrontabFile})
	})
	os.Exit(m.Run())
}
