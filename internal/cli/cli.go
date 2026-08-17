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

	"polymetrics.ai/internal/agentmode"
	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/perf"
	"polymetrics.ai/internal/runtimecheck"
	"polymetrics.ai/internal/safety"
)

type envelope map[string]any

const maxConnectorCommandLimit = 10000

func Run(args []string, stdout, stderr io.Writer) int {
	ctx := context.Background()
	root, jsonOut, cleanArgs := parseGlobal(args)
	opts := config.Options{Root: root, Flags: globalConfigFlags(args, root, jsonOut)}
	bootstrap, err := config.ResolveBootstrap(opts)
	if err != nil {
		return writeError(stdout, stderr, validationErrorf("%v", err), bootstrap.JSON)
	}
	cfg, err := config.Load(opts)
	if err != nil {
		return writeError(stdout, stderr, validationErrorf("%v", err), bootstrap.JSON)
	}
	cmd := newRootCmd(ctx, cfg, stdout, stderr)
	if err := executeRootCmd(cmd, cleanArgs); err != nil {
		return writeError(stdout, stderr, mapCobraErr(err), cfg.JSON)
	}
	return 0
}

func globalConfigFlags(args []string, root string, jsonOut bool) map[string]config.FlagValue {
	flags := map[string]config.FlagValue{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--json":
			flags["json"] = config.StaticFlag{FlagName: "json", Value: "true", Type: "bool", Changed: true}
		case arg == "--root" && i+1 < len(args):
			flags["root"] = config.StaticFlag{FlagName: "root", Value: root, Type: "string", Changed: true}
			i++
		case strings.HasPrefix(arg, "--root="):
			flags["root"] = config.StaticFlag{FlagName: "root", Value: root, Type: "string", Changed: true}
		}
	}
	if jsonOut {
		flags["json"] = config.StaticFlag{FlagName: "json", Value: "true", Type: "bool", Changed: true}
	}
	return flags
}

func writeRootManual(stdout io.Writer, jsonOut bool) error {
	manual := rootManual()
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "CommandManual", "command": "pm", "manual": manual})
	}
	_, err := fmt.Fprint(stdout, manual)
	return err
}

func rootManual() string {
	section := dynamicConnectorCommandsSection(appRegistry())
	if section == "" {
		return rootHelp
	}
	return strings.TrimRight(rootHelp, "\n") + "\n\n" + section
}

func dynamicConnectorCommandsSection(registry *connectors.Registry) string {
	if registry == nil {
		return ""
	}
	lines := []string{}
	for _, meta := range registry.List() {
		connector, ok := registry.Get(meta.Name)
		if !ok {
			continue
		}
		provider, ok := connector.(connectors.CommandSurfaceProvider)
		if !ok || provider.CommandSurface() == nil || strings.TrimSpace(provider.CommandSurface().Usage) == "" {
			continue
		}
		line := fmt.Sprintf("  pm %s <command>", meta.Name)
		if meta.DisplayName != "" {
			line += " - " + meta.DisplayName
		}
		if provider.CommandSurface().Tagline != "" {
			line += ": " + provider.CommandSurface().Tagline
		}
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("CONNECTOR COMMANDS\n")
	b.WriteString("  Some connectors expose provider-style command surfaces in addition to pm connectors inspect.\n")
	for _, line := range lines {
		b.WriteString(line + "\n")
	}
	b.WriteString("  Run pm <connector> --help for command groups and pm <connector> <path> --help for exact flags.\n")
	return b.String()
}

func runInit(root string, stdout io.Writer, jsonOut bool) error {
	if err := app.InitProject(root); err != nil {
		return err
	}
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "InitResult", "project_dir": filepath.Join(root, ".polymetrics")})
	}
	_, _ = fmt.Fprintf(stdout, "Initialized Polymetrics project at %s\n", filepath.Join(root, ".polymetrics"))
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
	_, _ = fmt.Fprint(stdout, text)
	return nil
}

func dynamicConnectorManual(name string) (string, bool) {
	connector, ok := dynamicConnectorWithCommandSurface(name)
	if !ok {
		return "", false
	}
	return connectors.RenderConnectorManual(connector), true
}

func isDynamicConnectorCommand(name string) bool {
	_, ok := dynamicConnectorWithCommandSurface(name)
	return ok
}

func dynamicConnectorWithCommandSurface(name string) (connectors.Connector, bool) {
	if err := safety.ValidateIdentifier(name, "connector"); err != nil {
		return nil, false
	}
	connector, ok := appRegistry().Get(name)
	if !ok {
		return nil, false
	}
	provider, ok := connector.(connectors.CommandSurfaceProvider)
	if !ok || provider.CommandSurface() == nil {
		return nil, false
	}
	return connector, true
}

