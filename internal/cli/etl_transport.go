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
	issueLabelTransportCommandSuffix                = "-issue-label"
	declarativeTypedDestinationTransportCommandName = "declarative-typed-destination"
	maxApprovalTokenStdinBytes                      = 4096
)

const etlTransportHelp = `NAME
  pm etl transport - run closed approval-bound transport lifecycles

SYNOPSIS
  pm etl transport %s plan --connection <name> [--json]
  pm etl transport %s preview <plan-id> [--json]
  pm etl transport postgres-managed-target plan --connection <name> --stream <stream> [--authorization-lifetime <24h..48h>] [--json]
  pm etl transport postgres-managed-target preview <plan-id> [--json]
  pm etl transport declarative-typed-destination plan --connection <name> --stream <stream> [--json]
  pm etl transport declarative-typed-destination preview <plan-id> [--json]
  pm etl run --connection <name> --stream <stream> --batch-size 1 [--max-in-flight-batches n] --approval-plan <plan-id> [--approval-token-stdin] --confirm destructive [--json]
  pm etl transport %s cleanup plan --connection <name> --forward-plan <plan-id> [--json]
  pm etl transport %s cleanup run <plan-id> --connection <name> --approval-token-stdin --confirm destructive [--json]

DESCRIPTION
  %s is a closed two-action %s label destination, not a generic writer. The
  connection owns the repository, source selection, target issue, label,
  action, and credential configuration. Its destination definition declares
  the admitted source executors, streams, and bounded record mappings. This
  command family accepts none of those command details.

  An input-fields source provides only the destination definition's
  target_issue and label inputs. full_append selects add_issue_labels;
  incremental_upsert selects set_issue_labels and requires
  transport_allow_keyed=true. The row-derived pair must match the plan-bound
  destination configuration, and null, malformed, mismatched, or tombstone
  rows stop before write I/O.

  Create a plan, preview it in human output to receive one ephemeral approval
  token, then send that token as one bounded line on standard input to pm etl
  run. The forward run remains the ordinary source -> durable warehouse ->
  reopen -> typed %s mutation and durable acknowledgement -> independent
  read-back -> checkpoint path.

  A successful non-additive run persists authorization for that exact plan
  scope. Later identical-scope runs reuse --approval-plan and --confirm
  destructive without --approval-token-stdin. A changed, expired, or revoked
  scope is refused before a provider write.

  A config-matched source selection and every independent read-back inspect
  only the first %s destination collection page. The transport fails instead
  of requesting another page when the configured source or target issue is not
  there.

  Cleanup is a separately planned, previewed, one-time approved typed inverse.
  A %s missing-label DELETE accepted by the declared missing_ok_status is an
  already-absent cleanup result, not a completed provider write; replaying a
  consumed cleanup approval is rejected before another provider request.

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

const postgresManagedTargetTransportHelp = `NAME
  pm etl transport postgres-managed-target - deliver sealed warehouse worksets to PostgreSQL

SYNOPSIS
  pm etl transport postgres-managed-target plan --connection <name> --stream <stream> [--authorization-lifetime <24h..48h>] [--json]
  pm etl transport postgres-managed-target preview <plan-id> [--json]
  pm etl run --connection <name> --stream <stream> --batch-size <n> [--max-in-flight-batches n] --approval-plan <plan-id> --approval-token-stdin --confirm destructive [--json]

DESCRIPTION
  This is a closed definition-selected source-to-PostgreSQL transport, not a
  generic SQL write surface. The saved connection owns both credentials, the
  source stream, primary key, mode, and immutable managed destination identity.
  Planning seals those bindings and the authoritative source schema before
  source records are read. PostgreSQL sources use their typed relation catalog;
  declared API sources use their JSON stream schema.

  The approved run reads through the registered source transport, stages
  bounded pages in the connection-owned warehouse, reopens the sealed Parquet
  workset, applies it through the managed PostgreSQL target, independently
  reads back the delivery receipt, and only then advances the source checkpoint.

  An incremental_upsert PostgreSQL source performs a gap-free logical-
  replication bootstrap when transport_bootstrap=true is set on its credential.
  After the snapshot barrier is durably delivered, committed pgoutput
  transactions follow the same warehouse and target acknowledgement path.

  --authorization-lifetime selects the durable authorization lifetime after
  the one-time preview token is consumed. It accepts 24h through 48h and
  defaults to 24h. It does not extend the preview token: every provider page
  fetch and staged destination apply remains independently deadline-bounded.

