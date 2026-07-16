package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"polymetrics.ai/internal/app"
)

type cobraLegacyHandler func(context.Context, string, []string, io.Writer, bool) error

type cobraLegacyCommand struct {
	name    string
	hidden  bool
	handler cobraLegacyHandler
}

func newRootCmd(ctx context.Context, root string, stdout, stderr io.Writer, jsonOut bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:                "pm",
		Short:              "local-first Polymetrics AI ETL and reverse ETL CLI",
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: true,
		SilenceErrors:      true,
		SilenceUsage:       true,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 || isRootManualArg(args[0]) {
				return writeRootManual(stdout, jsonOut)
			}
			return runMaybeConnectorCommand(ctx, root, args[0], args[1:], stdout, jsonOut)
		},
	}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	setManualHelp(cmd, "", stdout, jsonOut)
	for _, spec := range cobraLegacyCommands() {
		cmd.AddCommand(newLegacyCobraCommand(ctx, root, stdout, jsonOut, spec))
	}
	return cmd
}

func executeRootCmd(cmd *cobra.Command, args []string) error {
	if len(args) > 0 && lookupTopLevelCommand(cmd, args[0]) == nil {
		return cmd.RunE(cmd, args)
	}
	cmd.SetArgs(append([]string(nil), args...))
	if len(args) == 0 {
		cmd.SetArgs([]string{})
	}
	_, err := cmd.ExecuteC()
	return err
}

func cobraLegacyCommands() []cobraLegacyCommand {
	return []cobraLegacyCommand{
		{name: "init", handler: func(_ context.Context, root string, _ []string, stdout io.Writer, jsonOut bool) error {
			return runInit(root, stdout, jsonOut)
		}},
		{name: "help", handler: runManualAlias},
		{name: "man", handler: runManualAlias},
		{name: "connectors", handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return runConnectors(ctx, root, args, stdout, jsonOut)
		}},
		{name: "credentials", handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return withApp(root, func(a *app.App) error { return runCredentials(ctx, a, args, stdout, jsonOut) })
		}},
		{name: "connections", handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return withApp(root, func(a *app.App) error { return runConnections(ctx, a, args, stdout, jsonOut) })
		}},
		{name: "catalog", handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return withApp(root, func(a *app.App) error { return runCatalog(ctx, a, args, stdout, jsonOut) })
		}},
		{name: "etl", handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return withApp(root, func(a *app.App) error { return runETL(ctx, a, args, stdout, jsonOut) })
		}},
		{name: "query", handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return withApp(root, func(a *app.App) error { return runQuery(ctx, a, args, stdout, jsonOut) })
		}},
		{name: "reverse", handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return withApp(root, func(a *app.App) error { return runReverse(ctx, a, args, stdout, jsonOut) })
		}},
		{name: "agent", handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return runAgent(ctx, root, args, stdout, jsonOut)
		}},
		{name: "runtime", handler: func(ctx context.Context, _ string, args []string, stdout io.Writer, jsonOut bool) error {
			return runRuntime(ctx, args, stdout, jsonOut)
		}},
		{name: "flow", handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return withApp(root, func(a *app.App) error { return runFlow(ctx, a, args, stdout, jsonOut) })
		}},
		{name: "extract", hidden: true, handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return withApp(root, func(a *app.App) error { return runExtract(ctx, a, root, args, stdout, jsonOut) })
		}},
		{name: "perf", handler: func(ctx context.Context, _ string, args []string, stdout io.Writer, jsonOut bool) error {
			return runPerf(ctx, args, stdout, jsonOut)
		}},
		{name: "docs", handler: func(_ context.Context, _ string, args []string, stdout io.Writer, _ bool) error {
			return runDocs(args, stdout)
		}},
		{name: "skills", handler: func(_ context.Context, _ string, args []string, stdout io.Writer, jsonOut bool) error {
			return runSkills(args, stdout, jsonOut)
		}},
		{name: "version", handler: func(_ context.Context, _ string, args []string, stdout io.Writer, jsonOut bool) error {
			return runVersion(args, stdout, jsonOut)
		}},
		{name: "rlm", handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return runRLM(ctx, root, args, stdout, jsonOut)
		}},
		{name: "schedule", handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return runSchedule(ctx, root, args, stdout, jsonOut)
		}},
		{name: "worker", hidden: true, handler: func(ctx context.Context, _ string, args []string, stdout io.Writer, jsonOut bool) error {
			return runWorker(ctx, args, stdout, jsonOut)
		}},
	}
}

func newLegacyCobraCommand(ctx context.Context, root string, stdout io.Writer, jsonOut bool, spec cobraLegacyCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:                spec.name,
		Hidden:             spec.hidden,
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: true,
		SilenceErrors:      true,
		SilenceUsage:       true,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 && isHelpArg(args[0]) {
				return writeManual(spec.name, stdout, jsonOut)
			}
			if len(args) == 0 && isManualCommand(spec.name) {
				return writeManual(spec.name, stdout, jsonOut)
			}
			return spec.handler(ctx, root, args, stdout, jsonOut)
		},
	}
	setManualHelp(cmd, spec.name, stdout, jsonOut)
	return cmd
}

func runManualAlias(_ context.Context, _ string, args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 {
		return writeRootManual(stdout, jsonOut)
	}
	return runHelp(args, stdout)
}

func setManualHelp(cmd *cobra.Command, topic string, stdout io.Writer, jsonOut bool) {
	cmd.SetHelpFunc(func(_ *cobra.Command, _ []string) {
		_ = writeManualTopic(topic, stdout, jsonOut)
	})
	cmd.SetUsageFunc(func(_ *cobra.Command) error {
		return writeManualTopic(topic, stdout, jsonOut)
	})
}

func writeManualTopic(topic string, stdout io.Writer, jsonOut bool) error {
	if topic == "" {
		return writeRootManual(stdout, jsonOut)
	}
	return writeManual(topic, stdout, jsonOut)
}

func lookupTopLevelCommand(root *cobra.Command, name string) *cobra.Command {
	for _, cmd := range root.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}

func isRootManualArg(arg string) bool {
	return arg == "--help" || arg == "-h"
}

func isHelpArg(arg string) bool {
	return arg == "--help" || arg == "-h" || arg == "help"
}

func mapCobraErr(err error) error {
	if err == nil {
		return nil
	}
	var ce *cliError
	if errors.As(err, &ce) || errors.Is(err, errUsage) {
		return err
	}
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return err
	}
	if strings.Contains(message, "unknown command") || strings.Contains(message, "unknown flag") || strings.Contains(message, "unknown shorthand flag") {
		return usageErrorf("%s", message)
	}
	return fmt.Errorf("%s", message)
}
