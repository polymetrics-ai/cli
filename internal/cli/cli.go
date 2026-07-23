package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/term"

	"polymetrics.ai/internal/agentmode"
	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/events"
	pmlogging "polymetrics.ai/internal/logging"
	"polymetrics.ai/internal/perf"
	"polymetrics.ai/internal/runtimecheck"
	"polymetrics.ai/internal/safety"
	"polymetrics.ai/internal/telemetry"
	pmui "polymetrics.ai/internal/ui"
)

type envelope map[string]any

const maxConnectorCommandLimit = 10000

// RunMode controls how RunWithOptions evaluates the TTY gate.
type RunMode string

const (
	// ModeAuto evaluates the deterministic TTY gate.
	ModeAuto RunMode = "auto"
	// ModePlain forces the legacy plain path.
	ModePlain RunMode = "plain"
)

// RunOptions carries invocation-scoped UI detection facts for tests and future
// TTY renderers. Nil StdoutIsTerminal falls back to os.File + x/term detection.
type RunOptions struct {
	Mode                RunMode
	StdoutIsTerminal    *bool
	Env                 map[string]string
	ScheduleCrontabFile string
}

func Run(args []string, stdout, stderr io.Writer) int {
	return RunWithContext(context.Background(), args, stdout, stderr, RunOptions{Mode: ModePlain})
}

func RunWithOptions(args []string, stdout, stderr io.Writer, runOpts RunOptions) int {
	return RunWithContext(context.Background(), args, stdout, stderr, runOpts)
}

// RunWithContext preserves caller cancellation through the complete
// in-process command path. Run remains the compatibility wrapper for callers
// that do not have a context.
func RunWithContext(parent context.Context, args []string, stdout, stderr io.Writer, runOpts RunOptions) int {
	if parent == nil {
		parent = context.Background()
	}
	ctx := pmlogging.WithRegistry(parent, pmlogging.NewValueRegistry())
	globals, parseErr := parseGlobal(args)
	if parseErr != nil {
		return writeError(ctx, stdout, stderr, parseErr, globals.jsonOut)
	}
	if err := prevalidateCertifySafetyArgs(globals.clean); err != nil {
		return writeError(ctx, stdout, stderr, err, globals.jsonOut)
	}
	opts := config.Options{Root: globals.root, Flags: globalConfigFlags(args, globals.root, globals.jsonOut)}
	bootstrap, err := config.ResolveBootstrap(opts)
	if err != nil {
		return writeError(ctx, stdout, stderr, validationErrorf("%v", err), bootstrap.JSON)
	}
	cfg, err := config.Load(opts)
	if err != nil {
		return writeError(ctx, stdout, stderr, validationErrorf("%v", err), bootstrap.JSON)
	}
	if runOpts.ScheduleCrontabFile != "" {
		cfg.Schedule.CrontabFile = runOpts.ScheduleCrontabFile
	}
	logger, closeLogs := pmlogging.NewLogger(filepath.Join(cfg.Root, ".polymetrics"), stderr, pmlogging.LoggerOptions{Registry: pmlogging.RegistryFromContext(ctx)})
	defer func() { _ = closeLogs() }()
	ctx = pmlogging.WithLogger(ctx, logger)

	warnTelemetry := telemetryWarning(ctx, stderr)
	telemetryCfg, telemetryWarnings := telemetryConfig(cfg)
	for _, warning := range telemetryWarnings {
		warnTelemetry(warning)
	}
	ctx, telemetryHandle := telemetry.Init(ctx, telemetryCfg, warnTelemetry)
	_ = detectInvocationUI(globals, runOpts, stdout, cfg.JSON)
	if globals.progress == "ndjson" {
		ctx = events.WithEmitter(ctx, events.NewNDJSON(stderr))
	}

	ctx, commandSpan := telemetry.StartSpan(ctx, "pm.command", telemetry.StringAttr("pm.command.name", commandSpanName(globals.clean)))
	cmd := newRootCmd(ctx, cfg, stdout, stderr)
	exitCode := 0
	if err := executeRootCmd(cmd, globals.clean); err != nil {
		mapped := mapCobraErr(err)
		commandSpan.RecordError(classifyError(mapped))
		exitCode = writeError(ctx, stdout, stderr, mapped, cfg.JSON)
		commandSpan.SetAttributes(telemetry.StringAttr("pm.command.status", "error"), telemetry.IntAttr("pm.command.exit_code", exitCode))
	} else {
		commandSpan.SetAttributes(telemetry.StringAttr("pm.command.status", "ok"), telemetry.IntAttr("pm.command.exit_code", 0))
	}
	commandSpan.End()
	telemetry.Shutdown(ctx, telemetryHandle, warnTelemetry)
	return exitCode
}

