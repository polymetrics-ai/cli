package commandrunner

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/safety"
)

const (
	MaxDirectReadBytes          = 1 << 20
	MaxOperationDirectReadBytes = 16 << 20
	maxRecordPathArrayIndex     = 128
)

type Request struct {
	Path     []string
	Flags    map[string][]string
	Config   connectors.RuntimeConfig
	Limit    int
	MaxBytes int
	Preview  bool
	// DestRoot is the directory a binary_download command writes beneath.
	// Required for that intent and ignored by every other one.
	DestRoot string
	// FileName optionally names the downloaded file within DestRoot.
	FileName string
}

type Result struct {
	Connector      string                                    `json:"connector"`
	Command        string                                    `json:"command"`
	Stream         string                                    `json:"stream,omitempty"`
	Count          int                                       `json:"count,omitempty"`
	DirectRead     *connectors.DirectReadResult              `json:"direct_read,omitempty"`
	BinaryDownload *connectors.OperationBinaryDownloadResult `json:"binary_download,omitempty"`
}

type WriteCommand struct {
	Connector             string                   `json:"connector"`
	Command               string                   `json:"command"`
	Write                 string                   `json:"write"`
	MutationClass         string                   `json:"mutation_class"`
	TargetResource        string                   `json:"target_resource"`
	ApprovalRequired      bool                     `json:"approval_required"`
	Risk                  string                   `json:"risk,omitempty"`
	Approval              string                   `json:"approval,omitempty"`
	ConfirmationChallenge string                   `json:"confirmation_challenge,omitempty"`
	Record                connectors.Record        `json:"record,omitempty"`
	RedactedRecord        connectors.Record        `json:"redacted_record,omitempty"`
	Preview               *connectors.WritePreview `json:"preview,omitempty"`
}

var ErrNotWriteCommand = errors.New("connector command is not a reverse ETL write command")

type BlockedCommandError struct {
	Connector    string
	Command      string
	Intent       string
	Availability string
	Reason       string
}

func (e *BlockedCommandError) Error() string {
	parts := []string{fmt.Sprintf("connector command %q is blocked", e.Command)}
	if e.Intent != "" {
		parts = append(parts, "intent="+e.Intent)
	}
	if e.Availability != "" {
		parts = append(parts, "availability="+e.Availability)
	}
	if e.Reason != "" {
		parts = append(parts, e.Reason)
	}
	return strings.Join(parts, ": ")
}

func Preflight(connector connectors.Connector, path []string) error {
	_, _, err := resolvePreflightCommand(connector, path)
	return err
}

