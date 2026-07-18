package cli

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/config"
)

type cobraLegacyHandler func(context.Context, string, []string, io.Writer, bool) error

type cobraLegacyCommand struct {
	name    string
	hidden  bool
	handler cobraLegacyHandler
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

func newRootCmd(ctx context.Context, cfg config.Config, stdout, stderr io.Writer) *cobra.Command {
	return newRootCmdWithAgentImageRuntime(ctx, cfg, stdout, stderr, osAgentImageRuntime{})
}

func newRootCmdWithAgentImageRuntime(ctx context.Context, cfg config.Config, stdout, stderr io.Writer, imageRuntime agentImageRuntime) *cobra.Command {
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
				return markCobraLegacyError(writeRootManual(stdout, jsonOut))
			}
			if len(args) > 1 && isHelpArg(args[1]) {
				return markCobraLegacyError(writeManual(args[0], stdout, jsonOut))
			}
			return markCobraLegacyError(runMaybeConnectorCommand(ctx, root, args[0], args[1:], stdout, jsonOut))
		},
	}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.PersistentFlags().String("root", root, "project root (parsed by the legacy global parser)")
	cmd.PersistentFlags().Bool("json", jsonOut, "write machine-readable JSON output (parsed by the legacy global parser)")
	cmd.PersistentFlags().Bool("plain", false, "force plain non-TTY output (parsed by the legacy global parser)")
	cmd.PersistentFlags().Bool("no-input", false, "disable interactive prompting and TTY UI (parsed by the legacy global parser)")
	cmd.PersistentFlags().String("progress", "", "progress output format: ndjson writes sanitized events to stderr (parsed by the legacy global parser)")
	setManualHelp(cmd, "", stdout, jsonOut)
	for _, spec := range cobraLegacyCommands(cfg) {
		cmd.AddCommand(newLegacyCobraCommand(ctx, root, stdout, jsonOut, spec))
	}
	cmd.AddCommand(newCredentialsCobraCommand(ctx, root, stdout, jsonOut))
	cmd.AddCommand(newConnectionsCobraCommand(ctx, root, stdout, jsonOut))
	cmd.AddCommand(newCatalogCobraCommand(ctx, root, stdout, jsonOut))
	cmd.AddCommand(newETLCobraCommand(ctx, cfg, root, stdout, jsonOut))
	cmd.AddCommand(newQueryCobraCommand(ctx, root, stdout, jsonOut))
	cmd.AddCommand(newRuntimeCobraCommand(ctx, cfg, stdout, jsonOut))
	cmd.AddCommand(newPerfCobraCommand(ctx, cfg, stdout, jsonOut))
	cmd.AddCommand(newAgentCobraCommand(ctx, cfg, root, stdout, jsonOut, imageRuntime))
	cmd.AddCommand(newDocsCobraCommand(stdout, jsonOut))
	cmd.AddCommand(newSkillsCobraCommand(stdout, jsonOut))
	cmd.AddCommand(newVersionCobraCommand(stdout, jsonOut))
	return cmd
}