SECURITY
  The approval token is accepted only through --approval-token-stdin. It is
  single-use and binds the source schema plus both credential revisions. Stale,
  mismatched, replayed, authentication-refused, or permission-refused runs stop
  without advancing the checkpoint. PostgreSQL remains write=false in the
  public connector capability surface; this command does not expose raw SQL.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error
  3 validation error
`

const declarativeTypedDestinationTransportHelp = `NAME
  pm etl transport declarative-typed-destination - run a connector-declared typed reverse-ETL destination

SYNOPSIS
  pm etl transport declarative-typed-destination plan --connection <name> --stream <stream> [--json]
  pm etl transport declarative-typed-destination preview <plan-id> [--json]
  pm etl run --connection <name> --stream <stream> --batch-size <n> [--max-in-flight-batches n] --approval-plan <plan-id> [--approval-token-stdin] --confirm destructive [--json]

DESCRIPTION
  This command selects a destination only when the saved connection's
  destination descriptor declares declarative_typed_destination. The persisted
  stream owns the exact eligible writes.json action through destination_action;
  it is never an argument to this command or pm etl run.

  The connector definition owns the typed action, request method, route,
  record mapping, source-executor allowlist, acknowledgement, delivery
  contract, per-mode apply strategy, and conformance evidence. Shared Go only
  validates and dispatches that sealed declaration through the warehouse
  workset. It cannot write an arbitrary HTTP request.

  Planning and preview bind one connection, stream, destination_action,
  source mapping, credential revision, configuration digest, and strategy.
  The initial approval token becomes a revocable, bounded authorization for
  that exact shape. Each reopened workset obtains fresh typed write evidence;
  the generic orchestrator advances its checkpoint only after the declared
  durable acknowledgement and read-back.

SECURITY
  Approval tokens are accepted only through --approval-token-stdin. Callers
  cannot pass a connector, action, URL, method, body, mapping, or evidence.
  A missing declaration, foreign action, wrong source, malformed mapping,
  unavailable evidence, or unsupported mode is refused before provider I/O.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error
  3 validation error
