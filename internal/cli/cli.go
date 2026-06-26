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

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/registryset"
	"polymetrics.ai/internal/perf"
	pmruntime "polymetrics.ai/internal/runtime"
	"polymetrics.ai/internal/runtimecheck"
	"polymetrics.ai/internal/safety"
)

type envelope map[string]any

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
		if err := writeManual(cmd, stdout, jsonOut); err != nil {
			return writeError(stdout, stderr, err, jsonOut)
		}
		return 0
	}
	if len(rest) == 0 && isManualCommand(cmd) {
		if err := writeManual(cmd, stdout, jsonOut); err != nil {
			return writeError(stdout, stderr, err, jsonOut)
		}
		return 0
	}
	var err error
	switch cmd {
	case "init":
		err = runInit(root, stdout, jsonOut)
	case "help", "man":
		err = runHelp(rest, stdout)
	case "connectors":
		err = runConnectors(rest, stdout, jsonOut)
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
		err = runAgent(rest, stdout, jsonOut)
	case "runtime":
		err = runRuntime(ctx, rest, stdout, jsonOut)
	case "perf":
		err = runPerf(ctx, rest, stdout, jsonOut)
	case "docs":
		err = runDocs(rest, stdout)
	case "skills":
		err = runSkills(rest, stdout, jsonOut)
	default:
		err = usageErrorf("unknown command %q", cmd)
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

func runHelp(args []string, stdout io.Writer) error {
	topic := ""
	if len(args) > 0 {
		topic = args[0]
	}
	text, ok := docs[topic]
	if !ok {
		return fmt.Errorf("help topic %q not found", topic)
	}
	fmt.Fprint(stdout, text)
	return nil
}

func isManualCommand(cmd string) bool {
	if cmd == "init" || cmd == "help" || cmd == "man" {
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

func runConnectors(args []string, stdout io.Writer, jsonOut bool) error {
	registry := appRegistry()
	if len(args) == 0 {
		return errUsage
	}
	switch args[0] {
	case "list":
		flags := parseFlags(args[1:])
		if flags.first("all") != "" {
			defs := connectors.ConnectorCatalog()
			if jsonOut {
				return writeJSON(stdout, envelope{"kind": "ConnectorCatalog", "count": len(defs), "summary": connectors.ConnectorCatalogCounts(defs), "connectors": defs})
			}
			for _, item := range defs {
				fmt.Fprintf(stdout, "%s\t%s\t%s\t%s\t%s\n", item.Slug, item.Type, item.ImplementationStatus, item.RuntimeKind, item.DocumentationURL)
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
		filter, err := connectorCatalogFilter(flags)
		if err != nil {
			return err
		}
		defs := connectors.FilterConnectorCatalog(connectors.ConnectorCatalog(), filter)
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "ConnectorCatalog", "count": len(defs), "summary": connectors.ConnectorCatalogCounts(defs), "connectors": defs})
		}
		for _, item := range defs {
			fmt.Fprintf(stdout, "%s\t%s\t%s\t%s\t%s\n", item.Slug, item.Type, item.ImplementationStatus, item.RuntimeKind, item.DocumentationURL)
		}
		return nil
	case "port-plan":
		return runConnectorPortPlan(args[1:], stdout, jsonOut)
	case "inspect", "help", "man", "docs":
		if len(args) < 2 {
			return errUsage
		}
		if err := safety.ValidateIdentifier(args[1], "connector"); err != nil {
			return validationErrorf("%v", err)
		}
		// Prefer a live, registered connector: its manifest is more authoritative
		// than the catalog stub. Accepts bare names ("github") and legacy slugs
		// ("source-github") since both are registered. Fall back to the catalog
		// definition for connectors that are not yet natively ported.
		if c, ok := registry.Get(args[1]); ok {
			if jsonOut {
				return writeJSON(stdout, envelope{"kind": "Connector", "connector": c.Metadata(), "manifest": connectors.ManifestOf(c)})
			}
			fmt.Fprint(stdout, connectors.RenderConnectorManual(c))
			return nil
		}
		if def, found := connectors.ConnectorDefinitionBySlug(args[1]); found {
			if jsonOut {
				return writeJSON(stdout, envelope{"kind": "ConnectorDefinition", "connector": def})
			}
			fmt.Fprint(stdout, connectors.RenderConnectorDefinitionManual(def))
			return nil
		}
		return fmt.Errorf("connector %q not found", args[1])
	default:
		return errUsage
	}
}

func runConnectorPortPlan(args []string, stdout io.Writer, jsonOut bool) error {
	flags := parseFlags(args)
	positionals := flags.values["_"]
	if flags.first("all") != "" {
		plans := connectors.NativePortPlans(connectors.ConnectorCatalog())
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "NativePortPlanList", "count": len(plans), "summary": connectors.NativePortPlanCounts(plans), "plans": plans})
		}
		for _, plan := range plans {
			fmt.Fprintf(stdout, "%s\t%s\t%s\twave_%d\t%s\n", plan.Slug, plan.Type, plan.Family, plan.PriorityWave, plan.ImplementationStatus)
		}
		return nil
	}
	if len(positionals) != 1 {
		return errUsage
	}
	if err := safety.ValidateIdentifier(positionals[0], "connector"); err != nil {
		return validationErrorf("%v", err)
	}
	plan, ok := connectors.NativePortPlanBySlug(positionals[0])
	if !ok {
		return fmt.Errorf("connector %q not found", positionals[0])
	}
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "NativePortPlan", "plan": plan})
	}
	fmt.Fprint(stdout, connectors.RenderNativePortPlanManual(plan))
	return nil
}

