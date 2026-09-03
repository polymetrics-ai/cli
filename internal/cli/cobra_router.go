package cli

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/connectors"
)

type cobraLegacyHandler func(context.Context, string, []string, io.Writer, bool) error
type cobraLegacyManualResolver func([]string) (string, bool)

type cobraLegacyCommand struct {
	name           string
	hidden         bool
	manualResolver cobraLegacyManualResolver
	handler        cobraLegacyHandler
}

type cobraLegacyError struct {
	err error
}

func (e *cobraLegacyError) Error() string { return e.err.Error() }

func (e *cobraLegacyError) Unwrap() error { return e.err }

func markCobraLegacyError(err error) error {
	if err == nil {
		return nil
	}
	var legacy *cobraLegacyError
	if errors.As(err, &legacy) {
		return err
	}
	return &cobraLegacyError{err: err}
}

func newRootCmd(ctx context.Context, cfg config.Config, stdout, stderr io.Writer, candidates ...appOpeners) *cobra.Command {
	openers := selectAppOpeners(candidates...)
	if openers.mode == appOpenerProduction && openers.registry == nil {
		openers.registry = appRegistry()
	}
	registry := openers.registry
	if registry == nil {
		registry = appRegistry()
	}
	root := cfg.Root
	jsonOut := cfg.JSON
	cmd := &cobra.Command{
		Use:                "pm",
		Short:              "local-first Polymetrics AI ETL and reverse ETL CLI",
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: true,
		SilenceErrors:      true,
		SilenceUsage:       true,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 || isRootManualArg(args[0]) {
				return markCobraLegacyError(writeRootManualWithRegistry(stdout, jsonOut, registry))
			}
			if len(args) > 1 && isHelpArg(args[1]) {
				if isDynamicConnectorCommandWithRegistry(args[0], registry) {
					return markCobraLegacyError(runMaybeConnectorCommandWithRegistry(ctx, root, args[0], args[1:], stdout, stderr, jsonOut, registry, openers))
				}
				return markCobraLegacyError(usageErrorf("unknown command %q", args[0]))
			}
			return markCobraLegacyError(runMaybeConnectorCommandWithRegistry(ctx, root, args[0], args[1:], stdout, stderr, jsonOut, registry, openers))
		},
	}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.PersistentFlags().String("root", root, "project root (parsed by the legacy global parser)")
	cmd.PersistentFlags().Bool("json", jsonOut, "write machine-readable JSON output (parsed by the legacy global parser)")
	setManualHelp(cmd, "", stdout, jsonOut, registry)
	for _, spec := range cobraLegacyCommandsWithRegistry(cfg, openers, registry, stderr) {
		cmd.AddCommand(newLegacyCobraCommand(ctx, root, stdout, jsonOut, spec, registry))
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

func cobraLegacyCommands(cfg config.Config, openers appOpeners, stderrWriters ...io.Writer) []cobraLegacyCommand {
	registry := openers.registry
	if registry == nil {
		registry = appRegistry()
	}
	return cobraLegacyCommandsWithRegistry(cfg, openers, registry, stderrWriters...)
}

func cobraLegacyCommandsWithRegistry(cfg config.Config, openers appOpeners, registry *connectors.Registry, stderrWriters ...io.Writer) []cobraLegacyCommand {
	stderr := io.Discard
	if len(stderrWriters) > 0 && stderrWriters[0] != nil {
		stderr = stderrWriters[0]
	}
	return []cobraLegacyCommand{
		{name: "init", handler: func(_ context.Context, root string, _ []string, stdout io.Writer, jsonOut bool) error {
			return runInit(root, stdout, jsonOut)
		}},
		{name: "help", handler: func(_ context.Context, _ string, args []string, stdout io.Writer, jsonOut bool) error {
			return runManualAliasWithRegistry(args, stdout, jsonOut, registry)
		}},
		{name: "man", handler: func(_ context.Context, _ string, args []string, stdout io.Writer, jsonOut bool) error {
			return runManualAliasWithRegistry(args, stdout, jsonOut, registry)
		}},
		{name: "connectors", handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return runConnectorsWithRegistry(ctx, root, args, stdout, stderr, jsonOut, registry)
		}},
		{name: "credentials", handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return withApp(openers, root, func(a *app.App) error { return runCredentials(ctx, a, args, stdout, jsonOut) })
		}},
		{name: "connections", handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return withApp(openers, root, func(a *app.App) error { return runConnections(ctx, a, args, stdout, jsonOut) })
		}},
		{name: "catalog", handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return withApp(openers, root, func(a *app.App) error { return runCatalog(ctx, a, args, stdout, jsonOut) })
		}},
		{name: "etl", manualResolver: etlTransportManualCommand, handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return withApp(openers, root, func(a *app.App) error { return runETL(ctx, a, args, stdout, jsonOut, cfg) })
		}},
		{name: "query", handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return withApp(openers, root, func(a *app.App) error { return runQuery(ctx, a, args, stdout, jsonOut) })
		}},
		{name: "reverse", handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			approval, err := prepareReverseRunApproval(args, os.Stdin)
			if err != nil {
				return err
			}
			if approval.supplied {
				return withReverseExecutionApp(openers, root, func(a *app.App) error { return runReverse(ctx, a, args, approval, stdout, jsonOut) })
			}
			return withApp(openers, root, func(a *app.App) error { return runReverse(ctx, a, args, approval, stdout, jsonOut) })
		}},
		{name: "agent", handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return runAgent(ctx, cfg, root, args, stdout, jsonOut)
		}},
		{name: "runtime", handler: func(ctx context.Context, _ string, args []string, stdout io.Writer, jsonOut bool) error {
			return runRuntime(ctx, cfg, args, stdout, jsonOut)
		}},
		{name: "flow", handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return withApp(openers, root, func(a *app.App) error { return runFlow(ctx, cfg, a, args, stdout, jsonOut) })
		}},
		{name: "extract", hidden: true, handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return withApp(openers, root, func(a *app.App) error { return runExtract(ctx, a, cfg, root, args, stdout, jsonOut) })
		}},
		{name: "perf", handler: func(ctx context.Context, _ string, args []string, stdout io.Writer, jsonOut bool) error {
			return runPerf(ctx, cfg, args, stdout, jsonOut)
		}},
		{name: "docs", handler: func(_ context.Context, _ string, args []string, stdout io.Writer, _ bool) error {
			return runDocsWithRegistry(args, stdout, registry)
		}},
		{name: "skills", handler: func(_ context.Context, _ string, args []string, stdout io.Writer, jsonOut bool) error {
			return runSkillsWithRegistry(args, stdout, jsonOut, registry)
		}},
		{name: "version", handler: func(_ context.Context, _ string, args []string, stdout io.Writer, jsonOut bool) error {
			return runVersion(args, stdout, jsonOut)
		}},
		{name: "rlm", handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return runRLM(ctx, cfg, root, args, stdout, jsonOut)
		}},
		{name: "schedule", handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return runSchedule(ctx, cfg, root, args, stdout, jsonOut, openers)
		}},
		{name: "worker", hidden: true, handler: func(ctx context.Context, _ string, args []string, stdout io.Writer, jsonOut bool) error {
			return runWorker(ctx, cfg, args, stdout, jsonOut)
		}},
	}
}