func executeRootCmd(cmd *cobra.Command, args []string) error {
	var credentialsState credentialsCommandState
	args = normalizeNativeStringArrayArgs(args, &credentialsState)
	if credentialsState.rawCarrier {
		return errUsage
	}
	if credentialsState.boundedName != "" {
		if err := validateCredentialIdentifier(credentialsState.boundedName, "credential"); err != nil {
			return err
		}
	}
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	cmd.SetContext(context.WithValue(ctx, credentialsCommandStateKey{}, credentialsState))
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

func normalizeNativeStringArrayArgs(args []string, credentialsState *credentialsCommandState) []string {
	if len(args) >= 3 && args[0] == "catalog" && (args[1] == "refresh" || args[1] == "show") {
		return normalizeStringArraySpaceValues(args, 2, map[string]struct{}{"connection": {}})
	}
	if len(args) >= 2 && args[0] == "connections" && args[1] == "create" {
		return normalizeStringArraySpaceValues(args, 2, connectionsCreateFlagNames)
	}
	if len(args) >= 2 && args[0] == "query" && args[1] == "run" {
		return normalizeStringArraySpaceValues(args, 2, queryRunFlagNames)
	}
	if len(args) >= 2 && args[0] == "perf" && args[1] == "compare" {
		return normalizeStringArraySpaceValues(args, 2, perfCompareFlagNames)
	}
	if len(args) >= 2 && args[0] == "perf" && args[1] == "sync-modes" {
		return normalizeStringArraySpaceValues(args, 2, perfSyncModesFlagNames)
	}
	if len(args) >= 2 && args[0] == "skills" && args[1] == "generate" {
		return normalizeStringArraySpaceValues(args, 2, skillsGenerateFlagNames)
	}
	if len(args) >= 2 && args[0] == "etl" {
		if isHelpArg(args[1]) {
			return append([]string(nil), args[:2]...)
		}
		switch args[1] {
		case "check", "catalog", "read", "run", "status":
			args = normalizeStringArraySpaceValues(args, 2, etlFlagNames)
			return normalizeETLLegacyActionArgs(args, 2)
		}
	}
	if len(args) >= 2 && args[0] == "credentials" {
		if credentialsActionTakesName(args[1]) && credentialsArgsContainRawCarrier(args[2:]) {
			credentialsState.rawCarrier = true
			return args
		}
		if isHelpArg(args[1]) {
			return append([]string(nil), args[:2]...)
		}
		args, bounded := normalizeCredentialsActionBoundary(args, credentialsState)
		if args[1] == "add" {
			args = normalizeStringArraySpaceValues(args, 2, credentialsAddFlagNames)
		}
		if bounded {
			if credentialsActionTakesName(args[1]) {
				return normalizeCredentialsLegacyActionArgs(args, 2)
			}
			return args
		}
		if args[1] != "help" && !isHelpArg(args[1]) {
			return normalizeCredentialsLegacyActionArgs(args, 2)
		}
	}
	if len(args) >= 2 && args[0] == "agent" {
		var bounded bool
		args, bounded = normalizeAgentActionBoundary(args)
		if bounded {
			return args
		}
		switch args[1] {
		case "plan":
			args = normalizeStringArraySpaceValues(args, 2, agentPlanFlagNames)
			return normalizeAgentLegacyActionArgs(args, 2)
		case "image":
			if len(args) >= 3 && isLegacyHelpFlag(args[2]) {
				out := append([]string(nil), args[:2]...)
				out = append(out, "--")
				out = append(out, args[2:]...)
				return out
			}
			if len(args) >= 3 {
				return normalizeAgentLegacyActionArgs(args, 3)
			}
		default:
			if !isHelpArg(args[1]) {
				return normalizeAgentLegacyActionArgs(args, 2)
			}
		}
	}
	if len(args) >= 2 && args[0] == "docs" {
		if args[1] == "generate" || args[1] == "validate" {
			args = normalizeStringArraySpaceValues(args, 2, docsFlagNames)
		}
		if args[1] != "help" && !isHelpArg(args[1]) {
			return normalizeDocsLegacyActionArgs(args, 2)
		}
	}
	return args
}

// normalizeETLLegacyActionArgs keeps tokens ignored by the old ETL parser from becoming Cobra controls.
func normalizeETLLegacyActionArgs(args []string, start int) []string {
	out := make([]string, 0, len(args))
	out = append(out, args[:start]...)
	for _, arg := range args[start:] {
		if arg == "--" || isLegacyHelpFlag(arg) || (strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--")) {
			continue
		}
		out = append(out, arg)
	}
	return out
}

func normalizeCredentialsActionBoundary(args []string, state *credentialsCommandState) ([]string, bool) {
	switch args[1] {
	case "add", "inspect", "test", "remove":
		if len(args) >= 3 && strings.HasPrefix(args[2], "-") {
			state.boundedName = args[2]
			out := make([]string, 0, len(args)-1)
			out = append(out, args[:2]...)
			out = append(out, args[3:]...)
			return out, true
		}
		return args, false
	case "list", "help", "-h", "--help":
		return args, false
	}

	out := make([]string, 0, len(args)+1)
	out = append(out, args[0], "--")
	out = append(out, args[1:]...)
	return out, true
}

func credentialsActionTakesName(action string) bool {
	switch action {
	case "add", "inspect", "test", "remove":
		return true
	default:
		return false
	}
}

func credentialsArgsContainRawCarrier(args []string) bool {
	const rawCarrier = "--pm-internal-credentials-name"
	for _, arg := range args {
		if arg == rawCarrier || strings.HasPrefix(arg, rawCarrier+"=") {
			return true
		}
	}
	return false
}

// normalizeCredentialsLegacyActionArgs keeps tokens ignored by the old credentials parser from becoming Cobra controls.
func normalizeCredentialsLegacyActionArgs(args []string, start int) []string {
	out := make([]string, 0, len(args))
	out = append(out, args[:start]...)
	for _, arg := range args[start:] {
		if arg == "--" || arg == "-h" || arg == "--help" || strings.HasPrefix(arg, "--help=") ||
			(strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--")) {
			continue
		}
		out = append(out, arg)
	}
	return out
}

func normalizeAgentActionBoundary(args []string) ([]string, bool) {
	boundary := 1
	switch args[1] {
	case "plan", "help", "-h", "--help":
		return args, false
	case "image":
		if len(args) < 3 {
			return args, false
		}
		boundary = 2
		switch args[2] {
		case "build", "pull", "ensure":
			return args, false
		}
	}

	out := make([]string, 0, len(args)+1)
	out = append(out, args[:boundary]...)
	out = append(out, "--")
	out = append(out, args[boundary:]...)
	return out, true
}

// normalizeAgentLegacyActionArgs keeps tokens ignored by the old agent parser from becoming Cobra control flags.
func normalizeAgentLegacyActionArgs(args []string, start int) []string {
	out := make([]string, 0, len(args))
	out = append(out, args[:start]...)
	for _, arg := range args[start:] {
		if arg == "--" || isLegacyHelpFlag(arg) {
			continue
		}
		out = append(out, arg)
	}
	return out
}

func isLegacyHelpFlag(arg string) bool {
	return arg == "-h" || arg == "--help" || strings.HasPrefix(arg, "--help=") ||
		(len(arg) > 2 && arg[0] == '-' && arg[1] != '-' && strings.ContainsRune(arg[1:], 'h'))
}

func normalizeDocsLegacyActionArgs(args []string, start int) []string {
	out := make([]string, 0, len(args))
	out = append(out, args[:start]...)
	for _, arg := range args[start:] {
		if arg == "--" || arg == "-h" || arg == "--help" || strings.HasPrefix(arg, "--help=") {
			continue
		}
		out = append(out, arg)
	}
	return out
}

var etlFlagNames = map[string]struct{}{
	"connector":  {},
	"config":     {},
	"stream":     {},
	"limit":      {},
	"connection": {},
	"batch-size": {},
	"runtime":    {},
	"root":       {},
	"progress":   {},
}

var connectionsCreateFlagNames = map[string]struct{}{
	"source":             {},
	"destination":        {},
	"stream":             {},
	"sync-mode":          {},
	"cursor":             {},
	"primary-key":        {},
	"table":              {},
	"source-config":      {},
	"destination-config": {},
}

var queryRunFlagNames = map[string]struct{}{
	"table":      {},
	"sql":        {},
	"limit":      {},
	"fields":     {},
	"agent-mode": {},
	"sample":     {},
}

var perfCompareFlagNames = map[string]struct{}{
	"iterations": {},
	"runtime":    {},
}

var perfSyncModesFlagNames = map[string]struct{}{
	"records": {},
}

var skillsGenerateFlagNames = map[string]struct{}{
	"dir": {},
}

var credentialsAddFlagNames = map[string]struct{}{
	"connector":   {},
	"from-env":    {},
	"value-stdin": {},
	"config":      {},
	"root":        {},
	"progress":    {},
}

var agentPlanFlagNames = map[string]struct{}{
	"request": {},
}

var docsFlagNames = map[string]struct{}{
	"dir":            {},
	"connectors-dir": {},
}

func normalizeStringArraySpaceValues(args []string, start int, flagNames map[string]struct{}) []string {
	out := make([]string, 0, len(args))
	out = append(out, args[:start]...)
	for i := start; i < len(args); i++ {
		arg := args[i]
		if flagName, ok := nativeStringArrayFlagName(arg); ok {
			if _, known := flagNames[flagName]; known && i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				out = append(out, arg+"="+args[i+1])
				i++
				continue
			}
		}
		out = append(out, arg)
	}
	return out
}

