package commandrunner

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/failures"
	"polymetrics.ai/internal/safety"
)

const (
	MaxDirectReadBytes          = 1 << 20
	MaxOperationDirectReadBytes = 16 << 20
	maxRecordPathArrayIndex     = 128
	maxStructuredJSONFlagBytes  = 1 << 20
)

type Request struct {
	Path     []string
	Flags    map[string][]string
	Config   connectors.RuntimeConfig
	Limit    int
	MaxBytes int
	// Page and PageCursor navigate a direct read's declared pagination. A
	// direct read returns ONE page; these say which one. They are ignored by
	// every other intent.
	Page       int
	PageCursor string
	Preview    bool
	// DestRoot is the directory a binary_download command writes beneath.
	// Required for that intent and ignored by every other one.
	DestRoot string
	// FileName optionally names the downloaded file within DestRoot.
	FileName string
	// PlanContinuation permits already-planned required values to remain
	// withheld while previewing or executing a persisted reverse plan. It does
	// not relax unknown-flag, type, enum, format, size, range, cardinality, or
	// effective-config validation, all of which still run before state lookup.
	PlanContinuation bool
}

type Result struct {
	Connector      string                                    `json:"connector"`
	Command        string                                    `json:"command"`
	Stream         string                                    `json:"stream,omitempty"`
	Count          int                                       `json:"count,omitempty"`
	DirectRead     *connectors.DirectReadResult              `json:"direct_read,omitempty"`
	BinaryDownload *connectors.OperationBinaryDownloadResult `json:"binary_download,omitempty"`
	StatusCheck    *connectors.OperationStatusCheckResult    `json:"status_check,omitempty"`
}

type WriteCommand struct {
	Connector string `json:"connector"`
	Command   string `json:"command"`
	Intent    string `json:"intent"`
	Write     string `json:"write"`
	// Operation is set only for a typed direct_write command. Write remains
	// the plan action for backward-compatible reverse_etl command plans.
	Operation             string                   `json:"operation,omitempty"`
	MutationClass         string                   `json:"mutation_class"`
	TargetResource        string                   `json:"target_resource"`
	ApprovalRequired      bool                     `json:"approval_required"`
	Risk                  string                   `json:"risk,omitempty"`
	Approval              string                   `json:"approval,omitempty"`
	ConfirmationChallenge string                   `json:"confirmation_challenge,omitempty"`
	Record                connectors.Record        `json:"record,omitempty"`
	RedactedRecord        connectors.Record        `json:"redacted_record,omitempty"`
	PathParams            map[string]string        `json:"path_params,omitempty"`
	Query                 map[string]string        `json:"query,omitempty"`
	Headers               map[string]string        `json:"headers,omitempty"`
	HeaderValues          map[string][]string      `json:"header_values,omitempty"`
	Batchable             bool                     `json:"batchable"`
	Preview               *connectors.WritePreview `json:"preview,omitempty"`
}

var ErrNotWriteCommand = errors.New("connector command is not a reverse ETL write command")

// declarativeWritePreflighter is implemented by the declarative engine. It
// lets runtime preflight reject a write schema that the engine could not
// faithfully expose as typed CLI flags, without giving native connectors a
// second, hand-maintained schema contract.
type declarativeWritePreflighter interface {
	PreflightWriteAction(name string) error
}

// binaryUploadActionPreflighter is intentionally more specific than the
// ordinary write preflight. A public binary_upload command may expose only a
// declaration-owned, bounded file source with an explicit media policy; it is
// never a route to an arbitrary request body or URL.
type binaryUploadActionPreflighter interface {
	PreflightBinaryUploadAction(name string) ([]connectors.BinaryUploadSource, error)
}

// deferredCommandPreflighter is the declaration-runtime seam for a command
// that is intentionally not executable yet. It proves an honest exact target
// before commandrunner returns missing_foundation, without resolving a
// credential or making a provider request.
type deferredCommandPreflighter interface {
	PreflightDeferredCommand(connectors.CommandSurfaceCommand) error
}

// implementedCommandPreflighter is the engine-owned, lane-complete binding
// resolver shared with declaration admission.
type implementedCommandPreflighter interface {
	PreflightImplementedCommand(connectors.CommandSurfaceCommand) error
}

// structuredJSONRecordPreflighter is deliberately narrower than a generic
// JSON parser. The engine owns the raw record schema, so it alone decides
// whether a named, top-level record field may accept one structured JSON
// value. Native connectors do not opt into this declarative flag type merely
// by exposing a similarly named command.
type structuredJSONRecordPreflighter interface {
	PreflightStructuredJSONRecordField(actionName, field string) error
}

// structuredJSONRecordStringArmPreflighter is an even narrower opt-in for a
// source-declared string arm of a named multi-kind record field. It cannot
// authorize a raw body, open object, or a direct operation input.
type structuredJSONRecordStringArmPreflighter interface {
	PreflightStructuredJSONRecordStringArm(actionName, field string) error
}

// structuredJSONOperationBodyPreflighter is intentionally an operation
// contract rather than a generic JSON parser. The engine owns the source
// schema, body mapping, and recursive bounds for the named field.
type structuredJSONOperationBodyPreflighter interface {
	PreflightOperationStructuredJSONBodyField(operation, field string) error
}

type BlockedCommandError struct {
	Connector    string
	Command      string
	Intent       string
	Availability string
	Reason       string
	// Failure carries a typed, safe classification for a dispatch refusal when
	// the dispatch layer can identify one. It is optional because many existing
	// preflight refusals are not one of the closed #3991 dispatch terminals.
	Failure *failures.Classification
}

type MinimumFlagError struct {
	Parameter string
	Minimum   connectors.ExactNumber
}

type MaximumFlagError struct {
	Parameter string
	Maximum   connectors.ExactNumber
}

// ConfiguredFlagValueError preserves the declaration and config key that
// failed without exposing the configured value. The underlying typed error is
// available to callers through Unwrap for classification, but Error remains a
// safe public message for pre-credential CLI validation.
type ConfiguredFlagValueError struct {
	Command string
	Flag    string
	Target  string
	err     error
}

func (e *ConfiguredFlagValueError) Error() string {
	return fmt.Sprintf("configured value for config.%s (--%s) is invalid for command %q", e.Target, e.Flag, e.Command)
}

func (e *ConfiguredFlagValueError) Unwrap() error {
	return e.err
}

func (e *MaximumFlagError) Error() string {
	return fmt.Sprintf("invalid --%s: value must be at most %s", e.Parameter, e.Maximum.String())
}

func (e *MinimumFlagError) Error() string {
	return fmt.Sprintf("invalid --%s: value must be at least %s", e.Parameter, e.Minimum.String())
}

// MissingRequiredFlagError is a caller-correctable command-input refusal.
// Its fields let the CLI preserve the usage-error category without parsing an
// error string, and commandrunner returns it before any executor is called.
type MissingRequiredFlagError struct {
	Command string
	Flag    string
}

func (e *MissingRequiredFlagError) Error() string {
	return fmt.Sprintf("missing required flag --%s for command %q", e.Flag, e.Command)
}

func (e *BlockedCommandError) Error() string {
	parts := []string{fmt.Sprintf("connector command %q is blocked", e.Command)}
	if e.Intent != "" {
		parts = append(parts, "intent="+e.Intent)
	}
	if e.Availability != "" {
		parts = append(parts, "availability="+e.Availability)
	}
	reason := e.Reason
	if reason == "" && e.Failure != nil {
		reason = e.Failure.Error()
	}
	if reason != "" {
		parts = append(parts, reason)
	}
	return strings.Join(parts, ": ")
}

// Unwrap exposes an optional typed dispatch classification to callers without
// requiring them to parse the blocked-command text.
func (e *BlockedCommandError) Unwrap() error {
	if e == nil || e.Failure == nil {
		return nil
	}
	return e.Failure
}

func Preflight(connector connectors.Connector, path []string) error {
	_, _, err := resolvePreflightCommand(connector, path)
	return err
}

// PreflightSourceBoundOrigin refuses a caller-selected origin before the App
// resolves credentials or constructs authentication state. It is deliberately
// narrower than generic configuration validation: only a declared
// source_operation uses the engine's closed origin equivalence.
func PreflightSourceBoundOrigin(connector connectors.Connector, path []string, config map[string]string) error {
	cmd, command, err := resolvePreflightCommand(connector, path)
	if err != nil {
		return err
	}
	if strings.TrimSpace(cmd.SourceOperation) == "" {
		return nil
	}
	preflighter, ok := connector.(connectors.SourceBoundOriginPreflighter)
	if !ok {
		return &BlockedCommandError{Connector: connector.Name(), Command: command, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "source-bound command does not expose declared origin preflight"}
	}
	cfg := connectors.RuntimeConfig{Config: config}
	switch cmd.Intent {
	case "direct_read":
		if strings.TrimSpace(cmd.Operation) == "" {
			return &BlockedCommandError{Connector: connector.Name(), Command: command, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "source-bound direct read has no operation"}
		}
		return preflighter.PreflightSourceBoundOperationOrigin(cmd.Operation, cfg)
	case "etl":
		if strings.TrimSpace(cmd.Stream) == "" {
			return &BlockedCommandError{Connector: connector.Name(), Command: command, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "source-bound ETL read has no stream"}
		}
		return preflighter.PreflightSourceBoundStreamOrigin(cmd.Stream, cfg)
	default:
		return &BlockedCommandError{Connector: connector.Name(), Command: command, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "source-bound origin preflight is available only for read commands"}
	}
}

// PreflightRequest validates one public command invocation without credentials
// and without calling an executor. It resolves the same executable command as
// Preflight and shares the exact flag coercion/requiredness rules used while
// constructing runtime requests.
func PreflightRequest(connector connectors.Connector, req Request) error {
	cmd, _, err := resolvePreflightCommand(connector, req.Path)
	if err != nil {
		return err
	}
	flags := req.Flags
	if cmd.Intent == "direct_read" {
		flags = withoutDirectReadPageFlags(flags)
	}
	if _, err := validateCommandFlagSetForPreflight(cmd, flags, req.PlanContinuation); err != nil {
		return err
	}
	return validateConfiguredFlagValues(cmd, req.Config, flags)
}

func BuildWriteCommand(ctx context.Context, connector connectors.Connector, req Request) (WriteCommand, error) {
	cmd, command, err := resolvePreflightCommand(connector, req.Path)
	if err != nil {
		return WriteCommand{}, err
	}
	if cmd.Intent == "direct_write" {
		return buildOperationDirectWriteCommand(ctx, connector, cmd, command, req)
	}
	if cmd.Intent != "reverse_etl" && cmd.Intent != "binary_upload" {
		return WriteCommand{}, ErrNotWriteCommand
	}
	if cmd.Availability != "implemented" || cmd.Write == "" {
		return WriteCommand{}, &BlockedCommandError{
			Connector:    connector.Name(),
			Command:      command,
			Intent:       cmd.Intent,
			Availability: cmd.Availability,
			Reason:       "implemented reverse ETL and binary_upload commands must reference a write action",
		}
	}
	action, ok := findWriteAction(connectors.ManifestOf(connector), cmd.Write)
	if !ok {
		return WriteCommand{}, &BlockedCommandError{
			Connector:    connector.Name(),
			Command:      command,
			Intent:       cmd.Intent,
			Availability: cmd.Availability,
			Reason:       fmt.Sprintf("write action %q is not declared in connector manifest", cmd.Write),
		}
	}
	record, err := recordOverrides(cmd, req.Flags)
	if err != nil {
		return WriteCommand{}, err
	}
	writeReq := connectors.WriteRequest{Action: cmd.Write, Config: req.Config}
	records := []connectors.Record{record}
	if validator, ok := connector.(connectors.WriteValidator); ok {
		if err := validator.ValidateWrite(ctx, writeReq, records); err != nil {
			return WriteCommand{}, err
		}
	}
	out := WriteCommand{
		Connector:             connector.Name(),
		Command:               command,
		Intent:                cmd.Intent,
		Write:                 cmd.Write,
		MutationClass:         mutationClassOf(action),
		TargetResource:        targetResourceOf(cmd),
		ApprovalRequired:      true,
		Risk:                  firstNonEmpty(cmd.Risk, action.Risk),
		Approval:              firstNonEmpty(cmd.Approval, writeCommandApproval(cmd.Intent)),
		ConfirmationChallenge: string(connectors.ConfirmationForWriteAction(action).Kind),
		Record:                cloneRecord(record),
		RedactedRecord:        cloneRecord(record),
		Batchable:             action.IsBatchable(),
	}
	if req.Preview {
		dryRunner, ok := connector.(connectors.DryRunWriter)
		if !ok {
			return WriteCommand{}, &BlockedCommandError{
				Connector:    connector.Name(),
				Command:      command,
				Intent:       cmd.Intent,
				Availability: cmd.Availability,
				Reason:       "connector does not support reverse ETL previews",
			}
		}
		preview, err := dryRunner.DryRunWrite(ctx, writeReq, records)
		if err != nil {
			return WriteCommand{}, err
		}
		out.Preview = &preview
	}
	return out, nil
}

func writeCommandApproval(intent string) string {
	if intent == "binary_upload" {
		return "binary uploads require plan, preview, approval, execute"
	}
	return "reverse ETL writes require plan, preview, approval, execute"
}