func runConnectors(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
	registry := appRegistry()
	if len(args) == 0 {
		return errUsage
	}
	switch args[0] {
	case "certify":
		return runCertify(ctx, root, args[1:], stdout, jsonOut)
	case "list":
		flags := parseFlags(args[1:])
		if flags.first("all") != "" {
			defs, err := connectorCatalogEntries(registry, flags)
			if err != nil {
				return err
			}
			if jsonOut {
				return writeJSON(stdout, envelope{"kind": "ConnectorCatalog", "count": len(defs), "connectors": defs})
			}
			for _, item := range defs {
				_, _ = fmt.Fprintf(stdout, "%s\t%s\tread=%t\twrite=%t\tquery=%t\n", item.Name, item.IntegrationType, item.Capabilities.Read, item.Capabilities.Write, item.Capabilities.Query)
			}
			return nil
		}
		list := registry.List()
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "ConnectorList", "connectors": list})
		}
		for _, item := range list {
			_, _ = fmt.Fprintf(stdout, "%s\t%s\t%+v\n", item.Name, item.IntegrationType, item.Capabilities)
		}
		return nil
	case "catalog":
		flags := parseFlags(args[1:])
		defs, err := connectorCatalogEntries(registry, flags)
		if err != nil {
			return err
		}
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "ConnectorCatalog", "count": len(defs), "connectors": defs})
		}
		for _, item := range defs {
			_, _ = fmt.Fprintf(stdout, "%s\t%s\tread=%t\twrite=%t\tquery=%t\n", item.Name, item.IntegrationType, item.Capabilities.Read, item.Capabilities.Write, item.Capabilities.Query)
		}
		return nil
	case "inspect", "help", "man", "docs":
		if len(args) < 2 {
			return errUsage
		}
		if err := safety.ValidateIdentifier(args[1], "connector"); err != nil {
			return validationErrorf("%v", err)
		}
		if err := connectors.RejectLegacyConnectorName(args[1]); err != nil {
			return err
		}
		if c, ok := registry.Get(args[1]); ok {
			if jsonOut {
				return writeJSON(stdout, envelope{"kind": "Connector", "connector": connectors.MetadataWithIcon(c.Metadata()), "manifest": connectors.ManifestOf(c)})
			}
			_, _ = fmt.Fprint(stdout, connectors.RenderConnectorManual(c))
			return nil
		}
		return fmt.Errorf("connector %q not found", args[1])
	default:
		return errUsage
	}
}

func connectorCatalogEntries(registry *connectors.Registry, flags parsedFlags) ([]connectors.Definition, error) {
	if flags.first("type") != "" {
		return nil, validationErrorf("legacy --type source|destination was removed; use --capability read|write|cdc|query")
	}
	capability := strings.TrimSpace(strings.ToLower(flags.first("capability")))
	switch capability {
	case "", "read", "write", "cdc", "query":
	default:
		return nil, validationErrorf("invalid --capability %q, want read|write|cdc|query", capability)
	}
	stage := strings.TrimSpace(flags.first("stage"))
	defs := registry.CatalogEntries()
	out := make([]connectors.Definition, 0, len(defs))
	for _, def := range defs {
		if stage != "" && def.ReleaseStage != stage {
			continue
		}
		if !definitionHasCapability(registry, def, capability) {
			continue
		}
		out = append(out, def)
	}
	return out, nil
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

func runCredentials(ctx context.Context, a *app.App, args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 {
		return errUsage
	}
	switch args[0] {
	case "add":
		if len(args) < 2 {
			return errUsage
		}
		flags := parseFlags(args[2:])
		connector := flags.first("connector")
		if connector == "" {
			return errors.New("missing --connector")
		}
		if err := safety.ValidateIdentifier(args[1], "credential"); err != nil {
			return validationErrorf("%v", err)
		}
		if err := safety.ValidateIdentifier(connector, "connector"); err != nil {
			return validationErrorf("%v", err)
		}
		if err := connectors.RejectLegacyConnectorName(connector); err != nil {
			return err
		}
		secrets := map[string]string{}
		for _, spec := range flags.values["from-env"] {
			key, env, ok := strings.Cut(spec, "=")
			if !ok || key == "" || env == "" {
				return fmt.Errorf("invalid --from-env %q, want field=ENV", spec)
			}
			secrets[key] = os.Getenv(env)
			if secrets[key] == "" {
				return fmt.Errorf("environment variable %s is empty", env)
			}
		}
		if field := flags.first("value-stdin"); field != "" {
			b, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("read stdin secret: %w", err)
			}
			secrets[field] = strings.TrimRight(string(b), "\r\n")
		}
		config, err := keyValues(flags.values["config"])
		if err != nil {
			return err
		}
		if err := validateCredentialConfig(a, connector, config); err != nil {
			return err
		}
		cred, err := a.AddCredential(ctx, app.AddCredentialRequest{
			Name:      args[1],
			Connector: connector,
			Config:    config,
			Secrets:   secrets,
		})
		if err != nil {
			return err
		}
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "Credential", "credential": cred})
		}
		_, _ = fmt.Fprintf(stdout, "Saved credential %s for connector %s\n", cred.Name, cred.Connector)
		return nil
	case "list":
		creds := a.ListCredentials()
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "CredentialList", "credentials": creds})
		}
		for _, cred := range creds {
			_, _ = fmt.Fprintf(stdout, "%s\t%s\t%s\n", cred.Name, cred.ID, cred.Connector)
		}
		return nil
	case "inspect":
		if len(args) < 2 {
			return errUsage
		}
		cred, err := a.InspectCredential(args[1])
		if err != nil {
			return err
		}
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "Credential", "credential": cred})
		}
		b, _ := json.MarshalIndent(cred, "", "  ")
		_, _ = fmt.Fprintln(stdout, string(b))
		return nil
	case "test":
		if len(args) < 2 {
			return errUsage
		}
		cred, err := a.TestCredential(ctx, args[1])
		if err != nil {
			return err
		}
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "CredentialTest", "status": "ok", "credential": cred})
		}
		_, _ = fmt.Fprintf(stdout, "Credential %s validated\n", cred.Name)
		return nil
	case "remove":
		if len(args) < 2 {
			return errUsage
		}
		if err := a.RemoveCredential(ctx, args[1]); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(stdout, "Removed credential %s\n", args[1])
		return nil
	default:
		return errUsage
	}
}

