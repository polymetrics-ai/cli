package commandrunner

import (
	"context"
	"fmt"
	"net/http"
	"sort"
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
}

type Result struct {
	Connector  string                       `json:"connector"`
	Command    string                       `json:"command"`
	Stream     string                       `json:"stream,omitempty"`
	Count      int                          `json:"count,omitempty"`
	DirectRead *connectors.DirectReadResult `json:"direct_read,omitempty"`
}

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
	_, _, err := resolveRunnableCommand(connector, path)
	return err
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
	case "github_contents_file_metadata", "github_contents_directory":
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
	case "", "string", "boolean":
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

func isAbsoluteHTTPURL(raw string) bool {
	lower := strings.ToLower(strings.TrimSpace(raw))
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}
