package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/safety"
	"polymetrics.ai/internal/synctransport"
)

const (
	githubIssueLabelTransportCommand = "github-issue-label"
	maxApprovalTokenStdinBytes       = 4096
)

const etlTransportHelp = `NAME
  pm etl transport - run closed approval-bound transport lifecycles

SYNOPSIS
  pm etl transport github-issue-label plan --connection <name> [--json]
  pm etl transport github-issue-label preview <plan-id> [--json]
  pm etl run --connection <name> --stream issues --batch-size 1 --approval-plan <plan-id> --approval-token-stdin --confirm destructive [--json]
  pm etl transport github-issue-label cleanup plan --connection <name> --forward-plan <plan-id> [--json]
  pm etl transport github-issue-label cleanup run <plan-id> --connection <name> --approval-token-stdin --confirm destructive [--json]

DESCRIPTION
  github-issue-label is the one closed GitHub transport walking slice. The
  connection owns the repository, source issue, target issue, label, action,
  and credential configuration. This command family accepts none of those as
  command arguments.

  Create a plan, preview it in human output to receive one ephemeral approval
  token, then send that token as one bounded line on standard input to pm etl
  run. The forward run remains the ordinary source -> durable warehouse ->
  reopen -> typed GitHub mutation -> independent read-back -> checkpoint path.

  Cleanup is a separately planned, previewed, one-time approved typed inverse.
  A GitHub missing-label DELETE accepted by the declared missing_ok_status is a
  successful cleanup; replaying a consumed cleanup approval is rejected before
  another provider request.

SECURITY
  Raw approval tokens are accepted only through --approval-token-stdin. They
  are never accepted in argv, environment variables, files, JSON output, or
  persisted project state. JSON previews report only that a token was issued.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error
  3 validation error
`

type etlTransportPlanOutput struct {
	ID            string                       `json:"id"`
	Status        string                       `json:"status"`
	Mode          string                       `json:"mode"`
	ConnectionID  string                       `json:"connection_id"`
	Action        string                       `json:"action"`
	RecordCount   int                          `json:"record_count"`
	Confirmation  connectors.WriteConfirmation `json:"confirmation"`
	CreatedAt     time.Time                    `json:"created_at"`
	ExpiresAt     time.Time                    `json:"expires_at"`
	ForwardPlanID string                       `json:"forward_plan_id,omitempty"`
}

type etlTransportWritePreviewOutput struct {
	RecordsStaged int    `json:"records_staged"`
	Action        string `json:"action"`
	Digest        string `json:"digest"`
}

type strictTransportFlags struct {
	values map[string]string
}

func (p strictTransportFlags) value(name string) string { return p.values[name] }

type strictTransportFlagSpec struct {
	allowBare bool
	bareOnly  bool
}

func runETLTransport(ctx context.Context, a *app.App, args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 || isOnlyTransportHelp(args) {
		return writeETLTransportManual(stdout, jsonOut, "etl transport")
	}
	if args[0] != githubIssueLabelTransportCommand {
		return usageErrorf("unknown etl transport %q", args[0])
	}
	return runGitHubIssueLabelTransport(ctx, a, args[1:], stdout, jsonOut)
}

func runGitHubIssueLabelTransport(ctx context.Context, a *app.App, args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 || isOnlyTransportHelp(args) {
		return writeETLTransportManual(stdout, jsonOut, "etl transport github-issue-label")
	}
	switch args[0] {
	case "plan":
		if isOnlyTransportHelp(args[1:]) {
			return writeETLTransportManual(stdout, jsonOut, "etl transport github-issue-label")
		}
		flags, err := parseStrictTransportFlags(args[1:], map[string]strictTransportFlagSpec{
			"connection": {},
		})
		if err != nil {
			return err
		}
		connectionID, err := githubIssueLabelTransportConnectionID(a, flags.value("connection"))
		if err != nil {
			return err
		}
		plan, err := a.PlanGitHubIssueLabelTransport(ctx, connectionID)
		if err != nil {
			return err
		}
		return writeETLTransportPlan(stdout, jsonOut, plan)
	case "preview":
		if len(args) == 3 && isOnlyTransportHelp(args[2:]) {
			return writeETLTransportManual(stdout, jsonOut, "etl transport github-issue-label")
		}
		if len(args) != 2 || isOnlyTransportHelp(args[1:]) {
			if isOnlyTransportHelp(args[1:]) {
				return writeETLTransportManual(stdout, jsonOut, "etl transport github-issue-label")
			}
			return errUsage
		}
		if err := validateTransportPlanID(args[1]); err != nil {
			return err
		}
		plan, preview, err := a.PreviewGitHubIssueLabelTransport(ctx, args[1])
		if err != nil {
			return err
		}
		return writeETLTransportPreview(stdout, jsonOut, plan, preview)
	case "cleanup":
		return runGitHubIssueLabelTransportCleanup(ctx, a, args[1:], stdout, jsonOut)
	default:
		return usageErrorf("unknown GitHub issue-label transport command %q", args[0])
	}
}