func nativeStringArrayFlagName(arg string) (string, bool) {
	if !strings.HasPrefix(arg, "--") || strings.Contains(arg, "=") {
		return "", false
	}
	name := strings.TrimPrefix(arg, "--")
	if name == "" {
		return "", false
	}
	return name, true
}

func cobraLegacyCommands(cfg config.Config) []cobraLegacyCommand {
	return []cobraLegacyCommand{
		{name: "init", handler: func(_ context.Context, root string, _ []string, stdout io.Writer, jsonOut bool) error {
			return runInit(root, stdout, jsonOut)
		}},
		{name: "help", handler: runManualAlias},
		{name: "man", handler: runManualAlias},
		{name: "connectors", handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return runConnectors(ctx, root, args, stdout, jsonOut)
		}},
		{name: "reverse", handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return withApp(root, func(a *app.App) error { return runReverse(ctx, a, args, stdout, jsonOut) })
		}},
		{name: "flow", handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return withApp(root, func(a *app.App) error { return runFlow(ctx, cfg, a, args, stdout, jsonOut) })
		}},
		{name: "extract", hidden: true, handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return withApp(root, func(a *app.App) error { return runExtract(ctx, a, cfg, root, args, stdout, jsonOut) })
		}},
		{name: "rlm", handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return runRLM(ctx, cfg, root, args, stdout, jsonOut)
		}},
		{name: "schedule", handler: func(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
			return runSchedule(ctx, cfg, root, args, stdout, jsonOut)
		}},
		{name: "worker", hidden: true, handler: func(ctx context.Context, _ string, args []string, stdout io.Writer, jsonOut bool) error {
			return runWorker(ctx, cfg, args, stdout, jsonOut)
		}},
	}
}