func telemetryConfig(cfg config.Config) (telemetry.Config, []string) {
	dir := cfg.Telemetry.Directory
	if dir == "" {
		dir = filepath.Join(".polymetrics", "telemetry")
	}
	exporter := strings.TrimSpace(cfg.Telemetry.Exporter)
	endpoint := strings.TrimSpace(cfg.Telemetry.Endpoint)
	var warnings []string
	if strings.EqualFold(exporter, "otlp") {
		if !trustedTelemetrySource(cfg.Source("telemetry.exporter")) {
			exporter = string(telemetry.ExporterNone)
			endpoint = ""
			warnings = append(warnings, "config-sourced OTLP telemetry exporter ignored; set PM_TELEMETRY=otlp or POLYMETRICS_TELEMETRY=otlp to enable network telemetry")
		} else if endpoint != "" && !trustedTelemetrySource(cfg.Source("telemetry.endpoint")) {
			endpoint = ""
			warnings = append(warnings, "config-sourced OTLP telemetry endpoint ignored; set OTEL_EXPORTER_OTLP_ENDPOINT or PM_TELEMETRY_ENDPOINT to choose a collector")
		}
	}
	return telemetry.Config{
		Exporter:    telemetry.Exporter(exporter),
		Endpoint:    endpoint,
		Directory:   dir,
		ProjectRoot: cfg.Root,
		Capture:     cfg.Telemetry.Capture,
		RunID:       commandRunID(),
		ServiceName: "pm",
	}, warnings
}

func trustedTelemetrySource(source string) bool {
	return source == "env" || source == "flag"
}

func telemetryWarning(ctx context.Context, stderr io.Writer) telemetry.WarningFunc {
	return func(msg string) {
		fmt.Fprintf(stderr, "warning: telemetry: %s\n", pmlogging.RedactLine(ctx, msg))
	}
}

func commandRunID() string {
	return "trace-" + time.Now().UTC().Format("20060102T150405.000000000Z")
}

func commandSpanName(args []string) string {
	if len(args) == 0 {
		return "pm"
	}
	name := args[0]
	if name == "help" || name == "man" || name == "init" || name == "version" || name == "connectors" || name == "credentials" || name == "connections" || name == "catalog" || name == "etl" || name == "reverse" || name == "flow" || name == "query" || name == "runtime" || name == "perf" || name == "agent" || name == "rlm" || name == "schedule" || name == "docs" || name == "skills" || name == "worker" || name == "extract" {
		return name
	}
	return "connector"
}

func detectInvocationUI(globals globalOptions, runOpts RunOptions, stdout io.Writer, jsonOut bool) pmui.Detection {
	mode := runOpts.Mode
	if mode == "" {
		mode = ModeAuto
	}
	plain := globals.plain || mode == ModePlain
	return pmui.Detect(pmui.DetectOptions{
		StdoutTTY: stdoutIsTerminal(stdout, runOpts.StdoutIsTerminal),
		JSON:      jsonOut,
		Plain:     plain,
		NoInput:   globals.noInput,
		Env:       invocationEnv(runOpts.Env),
	})
}

func stdoutIsTerminal(stdout io.Writer, override *bool) bool {
	if override != nil {
		return *override
	}
	file, ok := stdout.(*os.File)
	if !ok || file == nil {
		return false
	}
	return term.IsTerminal(int(file.Fd()))
}

func invocationEnv(override map[string]string) map[string]string {
	keys := []string{"TERM", "PM_NO_TUI", "CI", "NO_COLOR", "CLICOLOR", "PM_ASCII"}
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if override != nil {
			out[key] = override[key]
			continue
		}
		out[key] = os.Getenv(key)
	}
	return out
}