func runGitHubIssueLabelTransportCleanup(ctx context.Context, a *app.App, args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 || isOnlyTransportHelp(args) {
		return writeETLTransportManual(stdout, jsonOut, "etl transport github-issue-label")
	}
	switch args[0] {
	case "plan":
		if isOnlyTransportHelp(args[1:]) {
			return writeETLTransportManual(stdout, jsonOut, "etl transport github-issue-label")
		}
		flags, err := parseStrictTransportFlags(args[1:], map[string]strictTransportFlagSpec{
			"connection":   {},
			"forward-plan": {},
		})
		if err != nil {
			return err
		}
		connectionID, err := githubIssueLabelTransportConnectionID(a, flags.value("connection"))
		if err != nil {
			return err
		}
		if err := validateTransportPlanID(flags.value("forward-plan")); err != nil {
			return err
		}
		plan, err := a.PlanGitHubIssueLabelTransportCleanup(ctx, connectionID, flags.value("forward-plan"))
		if err != nil {
			return err
		}
		return writeETLTransportPlan(stdout, jsonOut, plan)
	case "run":
		if len(args) < 2 || isOnlyTransportHelp(args[1:]) || isOnlyTransportHelp(args[2:]) {
			if isOnlyTransportHelp(args[1:]) || isOnlyTransportHelp(args[2:]) {
				return writeETLTransportManual(stdout, jsonOut, "etl transport github-issue-label")
			}
			return errUsage
		}
		if err := validateTransportPlanID(args[1]); err != nil {
			return err
		}
		flags, err := parseStrictTransportFlags(args[2:], map[string]strictTransportFlagSpec{
			"connection":           {},
			"approval-token-stdin": {allowBare: true, bareOnly: true},
			"confirm":              {},
		})
		if err != nil {
			return err
		}
		connectionID, err := githubIssueLabelTransportConnectionID(a, flags.value("connection"))
		if err != nil {
			return err
		}
		approval, err := approvalFromStrictTransportFlags(args[1], flags, true)
		if err != nil {
			return err
		}
		approval.ApprovalToken, err = readApprovalTokenFromStdin(os.Stdin)
		if err != nil {
			return err
		}
		plan, err := a.GetReversePlan(args[1])
		if err != nil {
			return err
		}
		if plan.Mode != "github_issue_label_transport_cleanup" {
			return validationErrorf("plan %q is not a GitHub issue-label transport cleanup plan", args[1])
		}
		result, err := a.ApplyGitHubIssueLabelTransportCleanup(ctx, connectionID, approval)
		if err != nil {
			return err
		}
		return writeETLTransportCleanupRun(stdout, jsonOut, plan, result)
	default:
		return usageErrorf("unknown GitHub issue-label transport cleanup command %q", args[0])
	}
}