func runConnections(ctx context.Context, a *app.App, args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 {
		return errUsage
	}
	switch args[0] {
	case "create":
		if len(args) < 2 {
			return errUsage
		}
		flags := parseFlags(args[2:])
		source, err := parseEndpoint(flags.first("source"))
		if err != nil {
			return err
		}
		dest, err := parseEndpoint(flags.first("destination"))
		if err != nil {
			return err
		}
		stream := flags.first("stream")
		if stream == "" {
			return errors.New("missing --stream")
		}
		sourceConfig, err := keyValues(flags.values["source-config"])
		if err != nil {
			return err
		}
		destConfig, err := keyValues(flags.values["destination-config"])
		if err != nil {
			return err
		}
		source.Config = sourceConfig
		dest.Config = destConfig
		streamCfg := app.StreamConfig{
			SyncMode:         valueOr(flags.first("sync-mode"), "full_refresh_overwrite"),
			CursorField:      flags.first("cursor"),
			PrimaryKey:       flags.values["primary-key"],
			DestinationTable: valueOr(flags.first("table"), stream),
		}
		conn, err := a.CreateConnection(ctx, app.CreateConnectionRequest{
			Name:        args[1],
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
		_, _ = fmt.Fprintf(stdout, "Created connection %s\n", conn.Name)
		return nil
	case "list":
		conns := a.ListConnections()
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "ConnectionList", "connections": conns})
		}
		for _, conn := range conns {
			_, _ = fmt.Fprintf(stdout, "%s\t%s:%s -> %s:%s\n", conn.Name, conn.Source.Connector, conn.Source.Credential, conn.Destination.Connector, conn.Destination.Credential)
		}
		return nil
	default:
		return errUsage
	}
}

func runCatalog(ctx context.Context, a *app.App, args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 {
		return errUsage
	}
	flags := parseFlags(args[1:])
	connection := flags.first("connection")
	if connection == "" {
		return errors.New("missing --connection")
	}
	var snapshot app.CatalogSnapshot
	var err error
	switch args[0] {
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
		_, _ = fmt.Fprintf(stdout, "%s\t%s\n", stream.Name, stream.Description)
	}
	return nil
}