func globalConfigFlags(args []string, root string, jsonOut bool) map[string]config.FlagValue {
	flags := map[string]config.FlagValue{}
	jsonChanged := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--json" || strings.HasPrefix(arg, "--json="):
			jsonChanged = true
		case arg == "--root" && i+1 < len(args):
			flags["root"] = config.StaticFlag{FlagName: "root", Value: root, Type: "string", Changed: true}
			i++
		case strings.HasPrefix(arg, "--root="):
			flags["root"] = config.StaticFlag{FlagName: "root", Value: root, Type: "string", Changed: true}
		}
	}
	if jsonChanged {
		value := "false"
		if jsonOut {
			value = "true"
		}
		flags["json"] = config.StaticFlag{FlagName: "json", Value: value, Type: "bool", Changed: true}
	}
	return flags
}

func writeRootManual(stdout io.Writer, jsonOut bool) error {
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "CommandManual", "command": "pm", "manual": rootHelp})
	}
	fmt.Fprint(stdout, rootHelp)
	return nil
}

func runInit(root string, stdout io.Writer, jsonOut bool) error {
	if err := app.InitProject(root); err != nil {
		return err
	}
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "InitResult", "project_dir": filepath.Join(root, ".polymetrics")})
	}
	fmt.Fprintf(stdout, "Initialized Polymetrics project at %s\n", filepath.Join(root, ".polymetrics"))
	return nil
}

func runHelp(args []string, stdout io.Writer, jsonOut bool) error {
	topic := ""
	if len(args) > 0 {
		topic = args[0]
	}
	return writeManual(topic, stdout, jsonOut)
}

func isManualCommand(cmd string) bool {
	if cmd == "init" || cmd == "help" || cmd == "man" || cmd == "version" {
		return false
	}
	_, ok := docs[cmd]
	return ok
}

func writeManual(topic string, stdout io.Writer, jsonOut bool) error {
	text, ok := docs[topic]
	if !ok {
		text, ok = dynamicConnectorManual(topic)
	}
	if !ok {
		return fmt.Errorf("help topic %q not found", topic)
	}
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "CommandManual", "command": topic, "manual": text})
	}
	fmt.Fprint(stdout, text)
	return nil
}

func dynamicConnectorManual(name string) (string, bool) {
	if err := safety.ValidateIdentifier(name, "connector"); err != nil {
		return "", false
	}
	connector, ok := appRegistry().Get(name)
	if !ok {
		return "", false
	}
	provider, ok := connector.(connectors.CommandSurfaceProvider)
	if !ok || provider.CommandSurface() == nil {
		return "", false
	}
	return connectors.RenderConnectorManual(connector), true
}

func definitionHasCapability(registry *connectors.Registry, def connectors.Definition, capability string) bool {
	switch capability {
	case "":
		return true
	case "read":
		return def.Capabilities.Read
	case "write":
		return def.Capabilities.Write
	case "query":
		return def.Capabilities.Query
	case "cdc":
		connector, ok := registry.Get(def.Name)
		if !ok {
			return false
		}
		_, ok = connector.(connectors.CDCReader)
		return ok
	default:
		return false
	}
}

type connectionsCreateFlags struct {
	Sources            []string
	Destinations       []string
	Streams            []string
	SyncModes          []string
	Cursors            []string
	PrimaryKeys        []string
	Tables             []string
	SourceConfigs      []string
	DestinationConfigs []string
}

func runConnectionsCreate(ctx context.Context, a *app.App, name string, flags connectionsCreateFlags, stdout io.Writer, jsonOut bool) error {
	source, err := parseEndpoint(lastString(flags.Sources))
	if err != nil {
		return err
	}
	dest, err := parseEndpoint(lastString(flags.Destinations))
	if err != nil {
		return err
	}
	stream := lastString(flags.Streams)
	if stream == "" {
		return errors.New("missing --stream")
	}
	sourceConfig, err := keyValues(flags.SourceConfigs)
	if err != nil {
		return err
	}
	destConfig, err := keyValues(flags.DestinationConfigs)
	if err != nil {
		return err
	}
	source.Config = sourceConfig
	dest.Config = destConfig
	streamCfg := app.StreamConfig{
		SyncMode:         valueOr(lastString(flags.SyncModes), "full_refresh_overwrite"),
		CursorField:      lastString(flags.Cursors),
		PrimaryKey:       flags.PrimaryKeys,
		DestinationTable: valueOr(lastString(flags.Tables), stream),
	}
	conn, err := a.CreateConnection(ctx, app.CreateConnectionRequest{
		Name:        name,
		Source:      source,
		Destination: dest,
		Streams:     map[string]app.StreamConfig{stream: streamCfg},
	})
	if err != nil {
		return err
	}
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "Connection", "connection": conn})
	}
	fmt.Fprintf(stdout, "Created connection %s\n", conn.Name)
	return nil
}