func parseETLRunTransportApproval(args []string, stdin io.Reader) (synctransport.DestinationApproval, bool, strictTransportFlags, error) {
	if !hasTransportApprovalInput(args) {
		return synctransport.DestinationApproval{}, false, strictTransportFlags{}, nil
	}
	flags, err := parseStrictTransportFlags(args, map[string]strictTransportFlagSpec{
		"connection":           {},
		"stream":               {},
		"batch-size":           {},
		"approval-plan":        {},
		"approval-token-stdin": {allowBare: true, bareOnly: true},
		"confirm":              {},
	})
	if err != nil {
		return synctransport.DestinationApproval{}, true, strictTransportFlags{}, err
	}
	if flags.value("connection") == "" || flags.value("stream") == "" || flags.value("batch-size") == "" {
		return synctransport.DestinationApproval{}, true, strictTransportFlags{}, validationErrorf("approved GitHub transport ETL requires --connection, --stream issues, and --batch-size 1")
	}
	if flags.value("stream") != "issues" {
		return synctransport.DestinationApproval{}, true, strictTransportFlags{}, validationErrorf("approved GitHub transport ETL requires --stream issues")
	}
	batchSize, err := parseIntFlag("batch-size", flags.value("batch-size"), 0)
	if err != nil {
		return synctransport.DestinationApproval{}, true, strictTransportFlags{}, err
	}
	if batchSize != 1 {
		return synctransport.DestinationApproval{}, true, strictTransportFlags{}, validationErrorf("approved GitHub transport ETL requires --batch-size 1")
	}
	if _, err := githubIssueLabelTransportConnectionIDFromName(flags.value("connection")); err != nil {
		return synctransport.DestinationApproval{}, true, strictTransportFlags{}, err
	}
	approval, err := approvalFromStrictTransportFlags(flags.value("approval-plan"), flags, false)
	if err != nil {
		return synctransport.DestinationApproval{}, true, strictTransportFlags{}, err
	}
	token, err := readApprovalTokenFromStdin(stdin)
	if err != nil {
		return synctransport.DestinationApproval{}, true, strictTransportFlags{}, err
	}
	approval.ApprovalToken = token
	return approval, true, flags, nil
}

func runApprovedGitHubIssueLabelTransportETL(ctx context.Context, a *app.App, flags strictTransportFlags, approval synctransport.DestinationApproval, stdout io.Writer, jsonOut bool) error {
	batchSize, err := parseIntFlag("batch-size", flags.value("batch-size"), 0)
	if err != nil {
		return err
	}
	run, err := a.RunETL(ctx, app.RunETLRequest{
		Connection:          flags.value("connection"),
		Stream:              flags.value("stream"),
		BatchSize:           batchSize,
		DestinationApproval: approval,
	})
	if err != nil {
		return err
	}
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "ETLRun", "run": run, "runtime_recorded": false})
	}
	_, _ = fmt.Fprintf(stdout, "ETL run %s completed: read=%d loaded=%d failed=%d\n", run.ID, run.RecordsRead, run.RecordsLoaded, run.RecordsFailed)
	return nil
}

func approvalFromStrictTransportFlags(planID string, flags strictTransportFlags, cleanup bool) (synctransport.DestinationApproval, error) {
	if err := validateTransportPlanID(planID); err != nil {
		return synctransport.DestinationApproval{}, err
	}
	if flags.value("approval-token-stdin") != "true" || flags.value("confirm") == "" {
		if cleanup {
			return synctransport.DestinationApproval{}, validationErrorf("GitHub transport cleanup requires --approval-token-stdin and --confirm destructive")
		}
		return synctransport.DestinationApproval{}, validationErrorf("approved GitHub transport ETL requires --approval-plan, --approval-token-stdin, and --confirm destructive")
	}
	confirmation, err := connectors.ParseWriteConfirmation(flags.value("confirm"))
	if err != nil {
		return synctransport.DestinationApproval{}, validationErrorf("invalid --confirm: %v", err)
	}
	if confirmation.Kind != connectors.ConfirmationKindDestructive {
		return synctransport.DestinationApproval{}, validationErrorf("GitHub transport requires --confirm destructive")
	}
	return synctransport.DestinationApproval{PlanID: planID, Confirmation: confirmation}, nil
}