func runETL(ctx context.Context, a *app.App, args []string, stdout io.Writer, jsonOut bool, cfg config.Config) error {
	if len(args) == 0 {
		return errUsage
	}
	switch args[0] {
	case "check":
		connector, cfg, err := directConnector(a, args[1:])
		if err != nil {
			return err
		}
		if err := connector.Check(ctx, cfg); err != nil {
			return err
		}
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "ETLCheck", "connector": connector.Name(), "status": "ok"})
		}
		_, _ = fmt.Fprintf(stdout, "Connector %s check ok\n", connector.Name())
		return nil
	case "catalog":
		connector, cfg, err := directConnector(a, args[1:])
		if err != nil {
			return err
		}
		catalog, err := connector.Catalog(ctx, cfg)
		if err != nil {
			return err
		}
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "ETLCatalog", "connector": connector.Name(), "catalog": catalog})
		}
		for _, stream := range catalog.Streams {
			_, _ = fmt.Fprintf(stdout, "%s\t%s\n", stream.Name, stream.Description)
		}
		return nil
	case "read":
		flags := parseFlags(args[1:])
		connector, cfg, err := directConnector(a, args[1:])
		if err != nil {
			return err
		}
		stream := flags.first("stream")
		limit, err := parseIntFlag("limit", valueOr(flags.first("limit"), "100"), 100)
		if err != nil {
			return err
		}
		if limit <= 0 {
			limit = 100
		}
		rows := make([]connectors.Record, 0, limit)
		err = connector.Read(ctx, connectors.ReadRequest{Stream: stream, Config: cfg, Limit: limit}, connectors.LimitEmitter(limit, func(record connectors.Record) error {
			rows = append(rows, record)
			return nil
		}))
		if err := connectors.IgnoreReadLimit(err); err != nil {
			return err
		}
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "ETLRead", "connector": connector.Name(), "stream": stream, "count": len(rows), "records": rows})
		}
		for _, row := range rows {
			b, _ := json.Marshal(row)
			_, _ = fmt.Fprintln(stdout, string(b))
		}
		return nil
	case "run":
		flags := parseFlags(args[1:])
		batchSize, err := parseIntFlag("batch-size", flags.first("batch-size"), 0)
		if err != nil {
			return err
		}
		run, err := a.RunETL(ctx, app.RunETLRequest{
			Connection: flags.first("connection"),
			Stream:     flags.first("stream"),
			BatchSize:  batchSize,
		})
		if err != nil {
			return err
		}
		runtimeRecorded := false
		if flags.first("runtime") == "true" {
			if err := recordRuntimeETL(ctx, run, cfg); err != nil {
				return err
			}
			runtimeRecorded = true
		}
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "ETLRun", "run": run, "runtime_recorded": runtimeRecorded})
		}
		if runtimeRecorded {
			_, _ = fmt.Fprintf(stdout, "ETL run %s completed: read=%d loaded=%d failed=%d runtime_recorded=true\n", run.ID, run.RecordsRead, run.RecordsLoaded, run.RecordsFailed)
			return nil
		}
		_, _ = fmt.Fprintf(stdout, "ETL run %s completed: read=%d loaded=%d failed=%d\n", run.ID, run.RecordsRead, run.RecordsLoaded, run.RecordsFailed)
		return nil
	case "status":
		if len(args) < 2 {
			return errUsage
		}
		run, err := a.GetRun(args[1])
		if err != nil {
			return err
		}
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "ETLRun", "run": run})
		}
		_, _ = fmt.Fprintf(stdout, "%s\t%s\tread=%d loaded=%d failed=%d\n", run.ID, run.Status, run.RecordsRead, run.RecordsLoaded, run.RecordsFailed)
		return nil
	default:
		return errUsage
	}
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
		command, manual := renderConnectorCommandManual(connectorName, connector, surface, args)
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "CommandManual", "command": command, "manual": manual})
		}
		_, _ = fmt.Fprint(stdout, manual)
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
			return connectorCommandBlockedError(withConnectorCommandSuggestion(blocked, surface, path))
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
	if len(path) > 0 && (path[0] == "help" || path[len(path)-1] == "help") {
		return true
	}
	if len(path) == 0 {
		return connectorHelpFlagsArePassive(flags, surface)
	}
	if len(path) == 1 && path[0] == "help" {
		return true
	}
	if connectorBareCommandGroupHelpRequested(flags, surface, path) {
		return true
	}
	for _, part := range path {
		if part == "-h" {
			return true
		}
	}
	return false
}

func connectorBareCommandGroupHelpRequested(flags parsedFlags, surface *connectors.CommandSurface, path []string) bool {
	if len(path) != 1 {
		return false
	}
	if _, ok := connectorSurfaceCommand(surface, path[0]); ok {
		return false
	}
	if !connectorSurfaceHasPrefix(surface, path[0]) {
		return false
	}
	return connectorHelpFlagsArePassive(flags, surface)
}

func connectorHelpFlagsArePassive(flags parsedFlags, surface *connectors.CommandSurface) bool {
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
	return !truthyFlag(flags.first("preview"))
}

func renderConnectorCommandManual(connectorName string, connector connectors.Connector, surface *connectors.CommandSurface, args []string) (string, string) {
	flags := parseFlags(args)
	path := connectorHelpPath(flags.values["_"])
	if len(path) > 0 {
		command := strings.Join(path, " ")
		if cmd, ok := connectorSurfaceCommand(surface, command); ok {
			return connectorName + " " + command, renderConnectorCommandDetail(connectorName, surface, cmd)
		}
		if len(path) == 1 && connectorSurfaceHasPrefix(surface, path[0]) {
			return connectorName + " " + path[0], renderConnectorCommandGroup(connectorName, connector, surface, path[0])
		}
	}
	return connectorName, renderConnectorCommandRoot(connectorName, connector, surface)
}

func connectorHelpPath(path []string) []string {
	out := make([]string, 0, len(path))
	for i, part := range path {
		switch {
		case part == "-h" || part == "--help":
			continue
		case part == "help" && (i == 0 || i == len(path)-1):
			continue
		default:
			out = append(out, part)
		}
	}
	return out
}

