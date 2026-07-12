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
	if len(cleanArgs) == 0 {
		if err := writeRootManual(stdout, jsonOut); err != nil {
			return writeError(stdout, stderr, err, jsonOut)
		}
		return 0
	}
	cmd := cleanArgs[0]
	rest := cleanArgs[1:]
	if cmd == "--help" || cmd == "-h" || (cmd == "help" && len(rest) == 0) || cmd == "man" && len(rest) == 0 {
		if err := writeRootManual(stdout, jsonOut); err != nil {
			return writeError(stdout, stderr, err, jsonOut)
		}
		return 0
	}
	if len(rest) > 0 && (rest[0] == "--help" || rest[0] == "-h" || rest[0] == "help") {
		if err := writeCommandOrConnectorManual(cmd, stdout, jsonOut); err != nil {
			return writeError(stdout, stderr, err, jsonOut)
		}
		return 0
	}
	if len(rest) == 0 && isManualCommand(cmd) {
		if err := writeCommandOrConnectorManual(cmd, stdout, jsonOut); err != nil {
			return writeError(stdout, stderr, err, jsonOut)
		}
		return 0
	}
	var err error
	switch cmd {
	case "init":
		err = runInit(root, stdout, jsonOut)
	case "help", "man":
		err = runHelp(rest, stdout, jsonOut)
	case "connectors":
		err = runConnectors(ctx, root, rest, stdout, jsonOut)
	case "credentials":
		err = withApp(root, func(a *app.App) error { return runCredentials(ctx, a, rest, stdout, jsonOut) })
	case "connections":
		err = withApp(root, func(a *app.App) error { return runConnections(ctx, a, rest, stdout, jsonOut) })
	case "catalog":
		err = withApp(root, func(a *app.App) error { return runCatalog(ctx, a, rest, stdout, jsonOut) })
	case "etl":
		err = withApp(root, func(a *app.App) error { return runETL(ctx, a, rest, stdout, jsonOut) })
	case "query":
		err = withApp(root, func(a *app.App) error { return runQuery(ctx, a, rest, stdout, jsonOut) })
	case "reverse":
		err = withApp(root, func(a *app.App) error { return runReverse(ctx, a, rest, stdout, jsonOut) })
	case "agent":
		err = runAgent(ctx, root, rest, stdout, jsonOut)
	case "runtime":
		err = runRuntime(ctx, rest, stdout, jsonOut)
	case "flow":
		err = withApp(root, func(a *app.App) error { return runFlow(ctx, a, rest, stdout, jsonOut) })
	case "extract":
		err = withApp(root, func(a *app.App) error { return runExtract(ctx, a, root, rest, stdout, jsonOut) })
	case "perf":
		err = runPerf(ctx, rest, stdout, jsonOut)
	case "docs":
		err = runDocs(rest, stdout)
	case "skills":
		err = runSkills(rest, stdout, jsonOut)
	case "version":
		err = runVersion(rest, stdout, jsonOut)
	case "rlm":
		err = runRLM(ctx, root, rest, stdout, jsonOut)
	case "schedule":
		err = runSchedule(ctx, root, rest, stdout, jsonOut)
	case "worker":
		err = runWorker(ctx, rest, stdout, jsonOut)
	default:
		err = runMaybeConnectorCommand(ctx, root, cmd, rest, stdout, jsonOut)
	}
	if err != nil {
		return writeError(stdout, stderr, err, jsonOut)
	}
	return 0
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
	if topic == "" {
		return writeRootManual(stdout, jsonOut)
	}
	return writeCommandOrConnectorManual(topic, stdout, jsonOut)
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
		return fmt.Errorf("help topic %q not found", topic)
	}
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "CommandManual", "command": topic, "manual": text})
	}
	fmt.Fprint(stdout, text)
	return nil
}

func writeCommandOrConnectorManual(topic string, stdout io.Writer, jsonOut bool) error {
	if _, ok := docs[topic]; ok {
		return writeManual(topic, stdout, jsonOut)
	}
	return writeConnectorCommandManual(topic, stdout, jsonOut)
}