func BuildWriteCommand(ctx context.Context, connector connectors.Connector, req Request) (WriteCommand, error) {
	cmd, command, err := resolvePreflightCommand(connector, req.Path)
	if err != nil {
		return WriteCommand{}, err
	}
	if cmd.Intent != "reverse_etl" {
		return WriteCommand{}, ErrNotWriteCommand
	}
	if cmd.Availability != "implemented" || cmd.Write == "" {
		return WriteCommand{}, &BlockedCommandError{
			Connector:    connector.Name(),
			Command:      command,
			Intent:       cmd.Intent,
			Availability: cmd.Availability,
			Reason:       "implemented reverse ETL commands must reference write action",
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
		Write:                 cmd.Write,
		MutationClass:         mutationClassOf(action),
		TargetResource:        targetResourceOf(cmd),
		ApprovalRequired:      true,
		Risk:                  firstNonEmpty(cmd.Risk, action.Risk),
		Approval:              firstNonEmpty(cmd.Approval, "reverse ETL writes require plan, preview, approval, execute"),
		ConfirmationChallenge: strings.TrimSpace(action.Confirm),
		Record:                cloneRecord(record),
		RedactedRecord:        redactRecordWithFields(record, cmd.RedactFields),
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

func Run(ctx context.Context, connector connectors.Connector, req Request, emit func(connectors.Record) error) (Result, error) {
	cmd, command, err := resolveRunnableCommand(connector, req.Path)
	if err != nil {
		return Result{}, err
	}
	if cmd.Intent == "direct_read" {
		return runDirectRead(ctx, connector, cmd, req)
	}
	if cmd.Intent == "binary_download" {
		return runBinaryDownload(ctx, connector, cmd, req)
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
		return emit(redactRecordWithFields(record, cmd.RedactFields))
	}))
	if err := connectors.IgnoreReadLimit(err); err != nil {
		return Result{}, redactCommandError(err, cmd.RedactFields, req)
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
	if cmd.Intent == "binary_download" && cmd.Availability == "implemented" {
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
	if cmd.Operation != "" && cmd.Intent != "binary_download" && (cmd.Intent != "direct_read" || cmd.Availability != "implemented") {
		return connectors.CommandSurfaceCommand{}, command, &BlockedCommandError{
			Connector:    connector.Name(),
			Command:      command,
			Intent:       cmd.Intent,
			Availability: cmd.Availability,
			Reason:       fmt.Sprintf("operation %s executor is not implemented in this slice", cmd.Operation),
		}
	}
	if cmd.Intent == "binary_download" && cmd.Availability == "implemented" {
		if err := validateBinaryDownloadCommand(connector, cmd); err != nil {
			return connectors.CommandSurfaceCommand{}, command, err
		}
		return cmd, command, nil
	}
	if cmd.Intent == "direct_read" && cmd.Availability == "implemented" && cmd.Operation != "" {
		if err := validateOperationDirectReadCommand(connector, cmd); err != nil {
			return connectors.CommandSurfaceCommand{}, command, err
		}
		return cmd, command, nil
	}
	if cmd.Intent == "direct_read" && cmd.Availability == "implemented" {
		if err := validateDirectReadCommand(connector, cmd); err != nil {
			return connectors.CommandSurfaceCommand{}, command, err
		}
		return cmd, command, nil
	}
	if cmd.Intent == "etl" && cmd.Availability == "implemented" && cmd.Stream != "" {
		return cmd, command, nil
	}
	if cmd.Intent == "reverse_etl" && cmd.Availability == "implemented" && cmd.Write != "" {
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

func runDirectRead(ctx context.Context, connector connectors.Connector, cmd connectors.CommandSurfaceCommand, req Request) (Result, error) {
	if cmd.Operation != "" {
		return runOperationDirectRead(ctx, connector, cmd, req)
	}
	if err := validateDirectReadCommand(connector, cmd); err != nil {
		return Result{}, err
	}
	pathParams, query, err := directReadOverrides(cmd, req.Flags)
	if err != nil {
		return Result{}, redactCommandError(err, cmd.RedactFields, req)
	}
	if err := validateCommandInputs(cmd, req.Config, mappedCommandInputs{Query: query}); err != nil {
		return Result{}, redactCommandError(err, cmd.RedactFields, req)
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
		RedactFields: cmd.RedactFields,
	})
	if err != nil {
		return Result{}, redactCommandError(err, cmd.RedactFields, req)
	}
	return Result{
		Connector:  connector.Name(),
		Command:    cmd.Path,
		DirectRead: &direct,
	}, nil
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
	bodySchema, err := operationRequestBodySchema(connector, cmd)
	if err != nil {
		return Result{}, err
	}
	pathParams, query, body, err := operationDirectReadOverrides(cmd, req.Flags, bodySchema)
	if err != nil {
		return Result{}, redactCommandError(err, cmd.RedactFields, req)
	}
	if err := validateCommandInputs(cmd, req.Config, mappedCommandInputs{Query: query, Body: body}); err != nil {
		return Result{}, redactCommandError(err, cmd.RedactFields, req)
	}
	maxBytes := req.MaxBytes
	if maxBytes <= 0 {
		maxBytes = MaxOperationDirectReadBytes
	}
	if maxBytes > MaxOperationDirectReadBytes {
		maxBytes = MaxOperationDirectReadBytes
	}
	direct, err := reader.OperationDirectRead(ctx, connectors.OperationDirectReadRequest{
		Operation:    cmd.Operation,
		Config:       req.Config,
		PathParams:   pathParams,
		Query:        query,
		Body:         body,
		MaxBytes:     maxBytes,
		OutputPolicy: cmd.OutputPolicy,
		RedactFields: cmd.RedactFields,
	})
	if err != nil {
		return Result{}, redactCommandError(err, cmd.RedactFields, req)
	}
	return Result{Connector: connector.Name(), Command: cmd.Path, DirectRead: &direct}, nil
}

type operationRequestBodySchemaProvider interface {
	OperationRequestBodySchema(string) (*engine.Schema, error)
}

func operationRequestBodySchema(connector connectors.Connector, cmd connectors.CommandSurfaceCommand) (*engine.Schema, error) {
	needsSchema := false
	for _, flag := range cmd.Flags {
		namespace, _, err := engine.ParseRequestFieldPointer(flag.MapsTo)
		if err == nil && namespace == "body" {
			needsSchema = true
			break
		}
	}
	if !needsSchema {
		return nil, nil
	}
	provider, ok := connector.(operationRequestBodySchemaProvider)
	if !ok {
		return nil, &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "operation direct_read body mappings require the operation body schema"}
	}
	schema, err := provider.OperationRequestBodySchema(cmd.Operation)
	if err != nil {
		return nil, &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("operation direct_read body schema is unavailable: %v", err)}
	}
	if schema == nil {
		return nil, &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "operation direct_read body mappings require a declared body_schema"}
	}
	return schema, nil
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
	return nil
}

func validateOperationDirectReadCommand(connector connectors.Connector, cmd connectors.CommandSurfaceCommand) error {
	if _, ok := connector.(connectors.OperationDirectReader); !ok {
		return &BlockedCommandError{Connector: connector.Name(), Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: "connector does not support operation direct reads"}
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
	return nil
}

func isSupportedDirectReadOutputPolicy(policy string) bool {
	switch policy {
	case "repository_contents_file_metadata", "repository_contents_directory", "json_redacted", "clinical_json_redacted":
		return true
	default:
		return false
	}
}

func commandPath(path []string) string {
	return strings.Join(path, " ")
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
	case cmd.Operation != "":
		return fmt.Sprintf("operation %s executor is not implemented in this slice", cmd.Operation)
	case cmd.Intent == "reverse_etl" && cmd.Write == "":
		return "implemented reverse ETL commands must reference write action"
	case cmd.Intent == "reverse_etl":
		if cmd.Approval != "" {
			return cmd.Approval
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
	allowed := map[string]connectors.CommandSurfaceFlag{}
	for _, flag := range cmd.Flags {
		if err := safety.ValidateIdentifier(flag.Name, "flag name"); err != nil {
			return connectors.RuntimeConfig{}, nil, err
		}
		allowed[flag.Name] = flag
	}
	if err := validateRequiredCommandFlags(cmd, flags); err != nil {
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
			if err := safety.ValidateQueryParameterName(target, "query parameter"); err != nil {
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
	return runtimeConfigWithOverrides(cfg, configOverrides), query, nil
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

type mappedCommandInputs struct {
	Query map[string]string
	Body  map[string]any
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
	default:
		return &BlockedCommandError{Command: "unknown", Reason: fmt.Sprintf("unsupported command constraint kind %q", constraint.Kind)}
	}
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
	case strings.HasPrefix(target, "/"):
		namespace, tokens, err := engine.ParseRequestFieldPointer(target)
		if err != nil {
			return "", false, target, &BlockedCommandError{Command: "unknown", Reason: fmt.Sprintf("invalid command validation target %q: %v", target, err)}
		}
		switch namespace {
		case "query":
			value, present := inputs.Query[tokens[0]]
			return strings.TrimSpace(value), present, target, nil
		case "body":
			value, present := nestedBodyPointerValue(inputs.Body, tokens)
			return strings.TrimSpace(fmt.Sprint(value)), present, target, nil
		default:
			return "", false, target, &BlockedCommandError{Command: "unknown", Reason: fmt.Sprintf("unsupported command validation target %q", target)}
		}
	case strings.HasPrefix(target, "query."):
		key := strings.TrimPrefix(target, "query.")
		value, present := inputs.Query[key]
		return strings.TrimSpace(value), present, target, nil
	case strings.HasPrefix(target, "body."):
		value, present := nestedBodyValue(inputs.Body, strings.TrimPrefix(target, "body."))
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
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = m[part]
		if !ok || cur == nil {
			return nil, false
		}
	}
	return cur, true
}

func nestedBodyPointerValue(body map[string]any, tokens []string) (any, bool) {
	if body == nil || len(tokens) == 0 {
		return nil, false
	}
	var current any = body
	for _, token := range tokens {
		switch value := current.(type) {
		case map[string]any:
			var ok bool
			current, ok = value[token]
			if !ok || current == nil {
				return nil, false
			}
		case []any:
			index, ok, err := pathArrayIndex(token)
			if err != nil || !ok || index >= len(value) {
				return nil, false
			}
			current = value[index]
		default:
			return nil, false
		}
	}
	return current, true
}

func parseDateTimeValue(value, label string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid %s, want ISO-8601/RFC3339 timestamp", label)
	}
	return parsed, nil
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
		value, err := coerceFlagValue(flag, values)
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
		return len(typed) == 0
	default:
		return false
	}
}

func missingRequiredFlagError(cmd connectors.CommandSurfaceCommand, name string) error {
	return fmt.Errorf("missing required flag --%s for command %q", name, cmd.Path)
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
	case "", "string", "boolean", "integer", "string_array":
		return nil
	case "enum":
		for _, allowed := range flag.Values {
			if value == allowed {
				return nil
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
	allowed := map[string]connectors.CommandSurfaceFlag{}
	for _, flag := range cmd.Flags {
		if err := safety.ValidateIdentifier(flag.Name, "flag name"); err != nil {
			return nil, nil, err
		}
		allowed[flag.Name] = flag
	}
	if err := validateRequiredCommandFlags(cmd, flags); err != nil {
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
			if err := safety.ValidateQueryParameterName(target, "query parameter"); err != nil {
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

func operationDirectReadOverrides(cmd connectors.CommandSurfaceCommand, flags map[string][]string, bodySchema *engine.Schema) (map[string]string, map[string]string, map[string]any, error) {
	allowed := map[string]connectors.CommandSurfaceFlag{}
	for _, flag := range cmd.Flags {
		if err := safety.ValidateIdentifier(flag.Name, "flag name"); err != nil {
			return nil, nil, nil, err
		}
		allowed[flag.Name] = flag
	}
	if err := validateRequiredCommandFlags(cmd, flags); err != nil {
		return nil, nil, nil, err
	}

	pathParams := map[string]string{}
	query := map[string]string{}
	body := map[string]any{}
	names := make([]string, 0, len(flags))
	for name := range flags {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		left := allowed[names[i]].MapsTo
		right := allowed[names[j]].MapsTo
		if left == right {
			return names[i] < names[j]
		}
		return requestPointerLess(left, right)
	})
	for _, name := range names {
		values := flags[name]
		if len(values) == 0 {
			continue
		}
		if err := safety.ValidateIdentifier(name, "flag name"); err != nil {
			return nil, nil, nil, err
		}
		flag, ok := allowed[name]
		if !ok {
			return nil, nil, nil, fmt.Errorf("unknown flag --%s for command %q", name, cmd.Path)
		}
		value, err := coerceFlagValue(flag, values)
		if err != nil {
			return nil, nil, nil, err
		}
		namespace, tokens, err := engine.ParseRequestFieldPointer(flag.MapsTo)
		if err != nil {
			return nil, nil, nil, &BlockedCommandError{Command: cmd.Path, Intent: cmd.Intent, Availability: cmd.Availability, Reason: fmt.Sprintf("flag --%s maps to unsupported target %q", name, flag.MapsTo)}
		}
		switch namespace {
		case "path":
			pathParams[tokens[0]] = stringifyCommandValue(value)
		case "query":
			query[tokens[0]] = stringifyCommandValue(value)
		case "body":
			if err := setRequestBodyValue(bodySchema, body, flag.MapsTo, value); err != nil {
				return nil, nil, nil, err
			}
		}
	}
	return pathParams, query, body, nil
}

func requestPointerLess(left, right string) bool {
	leftNamespace, leftTokens, leftErr := engine.ParseRequestFieldPointer(left)
	rightNamespace, rightTokens, rightErr := engine.ParseRequestFieldPointer(right)
	if leftErr != nil || rightErr != nil {
		return left < right
	}
	if leftNamespace != rightNamespace {
		return leftNamespace < rightNamespace
	}
	limit := len(leftTokens)
	if len(rightTokens) < limit {
		limit = len(rightTokens)
	}
	for i := 0; i < limit; i++ {
		leftIndex, leftNumeric, leftIndexErr := pathArrayIndex(leftTokens[i])
		rightIndex, rightNumeric, rightIndexErr := pathArrayIndex(rightTokens[i])
		if leftIndexErr == nil && rightIndexErr == nil && leftNumeric && rightNumeric && leftIndex != rightIndex {
			return leftIndex < rightIndex
		}
		if leftTokens[i] != rightTokens[i] {
			return leftTokens[i] < rightTokens[i]
		}
	}
	return len(leftTokens) < len(rightTokens)
}

func setRequestBodyValue(schema *engine.Schema, body map[string]any, pointer string, value any) error {
	if schema == nil {
		return fmt.Errorf("body field %q requires body_schema", pointer)
	}
	return schema.SetRequestBodyPointer(body, pointer, value)
}

func setBodyValue(body map[string]any, path string, value any) error {
	parts, err := validateDottedTargetPath(path, "body field")
	if err != nil {
		return err
	}
	_, err = setDottedValue(body, parts, value, path)
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

func setDottedValue(current any, parts []string, value any, fullPath string) (any, error) {
	if len(parts) == 0 {
		return value, nil
	}
	part := parts[0]
	if index, ok, err := pathArrayIndex(part); err != nil {
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
			child = newDottedContainer(parts[1])
		}
		updated, err := setDottedValue(child, parts[1:], value, fullPath)
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
		child = newDottedContainer(parts[1])
	}
	updated, err := setDottedValue(child, parts[1:], value, fullPath)
	if err != nil {
		return nil, err
	}
	object[part] = updated
	return object, nil
}

func newDottedContainer(nextPart string) any {
	if _, ok, _ := pathArrayIndex(nextPart); ok {
		return []any{}
	}
	return map[string]any{}
}

func pathArrayIndex(part string) (int, bool, error) {
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
	if index > maxRecordPathArrayIndex {
		return 0, false, fmt.Errorf("body field array index %d exceeds max %d", index, maxRecordPathArrayIndex)
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
	allowed := map[string]connectors.CommandSurfaceFlag{}
	for _, flag := range cmd.Flags {
		if err := safety.ValidateIdentifier(flag.Name, "flag name"); err != nil {
			return nil, err
		}
		allowed[flag.Name] = flag
	}
	if err := validateRequiredCommandFlags(cmd, flags); err != nil {
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
		value, err := coerceFlagValue(app.flag, app.values)
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

func coerceFlagValue(flag connectors.CommandSurfaceFlag, values []string) (any, error) {
	clean := make([]string, 0, len(values))
	for _, value := range values {
		if err := safety.RejectDangerousChars(value, "flag value"); err != nil {
			return nil, err
		}
		clean = append(clean, value)
	}
	value := clean[len(clean)-1]
	if err := validateFlagValue(flag, value); err != nil {
		return nil, err
	}
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
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("invalid --%s %q, want integer", flag.Name, value)
		}
		return parsed, nil
	case "string_array":
		var out []string
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
		return out, nil
	default:
		return nil, &BlockedCommandError{
			Command: "unknown",
			Reason:  fmt.Sprintf("flag --%s has unsupported type %q", flag.Name, flag.Type),
		}
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

func redactRecord(in connectors.Record) connectors.Record {
	return redactRecordWithFields(in, nil)
}

func redactCommandError(err error, fields []string, req Request) error {
	if err == nil {
		return nil
	}
	text := safety.RedactErrorText(err.Error())
	for _, values := range req.Flags {
		for _, value := range values {
			text = redactLiteral(text, value)
		}
	}
	explicit := explicitRedactFieldSet(fields)
	for key, value := range req.Config.Config {
		if explicit[normalizeRecordFieldName(key)] || explicit[compactRecordFieldName(key)] || isSensitiveRecordField(key) || strings.Contains(compactRecordFieldName(key), "patient") {
			text = redactLiteral(text, value)
		}
	}
	return errors.New(text)
}

func redactLiteral(text, value string) string {
	value = strings.TrimSpace(value)
	if value == "" || value == "***" {
		return text
	}
	return strings.ReplaceAll(text, value, "***")
}

func redactRecordWithFields(in connectors.Record, fields []string) connectors.Record {
	explicit := explicitRedactFieldSet(fields)
	out := make(connectors.Record, len(in))
	for k, v := range in {
		out[k] = redactValueForField(k, v, explicit)
	}
	return out
}

func redactValueForField(field string, value any, explicit map[string]bool) any {
	if isSensitiveRecordField(field) || explicit[normalizeRecordFieldName(field)] || explicit[compactRecordFieldName(field)] {
		return "***"
	}
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for k, v := range typed {
			out[k] = redactValueForField(k, v, explicit)
		}
		return out
	case connectors.Record:
		return redactRecordWithFields(typed, redactFieldsFromSet(explicit))
	case []any:
		out := make([]any, len(typed))
		for i, v := range typed {
			out[i] = redactValueForField("", v, explicit)
		}
		return out
	case []connectors.Record:
		out := make([]connectors.Record, len(typed))
		for i, v := range typed {
			out[i] = redactRecordWithFields(v, redactFieldsFromSet(explicit))
		}
		return out
	default:
		return value
	}
}

func redactFieldsFromSet(explicit map[string]bool) []string {
	out := make([]string, 0, len(explicit))
	for field := range explicit {
		out = append(out, field)
	}
	return out
}

func explicitRedactFieldSet(fields []string) map[string]bool {
	out := make(map[string]bool, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		out[normalizeRecordFieldName(field)] = true
		out[compactRecordFieldName(field)] = true
	}
	return out
}

func isSensitiveRecordField(name string) bool {
	normalized := normalizeRecordFieldName(name)
	for _, marker := range []string{"token", "secret", "password", "private_key", "api_key", "key", "body", "comment", "content", "payload", "inputs", "download", "clone", "media_url", "data_file", "media_file", "file_path"} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func normalizeRecordFieldName(name string) string {
	return strings.ToLower(strings.NewReplacer("-", "_", " ", "_", ".", "_").Replace(name))
}

func compactRecordFieldName(name string) string {
	return strings.ReplaceAll(normalizeRecordFieldName(name), "_", "")
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
	pathParams, query, err := directReadOverrides(cmd, req.Flags)
	if err != nil {
		return Result{}, redactCommandError(err, cmd.RedactFields, req)
	}
	if err := validateCommandInputs(cmd, req.Config, mappedCommandInputs{Query: query}); err != nil {
		return Result{}, redactCommandError(err, cmd.RedactFields, req)
	}
	download, err := downloader.OperationBinaryDownload(ctx, connectors.OperationBinaryDownloadRequest{
		Operation:    cmd.Operation,
		Config:       req.Config,
		PathParams:   pathParams,
		Query:        query,
		MaxBytes:     int64(req.MaxBytes),
		DestRoot:     req.DestRoot,
		FileName:     req.FileName,
		RedactFields: cmd.RedactFields,
	})
	if err != nil {
		return Result{}, redactCommandError(err, cmd.RedactFields, req)
	}
	return Result{Connector: connector.Name(), Command: cmd.Path, BinaryDownload: &download}, nil
}