func renderConnectorCommandRoot(connectorName string, connector connectors.Connector, surface *connectors.CommandSurface) string {
	meta := connector.Metadata()
	var b strings.Builder
	b.WriteString("NAME\n")
	fmt.Fprintf(&b, "  pm %s - %s command surface\n\n", connectorName, valueOr(meta.DisplayName, connectorName))
	b.WriteString("SYNOPSIS\n")
	fmt.Fprintf(&b, "  pm %s <command> [flags]\n", connectorName)
	fmt.Fprintf(&b, "  pm %s <group> --help\n", connectorName)
	fmt.Fprintf(&b, "  pm %s <group> <command> --help\n\n", connectorName)
	b.WriteString("DESCRIPTION\n")
	if surface.Tagline != "" {
		fmt.Fprintf(&b, "  %s\n", surface.Tagline)
	} else if meta.Description != "" {
		fmt.Fprintf(&b, "  %s\n", meta.Description)
	}
	b.WriteString("\nCOMMAND GROUPS\n")
	rendered := map[string]bool{}
	for _, group := range surface.Groups {
		for _, prefix := range group.Commands {
			if !connectorSurfaceHasPrefix(surface, prefix) {
				continue
			}
			rendered[prefix] = true
			fmt.Fprintf(&b, "  %s - %s; see pm %s %s --help\n", prefix, valueOr(group.Title, prefix), connectorName, prefix)
		}
	}
	for _, prefix := range connectorSurfacePrefixes(surface) {
		if rendered[prefix] {
			continue
		}
		fmt.Fprintf(&b, "  %s - see pm %s %s --help\n", prefix, connectorName, prefix)
	}
	writeConnectorGlobalFlags(&b, surface)
	if len(surface.HelpTopics) > 0 {
		b.WriteString("\nHELP TOPICS\n")
		for _, topic := range surface.HelpTopics {
			if topic.Name == "" {
				continue
			}
			fmt.Fprintf(&b, "  %s", topic.Name)
			if topic.Summary != "" {
				fmt.Fprintf(&b, " - %s", topic.Summary)
			}
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func renderConnectorCommandGroup(connectorName string, connector connectors.Connector, surface *connectors.CommandSurface, prefix string) string {
	var b strings.Builder
	displayName := valueOr(connector.Metadata().DisplayName, connectorName)
	b.WriteString("NAME\n")
	fmt.Fprintf(&b, "  pm %s %s - %s %s commands\n\n", connectorName, prefix, displayName, prefix)
	b.WriteString("SYNOPSIS\n")
	fmt.Fprintf(&b, "  pm %s %s <command> [flags]\n\n", connectorName, prefix)
	if title := connectorSurfaceGroupTitle(surface, prefix); title != "" {
		b.WriteString("DESCRIPTION\n")
		fmt.Fprintf(&b, "  %s commands for %s.\n\n", title, prefix)
	}
	b.WriteString("COMMANDS\n")
	for _, cmd := range surface.Commands {
		if commandSurfacePrefix(cmd.Path) != prefix {
			continue
		}
		fmt.Fprintf(&b, "  %s", cmd.Path)
		if cmd.Summary != "" {
			fmt.Fprintf(&b, " - %s", cmd.Summary)
		}
		if cmd.Availability != "" {
			fmt.Fprintf(&b, " [availability=%s]", cmd.Availability)
		}
		b.WriteByte('\n')
	}
	writeConnectorGlobalFlags(&b, surface)
	return b.String()
}

func renderConnectorCommandDetail(connectorName string, surface *connectors.CommandSurface, cmd connectors.CommandSurfaceCommand) string {
	var b strings.Builder
	b.WriteString("NAME\n")
	fmt.Fprintf(&b, "  pm %s %s", connectorName, cmd.Path)
	if cmd.Summary != "" {
		fmt.Fprintf(&b, " - %s", cmd.Summary)
	}
	b.WriteString("\n\nSYNOPSIS\n")
	fmt.Fprintf(&b, "  pm %s %s [flags]\n", connectorName, cmd.Path)
	if len(cmd.Examples) > 0 {
		for _, example := range cmd.Examples {
			fmt.Fprintf(&b, "  %s\n", example)
		}
	}
	if cmd.Summary != "" {
		b.WriteString("\nDESCRIPTION\n")
		fmt.Fprintf(&b, "  %s\n", cmd.Summary)
	}
	writeConnectorField(&b, "INTENT", cmd.Intent)
	writeConnectorField(&b, "AVAILABILITY", cmd.Availability)
	writeConnectorField(&b, "STREAM", cmd.Stream)
	writeConnectorField(&b, "WRITE", cmd.Write)
	writeConnectorField(&b, "OPERATION", cmd.Operation)
	writeConnectorField(&b, "APPROVAL", cmd.Approval)
	writeConnectorField(&b, "RISK", cmd.Risk)
	writeConnectorField(&b, "OUTPUT POLICY", cmd.OutputPolicy)
	writeConnectorField(&b, "NOTES", cmd.Notes)
	b.WriteString("\nFLAGS\n")
	if len(cmd.Flags) == 0 {
		b.WriteString("  No command-specific flags.\n")
	} else {
		for _, flag := range cmd.Flags {
			writeConnectorFlag(&b, flag)
		}
	}
	writeConnectorGlobalFlags(&b, surface)
	return b.String()
}

func writeConnectorField(b *strings.Builder, title, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	fmt.Fprintf(b, "\n%s\n  %s\n", title, value)
}

func writeConnectorGlobalFlags(b *strings.Builder, surface *connectors.CommandSurface) {
	if len(surface.GlobalFlags) == 0 {
		return
	}
	b.WriteString("\nGLOBAL FLAGS\n")
	for _, flag := range surface.GlobalFlags {
		writeConnectorFlag(b, flag)
	}
}

func writeConnectorFlag(b *strings.Builder, flag connectors.CommandSurfaceFlag) {
	fmt.Fprintf(b, "  --%s", strings.TrimLeft(flag.Name, "-"))
	if flag.Type != "" {
		fmt.Fprintf(b, " (%s)", flag.Type)
	}
	if flag.Required {
		b.WriteString(" required")
	}
	if flag.Summary != "" {
		fmt.Fprintf(b, ": %s", flag.Summary)
	}
	if len(flag.Values) > 0 {
		fmt.Fprintf(b, " values=%s", strings.Join(flag.Values, "|"))
	}
	if flag.MapsTo != "" {
		fmt.Fprintf(b, " maps_to=%s", flag.MapsTo)
	}
	b.WriteByte('\n')
}

func connectorSurfaceCommand(surface *connectors.CommandSurface, command string) (connectors.CommandSurfaceCommand, bool) {
	if surface == nil {
		return connectors.CommandSurfaceCommand{}, false
	}
	for _, cmd := range surface.Commands {
		if cmd.Path == command {
			return cmd, true
		}
	}
	return connectors.CommandSurfaceCommand{}, false
}

func connectorSurfaceHasPrefix(surface *connectors.CommandSurface, prefix string) bool {
	for _, cmd := range surface.Commands {
		if commandSurfacePrefix(cmd.Path) == prefix {
			return true
		}
	}
	return false
}

func connectorSurfacePrefixes(surface *connectors.CommandSurface) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, cmd := range surface.Commands {
		prefix := commandSurfacePrefix(cmd.Path)
		if prefix == "" || seen[prefix] {
			continue
		}
		seen[prefix] = true
		out = append(out, prefix)
	}
	return out
}

func connectorSurfaceGroupTitle(surface *connectors.CommandSurface, prefix string) string {
	for _, group := range surface.Groups {
		for _, candidate := range group.Commands {
			if candidate == prefix {
				return valueOr(group.Title, group.ID)
			}
		}
	}
	return ""
}

func commandSurfacePrefix(path string) string {
	fields := strings.Fields(path)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func withConnectorCommandSuggestion(blocked *commandrunner.BlockedCommandError, surface *connectors.CommandSurface, path []string) error {
	if blocked == nil || blocked.Reason != "unknown command" {
		return blocked
	}
	if suggestion := connectorCommandSuggestion(surface, path); suggestion != "" {
		copy := *blocked
		copy.Reason = fmt.Sprintf("%s; did you mean %q", copy.Reason, suggestion)
		return &copy
	}
	return blocked
}

func connectorCommandSuggestion(surface *connectors.CommandSurface, path []string) string {
	if len(path) == 0 {
		return ""
	}
	aliases := map[string]string{
		"appoint":     "appointments",
		"appointment": "appointments",
	}
	replacement, ok := aliases[path[0]]
	if !ok {
		return ""
	}
	candidate := append([]string{replacement}, path[1:]...)
	command := strings.Join(candidate, " ")
	if _, ok := connectorSurfaceCommand(surface, command); ok {
		return command
	}
	if connectorSurfaceHasPrefix(surface, replacement) {
		return replacement + " --help"
	}
	return ""
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
		_, _ = fmt.Fprintln(stdout, string(b))
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
		_, _ = fmt.Fprintln(stdout, string(b))
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
	_, _ = fmt.Fprintf(stdout, "Created connector command plan %s for %s\nApproval token: %s\n", plan.ID, plan.ConnectorCommand, plan.ApprovalToken)
	if plan.ConfirmationChallenge != "" {
		_, _ = fmt.Fprintf(stdout, "Confirmation required: --confirm %s\n", plan.ConfirmationChallenge)
	}
	if writePreview != nil {
		for _, warning := range writePreview.Warnings {
			_, _ = fmt.Fprintf(stdout, "- %s\n", warning)
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
		_, _ = fmt.Fprintf(stdout, "Reverse ETL run %s completed: succeeded=%d failed=%d\n", run.ID, run.RecordsSucceeded, run.RecordsFailed)
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
		_, _ = fmt.Fprintf(stdout, "Reverse plan %s previews %s via %s\n", plan.ID, plan.ConnectorCommand, plan.Action)
		for _, warning := range writePreview.Warnings {
			_, _ = fmt.Fprintf(stdout, "- %s\n", warning)
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
	plan.Sample = app.RedactReversePlanRecords(plan.Sample, plan.RedactFields)
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

func directConnector(a *app.App, args []string) (connectors.Connector, connectors.RuntimeConfig, error) {
	flags := parseFlags(args)
	name := flags.first("connector")
	if name == "" {
		return nil, connectors.RuntimeConfig{}, errors.New("missing --connector")
	}
	if err := safety.ValidateIdentifier(name, "connector"); err != nil {
		return nil, connectors.RuntimeConfig{}, validationErrorf("%v", err)
	}
	if err := connectors.RejectLegacyConnectorName(name); err != nil {
		return nil, connectors.RuntimeConfig{}, err
	}
	connector, ok := a.Registry().Get(name)
	if !ok {
		return nil, connectors.RuntimeConfig{}, fmt.Errorf("connector %q not found", name)
	}
	config, err := keyValues(flags.values["config"])
	if err != nil {
		return nil, connectors.RuntimeConfig{}, err
	}
	return connector, connectors.RuntimeConfig{
		ProjectDir: a.ProjectDir(),
		Config:     config,
	}, nil
}

func runQuery(ctx context.Context, a *app.App, args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 || args[0] != "run" {
		return errUsage
	}
	flags := parseFlags(args[1:])
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
		_, _ = fmt.Fprintln(stdout, string(b))
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

func runReverse(ctx context.Context, a *app.App, args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 {
		return errUsage
	}
	switch args[0] {
	case "list":
		plans := a.ListReversePlans()
		runs := a.ListReverseRuns()
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "ReversePlanList", "plans": safeReversePlansForOutput(plans), "runs": runs})
		}
		for _, plan := range plans {
			_, _ = fmt.Fprintf(stdout, "%s\t%s\t%s\trecords=%d\n", plan.ID, plan.Status, plan.Name, plan.RecordCount)
		}
		if len(runs) > 0 {
			_, _ = fmt.Fprintln(stdout, "\nRUNS")
			for _, run := range runs {
				_, _ = fmt.Fprintf(stdout, "%s\t%s\tplan=%s\tsucceeded=%d failed=%d\n", run.ID, run.Status, run.PlanID, run.RecordsSucceeded, run.RecordsFailed)
			}
		}
		return nil
	case "plan":
		if len(args) < 2 {
			return errUsage
		}
		flags := parseFlags(args[2:])
		dest, err := parseEndpoint(flags.first("destination"))
		if err != nil {
			return err
		}
		mappings, err := colonValues(flags.values["map"])
		if err != nil {
			return err
		}
		limit, err := parseIntFlag("limit", flags.first("limit"), 0)
		if err != nil {
			return err
		}
		plan, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
			Name:                  args[1],
			SourceTable:           flags.first("source-table"),
			DestinationConnector:  dest.Connector,
			DestinationCredential: dest.Credential,
			DestinationConfig:     dest.Config,
			Action:                valueOr(flags.first("action"), "upsert"),
			Mappings:              mappings,
			Limit:                 limit,
		})
		if err != nil {
			return err
		}
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "ReversePlan", "plan": safeReversePlanForOutput(plan), "approval_required": true})
		}
		_, _ = fmt.Fprintf(stdout, "Created reverse plan %s with %d records\nApproval token: %s\n", plan.ID, plan.RecordCount, plan.ApprovalToken)
		if plan.ConfirmationChallenge != "" {
			_, _ = fmt.Fprintf(stdout, "Confirmation required: --confirm %s\n", plan.ConfirmationChallenge)
		}
		return nil
	case "preview":
		if len(args) < 2 {
			return errUsage
		}
		plan, err := a.GetReversePlan(args[1])
		if err != nil {
			return err
		}
		if jsonOut {
			env := envelope{"kind": "ReversePlanPreview", "plan": safeReversePlanForOutput(plan)}
			if plan.ConnectorCommand != "" {
				safePlan, writePreview, err := a.PreviewConnectorCommandPlan(ctx, args[1])
				if err != nil {
					return err
				}
				env["plan"] = safeReversePlanForOutput(safePlan)
				env["write_preview"] = writePreview
			}
			return writeJSON(stdout, env)
		}
		b, _ := json.MarshalIndent(safeReversePlanForOutput(plan), "", "  ")
		_, _ = fmt.Fprintln(stdout, string(b))
		return nil
	case "run":
		if len(args) < 2 {
			return errUsage
		}
		flags := parseFlags(args[2:])
		run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: args[1], ApprovalToken: flags.first("approve"), Confirmation: flags.first("confirm")})
		if err != nil {
			return err
		}
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "ReverseRun", "run": run})
		}
		_, _ = fmt.Fprintf(stdout, "Reverse ETL run %s completed: succeeded=%d failed=%d\n", run.ID, run.RecordsSucceeded, run.RecordsFailed)
		return nil
	case "status":
		if len(args) < 2 {
			return errUsage
		}
		run, err := a.GetReverseRun(args[1])
		if err != nil {
			return err
		}
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "ReverseRun", "run": run})
		}
		_, _ = fmt.Fprintf(stdout, "%s\t%s\tplan=%s\tstaged=%d succeeded=%d failed=%d\n", run.ID, run.Status, run.PlanID, run.RecordsStaged, run.RecordsSucceeded, run.RecordsFailed)
		return nil
	default:
		return errUsage
	}
}