`

type etlTransportPlanOutput struct {
	ID                      string                       `json:"id"`
	Status                  string                       `json:"status"`
	Mode                    string                       `json:"mode"`
	ConnectionID            string                       `json:"connection_id"`
	Action                  string                       `json:"action"`
	RecordCount             int                          `json:"record_count"`
	Confirmation            connectors.WriteConfirmation `json:"confirmation"`
	CreatedAt               time.Time                    `json:"created_at"`
	ExpiresAt               time.Time                    `json:"expires_at"`
	AuthorizationLifetime   time.Duration                `json:"authorization_lifetime_ns,omitempty"`
	ForwardPlanID           string                       `json:"forward_plan_id,omitempty"`
	TargetCopyWorkers       int                          `json:"target_copy_workers,omitempty"`
	TargetCopyWorkerMaximum int                          `json:"target_copy_worker_maximum,omitempty"`
}

type etlTransportWritePreviewOutput struct {
	RecordsStaged           int    `json:"records_staged"`
	Action                  string `json:"action"`
	Digest                  string `json:"digest"`
	TargetCopyWorkers       int    `json:"target_copy_workers,omitempty"`
	TargetCopyWorkerMaximum int    `json:"target_copy_worker_maximum,omitempty"`
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
	if args[0] == "postgres-managed-target" {
		return runPostgresManagedTargetTransport(ctx, a, args[1:], stdout, jsonOut)
	}
	if args[0] == declarativeTypedDestinationTransportCommandName {
		return runDeclarativeTypedDestinationTransport(ctx, a, args[1:], stdout, jsonOut)
	}
	connectorName, ok := parseIssueLabelTransportCommand(args[0])
	if !ok {
		return usageErrorf("unknown etl transport %q", args[0])
	}
	return runIssueLabelTransport(ctx, a, connectorName, args[1:], stdout, jsonOut)
}

func runDeclarativeTypedDestinationTransport(ctx context.Context, a *app.App, args []string, stdout io.Writer, jsonOut bool) error {
	command := "etl transport " + declarativeTypedDestinationTransportCommandName
	if len(args) == 0 || isOnlyTransportHelp(args) {
		return writeETLTransportManual(stdout, jsonOut, command)
	}
	switch args[0] {
	case "plan":
		if isOnlyTransportHelp(args[1:]) {
			return writeETLTransportManual(stdout, jsonOut, command)
		}
		flags, err := parseStrictTransportFlags(args[1:], map[string]strictTransportFlagSpec{"connection": {}, "stream": {}})
		if err != nil {
			return err
		}
		if _, err := issueLabelTransportConnectionIDFromName(flags.value("connection")); err != nil {
			return err
		}
		if strings.TrimSpace(flags.value("stream")) == "" {
			return validationErrorf("missing --stream")
		}
		plan, err := a.PlanDeclarativeTypedDestinationTransport(ctx, flags.value("connection"), flags.value("stream"))
		if err != nil {
			return err
		}
		return writeDeclarativeTypedDestinationTransportPlan(stdout, jsonOut, plan)
	case "preview":
		if len(args) != 2 {
			return errUsage
		}
		if err := validateTransportPlanID(args[1]); err != nil {
			return err
		}
		plan, preview, err := a.PreviewDeclarativeTypedDestinationTransport(ctx, args[1])
		if err != nil {
			return err
		}
		return writeDeclarativeTypedDestinationTransportPreview(stdout, jsonOut, plan, preview)
	default:
		return usageErrorf("unknown declarative typed destination transport command %q", args[0])
	}
}

func runPostgresManagedTargetTransport(ctx context.Context, a *app.App, args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 || isOnlyTransportHelp(args) {
		return writeETLTransportManual(stdout, jsonOut, "etl transport postgres-managed-target")
	}
	switch args[0] {
	case "plan":
		flags, err := parseStrictTransportFlags(args[1:], map[string]strictTransportFlagSpec{"connection": {}, "stream": {}, "authorization-lifetime": {}})
		if err != nil {
			return err
		}
		if _, err := issueLabelTransportConnectionIDFromName(flags.value("connection")); err != nil {
			return err
		}
		if strings.TrimSpace(flags.value("stream")) == "" {
			return validationErrorf("missing --stream")
		}
		var authorizationLifetime time.Duration
		if raw := strings.TrimSpace(flags.value("authorization-lifetime")); raw != "" {
			authorizationLifetime, err = time.ParseDuration(raw)
			if err != nil {
				return validationErrorf("invalid --authorization-lifetime %q: %v", raw, err)
			}
		}
		plan, err := a.PlanPostgresManagedTargetTransportWithAuthorizationLifetime(ctx, flags.value("connection"), flags.value("stream"), authorizationLifetime)
		if err != nil {
			return err
		}
		return writeETLTransportPlan(stdout, jsonOut, plan)
	case "preview":
		if len(args) != 2 {
			return errUsage
		}
		if err := validateTransportPlanID(args[1]); err != nil {
			return err
		}
		plan, preview, err := a.PreviewPostgresManagedTargetTransport(ctx, args[1])
		if err != nil {
			return err
		}
		return writeETLTransportPreview(stdout, jsonOut, plan, preview)
	default:
		return usageErrorf("unknown PostgreSQL managed target transport command %q", args[0])
	}
}

func runIssueLabelTransport(ctx context.Context, a *app.App, connectorName string, args []string, stdout io.Writer, jsonOut bool) error {
	command := "etl transport " + issueLabelTransportCommand(connectorName)
	if len(args) == 0 || isOnlyTransportHelp(args) {
		return writeETLTransportManual(stdout, jsonOut, command)
	}
	switch args[0] {
	case "plan":
		if isOnlyTransportHelp(args[1:]) {
			return writeETLTransportManual(stdout, jsonOut, command)
		}
		flags, err := parseStrictTransportFlags(args[1:], map[string]strictTransportFlagSpec{
			"connection": {},
		})
		if err != nil {
			return err
		}
		connectionID, err := issueLabelTransportConnectionID(a, connectorName, flags.value("connection"))
		if err != nil {
			return err
		}
		plan, err := a.PlanIssueLabelTransport(ctx, connectionID)
		if err != nil {
			return err
		}
		return writeETLTransportPlan(stdout, jsonOut, plan)
	case "preview":
		if len(args) == 3 && isOnlyTransportHelp(args[2:]) {
			return writeETLTransportManual(stdout, jsonOut, command)
		}
		if len(args) != 2 || isOnlyTransportHelp(args[1:]) {
			if isOnlyTransportHelp(args[1:]) {
				return writeETLTransportManual(stdout, jsonOut, command)
			}
			return errUsage
		}
		if err := validateTransportPlanID(args[1]); err != nil {
			return err
		}
		plan, preview, err := a.PreviewIssueLabelTransport(ctx, args[1])
		if err != nil {
			return err
		}
		if err := issueLabelTransportPlanUsesConnector(a, plan, connectorName); err != nil {
			return err
		}
		return writeETLTransportPreview(stdout, jsonOut, plan, preview)
	case "cleanup":
		return runIssueLabelTransportCleanup(ctx, a, connectorName, args[1:], stdout, jsonOut)
	default:
		return usageErrorf("unknown issue-label transport command %q", args[0])
	}
}

func runIssueLabelTransportCleanup(ctx context.Context, a *app.App, connectorName string, args []string, stdout io.Writer, jsonOut bool) error {
	command := "etl transport " + issueLabelTransportCommand(connectorName)
	if len(args) == 0 || isOnlyTransportHelp(args) {
		return writeETLTransportManual(stdout, jsonOut, command)
	}
	switch args[0] {
	case "plan":
		if isOnlyTransportHelp(args[1:]) {
			return writeETLTransportManual(stdout, jsonOut, command)
		}
		flags, err := parseStrictTransportFlags(args[1:], map[string]strictTransportFlagSpec{
			"connection":   {},
			"forward-plan": {},
		})
		if err != nil {
			return err
		}
		connectionID, err := issueLabelTransportConnectionID(a, connectorName, flags.value("connection"))
		if err != nil {
			return err
		}
		if err := validateTransportPlanID(flags.value("forward-plan")); err != nil {
			return err
		}
		plan, err := a.PlanIssueLabelTransportCleanup(ctx, connectionID, flags.value("forward-plan"))
		if err != nil {
			return err
		}
		return writeETLTransportPlan(stdout, jsonOut, plan)
	case "run":
		if len(args) < 2 || isOnlyTransportHelp(args[1:]) || isOnlyTransportHelp(args[2:]) {
			if isOnlyTransportHelp(args[1:]) || isOnlyTransportHelp(args[2:]) {
				return writeETLTransportManual(stdout, jsonOut, command)
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
		connectionID, err := issueLabelTransportConnectionID(a, connectorName, flags.value("connection"))
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
		if err := issueLabelTransportPlanUsesConnector(a, plan, connectorName); err != nil {
			return err
		}
		result, err := a.ApplyIssueLabelTransportCleanup(ctx, connectionID, approval)
		if err != nil {
			return err
		}
		return writeETLTransportCleanupRun(stdout, jsonOut, plan, result)
	default:
		return usageErrorf("unknown issue-label transport cleanup command %q", args[0])
	}
}

func parseETLRunTransportApproval(args []string, stdin io.Reader) (synctransport.DestinationApproval, bool, strictTransportFlags, error) {
	if !hasTransportApprovalInput(args) {
		return synctransport.DestinationApproval{}, false, strictTransportFlags{}, nil
	}
	flags, err := parseStrictTransportFlags(args, map[string]strictTransportFlagSpec{
		"connection":            {},
		"stream":                {},
		"batch-size":            {},
		"max-in-flight-batches": {},
		"approval-plan":         {},
		"approval-token-stdin":  {allowBare: true, bareOnly: true},
		"confirm":               {},
	})
	if err != nil {
		return synctransport.DestinationApproval{}, true, strictTransportFlags{}, err
	}
	if flags.value("connection") == "" || flags.value("stream") == "" || flags.value("batch-size") == "" {
		return synctransport.DestinationApproval{}, true, strictTransportFlags{}, validationErrorf("approved transport ETL requires --connection, --stream, and --batch-size")
	}
	batchSize, err := parseIntFlag("batch-size", flags.value("batch-size"), 0)
	if err != nil {
		return synctransport.DestinationApproval{}, true, strictTransportFlags{}, err
	}
	if batchSize <= 0 {
		return synctransport.DestinationApproval{}, true, strictTransportFlags{}, validationErrorf("approved transport ETL requires a positive --batch-size")
	}
	if _, err := parseMaxInFlightBatches(flags.value("max-in-flight-batches")); err != nil {
		return synctransport.DestinationApproval{}, true, strictTransportFlags{}, err
	}
	if _, err := issueLabelTransportConnectionIDFromName(flags.value("connection")); err != nil {
		return synctransport.DestinationApproval{}, true, strictTransportFlags{}, err
	}
	approval, err := approvalFromStrictTransportFlags(flags.value("approval-plan"), flags, false)
	if err != nil {
		return synctransport.DestinationApproval{}, true, strictTransportFlags{}, err
	}
	if flags.value("approval-token-stdin") == "true" {
		token, err := readApprovalTokenFromStdin(stdin)
		if err != nil {
			return synctransport.DestinationApproval{}, true, strictTransportFlags{}, err
		}
		approval.ApprovalToken = token
	}
	return approval, true, flags, nil
}

// runApprovedTransportETL carries an approval only to the persisted App route.
// It does not select a destination action; App preflight resolves that from the
// saved stream descriptor before any source or provider I/O.
func runApprovedTransportETL(ctx context.Context, a *app.App, flags strictTransportFlags, approval synctransport.DestinationApproval, stdout io.Writer, jsonOut bool) error {
	batchSize, err := parseIntFlag("batch-size", flags.value("batch-size"), 0)
	if err != nil {
		return err
	}
	maxInFlightBatches, err := parseMaxInFlightBatches(flags.value("max-in-flight-batches"))
	if err != nil {
		return err
	}
	run, err := a.RunETL(ctx, app.RunETLRequest{
		Connection:          flags.value("connection"),
		Stream:              flags.value("stream"),
		BatchSize:           batchSize,
		MaxInFlightBatches:  maxInFlightBatches,
		DestinationApproval: approval,
	})
	if err != nil && run.ID == "" {
		return err
	}
	if jsonOut {
		if outputErr := writeJSON(stdout, envelope{"kind": "ETLRun", "run": run, "runtime_recorded": false}); outputErr != nil {
			return outputErr
		}
		if err != nil {
			return alreadyReportedExecutionError(err)
		}
		return nil
	}
	if err != nil {
		_, _ = fmt.Fprintf(stdout, "ETL run %s ended with status %s; inspect it with pm etl status %s\n", run.ID, run.Status, run.ID)
		return alreadyReportedExecutionError(err)
	}
	_, _ = fmt.Fprintf(stdout, "ETL run %s completed: read=%d loaded=%d failed=%d\n", run.ID, run.RecordsRead, run.RecordsLoaded, run.RecordsFailed)
	return nil
}

func approvalFromStrictTransportFlags(planID string, flags strictTransportFlags, cleanup bool) (synctransport.DestinationApproval, error) {
	if err := validateTransportPlanID(planID); err != nil {
		return synctransport.DestinationApproval{}, err
	}
	if cleanup && flags.value("approval-token-stdin") != "true" {
		return synctransport.DestinationApproval{}, validationErrorf("issue-label transport cleanup requires --approval-token-stdin and --confirm destructive")
	}
	if flags.value("confirm") == "" {
		if cleanup {
			return synctransport.DestinationApproval{}, validationErrorf("issue-label transport cleanup requires --approval-token-stdin and --confirm destructive")
		}
		return synctransport.DestinationApproval{}, validationErrorf("approved transport ETL requires --approval-plan and --confirm destructive")
	}
	confirmation, err := connectors.ParseWriteConfirmation(flags.value("confirm"))
	if err != nil {
		return synctransport.DestinationApproval{}, validationErrorf("invalid --confirm: %v", err)
	}
	if confirmation.Kind != connectors.ConfirmationKindDestructive {
		return synctransport.DestinationApproval{}, validationErrorf("approved transport requires --confirm destructive")
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
		case "approval-plan", "approval-token-stdin", "confirm", "approval-token", "approve", "connector", "action", "destination-action", "destination_action", "method", "path", "url", "route", "verb", "body", "payload", "query", "headers", "header", "mapping", "map", "input-fields", "input_fields", "record", "issue", "label", "evidence", "destination-config", "destination_config", "destination", "credential", "source-config", "source_config", "sql", "shell", "http":
			return true
		}
	}
	return false
}

func validateLegacyETLRunFlags(args []string) error {
	_, err := parseStrictTransportFlags(args, map[string]strictTransportFlagSpec{
		"connection":            {},
		"stream":                {},
		"batch-size":            {},
		"max-in-flight-batches": {},
		"runtime":               {allowBare: true},
	})
	return err
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
	if len(contents) == 0 {
		return "", validationErrorf("approval token stdin must contain one bounded line")
	}
	if len(contents) > maxApprovalTokenStdinBytes+2 {
		return "", validationErrorf("approval token stdin is too large")
	}
	line, rest, found := strings.Cut(string(contents), "\n")
	if !found || rest != "" {
		return "", validationErrorf("approval token stdin must contain exactly one line")
	}
	line = strings.TrimSuffix(line, "\r")
	if len(line) > maxApprovalTokenStdinBytes {
		return "", validationErrorf("approval token stdin is too large")
	}
	if strings.TrimSpace(line) == "" || strings.ContainsAny(line, "\r\n") {
		return "", validationErrorf("approval token stdin must contain a non-empty token line")
	}
	return line, nil
}

func issueLabelTransportConnectionID(a *app.App, connectorName, name string) (string, error) {
	if _, err := issueLabelTransportConnectionIDFromName(name); err != nil {
		return "", err
	}
	for _, connection := range a.ListConnections() {
		if connection.Name == name {
			// App owns exact source admission from the registered transport
			// descriptor. The CLI command selects the definition-owned destination
			// only, so every admitted source reaches the same app-level closed
			// validation rather than being rejected by a stale shortcut.
			if connection.Destination.Connector != connectorName {
				return "", validationErrorf("connection %q does not own the selected issue-label transport command", name)
			}
			return connection.ID, nil
		}
	}
	return "", validationErrorf("connection %q not found", name)
}

func issueLabelTransportPlanUsesConnector(a *app.App, plan app.ReversePlan, connectorName string) error {
	for _, connection := range a.ListConnections() {
		if connection.ID != plan.TransportConnectionID {
			continue
		}
		if connection.Destination.Connector != connectorName || plan.DestinationConnector != connectorName {
			return validationErrorf("transport plan %q does not belong to the selected issue-label transport command", plan.ID)
		}
		return nil
	}
	return validationErrorf("transport plan %q has no connection-owned command route", plan.ID)
}

func issueLabelTransportConnectionIDFromName(name string) (string, error) {
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
		ID:                      plan.ID,
		Status:                  plan.Status,
		Mode:                    plan.Mode,
		ConnectionID:            plan.TransportConnectionID,
		Action:                  plan.Action,
		RecordCount:             plan.RecordCount,
		Confirmation:            plan.ConfirmationPolicy,
		CreatedAt:               plan.CreatedAt,
		ExpiresAt:               plan.ExpiresAt,
		AuthorizationLifetime:   plan.AuthorizationLifetime,
		ForwardPlanID:           plan.TransportForwardPlanID,
		TargetCopyWorkers:       plan.TargetCopyWorkers,
		TargetCopyWorkerMaximum: plan.TargetCopyWorkerMaximum,
	}
}

func writeETLTransportPlan(stdout io.Writer, jsonOut bool, plan app.ReversePlan) error {
	safe := safeETLTransportPlan(plan)
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "ETLTransportPlan", "approval_required": true, "plan": safe})
	}
	_, _ = fmt.Fprintf(stdout, "Issue-label transport plan %s\n", safe.ID)
	_, _ = fmt.Fprintf(stdout, "Mode: %s\nAction: %s\nRecords: %d\n", safe.Mode, safe.Action, safe.RecordCount)
	if safe.TargetCopyWorkers > 0 {
		_, _ = fmt.Fprintf(stdout, "Target COPY workers: %d (declared pool maximum %d)\n", safe.TargetCopyWorkers, safe.TargetCopyWorkerMaximum)
	}
	_, _ = fmt.Fprintln(stdout, "Preview required before an approval token is issued.")
	_, err := fmt.Fprintln(stdout, "Confirmation required: --confirm destructive")
	return err
}

func writeETLTransportPreview(stdout io.Writer, jsonOut bool, plan app.ReversePlan, preview connectors.WritePreview) error {
	safePlan := safeETLTransportPlan(plan)
	safePreview := etlTransportWritePreviewOutput{
		RecordsStaged:           preview.RecordsStaged,
		Action:                  preview.Action,
		Digest:                  preview.Digest,
		TargetCopyWorkers:       safePlan.TargetCopyWorkers,
		TargetCopyWorkerMaximum: safePlan.TargetCopyWorkerMaximum,
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
	_, _ = fmt.Fprintf(stdout, "Issue-label transport preview %s\n", safePlan.ID)
	_, _ = fmt.Fprintf(stdout, "Mode: %s\nAction: %s\nRecords staged: %d\n", safePlan.Mode, safePlan.Action, safePreview.RecordsStaged)
	if safePreview.TargetCopyWorkers > 0 {
		_, _ = fmt.Fprintf(stdout, "Target COPY workers: %d (declared pool maximum %d)\n", safePreview.TargetCopyWorkers, safePreview.TargetCopyWorkerMaximum)
	}
	if plan.ApprovalToken == "" {
		return fmt.Errorf("issue-label transport preview did not issue an approval token")
	}
	_, _ = fmt.Fprintf(stdout, "Approval token: %s\n", plan.ApprovalToken)
	_, err := fmt.Fprintln(stdout, "Confirmation required: --confirm destructive")
	return err
}

func writeDeclarativeTypedDestinationTransportPlan(stdout io.Writer, jsonOut bool, plan app.ReversePlan) error {
	safe := safeETLTransportPlan(plan)
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "DeclarativeTypedDestinationTransportPlan", "approval_required": true, "plan": safe})
	}
	_, _ = fmt.Fprintf(stdout, "Declarative typed destination transport plan %s\n", safe.ID)
	_, _ = fmt.Fprintf(stdout, "Mode: %s\nAction: %s\nRecords: %d\n", safe.Mode, safe.Action, safe.RecordCount)
	_, _ = fmt.Fprintln(stdout, "Preview required before an approval token is issued.")
	_, err := fmt.Fprintln(stdout, "Confirmation required: --confirm destructive")
	return err
}

func writeDeclarativeTypedDestinationTransportPreview(stdout io.Writer, jsonOut bool, plan app.ReversePlan, preview connectors.WritePreview) error {
	safePlan := safeETLTransportPlan(plan)
	safePreview := etlTransportWritePreviewOutput{RecordsStaged: preview.RecordsStaged, Action: preview.Action, Digest: preview.Digest}
	if jsonOut {
		return writeJSON(stdout, envelope{
			"kind":                  "DeclarativeTypedDestinationTransportPreview",
			"approval_required":     true,
			"approval_token_issued": plan.ApprovalToken != "",
			"plan":                  safePlan,
			"write_preview":         safePreview,
		})
	}
	_, _ = fmt.Fprintf(stdout, "Declarative typed destination transport preview %s\n", safePlan.ID)
	_, _ = fmt.Fprintf(stdout, "Mode: %s\nAction: %s\nRecords staged: %d\n", safePlan.Mode, safePlan.Action, safePreview.RecordsStaged)
	if plan.ApprovalToken == "" {
		return fmt.Errorf("declarative typed destination transport preview did not issue an approval token")
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
	_, err := fmt.Fprintf(stdout, "Issue-label transport cleanup %s completed: written=%d unchanged=%d failed=%d\n", safe.ID, result.RecordsWritten, result.RecordsUnchanged, result.RecordsFailed)
	return err
}

func writeETLTransportManual(stdout io.Writer, jsonOut bool, command string) error {
	manual, err := etlTransportManual(command)
	if err != nil {
		return err
	}
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "CommandManual", "command": command, "manual": manual})
	}
	_, err = fmt.Fprint(stdout, manual)
	return err
}

func etlTransportManual(command string) (string, error) {
	if command == "etl transport postgres-managed-target" {
		return postgresManagedTargetTransportHelp, nil
	}
	if command == "etl transport "+declarativeTypedDestinationTransportCommandName {
		return declarativeTypedDestinationTransportHelp, nil
	}
	identity, err := app.DefaultIssueLabelTransportIdentity()
	if err != nil {
		return "", fmt.Errorf("resolve closed issue-label transport manual: %w", err)
	}
	name := issueLabelTransportCommand(identity.ConnectorName)
	parts := strings.Fields(command)
	if len(parts) == 3 {
		name = parts[2]
	}
	if name != issueLabelTransportCommand(identity.ConnectorName) {
		return "", usageErrorf("unknown etl transport %q", name)
	}
	return fmt.Sprintf(etlTransportHelp, name, name, name, name, name, identity.DisplayName, identity.DisplayName, identity.DisplayName, identity.DisplayName), nil
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
	if args[1] == "postgres-managed-target" {
		command := "etl transport postgres-managed-target"
		if len(args) == 2 || isOnlyTransportHelp(args[2:]) {
			return command, true
		}
		if len(args) == 4 && (args[2] == "plan" || args[2] == "preview") && isHelpArg(args[3]) {
			return command, true
		}
		if len(args) >= 5 && args[2] == "preview" && isHelpArg(args[len(args)-1]) {
			return command, true
		}
		return "", false
	}
	if args[1] == declarativeTypedDestinationTransportCommandName {
		command := "etl transport " + declarativeTypedDestinationTransportCommandName
		if len(args) == 2 || isOnlyTransportHelp(args[2:]) {
			return command, true
		}
		if len(args) == 4 && (args[2] == "plan" || args[2] == "preview") && isHelpArg(args[3]) {
			return command, true
		}
		if len(args) >= 5 && args[2] == "preview" && isHelpArg(args[len(args)-1]) {
			return command, true
		}
		return "", false
	}
	connectorName, ok := parseIssueLabelTransportCommand(args[1])
	if !ok {
		return "", false
	}
	command := "etl transport " + issueLabelTransportCommand(connectorName)
	if len(args) == 2 || isOnlyTransportHelp(args[2:]) {
		return command, true
	}
	if len(args) == 3 && args[2] == "cleanup" {
		return command, true
	}
	if len(args) == 3 && args[2] == "plan" {
		return "", false
	}
	if len(args) == 4 && (args[2] == "plan" || args[2] == "cleanup") && isHelpArg(args[3]) {
		return command, true
	}
	if len(args) >= 4 && args[2] == "preview" && isHelpArg(args[len(args)-1]) {
		return command, true
	}
	if len(args) >= 5 && args[2] == "cleanup" && (args[3] == "plan" || args[3] == "run") && isHelpArg(args[len(args)-1]) {
		return command, true
	}
	return "", false
}

func issueLabelTransportCommand(connectorName string) string {
	return connectorName + issueLabelTransportCommandSuffix
}

func parseIssueLabelTransportCommand(value string) (string, bool) {
	identity, err := app.DefaultIssueLabelTransportIdentity()
	if err != nil || value != issueLabelTransportCommand(identity.ConnectorName) {
		return "", false
	}
	if safety.ValidateIdentifier(identity.ConnectorName, "connector") != nil {
		return "", false
	}
	return identity.ConnectorName, true
}
