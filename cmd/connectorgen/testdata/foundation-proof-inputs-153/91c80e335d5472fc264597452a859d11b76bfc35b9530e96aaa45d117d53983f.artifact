package engine

import (
	"fmt"
	"strings"

	"polymetrics.ai/internal/connectors"
)

// PreflightDeferredCommand proves a deferred command has one bounded execution
// target before commandrunner reports its concrete missing foundation. It is
// deliberately no-network and credential-free; historical source admission,
// offline evidence and retention state cannot hide a documented command.
func (c *Connector) PreflightDeferredCommand(cmd connectors.CommandSurfaceCommand) error {
	return PreflightDeferredCommand(c.bundle, cmd)
}

// PreflightDeferredCommand gives native connectors embedding Base the same
// exact, bundle-owned deferred target validation as a declarative Connector.
func (b Base) PreflightDeferredCommand(cmd connectors.CommandSurfaceCommand) error {
	return PreflightDeferredCommand(b.bundle, cmd)
}

// PreflightDeferredCommand validates a deferred command against its execution
// bundle and named absence predicate. It must stay separate from executable
// preflight: success means only that the command is honestly deferred, never
// that its provider operation can run.
func PreflightDeferredCommand(b Bundle, cmd connectors.CommandSurfaceCommand) error {
	foundation := cmd.Foundation
	if foundation == nil || strings.TrimSpace(foundation.ID) == "" || strings.TrimSpace(foundation.Reason) == "" ||
		!connectors.ValidCommandFoundationComponent(foundation.Component) ||
		!connectors.ValidCommandFoundationEvidence(foundation.Component, foundation.Evidence) {
		return fmt.Errorf("deferred command requires a typed foundation component and evidence")
	}
	if strings.TrimSpace(foundation.Target.Method) == "" || strings.TrimSpace(foundation.Target.Path) == "" {
		return fmt.Errorf("deferred command foundation requires one exact target")
	}
	if err := ValidateCommandEndpoint(foundation.Target.Method, foundation.Target.Path); err != nil {
		return fmt.Errorf("deferred command foundation target: %w", err)
	}
	if !validCommandBinding(foundation.Target.Binding.Kind, foundation.Target.Binding.ID) {
		return fmt.Errorf("deferred command foundation target requires one execution binding")
	}
	if foundation.Target.DestructiveKind != "" {
		switch foundation.Target.DestructiveKind {
		case "none", "delete", "destructive":
		default:
			return fmt.Errorf("deferred command foundation target has invalid destructive semantic %q", foundation.Target.DestructiveKind)
		}
	}
	if len(cmd.APISurface) > 1 {
		return fmt.Errorf("deferred command must not reference ambiguous execution endpoints")
	}
	if len(cmd.APISurface) == 1 {
		endpointRef := cmd.APISurface[0]
		if !strings.EqualFold(endpointRef.Method, foundation.Target.Method) || endpointRef.Path != foundation.Target.Path {
			return fmt.Errorf("deferred command foundation target does not match its execution endpoint")
		}
	}
	return deferredCommandEvidenceMatchesBundle(b, cmd)
}

func deferredCommandEvidenceMatchesBundle(b Bundle, cmd connectors.CommandSurfaceCommand) error {
	return deferredCommandIdentityEvidenceMatchesBundle(b, cmd)
}

func validCommandBinding(kind, id string) bool {
	if id == "" || id != strings.TrimSpace(id) {
		return false
	}
	switch kind {
	case connectors.CommandBindingCommand, connectors.CommandBindingStream, connectors.CommandBindingWrite, connectors.CommandBindingOperation:
		return true
	default:
		return false
	}
}