// buildOperationDirectWriteCommand shapes a declared direct_write command for
// the existing connector-command plan lifecycle. It does not execute or
// preview the write: those actions belong to App so the plan, preview digest,
// typed confirmation, and single-use grant cannot be bypassed by a direct
// commandrunner call.
func buildOperationDirectWriteCommand(ctx context.Context, connector connectors.Connector, cmd connectors.CommandSurfaceCommand, command string, req Request) (WriteCommand, error) {
	if cmd.Availability != "implemented" || strings.TrimSpace(cmd.Operation) == "" {
		return WriteCommand{}, &BlockedCommandError{Connector: connector.Name(), Command: command, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "implemented direct_write commands must reference an operation"}
	}
	if err := validateOperationDirectWriteCommand(connector, cmd); err != nil {
		return WriteCommand{}, err
	}
	bodyMaterializer, ok := connector.(connectors.OperationDirectWriteBodyMaterializer)
	if !ok {
		return WriteCommand{}, &BlockedCommandError{Connector: connector.Name(), Command: command, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "connector does not materialize declared direct-write bodies"}
	}
	bodyValueResolver, ok := connector.(connectors.OperationDirectWriteBodyValueResolver)
	if !ok {
		return WriteCommand{}, &BlockedCommandError{Connector: connector.Name(), Command: command, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "connector does not resolve declared direct-write body values"}
	}
	pathParams, query, headers, headerValues, body, _, err := operationDirectWriteOverrides(cmd, req.Flags, bodyMaterializer.MaterializeOperationDirectWriteBody)
	if err != nil {
		return WriteCommand{}, err
	}
	if err := validateCommandInputs(cmd, req.Config, mappedCommandInputs{
		Query: query,
		Body:  body,
		BodyValue: func(path string) (any, bool, error) {
			return bodyValueResolver.ResolveOperationDirectWriteBodyValue(cmd.Operation, body, path)
		},
	}); err != nil {
		return WriteCommand{}, err
	}
	metadata, err := connector.(connectors.OperationDirectWriteMetadataProvider).OperationDirectWriteMetadata(cmd.Operation)
	if err != nil {
		return WriteCommand{}, err
	}
	if metadata.Operation != cmd.Operation {
		return WriteCommand{}, &BlockedCommandError{Connector: connector.Name(), Command: command, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "operation direct_write metadata did not match command operation"}
	}
	if metadata.OutputPolicy != cmd.OutputPolicy {
		return WriteCommand{}, &BlockedCommandError{Connector: connector.Name(), Command: command, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "direct_write command output_policy does not match declared operation"}
	}
	record := connectors.Record(body)
	return WriteCommand{
		Connector:             connector.Name(),
		Command:               command,
		Intent:                cmd.Intent,
		Write:                 cmd.Operation,
		Operation:             cmd.Operation,
		MutationClass:         metadata.MutationClass,
		TargetResource:        targetResourceOf(cmd),
		ApprovalRequired:      true,
		Risk:                  firstNonEmpty(cmd.Risk, metadata.Risk),
		Approval:              firstNonEmpty(cmd.Approval, metadata.Approval, "direct writes require plan, preview, approval, execute"),
		ConfirmationChallenge: metadata.ConfirmationChallenge,
		Record:                cloneRecord(record),
		RedactedRecord:        cloneRecord(record),
		PathParams:            cloneStringMap(pathParams),
		Query:                 cloneStringMap(query),
		Headers:               cloneStringMap(headers),
		HeaderValues:          cloneStringSliceMap(headerValues),
		Batchable:             metadata.Batchable,
	}, nil
}

func Run(ctx context.Context, connector connectors.Connector, req Request, emit func(connectors.Record) error) (Result, error) {
	cmd, command, err := resolveRunnableCommand(connector, req.Path)
	if err != nil {
		return Result{}, err
	}
	if cmd.Intent == "direct_read" {
		if err := connectors.ValidateDirectReadPageCursor(req.PageCursor); err != nil {
			return Result{}, err
		}
		return runDirectRead(ctx, connector, cmd, req)
	}
	if cmd.Intent == "binary_download" || cmd.Intent == "text_export" {
		return runBinaryDownload(ctx, connector, cmd, req)
	}
	if cmd.Intent == "status_check" {
		return runStatusCheck(ctx, connector, cmd, req)
	}
	if cmd.Intent != "etl" || cmd.Availability != "implemented" || cmd.Stream == "" {
		return Result{}, &BlockedCommandError{
			Connector:    connector.Name(),
			Command:      command,
			Intent:       cmd.Intent,
			Availability: cmd.Availability,
			Reason:       blockReason(cmd),
		}
	}

	runtimeConfig, query, err := streamOverrides(cmd, req.Config, req.Flags)
	if err != nil {
		return Result{}, err
	}
	if err := validateCommandInputs(cmd, runtimeConfig, mappedCommandInputs{Query: query}); err != nil {
		return Result{}, err
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}
	result := Result{Connector: connector.Name(), Command: command, Stream: cmd.Stream}
	readReq := connectors.ReadRequest{
		Stream: cmd.Stream,
		Config: runtimeConfig,
		Query:  query,
		Limit:  limit,
	}
	err = connector.Read(ctx, readReq, connectors.LimitEmitter(limit, func(record connectors.Record) error {
		result.Count++
		return emit(record)
	}))
	if err := connectors.IgnoreReadLimit(err); err != nil {
		return Result{}, err
	}
	return result, nil
}

func resolveRunnableCommand(connector connectors.Connector, path []string) (connectors.CommandSurfaceCommand, string, error) {
	cmd, command, err := resolvePreflightCommand(connector, path)
	if err != nil {
		return connectors.CommandSurfaceCommand{}, command, err
	}
	if cmd.Intent == "etl" && cmd.Availability == "implemented" && cmd.Stream != "" {
		return cmd, command, nil
	}
	if cmd.Intent == "direct_read" && cmd.Availability == "implemented" {
		return cmd, command, nil
	}
	if (cmd.Intent == "binary_download" || cmd.Intent == "text_export") && cmd.Availability == "implemented" {
		return cmd, command, nil
	}
	if cmd.Intent == "status_check" && cmd.Availability == "implemented" {
		return cmd, command, nil
	}
	return connectors.CommandSurfaceCommand{}, command, &BlockedCommandError{
		Connector:    connector.Name(),
		Command:      command,
		Intent:       cmd.Intent,
		Availability: cmd.Availability,
		Reason:       blockReason(cmd),
	}
}

func resolvePreflightCommand(connector connectors.Connector, path []string) (connectors.CommandSurfaceCommand, string, error) {
	command := commandPath(path)
	if connector == nil {
		return connectors.CommandSurfaceCommand{}, command, &BlockedCommandError{Command: command, Reason: "connector is nil"}
	}
	if err := validateCommandPath(path); err != nil {
		return connectors.CommandSurfaceCommand{}, command, err
	}
	command = commandPath(path)
	surfaceProvider, ok := connector.(connectors.CommandSurfaceProvider)
	if !ok || surfaceProvider.CommandSurface() == nil {
		return connectors.CommandSurfaceCommand{}, command, &BlockedCommandError{Connector: connector.Name(), Command: command, Reason: "connector has no command surface"}
	}

	cmd, ok := findCommand(surfaceProvider.CommandSurface(), command)
	if !ok {
		return connectors.CommandSurfaceCommand{}, command, &BlockedCommandError{Connector: connector.Name(), Command: command, Reason: "unknown command"}
	}
	if cmd.Availability == "unsupported_with_provider_evidence" {
		if cmd.Unsupported == nil || strings.TrimSpace(cmd.Unsupported.Reason) == "" ||
			strings.TrimSpace(cmd.Unsupported.Target.SourceID) == "" ||
			strings.TrimSpace(cmd.Unsupported.Target.Method) == "" || strings.TrimSpace(cmd.Unsupported.Target.Path) == "" ||
			cmd.Foundation != nil || cmd.Stream != "" || cmd.Write != "" || cmd.Operation != "" {
			return connectors.CommandSurfaceCommand{}, command, &BlockedCommandError{
				Connector: connector.Name(), Command: command, Intent: cmd.Intent, Availability: cmd.Availability,
				Reason: "provider-evidenced unsupported command has invalid source disposition metadata",
			}
		}
		return connectors.CommandSurfaceCommand{}, command, unsupportedCommandError(connector.Name(), command, cmd)
	}
	if strings.HasPrefix(strings.TrimSpace(cmd.Notes), "missing_foundation=source-bound-read-execution-r1:") {
		return connectors.CommandSurfaceCommand{}, command, &BlockedCommandError{
			Connector:    connector.Name(),
			Command:      command,
			Intent:       cmd.Intent,
			Availability: cmd.Availability,
			Reason:       strings.TrimSpace(cmd.Notes),
		}
	}
	if cmd.Availability == "deferred" {
		implemented := cmd
		implemented.Availability = "implemented"
		implemented.Foundation = nil
		if _, err := preflightImplementedCommand(connector, implemented, command); err == nil {
			return connectors.CommandSurfaceCommand{}, command, &BlockedCommandError{
				Connector: connector.Name(), Command: command, Intent: cmd.Intent, Availability: cmd.Availability,
				Reason: "deferred command is stale: the exact command passes implemented runtime preflight",
			}
		}
		preflighter, supported := connector.(deferredCommandPreflighter)
		if !supported {
			return connectors.CommandSurfaceCommand{}, command, &BlockedCommandError{
				Connector: connector.Name(), Command: command, Intent: cmd.Intent, Availability: cmd.Availability,
				Reason: "connector does not expose exact deferred-command target preflight",
			}
		}
		if err := preflighter.PreflightDeferredCommand(cmd); err != nil {
			return connectors.CommandSurfaceCommand{}, command, &BlockedCommandError{
				Connector: connector.Name(), Command: command, Intent: cmd.Intent, Availability: cmd.Availability,
				Reason: fmt.Sprintf("deferred command target is not admissible: %v", err),
			}
		}
		return connectors.CommandSurfaceCommand{}, command, deferredCommandError(connector.Name(), command, cmd)
	}
	if cmd.Availability == "implemented" {
		implemented, err := preflightImplementedCommand(connector, cmd, command)
		return implemented, command, err
	}
	return connectors.CommandSurfaceCommand{}, command, &BlockedCommandError{
		Connector: connector.Name(), Command: command, Intent: cmd.Intent, Availability: cmd.Availability, Reason: blockReason(cmd),
	}
}

func unsupportedCommandError(connectorName, command string, cmd connectors.CommandSurfaceCommand) *BlockedCommandError {
	reason := boundedFailureMessage(cmd.Unsupported.Reason, 700)
	message := "provider-evidenced unsupported operation: " + reason
	classification, err := failures.New(failures.Input{
		Domain: failures.DomainSystem, Code: "provider_evidenced_unsupported", Message: message,
		DispatchKind: failures.DispatchKindDeclaredButUnroutableCommand,
		References: []failures.Reference{
			{Kind: failures.ReferenceKindConnector, Value: boundedFailureReference(connectorName)},
			{Kind: failures.ReferenceKindCommand, Value: boundedFailureReference(strings.ReplaceAll(command, " ", "_"))},
			{Kind: failures.ReferenceKindSource, Value: boundedFailureReference(cmd.Unsupported.Target.SourceID)},
		},
	}, nil)
	if err != nil {
		classification, _ = failures.New(failures.Input{
			Domain: failures.DomainSystem, Code: "provider_evidenced_unsupported", Message: "provider-evidenced operation is unsupported",
			DispatchKind: failures.DispatchKindDeclaredButUnroutableCommand,
		}, nil)
	}
	return &BlockedCommandError{
		Connector: connectorName, Command: command, Intent: cmd.Intent, Availability: cmd.Availability,
		Reason: message, Failure: classification,
	}
}

// preflightImplementedCommand is the executable commandrunner contract. A
// deferred row is tested through this exact helper before it may report a
// missing foundation, preventing runnable commands from being relabelled.
func preflightImplementedCommand(connector connectors.Connector, cmd connectors.CommandSurfaceCommand, command string) (connectors.CommandSurfaceCommand, error) {
	if preflighter, supported := connector.(implementedCommandPreflighter); supported {
		if err := preflighter.PreflightImplementedCommand(cmd); err != nil {
			return connectors.CommandSurfaceCommand{}, &BlockedCommandError{
				Connector: connector.Name(), Command: command, Intent: cmd.Intent, Availability: cmd.Availability,
				Reason: fmt.Sprintf("implemented command binding is not admissible: %v", err),
			}
		}
	}
	if err := preflightStructuredJSONFlags(connector, cmd); err != nil {
		return connectors.CommandSurfaceCommand{}, &BlockedCommandError{
			Connector: connector.Name(), Command: command, Intent: cmd.Intent, Availability: cmd.Availability, Reason: err.Error(),
		}
	}
	if cmd.Intent == "etl" && cmd.Availability == "implemented" {
		if cmd.SourceOperation != "" {
			if err := validateSourceBoundStreamReadCommand(connector, cmd); err != nil {
				return connectors.CommandSurfaceCommand{}, err
			}
			return cmd, nil
		}
		if cmd.Stream != "" {
			return cmd, nil
		}
	}
	if cmd.Operation != "" && cmd.Intent != "binary_download" && cmd.Intent != "text_export" &&
		cmd.Intent != "status_check" &&
		cmd.Intent != "direct_read" && cmd.Intent != "direct_write" {
		return connectors.CommandSurfaceCommand{}, &BlockedCommandError{
			Connector:    connector.Name(),
			Command:      command,
			Intent:       cmd.Intent,
			Availability: cmd.Availability,
			Reason:       fmt.Sprintf("operation %s executor is not implemented in this slice", cmd.Operation),
		}
	}
	if (cmd.Intent == "binary_download" || cmd.Intent == "text_export") && cmd.Availability == "implemented" {
		if err := validateBinaryDownloadCommand(connector, cmd); err != nil {
			return connectors.CommandSurfaceCommand{}, err
		}
		return cmd, nil
	}
	if cmd.Intent == "status_check" && cmd.Availability == "implemented" {
		if err := validateStatusCheckCommand(connector, cmd); err != nil {
			return connectors.CommandSurfaceCommand{}, err
		}
		return cmd, nil
	}
	if cmd.Intent == "direct_read" && cmd.Availability == "implemented" && cmd.Operation != "" {
		if err := validateOperationDirectReadCommand(connector, cmd); err != nil {
			return connectors.CommandSurfaceCommand{}, err
		}
		return cmd, nil
	}
	if cmd.Intent == "direct_read" && cmd.Availability == "implemented" {
		if err := validateDirectReadCommand(connector, cmd); err != nil {
			return connectors.CommandSurfaceCommand{}, err
		}
		return cmd, nil
	}
	if cmd.Intent == "direct_write" && cmd.Availability == "implemented" {
		if err := validateOperationDirectWriteCommand(connector, cmd); err != nil {
			return connectors.CommandSurfaceCommand{}, err
		}
		return cmd, nil
	}
	if cmd.Intent == "binary_upload" && cmd.Availability == "implemented" && cmd.Write != "" {
		if err := validateBinaryUploadCommand(connector, cmd); err != nil {
			return connectors.CommandSurfaceCommand{}, err
		}
		return cmd, nil
	}
	if cmd.Intent == "reverse_etl" && cmd.Availability == "implemented" && cmd.Write != "" {
		if preflighter, ok := connector.(declarativeWritePreflighter); ok {
			if err := preflighter.PreflightWriteAction(cmd.Write); err != nil {
				return connectors.CommandSurfaceCommand{}, &BlockedCommandError{
					Connector:    connector.Name(),
					Command:      command,
					Intent:       cmd.Intent,
					Availability: cmd.Availability,
					Reason:       fmt.Sprintf("write action %q is not promotable: %v", cmd.Write, err),
				}
			}
		}
		return cmd, nil
	}
	return connectors.CommandSurfaceCommand{}, &BlockedCommandError{
		Connector:    connector.Name(),
		Command:      command,
		Intent:       cmd.Intent,
		Availability: cmd.Availability,
		Reason:       blockReason(cmd),
	}
}

