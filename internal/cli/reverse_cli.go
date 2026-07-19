package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"polymetrics.ai/internal/app"
)

type reversePlanFlags struct {
	SourceTables []string
	Destinations []string
	Mappings     []string
	Actions      []string
	Limits       []string
}

type reverseRunFlags struct {
	Approvals     []string
	Confirmations []string
}

func newReverseCobraCommand(ctx context.Context, root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "reverse",
		Args:              cobra.ArbitraryArgs,
		SilenceErrors:     true,
		SilenceUsage:      true,
		ValidArgsFunction: completeNoFile,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				return errUsage
			}
			return markCobraLegacyError(writeManual("reverse", stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "reverse", stdout, jsonOut)
	cmd.AddCommand(newReverseListCobraCommand(root, stdout, jsonOut))
	cmd.AddCommand(newReversePlanCobraCommand(ctx, root, stdout, jsonOut))
	cmd.AddCommand(newReversePreviewCobraCommand(ctx, root, stdout, jsonOut))
	cmd.AddCommand(newReverseRunCobraCommand(ctx, root, stdout, jsonOut))
	cmd.AddCommand(newReverseStatusCobraCommand(root, stdout, jsonOut))
	cmd.AddCommand(newReverseHelpCobraCommand(stdout, jsonOut))
	return cmd
}

func newReverseListCobraCommand(root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := newReverseActionCobraCommand("list", func(_ *cobra.Command, _ []string) error {
		return markCobraLegacyError(withApp(root, func(a *app.App) error {
			return runReverseList(a, stdout, jsonOut)
		}))
	})
	setManualHelp(cmd, "reverse", stdout, jsonOut)
	return cmd
}

func newReversePlanCobraCommand(ctx context.Context, root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	var flags reversePlanFlags
	cmd := newReverseActionCobraCommand("plan <name>", func(cmd *cobra.Command, _ []string) error {
		name, ok := reverseOperand(cmd)
		if !ok {
			return errUsage
		}
		return markCobraLegacyError(withApp(root, func(a *app.App) error {
			return runReversePlan(ctx, a, name, flags, stdout, jsonOut)
		}))
	})
	setManualHelp(cmd, "reverse", stdout, jsonOut)
	addReverseStringArrayFlag(cmd, &flags.SourceTables, "source-table", "local warehouse table to read")
	addReverseStringArrayFlag(cmd, &flags.Destinations, "destination", "destination connector:credential")
	addReverseStringArrayFlag(cmd, &flags.Mappings, "map", "field mapping source:destination")
	addReverseStringArrayFlag(cmd, &flags.Actions, "action", "destination write action")
	addReverseStringArrayFlag(cmd, &flags.Limits, "limit", "maximum source rows in the plan")
	return cmd
}

func newReversePreviewCobraCommand(ctx context.Context, root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := newReverseActionCobraCommand("preview <plan-id>", func(cmd *cobra.Command, _ []string) error {
		planID, ok := reverseOperand(cmd)
		if !ok {
			return errUsage
		}
		return markCobraLegacyError(withApp(root, func(a *app.App) error {
			return runReversePreview(ctx, a, planID, stdout, jsonOut)
		}))
	})
	setManualHelp(cmd, "reverse", stdout, jsonOut)
	return cmd
}

func newReverseRunCobraCommand(ctx context.Context, root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	var flags reverseRunFlags
	cmd := newReverseActionCobraCommand("run <plan-id>", func(cmd *cobra.Command, _ []string) error {
		planID, ok := reverseOperand(cmd)
		if !ok {
			return errUsage
		}
		return markCobraLegacyError(withApp(root, func(a *app.App) error {
			return runReverseRun(ctx, a, planID, flags, stdout, jsonOut)
		}))
	})
	setManualHelp(cmd, "reverse", stdout, jsonOut)
	addReverseStringArrayFlag(cmd, &flags.Approvals, "approve", "approval token from human plan output")
	addReverseStringArrayFlag(cmd, &flags.Confirmations, "confirm", "typed confirmation challenge")
	return cmd
}