func newConnectionsCobraCommand(ctx context.Context, root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "connections",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completeConnectionNames(cmd, args, toComplete)
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return markCobraLegacyError(writeManual("connections", stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "connections", stdout, jsonOut)
	cmd.AddCommand(newConnectionsCreateCobraCommand(ctx, root, stdout, jsonOut))
	cmd.AddCommand(newConnectionsListCobraCommand(ctx, root, stdout, jsonOut))
	return cmd
}

func newConnectionsCreateCobraCommand(ctx context.Context, root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	var flags connectionsCreateFlags
	cmd := &cobra.Command{
		Use:           "create <name>",
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errUsage
			}
			return markCobraLegacyError(withApp(root, func(a *app.App) error {
				return runConnectionsCreate(ctx, a, args[0], flags, stdout, jsonOut)
			}))
		},
	}
	setManualHelp(cmd, "connections", stdout, jsonOut)
	addConnectionsStringArrayFlag(cmd, &flags.Sources, "source", "source connector:credential")
	addConnectionsStringArrayFlag(cmd, &flags.Destinations, "destination", "destination connector:credential")
	addConnectionsStringArrayFlag(cmd, &flags.Streams, "stream", "stream name")
	addConnectionsStringArrayFlag(cmd, &flags.SyncModes, "sync-mode", "sync mode")
	addConnectionsStringArrayFlag(cmd, &flags.Cursors, "cursor", "cursor field")
	addConnectionsStringArrayFlag(cmd, &flags.PrimaryKeys, "primary-key", "primary key field")
	addConnectionsStringArrayFlag(cmd, &flags.Tables, "table", "destination table")
	addConnectionsStringArrayFlag(cmd, &flags.SourceConfigs, "source-config", "source config key=value")
	addConnectionsStringArrayFlag(cmd, &flags.DestinationConfigs, "destination-config", "destination config key=value")
	return cmd
}