func deferredCommandIdentityEvidenceMatchesBundle(b Bundle, cmd connectors.CommandSurfaceCommand) error {
	foundation := cmd.Foundation
	target := foundation.Target
	actualBinding, bindings, err := deferredCommandCommandBinding(cmd)
	if err != nil {
		return err
	}
	if bindings == 1 && actualBinding != target.Binding {
		return fmt.Errorf("deferred command runtime binding does not match execution binding %s/%s", target.Binding.Kind, target.Binding.ID)
	}
	resolved, err := deferredCommandResolveBinding(b, target)
	if err != nil {
		return err
	}
	if !resolved.exists && deferredCommandEndpointClaimedByAnotherBinding(b, target) {
		return fmt.Errorf("deferred command target is claimed by a different runtime binding")
	}

	switch foundation.Component {
	case connectors.FoundationComponentTypedWriteAction:
		if cmd.Intent != "reverse_etl" && cmd.Intent != "binary_upload" {
			return fmt.Errorf("typed_write_action foundation does not apply to %q", cmd.Intent)
		}
		if target.Binding.Kind != connectors.CommandBindingWrite {
			return fmt.Errorf("typed_write_action foundation requires a write binding identity")
		}
		if bindings != 0 || resolved.exists {
			return fmt.Errorf("typed_write_action foundation is stale: the execution write action already exists")
		}
	case connectors.FoundationComponentTypedRecordSchema:
		if target.Binding.Kind != connectors.CommandBindingWrite || !resolved.exists || resolved.action == nil || bindings != 1 {
			return fmt.Errorf("typed_record_schema foundation requires its exact execution write action")
		}
		if len(resolved.action.RecordSchema) != 0 {
			return fmt.Errorf("typed_record_schema foundation is stale: the target action has a record schema")
		}
	case connectors.FoundationComponentTypedRequestBody:
		if target.Binding.Kind != connectors.CommandBindingOperation || !resolved.exists || resolved.operation == nil || bindings != 1 || resolved.operation.REST == nil {
			return fmt.Errorf("typed_request_body foundation requires its exact execution REST operation")
		}
		if len(resolved.operation.REST.BodySchema) != 0 {
			return fmt.Errorf("typed_request_body foundation is stale: the target operation has a request body schema")
		}
	case connectors.FoundationComponentTypedResponseDescriptor:
		if target.Binding.Kind != connectors.CommandBindingOperation || !resolved.exists || resolved.operation == nil || bindings != 1 ||
			(resolved.operation.REST == nil && resolved.operation.Binary == nil) {
			return fmt.Errorf("typed_response_descriptor foundation requires its exact execution operation")
		}
		if (resolved.operation.REST != nil && resolved.operation.REST.Response != nil) ||
			(resolved.operation.Binary != nil && resolved.operation.Binary.Response != nil) {
			return fmt.Errorf("typed_response_descriptor foundation is stale: the target operation has a response descriptor")
		}
	case connectors.FoundationComponentBinaryTransferBinding:
		switch cmd.Intent {
		case "binary_download":
			if target.Binding.Kind != connectors.CommandBindingOperation {
				return fmt.Errorf("binary download foundation requires an operation binding identity")
			}
			if !resolved.exists {
				if bindings != 0 {
					return fmt.Errorf("binary_transfer_binding foundation references an unknown runtime binding")
				}
				return nil
			}
			if bindings != 1 || resolved.operation == nil {
				return fmt.Errorf("binary_transfer_binding foundation does not reference its exact execution operation")
			}
			if resolved.operation.Binary != nil {
				return fmt.Errorf("binary_transfer_binding foundation is stale: the target has a binary operation binding")
			}
		case "binary_upload":
			if target.Binding.Kind != connectors.CommandBindingWrite {
				return fmt.Errorf("binary upload foundation requires a write binding identity")
			}
			if !resolved.exists {
				if bindings != 0 {
					return fmt.Errorf("binary_transfer_binding foundation references an unknown runtime binding")
				}
				return nil
			}
			if bindings != 1 || resolved.action == nil {
				return fmt.Errorf("binary_transfer_binding foundation does not reference its exact execution write action")
			}
			if resolved.action.BinaryUpload != nil {
				return fmt.Errorf("binary_transfer_binding foundation is stale: the target has a binary upload binding")
			}
		default:
			return fmt.Errorf("binary_transfer_binding foundation does not apply to %q", cmd.Intent)
		}
	case connectors.FoundationComponentRuntimeExecutor:
		if bindings != 0 {
			return fmt.Errorf("runtime_executor foundation must not reference an undeclared runtime binding")
		}
		if resolved.exists {
			return fmt.Errorf("runtime_executor foundation is stale: the execution binding already exists")
		}
	default:
		return fmt.Errorf("unknown deferred foundation component %q", foundation.Component)
	}
	return nil
}

