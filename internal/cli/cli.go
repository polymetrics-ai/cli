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
	"polymetrics.ai/internal/connectors/certifications"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/coordination"
	"polymetrics.ai/internal/credential"
	"polymetrics.ai/internal/perf"
	"polymetrics.ai/internal/runtimecheck"
	"polymetrics.ai/internal/safety"
	"polymetrics.ai/internal/warehouse"
)

type envelope map[string]any

const maxConnectorCommandLimit = 10000

func Run(args []string, stdout, stderr io.Writer) int {
	ctx := context.Background()
	root, jsonOut, cleanArgs := parseGlobal(args)
	opts := config.Options{Root: root, Flags: globalConfigFlags(args, root, jsonOut)}
	if err := validateRawApprovalCarrierArgs(args); err != nil {
		jsonError := jsonOut
		bootstrap, bootstrapErr := config.ResolveBootstrap(opts)
		if bootstrapErr == nil {
			jsonError = bootstrap.JSON
			if !rawApprovalCarrierRoot(bootstrap.Root) {
				if cfg, loadErr := config.Load(opts); loadErr == nil {
					jsonError = cfg.JSON
				}
			}
		}
		return writeError(stdout, stderr, err, jsonError)
	}
	bootstrap, err := config.ResolveBootstrap(opts)
	if err != nil {
		return writeError(stdout, stderr, validationErrorf("%v", err), bootstrap.JSON)
	}
	cfg, err := config.Load(opts)
	if err != nil {
		return writeError(stdout, stderr, validationErrorf("%v", err), bootstrap.JSON)
	}
	if err := validateApprovalCarrierBeforeDispatch(cleanArgs); err != nil {
		return writeError(stdout, stderr, err, cfg.JSON)
	}
	engine.ConfigureSharedRateLimitRegistry(coordination.OpenSharedRateLimitRegistry(cfg.Runtime.DragonflyAddr))
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
	if len(args) >= 2 && args[0] == "etl" && args[1] == "transport" {
		return writeETLTransportManual(stdout, jsonOut, "etl transport")
	}
	topic := ""
	if len(args) > 0 {
		topic = args[0]
	}
	return writeManual(topic, stdout, jsonOut)
}

func isManualCommand(cmd string) bool {
	if cmd == "init" || cmd == "help" || cmd == "man" || cmd == "version" || cmd == "extract" {
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

func runConnectors(ctx context.Context, root string, args []string, stdout, stderr io.Writer, jsonOut bool) error {
	registry := appRegistry()
	if len(args) == 0 {
		return errUsage
	}
	switch args[0] {
	case "certify":
		if len(args) == 2 && isHelpArg(args[1]) {
			return writeManual("connectors", stdout, jsonOut)
		}
		return runCertify(ctx, root, args[1:], stdout, stderr, jsonOut)
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
			status, err := certifications.StatusForRegistered(args[1], true)
			if err != nil {
				return fmt.Errorf("read connector certification status: %w", err)
			}
			if jsonOut {
				response := envelope{
					"kind":           "Connector",
					"connector":      connectors.MetadataOf(c),
					"manifest":       connectors.ManifestOf(c),
					"certification":  status,
					"sync_transport": connectors.SyncTransportEligibilityOf(c),
				}
				if rateLimits, ok := connectors.RateLimitCoordinationOf(c); ok {
					response["rate_limit_coordination"] = rateLimits
				}
				if executionLimits := connectors.RequestExecutionLimitsOf(c); len(executionLimits) > 0 {
					response["request_execution_limits"] = executionLimits
				}
				if def, ok := connectors.DefinitionOf(c); ok && def.Changefeed != nil {
					response["changefeed"] = def.Changefeed
				}
				if def, ok := connectors.DefinitionOf(c); ok && def.PollingWatermark != nil {
					response["polling_watermark"] = def.PollingWatermark
				}
				return writeJSON(stdout, response)
			}
			_, _ = fmt.Fprint(stdout, connectors.RenderConnectorManual(c))
			if rateLimits, ok := connectors.RateLimitCoordinationOf(c); ok {
				_, _ = fmt.Fprintf(stdout, "\nRATE LIMIT COORDINATION\n  %s\n", rateLimits.Message)
			}
			_, _ = fmt.Fprintf(stdout, "\nCERTIFICATION\n  %s\n", status.Label)
			if status.Warning != "" {
				_, _ = fmt.Fprintf(stdout, "  %s\n", status.Warning)
			}
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
		if !definitionHasCapability(def, capability) {
			continue
		}
		out = append(out, def)
	}
	return out, nil
}

func definitionHasCapability(def connectors.Definition, capability string) bool {
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
		return def.Capabilities.CDC
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
		if flags.isBare("provider-family") {
			return usageErrorf("missing value for --provider-family")
		}
		if flags.isBare("auth-profile") {
			return usageErrorf("missing value for --auth-profile")
		}
		if flags.isBare("link-credential") || flags.hasBlankValue("link-credential") {
			return usageErrorf("--link-credential requires a credential identifier")
		}
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
			if err := credential.RequirePersistentValue(key, secrets[key]); err != nil {
				return validationErrorf("%v", err)
			}
		}
		if field := flags.first("value-stdin"); field != "" {
			b, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("read stdin secret: %w", err)
			}
			secrets[field] = credential.NormalizeStdin(string(b))
			if err := credential.RequirePersistentValue(field, secrets[field]); err != nil {
				return validationErrorf("%v", err)
			}
		}
		config, err := keyValues(flags.values["config"])
		if err != nil {
			return err
		}
		if err := validateCredentialConfig(a, connector, config); err != nil {
			return err
		}
		cred, err := a.AddCredential(ctx, app.AddCredentialRequest{
			Name:           args[1],
			Connector:      connector,
			Config:         config,
			Secrets:        secrets,
			ProviderFamily: flags.first("provider-family"),
			AuthProfile:    flags.first("auth-profile"),
			LinkCredential: flags.first("link-credential"),
		})
		if err != nil {
			return credentialCoordinationInputError(err)
		}
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "Credential", "credential": cred})
		}
		_, _ = fmt.Fprintf(stdout, "Saved credential %s for connector %s\n", cred.Name, cred.Connector)
		return nil
	case "link":
		if len(args) < 2 {
			return errUsage
		}
		if err := safety.ValidateIdentifier(args[1], "credential"); err != nil {
			return validationErrorf("%v", err)
		}
		flags := parseFlags(args[2:])
		if flags.isBare("to") {
			return usageErrorf("--to requires a credential identifier")
		}
		target := flags.first("to")
		if target == "" {
			return usageErrorf("missing --to")
		}
		if err := safety.ValidateIdentifier(target, "credential"); err != nil {
			return validationErrorf("%v", err)
		}
		cred, err := a.LinkCredential(args[1], target)
		if err != nil {
			return credentialCoordinationInputError(err)
		}
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "Credential", "credential": cred})
		}
		_, _ = fmt.Fprintf(stdout, "Linked credential %s to compatible credential %s\n", cred.Name, target)
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