func writeConnectorCommandManual(name string, stdout io.Writer, jsonOut bool) error {
	if err := safety.ValidateIdentifier(name, "connector"); err != nil {
		return fmt.Errorf("help topic %q not found", name)
	}
	if err := connectors.RejectLegacyConnectorName(name); err != nil {
		return err
	}
	connector, ok := appRegistry().Get(name)
	if !ok {
		return fmt.Errorf("help topic %q not found", name)
	}
	provider, ok := connector.(connectors.CommandSurfaceProvider)
	if !ok || provider.CommandSurface() == nil {
		return fmt.Errorf("help topic %q not found", name)
	}
	manual := connectors.RenderConnectorManual(connector)
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "CommandManual", "command": name, "manual": manual})
	}
	fmt.Fprint(stdout, manual)
	return nil
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
				fmt.Fprintf(stdout, "%s\t%s\tread=%t\twrite=%t\tquery=%t\n", item.Name, item.IntegrationType, item.Capabilities.Read, item.Capabilities.Write, item.Capabilities.Query)
			}
			return nil
		}
		list := registry.List()
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "ConnectorList", "connectors": list})
		}
		for _, item := range list {
			fmt.Fprintf(stdout, "%s\t%s\t%+v\n", item.Name, item.IntegrationType, item.Capabilities)
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
			fmt.Fprintf(stdout, "%s\t%s\tread=%t\twrite=%t\tquery=%t\n", item.Name, item.IntegrationType, item.Capabilities.Read, item.Capabilities.Write, item.Capabilities.Query)
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
			fmt.Fprint(stdout, connectors.RenderConnectorManual(c))
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
		fmt.Fprintf(stdout, "Saved credential %s for connector %s\n", cred.Name, cred.Connector)
		return nil
	case "list":
		creds := a.ListCredentials()
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "CredentialList", "credentials": creds})
		}
		for _, cred := range creds {
			fmt.Fprintf(stdout, "%s\t%s\t%s\n", cred.Name, cred.ID, cred.Connector)
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
		fmt.Fprintln(stdout, string(b))
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
		fmt.Fprintf(stdout, "Credential %s validated\n", cred.Name)
		return nil
	case "remove":
		if len(args) < 2 {
			return errUsage
		}
		if err := a.RemoveCredential(ctx, args[1]); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Removed credential %s\n", args[1])
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
		fmt.Fprintf(stdout, "Created connection %s\n", conn.Name)
		return nil
	case "list":
		conns := a.ListConnections()
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "ConnectionList", "connections": conns})
		}
		for _, conn := range conns {
			fmt.Fprintf(stdout, "%s\t%s:%s -> %s:%s\n", conn.Name, conn.Source.Connector, conn.Source.Credential, conn.Destination.Connector, conn.Destination.Credential)
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
		fmt.Fprintf(stdout, "%s\t%s\n", stream.Name, stream.Description)
	}
	return nil
}