func parseStrictTransportFlags(args []string, allowed map[string]strictTransportFlagSpec) (strictTransportFlags, error) {
	parsed := strictTransportFlags{values: make(map[string]string, len(allowed))}
	for index := 0; index < len(args); index++ {
		raw := args[index]
		if !strings.HasPrefix(raw, "--") || raw == "--" {
			return strictTransportFlags{}, usageErrorf("unexpected transport argument %q", raw)
		}
		keyValue := strings.TrimPrefix(raw, "--")
		key, value, equals := strings.Cut(keyValue, "=")
		spec, ok := allowed[key]
		if !ok {
			return strictTransportFlags{}, usageErrorf("unknown transport flag --%s", key)
		}
		if _, duplicate := parsed.values[key]; duplicate {
			return strictTransportFlags{}, usageErrorf("transport flag --%s may be supplied only once", key)
		}
		if equals {
			if spec.bareOnly {
				return strictTransportFlags{}, usageErrorf("--%s must be a bare stdin marker", key)
			}
			if strings.TrimSpace(value) == "" {
				return strictTransportFlags{}, validationErrorf("--%s must not be blank", key)
			}
			parsed.values[key] = value
			continue
		}
		if index+1 < len(args) && !strings.HasPrefix(args[index+1], "--") {
			if spec.bareOnly {
				return strictTransportFlags{}, usageErrorf("--%s must be a bare stdin marker", key)
			}
			value = args[index+1]
			index++
			if strings.TrimSpace(value) == "" {
				return strictTransportFlags{}, validationErrorf("--%s must not be blank", key)
			}
			parsed.values[key] = value
			continue
		}
		if !spec.allowBare {
			return strictTransportFlags{}, usageErrorf("--%s requires a value", key)
		}
		parsed.values[key] = "true"
	}
	return parsed, nil
}

func hasTransportApprovalInput(args []string) bool {
	for _, raw := range args {
		if !strings.HasPrefix(raw, "--") {
			continue
		}
		key, _, _ := strings.Cut(strings.TrimPrefix(raw, "--"), "=")
		switch key {
		case "approval-plan", "approval-token-stdin", "confirm", "approval-token", "approve", "method", "path", "url", "action", "record", "issue", "label", "destination-config":
			return true
		}
	}
	return false
}

func readApprovalTokenFromStdin(stdin io.Reader) (string, error) {
	if stdin == nil {
		return "", validationErrorf("approval token stdin is unavailable")
	}
	// Permit a full 4 KiB token plus its one LF/CRLF terminator, while reading
	// one further byte to distinguish that bounded line from an overlong input.
	contents, err := io.ReadAll(io.LimitReader(stdin, maxApprovalTokenStdinBytes+3))
	if err != nil {
		return "", validationErrorf("read approval token from stdin: %v", err)
	}
	if len(contents) == 0 || len(contents) > maxApprovalTokenStdinBytes+2 {
		return "", validationErrorf("approval token stdin must contain one bounded line")
	}
	line, rest, found := strings.Cut(string(contents), "\n")
	if !found || rest != "" {
		return "", validationErrorf("approval token stdin must contain exactly one line")
	}
	line = strings.TrimSuffix(line, "\r")
	if len(line) > maxApprovalTokenStdinBytes || strings.TrimSpace(line) == "" || strings.ContainsAny(line, "\r\n") {
		return "", validationErrorf("approval token stdin must contain a non-empty token line")
	}
	return line, nil
}

func githubIssueLabelTransportConnectionID(a *app.App, name string) (string, error) {
	if _, err := githubIssueLabelTransportConnectionIDFromName(name); err != nil {
		return "", err
	}
	for _, connection := range a.ListConnections() {
		if connection.Name == name {
			return connection.ID, nil
		}
	}
	return "", validationErrorf("connection %q not found", name)
}

func githubIssueLabelTransportConnectionIDFromName(name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", validationErrorf("missing --connection")
	}
	if err := safety.ValidateIdentifier(name, "connection"); err != nil {
		return "", validationErrorf("%v", err)
	}
	return name, nil
}

func validateTransportPlanID(planID string) error {
	if strings.TrimSpace(planID) == "" {
		return validationErrorf("transport plan ID is required")
	}
	if err := safety.ValidateIdentifier(planID, "transport plan"); err != nil {
		return validationErrorf("%v", err)
	}
	return nil
}

func safeETLTransportPlan(plan app.ReversePlan) etlTransportPlanOutput {
	return etlTransportPlanOutput{
		ID:            plan.ID,
		Status:        plan.Status,
		Mode:          plan.Mode,
		ConnectionID:  plan.TransportConnectionID,
		Action:        plan.Action,
		RecordCount:   plan.RecordCount,
		Confirmation:  plan.ConfirmationPolicy,
		CreatedAt:     plan.CreatedAt,
		ExpiresAt:     plan.ExpiresAt,
		ForwardPlanID: plan.TransportForwardPlanID,
	}
}

