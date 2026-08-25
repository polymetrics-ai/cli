package engine

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"polymetrics.ai/internal/connectors"
)

var commandBindingTemplateRE = regexp.MustCompile(`\{\{\s*(?:config|record|fanout)\.([-A-Za-z0-9_]+)(?:\s*\|[^}]*)?\s*\}\}`)

// ResolvedCommandBinding is the one declaration and provider target selected
// by the runtime for an implemented CLI command.
type ResolvedCommandBinding struct {
	Binding     connectors.CommandBindingIdentity
	Method      string
	Path        string
	Destructive DestructiveTarget
}

// ResolveImplementedCommandBinding is shared by runtime preflight and source
// admission. It is the sole lane-to-declaration resolver; certification must
// not copy its own incomplete stream/write/operation switch.
func ResolveImplementedCommandBinding(b Bundle, cmd connectors.CommandSurfaceCommand) (ResolvedCommandBinding, error) {
	bindings := 0
	for _, value := range []string{cmd.Stream, cmd.Write, cmd.Operation} {
		if strings.TrimSpace(value) != "" {
			bindings++
		}
	}
	if bindings > 1 {
		return ResolvedCommandBinding{}, fmt.Errorf("implemented command %q references more than one runtime binding", cmd.Path)
	}

	var resolved ResolvedCommandBinding
	switch {
	case cmd.Stream != "":
		stream, ok := commandBindingStream(b, cmd.Stream)
		if !ok {
			return ResolvedCommandBinding{}, fmt.Errorf("implemented command %q references missing stream %q", cmd.Path, cmd.Stream)
		}
		method := strings.ToUpper(strings.TrimSpace(stream.Method))
		if method == "" {
			method = http.MethodGet
		}
		resolved = ResolvedCommandBinding{
			Binding: connectors.CommandBindingIdentity{Kind: connectors.CommandBindingStream, ID: stream.Name},
			Method:  method, Path: normalizeBindingPathTemplate(stream.Path),
		}
	case cmd.Write != "":
		action, ok := commandBindingWrite(b, cmd.Write)
		if !ok {
			return ResolvedCommandBinding{}, fmt.Errorf("implemented command %q references missing write action %q", cmd.Path, cmd.Write)
		}
		resolved = ResolvedCommandBinding{
			Binding:     connectors.CommandBindingIdentity{Kind: connectors.CommandBindingWrite, ID: action.Name},
			Method:      strings.ToUpper(strings.TrimSpace(action.Method)),
			Path:        normalizeBindingPathTemplate(action.Path),
			Destructive: DestructiveTargetForWrite(b.Name, action),
		}
	case cmd.Operation != "":
		operation, ok := commandBindingOperation(b, cmd.Operation)
		if !ok {
			return ResolvedCommandBinding{}, fmt.Errorf("implemented command %q references missing operation %q", cmd.Path, cmd.Operation)
		}
		method, path, err := commandBindingOperationEndpoint(operation)
		if err != nil {
			return ResolvedCommandBinding{}, err
		}
		resolved = ResolvedCommandBinding{
			Binding:     connectors.CommandBindingIdentity{Kind: connectors.CommandBindingOperation, ID: operation.ID},
			Method:      method,
			Path:        path,
			Destructive: DestructiveTargetForOperation(b.Name, operation),
		}
	default:
		if cmd.Intent != "direct_read" || len(cmd.APISurface) != 1 {
			return ResolvedCommandBinding{}, fmt.Errorf("implemented command %q has no resolvable runtime binding", cmd.Path)
		}
		resolved = ResolvedCommandBinding{
			Binding: connectors.CommandBindingIdentity{Kind: connectors.CommandBindingCommand, ID: cmd.Path},
			Method:  cmd.APISurface[0].Method, Path: cmd.APISurface[0].Path,
		}
	}

	if len(cmd.APISurface) == 1 {
		reference := cmd.APISurface[0]
		// The command projection owns the provider-document endpoint identity.
		// Runtime declarations frequently keep a path relative to a configurable
		// base URL, or use a different local placeholder name for the same path
		// slot. Those are intentionally valid runtime bindings and must not be
		// reclassified by a literal comparison that commandrunner never made.
		// A GRAPHQL reference can instead name the fixed document/field; retain
		// the POST transport in that representation and let Binding disambiguate
		// operations sharing it.
		if reference.Method != "GRAPHQL" {
			resolved.Method = strings.ToUpper(strings.TrimSpace(reference.Method))
			resolved.Path = reference.Path
		}
	}
	return resolved, nil
}

// ResolveImplementedCommandPath resolves one bundle command through the same
// projected command surface commandrunner receives.
func ResolveImplementedCommandPath(b Bundle, path string) (ResolvedCommandBinding, error) {
	surface := synthesizeCommandSurface(b)
	if surface == nil {
		return ResolvedCommandBinding{}, fmt.Errorf("bundle %q has no command surface", b.Name)
	}
	matches := 0
	var command connectors.CommandSurfaceCommand
	for _, candidate := range surface.Commands {
		if candidate.Path == path {
			matches++
			command = candidate
		}
	}
	if matches != 1 {
		return ResolvedCommandBinding{}, fmt.Errorf("bundle %q command %q resolves %d times", b.Name, path, matches)
	}
	return ResolveImplementedCommandBinding(b, command)
}

// PreflightImplementedCommand lets commandrunner use the same resolver as
// admission before selecting an executor.
func (c *Connector) PreflightImplementedCommand(cmd connectors.CommandSurfaceCommand) error {
	_, err := ResolveImplementedCommandBinding(c.bundle, cmd)
	return err
}

func (b Base) PreflightImplementedCommand(cmd connectors.CommandSurfaceCommand) error {
	_, err := ResolveImplementedCommandBinding(b.bundle, cmd)
	return err
}

func commandBindingStream(b Bundle, name string) (StreamSpec, bool) {
	for _, stream := range b.Streams {
		if stream.Name == name {
			return stream, true
		}
	}
	return StreamSpec{}, false
}

func commandBindingWrite(b Bundle, name string) (WriteAction, bool) {
	for _, action := range b.Writes {
		if action.Name == name {
			return action, true
		}
	}
	return WriteAction{}, false
}

func commandBindingOperation(b Bundle, id string) (OperationSpec, bool) {
	for _, operation := range b.Operations {
		if operation.ID == id {
			return operation, true
		}
	}
	return OperationSpec{}, false
}

func commandBindingOperationEndpoint(operation OperationSpec) (string, string, error) {
	switch {
	case operation.REST != nil:
		return strings.ToUpper(strings.TrimSpace(operation.REST.Method)), operation.REST.Path, nil
	case operation.Binary != nil:
		return strings.ToUpper(strings.TrimSpace(operation.Binary.Method)), operation.Binary.Path, nil
	case operation.GraphQL != nil:
		return http.MethodPost, operation.GraphQL.Path, nil
	default:
		return "", "", fmt.Errorf("operation %q has no command-addressable endpoint", operation.ID)
	}
}

func normalizeBindingPathTemplate(path string) string {
	return commandBindingTemplateRE.ReplaceAllString(path, `{$1}`)
}
