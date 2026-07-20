package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/events"
	"polymetrics.ai/internal/safety"
	pmui "polymetrics.ai/internal/ui"
	rundashboard "polymetrics.ai/internal/ui/run"
)

type etlDirectFlags struct {
	Connectors []string
	Configs    []string
}

type etlReadFlags struct {
	Connectors []string
	Streams    []string
	Limits     []string
	Configs    []string
}

type etlRunFlags struct {
	Connections []string
	Streams     []string
	BatchSizes  []string
	Runtimes    []string
}

func newETLCobraCommand(ctx context.Context, cfg config.Config, root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "etl",
		Args:              cobra.ArbitraryArgs,
		SilenceErrors:     true,
		SilenceUsage:      true,
		ValidArgsFunction: completeNoFile,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				return errUsage
			}
			return markCobraLegacyError(writeManual("etl", stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "etl", stdout, jsonOut)
	cmd.AddCommand(newETLCheckCobraCommand(ctx, root, stdout, jsonOut))
	cmd.AddCommand(newETLCatalogCobraCommand(ctx, root, stdout, jsonOut))
	cmd.AddCommand(newETLReadCobraCommand(ctx, root, stdout, jsonOut))
	cmd.AddCommand(newETLRunCobraCommand(ctx, cfg, root, stdout, jsonOut))
	cmd.AddCommand(newETLStatusCobraCommand(root, stdout, jsonOut))
	cmd.AddCommand(newETLHelpCobraCommand(stdout, jsonOut))
	return cmd
}

func newETLCheckCobraCommand(ctx context.Context, root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	var flags etlDirectFlags
	cmd := newETLActionCobraCommand("check", func(_ *cobra.Command, _ []string) error {
		return markCobraLegacyError(withApp(root, func(a *app.App) error {
			return runETLCheck(ctx, a, flags, stdout, jsonOut)
		}))
	})
	setManualHelp(cmd, "etl", stdout, jsonOut)
	addETLStringArrayFlag(cmd, &flags.Connectors, "connector", "connector name")
	addETLStringArrayFlag(cmd, &flags.Configs, "config", "connector config key=value")
	return cmd
}

func newETLCatalogCobraCommand(ctx context.Context, root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	var flags etlDirectFlags
	cmd := newETLActionCobraCommand("catalog", func(_ *cobra.Command, _ []string) error {
		return markCobraLegacyError(withApp(root, func(a *app.App) error {
			return runETLCatalog(ctx, a, flags, stdout, jsonOut)
		}))
	})
	setManualHelp(cmd, "etl", stdout, jsonOut)
	addETLStringArrayFlag(cmd, &flags.Connectors, "connector", "connector name")
	addETLStringArrayFlag(cmd, &flags.Configs, "config", "connector config key=value")
	return cmd
}

func newETLReadCobraCommand(ctx context.Context, root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	var flags etlReadFlags
	cmd := newETLActionCobraCommand("read", func(_ *cobra.Command, _ []string) error {
		return markCobraLegacyError(withApp(root, func(a *app.App) error {
			return runETLRead(ctx, a, flags, stdout, jsonOut)
		}))
	})
	setManualHelp(cmd, "etl", stdout, jsonOut)
	addETLStringArrayFlag(cmd, &flags.Connectors, "connector", "connector name")
	addETLStringArrayFlag(cmd, &flags.Streams, "stream", "stream name")
	addETLStringArrayFlag(cmd, &flags.Limits, "limit", "maximum records to read")
	addETLStringArrayFlag(cmd, &flags.Configs, "config", "connector config key=value")
	return cmd
}

func newETLRunCobraCommand(ctx context.Context, cfg config.Config, root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	var flags etlRunFlags
	cmd := newETLActionCobraCommand("run", func(_ *cobra.Command, _ []string) error {
		return markCobraLegacyError(withApp(root, func(a *app.App) error {
			return runETLRun(ctx, cfg, a, flags, stdout, jsonOut)
		}))
	})
	setManualHelp(cmd, "etl", stdout, jsonOut)
	addETLStringArrayFlag(cmd, &flags.Connections, "connection", "configured connection name")
	addETLStringArrayFlag(cmd, &flags.Streams, "stream", "configured stream name")
	addETLStringArrayFlag(cmd, &flags.BatchSizes, "batch-size", "records per bounded write batch")
	addETLStringArrayFlag(cmd, &flags.Runtimes, "runtime", "record the completed run through optional runtime services")
	return cmd
}

func newETLStatusCobraCommand(root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := newETLActionCobraCommand("status <run-id>", func(cmd *cobra.Command, _ []string) error {
		state, ok := cmd.Context().Value(etlCommandStateKey{}).(etlCommandState)
		if !ok || !state.statusRunIDSet {
			return errUsage
		}
		return markCobraLegacyError(withApp(root, func(a *app.App) error {
			return runETLStatus(a, state.statusRunID, stdout, jsonOut)
		}))
	})
	setManualHelp(cmd, "etl", stdout, jsonOut)
	return cmd
}

func newETLHelpCobraCommand(stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := newETLActionCobraCommand("help", func(_ *cobra.Command, _ []string) error {
		return markCobraLegacyError(writeManual("etl", stdout, jsonOut))
	})
	cmd.Hidden = true
	setManualHelp(cmd, "etl", stdout, jsonOut)
	return cmd
}