func runConnectionsList(a *app.App, stdout io.Writer, jsonOut bool) error {
	conns := a.ListConnections()
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "ConnectionList", "connections": conns})
	}
	for _, conn := range conns {
		fmt.Fprintf(stdout, "%s\t%s:%s -> %s:%s\n", conn.Name, conn.Source.Connector, conn.Source.Credential, conn.Destination.Connector, conn.Destination.Credential)
	}
	return nil
}

func lastString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[len(values)-1]
}

func runCatalogAction(ctx context.Context, a *app.App, action string, connection string, stdout io.Writer, jsonOut bool) error {
	if connection == "" {
		return errors.New("missing --connection")
	}
	var snapshot app.CatalogSnapshot
	var err error
	switch action {
	case "refresh":
		snapshot, err = a.RefreshCatalog(ctx, connection)
	case "show":
		snapshot, err = a.ShowCatalog(ctx, connection)
	default:
		return errUsage
	}
	if err != nil {
		return err
	}
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "Catalog", "catalog": snapshot})
	}
	for _, stream := range snapshot.Catalog.Streams {
		fmt.Fprintf(stdout, "%s\t%s\n", stream.Name, stream.Description)
	}
	return nil
}

func runMaybeConnectorCommand(ctx context.Context, root, connectorName string, args []string, stdout io.Writer, jsonOut bool) error {
	if err := safety.ValidateIdentifier(connectorName, "connector"); err != nil {
		return usageErrorf("unknown command %q", connectorName)
	}
	if err := connectors.RejectLegacyConnectorName(connectorName); err != nil {
		return err
	}
	registry := appRegistry()
	connector, ok := registry.Get(connectorName)
	if !ok {
		return usageErrorf("unknown command %q", connectorName)
	}
	surfaceProvider, ok := connector.(connectors.CommandSurfaceProvider)
	if !ok || surfaceProvider.CommandSurface() == nil {
		return usageErrorf("unknown command %q", connectorName)
	}
	surface := surfaceProvider.CommandSurface()
	if len(args) == 0 || connectorHelpRequested(args, surface) {
		manual := connectors.RenderConnectorManual(connector)
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "CommandManual", "command": connectorName, "manual": manual})
		}
		fmt.Fprint(stdout, manual)
		return nil
	}
	flags := parseFlags(args)
	path := flags.values["_"]
	if len(path) == 0 {
		return usageErrorf("missing connector command path")
	}
	if err := validateConnectorLifecycleFlagValues(flags); err != nil {
		return err
	}
	if err := commandrunner.Preflight(connector, path); err != nil {
		var blocked *commandrunner.BlockedCommandError
		if errors.As(err, &blocked) {
			return connectorCommandBlockedError(err)
		}
		return err
	}
	return withApp(root, func(a *app.App) error {
		return runConnectorCommand(ctx, a, connectorName, args, stdout, jsonOut)
	})
}

func connectorHelpRequested(args []string, surface *connectors.CommandSurface) bool {
	flags := parseFlags(args)
	if _, ok := flags.values["help"]; ok {
		return true
	}
	path := flags.values["_"]
	if len(path) == 0 {
		declared := map[string]bool{
			"credential": true, "connection": true, "config": true,
			"limit": true, "max-bytes": true,
		}
		for _, flag := range surface.GlobalFlags {
			declared[flag.Name] = true
		}
		for name := range flags.values {
			if name != "_" && !declared[name] {
				return false
			}
		}
		for _, name := range []string{"plan", "approve", "confirm"} {
			if _, ok := flags.values[name]; ok {
				return false
			}
		}
		if truthyFlag(flags.first("preview")) {
			return false
		}
		return true
	}
	if len(path) == 1 && path[0] == "help" {
		return true
	}
	for _, part := range path {
		if part == "-h" {
			return true
		}
	}
	return false
}