func deferredCommandError(connectorName, command string, cmd connectors.CommandSurfaceCommand) *BlockedCommandError {
	reason := "deferred command has no named missing foundation"
	var failure *failures.Classification
	if cmd.Foundation != nil && strings.TrimSpace(cmd.Foundation.ID) != "" && strings.TrimSpace(cmd.Foundation.Reason) != "" {
		foundationID := boundedFailureIdentifier(cmd.Foundation.ID)
		foundationReason := boundedFailureMessage(cmd.Foundation.Reason, 700)
		reason = fmt.Sprintf("missing foundation %q: %s", foundationID, foundationReason)
		classification, err := failures.New(failures.Input{
			Domain:       failures.DomainSystem,
			Code:         "missing_foundation",
			Message:      reason,
			DispatchKind: failures.DispatchKindDeclaredButUnroutableCommand,
			References: []failures.Reference{
				{Kind: failures.ReferenceKindConnector, Value: boundedFailureReference(connectorName)},
				{Kind: failures.ReferenceKindCommand, Value: boundedFailureReference(strings.ReplaceAll(command, " ", "_"))},
			},
		}, nil)
		if err != nil {
			classification, _ = failures.New(failures.Input{
				Domain: failures.DomainSystem, Code: "missing_foundation", Message: "missing declared foundation",
				DispatchKind: failures.DispatchKindDeclaredButUnroutableCommand,
			}, nil)
		}
		failure = classification
	}
	return &BlockedCommandError{
		Connector: connectorName, Command: command, Intent: cmd.Intent, Availability: cmd.Availability, Reason: reason, Failure: failure,
	}
}

func boundedFailureIdentifier(raw string) string {
	value := strings.TrimSpace(safety.SanitizeTerminal(raw))
	if value != "" && len(value) <= 128 && failureReferenceSafe(value) {
		return value
	}
	return failureDigest(raw)
}

func boundedFailureReference(raw string) string {
	value := strings.TrimSpace(safety.SanitizeTerminal(raw))
	if value != "" && len(value) <= 256 && failureReferenceSafe(value) {
		return value
	}
	return failureDigest(raw)
}

func boundedFailureMessage(raw string, maximum int) string {
	value := strings.TrimSpace(safety.SanitizeTerminal(raw))
	if value == "" {
		return "declared foundation is unavailable"
	}
	for len(value) > maximum {
		_, size := utf8.DecodeLastRuneInString(value)
		if size <= 0 {
			return "declared foundation is unavailable"
		}
		value = value[:len(value)-size]
	}
	return value
}

func failureReferenceSafe(value string) bool {
	for _, character := range value {
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || strings.ContainsRune("._:/-", character) {
			continue
		}
		return false
	}
	return true
}

func failureDigest(raw string) string {
	digest := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("sha256-%x", digest[:12])
}

func validateBinaryUploadCommand(connector connectors.Connector, cmd connectors.CommandSurfaceCommand) error {
	if strings.TrimSpace(cmd.Write) == "" {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "binary_upload commands require a declared write action"}
	}
	if cmd.Stream != "" || cmd.Operation != "" {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "binary_upload commands may bind only one declared write action"}
	}
	preflighter, ok := connector.(binaryUploadActionPreflighter)
	if !ok {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "connector does not expose a declarative binary upload action"}
	}
	sources, err := preflighter.PreflightBinaryUploadAction(cmd.Write)
	if err != nil {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("binary upload action %q is not safely promotable: %v", cmd.Write, err)}
	}
	for _, source := range sources {
		declaredTarget := "record." + source.Field
		bound := false
		for _, flag := range cmd.Flags {
			if flag.Required && flag.MapsTo == declaredTarget {
				bound = true
				break
			}
		}
		if !bound {
			return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("binary upload source %q must be one required declared command flag", source.Field)}
		}
	}
	return nil
}

// preflightStructuredJSONFlags preserves the original reverse-ETL record
// boundary and adds operation-specific exceptions. A JSON flag never means
// "take an arbitrary request body": direct writes and reads can name only a
// declared body field from their fixed operation. The engine proves the named
// value is a closed, bounded object or array in that exact declaration.
func preflightStructuredJSONFlags(connector connectors.Connector, cmd connectors.CommandSurfaceCommand) error {
	var preflighter structuredJSONRecordPreflighter
	var stringArmPreflighter structuredJSONRecordStringArmPreflighter
	var operationPreflighter structuredJSONOperationBodyPreflighter
	for _, flag := range cmd.Flags {
		if flag.AllowBareString && flag.Type != "json" {
			return fmt.Errorf("bare-string flag --%s must use the bounded json flag type", flag.Name)
		}
		if flag.Type != "json" {
			continue
		}
		switch {
		case (cmd.Intent == "reverse_etl" || cmd.Intent == "binary_upload") && strings.TrimSpace(cmd.Write) != "":
			field, ok := strings.CutPrefix(flag.MapsTo, "record.")
			if !ok || field == "" {
				return fmt.Errorf("structured JSON flag --%s must map to a record field", flag.Name)
			}
			if preflighter == nil {
				var supported bool
				preflighter, supported = connector.(structuredJSONRecordPreflighter)
				if !supported {
					return fmt.Errorf("structured JSON flag --%s requires a declarative record-schema preflight", flag.Name)
				}
			}
			if err := preflighter.PreflightStructuredJSONRecordField(cmd.Write, field); err != nil {
				return fmt.Errorf("structured JSON flag --%s is not declared safely: %w", flag.Name, err)
			}
			if flag.AllowBareString {
				if stringArmPreflighter == nil {
					var supported bool
					stringArmPreflighter, supported = connector.(structuredJSONRecordStringArmPreflighter)
					if !supported {
						return fmt.Errorf("bare-string flag --%s requires a declarative string-arm preflight", flag.Name)
					}
				}
				if err := stringArmPreflighter.PreflightStructuredJSONRecordStringArm(cmd.Write, field); err != nil {
					return fmt.Errorf("bare-string flag --%s is not declared safely: %w", flag.Name, err)
				}
			}
		case (cmd.Intent == "direct_write" || cmd.Intent == "direct_read") && strings.TrimSpace(cmd.Operation) != "":
			if flag.AllowBareString {
				return fmt.Errorf("bare-string flag --%s is allowed only on a declared reverse-ETL record field", flag.Name)
			}
			variable, ok := strings.CutPrefix(flag.MapsTo, "body.")
			if !ok || variable == "" {
				return fmt.Errorf("structured JSON flag --%s must map to a declared body field of its operation", flag.Name)
			}
			if cmd.Intent == "direct_read" && strings.Contains(variable, ".") {
				return fmt.Errorf("structured JSON flag --%s must map to one top-level body field of a fixed operation", flag.Name)
			}
			if operationPreflighter == nil {
				var supported bool
				operationPreflighter, supported = connector.(structuredJSONOperationBodyPreflighter)
				if !supported {
					return fmt.Errorf("structured JSON flag --%s requires declaration-backed operation body preflight", flag.Name)
				}
			}
			if err := operationPreflighter.PreflightOperationStructuredJSONBodyField(cmd.Operation, variable); err != nil {
				return fmt.Errorf("structured JSON flag --%s is not declared safely: %w", flag.Name, err)
			}
		default:
			return fmt.Errorf("structured JSON flag --%s is allowed only on a declared reverse-ETL record field, fixed REST body field, or fixed GraphQL variable", flag.Name)
		}
	}
	return nil
}

/*
	The structured body preflight above intentionally shares the REST and
	GraphQL declaration boundary. Do not route json flags through a transport
	method/path/body escape hatch: the fixed operation and its top-level mapped
	field remain the only authority.
*/

// directReadPageFlagNames are consumed as Request.Page/Request.PageCursor
// rather than as command flags. Only this intent can honour them, so they are
// dropped here and nowhere earlier: every other intent then keeps its existing
// "unknown flag --page" refusal instead of accepting and ignoring them.
var directReadPageFlagNames = []string{"page", "page-cursor"}

func withoutDirectReadPageFlags(flags map[string][]string) map[string][]string {
	out := make(map[string][]string, len(flags))
	for name, values := range flags {
		out[name] = values
	}
	for _, name := range directReadPageFlagNames {
		delete(out, name)
	}
	return out
}

func runDirectRead(ctx context.Context, connector connectors.Connector, cmd connectors.CommandSurfaceCommand, req Request) (Result, error) {
	req.Flags = withoutDirectReadPageFlags(req.Flags)
	if cmd.Operation != "" {
		return runOperationDirectRead(ctx, connector, cmd, req)
	}
	if err := validateDirectReadCommand(connector, cmd); err != nil {
		return Result{}, err
	}
	pathParams, query, err := directReadOverrides(cmd, req.Flags)
	if err != nil {
		return Result{}, err
	}
	if err := validateCommandInputs(cmd, req.Config, mappedCommandInputs{Query: query}); err != nil {
		return Result{}, err
	}
	maxBytes := req.MaxBytes
	if maxBytes <= 0 {
		maxBytes = MaxDirectReadBytes
	}
	if maxBytes > MaxDirectReadBytes {
		maxBytes = MaxDirectReadBytes
	}
	endpoint := cmd.APISurface[0]
	method := strings.ToUpper(strings.TrimSpace(endpoint.Method))
	direct, err := connector.(connectors.DirectReader).DirectRead(ctx, connectors.DirectReadRequest{
		Method:       method,
		Path:         endpoint.Path,
		Config:       req.Config,
		PathParams:   pathParams,
		Query:        query,
		MaxBytes:     maxBytes,
		OutputPolicy: cmd.OutputPolicy,
		Page:         req.Page,
		PageCursor:   req.PageCursor,
	})
	direct = connectors.SanitizeDirectReadResultForOutput(direct, req.Config.Secrets)
	if err != nil {
		if !directReadHasProviderEvidence(direct) {
			return Result{}, err
		}
		return Result{Connector: connector.Name(), Command: cmd.Path, DirectRead: &direct}, err
	}
	if err := assertDirectReadNavigated(connector, cmd, req, direct); err != nil {
		if !directReadHasProviderEvidence(direct) {
			return Result{}, err
		}
		return Result{Connector: connector.Name(), Command: cmd.Path, DirectRead: &direct}, err
	}
	return Result{
		Connector:  connector.Name(),
		Command:    cmd.Path,
		DirectRead: &direct,
	}, nil
}

// assertDirectReadNavigated is the general guard against a direct-read
// executor that ACCEPTS page navigation and ignores it.
//
// Page and PageCursor are handed to whatever DirectReader/OperationDirectReader
// a connector supplies, and nothing in the type system makes an implementation
// honour them. One that does not returns a zero-value DirectReadPage, so the
// caller gets page one at status 200 with nothing saying the request was
// discarded — precisely the accepted-and-ignored wrongness --page exists to
// remove, pointed at navigation instead of at records.
//
// The check is on the reported page rather than on an opt-in interface so a
// future executor cannot regress by forgetting to declare anything.
func assertDirectReadNavigated(connector connectors.Connector, cmd connectors.CommandSurfaceCommand, req Request, result connectors.DirectReadResult) error {
	if req.Page <= 1 && req.PageCursor == "" {
		return nil
	}
	if result.Page.Strategy != "" {
		return nil
	}
	return fmt.Errorf("connector %q command %q accepted page navigation but reported no page context; its direct read cannot address another page", connector.Name(), cmd.Path)
}

func runOperationDirectRead(ctx context.Context, connector connectors.Connector, cmd connectors.CommandSurfaceCommand, req Request) (Result, error) {
	reader, ok := connector.(connectors.OperationDirectReader)
	if !ok {
		return Result{}, &BlockedCommandError{
			Connector:    connector.Name(),
			Command:      cmd.Path,
			Intent:       cmd.Intent,
			Availability: cmd.Availability,
			Reason:       "connector does not support operation direct reads",
		}
	}
	if err := validateOperationDirectReadCommand(connector, cmd); err != nil {
		return Result{}, err
	}
	pathParams, query, headers, headerValues, body, rawBody, err := operationDirectReadOverrides(cmd, req.Flags)
	if err != nil {
		return Result{}, err
	}
	if err := validateCommandInputs(cmd, req.Config, mappedCommandInputs{Query: query, Body: body}); err != nil {
		return Result{}, err
	}
	maxBytes := req.MaxBytes
	if maxBytes <= 0 {
		maxBytes = MaxOperationDirectReadBytes
	}
	if maxBytes > MaxOperationDirectReadBytes {
		maxBytes = MaxOperationDirectReadBytes
	}
	direct, err := reader.OperationDirectRead(ctx, connectors.OperationDirectReadRequest{
		Operation:  cmd.Operation,
		Config:     req.Config,
		PathParams: pathParams,
		Query:      query,
		CommandBindings: &connectors.OperationDirectReadBindings{
			Path:  operationMappedFields(cmd, "path."),
			Query: operationMappedFields(cmd, "query."),
			Body:  operationMappedFields(cmd, "body."),
			RawBody: slices.ContainsFunc(cmd.Flags, func(flag connectors.CommandSurfaceFlag) bool {
				return strings.TrimSpace(flag.MapsTo) == "body"
			}),
		},
		Headers:      headers,
		HeaderValues: headerValues,
		Body:         body,
		RawBody:      rawBody,
		MaxBytes:     maxBytes,
		OutputPolicy: cmd.OutputPolicy,
		Page:         req.Page,
		PageCursor:   req.PageCursor,
	})
	direct = connectors.SanitizeDirectReadResultForOutput(direct, req.Config.Secrets)
	if err != nil {
		if !directReadHasProviderEvidence(direct) {
			return Result{}, err
		}
		return Result{Connector: connector.Name(), Command: cmd.Path, DirectRead: &direct}, err
	}
	if err := assertDirectReadNavigated(connector, cmd, req, direct); err != nil {
		if !directReadHasProviderEvidence(direct) {
			return Result{}, err
		}
		return Result{Connector: connector.Name(), Command: cmd.Path, DirectRead: &direct}, err
	}
	return Result{Connector: connector.Name(), Command: cmd.Path, DirectRead: &direct}, nil
}