func newETLActionCobraCommand(use string, run func(*cobra.Command, []string) error) *cobra.Command {
	return &cobra.Command{
		Use:           use,
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		ValidArgsFunction: completeNoFile,
		RunE:              run,
	}
}

func addETLStringArrayFlag(cmd *cobra.Command, target *[]string, name, usage string) {
	cmd.Flags().StringArrayVar(target, name, nil, usage)
	if flag := cmd.Flags().Lookup(name); flag != nil {
		flag.NoOptDefVal = "true"
	}
}

func runETLCheck(ctx context.Context, a *app.App, flags etlDirectFlags, stdout io.Writer, jsonOut bool) error {
	connector, runtime, err := directETLConnector(a, flags.Connectors, flags.Configs)
	if err != nil {
		return err
	}
	if err := connector.Check(ctx, runtime); err != nil {
		return err
	}
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "ETLCheck", "connector": connector.Name(), "status": "ok"})
	}
	fmt.Fprintf(stdout, "Connector %s check ok\n", connector.Name())
	return nil
}

func runETLCatalog(ctx context.Context, a *app.App, flags etlDirectFlags, stdout io.Writer, jsonOut bool) error {
	connector, runtime, err := directETLConnector(a, flags.Connectors, flags.Configs)
	if err != nil {
		return err
	}
	catalog, err := connector.Catalog(ctx, runtime)
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
}

func runETLRead(ctx context.Context, a *app.App, flags etlReadFlags, stdout io.Writer, jsonOut bool) error {
	connector, runtime, err := directETLConnector(a, flags.Connectors, flags.Configs)
	if err != nil {
		return err
	}
	stream := lastString(flags.Streams)
	limit, err := parseIntFlag("limit", valueOr(lastString(flags.Limits), "100"), 100)
	if err != nil {
		return err
	}
	if limit <= 0 {
		limit = 100
	}
	rows := make([]connectors.Record, 0, limit)
	err = connector.Read(ctx, connectors.ReadRequest{Stream: stream, Config: runtime, Limit: limit}, connectors.LimitEmitter(limit, func(record connectors.Record) error {
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
		value, err := json.Marshal(row)
		if err != nil {
			return fmt.Errorf("encode ETL record: %w", err)
		}
		fmt.Fprintln(stdout, string(value))
	}
	return nil
}

func runETLRun(ctx context.Context, cfg config.Config, a *app.App, flags etlRunFlags, stdout io.Writer, jsonOut bool) error {
	batchSize, err := parseIntFlag("batch-size", lastString(flags.BatchSizes), 0)
	if err != nil {
		return err
	}
	if uiDetectionFromContext(ctx).Mode == pmui.ModeTUI {
		return runETLRunDashboard(ctx, cfg, a, flags, batchSize, stdout)
	}
	run, err := a.RunETL(ctx, app.RunETLRequest{
		Connection: lastString(flags.Connections),
		Stream:     lastString(flags.Streams),
		BatchSize:  batchSize,
	})
	if err != nil {
		return err
	}
	runtimeRecorded := false
	if lastString(flags.Runtimes) == "true" {
		if err := recordRuntimeETL(ctx, run, cfg); err != nil {
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
}

func runETLRunDashboard(ctx context.Context, cfg config.Config, a *app.App, flags etlRunFlags, batchSize int, stdout io.Writer) error {
	detection := uiDetectionFromContext(ctx)
	stream := lastString(flags.Streams)
	width, height := dashboardDimensions(stdout)
	session := rundashboard.NewSession(ctx, rundashboard.SessionOptions{
		Config: rundashboard.Config{
			Title:         "ETL",
			Name:          stream,
			Steps:         []rundashboard.Step{{ID: stream, Kind: "stream", Detail: lastString(flags.Connections)}},
			Width:         width,
			Height:        height,
			ASCII:         detection.ASCII,
			NoColor:       !detection.Color,
			StartedAt:     time.Now(),
			ResumeCommand: fmt.Sprintf("pm etl run --connection %s --stream %s", lastString(flags.Connections), stream),
		},
		Upstream: events.FromContext(ctx),
		Interval: 100 * time.Millisecond,
		Output:   stdout,
	})
	var run app.Run
	err := session.Execute(func(runCtx context.Context) error {
		var runErr error
		run, runErr = a.RunETL(runCtx, app.RunETLRequest{
			Connection: lastString(flags.Connections),
			Stream:     stream,
			BatchSize:  batchSize,
		})
		if runErr == nil && lastString(flags.Runtimes) == "true" {
			runErr = recordRuntimeETL(runCtx, run, cfg)
		}
		return runErr
	})
	return err
}

func runETLStatus(a *app.App, runID string, stdout io.Writer, jsonOut bool) error {
	run, err := a.GetRun(runID)
	if err != nil {
		return err
	}
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "ETLRun", "run": run})
	}
	fmt.Fprintf(stdout, "%s\t%s\tread=%d loaded=%d failed=%d\n", run.ID, run.Status, run.RecordsRead, run.RecordsLoaded, run.RecordsFailed)
	return nil
}

func directETLConnector(a *app.App, connectorsFlag, configs []string) (connectors.Connector, connectors.RuntimeConfig, error) {
	name := lastString(connectorsFlag)
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
	connectorConfig, err := keyValues(configs)
	if err != nil {
		return nil, connectors.RuntimeConfig{}, err
	}
	return connector, connectors.RuntimeConfig{
		ProjectDir: a.ProjectDir(),
		Config:     connectorConfig,
	}, nil
}