func runConnectorCommand(ctx context.Context, a *app.App, connectorName string, args []string, stdout io.Writer, jsonOut bool) error {
	flags := parseFlags(args)
	path := flags.values["_"]
	if len(path) == 0 {
		return usageErrorf("missing connector command path")
	}
	credential := flags.first("credential")
	if credential == "" {
		credential = flags.first("connection")
	}
	config, err := keyValues(flags.values["config"])
	if err != nil {
		return err
	}
	limit, err := parseIntFlag("limit", flags.first("limit"), 100)
	if err != nil {
		return err
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > maxConnectorCommandLimit {
		limit = maxConnectorCommandLimit
	}
	maxBytes, err := connectorCommandMaxBytes(flags)
	if err != nil {
		return err
	}
	commandFlags := map[string][]string{}
	for name, values := range flags.values {
		switch name {
		case "_", "credential", "connection", "config", "limit", "max-bytes", "plan", "preview", "approve", "confirm", "plan-name":
			continue
		default:
			commandFlags[name] = values
		}
	}

	if flags.first("plan") != "" {
		return runConnectorWriteCommandFromPlan(ctx, a, connectorName, path, flags, stdout, jsonOut)
	}

	connector, cfg, err := a.ResolveConnectorCredential(ctx, connectorName, credential, config)
	if err != nil {
		return err
	}

	if err := runConnectorWriteCommand(ctx, a, connectorName, credential, config, path, commandFlags, flags, stdout, jsonOut); err != commandrunner.ErrNotWriteCommand {
		if err != nil {
			var blocked *commandrunner.BlockedCommandError
			if errors.As(err, &blocked) {
				return connectorCommandBlockedError(err)
			}
			return err
		}
		return nil
	}

	rows := make([]connectors.Record, 0, limit)
	result, err := commandrunner.Run(ctx, connector, commandrunner.Request{
		Path:     path,
		Flags:    commandFlags,
		Config:   cfg,
		Limit:    limit,
		MaxBytes: maxBytes,
	}, func(record connectors.Record) error {
		rows = append(rows, record)
		return nil
	})
	if err != nil {
		var blocked *commandrunner.BlockedCommandError
		if errors.As(err, &blocked) {
			return connectorCommandBlockedError(err)
		}
		return err
	}
	if result.DirectRead != nil {
		if jsonOut {
			return writeJSON(stdout, envelope{
				"kind":      "ConnectorCommandDirectRead",
				"connector": result.Connector,
				"command":   result.Command,
				"method":    result.DirectRead.Method,
				"path":      result.DirectRead.Path,
				"status":    result.DirectRead.Status,
				"response":  result.DirectRead.Body,
			})
		}
		b, _ := json.MarshalIndent(result.DirectRead.Body, "", "  ")
		fmt.Fprintln(stdout, string(b))
		return nil
	}
	if jsonOut {
		return writeJSON(stdout, envelope{
			"kind":      "ConnectorCommandRead",
			"connector": result.Connector,
			"command":   result.Command,
			"stream":    result.Stream,
			"count":     result.Count,
			"records":   rows,
		})
	}
	for _, row := range rows {
		b, _ := json.Marshal(row)
		fmt.Fprintln(stdout, string(b))
	}
	return nil
}

func validateConnectorLifecycleFlagValues(flags parsedFlags) error {
	for _, name := range []string{"plan", "approve", "confirm"} {
		if _, ok := flags.values[name]; !ok {
			continue
		}
		for _, raw := range flags.values[name] {
			value := strings.TrimSpace(raw)
			if value == "" || value == "true" {
				return usageErrorf("--%s requires a value", name)
			}
		}
	}
	return nil
}

func connectorCommandMaxBytes(flags parsedFlags) (int, error) {
	maxBytes, err := parseIntFlag("max-bytes", flags.first("max-bytes"), commandrunner.MaxOperationDirectReadBytes)
	if err != nil {
		return 0, err
	}
	if maxBytes <= 0 {
		maxBytes = commandrunner.MaxOperationDirectReadBytes
	}
	if maxBytes > commandrunner.MaxOperationDirectReadBytes {
		maxBytes = commandrunner.MaxOperationDirectReadBytes
	}
	return maxBytes, nil
}

func runConnectorWriteCommand(ctx context.Context, a *app.App, connectorName, credential string, config map[string]string, path []string, commandFlags map[string][]string, flags parsedFlags, stdout io.Writer, jsonOut bool) error {
	preview := truthyFlag(flags.first("preview"))
	plan, writePreview, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
		Name:       flags.first("plan-name"),
		Connector:  connectorName,
		Credential: credential,
		Config:     config,
		Path:       path,
		Flags:      commandFlags,
		Preview:    preview,
	})
	if err != nil {
		return err
	}
	if jsonOut {
		env := envelope{"kind": "ConnectorCommandWritePlan", "plan": safeReversePlanForOutput(plan), "approval_required": true}
		if writePreview != nil {
			env["write_preview"] = writePreview
		}
		return writeJSON(stdout, env)
	}
	fmt.Fprintf(stdout, "Created connector command plan %s for %s\nApproval token: %s\n", plan.ID, plan.ConnectorCommand, plan.ApprovalToken)
	if plan.ConfirmationChallenge != "" {
		fmt.Fprintf(stdout, "Confirmation required: --confirm %s\n", plan.ConfirmationChallenge)
	}
	if writePreview != nil {
		for _, warning := range writePreview.Warnings {
			fmt.Fprintf(stdout, "- %s\n", warning)
		}
	}
	return nil
}