func newConnectionsListCobraCommand(ctx context.Context, root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "list",
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return markCobraLegacyError(withApp(root, func(a *app.App) error {
				return runConnectionsList(a, stdout, jsonOut)
			}))
		},
	}
	setManualHelp(cmd, "connections", stdout, jsonOut)
	return cmd
}

func addConnectionsStringArrayFlag(cmd *cobra.Command, target *[]string, name string, usage string) {
	cmd.Flags().StringArrayVar(target, name, nil, usage)
	if flag := cmd.Flags().Lookup(name); flag != nil {
		flag.NoOptDefVal = "true"
	}
}

func completeConnectionNames(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}

type queryRunFlags struct {
	Tables     []string
	SQLs       []string
	Limits     []string
	Fields     []string
	AgentModes []string
	Samples    []string
}

func (f queryRunFlags) parsed() parsedFlags {
	values := map[string][]string{}
	add := func(name string, vals []string) {
		if len(vals) > 0 {
			values[name] = append([]string(nil), vals...)
		}
	}
	add("table", f.Tables)
	add("sql", f.SQLs)
	add("limit", f.Limits)
	add("fields", f.Fields)
	add("agent-mode", f.AgentModes)
	add("sample", f.Samples)
	return parsedFlags{values: values}
}

func newQueryCobraCommand(ctx context.Context, root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "query",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return markCobraLegacyError(writeManual("query", stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "query", stdout, jsonOut)
	cmd.AddCommand(newQueryRunCobraCommand(ctx, root, stdout, jsonOut))
	return cmd
}

func newQueryRunCobraCommand(ctx context.Context, root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	var flags queryRunFlags
	cmd := &cobra.Command{
		Use:           "run",
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		ValidArgsFunction: completeNoFile,
		RunE: func(_ *cobra.Command, _ []string) error {
			return markCobraLegacyError(withApp(root, func(a *app.App) error {
				return runQueryRun(ctx, a, flags.parsed(), stdout, jsonOut)
			}))
		},
	}
	setManualHelp(cmd, "query", stdout, jsonOut)
	addQueryStringArrayFlag(cmd, &flags.Tables, "table", "local warehouse table to scan")
	addQueryStringArrayFlag(cmd, &flags.SQLs, "sql", "read-only SQL query")
	addQueryStringArrayFlag(cmd, &flags.Limits, "limit", "maximum rows to read")
	addQueryStringArrayFlag(cmd, &flags.Fields, "fields", "comma-separated fields to project")
	addQueryStringArrayFlag(cmd, &flags.AgentModes, "agent-mode", "agent output mode")
	addQueryStringArrayFlag(cmd, &flags.Samples, "sample", "summary sample size")
	return cmd
}

func addQueryStringArrayFlag(cmd *cobra.Command, target *[]string, name string, usage string) {
	cmd.Flags().StringArrayVar(target, name, nil, usage)
	if flag := cmd.Flags().Lookup(name); flag != nil {
		flag.NoOptDefVal = "true"
	}
}

func completeNoFile(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func newRuntimeCobraCommand(ctx context.Context, cfg config.Config, stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "runtime",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return markCobraLegacyError(writeManual("runtime", stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "runtime", stdout, jsonOut)
	cmd.AddCommand(newRuntimeDoctorCobraCommand(ctx, cfg, stdout, jsonOut))
	cmd.AddCommand(newRuntimeHelpCobraCommand(stdout, jsonOut))
	return cmd
}

func newRuntimeHelpCobraCommand(stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "help",
		Hidden:        true,
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		ValidArgsFunction: completeNoFile,
		RunE: func(_ *cobra.Command, _ []string) error {
			return markCobraLegacyError(writeManual("runtime", stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "runtime", stdout, jsonOut)
	return cmd
}

func newRuntimeDoctorCobraCommand(ctx context.Context, cfg config.Config, stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "doctor",
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		ValidArgsFunction: completeNoFile,
		RunE: func(_ *cobra.Command, _ []string) error {
			return markCobraLegacyError(runRuntimeDoctor(ctx, cfg, stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "runtime", stdout, jsonOut)
	return cmd
}

type agentPlanFlags struct {
	Requests []string
}

func newAgentCobraCommand(ctx context.Context, cfg config.Config, root string, stdout io.Writer, jsonOut bool, imageRuntime agentImageRuntime) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "agent",
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				return errUsage
			}
			return markCobraLegacyError(writeManual("agent", stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "agent", stdout, jsonOut)
	cmd.AddCommand(newAgentPlanCobraCommand(stdout, jsonOut))
	cmd.AddCommand(newAgentImageCobraCommand(ctx, cfg, root, stdout, jsonOut, imageRuntime))
	cmd.AddCommand(newAgentHelpCobraCommand(stdout, jsonOut))
	return cmd
}

func newAgentPlanCobraCommand(stdout io.Writer, jsonOut bool) *cobra.Command {
	var flags agentPlanFlags
	cmd := &cobra.Command{
		Use:           "plan",
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		ValidArgsFunction: completeNoFile,
		RunE: func(_ *cobra.Command, _ []string) error {
			return markCobraLegacyError(runAgentPlan(lastString(flags.Requests), stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "agent", stdout, jsonOut)
	cmd.Flags().StringArrayVar(&flags.Requests, "request", nil, "natural-language planning request")
	if flag := cmd.Flags().Lookup("request"); flag != nil {
		flag.NoOptDefVal = "true"
	}
	return cmd
}

func newAgentImageCobraCommand(ctx context.Context, cfg config.Config, root string, stdout io.Writer, jsonOut bool, imageRuntime agentImageRuntime) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "image",
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		ValidArgsFunction: completeNoFile,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				return usageErrorf("agent image: unknown subcommand %q (want build|pull|ensure)", args[0])
			}
			return markCobraLegacyError(runAgentImage(ctx, cfg, root, args, stdout, jsonOut, imageRuntime))
		},
	}
	setManualHelp(cmd, "agent", stdout, jsonOut)
	for _, action := range []string{"build", "pull", "ensure"} {
		cmd.AddCommand(newAgentImageActionCobraCommand(ctx, cfg, root, stdout, jsonOut, action, imageRuntime))
	}
	return cmd
}

func newAgentImageActionCobraCommand(ctx context.Context, cfg config.Config, root string, stdout io.Writer, jsonOut bool, action string, imageRuntime agentImageRuntime) *cobra.Command {
	cmd := &cobra.Command{
		Use:           action,
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		ValidArgsFunction: completeNoFile,
		RunE: func(_ *cobra.Command, _ []string) error {
			return markCobraLegacyError(runAgentImageAction(ctx, cfg, root, action, stdout, jsonOut, imageRuntime))
		},
	}
	setManualHelp(cmd, "agent", stdout, jsonOut)
	return cmd
}

func newAgentHelpCobraCommand(stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "help",
		Hidden:        true,
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		ValidArgsFunction: completeNoFile,
		RunE: func(_ *cobra.Command, _ []string) error {
			return markCobraLegacyError(writeManual("agent", stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "agent", stdout, jsonOut)
	return cmd
}

func newDocsCobraCommand(stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "docs",
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				return errUsage
			}
			return markCobraLegacyError(writeManual("docs", stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "docs", stdout, jsonOut)
	cmd.AddCommand(newDocsActionCobraCommand(stdout, jsonOut, "generate"))
	cmd.AddCommand(newDocsActionCobraCommand(stdout, jsonOut, "validate"))
	cmd.AddCommand(newDocsHelpCobraCommand(stdout, jsonOut))
	return cmd
}

func newDocsActionCobraCommand(stdout io.Writer, jsonOut bool, action string) *cobra.Command {
	var flags docsCommandFlags
	cmd := &cobra.Command{
		Use:           action,
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		ValidArgsFunction: completeNoFile,
		RunE: func(_ *cobra.Command, _ []string) error {
			return markCobraLegacyError(runDocs(action, flags, stdout))
		},
	}
	setManualHelp(cmd, "docs", stdout, jsonOut)
	addDocsStringArrayFlag(cmd, &flags.Dirs, "dir", "command or connector docs output directory")
	addDocsStringArrayFlag(cmd, &flags.ConnectorsDirs, "connectors-dir", "connector docs output directory")
	return cmd
}

func addDocsStringArrayFlag(cmd *cobra.Command, target *[]string, name, usage string) {
	cmd.Flags().StringArrayVar(target, name, nil, usage)
	if flag := cmd.Flags().Lookup(name); flag != nil {
		flag.NoOptDefVal = "true"
	}
}

func newDocsHelpCobraCommand(stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "help",
		Hidden:        true,
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		ValidArgsFunction: completeNoFile,
		RunE: func(_ *cobra.Command, _ []string) error {
			return markCobraLegacyError(writeManual("docs", stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "docs", stdout, jsonOut)
	return cmd
}

func newSkillsCobraCommand(stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "skills",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return markCobraLegacyError(writeManual("skills", stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "skills", stdout, jsonOut)
	cmd.AddCommand(newSkillsGenerateCobraCommand(stdout, jsonOut))
	cmd.AddCommand(newSkillsHelpCobraCommand(stdout, jsonOut))
	return cmd
}

func newSkillsGenerateCobraCommand(stdout io.Writer, jsonOut bool) *cobra.Command {
	var dirs []string
	cmd := &cobra.Command{
		Use:           "generate",
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		ValidArgsFunction: completeNoFile,
		RunE: func(_ *cobra.Command, _ []string) error {
			dir := ""
			if len(dirs) > 0 {
				dir = dirs[len(dirs)-1]
			}
			return markCobraLegacyError(runSkills(dir, stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "skills", stdout, jsonOut)
	cmd.Flags().StringArrayVar(&dirs, "dir", nil, "destination directory for generated skills")
	if flag := cmd.Flags().Lookup("dir"); flag != nil {
		flag.NoOptDefVal = "true"
	}
	return cmd
}

func newSkillsHelpCobraCommand(stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "help",
		Hidden:        true,
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		ValidArgsFunction: completeNoFile,
		RunE: func(_ *cobra.Command, _ []string) error {
			return markCobraLegacyError(writeManual("skills", stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "skills", stdout, jsonOut)
	return cmd
}

func newVersionCobraCommand(stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "version",
		Args:              cobra.NoArgs,
		SilenceErrors:     true,
		SilenceUsage:      true,
		ValidArgsFunction: completeNoFile,
		RunE: func(_ *cobra.Command, _ []string) error {
			return markCobraLegacyError(runVersion(stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "version", stdout, jsonOut)
	cmd.AddCommand(newVersionHelpCobraCommand(stdout, jsonOut))
	return cmd
}

func newVersionHelpCobraCommand(stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "help",
		Hidden:        true,
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		ValidArgsFunction: completeNoFile,
		RunE: func(_ *cobra.Command, _ []string) error {
			return markCobraLegacyError(writeManual("version", stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "version", stdout, jsonOut)
	return cmd
}

type perfCompareFlags struct {
	Iterations []string
	Runtime    []string
}

func (f perfCompareFlags) parsed() parsedFlags {
	values := map[string][]string{}
	if len(f.Iterations) > 0 {
		values["iterations"] = append([]string(nil), f.Iterations...)
	}
	if len(f.Runtime) > 0 {
		values["runtime"] = append([]string(nil), f.Runtime...)
	}
	return parsedFlags{values: values}
}

type perfSyncModesFlags struct {
	Records []string
}

func (f perfSyncModesFlags) parsed() parsedFlags {
	values := map[string][]string{}
	if len(f.Records) > 0 {
		values["records"] = append([]string(nil), f.Records...)
	}
	return parsedFlags{values: values}
}

func newPerfCobraCommand(ctx context.Context, cfg config.Config, stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "perf",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return markCobraLegacyError(writeManual("perf", stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "perf", stdout, jsonOut)
	cmd.AddCommand(newPerfCompareCobraCommand(ctx, cfg, stdout, jsonOut))
	cmd.AddCommand(newPerfSyncModesCobraCommand(ctx, stdout, jsonOut))
	return cmd
}

func newPerfCompareCobraCommand(ctx context.Context, cfg config.Config, stdout io.Writer, jsonOut bool) *cobra.Command {
	var flags perfCompareFlags
	cmd := &cobra.Command{
		Use:           "compare",
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		ValidArgsFunction: completeNoFile,
		RunE: func(_ *cobra.Command, _ []string) error {
			return markCobraLegacyError(runPerfCompare(ctx, cfg, flags.parsed(), stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "perf", stdout, jsonOut)
	addPerfStringArrayFlag(cmd, &flags.Iterations, "iterations", "number of benchmark iterations")
	addPerfStringArrayFlag(cmd, &flags.Runtime, "runtime", "also compare runtime-backed checks")
	return cmd
}

func newPerfSyncModesCobraCommand(ctx context.Context, stdout io.Writer, jsonOut bool) *cobra.Command {
	var flags perfSyncModesFlags
	cmd := &cobra.Command{
		Use:           "sync-modes",
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		ValidArgsFunction: completeNoFile,
		RunE: func(_ *cobra.Command, _ []string) error {
			return markCobraLegacyError(runPerfSyncModes(ctx, flags.parsed(), stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "perf", stdout, jsonOut)
	addPerfStringArrayFlag(cmd, &flags.Records, "records", "number of synthetic records")
	return cmd
}

func addPerfStringArrayFlag(cmd *cobra.Command, target *[]string, name string, usage string) {
	cmd.Flags().StringArrayVar(target, name, nil, usage)
	if flag := cmd.Flags().Lookup(name); flag != nil {
		flag.NoOptDefVal = "true"
	}
}

func newCatalogCobraCommand(ctx context.Context, root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "catalog",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return markCobraLegacyError(writeManual("catalog", stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "catalog", stdout, jsonOut)
	cmd.AddCommand(newCatalogActionCobraCommand(ctx, root, stdout, jsonOut, "refresh"))
	cmd.AddCommand(newCatalogActionCobraCommand(ctx, root, stdout, jsonOut, "show"))
	return cmd
}

func newCatalogActionCobraCommand(ctx context.Context, root string, stdout io.Writer, jsonOut bool, action string) *cobra.Command {
	var connections []string
	cmd := &cobra.Command{
		Use:           action,
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			connection := ""
			if len(connections) > 0 {
				connection = connections[len(connections)-1]
			}
			return markCobraLegacyError(withApp(root, func(a *app.App) error {
				return runCatalogAction(ctx, a, action, connection, stdout, jsonOut)
			}))
		},
	}
	setManualHelp(cmd, "catalog", stdout, jsonOut)
	cmd.Flags().StringArrayVar(&connections, "connection", nil, "connection name")
	if flag := cmd.Flags().Lookup("connection"); flag != nil {
		flag.NoOptDefVal = "true"
	}
	return cmd
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
				return markCobraLegacyError(writeManual(spec.name, stdout, jsonOut))
			}
			if len(args) == 0 && isManualCommand(spec.name) {
				return markCobraLegacyError(writeManual(spec.name, stdout, jsonOut))
			}
			return markCobraLegacyError(spec.handler(ctx, root, args, stdout, jsonOut))
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
	var legacy *cobraLegacyError
	if errors.As(err, &legacy) {
		return err
	}
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return err
	}
	// Non-legacy errors reaching this shim are produced by Cobra/pflag before a
	// command RunE handler starts. Treat all such parse/selection failures as
	// usage errors so native commands keep the CLI exit-code taxonomy.
	return usageErrorf("%s", message)
}