func operationMappedFields(cmd connectors.CommandSurfaceCommand, prefix string) []string {
	fields := make([]string, 0, len(cmd.Flags))
	for _, flag := range cmd.Flags {
		mapsTo := strings.TrimSpace(flag.MapsTo)
		if strings.HasPrefix(mapsTo, prefix) {
			fields = append(fields, strings.TrimPrefix(mapsTo, prefix))
		}
	}
	return fields
}

func validateDirectReadCommand(connector connectors.Connector, cmd connectors.CommandSurfaceCommand) error {
	if _, ok := connector.(connectors.DirectReader); !ok {
		return &BlockedCommandError{
			Connector:    connector.Name(),
			Command:      cmd.Path,
			Intent:       cmd.Intent,
			Availability: cmd.Availability,
			Reason:       "connector does not support direct reads",
		}
	}
	if len(cmd.APISurface) != 1 {
		return &BlockedCommandError{
			Connector:    connector.Name(),
			Command:      cmd.Path,
			Intent:       cmd.Intent,
			Availability: cmd.Availability,
			Reason:       "direct_read commands require exactly one api_surface endpoint",
		}
	}
	endpoint := cmd.APISurface[0]
	method := strings.ToUpper(strings.TrimSpace(endpoint.Method))
	if method != http.MethodGet {
		return &BlockedCommandError{
			Connector:    connector.Name(),
			Command:      cmd.Path,
			Intent:       cmd.Intent,
			Availability: cmd.Availability,
			Reason:       fmt.Sprintf("direct_read commands require GET api_surface endpoints, got %s", method),
		}
	}
	if isAbsoluteHTTPURL(endpoint.Path) {
		return &BlockedCommandError{
			Connector:    connector.Name(),
			Command:      cmd.Path,
			Intent:       cmd.Intent,
			Availability: cmd.Availability,
			Reason:       "direct_read commands must not reference an absolute URL",
		}
	}
	if !isSupportedDirectReadOutputPolicy(cmd.OutputPolicy) {
		return &BlockedCommandError{
			Connector:    connector.Name(),
			Command:      cmd.Path,
			Intent:       cmd.Intent,
			Availability: cmd.Availability,
			Reason:       "direct_read commands require an explicit supported output_policy",
		}
	}
	if isOperationOnlyDirectReadOutputPolicy(cmd.OutputPolicy) {
		return &BlockedCommandError{
			Connector:    connector.Name(),
			Command:      cmd.Path,
			Intent:       cmd.Intent,
			Availability: cmd.Availability,
			Reason:       fmt.Sprintf("direct_read output policy %q requires an operation-backed command", cmd.OutputPolicy),
		}
	}
	return nil
}

func validateOperationDirectReadCommand(connector connectors.Connector, cmd connectors.CommandSurfaceCommand) error {
	if _, ok := connector.(connectors.OperationDirectReader); !ok {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "connector does not support operation direct reads"}
	}
	preflighter, ok := connector.(connectors.OperationDirectReadPreflighter)
	if !ok {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "connector does not expose operation direct-read metadata"}
	}
	bindingPreflighter, ok := connector.(connectors.OperationDirectReadBindingPreflighter)
	if !ok {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "connector does not expose operation direct-read binding metadata"}
	}
	if strings.TrimSpace(cmd.Operation) == "" {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "operation direct_read commands require operation"}
	}
	if len(cmd.APISurface) != 1 {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "operation direct_read commands require exactly one api_surface endpoint"}
	}
	method := strings.ToUpper(strings.TrimSpace(cmd.APISurface[0].Method))
	if method != http.MethodGet && method != http.MethodPost {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("operation direct_read commands require GET or POST api_surface endpoints, got %s", method)}
	}
	if isAbsoluteHTTPURL(cmd.APISurface[0].Path) {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "operation direct_read commands must not reference an absolute URL"}
	}
	if !isSupportedDirectReadOutputPolicy(cmd.OutputPolicy) {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "operation direct_read commands require an explicit supported output_policy"}
	}
	if err := preflighter.PreflightOperationDirectRead(cmd.Operation, method, cmd.APISurface[0].Path, MaxOperationDirectReadBytes, cmd.OutputPolicy); err != nil {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("operation direct read metadata is not executable: %v", err)}
	}
	if cmd.SourceOperation != "" {
		sourcePreflighter, ok := connector.(connectors.SourceBoundReadPreflighter)
		if !ok {
			return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "connector does not expose source-bound read metadata"}
		}
		if err := sourcePreflighter.PreflightSourceBoundRead(cmd.Operation, cmd.SourceOperation, method, cmd.APISurface[0].Path); err != nil {
			return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("source-bound read metadata is not executable: %v", err)}
		}
	}
	pathFields := make([]string, 0, len(cmd.Flags))
	queryFields := make([]string, 0, len(cmd.Flags))
	bodyFields := make([]string, 0, len(cmd.Flags))
	rawBody := false
	for _, flag := range cmd.Flags {
		mapsTo := strings.TrimSpace(flag.MapsTo)
		switch {
		case strings.HasPrefix(mapsTo, "path."):
			pathFields = append(pathFields, strings.TrimPrefix(mapsTo, "path."))
		case strings.HasPrefix(mapsTo, "query."):
			queryFields = append(queryFields, strings.TrimPrefix(mapsTo, "query."))
		case strings.HasPrefix(mapsTo, "header."):
		case strings.HasPrefix(mapsTo, "body."):
			bodyFields = append(bodyFields, strings.TrimPrefix(mapsTo, "body."))
		case mapsTo == "body":
			rawBody = true
		default:
			return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("operation direct_read flag --%s maps to unsupported target %q", flag.Name, flag.MapsTo)}
		}
	}
	if err := bindingPreflighter.PreflightOperationDirectReadBindings(cmd.Operation, pathFields, queryFields, bodyFields, rawBody); err != nil {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("operation direct read bindings are not executable: %v", err)}
	}
	return nil
}

func validateSourceBoundStreamReadCommand(connector connectors.Connector, cmd connectors.CommandSurfaceCommand) error {
	if cmd.SourceOperation == "" {
		return nil
	}
	if cmd.Stream == "" || len(cmd.APISurface) != 1 {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "source-bound ETL commands require one declared stream and exactly one api_surface endpoint"}
	}
	method := strings.ToUpper(strings.TrimSpace(cmd.APISurface[0].Method))
	if method != http.MethodGet || isAbsoluteHTTPURL(cmd.APISurface[0].Path) {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "source-bound ETL commands require one fixed relative GET endpoint"}
	}
	preflighter, ok := connector.(connectors.SourceBoundStreamReadPreflighter)
	if !ok {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "connector does not expose source-bound stream metadata"}
	}
	if err := preflighter.PreflightSourceBoundStreamRead(cmd.Stream, cmd.SourceOperation, method, cmd.APISurface[0].Path); err != nil {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("source-bound stream metadata is not executable: %v", err)}
	}
	return nil
}

// validateOperationDirectWriteCommand is deliberately limited to the shape
// the REST/fixed-GraphQL engine executor can prove safe. Its result makes a
// command eligible for the plan lifecycle only; resolveRunnableCommand still
// refuses direct execution so every write traverses plan -> preview -> approval.
func validateOperationDirectWriteCommand(connector connectors.Connector, cmd connectors.CommandSurfaceCommand) error {
	if _, ok := connector.(connectors.OperationDirectWriter); !ok {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "connector does not support operation direct writes"}
	}
	preflighter, ok := connector.(connectors.OperationDirectWritePreflighter)
	if !ok {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "connector does not expose operation direct-write metadata"}
	}
	bindingPreflighter, ok := connector.(connectors.OperationDirectWriteBindingPreflighter)
	if !ok {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "connector does not expose operation direct-write binding metadata"}
	}
	if _, ok := connector.(connectors.OperationDirectWriteBodyMaterializer); !ok {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "connector does not expose operation direct-write body materialization"}
	}
	if _, ok := connector.(connectors.OperationDirectWriteBodyValueResolver); !ok {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "connector does not expose operation direct-write body resolution"}
	}
	metadataProvider, ok := connector.(connectors.OperationDirectWriteMetadataProvider)
	if !ok {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "connector does not expose operation direct-write metadata"}
	}
	if strings.TrimSpace(cmd.Operation) == "" {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "direct_write commands require operation"}
	}
	if len(cmd.APISurface) != 1 {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "direct_write commands require exactly one api_surface endpoint"}
	}
	method := strings.ToUpper(strings.TrimSpace(cmd.APISurface[0].Method))
	if !isOperationDirectWriteMethod(method) {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("direct_write commands require POST, PUT, PATCH, or DELETE api_surface endpoints, got %s", method)}
	}
	if isAbsoluteHTTPURL(cmd.APISurface[0].Path) {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "direct_write commands must not reference an absolute URL"}
	}
	if !isSupportedDirectWriteOutputPolicy(cmd.OutputPolicy) {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "direct_write commands require an explicit supported output_policy"}
	}
	metadata, err := metadataProvider.OperationDirectWriteMetadata(cmd.Operation)
	if err != nil {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("direct_write operation metadata is not executable: %v", err)}
	}
	if metadata.Operation != cmd.Operation {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "direct_write operation metadata did not match command operation"}
	}
	if metadata.OutputPolicy != cmd.OutputPolicy {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "direct_write command output_policy does not match declared operation"}
	}
	queryFields := make([]string, 0, len(cmd.Flags))
	pathFields := make([]string, 0, len(cmd.Flags))
	bodyFields := make([]string, 0, len(cmd.Flags))
	for _, flag := range cmd.Flags {
		mapsTo := strings.TrimSpace(flag.MapsTo)
		switch {
		case strings.HasPrefix(mapsTo, "query."):
			field, _ := strings.CutPrefix(mapsTo, "query.")
			queryFields = append(queryFields, field)
		case strings.HasPrefix(mapsTo, "path."):
			field, _ := strings.CutPrefix(mapsTo, "path.")
			pathFields = append(pathFields, field)
		case strings.HasPrefix(mapsTo, "body."):
			field, _ := strings.CutPrefix(mapsTo, "body.")
			bodyFields = append(bodyFields, field)
		case strings.HasPrefix(mapsTo, "header."):
			// The declared operation remains the authority for request-header
			// names, repeatability, schemas, and byte caps. Its engine-owned
			// pre-I/O preparation validates those bindings against the exact
			// supplied values; commandrunner only admits the closed header
			// mapping shape here, rather than treating it as a raw HTTP escape.
			if err := validateOperationHeaderTarget(strings.TrimPrefix(mapsTo, "header.")); err != nil {
				return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("direct_write flag --%s maps to invalid request header: %v", flag.Name, err)}
			}
		default:
			return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("direct_write flag --%s maps to unsupported target %q", flag.Name, flag.MapsTo)}
		}
	}
	if err := preflighter.PreflightOperationDirectWrite(cmd.Operation, method, cmd.APISurface[0].Path, cmd.OutputPolicy, queryFields...); err != nil {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("operation direct write metadata is not executable: %v", err)}
	}
	if err := bindingPreflighter.PreflightOperationDirectWriteBindings(cmd.Operation, pathFields, bodyFields); err != nil {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("operation direct write bindings are not executable: %v", err)}
	}
	return nil
}

// Keep these closed policy sets enumerable: the CLI schema regression test
// compares their union with its output_policy enum so declaration and runtime
// support cannot silently drift apart.
var (
	supportedDirectReadOutputPolicies = map[string]struct{}{
		"repository_contents_file_metadata": {},
		"repository_contents_directory":     {},
		"json_redacted":                     {},
		"clinical_json_redacted":            {},
		"none":                              {},
		"text":                              {},
	}
	supportedDirectWriteOutputPolicies = map[string]struct{}{
		"none":                        {},
		"json":                        {},
		"json_redacted":               {},
		"write_result_redacted":       {},
		"gong_bounded_input_redacted": {},
		"secret_stored":               {},
	}
)

func isSupportedDirectReadOutputPolicy(policy string) bool {
	_, ok := supportedDirectReadOutputPolicies[policy]
	return ok
}

func isOperationOnlyDirectReadOutputPolicy(policy string) bool {
	return policy == "none" || policy == "text"
}

func isSupportedDirectWriteOutputPolicy(policy string) bool {
	_, ok := supportedDirectWriteOutputPolicies[policy]
	return ok
}

func isOperationDirectWriteMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func commandPath(path []string) string {
	return strings.Join(path, " ")
}

// CommandPathSegments parses a declaration-owned command path into the exact
// segments commandrunner resolves. It rejects alternate whitespace spellings,
// empty segments, and unsafe identifiers so a certification row cannot claim
// a command that the runtime parser can never dispatch.
func CommandPathSegments(raw string) ([]string, error) {
	if raw == "" || raw != strings.TrimSpace(raw) {
		return nil, fmt.Errorf("command path must be non-empty and trimmed")
	}
	parts := strings.Split(raw, " ")
	if strings.Join(parts, " ") != raw {
		return nil, fmt.Errorf("command path must use one ASCII space between segments")
	}
	if err := validateCommandPath(parts); err != nil {
		return nil, err
	}
	return parts, nil
}

func validateCommandPath(path []string) error {
	if len(path) == 0 {
		return &BlockedCommandError{Reason: "missing command path"}
	}
	for i, part := range path {
		if err := safety.ValidateIdentifier(part, fmt.Sprintf("command path segment %d", i+1)); err != nil {
			return err
		}
	}
	return nil
}

func findCommand(surface *connectors.CommandSurface, path string) (connectors.CommandSurfaceCommand, bool) {
	for _, cmd := range surface.Commands {
		if cmd.Path == path {
			return cmd, true
		}
	}
	return connectors.CommandSurfaceCommand{}, false
}

func blockReason(cmd connectors.CommandSurfaceCommand) string {
	switch {
	case cmd.Intent == "direct_write":
		return "direct_write commands require plan, preview, approval, execute"
	case cmd.Operation != "":
		return fmt.Sprintf("operation %s executor is not implemented in this slice", cmd.Operation)
	case (cmd.Intent == "reverse_etl" || cmd.Intent == "binary_upload") && cmd.Write == "":
		return "implemented reverse ETL commands must reference write action"
	case cmd.Intent == "reverse_etl" || cmd.Intent == "binary_upload":
		if cmd.Approval != "" {
			return cmd.Approval
		}
		if cmd.Intent == "binary_upload" {
			return "binary uploads require plan, preview, approval, execute"
		}
		return "reverse ETL writes require plan, preview, approval, execute"
	case cmd.Intent == "local_workflow":
		if cmd.Notes != "" {
			return cmd.Notes
		}
		return "local workflow commands are not connector API operations"
	case cmd.Risk != "":
		return cmd.Risk
	case cmd.Notes != "":
		return cmd.Notes
	default:
		return "only implemented ETL stream commands are executable"
	}
}