type deferredResolvedBinding struct {
	action    *WriteAction
	operation *OperationSpec
	stream    *StreamSpec
	exists    bool
}

func deferredCommandCommandBinding(cmd connectors.CommandSurfaceCommand) (connectors.CommandBindingIdentity, int, error) {
	values := []connectors.CommandBindingIdentity{
		{Kind: connectors.CommandBindingStream, ID: cmd.Stream},
		{Kind: connectors.CommandBindingWrite, ID: cmd.Write},
		{Kind: connectors.CommandBindingOperation, ID: cmd.Operation},
	}
	var binding connectors.CommandBindingIdentity
	count := 0
	for _, candidate := range values {
		if strings.TrimSpace(candidate.ID) == "" {
			continue
		}
		binding = candidate
		count++
	}
	if count > 1 {
		return connectors.CommandBindingIdentity{}, count, fmt.Errorf("deferred command references more than one runtime binding")
	}
	return binding, count, nil
}

func deferredCommandResolveBinding(b Bundle, target connectors.CommandFoundationTarget) (deferredResolvedBinding, error) {
	switch target.Binding.Kind {
	case connectors.CommandBindingCommand:
		return deferredResolvedBinding{}, nil
	case connectors.CommandBindingStream:
		stream, ok := commandBindingStream(b, target.Binding.ID)
		if !ok {
			return deferredResolvedBinding{}, nil
		}
		method := stream.Method
		if method == "" {
			method = "GET"
		}
		if !strings.EqualFold(method, target.Method) || normalizeBindingPathTemplate(stream.Path) != target.Path {
			return deferredResolvedBinding{}, fmt.Errorf("execution stream binding resolves to a different provider target")
		}
		return deferredResolvedBinding{stream: &stream, exists: true}, nil
	case connectors.CommandBindingWrite:
		action, ok := commandBindingWrite(b, target.Binding.ID)
		if !ok {
			return deferredResolvedBinding{}, nil
		}
		if !strings.EqualFold(action.Method, target.Method) || normalizeBindingPathTemplate(action.Path) != target.Path {
			return deferredResolvedBinding{}, fmt.Errorf("execution write binding resolves to a different provider target")
		}
		return deferredResolvedBinding{action: &action, exists: true}, nil
	case connectors.CommandBindingOperation:
		operation, ok := commandBindingOperation(b, target.Binding.ID)
		if !ok {
			return deferredResolvedBinding{}, nil
		}
		method, path, err := commandBindingOperationEndpoint(operation)
		if err != nil {
			return deferredResolvedBinding{}, err
		}
		if !strings.EqualFold(method, target.Method) || path != target.Path {
			return deferredResolvedBinding{}, fmt.Errorf("execution operation binding resolves to a different provider target")
		}
		return deferredResolvedBinding{operation: &operation, exists: true}, nil
	default:
		return deferredResolvedBinding{}, fmt.Errorf("unknown execution binding kind %q", target.Binding.Kind)
	}
}

func deferredCommandEndpointClaimedByAnotherBinding(b Bundle, target connectors.CommandFoundationTarget) bool {
	for _, stream := range b.Streams {
		method := stream.Method
		if method == "" {
			method = "GET"
		}
		if strings.EqualFold(method, target.Method) && normalizeBindingPathTemplate(stream.Path) == target.Path &&
			(target.Binding.Kind != connectors.CommandBindingStream || target.Binding.ID != stream.Name) {
			return true
		}
	}
	for _, action := range b.Writes {
		if strings.EqualFold(action.Method, target.Method) && normalizeBindingPathTemplate(action.Path) == target.Path &&
			(target.Binding.Kind != connectors.CommandBindingWrite || target.Binding.ID != action.Name) {
			return true
		}
	}
	for _, operation := range b.Operations {
		method, path, err := commandBindingOperationEndpoint(operation)
		if err == nil && strings.EqualFold(method, target.Method) && path == target.Path &&
			(target.Binding.Kind != connectors.CommandBindingOperation || target.Binding.ID != operation.ID) {
			return true
		}
	}
	return false
}
