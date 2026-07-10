package commandrunner

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/safety"
)

const MaxDirectReadBytes = 1 << 20

type Request struct {
	Path     []string
	Flags    map[string][]string
	Config   connectors.RuntimeConfig
	Limit    int
	MaxBytes int
	Preview  bool
}

type Result struct {
	Connector  string                       `json:"connector"`
	Command    string                       `json:"command"`
	Stream     string                       `json:"stream,omitempty"`
	Count      int                          `json:"count,omitempty"`
	DirectRead *connectors.DirectReadResult `json:"direct_read,omitempty"`
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
		RedactedRecord:        redactRecord(record),
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
	if cmd.Intent != "etl" || cmd.Availability != "implemented" || cmd.Stream == "" {
		return Result{}, &BlockedCommandError{
			Connector:    connector.Name(),
			Command:      command,
			Intent:       cmd.Intent,
			Availability: cmd.Availability,
			Reason:       blockReason(cmd),
		}
	}

	query, err := queryOverrides(cmd, req.Flags)
	if err != nil {
		return Result{}, err
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}
	result := Result{Connector: connector.Name(), Command: command, Stream: cmd.Stream}
	readReq := connectors.ReadRequest{
		Stream: cmd.Stream,
		Config: req.Config,
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
	if cmd.Operation != "" {
		return connectors.CommandSurfaceCommand{}, command, &BlockedCommandError{
			Connector:    connector.Name(),
			Command:      command,
			Intent:       cmd.Intent,
			Availability: cmd.Availability,
			Reason:       fmt.Sprintf("operation %s executor is not implemented in this slice", cmd.Operation),
		}
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
	if err := validateDirectReadCommand(connector, cmd); err != nil {
		return Result{}, err
	}
	pathParams, query, err := directReadOverrides(cmd, req.Flags)
	if err != nil {
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
	})
	if err != nil {
		return Result{}, err
	}
	return Result{
		Connector:  connector.Name(),
		Command:    cmd.Path,
		DirectRead: &direct,
	}, nil
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

func isSupportedDirectReadOutputPolicy(policy string) bool {
	switch policy {
	case "github_contents_file_metadata", "github_contents_directory", "json_response":
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

func queryOverrides(cmd connectors.CommandSurfaceCommand, flags map[string][]string) (map[string]string, error) {
	allowed := map[string]connectors.CommandSurfaceFlag{}
	for _, flag := range cmd.Flags {
		if err := safety.ValidateIdentifier(flag.Name, "flag name"); err != nil {
			return nil, err
		}
		allowed[flag.Name] = flag
	}

	query := map[string]string{}
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
		value := values[len(values)-1]
		if err := safety.RejectDangerousChars(value, "flag value"); err != nil {
			return nil, err
		}
		if err := validateFlagValue(flag, value); err != nil {
			return nil, err
		}
		target, ok := strings.CutPrefix(flag.MapsTo, "query.")
		if !ok || target == "" {
			return nil, &BlockedCommandError{
				Command:      cmd.Path,
				Intent:       cmd.Intent,
				Availability: cmd.Availability,
				Reason:       fmt.Sprintf("flag --%s maps to unsupported target %q", name, flag.MapsTo),
			}
		}
		if err := safety.ValidateIdentifier(target, "query parameter"); err != nil {
			return nil, err
		}
		query[target] = value
	}
	return query, nil
}

func validateFlagValue(flag connectors.CommandSurfaceFlag, value string) error {
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

func directReadOverrides(cmd connectors.CommandSurfaceCommand, flags map[string][]string) (map[string]string, map[string]string, error) {
	allowed := map[string]connectors.CommandSurfaceFlag{}
	for _, flag := range cmd.Flags {
		if err := safety.ValidateIdentifier(flag.Name, "flag name"); err != nil {
			return nil, nil, err
		}
		allowed[flag.Name] = flag
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

func recordOverrides(cmd connectors.CommandSurfaceCommand, flags map[string][]string) (connectors.Record, error) {
	allowed := map[string]connectors.CommandSurfaceFlag{}
	for _, flag := range cmd.Flags {
		if err := safety.ValidateIdentifier(flag.Name, "flag name"); err != nil {
			return nil, err
		}
		allowed[flag.Name] = flag
	}
	record := connectors.Record{}
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
		if err := safety.ValidateIdentifier(target, "record field"); err != nil {
			return nil, err
		}
		value, err := coerceFlagValue(flag, values)
		if err != nil {
			return nil, err
		}
		record[target] = value
	}
	return record, nil
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
	out := make(connectors.Record, len(in))
	for k, v := range in {
		if isSensitiveRecordField(k) {
			out[k] = "***"
			continue
		}
		out[k] = v
	}
	return out
}

func isSensitiveRecordField(name string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(name, "-", "_"))
	for _, marker := range []string{"token", "secret", "password", "private_key", "api_key", "key", "body", "comment", "content", "payload", "inputs"} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
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
