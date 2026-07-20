package main

import (
	"context"
	"io"
	"os"

	"polymetrics.ai/internal/cli"
	"polymetrics.ai/internal/connectors/certify"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	// certify cannot import internal/cli directly (internal/cli's own `pm
	// connectors certify` dispatch imports certify, and Go forbids the
	// resulting cycle), so this is the one place — a leaf package that can
	// safely see both — that wires the real in-process CLI entrypoint into
	// certify's harness before any command runs (see
	// internal/connectors/certify/cliharness.go SetCLIRunFunc).
	certify.SetCLIRunFunc(cli.Run)
	certify.SetCLIRunContextFunc(func(ctx context.Context, args []string, stdout, stderr io.Writer, opts certify.CLIInvocationOptions) int {
		return cli.RunWithContext(ctx, args, stdout, stderr, cli.RunOptions{
			Mode:                cli.ModePlain,
			ScheduleCrontabFile: opts.CrontabFile,
		})
	})
	return cli.Run(args, stdout, stderr)
}