func runConnectorWriteCommandFromPlan(ctx context.Context, a *app.App, connectorName string, path []string, flags parsedFlags, stdout io.Writer, jsonOut bool) error {
	planID := strings.TrimSpace(flags.first("plan"))
	approvalToken := strings.TrimSpace(flags.first("approve"))
	preview := truthyFlag(flags.first("preview"))
	plan, err := connectorCommandPlanForPath(a, planID, connectorName, path)
	if err != nil {
		return err
	}
	if approvalToken != "" {
		run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: approvalToken, Confirmation: flags.first("confirm")})
		if err != nil {
			return err
		}
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "ReverseRun", "run": run})
		}
		fmt.Fprintf(stdout, "Reverse ETL run %s completed: succeeded=%d failed=%d\n", run.ID, run.RecordsSucceeded, run.RecordsFailed)
		return nil
	}
	if preview {
		plan, writePreview, err := a.PreviewConnectorCommandPlan(ctx, plan.ID)
		if err != nil {
			return err
		}
		if jsonOut {
			return writeJSON(stdout, envelope{
				"kind":          "ConnectorCommandWritePreview",
				"plan":          safeReversePlanForOutput(plan),
				"write_preview": writePreview,
			})
		}
		fmt.Fprintf(stdout, "Reverse plan %s previews %s via %s\n", plan.ID, plan.ConnectorCommand, plan.Action)
		for _, warning := range writePreview.Warnings {
			fmt.Fprintf(stdout, "- %s\n", warning)
		}
		return nil
	}
	return usageErrorf("connector write command with --plan requires --preview or --approve")
}

func connectorCommandPlanForPath(a *app.App, planID, connectorName string, path []string) (app.ReversePlan, error) {
	plan, err := a.GetReversePlan(planID)
	if err != nil {
		return app.ReversePlan{}, err
	}
	if plan.ConnectorCommand == "" || len(plan.ConnectorCommandPath) == 0 {
		return app.ReversePlan{}, usageErrorf("reverse plan %q is not a connector command plan", planID)
	}
	if plan.DestinationConnector != connectorName {
		return app.ReversePlan{}, validationErrorf("reverse plan %q targets connector %q, not %q", planID, plan.DestinationConnector, connectorName)
	}
	if !sameStringSlice(plan.ConnectorCommandPath, path) {
		return app.ReversePlan{}, validationErrorf("reverse plan %q targets command %q, not %q", planID, strings.Join(plan.ConnectorCommandPath, " "), strings.Join(path, " "))
	}
	return plan, nil
}

func connectorCommandBlockedError(err error) error {
	return &cliError{
		category: categoryPolicy,
		code:     "connector_command_blocked",
		message:  err.Error(),
		err:      err,
	}
}

func truthyFlag(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "false", "0", "no":
		return false
	default:
		return true
	}
}

func safeReversePlansForOutput(plans []app.ReversePlan) []app.ReversePlan {
	out := make([]app.ReversePlan, 0, len(plans))
	for _, plan := range plans {
		out = append(out, safeReversePlanForOutput(plan))
	}
	return out
}

func safeReversePlanForOutput(plan app.ReversePlan) app.ReversePlan {
	plan.ApprovalToken = ""
	plan.ApprovalTokenHash = ""
	plan.ConnectorCommandRecord = nil
	return plan
}

func sameStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func runQueryRun(ctx context.Context, a *app.App, flags parsedFlags, stdout io.Writer, jsonOut bool) error {
	limit, err := parseIntFlag("limit", valueOr(flags.first("limit"), "100"), 100)
	if err != nil {
		return err
	}
	var rows []connectors.Record
	if sql := flags.first("sql"); sql != "" {
		rows, err = a.QuerySQL(ctx, sql, limit)
	} else {
		rows, err = a.QueryTable(ctx, app.QueryTableRequest{Table: flags.first("table"), Limit: limit})
	}
	if err != nil {
		return err
	}
	if fields := parseCSVFlags(flags.values["fields"]); len(fields) > 0 {
		rows = agentmode.FieldsProjection(rows, fields)
	}
	if mode := strings.TrimSpace(flags.first("agent-mode")); mode != "" {
		return writeAgentModeQuery(stdout, rows, mode, flags)
	}
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "QueryResult", "rows": rows, "count": len(rows)})
	}
	for _, row := range rows {
		b, _ := json.Marshal(row)
		fmt.Fprintln(stdout, string(b))
	}
	return nil
}

func parseCSVFlags(values []string) []string {
	var out []string
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			item = strings.TrimSpace(item)
			if item != "" {
				out = append(out, item)
			}
		}
	}
	return out
}

func writeAgentModeQuery(stdout io.Writer, rows []connectors.Record, mode string, flags parsedFlags) error {
	switch mode {
	case "summary":
		sampleN, err := parseIntFlag("sample", valueOr(flags.first("sample"), "3"), 3)
		if err != nil {
			return err
		}
		payload, err := agentmode.Summarize("QueryResult", rows, sampleN)
		if err != nil {
			return err
		}
		if _, err := stdout.Write(payload); err != nil {
			return err
		}
		_, err = stdout.Write([]byte{'\n'})
		return err
	case "stream":
		return agentmode.EncodeStream(stdout, rows)
	default:
		return usageErrorf("query run: unknown --agent-mode %q (want summary|stream)", mode)
	}
}

func runAgentPlan(request string, stdout io.Writer, jsonOut bool) error {
	if err := safety.RejectDangerousChars(request, "agent request"); err != nil {
		return validationErrorf("%v", err)
	}
	req := strings.ToLower(request)
	steps := []string{"pm connectors list --json", "pm help etl"}
	if strings.Contains(req, "sample") && strings.Contains(req, "customers") {
		steps = []string{
			"pm credentials add sample-local --connector sample",
			"pm credentials add warehouse-local --connector warehouse",
			"pm connections create sample_to_warehouse --source sample:sample-local --destination warehouse:warehouse-local --stream customers --primary-key id --table sample_customers",
			"pm etl run --connection sample_to_warehouse --stream customers --json",
		}
	}
	result := envelope{"kind": "AgentPlan", "risk": "read_local", "steps": steps, "safety": "No secrets or approval tokens are returned."}
	if jsonOut {
		return writeJSON(stdout, result)
	}
	for _, step := range steps {
		fmt.Fprintln(stdout, step)
	}
	return nil
}

type docsCommandFlags struct {
	Dirs           []string
	ConnectorsDirs []string
}

func (f docsCommandFlags) dir() string {
	return lastDocsFlag(f.Dirs)
}

func (f docsCommandFlags) connectorsDir() string {
	return lastDocsFlag(f.ConnectorsDirs)
}

func lastDocsFlag(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[len(values)-1]
}