func newLegacyCobraCommand(ctx context.Context, root string, stdout io.Writer, jsonOut bool, spec cobraLegacyCommand, registries ...*connectors.Registry) *cobra.Command {
	registry := appRegistry()
	if len(registries) == 1 && registries[0] != nil {
		registry = registries[0]
	}
	cmd := &cobra.Command{
		Use:                spec.name,
		Hidden:             spec.hidden,
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: true,
		SilenceErrors:      true,
		SilenceUsage:       true,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 && isHelpArg(args[0]) {
				return markCobraLegacyError(writeManualWithRegistry(spec.name, stdout, jsonOut, registry))
			}
			if spec.manualResolver != nil {
				if command, ok := spec.manualResolver(args); ok {
					return markCobraLegacyError(writeETLTransportManual(stdout, jsonOut, command))
				}
			}
			if containsHelpFlag(args) {
				if manualDocumentsInvocation(spec.name, args) {
					return markCobraLegacyError(writeManualWithRegistry(spec.name, stdout, jsonOut, registry))
				}
				return markCobraLegacyError(usageErrorf("unknown command %q", strings.Join(commandPath(args), " ")))
			}
			if len(args) == 0 && isManualCommand(spec.name) {
				return markCobraLegacyError(writeManualWithRegistry(spec.name, stdout, jsonOut, registry))
			}
			return markCobraLegacyError(spec.handler(ctx, root, args, stdout, jsonOut))
		},
	}
	setManualHelp(cmd, spec.name, stdout, jsonOut, registry)
	return cmd
}