func newReverseStatusCobraCommand(root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := newReverseActionCobraCommand("status <run-id>", func(cmd *cobra.Command, _ []string) error {
		runID, ok := reverseOperand(cmd)
		if !ok {
			return errUsage
		}
		return markCobraLegacyError(withApp(root, func(a *app.App) error {
			return runReverseStatus(a, runID, stdout, jsonOut)
		}))
	})
	setManualHelp(cmd, "reverse", stdout, jsonOut)
	return cmd
}

func newReverseHelpCobraCommand(stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := newReverseActionCobraCommand("help", func(_ *cobra.Command, _ []string) error {
		return markCobraLegacyError(writeManual("reverse", stdout, jsonOut))
	})
	cmd.Hidden = true
	setManualHelp(cmd, "reverse", stdout, jsonOut)
	return cmd
}

func newReverseActionCobraCommand(use string, run func(*cobra.Command, []string) error) *cobra.Command {
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

func addReverseStringArrayFlag(cmd *cobra.Command, target *[]string, name, usage string) {
	cmd.Flags().StringArrayVar(target, name, nil, usage)
	if flag := cmd.Flags().Lookup(name); flag != nil {
		flag.NoOptDefVal = "true"
	}
}

func reverseOperand(cmd *cobra.Command) (string, bool) {
	state, ok := cmd.Context().Value(reverseCommandStateKey{}).(reverseCommandState)
	if !ok || !state.operandSet {
		return "", false
	}
	return state.operand, true
}

func runReverseList(a *app.App, stdout io.Writer, jsonOut bool) error {
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
}

func runReversePlan(ctx context.Context, a *app.App, name string, flags reversePlanFlags, stdout io.Writer, jsonOut bool) error {
	dest, err := parseEndpoint(lastString(flags.Destinations))
	if err != nil {
		return err
	}
	mappings, err := colonValues(flags.Mappings)
	if err != nil {
		return err
	}
	limit, err := parseIntFlag("limit", lastString(flags.Limits), 0)
	if err != nil {
		return err
	}
	plan, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
		Name:                  name,
		SourceTable:           lastString(flags.SourceTables),
		DestinationConnector:  dest.Connector,
		DestinationCredential: dest.Credential,
		DestinationConfig:     dest.Config,
		Action:                valueOr(lastString(flags.Actions), "upsert"),
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
}

func runReversePreview(ctx context.Context, a *app.App, planID string, stdout io.Writer, jsonOut bool) error {
	plan, err := a.GetReversePlan(planID)
	if err != nil {
		return err
	}
	if jsonOut {
		env := envelope{"kind": "ReversePlanPreview", "plan": safeReversePlanForOutput(plan)}
		if plan.ConnectorCommand != "" {
			safePlan, writePreview, err := a.PreviewConnectorCommandPlan(ctx, planID)
			if err != nil {
				return err
			}
			env["plan"] = safeReversePlanForOutput(safePlan)
			env["write_preview"] = writePreview
		}
		return writeJSON(stdout, env)
	}
	value, _ := json.MarshalIndent(safeReversePlanForOutput(plan), "", "  ")
	fmt.Fprintln(stdout, string(value))
	return nil
}

func runReverseRun(ctx context.Context, a *app.App, planID string, flags reverseRunFlags, stdout io.Writer, jsonOut bool) error {
	run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        planID,
		ApprovalToken: lastString(flags.Approvals),
		Confirmation:  lastString(flags.Confirmations),
	})
	if err != nil {
		return err
	}
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "ReverseRun", "run": run})
	}
	fmt.Fprintf(stdout, "Reverse ETL run %s completed: succeeded=%d failed=%d\n", run.ID, run.RecordsSucceeded, run.RecordsFailed)
	return nil
}

func runReverseStatus(a *app.App, runID string, stdout io.Writer, jsonOut bool) error {
	run, err := a.GetReverseRun(runID)
	if err != nil {
		return err
	}
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "ReverseRun", "run": run})
	}
	fmt.Fprintf(stdout, "%s\t%s\tplan=%s\tstaged=%d succeeded=%d failed=%d\n", run.ID, run.Status, run.PlanID, run.RecordsStaged, run.RecordsSucceeded, run.RecordsFailed)
	return nil
}