func streamOverrides(cmd connectors.CommandSurfaceCommand, cfg connectors.RuntimeConfig, flags map[string][]string) (connectors.RuntimeConfig, map[string]string, error) {
	allowed, err := validateCommandFlagSet(cmd, flags)
	if err != nil {
		return connectors.RuntimeConfig{}, nil, err
	}

	query := map[string]string{}
	configOverrides := map[string]string{}
	for name, values := range flags {
		if len(values) == 0 {
			continue
		}
		if err := safety.ValidateIdentifier(name, "flag name"); err != nil {
			return connectors.RuntimeConfig{}, nil, err
		}
		flag, ok := allowed[name]
		if !ok {
			return connectors.RuntimeConfig{}, nil, fmt.Errorf("unknown flag --%s for command %q", name, cmd.Path)
		}
		value := values[len(values)-1]
		if err := safety.RejectDangerousChars(value, "flag value"); err != nil {
			return connectors.RuntimeConfig{}, nil, err
		}
		if err := validateFlagValue(flag, value); err != nil {
			return connectors.RuntimeConfig{}, nil, err
		}
		switch {
		case strings.HasPrefix(flag.MapsTo, "query."):
			target := strings.TrimPrefix(flag.MapsTo, "query.")
			if err := safety.ValidateIdentifier(target, "query parameter"); err != nil {
				return connectors.RuntimeConfig{}, nil, err
			}
			query[target] = value
		case strings.HasPrefix(flag.MapsTo, "config."):
			target := strings.TrimPrefix(flag.MapsTo, "config.")
			if err := safety.ValidateIdentifier(target, "config parameter"); err != nil {
				return connectors.RuntimeConfig{}, nil, err
			}
			configOverrides[target] = value
		default:
			return connectors.RuntimeConfig{}, nil, &BlockedCommandError{
				Command:      cmd.Path,
				Intent:       cmd.Intent,
				Availability: cmd.Availability,
				Reason:       fmt.Sprintf("flag --%s maps to unsupported target %q", name, flag.MapsTo),
			}
		}
	}
	runtimeConfig := runtimeConfigWithOverrides(cfg, configOverrides)
	if err := validateConfiguredFlagValues(cmd, runtimeConfig, nil); err != nil {
		return connectors.RuntimeConfig{}, nil, err
	}
	return runtimeConfig, query, nil
}

func runtimeConfigWithOverrides(cfg connectors.RuntimeConfig, overrides map[string]string) connectors.RuntimeConfig {
	if len(overrides) == 0 {
		return cfg
	}
	out := cfg
	out.Config = make(map[string]string, len(cfg.Config)+len(overrides))
	for key, value := range cfg.Config {
		out.Config[key] = value
	}
	for key, value := range overrides {
		out.Config[key] = value
	}
	return out
}

func validateConfiguredFlagValues(cmd connectors.CommandSurfaceCommand, cfg connectors.RuntimeConfig, explicitFlags map[string][]string) error {
	overriddenTargets := map[string]struct{}{}
	for _, flag := range cmd.Flags {
		target, mapped := strings.CutPrefix(flag.MapsTo, "config.")
		if !mapped {
			continue
		}
		if err := safety.ValidateIdentifier(target, "config parameter"); err != nil {
			return err
		}
		if len(explicitFlags[flag.Name]) > 0 {
			overriddenTargets[target] = struct{}{}
		}
	}
	for _, flag := range cmd.Flags {
		target, mapped := strings.CutPrefix(flag.MapsTo, "config.")
		if !mapped {
			continue
		}
		if _, overridden := overriddenTargets[target]; overridden {
			continue
		}
		value, ok := cfg.Config[target]
		if !ok {
			continue
		}
		if _, err := coerceCommandFlagValue(cmd, flag, []string{value}); err != nil {
			return &ConfiguredFlagValueError{Command: cmd.Path, Flag: flag.Name, Target: target, err: err}
		}
	}
	return nil
}

type mappedCommandInputs struct {
	Query     map[string]string
	Body      map[string]any
	BodyValue func(string) (any, bool, error)
}

func validateCommandInputs(cmd connectors.CommandSurfaceCommand, cfg connectors.RuntimeConfig, inputs mappedCommandInputs) error {
	for _, constraint := range cmd.Constraints {
		if err := validateCommandConstraint(constraint, cfg, inputs); err != nil {
			return err
		}
	}
	return nil
}

func validateCommandConstraint(constraint connectors.CommandSurfaceConstraint, cfg connectors.RuntimeConfig, inputs mappedCommandInputs) error {
	switch constraint.Kind {
	case "order":
		return validateOrderConstraint(constraint, cfg, inputs)
	case "exactly_one":
		return validateExactlyOneConstraint(constraint, cfg, inputs)
	default:
		return &BlockedCommandError{Command: "unknown", Reason: fmt.Sprintf("unsupported command constraint kind %q", constraint.Kind)}
	}
}

func validateExactlyOneConstraint(constraint connectors.CommandSurfaceConstraint, cfg connectors.RuntimeConfig, inputs mappedCommandInputs) error {
	present := 0
	for _, target := range constraint.Fields {
		_, targetPresent, _, err := validationTargetValue(target, cfg, inputs)
		if err != nil {
			return err
		}
		if targetPresent {
			present++
		}
	}
	if present == 1 {
		return nil
	}
	if strings.TrimSpace(constraint.Message) != "" {
		return errors.New(constraint.Message)
	}
	return fmt.Errorf("invalid command constraint: exactly one of %s must be provided", strings.Join(constraint.Fields, ", "))
}

func validateOrderConstraint(constraint connectors.CommandSurfaceConstraint, cfg connectors.RuntimeConfig, inputs mappedCommandInputs) error {
	leftValue, leftPresent, leftLabel, err := validationValueWithFallback(constraint.Left, constraint.LeftFallback, cfg, inputs)
	if err != nil {
		return err
	}
	rightValue, rightPresent, rightLabel, err := validationValueWithFallback(constraint.Right, constraint.RightFallback, cfg, inputs)
	if err != nil {
		return err
	}
	if !leftPresent || !rightPresent {
		return nil
	}
	if constraint.Op != "lt" {
		return &BlockedCommandError{Command: "unknown", Reason: fmt.Sprintf("unsupported command constraint operator %q", constraint.Op)}
	}
	if constraint.ValueType != "date-time" {
		return &BlockedCommandError{Command: "unknown", Reason: fmt.Sprintf("unsupported command constraint value_type %q", constraint.ValueType)}
	}
	left, err := parseDateTimeValue(leftValue, leftLabel)
	if err != nil {
		return err
	}
	right, err := parseDateTimeValue(rightValue, rightLabel)
	if err != nil {
		return err
	}
	if !left.Before(right) {
		if strings.TrimSpace(constraint.Message) != "" {
			return errors.New(constraint.Message)
		}
		return fmt.Errorf("invalid command constraint: %s must be before %s", leftLabel, rightLabel)
	}
	return nil
}

func validationValueWithFallback(primary, fallback string, cfg connectors.RuntimeConfig, inputs mappedCommandInputs) (string, bool, string, error) {
	value, present, label, err := validationTargetValue(primary, cfg, inputs)
	if err != nil || present || fallback == "" {
		return value, present, label, err
	}
	return validationTargetValue(fallback, cfg, inputs)
}

func validationTargetValue(target string, cfg connectors.RuntimeConfig, inputs mappedCommandInputs) (string, bool, string, error) {
	switch {
	case strings.HasPrefix(target, "query."):
		key := strings.TrimPrefix(target, "query.")
		value, present := inputs.Query[key]
		return strings.TrimSpace(value), present, target, nil
	case strings.HasPrefix(target, "body."):
		path := strings.TrimPrefix(target, "body.")
		if inputs.BodyValue != nil {
			value, present, err := inputs.BodyValue(path)
			if err != nil {
				return "", false, target, err
			}
			return strings.TrimSpace(fmt.Sprint(value)), present, target, nil
		}
		value, present := nestedBodyValue(inputs.Body, path)
		return strings.TrimSpace(fmt.Sprint(value)), present, target, nil
	case strings.HasPrefix(target, "config."):
		key := strings.TrimPrefix(target, "config.")
		value := strings.TrimSpace(cfg.Config[key])
		return value, value != "", target, nil
	default:
		return "", false, target, &BlockedCommandError{Command: "unknown", Reason: fmt.Sprintf("unsupported command validation target %q", target)}
	}
}

func nestedBodyValue(body map[string]any, path string) (any, bool) {
	if body == nil || strings.TrimSpace(path) == "" {
		return nil, false
	}
	parts := strings.Split(path, ".")
	var cur any = body
	for _, part := range parts {
		switch value := cur.(type) {
		case map[string]any:
			next, ok := value[part]
			if !ok || next == nil {
				return nil, false
			}
			cur = next
		case []any:
			index, ok := bodyMappingArrayIndex(part)
			if !ok || index >= uint64(len(value)) || value[index] == nil {
				return nil, false
			}
			cur = value[index]
		default:
			return nil, false
		}
	}
	return cur, true
}

func parseDateTimeValue(value, label string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid %s, want ISO-8601/RFC3339 timestamp", label)
	}
	return parsed, nil
}

// validateCommandFlagSet is the one credential-free validation boundary for
// declared command inputs. Runtime mapping helpers use the returned declaration
// map, so the public preflight and executor request builders cannot drift on
// requiredness, multiplicity, unknown flags, types, enums, bounds, formats, or
// encoded-size limits.
func validateCommandFlagSet(cmd connectors.CommandSurfaceCommand, flags map[string][]string) (map[string]connectors.CommandSurfaceFlag, error) {
	return validateCommandFlagSetForPreflight(cmd, flags, false)
}

func validateCommandFlagSetForPreflight(cmd connectors.CommandSurfaceCommand, flags map[string][]string, allowOmittedRequired bool) (map[string]connectors.CommandSurfaceFlag, error) {
	allowed := make(map[string]connectors.CommandSurfaceFlag, len(cmd.Flags))
	for _, flag := range cmd.Flags {
		if err := safety.ValidateIdentifier(flag.Name, "flag name"); err != nil {
			return nil, err
		}
		if _, duplicate := allowed[flag.Name]; duplicate {
			return nil, &BlockedCommandError{Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("flag --%s is declared more than once", flag.Name)}
		}
		allowed[flag.Name] = flag
	}
	if err := validateCanonicalFlagOccurrences(cmd, flags); err != nil {
		return nil, err
	}
	if !allowOmittedRequired {
		if err := validateRequiredCommandFlags(cmd, flags); err != nil {
			return nil, err
		}
	}
	names := make([]string, 0, len(flags))
	for name := range flags {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		values := flags[name]
		if len(values) == 0 {
			continue
		}
		if err := safety.ValidateIdentifier(name, "flag name"); err != nil {
			return nil, err
		}
		flag, ok := allowed[name]
		if !ok {
			return nil, fmt.Errorf("unknown flag --%s for command %q", name, cmd.Path)
		}
		if _, err := coerceCommandFlagValue(cmd, flag, values); err != nil {
			return nil, err
		}
	}
	return allowed, nil
}

func validateRequiredCommandFlags(cmd connectors.CommandSurfaceCommand, flags map[string][]string) error {
	for _, flag := range cmd.Flags {
		if !flag.Required {
			continue
		}
		if err := safety.ValidateIdentifier(flag.Name, "flag name"); err != nil {
			return err
		}
		values, ok := flags[flag.Name]
		if !ok || len(values) == 0 {
			return missingRequiredFlagError(cmd, flag.Name)
		}
		value, err := coerceCommandFlagValue(cmd, flag, values)
		if err != nil {
			return err
		}
		if commandValueEmpty(value) {
			return missingRequiredFlagError(cmd, flag.Name)
		}
	}
	return nil
}

func commandValueEmpty(value any) bool {
	switch typed := value.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(typed) == ""
	case []string:
		// Required-flag raw presence is established before coercion. Array
		// cardinality is enforced by coerceFlagValue's min_items validation, so
		// an explicitly supplied, zero-minimum array may legitimately be empty.
		return false
	default:
		return false
	}
}

func missingRequiredFlagError(cmd connectors.CommandSurfaceCommand, name string) error {
	return &MissingRequiredFlagError{Command: cmd.Path, Flag: name}
}

// ReconstituteWithheldFields rebuilds the record fragment for fields a reverse
// plan withheld from disk. Values are read from the same command flags the plan
// accepted and coerced by the same rules, so a reconstituted record hashes
// identically to the record the plan bound. Fields with no supplied value are
// returned as operator-facing flag names rather than silently dropped.
//
// A declared sensitive field is not always a flag target. It can equally be an
// ancestor of several flag targets -- recurly's create_invoice_retry declares
// account.billing_infos while its command maps four separate leaves beneath it
// -- and withholding removes that whole subtree, which is the correct reading of
// the declaration. Such a field is rebuilt from its descendant flags in the same
// target order recordOverrides applies them, so the subtree is byte-identical to
// the one the plan hashed.
func ReconstituteWithheldFields(connector connectors.Connector, path []string, fields []string, flags map[string][]string) (connectors.Record, []string, error) {
	if len(fields) == 0 {
		return connectors.Record{}, nil, nil
	}
	cmd, _, err := resolvePreflightCommand(connector, path)
	if err != nil {
		return nil, nil, err
	}
	// A direct_write command's record IS its body, so its fields are declared
	// as body.<path>; a reverse_etl command declares record.<path>. Dispatch on
	// mode rather than trying both, so a plan can never resolve a field through
	// the other namespace.
	prefix := "record."
	if strings.TrimSpace(cmd.Operation) != "" {
		prefix = "body."
	}
	byTarget := map[string]connectors.CommandSurfaceFlag{}
	for _, flag := range cmd.Flags {
		if target, ok := strings.CutPrefix(flag.MapsTo, prefix); ok && target != "" {
			byTarget[target] = flag
		}
	}
	if strings.TrimSpace(cmd.Operation) != "" {
		metadataProvider, ok := connector.(connectors.OperationDirectWriteMetadataProvider)
		if !ok {
			return nil, nil, fmt.Errorf("connector %q does not expose direct-write metadata for operation %q", connector.Name(), cmd.Operation)
		}
		metadata, err := metadataProvider.OperationDirectWriteMetadata(cmd.Operation)
		if err != nil {
			return nil, nil, err
		}
		if metadata.StructuredBody {
			return reconstituteStructuredDirectWriteFields(connector, cmd, byTarget, fields, flags)
		}
	}
	record := connectors.Record{}
	missing := make([]string, 0, len(fields))
	for _, field := range fields {
		target := strings.TrimPrefix(strings.TrimSpace(field), prefix)
		if target == "" {
			continue
		}
		if flag, ok := byTarget[target]; ok {
			values := flags[flag.Name]
			if len(values) == 0 {
				missing = append(missing, "--"+flag.Name)
				continue
			}
			value, err := coerceRecordFlagValue(flag, values)
			if err != nil {
				return nil, nil, err
			}
			if err := setRecordValue(record, target, value); err != nil {
				return nil, nil, err
			}
			continue
		}
		unresolved, err := reconstituteWithheldSubtree(record, byTarget, target, flags)
		if err != nil {
			return nil, nil, err
		}
		missing = append(missing, unresolved...)
	}
	return record, missing, nil
}