func runAgent(ctx context.Context, cfg config.Config, root string, args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 {
		return errUsage
	}
	if args[0] == "image" {
		return runAgentImage(ctx, cfg, root, args[1:], stdout, jsonOut)
	}
	if args[0] != "plan" {
		return errUsage
	}
	flags := parseFlags(args[1:])
	req := strings.ToLower(flags.first("request"))
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
		_, _ = fmt.Fprintln(stdout, step)
	}
	return nil
}

func runDocs(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return errUsage
	}
	flags := parseFlags(args[1:])
	switch args[0] {
	case "generate":
		dir := flags.first("dir")
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
		connectorsDir := valueOr(flags.first("connectors-dir"), filepath.Join(filepath.Dir(dir), "connectors"))
		if err := writeConnectorDocs(connectorsDir, appRegistry()); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(stdout, "Generated docs in %s and connector docs in %s\n", dir, connectorsDir)
		return nil
	case "validate":
		dir := valueOr(flags.first("connectors-dir"), valueOr(flags.first("dir"), "docs/connectors"))
		if err := validateConnectorDocs(dir, appRegistry()); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(stdout, "Validated connector docs in %s\n", dir)
		return nil
	default:
		return errUsage
	}
}

func runRuntime(ctx context.Context, cfg config.Config, args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 || args[0] != "doctor" {
		return errUsage
	}
	runtimeCfg := runtimecheck.FromConfig(cfg)
	report := runtimecheck.Doctor(ctx, runtimeCfg)
	if jsonOut {
		return writeJSON(stdout, envelope{
			"kind":   "RuntimeDoctor",
			"config": runtimecheck.RedactedConfig(runtimeCfg),
			"report": report,
		})
	}
	_, _ = fmt.Fprintf(stdout, "mode=%s duration=%s\n", report.Mode, report.Duration)
	for _, check := range report.Checks {
		if check.Error != "" {
			_, _ = fmt.Fprintf(stdout, "%s\t%s\t%s\t%s\t%s\n", check.Name, check.Status, check.Endpoint, check.Latency, check.Error)
			continue
		}
		_, _ = fmt.Fprintf(stdout, "%s\t%s\t%s\t%s\n", check.Name, check.Status, check.Endpoint, check.Latency)
	}
	return nil
}