func connectorCatalogFilter(flags parsedFlags) (connectors.ConnectorCatalogFilter, error) {
	filter := connectors.ConnectorCatalogFilter{Stage: flags.first("stage")}
	switch value := flags.first("type"); value {
	case "":
	case string(connectors.ConnectorTypeSource):
		filter.Type = connectors.ConnectorTypeSource
	case string(connectors.ConnectorTypeDestination):
		filter.Type = connectors.ConnectorTypeDestination
	default:
		return filter, validationErrorf("invalid --type %q, want source or destination", value)
	}
	return filter, nil
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

func directConnector(a *app.App, args []string) (connectors.Connector, connectors.RuntimeConfig, error) {
	flags := parseFlags(args)
	name := flags.first("connector")
	if name == "" {
		return nil, connectors.RuntimeConfig{}, errors.New("missing --connector")
	}
	if err := safety.ValidateIdentifier(name, "connector"); err != nil {
		return nil, connectors.RuntimeConfig{}, validationErrorf("%v", err)
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
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "QueryResult", "rows": rows, "count": len(rows)})
	}
	for _, row := range rows {
		b, _ := json.Marshal(row)
		fmt.Fprintln(stdout, string(b))
	}
	return nil
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
			return writeJSON(stdout, envelope{"kind": "ReversePlanList", "plans": plans, "runs": runs})
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
			safe := plan
			safe.ApprovalToken = ""
			return writeJSON(stdout, envelope{"kind": "ReversePlan", "plan": safe, "approval_required": true})
		}
		fmt.Fprintf(stdout, "Created reverse plan %s with %d records\nApproval token: %s\n", plan.ID, plan.RecordCount, plan.ApprovalToken)
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
			return writeJSON(stdout, envelope{"kind": "ReversePlanPreview", "plan": plan})
		}
		b, _ := json.MarshalIndent(plan, "", "  ")
		fmt.Fprintln(stdout, string(b))
		return nil
	case "run":
		if len(args) < 2 {
			return errUsage
		}
		flags := parseFlags(args[2:])
		run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: args[1], ApprovalToken: flags.first("approve")})
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

func runAgent(args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 || args[0] != "plan" {
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

func recordRuntimeETL(ctx context.Context, run app.Run) error {
	cfg := runtimecheck.FromEnv()
	report := runtimecheck.Doctor(ctx, cfg)
	if !runtimecheck.Healthy(report) {
		return fmt.Errorf("runtime dependencies are not healthy; run `pm runtime doctor --json` for details")
	}
	dragonfly := pmruntime.OpenDragonflyLeaseStore(cfg.DragonflyAddr)
	defer dragonfly.Close()
	pg, err := pmruntime.OpenPostgresRunLedger(ctx, cfg.PostgresURL)
	if err != nil {
		return err
	}
	defer pg.Close()
	if err := pg.Migrate(ctx); err != nil {
		return err
	}
	module := pmruntime.Module{Leases: dragonfly, Ledger: pg}
	return module.RecordRunWithLease(ctx, pmruntime.LeaseRequest{Key: "polymetrics:etl:" + run.ID, Value: "recording", TTL: 30 * time.Second}, pmruntime.RunRecord{
		ID:             run.ID,
		Mode:           "runtime-backed",
		Operation:      "etl",
		Status:         run.Status,
		RecordsRead:    run.RecordsRead,
		RecordsWritten: run.RecordsLoaded,
		Duration:       run.CompletedAt.Sub(run.StartedAt).Nanoseconds(),
		CreatedAt:      run.StartedAt,
	})
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
	return registryset.New()
}