func runManualAlias(_ context.Context, _ string, args []string, stdout io.Writer, jsonOut bool) error {
	return runManualAliasWithRegistry(args, stdout, jsonOut, appRegistry())
}

func runManualAliasWithRegistry(args []string, stdout io.Writer, jsonOut bool, registry *connectors.Registry) error {
	if len(args) == 0 {
		return writeRootManualWithRegistry(stdout, jsonOut, registry)
	}
	return runHelpWithRegistry(args, stdout, jsonOut, registry)
}

func setManualHelp(cmd *cobra.Command, topic string, stdout io.Writer, jsonOut bool, registries ...*connectors.Registry) {
	registry := appRegistry()
	if len(registries) == 1 && registries[0] != nil {
		registry = registries[0]
	}
	cmd.SetHelpFunc(func(_ *cobra.Command, _ []string) {
		_ = writeManualTopicWithRegistry(topic, stdout, jsonOut, registry)
	})
	cmd.SetUsageFunc(func(_ *cobra.Command) error {
		return writeManualTopicWithRegistry(topic, stdout, jsonOut, registry)
	})
}

func writeManualTopic(topic string, stdout io.Writer, jsonOut bool) error {
	return writeManualTopicWithRegistry(topic, stdout, jsonOut, appRegistry())
}

func writeManualTopicWithRegistry(topic string, stdout io.Writer, jsonOut bool, registry *connectors.Registry) error {
	if topic == "" {
		return writeRootManualWithRegistry(stdout, jsonOut, registry)
	}
	return writeManualWithRegistry(topic, stdout, jsonOut, registry)
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

func containsHelpFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

// manualDocumentsInvocation admits a help request only when its non-flag
// command prefix matches a documented invocation in the wrapper's own manual.
// The manual is the declaration shared by runtime help and generated CLI docs,
// so a newly documented leaf stays project-free without a per-command switch.
func manualDocumentsInvocation(topic string, args []string) bool {
	manual, ok := docs[topic]
	if !ok {
		return false
	}
	path := append([]string{topic}, commandPath(args)...)
	if len(path) == 1 {
		return true
	}
	for _, documented := range manualCommandPaths(manual) {
		if commandPathsOverlap(path, documented) {
			return true
		}
	}
	return false
}

func commandPath(args []string) []string {
	path := make([]string, 0, len(args))
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			break
		}
		path = append(path, arg)
	}
	return path
}

func manualCommandPaths(manual string) [][]string {
	var paths [][]string
	inCommandSection := false
	for _, line := range strings.Split(manual, "\n") {
		trimmed := strings.TrimSpace(line)
		switch trimmed {
		case "SYNOPSIS", "USAGE":
			inCommandSection = true
			continue
		}
		if isManualSectionHeader(trimmed) {
			inCommandSection = false
		}
		if !inCommandSection {
			continue
		}
		if path := manualCommandPath(trimmed); len(path) > 0 {
			paths = append(paths, path)
		}
	}
	return paths
}

func isManualSectionHeader(line string) bool {
	return line != "" && line == strings.ToUpper(line) && strings.IndexFunc(line, func(r rune) bool {
		return r >= 'A' && r <= 'Z'
	}) >= 0
}

func manualCommandPath(line string) []string {
	fields := strings.Fields(strings.TrimSpace(line))
	if len(fields) < 2 || fields[0] != "pm" {
		return nil
	}
	path := make([]string, 0, len(fields)-1)
	for _, field := range fields[1:] {
		if strings.HasPrefix(field, "[") || strings.HasPrefix(field, "<") || strings.HasPrefix(field, "--") {
			break
		}
		path = append(path, field)
	}
	return path
}

func commandPathsOverlap(path, documented []string) bool {
	limit := min(len(path), len(documented))
	for i := 0; i < limit; i++ {
		if path[i] != documented[i] {
			return false
		}
	}
	return true
}

func mapCobraErr(err error) error {
	if err == nil {
		return nil
	}
	var ce *cliError
	if errors.As(err, &ce) || errors.Is(err, errUsage) {
		return err
	}
	var legacy *cobraLegacyError
	if errors.As(err, &legacy) {
		return err
	}
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return err
	}
	if strings.Contains(message, "unknown command") || strings.Contains(message, "unknown flag") || strings.Contains(message, "unknown shorthand flag") {
		return usageErrorf("%s", message)
	}
	return err
}