func runPerf(ctx context.Context, cfg config.Config, args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 {
		return errUsage
	}
	switch args[0] {
	case "compare":
		flags := parseFlags(args[1:])
		iterations, err := parseIntFlag("iterations", valueOr(flags.first("iterations"), "25"), 25)
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
		_, _ = fmt.Fprintf(stdout, "\nDependency-free: %s\n", comparison.Explanation["dependency_free"])
		_, _ = fmt.Fprintf(stdout, "Runtime-backed: %s\n", comparison.Explanation["runtime_backed"])
		return nil
	case "sync-modes":
		flags := parseFlags(args[1:])
		records, err := parseIntFlag("records", valueOr(flags.first("records"), "1000"), 1000)
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
			_, _ = fmt.Fprintf(stdout, "%s\trecords=%d\tduration=%s\trecords_per_sec=%.2f", result.Mode, result.Records, result.Duration, result.RecordsPerSec)
			if result.Error != "" {
				_, _ = fmt.Fprintf(stdout, "\terror=%s", result.Error)
			}
			_, _ = fmt.Fprintln(stdout)
		}
		_, _ = fmt.Fprintln(stdout, benchmark.Explanation)
		return nil
	default:
		return errUsage
	}
}

func printPerfResult(stdout io.Writer, result perf.Result) {
	if result.Error != "" {
		_, _ = fmt.Fprintf(stdout, "%s\titerations=%d\terror=%s\n", result.Mode, result.Iterations, result.Error)
		return
	}
	_, _ = fmt.Fprintf(stdout, "%s\titerations=%d\trecords=%d\tduration=%s\tavg=%s\trecords_per_sec=%.2f\n",
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