func reconstituteStructuredDirectWriteFields(connector connectors.Connector, cmd connectors.CommandSurfaceCommand, byTarget map[string]connectors.CommandSurfaceFlag, fields []string, flags map[string][]string) (connectors.Record, []string, error) {
	materializer, ok := connector.(connectors.OperationDirectWriteBodyMaterializer)
	if !ok {
		return nil, nil, fmt.Errorf("connector %q does not expose direct-write body materialization", connector.Name())
	}
	transformer, ok := connector.(connectors.OperationDirectWriteBodyPlanTransformer)
	if !ok {
		return nil, nil, fmt.Errorf("connector %q does not expose direct-write body plan transformation", connector.Name())
	}
	mappings := map[string]any{}
	missing := make([]string, 0)
	add := func(target string, flag connectors.CommandSurfaceFlag) error {
		values := flags[flag.Name]
		if len(values) == 0 {
			missing = append(missing, "--"+flag.Name)
			return nil
		}
		value, err := coerceRecordFlagValue(flag, values)
		if err != nil {
			return err
		}
		mappings[target] = value
		return nil
	}
	for _, raw := range fields {
		target := strings.TrimPrefix(strings.TrimSpace(raw), "body.")
		if target == "" {
			continue
		}
		if flag, found := byTarget[target]; found {
			if err := add(target, flag); err != nil {
				return nil, nil, err
			}
			continue
		}
		ancestor := ""
		for candidate := range byTarget {
			contains, err := transformer.OperationDirectWriteBodyPathContains(cmd.Operation, candidate, target)
			if err != nil {
				return nil, nil, err
			}
			if contains && (ancestor == "" || len(candidate) > len(ancestor)) {
				ancestor = candidate
			}
		}
		if ancestor != "" {
			if err := add(ancestor, byTarget[ancestor]); err != nil {
				return nil, nil, err
			}
			continue
		}
		missing = append(missing, target)
	}
	if len(mappings) == 0 {
		sort.Strings(missing)
		return connectors.Record{}, missing, nil
	}
	body, err := materializer.MaterializeOperationDirectWriteBody(cmd.Operation, mappings)
	if err != nil {
		return nil, nil, err
	}
	sort.Strings(missing)
	return connectors.Record(body), missing, nil
}

// reconstituteWithheldSubtree rebuilds a withheld field that no single flag maps
// to, from the flags that map beneath it. Targets are applied in the sorted
// order recordOverrides uses, because setDottedValue refuses a sparse array
// index and so depends on that order to build the same shape.
//
// An optional descendant with no supplied value is skipped rather than demanded:
// recordOverrides skipped it too when the plan was built, so demanding it would
// strand the plan the same way the declared-but-unsupplied case once did. A
// required descendant is always demanded, because the plan could not have been
// created without it.
func reconstituteWithheldSubtree(record connectors.Record, byTarget map[string]connectors.CommandSurfaceFlag, target string, flags map[string][]string) ([]string, error) {
	descendants := make([]string, 0, len(byTarget))
	for candidate := range byTarget {
		if dottedPathPrefix(target, candidate) {
			descendants = append(descendants, candidate)
		}
	}
	if len(descendants) == 0 {
		return []string{target}, nil
	}
	sort.Strings(descendants)
	missing := make([]string, 0, len(descendants))
	applied := 0
	for _, descendant := range descendants {
		flag := byTarget[descendant]
		values := flags[flag.Name]
		if len(values) == 0 {
			if flag.Required {
				missing = append(missing, "--"+flag.Name)
			}
			continue
		}
		value, err := coerceRecordFlagValue(flag, values)
		if err != nil {
			return nil, err
		}
		if err := setRecordValue(record, descendant, value); err != nil {
			return nil, err
		}
		applied++
	}
	if applied == 0 && len(missing) == 0 {
		for _, descendant := range descendants {
			missing = append(missing, "--"+byTarget[descendant].Name)
		}
	}
	return missing, nil
}

func validateFlagValue(flag connectors.CommandSurfaceFlag, value string) error {
	trimmed := strings.TrimSpace(value)
	if flag.AllowEmpty != nil {
		if !*flag.AllowEmpty && trimmed == "" {
			return fmt.Errorf("invalid --%s %q, want non-empty value", flag.Name, value)
		}
		if *flag.AllowEmpty && trimmed == "" {
			return nil
		}
	}
	if err := validateFlagFormat(flag, trimmed); err != nil {
		return err
	}
	switch flag.Type {
	case "", "string", "boolean", "integer", "number", "string_array":
		return validateFlagNumericRange(flag, value)
	case "enum":
		for _, allowed := range flag.Values {
			if value == allowed {
				return validateFlagNumericRange(flag, value)
			}
		}
		values := append([]string(nil), flag.Values...)
		sort.Strings(values)
		return fmt.Errorf("invalid --%s %q, want one of %s", flag.Name, value, strings.Join(values, "|"))
	default:
		return &BlockedCommandError{
			Command: "unknown",
			Reason:  fmt.Sprintf("flag --%s has unsupported type %q", flag.Name, flag.Type),
		}
	}
}

func validateFlagNumericRange(flag connectors.CommandSurfaceFlag, value string) error {
	if err := validateFlagMinimum(flag, value); err != nil {
		return err
	}
	return validateFlagMaximum(flag, value)
}

func validateFlagMinimum(flag connectors.CommandSurfaceFlag, value string) error {
	if flag.Minimum == nil {
		return nil
	}
	minimum, ok := parseExactJSONNumber(flag.Minimum.String())
	if !ok {
		return &BlockedCommandError{Command: "unknown", Reason: fmt.Sprintf("flag --%s has invalid minimum", flag.Name)}
	}
	if flag.Type != "integer" && flag.Type != "number" {
		return &BlockedCommandError{Command: "unknown", Reason: fmt.Sprintf("flag --%s minimum requires integer or number type", flag.Name)}
	}
	parsed, ok := parseExactFlagNumber(flag.Type, value)
	if !ok {
		return &MinimumFlagError{Parameter: flag.Name, Minimum: *flag.Minimum}
	}
	if parsed.Cmp(minimum) < 0 {
		return &MinimumFlagError{Parameter: flag.Name, Minimum: *flag.Minimum}
	}
	return nil
}

func validateFlagMaximum(flag connectors.CommandSurfaceFlag, value string) error {
	if flag.Maximum == nil {
		return nil
	}
	maximum, ok := parseExactJSONNumber(flag.Maximum.String())
	if !ok {
		return &BlockedCommandError{Command: "unknown", Reason: fmt.Sprintf("flag --%s has invalid maximum", flag.Name)}
	}
	if flag.Type != "integer" && flag.Type != "number" {
		return &BlockedCommandError{Command: "unknown", Reason: fmt.Sprintf("flag --%s maximum requires integer or number type", flag.Name)}
	}
	parsed, ok := parseExactFlagNumber(flag.Type, value)
	if !ok {
		return &MaximumFlagError{Parameter: flag.Name, Maximum: *flag.Maximum}
	}
	if parsed.Cmp(maximum) > 0 {
		return &MaximumFlagError{Parameter: flag.Name, Maximum: *flag.Maximum}
	}
	return nil
}

func parseExactFlagNumber(flagType, value string) (*big.Rat, bool) {
	switch flagType {
	case "integer":
		integer, ok := parseExactJSONInteger(value)
		if !ok {
			return nil, false
		}
		return new(big.Rat).SetInt(integer), true
	case "number":
		return parseExactJSONNumber(value)
	default:
		return nil, false
	}
}

func validateFlagFormat(flag connectors.CommandSurfaceFlag, value string) error {
	switch flag.Format {
	case "":
		return nil
	case "date-time":
		if _, err := time.Parse(time.RFC3339, value); err != nil {
			return fmt.Errorf("invalid --%s %q, want ISO-8601/RFC3339 timestamp", flag.Name, value)
		}
		return nil
	default:
		return &BlockedCommandError{
			Command: "unknown",
			Reason:  fmt.Sprintf("flag --%s has unsupported format %q", flag.Name, flag.Format),
		}
	}
}

func directReadOverrides(cmd connectors.CommandSurfaceCommand, flags map[string][]string) (map[string]string, map[string]string, error) {
	allowed, err := validateCommandFlagSet(cmd, flags)
	if err != nil {
		return nil, nil, err
	}

	pathParams := map[string]string{}
	query := map[string]string{}
	for name, values := range flags {
		if len(values) == 0 {
			continue
		}
		if err := safety.ValidateIdentifier(name, "flag name"); err != nil {
			return nil, nil, err
		}
		flag, ok := allowed[name]
		if !ok {
			return nil, nil, fmt.Errorf("unknown flag --%s for command %q", name, cmd.Path)
		}
		value := values[len(values)-1]
		if err := safety.RejectDangerousChars(value, "flag value"); err != nil {
			return nil, nil, err
		}
		if err := validateFlagValue(flag, value); err != nil {
			return nil, nil, err
		}
		switch {
		case strings.HasPrefix(flag.MapsTo, "path."):
			target := strings.TrimPrefix(flag.MapsTo, "path.")
			if err := safety.ValidateIdentifier(target, "path parameter"); err != nil {
				return nil, nil, err
			}
			pathParams[target] = value
		case strings.HasPrefix(flag.MapsTo, "query."):
			target := strings.TrimPrefix(flag.MapsTo, "query.")
			if err := safety.ValidateIdentifier(target, "query parameter"); err != nil {
				return nil, nil, err
			}
			query[target] = value
		default:
			return nil, nil, &BlockedCommandError{
				Command:      cmd.Path,
				Intent:       cmd.Intent,
				Availability: cmd.Availability,
				Reason:       fmt.Sprintf("flag --%s maps to unsupported target %q", name, flag.MapsTo),
			}
		}
	}
	return pathParams, query, nil
}

// operationDirectReadOverrides shapes the explicitly declared request inputs
// shared by operation-backed reads and writes. A literal body is intentionally
// narrower than the dotted JSON body mapping: only an operation direct read
// may name the exact maps_to value "body", and the engine subsequently admits
// it only for its declared text/plain root-string contract.
func operationDirectReadOverrides(cmd connectors.CommandSurfaceCommand, flags map[string][]string) (map[string]string, map[string]string, map[string]string, map[string][]string, map[string]any, *string, error) {
	return operationDirectOverrides(cmd, flags, nil)
}

func operationDirectWriteOverrides(cmd connectors.CommandSurfaceCommand, flags map[string][]string, materialize func(string, map[string]any) (map[string]any, error)) (map[string]string, map[string]string, map[string]string, map[string][]string, map[string]any, *string, error) {
	return operationDirectOverrides(cmd, flags, materialize)
}

func operationDirectOverrides(cmd connectors.CommandSurfaceCommand, flags map[string][]string, materialize func(string, map[string]any) (map[string]any, error)) (map[string]string, map[string]string, map[string]string, map[string][]string, map[string]any, *string, error) {
	allowed, err := validateCommandFlagSet(cmd, flags)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	pathParams := map[string]string{}
	query := map[string]string{}
	headers := map[string]string{}
	headerValues := map[string][]string{}
	headerTargets := map[string]struct{}{}
	body := map[string]any{}
	var rawBody *string
	type bodyMapping struct {
		path  string
		value any
	}
	bodyMappings := make([]bodyMapping, 0)
	names := make([]string, 0, len(flags))
	for name := range flags {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		values := flags[name]
		if len(values) == 0 {
			continue
		}
		if err := safety.ValidateIdentifier(name, "flag name"); err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}
		flag, ok := allowed[name]
		if !ok {
			return nil, nil, nil, nil, nil, nil, fmt.Errorf("unknown flag --%s for command %q", name, cmd.Path)
		}
		value, err := coerceCommandFlagValue(cmd, flag, values)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}
		switch {
		case strings.HasPrefix(flag.MapsTo, "path."):
			target := strings.TrimPrefix(flag.MapsTo, "path.")
			if err := safety.ValidateIdentifier(target, "path parameter"); err != nil {
				return nil, nil, nil, nil, nil, nil, err
			}
			pathParams[target] = stringifyCommandValue(value)
		case strings.HasPrefix(flag.MapsTo, "query."):
			target := strings.TrimPrefix(flag.MapsTo, "query.")
			if err := safety.ValidateIdentifier(target, "query parameter"); err != nil {
				return nil, nil, nil, nil, nil, nil, err
			}
			query[target] = stringifyCommandValue(value)
		case strings.HasPrefix(flag.MapsTo, "header."):
			target := strings.TrimPrefix(flag.MapsTo, "header.")
			if err := validateOperationHeaderTarget(target); err != nil {
				return nil, nil, nil, nil, nil, nil, err
			}
			canonical, err := connectors.CanonicalOperationHeaderName(target)
			if err != nil {
				return nil, nil, nil, nil, nil, nil, err
			}
			if _, duplicate := headerTargets[canonical]; duplicate {
				return nil, nil, nil, nil, nil, nil, fmt.Errorf("command %q maps multiple flags to request header %q", cmd.Path, target)
			}
			if flag.Repeatable {
				if flag.Type != "" && flag.Type != "string" && flag.Type != "enum" {
					return nil, nil, nil, nil, nil, nil, &BlockedCommandError{Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("flag --%s maps to a repeatable header but is not a string", name)}
				}
				headerTargets[canonical] = struct{}{}
				headerValues[target] = append([]string(nil), values...)
				continue
			}
			text, ok := value.(string)
			if !ok {
				return nil, nil, nil, nil, nil, nil, &BlockedCommandError{Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("flag --%s maps to header but is not a string", name)}
			}
			headerTargets[canonical] = struct{}{}
			headers[target] = text
		case strings.HasPrefix(flag.MapsTo, "body."):
			target := strings.TrimPrefix(flag.MapsTo, "body.")
			bodyMappings = append(bodyMappings, bodyMapping{path: target, value: value})
		case flag.MapsTo == "body":
			if cmd.Intent != "direct_read" || flag.Type != "string" {
				return nil, nil, nil, nil, nil, nil, &BlockedCommandError{Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("flag --%s maps to unsupported target %q", name, flag.MapsTo)}
			}
			text, ok := value.(string)
			if !ok {
				return nil, nil, nil, nil, nil, nil, &BlockedCommandError{Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("flag --%s maps to body but is not a string", name)}
			}
			rawBody = &text
		default:
			return nil, nil, nil, nil, nil, nil, &BlockedCommandError{Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("flag --%s maps to unsupported target %q", name, flag.MapsTo)}
		}
	}
	if materialize != nil {
		mappings := make(map[string]any, len(bodyMappings))
		for _, mapping := range bodyMappings {
			if _, duplicate := mappings[mapping.path]; duplicate {
				return nil, nil, nil, nil, nil, nil, &BlockedCommandError{Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("multiple flags map to body.%s", mapping.path)}
			}
			mappings[mapping.path] = mapping.value
		}
		var err error
		body, err = materialize(cmd.Operation, mappings)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}
	} else {
		sort.SliceStable(bodyMappings, func(left, right int) bool {
			return bodyMappingPathLess(bodyMappings[left].path, bodyMappings[right].path)
		})
		for _, mapping := range bodyMappings {
			if err := setOperationBodyValue(body, mapping.path, mapping.value, false); err != nil {
				return nil, nil, nil, nil, nil, nil, err
			}
		}
	}
	if rawBody != nil && len(bodyMappings) != 0 {
		return nil, nil, nil, nil, nil, nil, &BlockedCommandError{Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "raw body cannot mix with JSON body fields"}
	}
	return pathParams, query, headers, headerValues, body, rawBody, nil
}