func runDocs(action string, flags docsCommandFlags, stdout io.Writer) error {
	switch action {
	case "generate":
		dir := flags.dir()
		if dir == "" {
			return errors.New("missing --dir")
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		for topic, text := range docs {
			if topic == "" || topic == "pm" {
				continue
			}
			path := filepath.Join(dir, topic+".md")
			if err := os.WriteFile(path, []byte("```\n"+text+"\n```\n"), 0o644); err != nil {
				return err
			}
		}
		connectorsDir := valueOr(flags.connectorsDir(), filepath.Join(filepath.Dir(dir), "connectors"))
		if err := writeConnectorDocs(connectorsDir, appRegistry()); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Generated docs in %s and connector docs in %s\n", dir, connectorsDir)
		return nil
	case "validate":
		dir := valueOr(flags.connectorsDir(), valueOr(flags.dir(), "docs/connectors"))
		if err := validateConnectorDocs(dir, appRegistry()); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Validated connector docs in %s\n", dir)
		return nil
	default:
		return errUsage
	}
}

func runRuntimeDoctor(ctx context.Context, cfg config.Config, stdout io.Writer, jsonOut bool) error {
	runtimeCfg := runtimecheck.FromConfig(cfg)
	report := runtimecheck.Doctor(ctx, runtimeCfg)
	if jsonOut {
		return writeJSON(stdout, envelope{
			"kind":   "RuntimeDoctor",
			"config": runtimecheck.RedactedConfig(runtimeCfg),
			"report": report,
		})
	}
	fmt.Fprintf(stdout, "mode=%s duration=%s\n", report.Mode, report.Duration)
	for _, check := range report.Checks {
		if check.Error != "" {
			fmt.Fprintf(stdout, "%s\t%s\t%s\t%s\t%s\n", check.Name, check.Status, check.Endpoint, check.Latency, check.Error)
			continue
		}
		fmt.Fprintf(stdout, "%s\t%s\t%s\t%s\n", check.Name, check.Status, check.Endpoint, check.Latency)
	}
	return nil
}

func runPerfCompare(ctx context.Context, cfg config.Config, flags parsedFlags, stdout io.Writer, jsonOut bool) error {
	iterations, err := parsePositiveIntFlag("iterations", flags.first("iterations"), perf.DefaultCompareIterations, perf.MaxCompareIterations)
	if err != nil {
		return err
	}
	comparison, err := perf.Compare(ctx, perf.CompareRequest{
		Iterations:    iterations,
		Runtime:       flags.first("runtime") == "true",
		RuntimeConfig: runtimecheck.FromConfig(cfg),
	})
	if err != nil {
		return err
	}
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "PerformanceComparison", "comparison": comparison})
	}
	printPerfResult(stdout, comparison.DependencyFree)
	if comparison.RuntimeBacked != nil {
		printPerfResult(stdout, *comparison.RuntimeBacked)
	}
	fmt.Fprintf(stdout, "\nDependency-free: %s\n", comparison.Explanation["dependency_free"])
	fmt.Fprintf(stdout, "Runtime-backed: %s\n", comparison.Explanation["runtime_backed"])
	return nil
}

func runPerfSyncModes(ctx context.Context, flags parsedFlags, stdout io.Writer, jsonOut bool) error {
	records, err := parsePositiveIntFlag("records", flags.first("records"), perf.DefaultSyncModeRecords, perf.MaxSyncModeRecords)
	if err != nil {
		return err
	}
	benchmark, err := perf.CompareSyncModes(ctx, perf.SyncModeBenchmarkRequest{
		Records: records,
	})
	if err != nil {
		return err
	}
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "SyncModeBenchmark", "benchmark": benchmark})
	}
	for _, result := range benchmark.Results {
		fmt.Fprintf(stdout, "%s\trecords=%d\tduration=%s\trecords_per_sec=%.2f", result.Mode, result.Records, result.Duration, result.RecordsPerSec)
		if result.Error != "" {
			fmt.Fprintf(stdout, "\terror=%s", result.Error)
		}
		fmt.Fprintln(stdout)
	}
	fmt.Fprintln(stdout, benchmark.Explanation)
	return nil
}

func printPerfResult(stdout io.Writer, result perf.Result) {
	if result.Error != "" {
		fmt.Fprintf(stdout, "%s\titerations=%d\terror=%s\n", result.Mode, result.Iterations, result.Error)
		return
	}
	fmt.Fprintf(stdout, "%s\titerations=%d\trecords=%d\tduration=%s\tavg=%s\trecords_per_sec=%.2f\n",
		result.Mode,
		result.Iterations,
		result.Records,
		result.Duration,
		result.Average,
		result.RecordsPerSec,
	)
}

func withApp(root string, fn func(*app.App) error) error {
	a, err := app.Open(root)
	if err != nil {
		return err
	}
	return fn(a)
}

func validateCredentialConfig(a *app.App, connector string, config map[string]string) error {
	path := config["path"]
	if path == "" {
		return nil
	}
	switch connector {
	case "warehouse", "outbox":
		allowExternal := strings.EqualFold(config["allow_external_path"], "true")
		if err := safety.ValidateLocalWritePath(filepath.Dir(a.ProjectDir()), path, connector+" path", allowExternal); err != nil {
			return validationErrorf("%v", err)
		}
	default:
		if err := safety.RejectDangerousChars(path, connector+" path"); err != nil {
			return validationErrorf("%v", err)
		}
	}
	return nil
}

func appRegistry() *connectors.Registry {
	return bundleregistry.New()
}