func credentialCoordinationInputError(err error) error {
	var emptySecret *credential.EmptySecretError
	if errors.As(err, &emptySecret) {
		return validationErrorf("%v", err)
	}
	var declarationErr *app.CredentialCoordinationDeclarationError
	if errors.As(err, &declarationErr) {
		return validationErrorf("%v", err)
	}
	var linkErr *app.CredentialLinkValidationError
	if errors.As(err, &linkErr) {
		return validationErrorf("%v", err)
	}
	return err
}

func runConnections(ctx context.Context, a *app.App, args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 {
		return errUsage
	}
	switch args[0] {
	case "create":
		if containsHelpFlag(args[1:]) {
			return writeManual("connections", stdout, jsonOut)
		}
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
		targetCopyWorkers, err := parseTargetCopyWorkers(flags.first("target-copy-workers"))
		if err != nil {
			return err
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
		if len(flags.values["destination-action"]) > 1 || flags.isBare("destination-action") {
			return validationErrorf("--destination-action must name exactly one declared action")
		}
		destinationAction := flags.first("destination-action")
		if destinationAction != "" {
			if err := safety.ValidateIdentifier(destinationAction, "destination action"); err != nil {
				return validationErrorf("%v", err)
			}
		}
		streamCfg := app.StreamConfig{
			SyncMode:          valueOr(flags.first("sync-mode"), "full_refresh_overwrite"),
			CursorField:       flags.first("cursor"),
			PrimaryKey:        flags.values["primary-key"],
			DestinationTable:  valueOr(flags.first("table"), stream),
			DestinationAction: destinationAction,
		}
		if transformFile := flags.first("transform-file"); transformFile != "" {
			plan, err := readTransformPlanFile(transformFile)
			if err != nil {
				return err
			}
			streamCfg.TransformPlan = string(plan.NormalizedJSON())
			streamCfg.TransformPlanHash = plan.Hash()
		}
		conn, err := a.CreateConnection(ctx, app.CreateConnectionRequest{
			Name:              args[1],
			Source:            source,
			Destination:       dest,
			Streams:           map[string]app.StreamConfig{stream: streamCfg},
			TargetCopyWorkers: targetCopyWorkers,
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
	if message := catalogStatusMessage(snapshot.Catalog.Discovery); message != "" {
		_, _ = fmt.Fprintln(stdout, message)
	}
	for _, stream := range snapshot.Catalog.Streams {
		_, _ = fmt.Fprintf(stdout, "%s\t%s\n", stream.Name, stream.Description)
	}
	return nil
}

func catalogStatusMessage(status *connectors.DiscoveryStatus) string {
	if status == nil {
		return ""
	}
	switch {
	case status.Stale:
		return "catalog status: stale; run pm catalog refresh --connection <name> before using this schema"
	case !status.Complete:
		return "catalog status: partial; refresh after the provider issue is resolved before relying on this schema"
	default:
		return "catalog status: current"
	}
}

// shouldPresentETLTerminalRun accepts only the App's durable terminal
// presentation proof. Successful state updates return the stored terminal run;
// may-have-committed updates return an exact reload; and a definite no-commit
// returns Run{}. The accompanying operational error controls the categorized
// nonzero exit, never whether a durable terminal envelope is emitted.
func shouldPresentETLTerminalRun(run app.Run) bool {
	if run.ID == "" || !app.IsTerminalETLRunStatus(run.Status) || run.CompletedAt.IsZero() {
		return false
	}
	return true
}

// pmcert:workflow etl
func runETL(ctx context.Context, a *app.App, args []string, stdout io.Writer, jsonOut bool, cfg config.Config) error {
	if len(args) == 0 {
		return errUsage
	}
	switch args[0] {
	case "check":
		connector, cfg, err := directConnector(ctx, a, args[1:])
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
		connector, cfg, err := directConnector(ctx, a, args[1:])
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
		connector, cfg, err := directConnector(ctx, a, args[1:])
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
		if containsHelpFlag(args[1:]) {
			return writeManual("etl", stdout, jsonOut)
		}
		if approval, transportApproval, strictFlags, err := parseETLRunTransportApproval(args[1:], os.Stdin); err != nil {
			return err
		} else if transportApproval {
			return runApprovedTransportETL(ctx, a, strictFlags, approval, stdout, jsonOut)
		}
		if err := validateLegacyETLRunFlags(args[1:]); err != nil {
			return err
		}
		flags := parseFlags(args[1:])
		batchSize, err := parseIntFlag("batch-size", flags.first("batch-size"), 0)
		if err != nil {
			return err
		}
		maxInFlightBatches, err := parseMaxInFlightBatches(flags.first("max-in-flight-batches"))
		if err != nil {
			return err
		}
		run, runErr := a.RunETL(ctx, app.RunETLRequest{
			Connection:         flags.first("connection"),
			Stream:             flags.first("stream"),
			BatchSize:          batchSize,
			MaxInFlightBatches: maxInFlightBatches,
		})
		if runErr != nil && !shouldPresentETLTerminalRun(run) {
			return runErr
		}
		runtimeRecorded := false
		if runErr == nil && flags.first("runtime") == "true" {
			if err := recordRuntimeETL(ctx, run, cfg); err != nil {
				runErr = err
			} else {
				runtimeRecorded = true
			}
		}
		if jsonOut {
			if err := writeJSON(stdout, envelope{"kind": "ETLRun", "run": run, "runtime_recorded": runtimeRecorded}); err != nil {
				return err
			}
			if runErr != nil {
				return alreadyReportedExecutionError(runErr)
			}
			return nil
		}
		if runErr != nil {
			_, _ = fmt.Fprintf(stdout, "ETL run %s ended with status %s; inspect it with pm etl status %s\n", run.ID, run.Status, run.ID)
			return alreadyReportedExecutionError(runErr)
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
	case "transport":
		return runETLTransport(ctx, a, args[1:], stdout, jsonOut)
	default:
		return errUsage
	}
}

func runMaybeConnectorCommand(ctx context.Context, root, connectorName string, args []string, stdout, stderr io.Writer, jsonOut bool) error {
	return runMaybeConnectorCommandWithRegistry(ctx, root, connectorName, args, stdout, stderr, jsonOut, appRegistry())
}

func runMaybeConnectorCommandWithRegistry(ctx context.Context, root, connectorName string, args []string, stdout, stderr io.Writer, jsonOut bool, registry *connectors.Registry) error {
	if err := safety.ValidateIdentifier(connectorName, "connector"); err != nil {
		return usageErrorf("unknown command %q", connectorName)
	}
	if err := connectors.RejectLegacyConnectorName(connectorName); err != nil {
		return err
	}
	connector, ok := registry.Get(connectorName)
	if !ok {
		return usageErrorf("unknown command %q", connectorName)
	}
	surfaceProvider, ok := connector.(connectors.CommandSurfaceProvider)
	if !ok || surfaceProvider.CommandSurface() == nil {
		return usageErrorf("unknown command %q", connectorName)
	}
	surface := surfaceProvider.CommandSurface()
	if len(args) == 0 {
		command, manual := renderConnectorCommandManual(connectorName, connector, surface, args)
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "CommandManual", "command": command, "manual": manual})
		}
		_, _ = fmt.Fprint(stdout, manual)
		return nil
	}
	flags := parseFlags(args)
	if _, err := validateApprovalTokenCarrierFlags(flags); err != nil {
		return err
	}
	if connectorHelpRequested(args, surface) {
		helpPath := connectorHelpPath(flags.values["_"])
		if len(helpPath) > 0 {
			command := strings.Join(helpPath, " ")
			_, isCommand := connectorSurfaceCommand(surface, command)
			isGroup := len(helpPath) == 1 && connectorSurfaceHasPrefix(surface, helpPath[0])
			if !isCommand && !isGroup {
				return connectorCommandUsageError(surface, helpPath)
			}
		}
		command, manual := renderConnectorCommandManual(connectorName, connector, surface, args)
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "CommandManual", "command": command, "manual": manual})
		}
		_, _ = fmt.Fprint(stdout, manual)
		return nil
	}
	path := flags.values["_"]
	if len(path) == 0 {
		return usageErrorf("missing connector command path")
	}
	if err := validateConnectorLifecycleFlagValues(flags); err != nil {
		return err
	}
	config, err := keyValues(flags.values["config"])
	if err != nil {
		return err
	}
	if _, found := connectorSurfaceCommand(surface, strings.Join(path, " ")); !found {
		return connectorCommandUsageError(surface, path)
	}
	var preparedCommandFlags map[string][]string
	preflight := func() error {
		resolvedFlags, resolveErr := resolveConnectorCommandEnvironmentOnlyFlags(surface, path, flags.values)
		if resolveErr != nil {
			return resolveErr
		}
		preparedCommandFlags = connectorCommandFlags(resolvedFlags)
		return commandrunner.PreflightRequest(connector, commandrunner.Request{
			Path: path, Flags: preparedCommandFlags, Config: connectors.RuntimeConfig{Config: config},
			PlanContinuation: flags.first("plan") != "",
		})
	}
	if err := preflight(); err != nil {
		var blocked *commandrunner.BlockedCommandError
		if errors.As(err, &blocked) {
			if blocked.Reason == "unknown command" {
				return connectorCommandUsageError(surface, path)
			}
			return connectorCommandBlockedError(blocked)
		}
		var missingRequired *commandrunner.MissingRequiredFlagError
		if errors.As(err, &missingRequired) {
			return err
		}
		return validationErrorf("%v", err)
	}
	approval, err := prepareReverseApprovalCarrier(flags, os.Stdin)
	if err != nil {
		return err
	}
	if approval.supplied {
		return withReverseExecutionApp(root, func(a *app.App) error {
			return runConnectorCommand(ctx, a, connectorName, args, preparedCommandFlags, approval, stdout, stderr, jsonOut)
		})
	}
	return withApp(root, func(a *app.App) error {
		return runConnectorCommand(ctx, a, connectorName, args, preparedCommandFlags, approval, stdout, stderr, jsonOut)
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
		"page": true, "page-cursor": true,
		"dest-root": true, "file-name": true,
	}
	for _, flag := range surface.GlobalFlags {
		declared[flag.Name] = true
	}
	for name := range flags.values {
		if name != "_" && !declared[name] {
			return false
		}
	}
	for _, name := range []string{"plan", "approve", "approval-token-stdin", "confirm"} {
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
			return connectorName + " " + command, renderConnectorCommandDetail(connectorName, connector, surface, cmd)
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

func renderConnectorCommandDetail(connectorName string, connector connectors.Connector, surface *connectors.CommandSurface, cmd connectors.CommandSurfaceCommand) string {
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
	writeConnectorField(&b, "CONFIRMATION", connectorCommandConfirmationHelp(connector, cmd))
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
	writeConnectorDownloadFlags(&b, cmd)
	writeConnectorPageFlags(&b, cmd)
	writeConnectorGlobalFlags(&b, surface)
	return b.String()
}

// writeConnectorPageFlags documents the page-navigation flags that only a
// direct_read command accepts. A direct read returns one page, so a caller
// told that more records remain needs these documented to fetch them.
//
// Like the download flags, they come from connectors.DirectReadPageFlags
// rather than from literals here, so runtime help and the generated
// manual/skill/website docs cannot document different flags.
func writeConnectorPageFlags(b *strings.Builder, cmd connectors.CommandSurfaceCommand) {
	if cmd.Intent != "direct_read" {
		return
	}
	b.WriteString("\nPAGE FLAGS\n")
	for _, flag := range connectors.DirectReadPageFlags() {
		writeConnectorFlag(b, flag)
	}
}

// connectorCommandConfirmationHelp states the typed confirmation the command
// will demand at run, phrased as the flag the operator has to type.
//
// It resolves through commandrunner rather than reading the bundle's notes, for
// the same reason writeConnectorDownloadFlags reads BinaryDownloadFlags: help
// and the runtime must not be able to disagree. A note is prose one author
// wrote on one command, so it is absent on every command nobody annotated, and
// an absent note reads exactly like "no confirmation needed".
func connectorCommandConfirmationHelp(connector connectors.Connector, cmd connectors.CommandSurfaceCommand) string {
	challenge := strings.TrimSpace(commandrunner.ConfirmationChallengeForCommand(connector, cmd))
	if challenge == "" {
		return ""
	}
	return "execution requires the typed confirmation --confirm " + challenge
}

// writeConnectorDownloadFlags documents the destination flags that only a
// binary_download command accepts. --dest-root is required: the destination is
// never inferred, so a user who does not see it documented cannot run the
// command at all.
//
// The flags come from connectors.BinaryDownloadFlags rather than from literals
// here, so runtime help and the generated manual/skill/website docs cannot
// document different flags.
func writeConnectorDownloadFlags(b *strings.Builder, cmd connectors.CommandSurfaceCommand) {
	if cmd.Intent != "binary_download" && cmd.Intent != "text_export" {
		return
	}
	b.WriteString("\nDOWNLOAD FLAGS\n")
	for _, flag := range connectors.BinaryDownloadFlags() {
		writeConnectorFlag(b, flag)
	}
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
		flagType := flag.Type
		if flag.Type == "json" && flag.AllowBareString {
			flagType = "json or string"
		}
		fmt.Fprintf(b, " (%s)", flagType)
	}
	if flag.Required {
		b.WriteString(" required")
	}
	if flag.Repeatable {
		b.WriteString(" repeatable")
	}
	if flag.EnvOnly {
		fmt.Fprintf(b, " env-only (use --from-env %s=ENV)", strings.TrimLeft(flag.Name, "-"))
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

func connectorCommandUsageError(surface *connectors.CommandSurface, path []string) error {
	message := fmt.Sprintf("unknown command %q", strings.Join(path, " "))
	if suggestion := connectorCommandSuggestion(surface, path); suggestion != "" {
		message = fmt.Sprintf("%s; did you mean %q", message, suggestion)
	}
	return usageErrorf("%s", message)
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

func runConnectorCommand(ctx context.Context, a *app.App, connectorName string, args []string, preparedCommandFlags map[string][]string, approval reverseApprovalCarrier, stdout, stderr io.Writer, jsonOut bool) error {
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
	page, pageCursor, err := connectorCommandPage(flags)
	if err != nil {
		return err
	}
	if flags.first("plan") != "" {
		return runConnectorWriteCommandFromPlan(ctx, a, connectorName, path, flags, approval, stdout, jsonOut)
	}

	connector, cfg, err := a.ResolveConnectorCredential(ctx, connectorName, credential, config)
	if err != nil {
		return err
	}
	// The page flags stay in commandFlags on purpose: only a direct_read can
	// honour them, and the runner drops them for that intent alone. Stripping
	// them here for every intent made `--page 3` on an ETL command
	// accepted-and-ignored, which is the same quiet wrongness --page exists to
	// remove.
	commandFlags := preparedCommandFlags
	if commandFlags == nil {
		return fmt.Errorf("connector command inputs were not preflighted")
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
		Path:       path,
		Flags:      commandFlags,
		Config:     cfg,
		Limit:      limit,
		MaxBytes:   maxBytes,
		Page:       page,
		PageCursor: pageCursor,
		DestRoot:   flags.first("dest-root"),
		FileName:   flags.first("file-name"),
	}, func(record connectors.Record) error {
		rows = append(rows, record)
		return nil
	})
	if err != nil {
		var blocked *commandrunner.BlockedCommandError
		if errors.As(err, &blocked) {
			return connectorCommandBlockedError(err)
		}
		// An executor may have a post-provider result before it can construct a
		// receipt (for example, a legacy parser failure). Emit that one bounded
		// result envelope before the categorized nonzero error instead of
		// erasing provider status/path evidence at the CLI boundary.
		if connectorCommandHasPostProviderResult(result) {
			if outputErr := writeConnectorCommandFailureResult(stdout, stderr, jsonOut, result, rows, err); outputErr != nil {
				return outputErr
			}
			return alreadyReportedExecutionError(err)
		}
		return err
	}
	return writeConnectorCommandResult(stdout, stderr, jsonOut, result, rows)
}

func connectorCommandHasPostProviderResult(result commandrunner.Result) bool {
	return result.DirectRead != nil || result.BinaryDownload != nil || result.StatusCheck != nil
}

func writeConnectorCommandResult(stdout, stderr io.Writer, jsonOut bool, result commandrunner.Result, rows []connectors.Record) error {
	return writeConnectorCommandResultEnvelope(stdout, stderr, jsonOut, result, rows, nil)
}

func writeConnectorCommandFailureResult(stdout, stderr io.Writer, jsonOut bool, result commandrunner.Result, rows []connectors.Record, executionErr error) error {
	return writeConnectorCommandResultEnvelope(stdout, stderr, jsonOut, result, rows, executionErr)
}

func writeConnectorCommandResultEnvelope(stdout, stderr io.Writer, jsonOut bool, result commandrunner.Result, rows []connectors.Record, executionErr error) error {
	if result.BinaryDownload != nil {
		out := envelope{
			"kind":      "ConnectorCommandBinaryDownload",
			"connector": result.Connector,
			"command":   result.Command,
			"operation": result.BinaryDownload.Operation,
			"method":    result.BinaryDownload.Method,
			"path":      result.BinaryDownload.Path,
			"status":    result.BinaryDownload.Status,
			"headers":   result.BinaryDownload.Headers,
			"record":    result.BinaryDownload.Record,
			"receipt":   result.BinaryDownload.Receipt,
		}
		if executionErr != nil {
			out["error"] = publicErrorEnvelope(executionErr)
		}
		if jsonOut {
			return writeJSON(stdout, out)
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		_, _ = fmt.Fprintln(stdout, string(b))
		return nil
	}
	if result.StatusCheck != nil {
		out := envelope{
			"kind":       "ConnectorCommandStatusCheck",
			"connector":  result.Connector,
			"command":    result.Command,
			"operation":  result.StatusCheck.Operation,
			"method":     result.StatusCheck.Method,
			"path":       result.StatusCheck.Path,
			"status":     result.StatusCheck.Status,
			"body_bytes": result.StatusCheck.BodyBytes,
			"headers":    result.StatusCheck.Headers,
			"receipt":    result.StatusCheck.Receipt,
		}
		if executionErr != nil {
			out["error"] = publicErrorEnvelope(executionErr)
		}
		if jsonOut {
			return writeJSON(stdout, out)
		}
		if len(result.StatusCheck.Headers) == 0 {
			_, err := fmt.Fprintf(stdout, "connector=%s command=%q operation=%s method=%s path=%s status=%d body_bytes=%d\n", result.Connector, result.Command, result.StatusCheck.Operation, result.StatusCheck.Method, result.StatusCheck.Path, result.StatusCheck.Status, result.StatusCheck.BodyBytes)
			return err
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		_, _ = fmt.Fprintln(stdout, string(b))
		return nil
	}
	if result.DirectRead != nil {
		out := envelope{
			"kind":      "ConnectorCommandDirectRead",
			"connector": result.Connector,
			"command":   result.Command,
			"operation": result.DirectRead.Operation,
			"method":    result.DirectRead.Method,
			"path":      result.DirectRead.Path,
			"status":    result.DirectRead.Status,
			"headers":   result.DirectRead.Headers,
			"response":  result.DirectRead.Body,
			"receipt":   result.DirectRead.Receipt,
		}
		if executionErr != nil {
			out["error"] = publicErrorEnvelope(executionErr)
		}
		if directReadPageIsReported(result.DirectRead.Page) {
			out["page"] = result.DirectRead.Page
		}
		if result.DirectRead.GraphQL != nil {
			out["graphql"] = result.DirectRead.GraphQL
		}
		if jsonOut {
			return writeJSON(stdout, out)
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		_, _ = fmt.Fprintln(stdout, string(b))
		// A human reading this must not have to infer completeness from the row
		// count. The notice goes to stderr so piping the body stays lossless.
		writeDirectReadPageNotice(stderr, result.DirectRead.Page)
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

func connectorCommandFlags(values map[string][]string) map[string][]string {
	commandFlags := map[string][]string{}
	for name, entries := range values {
		switch name {
		case "_", "credential", "connection", "config", "limit", "max-bytes", "plan", "preview", "approve", "approval-token-stdin", "confirm", "plan-name", "dest-root", "file-name", "from-env":
			continue
		default:
			commandFlags[name] = append([]string(nil), entries...)
		}
	}
	return commandFlags
}

// resolveConnectorCommandEnvironmentOnlyFlags resolves only fields the
// connector declaration marks env_only. It is intentionally command-specific:
// a generic --from-env escape hatch would let a caller bypass the declaration
// that identifies which values are sensitive and later withheld from state.
// The environment value is held only in memory and then enters the ordinary
// typed command flag coercion; this helper never formats it into an error.
func resolveConnectorCommandEnvironmentOnlyFlags(surface *connectors.CommandSurface, path []string, values map[string][]string) (map[string][]string, error) {
	if surface == nil {
		return nil, fmt.Errorf("connector command has no command surface")
	}
	command, found := connectorSurfaceCommand(surface, strings.Join(path, " "))
	if !found {
		return nil, fmt.Errorf("unknown connector command %q", strings.Join(path, " "))
	}
	envOnly := map[string]connectors.CommandSurfaceFlag{}
	for _, flag := range command.Flags {
		if flag.EnvOnly {
			envOnly[flag.Name] = flag
		}
	}
	resolved := make(map[string][]string, len(values))
	for name, entries := range values {
		if name == "from-env" {
			continue
		}
		resolved[name] = append([]string(nil), entries...)
	}
	for name := range envOnly {
		if len(resolved[name]) != 0 {
			return nil, validationErrorf("--%s must be supplied through --from-env %s=ENV", name, name)
		}
	}
	seen := map[string]struct{}{}
	for _, spec := range values["from-env"] {
		field, envName, ok := strings.Cut(spec, "=")
		if !ok || strings.TrimSpace(field) == "" || strings.TrimSpace(envName) == "" {
			return nil, usageErrorf("invalid --from-env %q, want field=ENV", spec)
		}
		field = strings.TrimSpace(field)
		envName = strings.TrimSpace(envName)
		if _, ok := envOnly[field]; !ok {
			return nil, validationErrorf("--from-env %s is not declared for connector command %q", field, command.Path)
		}
		if _, duplicate := seen[field]; duplicate {
			return nil, validationErrorf("--from-env supplies %s more than once", field)
		}
		if err := safety.ValidateIdentifier(envName, "environment variable"); err != nil {
			return nil, validationErrorf("%v", err)
		}
		value, present := os.LookupEnv(envName)
		if !present || value == "" {
			return nil, validationErrorf("environment variable %s is empty", envName)
		}
		resolved[field] = []string{value}
		seen[field] = struct{}{}
	}
	return resolved, nil
}

func resolveReversePlanEnvironmentOnlyFlags(plan app.ReversePlan, values map[string][]string) (map[string][]string, error) {
	if plan.ConnectorCommand == "" {
		return values, nil
	}
	connector, ok := appRegistry().Get(plan.DestinationConnector)
	if !ok {
		return nil, fmt.Errorf("unknown connector %q", plan.DestinationConnector)
	}
	surfaceProvider, ok := connector.(connectors.CommandSurfaceProvider)
	if !ok || surfaceProvider.CommandSurface() == nil {
		return nil, fmt.Errorf("connector %q has no command surface", plan.DestinationConnector)
	}
	return resolveConnectorCommandEnvironmentOnlyFlags(surfaceProvider.CommandSurface(), plan.ConnectorCommandPath, values)
}

func validateConnectorLifecycleFlagValues(flags parsedFlags) error {
	approvalSupplied, err := validateApprovalTokenCarrierFlags(flags)
	if err != nil {
		return err
	}
	if approvalSupplied {
		if strings.TrimSpace(flags.first("plan")) == "" {
			return usageErrorf("--approval-token-stdin requires --plan")
		}
	}
	for _, name := range []string{"plan", "confirm"} {
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

func validateApprovalTokenCarrierFlags(flags parsedFlags) (bool, error) {
	if _, supplied := flags.values["approve"]; supplied {
		return false, usageErrorf("approval tokens must be supplied with --approval-token-stdin")
	}
	values, supplied := flags.values["approval-token-stdin"]
	if !supplied {
		return false, nil
	}
	if !flags.isBare("approval-token-stdin") || len(values) != 1 {
		return false, usageErrorf("--approval-token-stdin must be a bare stdin marker")
	}
	return true, nil
}

// reverseApprovalTokenFromStdin accepts the only secret-bearing approval
// carrier shared by reverse-ETL execution paths. The marker is deliberately
// bare so an approval value can never be parsed from argv.
func reverseApprovalTokenFromStdin(flags parsedFlags, stdin io.Reader) (string, bool, error) {
	approvalSupplied, err := validateApprovalTokenCarrierFlags(flags)
	if err != nil {
		return "", false, err
	}
	if !approvalSupplied {
		return "", false, nil
	}
	token, err := readApprovalTokenFromStdin(stdin)
	if err != nil {
		return "", false, err
	}
	return token, true, nil
}

type reverseApprovalCarrier struct {
	token    string
	supplied bool
}

func prepareReverseApprovalCarrier(flags parsedFlags, stdin io.Reader) (reverseApprovalCarrier, error) {
	token, supplied, err := reverseApprovalTokenFromStdin(flags, stdin)
	if err != nil {
		return reverseApprovalCarrier{}, err
	}
	return reverseApprovalCarrier{token: token, supplied: supplied}, nil
}

func validateReverseApprovalCarrierFlags(command string, flags parsedFlags) error {
	approvalSupplied, err := validateApprovalTokenCarrierFlags(flags)
	if err != nil {
		return err
	}
	if approvalSupplied && command != "run" {
		return usageErrorf("--approval-token-stdin is only valid with reverse run")
	}
	return nil
}

func validateApprovalCarrierBeforeDispatch(args []string) error {
	if _, err := validateApprovalTokenCarrierFlags(parseFlags(args)); err != nil {
		return err
	}
	return validateReverseApprovalCarrierBeforeDispatch(args)
}

func validateRawApprovalCarrierArgs(args []string) error {
	for index, arg := range args {
		if err := validateRawApprovalCarrierArg(arg, false); err != nil {
			return err
		}
		if arg == "--root" && index+1 < len(args) {
			if err := validateRawApprovalCarrierArg(args[index+1], true); err != nil {
				return err
			}
		}
		if strings.HasPrefix(arg, "--root=") {
			if err := validateRawApprovalCarrierArg(strings.TrimPrefix(arg, "--root="), true); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateRawApprovalCarrierArg(arg string, rootValue bool) error {
	switch {
	case arg == "--approve" || strings.HasPrefix(arg, "--approve="):
		return usageErrorf("approval tokens must be supplied with --approval-token-stdin")
	case strings.HasPrefix(arg, "--approval-token-stdin="):
		return usageErrorf("--approval-token-stdin must be a bare stdin marker")
	case rootValue && arg == "--approval-token-stdin":
		return usageErrorf("--approval-token-stdin must be a bare stdin marker")
	default:
		return nil
	}
}

func rawApprovalCarrierRoot(root string) bool {
	return root == "--approve" || strings.HasPrefix(root, "--approve=") ||
		root == "--approval-token-stdin" || strings.HasPrefix(root, "--approval-token-stdin=")
}

func validateReverseApprovalCarrierBeforeDispatch(args []string) error {
	if len(args) == 0 {
		return nil
	}
	switch args[0] {
	case "reverse":
		command := ""
		if len(args) > 1 {
			command = args[1]
		}
		return validateReverseApprovalCarrierFlags(command, parseFlags(args[1:]))
	case "help", "man":
		if len(args) < 2 || args[1] != "reverse" {
			return nil
		}
		return validateReverseApprovalCarrierFlags("", parseFlags(args[2:]))
	default:
		return nil
	}
}

func prepareReverseRunApproval(args []string, stdin io.Reader) (reverseApprovalCarrier, error) {
	if len(args) == 0 {
		return reverseApprovalCarrier{}, errUsage
	}
	if err := validateReverseApprovalCarrierFlags(args[0], parseFlags(args)); err != nil {
		return reverseApprovalCarrier{}, err
	}
	if args[0] != "run" {
		return reverseApprovalCarrier{}, nil
	}
	if len(args) < 2 {
		return reverseApprovalCarrier{}, errUsage
	}
	approval, err := prepareReverseApprovalCarrier(parseFlags(args[2:]), stdin)
	if err != nil {
		return reverseApprovalCarrier{}, err
	}
	if !approval.supplied {
		return reverseApprovalCarrier{}, usageErrorf("reverse run requires --approval-token-stdin")
	}
	return approval, nil
}

// connectorCommandMaxBytes returns what the user asked for, and nothing else.
// Zero means "unset", which is how the runner is told to apply the intent's own
// default.
//
// It deliberately applies no default and no ceiling. Every intent already owns
// its own limit — commandrunner clamps direct reads to the direct-read ceiling
// and the engine clamps a binary download to its operation's declared
// max_bytes — and restating the direct-read ceiling here silently capped every
// binary download at 16 MiB, against operations declaring 100 MiB, while the
// help text promised the flag could only ever lower a cap.
func connectorCommandMaxBytes(flags parsedFlags) (int, error) {
	maxBytes, err := parseIntFlag("max-bytes", flags.first("max-bytes"), 0)
	if err != nil {
		return 0, err
	}
	if maxBytes < 0 {
		return 0, validationErrorf("invalid --max-bytes %d, want a positive integer", maxBytes)
	}
	return maxBytes, nil
}

// connectorCommandPage resolves --page and --page-cursor, the two ways to name
// which page of a direct read to return. A direct read is page-wise
// exploration, so the caller navigates; it never silently returns page one
// when it was asked for another.
//
// They are mutually exclusive by construction: a connector's declared strategy
// addresses pages either by number (page_number, offset_limit) or by opaque
// token, never both. The engine rejects the pairing that does not match the
// declared strategy.
func connectorCommandPage(flags parsedFlags) (int, string, error) {
	raw := flags.first("page")
	page, err := parseIntFlag("page", raw, 0)
	if err != nil {
		return 0, "", err
	}
	// Only an absent --page means "unset". Treating an explicit --page 0 as
	// unset returned page one for a page nobody has, and let
	// `--page 0 --page-cursor X` slip past the mutual-exclusion check below as
	// though only the cursor had been passed.
	if raw != "" && page < 1 {
		return 0, "", validationErrorf("invalid --page %d, want a positive page number", page)
	}
	cursor := flags.first("page-cursor")
	if err := connectors.ValidateDirectReadPageCursor(cursor); err != nil {
		return 0, "", validationErrorf("invalid --page-cursor: %v", err)
	}
	if raw != "" && cursor != "" {
		return 0, "", validationErrorf("--page and --page-cursor are mutually exclusive; a connector addresses pages either by number or by cursor")
	}
	return page, cursor, nil
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
	_, _ = fmt.Fprintf(stdout, "Created connector command plan %s for %s\n", plan.ID, plan.ConnectorCommand)
	if plan.ApprovalToken == "" {
		_, _ = fmt.Fprintln(stdout, "Preview required before an approval token is issued.")
	} else {
		_, _ = fmt.Fprintf(stdout, "Approval token: %s\n", plan.ApprovalToken)
	}
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

func runConnectorWriteCommandFromPlan(ctx context.Context, a *app.App, connectorName string, path []string, flags parsedFlags, approval reverseApprovalCarrier, stdout io.Writer, jsonOut bool) error {
	planID := strings.TrimSpace(flags.first("plan"))
	preview := truthyFlag(flags.first("preview"))
	plan, err := connectorCommandPlanForPath(a, planID, connectorName, path)
	if err != nil {
		return err
	}
	connector, ok := appRegistry().Get(connectorName)
	if !ok {
		return fmt.Errorf("unknown connector %q", connectorName)
	}
	surfaceProvider, ok := connector.(connectors.CommandSurfaceProvider)
	if !ok || surfaceProvider.CommandSurface() == nil {
		return fmt.Errorf("connector %q has no command surface", connectorName)
	}
	resolvedFlags, err := resolveConnectorCommandEnvironmentOnlyFlags(surfaceProvider.CommandSurface(), path, flags.values)
	if err != nil {
		return err
	}
	if approval.supplied {
		confirmation, err := connectors.ParseWriteConfirmation(flags.first("confirm"))
		if err != nil {
			return validationErrorf("invalid --confirm: %v", err)
		}
		run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: approval.token, Confirmation: confirmation, WithheldFlags: resolvedFlags})
		if err != nil && run.ID == "" {
			return err
		}
		if jsonOut {
			if outputErr := writeJSON(stdout, envelope{"kind": "ReverseRun", "run": run}); outputErr != nil {
				return outputErr
			}
			if err != nil {
				return alreadyReportedExecutionError(err)
			}
			return nil
		}
		if err != nil {
			_, _ = fmt.Fprintf(stdout, "Reverse ETL run %s ended with status %s; inspect it with pm reverse status %s\n", run.ID, run.Status, run.ID)
			return alreadyReportedExecutionError(err)
		}
		_, _ = fmt.Fprintf(stdout, "Reverse ETL run %s completed: succeeded=%d failed=%d\n", run.ID, run.RecordsSucceeded, run.RecordsFailed)
		return nil
	}
	if preview {
		plan, writePreview, err := a.PreviewConnectorCommandPlan(ctx, plan.ID, resolvedFlags)
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
		if plan.ApprovalToken != "" {
			_, _ = fmt.Fprintf(stdout, "Approval token: %s\n", plan.ApprovalToken)
		}
		return nil
	}
	return usageErrorf("connector write command with --plan requires --preview or --approval-token-stdin")
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
	var blocked *commandrunner.BlockedCommandError
	if errors.As(err, &blocked) && blocked.Failure != nil && blocked.Failure.Code() == "missing_foundation" {
		return &cliError{
			category: categoryInternal,
			code:     "missing_foundation",
			message:  blocked.Error(),
			err:      err,
		}
	}
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
	plan.ConnectorCommandPathParams = nil
	plan.ConnectorCommandQuery = nil
	plan.ConnectorCommandHeaders = nil
	plan.ConnectorCommandHeaderValues = nil
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

func directConnector(ctx context.Context, a *app.App, args []string) (connectors.Connector, connectors.RuntimeConfig, error) {
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
	if credential := flags.first("credential"); credential != "" {
		return a.ResolveConnectorCredential(ctx, name, credential, config)
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
		rows, err = a.QuerySQL(ctx, app.QuerySQLRequest{SQL: sql, Limit: limit})
		// --connection scopes --table reads only; the SQL path names its table
		// inside the query, so point at the surface that can resolve it rather
		// than at a flag that would be ignored here.
		err = warehouse.WithAmbiguityRemedy(err, "read it with `pm query run --table <table> --connection <name>`")
	} else {
		rows, err = a.QueryTable(ctx, app.QueryTableRequest{
			Table:      flags.first("table"),
			Connection: flags.first("connection"),
			Limit:      limit,
		})
		err = warehouse.WithAmbiguityRemedy(err, "pass --connection to choose one")
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

// pmcert:workflow reverse_etl
func runReverse(ctx context.Context, a *app.App, args []string, approval reverseApprovalCarrier, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 {
		return errUsage
	}
	if err := validateReverseApprovalCarrierFlags(args[0], parseFlags(args)); err != nil {
		return err
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
			SourceConnection:      flags.first("connection"),
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
		_, _ = fmt.Fprintf(stdout, "Created reverse plan %s with %d records\n", plan.ID, plan.RecordCount)
		if plan.ApprovalToken == "" {
			_, _ = fmt.Fprintln(stdout, "Preview required before an approval token is issued.")
		} else {
			_, _ = fmt.Fprintf(stdout, "Approval token: %s\n", plan.ApprovalToken)
		}
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
		flags := parseFlags(args[2:])
		resolvedFlags, err := resolveReversePlanEnvironmentOnlyFlags(plan, flags.values)
		if err != nil {
			return err
		}
		var writePreview *connectors.WritePreview
		if plan.ConnectorCommand != "" || plan.ConfirmationPolicy.Kind != "" || plan.ConfirmationChallenge != "" {
			previewedPlan, preview, err := a.PreviewReversePlan(ctx, args[1], resolvedFlags)
			if err != nil {
				return err
			}
			plan = previewedPlan
			writePreview = &preview
		}
		if jsonOut {
			env := envelope{"kind": "ReversePlanPreview", "plan": safeReversePlanForOutput(plan)}
			if writePreview != nil {
				env["write_preview"] = writePreview
			}
			return writeJSON(stdout, env)
		}
		b, _ := json.MarshalIndent(safeReversePlanForOutput(plan), "", "  ")
		_, _ = fmt.Fprintln(stdout, string(b))
		if writePreview != nil {
			for _, warning := range writePreview.Warnings {
				_, _ = fmt.Fprintf(stdout, "- %s\n", warning)
			}
		}
		if plan.ApprovalToken != "" {
			_, _ = fmt.Fprintf(stdout, "Approval token: %s\n", plan.ApprovalToken)
		}
		return nil
	case "run":
		if len(args) < 2 {
			return errUsage
		}
		flags := parseFlags(args[2:])
		if !approval.supplied {
			return usageErrorf("reverse run requires --approval-token-stdin")
		}
		plan, err := a.GetReversePlan(args[1])
		if err != nil {
			return err
		}
		resolvedFlags, err := resolveReversePlanEnvironmentOnlyFlags(plan, flags.values)
		if err != nil {
			return err
		}
		confirmation, err := connectors.ParseWriteConfirmation(flags.first("confirm"))
		if err != nil {
			return validationErrorf("invalid --confirm: %v", err)
		}
		run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: args[1], ApprovalToken: approval.token, Confirmation: confirmation, WithheldFlags: resolvedFlags})
		if err != nil && run.ID == "" {
			return err
		}
		if jsonOut {
			if outputErr := writeJSON(stdout, envelope{"kind": "ReverseRun", "run": run}); outputErr != nil {
				return outputErr
			}
			if err != nil {
				return alreadyReportedExecutionError(err)
			}
			return nil
		}
		if err != nil {
			_, _ = fmt.Fprintf(stdout, "Reverse ETL run %s ended with status %s; inspect it with pm reverse status %s\n", run.ID, run.Status, run.ID)
			return alreadyReportedExecutionError(err)
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

func withReverseExecutionApp(root string, fn func(*app.App) error) error {
	a, err := app.OpenForReverseExecution(root)
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

// writeDirectReadPageNotice tells a human reader when the page they just
// received is not the whole collection, and how to ask for the rest.
//
// Without it the human-readable path is exactly the failure this change
// exists to remove: a short result that looks complete. It writes to stderr so
// that piping the body stays lossless.
func writeDirectReadPageNotice(stderr io.Writer, page connectors.DirectReadPage) {
	if stderr == nil || page.Complete || !directReadPageIsReported(page) {
		return
	}
	switch {
	case page.NextNumber > 0:
		_, _ = fmt.Fprintf(stderr, "note: page %d of a paged result (%d records); more remain — rerun with --page %d\n", page.Number, page.Records, page.NextNumber)
	case page.NextCursor != "":
		_, _ = fmt.Fprintf(stderr, "note: partial result (%d records); more remain — rerun with --page-cursor %s\n", page.Records, page.NextCursor)
	case page.HasMore:
		// A caller who paged by setting the connector's own paging parameter
		// owns the position, so there is no --page or --page-cursor to offer.
		_, _ = fmt.Fprintf(stderr, "note: partial result (%d records); more remain — advance the paging parameter you supplied\n", page.Records)
	case page.Reason == connectors.DirectReadPageReasonAmbiguous:
		_, _ = fmt.Fprintf(stderr, "note: %d array elements returned across more than one top-level array; the paged collection cannot be identified, so completeness cannot be confirmed\n", page.Records)
	default:
		_, _ = fmt.Fprintf(stderr, "note: %d records returned; %s\n", page.Records, directReadPageIncompleteReason(page))
	}
}

// directReadPageIsReported separates "this read has no page information" from
// "this read is incomplete". Printing an incompleteness claim with an empty
// parenthetical for a read nothing measured says something untrue.
//
// It is also the invariant commandrunner enforces: a direct-read executor that
// reports no page context has not navigated, so it may not be handed --page or
// --page-cursor and quietly answer with page one.
func directReadPageIsReported(page connectors.DirectReadPage) bool {
	return page.Strategy != ""
}

func directReadPageIncompleteReason(page connectors.DirectReadPage) string {
	switch page.Reason {
	case connectors.DirectReadPageReasonNoPagination:
		return "this connector declares no pagination strategy, so completeness cannot be confirmed"
	case connectors.DirectReadPageReasonDeclaredNone:
		return `this connector declares pagination type "none", so completeness cannot be confirmed`
	case connectors.DirectReadPageReasonNotAddressable:
		return fmt.Sprintf("the declared %q pagination cannot page this request, so completeness cannot be confirmed", page.Strategy)
	case connectors.DirectReadPageReasonInvalidSpec:
		return fmt.Sprintf("this connector's declared %q pagination is unusable, so completeness cannot be confirmed", page.Strategy)
	case connectors.DirectReadPageReasonSizeNotRequested:
		return fmt.Sprintf("the declared %q pagination names no page-size parameter, so the provider chose the page size and completeness cannot be confirmed", page.Strategy)
	case "":
		return "result is not confirmed complete"
	default:
		return fmt.Sprintf("result is not confirmed complete (%s)", page.Reason)
	}
}