// validateOperationHeaderTarget does not decide whether a header is allowed —
// the engine does that against the selected operation and runtime-owned names.
// It only prevents a hand-authored command surface from smuggling a malformed
// field name across the commandrunner/engine boundary.
func validateOperationHeaderTarget(target string) error {
	_, err := connectors.CanonicalOperationHeaderName(target)
	return err
}

func bodyMappingPathLess(left, right string) bool {
	leftParts := strings.Split(left, ".")
	rightParts := strings.Split(right, ".")
	limit := len(leftParts)
	if len(rightParts) < limit {
		limit = len(rightParts)
	}
	for index := 0; index < limit; index++ {
		leftPart := leftParts[index]
		rightPart := rightParts[index]
		if leftPart == rightPart {
			continue
		}
		leftIndex, leftNumeric := bodyMappingArrayIndex(leftPart)
		rightIndex, rightNumeric := bodyMappingArrayIndex(rightPart)
		if leftNumeric && rightNumeric {
			return leftIndex < rightIndex
		}
		return leftPart < rightPart
	}
	return len(leftParts) < len(rightParts)
}

func bodyMappingArrayIndex(part string) (uint64, bool) {
	if part == "" || (len(part) > 1 && strings.HasPrefix(part, "0")) {
		return 0, false
	}
	for _, r := range part {
		if r < '0' || r > '9' {
			return 0, false
		}
	}
	index, err := strconv.ParseUint(part, 10, 64)
	if err != nil {
		return 0, false
	}
	return index, true
}

func validateCanonicalFlagOccurrences(cmd connectors.CommandSurfaceCommand, flags map[string][]string) error {
	allowed := make(map[string]connectors.CommandSurfaceFlag, len(cmd.Flags))
	for _, flag := range cmd.Flags {
		allowed[flag.Name] = flag
	}
	names := make([]string, 0, len(flags))
	for name := range flags {
		names = append(names, name)
	}
	sort.Strings(names)
	seen := make(map[string]string)
	for _, name := range names {
		values := flags[name]
		if len(values) == 0 {
			continue
		}
		flag, ok := allowed[name]
		if !ok {
			continue
		}
		target, err := normalizedCommandFlagTarget(flag)
		if err != nil {
			return &BlockedCommandError{Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: err.Error()}
		}
		if len(values) > 1 && !flag.Repeatable && flag.Type != "string_array" {
			return &BlockedCommandError{Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("target %q must accept exactly one value; it was supplied more than once via --%s", target, name)}
		}
		if previous, duplicate := seen[target]; duplicate {
			return &BlockedCommandError{Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("target %q must accept exactly one value; it was supplied more than once via --%s and --%s", target, previous, name)}
		}
		seen[target] = name
	}
	return nil
}

func normalizedCommandFlagTarget(flag connectors.CommandSurfaceFlag) (string, error) {
	target := strings.TrimSpace(flag.MapsTo)
	if target == "" {
		return "", fmt.Errorf("flag --%s has no request target", flag.Name)
	}
	if header, ok := strings.CutPrefix(target, "header."); ok {
		canonical, err := connectors.CanonicalOperationHeaderName(header)
		if err != nil {
			return "", err
		}
		return "header." + canonical, nil
	}
	return target, nil
}

func setBodyValue(body map[string]any, path string, value any) error {
	return setOperationBodyValue(body, path, value, false)
}

func setOperationBodyValue(body map[string]any, path string, value any, declarationBoundArrays bool) error {
	parts, err := validateDottedTargetPath(path, "body field")
	if err != nil {
		return err
	}
	maxArrayIndex := maxRecordPathArrayIndex
	if declarationBoundArrays {
		maxArrayIndex = 0
	}
	_, err = setDottedValue(body, parts, value, path, maxArrayIndex)
	return err
}

func validateDottedTargetPath(path, field string) ([]string, error) {
	parts := strings.Split(path, ".")
	for _, part := range parts {
		if err := safety.ValidateIdentifier(part, field); err != nil {
			return nil, err
		}
	}
	return parts, nil
}

func setDottedValue(current any, parts []string, value any, fullPath string, maxArrayIndex int) (any, error) {
	if len(parts) == 0 {
		return value, nil
	}
	part := parts[0]
	if index, ok, err := pathArrayIndex(part, maxArrayIndex); err != nil {
		return nil, err
	} else if ok {
		items, ok := current.([]any)
		if !ok {
			return nil, fmt.Errorf("body field %q conflicts with existing non-array value", fullPath)
		}
		if index > len(items) {
			return nil, fmt.Errorf("body field %q uses sparse array index %d", fullPath, index)
		}
		if index == len(items) {
			items = append(items, nil)
		}
		if len(parts) == 1 {
			items[index] = value
			return items, nil
		}
		child := items[index]
		if child == nil {
			child = newDottedContainer(parts[1], maxArrayIndex)
		}
		updated, err := setDottedValue(child, parts[1:], value, fullPath, maxArrayIndex)
		if err != nil {
			return nil, err
		}
		items[index] = updated
		return items, nil
	}

	object, ok := current.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("body field %q conflicts with existing non-object value", fullPath)
	}
	if len(parts) == 1 {
		object[part] = value
		return object, nil
	}
	child, ok := object[part]
	if !ok {
		child = newDottedContainer(parts[1], maxArrayIndex)
	}
	updated, err := setDottedValue(child, parts[1:], value, fullPath, maxArrayIndex)
	if err != nil {
		return nil, err
	}
	object[part] = updated
	return object, nil
}

func newDottedContainer(nextPart string, maxArrayIndex int) any {
	if _, ok, _ := pathArrayIndex(nextPart, maxArrayIndex); ok {
		return []any{}
	}
	return map[string]any{}
}

func pathArrayIndex(part string, maxArrayIndex int) (int, bool, error) {
	for _, r := range part {
		if r < '0' || r > '9' {
			return 0, false, nil
		}
	}
	if len(part) > 1 && strings.HasPrefix(part, "0") {
		return 0, false, fmt.Errorf("body field array index %q must not have leading zeroes", part)
	}
	index, err := strconv.Atoi(part)
	if err != nil {
		return 0, false, fmt.Errorf("body field array index %q is invalid", part)
	}
	if maxArrayIndex > 0 && index > maxArrayIndex {
		return 0, false, fmt.Errorf("body field array index %d exceeds max %d", index, maxArrayIndex)
	}
	return index, true, nil
}

func stringifyCommandValue(value any) string {
	if s, ok := value.(string); ok {
		return s
	}
	return fmt.Sprint(value)
}

func recordOverrides(cmd connectors.CommandSurfaceCommand, flags map[string][]string) (connectors.Record, error) {
	allowed, err := validateCommandFlagSet(cmd, flags)
	if err != nil {
		return nil, err
	}
	if err := validateRecordFlagTargets(cmd); err != nil {
		return nil, err
	}
	type flagApplication struct {
		name   string
		flag   connectors.CommandSurfaceFlag
		values []string
		target string
	}
	applications := make([]flagApplication, 0, len(flags))
	for name, values := range flags {
		if len(values) == 0 {
			continue
		}
		if err := safety.ValidateIdentifier(name, "flag name"); err != nil {
			return nil, err
		}
		flag, ok := allowed[name]
		if !ok {
			return nil, fmt.Errorf("unknown flag --%s for command %q", name, cmd.Path)
		}
		target, ok := strings.CutPrefix(flag.MapsTo, "record.")
		if !ok || target == "" {
			return nil, &BlockedCommandError{
				Command:      cmd.Path,
				Intent:       cmd.Intent,
				Availability: cmd.Availability,
				Reason:       fmt.Sprintf("flag --%s maps to unsupported target %q", name, flag.MapsTo),
			}
		}
		applications = append(applications, flagApplication{name: name, flag: flag, values: values, target: target})
	}
	sort.Slice(applications, func(i, j int) bool {
		if applications[i].target == applications[j].target {
			return applications[i].name < applications[j].name
		}
		return applications[i].target < applications[j].target
	})
	record := connectors.Record{}
	for _, app := range applications {
		value, err := coerceRecordFlagValue(app.flag, app.values)
		if err != nil {
			return nil, err
		}
		if err := setRecordValue(record, app.target, value); err != nil {
			return nil, err
		}
	}
	return record, nil
}

func validateRecordFlagTargets(cmd connectors.CommandSurfaceCommand) error {
	targets := map[string]string{}
	for _, flag := range cmd.Flags {
		target, ok := strings.CutPrefix(flag.MapsTo, "record.")
		if !ok || target == "" {
			continue
		}
		parts, err := validateDottedTargetPath(target, "record field")
		if err != nil {
			return err
		}
		normalized := strings.Join(parts, ".")
		if prior := targets[normalized]; prior != "" {
			return fmt.Errorf("flags --%s and --%s both map to record.%s", prior, flag.Name, normalized)
		}
		for existing, prior := range targets {
			if dottedPathPrefix(existing, normalized) || dottedPathPrefix(normalized, existing) {
				return fmt.Errorf("flags --%s and --%s have conflicting record mappings", prior, flag.Name)
			}
		}
		targets[normalized] = flag.Name
	}
	return nil
}

func dottedPathPrefix(parent, child string) bool {
	return strings.HasPrefix(child, parent+".")
}

func setRecordValue(record connectors.Record, path string, value any) error {
	return setBodyValue(map[string]any(record), path, value)
}

// coerceRecordFlagValue is the only route that admits the declarative `json`
// flag kind. It is intentionally unavailable to direct reads, direct writes,
// paths, queries, headers, and arbitrary body fields; preflight has already
// tied it to one engine-validated reverse-ETL record property.
func coerceRecordFlagValue(flag connectors.CommandSurfaceFlag, values []string) (any, error) {
	if flag.Type == "json" {
		return coerceDeclaredStructuredJSONRecordFlagValue(flag, values)
	}
	return coerceFlagValue(flag, values)
}

func coerceDeclaredStructuredJSONRecordFlagValue(flag connectors.CommandSurfaceFlag, values []string) (any, error) {
	if len(values) != 1 {
		return nil, fmt.Errorf("invalid --%s: structured JSON flags accept exactly one value", flag.Name)
	}
	raw := values[0]
	maxBytes := maxStructuredJSONFlagBytes
	if flag.MaxBytes > 0 && flag.MaxBytes < maxBytes {
		maxBytes = flag.MaxBytes
	}
	if len(raw) > maxBytes {
		return nil, fmt.Errorf("invalid --%s: structured JSON exceeds %d bytes", flag.Name, maxBytes)
	}
	if err := safety.RejectDangerousChars(raw, "flag value"); err != nil {
		return nil, err
	}
	if flag.AllowBareString && !structuredJSONRecordValueStartsContainer(raw) {
		// The declaration preflight has proved this exact named field has a
		// string arm. Preserve normal command-line text (including text that
		// resembles a JSON scalar) while object/array values keep their strict
		// JSON syntax and all values still face record-schema validation.
		return raw, nil
	}

	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("invalid JSON for --%s: %w", flag.Name, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("invalid JSON for --%s: must contain exactly one value", flag.Name)
		}
		return nil, fmt.Errorf("invalid JSON for --%s: %w", flag.Name, err)
	}
	// Runtime preflight has already proved the exact named record property is
	// structured or a provider-declared multi-kind union. Do not impose a
	// second object/array-only rule here: it would silently remove a scalar arm
	// from that declared union. This remains a named record value, never a raw
	// request body, and is still bounded above before decoding.
	return value, nil
}