func runETL(ctx context.Context, a *app.App, args []string, stdout io.Writer, jsonOut bool) error {
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
		fmt.Fprintf(stdout, "Connector %s check ok\n", connector.Name())
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
			fmt.Fprintf(stdout, "%s\t%s\n", stream.Name, stream.Description)
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
			fmt.Fprintln(stdout, string(b))
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
			if err := recordRuntimeETL(ctx, run); err != nil {
				return err
			}
			runtimeRecorded = true
		}
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "ETLRun", "run": run, "runtime_recorded": runtimeRecorded})
		}
		if runtimeRecorded {
			fmt.Fprintf(stdout, "ETL run %s completed: read=%d loaded=%d failed=%d runtime_recorded=true\n", run.ID, run.RecordsRead, run.RecordsLoaded, run.RecordsFailed)
			return nil
		}
		fmt.Fprintf(stdout, "ETL run %s completed: read=%d loaded=%d failed=%d\n", run.ID, run.RecordsRead, run.RecordsLoaded, run.RecordsFailed)
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
		fmt.Fprintf(stdout, "%s\t%s\tread=%d loaded=%d failed=%d\n", run.ID, run.Status, run.RecordsRead, run.RecordsLoaded, run.RecordsFailed)
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
	flags := parseFlags(args)
	path := flags.values["_"]
	if len(path) == 0 {
		return writeConnectorCommandManual(connectorName, stdout, jsonOut)
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
	maxBytes, err := parseIntFlag("max-bytes", flags.first("max-bytes"), 1<<20)
	if err != nil {
		return err
	}
	if maxBytes <= 0 {
		maxBytes = 1 << 20
	}
	if maxBytes > commandrunner.MaxDirectReadBytes {
		maxBytes = commandrunner.MaxDirectReadBytes
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
			fmt.Fprintf(stdout, "%s\t%s\t%s\trecords=%d\n", plan.ID, plan.Status, plan.Name, plan.RecordCount)
		}
		if len(runs) > 0 {
			fmt.Fprintln(stdout, "\nRUNS")
			for _, run := range runs {
				fmt.Fprintf(stdout, "%s\t%s\tplan=%s\tsucceeded=%d failed=%d\n", run.ID, run.Status, run.PlanID, run.RecordsSucceeded, run.RecordsFailed)
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
		fmt.Fprintf(stdout, "Created reverse plan %s with %d records\nApproval token: %s\n", plan.ID, plan.RecordCount, plan.ApprovalToken)
		if plan.ConfirmationChallenge != "" {
			fmt.Fprintf(stdout, "Confirmation required: --confirm %s\n", plan.ConfirmationChallenge)
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
		fmt.Fprintln(stdout, string(b))
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
		fmt.Fprintf(stdout, "Reverse ETL run %s completed: succeeded=%d failed=%d\n", run.ID, run.RecordsSucceeded, run.RecordsFailed)
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
		fmt.Fprintf(stdout, "%s\t%s\tplan=%s\tstaged=%d succeeded=%d failed=%d\n", run.ID, run.Status, run.PlanID, run.RecordsStaged, run.RecordsSucceeded, run.RecordsFailed)
		return nil
	default:
		return errUsage
	}
}

func runAgent(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 {
		return errUsage
	}
	if args[0] == "image" {
		return runAgentImage(ctx, root, args[1:], stdout, jsonOut)
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
		fmt.Fprintln(stdout, step)
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
		fmt.Fprintf(stdout, "Generated docs in %s and connector docs in %s\n", dir, connectorsDir)
		return nil
	case "validate":
		dir := valueOr(flags.first("connectors-dir"), valueOr(flags.first("dir"), "docs/connectors"))
		if err := validateConnectorDocs(dir, appRegistry()); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Validated connector docs in %s\n", dir)
		return nil
	default:
		return errUsage
	}
}

func runRuntime(ctx context.Context, args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 || args[0] != "doctor" {
		return errUsage
	}
	cfg := runtimecheck.FromEnv()
	report := runtimecheck.Doctor(ctx, cfg)
	if jsonOut {
		return writeJSON(stdout, envelope{
			"kind":   "RuntimeDoctor",
			"config": runtimecheck.RedactedConfig(cfg),
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

func runPerf(ctx context.Context, args []string, stdout io.Writer, jsonOut bool) error {
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
			Iterations: iterations,
			Runtime:    flags.first("runtime") == "true",
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
			fmt.Fprintf(stdout, "%s\trecords=%d\tduration=%s\trecords_per_sec=%.2f", result.Mode, result.Records, result.Duration, result.RecordsPerSec)
			if result.Error != "" {
				fmt.Fprintf(stdout, "\terror=%s", result.Error)
			}
			fmt.Fprintln(stdout)
		}
		fmt.Fprintln(stdout, benchmark.Explanation)
		return nil
	default:
		return errUsage
	}
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