func writeETLTransportPlan(stdout io.Writer, jsonOut bool, plan app.ReversePlan) error {
	safe := safeETLTransportPlan(plan)
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "ETLTransportPlan", "approval_required": true, "plan": safe})
	}
	_, _ = fmt.Fprintf(stdout, "GitHub issue-label transport plan %s\n", safe.ID)
	_, _ = fmt.Fprintf(stdout, "Mode: %s\nAction: %s\nRecords: %d\n", safe.Mode, safe.Action, safe.RecordCount)
	_, _ = fmt.Fprintln(stdout, "Preview required before an approval token is issued.")
	_, err := fmt.Fprintln(stdout, "Confirmation required: --confirm destructive")
	return err
}

func writeETLTransportPreview(stdout io.Writer, jsonOut bool, plan app.ReversePlan, preview connectors.WritePreview) error {
	safePlan := safeETLTransportPlan(plan)
	safePreview := etlTransportWritePreviewOutput{
		RecordsStaged: preview.RecordsStaged,
		Action:        preview.Action,
		Digest:        preview.Digest,
	}
	if jsonOut {
		return writeJSON(stdout, envelope{
			"kind":                  "ETLTransportPreview",
			"approval_required":     true,
			"approval_token_issued": plan.ApprovalToken != "",
			"plan":                  safePlan,
			"write_preview":         safePreview,
		})
	}
	_, _ = fmt.Fprintf(stdout, "GitHub issue-label transport preview %s\n", safePlan.ID)
	_, _ = fmt.Fprintf(stdout, "Mode: %s\nAction: %s\nRecords staged: %d\n", safePlan.Mode, safePlan.Action, safePreview.RecordsStaged)
	if plan.ApprovalToken == "" {
		return fmt.Errorf("GitHub issue-label transport preview did not issue an approval token")
	}
	_, _ = fmt.Fprintf(stdout, "Approval token: %s\n", plan.ApprovalToken)
	_, err := fmt.Fprintln(stdout, "Confirmation required: --confirm destructive")
	return err
}

func writeETLTransportCleanupRun(stdout io.Writer, jsonOut bool, plan app.ReversePlan, result connectors.WriteResult) error {
	safe := safeETLTransportPlan(plan)
	if jsonOut {
		return writeJSON(stdout, envelope{
			"kind":            "ETLTransportCleanupRun",
			"status":          "completed",
			"plan_id":         safe.ID,
			"forward_plan_id": safe.ForwardPlanID,
			"connection_id":   safe.ConnectionID,
			"action":          safe.Action,
			"result":          result,
		})
	}
	_, err := fmt.Fprintf(stdout, "GitHub issue-label transport cleanup %s completed: written=%d failed=%d\n", safe.ID, result.RecordsWritten, result.RecordsFailed)
	return err
}

func writeETLTransportManual(stdout io.Writer, jsonOut bool, command string) error {
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "CommandManual", "command": command, "manual": etlTransportHelp})
	}
	_, err := fmt.Fprint(stdout, etlTransportHelp)
	return err
}

func isOnlyTransportHelp(args []string) bool {
	return len(args) == 1 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help")
}

// etlTransportManualCommand keeps the closed manual available before App.Open.
// A manual must not require a project or credentials merely to explain the
// stdin-only approval boundary.
func etlTransportManualCommand(args []string) (string, bool) {
	if len(args) == 0 || args[0] != "transport" {
		return "", false
	}
	if len(args) == 1 || isOnlyTransportHelp(args[1:]) {
		return "etl transport", true
	}
	if args[1] != githubIssueLabelTransportCommand {
		return "", false
	}
	if len(args) == 2 || isOnlyTransportHelp(args[2:]) {
		return "etl transport github-issue-label", true
	}
	if len(args) == 3 && args[2] == "cleanup" {
		return "etl transport github-issue-label", true
	}
	if len(args) == 3 && args[2] == "plan" {
		return "", false
	}
	if len(args) == 4 && (args[2] == "plan" || args[2] == "cleanup") && isHelpArg(args[3]) {
		return "etl transport github-issue-label", true
	}
	if len(args) >= 4 && args[2] == "preview" && isHelpArg(args[len(args)-1]) {
		return "etl transport github-issue-label", true
	}
	if len(args) >= 5 && args[2] == "cleanup" && (args[3] == "plan" || args[3] == "run") && isHelpArg(args[len(args)-1]) {
		return "etl transport github-issue-label", true
	}
	return "", false
}