func structuredJSONRecordValueStartsContainer(raw string) bool {
	trimmed := strings.TrimSpace(raw)
	return strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[")
}

// coerceCommandFlagValue retains the normal command-line control-character
// guard everywhere except a literal text/plain operation body. Markdown and
// similar declared text documents need line breaks; they are never used as a
// URL, header, or JSON field because only the exact body mapping reaches this
// path. Other control characters remain refused.
func coerceCommandFlagValue(cmd connectors.CommandSurfaceCommand, flag connectors.CommandSurfaceFlag, values []string) (any, error) {
	if (cmd.Intent == "reverse_etl" || cmd.Intent == "binary_upload") && strings.HasPrefix(flag.MapsTo, "record.") {
		return coerceRecordFlagValue(flag, values)
	}
	if flag.Type == "json" && isDeclaredStructuredJSONOperationBodyFlag(cmd, flag) {
		return coerceDeclaredStructuredJSONRecordFlagValue(flag, values)
	}
	if cmd.Intent == "direct_read" && flag.MapsTo == "body" {
		return coerceDeclaredPlainTextBodyFlagValue(flag, values)
	}
	return coerceFlagValue(flag, values)
}

// isDeclaredStructuredJSONOperationBodyFlag mirrors the non-network preflight
// shape before admitting JSON parsing. Runtime has already tied the exact
// field to either a fixed GraphQL variable or a closed REST body property.
// Keeping this syntactic guard here ensures an internal caller cannot route a
// raw or nested path through the structured parser.
func isDeclaredStructuredJSONOperationBodyFlag(cmd connectors.CommandSurfaceCommand, flag connectors.CommandSurfaceFlag) bool {
	if (cmd.Intent != "direct_read" && cmd.Intent != "direct_write") || strings.TrimSpace(cmd.Operation) == "" {
		return false
	}
	variable, ok := strings.CutPrefix(flag.MapsTo, "body.")
	return ok && variable != ""
}

func coerceDeclaredPlainTextBodyFlagValue(flag connectors.CommandSurfaceFlag, values []string) (any, error) {
	if flag.Type != "string" {
		return nil, &BlockedCommandError{Command: "unknown", Reason: fmt.Sprintf("flag --%s maps to raw body but is not a string", flag.Name)}
	}
	for _, value := range values {
		for _, r := range value {
			if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
				return nil, fmt.Errorf("flag value contains unsupported control character for declared text body")
			}
		}
	}
	value := values[len(values)-1]
	if err := validateFlagValue(flag, value); err != nil {
		return nil, err
	}
	return value, nil
}

func coerceFlagValue(flag connectors.CommandSurfaceFlag, values []string) (any, error) {
	clean := make([]string, 0, len(values))
	for _, value := range values {
		if err := safety.RejectDangerousChars(value, "flag value"); err != nil {
			return nil, err
		}
		if err := validateCommandFlagEncodedBytes(flag, value); err != nil {
			return nil, err
		}
		if err := validateFlagValue(flag, value); err != nil {
			return nil, err
		}
		clean = append(clean, value)
	}
	value := clean[len(clean)-1]
	switch flag.Type {
	case "", "string", "enum":
		return value, nil
	case "boolean":
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("invalid --%s %q, want boolean", flag.Name, value)
		}
		return parsed, nil
	case "integer":
		if _, ok := parseExactJSONInteger(value); !ok {
			return nil, fmt.Errorf("invalid --%s %q, want integer", flag.Name, value)
		}
		return json.Number(value), nil
	case "number":
		if _, ok := parseExactJSONNumber(value); !ok {
			return nil, fmt.Errorf("invalid --%s %q, want finite number", flag.Name, value)
		}
		return json.Number(value), nil
	case "string_array":
		out := make([]string, 0)
		for _, raw := range clean {
			for _, item := range strings.Split(raw, ",") {
				item = strings.TrimSpace(item)
				if item != "" {
					out = append(out, item)
				}
			}
		}
		// Bounded here as well as in the body schema: the schema fires on the
		// assembled body, this fires on the flag the user typed, so the error can
		// name it.
		if flag.MaxItems > 0 && len(out) > flag.MaxItems {
			return nil, fmt.Errorf("invalid --%s: %d values exceeds the maximum of %d", flag.Name, len(out), flag.MaxItems)
		}
		if flag.MinItems > 0 && len(out) < flag.MinItems {
			return nil, fmt.Errorf("invalid --%s: %d values is below the minimum of %d", flag.Name, len(out), flag.MinItems)
		}
		if len(flag.Values) != 0 {
			allowed := make(map[string]struct{}, len(flag.Values))
			for _, value := range flag.Values {
				allowed[value] = struct{}{}
			}
			for _, item := range out {
				if _, ok := allowed[item]; !ok {
					return nil, fmt.Errorf("invalid --%s value %q", flag.Name, item)
				}
			}
		}
		return out, nil
	default:
		return nil, &BlockedCommandError{
			Command: "unknown",
			Reason:  fmt.Sprintf("flag --%s has unsupported type %q", flag.Name, flag.Type),
		}
	}
}

// parseExactJSONNumber accepts exactly one JSON numeric lexeme and represents
// it as a rational for comparisons. Unlike ParseFloat it preserves provider
// identifiers and decimal coefficients past 53 bits, including exponent
// forms, until the sealed JSON encoder sends the same lexeme onward.
func parseExactJSONNumber(value string) (*big.Rat, bool) {
	if !json.Valid([]byte(value)) {
		return nil, false
	}
	rational, ok := new(big.Rat).SetString(value)
	return rational, ok
}

func parseExactJSONInteger(value string) (*big.Int, bool) {
	if !json.Valid([]byte(value)) {
		return nil, false
	}
	integer, ok := new(big.Int).SetString(value, 10)
	return integer, ok
}

func validateCommandFlagEncodedBytes(flag connectors.CommandSurfaceFlag, value string) error {
	if flag.MaxBytes <= 0 {
		return nil
	}
	location, name, ok := strings.Cut(strings.TrimSpace(flag.MapsTo), ".")
	if !ok || name == "" {
		return nil
	}
	encoded := ""
	switch location {
	case "record":
		encoded = value
	case "config":
		encoded = value
	case "path":
		if name == "path" || name == "ref" {
			parts := strings.Split(value, "/")
			for index := range parts {
				parts[index] = url.PathEscape(parts[index])
			}
			encoded = strings.Join(parts, "/")
		} else {
			encoded = url.PathEscape(value)
		}
	case "query":
		encoded = url.QueryEscape(value)
	default:
		return nil
	}
	if len(encoded) > flag.MaxBytes {
		return fmt.Errorf("flag --%s encoded %s value exceeds byte cap %d", flag.Name, location, flag.MaxBytes)
	}
	return nil
}

// ConfirmationChallengeForCommand reports the typed confirmation a command's
// bound executor will demand at run, resolved from the same declarations
// buildWriteCommand and buildOperationDirectWriteCommand read. It takes no
// record and validates nothing, so help and documentation callers can answer
// "will this ask me to confirm?" without restating the rule or building a plan.
//
// A bundle's cli_surface notes cannot serve that purpose: they are prose an
// author writes per command, so they are silent on every command nobody
// annotated, and silence there is indistinguishable from "no confirmation".
func ConfirmationChallengeForCommand(connector connectors.Connector, cmd connectors.CommandSurfaceCommand) string {
	switch cmd.Intent {
	case "reverse_etl", "binary_upload":
		action, ok := findWriteAction(connectors.ManifestOf(connector), cmd.Write)
		if !ok {
			return ""
		}
		return string(connectors.ConfirmationForWriteAction(action).Kind)
	case "direct_write":
		provider, ok := connector.(connectors.OperationDirectWriteMetadataProvider)
		if !ok || strings.TrimSpace(cmd.Operation) == "" {
			return ""
		}
		metadata, err := provider.OperationDirectWriteMetadata(cmd.Operation)
		if err != nil {
			return ""
		}
		return metadata.ConfirmationChallenge
	default:
		return ""
	}
}

func findWriteAction(manifest connectors.Manifest, name string) (connectors.WriteActionSpec, bool) {
	for _, action := range manifest.WriteActions {
		if action.Name == name {
			return action, true
		}
	}
	return connectors.WriteActionSpec{}, false
}

func mutationClassOf(action connectors.WriteActionSpec) string {
	switch strings.ToUpper(strings.TrimSpace(action.Method)) {
	case http.MethodPost:
		return "create"
	case http.MethodPut, http.MethodPatch:
		return "update"
	case http.MethodDelete:
		return "delete"
	default:
		return "write"
	}
}

func targetResourceOf(cmd connectors.CommandSurfaceCommand) string {
	fields := strings.Fields(cmd.Path)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func cloneRecord(in connectors.Record) connectors.Record {
	out := make(connectors.Record, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func cloneStringSliceMap(in map[string][]string) map[string][]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string][]string, len(in))
	for key, values := range in {
		out[key] = append([]string(nil), values...)
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func isAbsoluteHTTPURL(raw string) bool {
	lower := strings.ToLower(strings.TrimSpace(raw))
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

// validateBinaryDownloadCommand mirrors, for the binary_download intent, what
// validateOperationDirectReadCommand does for direct reads: it refuses metadata
// the executor cannot honour BEFORE any network or filesystem access.
//
// The endpoint must be a single connector-relative GET. Unlike a direct read,
// no output_policy applies: the response never becomes a JSON body, it becomes
// a file, and the record that describes it carries no response bytes.
func validateBinaryDownloadCommand(connector connectors.Connector, cmd connectors.CommandSurfaceCommand) error {
	if _, ok := connector.(connectors.OperationBinaryDownloader); !ok {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "connector does not support binary downloads"}
	}
	if strings.TrimSpace(cmd.Operation) == "" {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "binary_download commands require operation"}
	}
	if len(cmd.APISurface) != 1 {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "binary_download commands require exactly one api_surface endpoint"}
	}
	method := strings.ToUpper(strings.TrimSpace(cmd.APISurface[0].Method))
	if method != http.MethodGet {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("binary_download commands require GET api_surface endpoints, got %s", method)}
	}
	if isAbsoluteHTTPURL(cmd.APISurface[0].Path) {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "binary_download commands must not reference an absolute URL"}
	}
	if preflighter, ok := connector.(connectors.OperationBinaryDownloadPreflighter); ok {
		if err := preflighter.PreflightOperationBinaryDownload(cmd.Operation, method, cmd.APISurface[0].Path); err != nil {
			return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("binary download metadata is not executable: %v", err)}
		}
	}
	return nil
}

// runBinaryDownload executes a binary_download command. The destination is
// caller-supplied and never inferred: without an explicit --dest-root the
// command is refused rather than defaulting to the working directory.
func runBinaryDownload(ctx context.Context, connector connectors.Connector, cmd connectors.CommandSurfaceCommand, req Request) (Result, error) {
	downloader, ok := connector.(connectors.OperationBinaryDownloader)
	if !ok {
		return Result{}, &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "connector does not support binary downloads"}
	}
	if err := validateBinaryDownloadCommand(connector, cmd); err != nil {
		return Result{}, err
	}
	if strings.TrimSpace(req.DestRoot) == "" {
		return Result{}, fmt.Errorf("binary download requires --dest-root")
	}
	pathParams, query, headers, headerValues, _, _, err := operationDirectReadOverrides(cmd, req.Flags)
	if err != nil {
		return Result{}, err
	}
	if err := validateCommandInputs(cmd, req.Config, mappedCommandInputs{Query: query}); err != nil {
		return Result{}, err
	}
	download, err := downloader.OperationBinaryDownload(ctx, connectors.OperationBinaryDownloadRequest{
		Operation:    cmd.Operation,
		Config:       req.Config,
		PathParams:   pathParams,
		Query:        query,
		Headers:      headers,
		HeaderValues: headerValues,
		MaxBytes:     int64(req.MaxBytes),
		DestRoot:     req.DestRoot,
		FileName:     req.FileName,
	})
	download = connectors.SanitizeOperationBinaryDownloadResultForOutput(download, req.Config.Secrets)
	if err != nil {
		if !binaryDownloadHasProviderEvidence(download) {
			return Result{}, err
		}
		return Result{Connector: connector.Name(), Command: cmd.Path, BinaryDownload: &download}, err
	}
	return Result{Connector: connector.Name(), Command: cmd.Path, BinaryDownload: &download}, nil
}

func validateStatusCheckCommand(connector connectors.Connector, cmd connectors.CommandSurfaceCommand) error {
	checker, ok := connector.(connectors.OperationStatusChecker)
	if !ok || checker == nil {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "connector does not support HEAD status checks"}
	}
	preflighter, ok := connector.(connectors.OperationStatusCheckPreflighter)
	if !ok || strings.TrimSpace(cmd.Operation) == "" || len(cmd.APISurface) != 1 {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "status_check commands require one declared operation endpoint"}
	}
	endpoint := cmd.APISurface[0]
	if !strings.EqualFold(endpoint.Method, http.MethodHead) || isAbsoluteHTTPURL(endpoint.Path) || cmd.OutputPolicy != "status" {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "status_check commands require a connector-relative HEAD endpoint and status output"}
	}
	if err := preflighter.PreflightOperationStatusCheck(cmd.Operation, endpoint.Method, endpoint.Path, cmd.OutputPolicy); err != nil {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("status check metadata is not executable: %v", err)}
	}
	return nil
}

func runStatusCheck(ctx context.Context, connector connectors.Connector, cmd connectors.CommandSurfaceCommand, req Request) (Result, error) {
	if err := validateStatusCheckCommand(connector, cmd); err != nil {
		return Result{}, err
	}
	pathParams, query, headers, headerValues, _, _, err := operationDirectReadOverrides(cmd, req.Flags)
	if err != nil {
		return Result{}, err
	}
	if err := validateCommandInputs(cmd, req.Config, mappedCommandInputs{Query: query}); err != nil {
		return Result{}, err
	}
	status, err := connector.(connectors.OperationStatusChecker).OperationStatusCheck(ctx, connectors.OperationStatusCheckRequest{Operation: cmd.Operation, Config: req.Config, PathParams: pathParams, Query: query, Headers: headers, HeaderValues: headerValues})
	status = connectors.SanitizeOperationStatusCheckResultForOutput(status, req.Config.Secrets)
	if err != nil {
		if !statusCheckHasProviderEvidence(status) {
			return Result{}, err
		}
		return Result{Connector: connector.Name(), Command: cmd.Path, StatusCheck: &status}, err
	}
	return Result{Connector: connector.Name(), Command: cmd.Path, StatusCheck: &status}, nil
}

func directReadHasProviderEvidence(result connectors.DirectReadResult) bool {
	return result.Status != 0 || (result.Receipt != nil && result.Receipt.ResponseReceived)
}

func binaryDownloadHasProviderEvidence(result connectors.OperationBinaryDownloadResult) bool {
	return result.Status != 0 || (result.Receipt != nil && result.Receipt.ResponseReceived)
}

func statusCheckHasProviderEvidence(result connectors.OperationStatusCheckResult) bool {
	return result.Status != 0 || (result.Receipt != nil && result.Receipt.ResponseReceived)
}
