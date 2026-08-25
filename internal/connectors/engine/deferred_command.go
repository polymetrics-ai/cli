package engine

import (
	"fmt"
	"regexp"
	"strings"

	"polymetrics.ai/internal/connectors"
)

var deferredCommandActionTemplateRE = regexp.MustCompile(`\{\{\s*(?:config|record)\.([-A-Za-z0-9_]+)\s*\}\}`)

// PreflightDeferredCommand proves a deferred command is still an exact,
// provider-declared target before commandrunner reports missing_foundation. It
// is deliberately no-network and credential-free. A policy/excluded endpoint
// is not a target binding, so it cannot use deferred state to bypass the
// normal API-surface contract.
func (c *Connector) PreflightDeferredCommand(cmd connectors.CommandSurfaceCommand) error {
	return PreflightDeferredCommand(c.bundle, cmd)
}

// PreflightDeferredCommand gives native connectors embedding Base the same
// exact, bundle-owned deferred target validation as a declarative Connector.
func (b Base) PreflightDeferredCommand(cmd connectors.CommandSurfaceCommand) error {
	return PreflightDeferredCommand(b.bundle, cmd)
}

// PreflightDeferredCommand validates a deferred command against its bundle's
// blocked operation ledger and named absence predicate. It must stay separate
// from executable preflight: success means only that the command is honestly
// deferred, never that its provider operation can run.
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
	admittedTarget, admitted, err := deferredCommandAdmittedTarget(b, foundation.Target)
	if err != nil {
		return err
	}
	if len(cmd.APISurface) != 1 {
		return fmt.Errorf("deferred command must reference exactly one API-surface endpoint")
	}
	endpointRef := cmd.APISurface[0]
	if !strings.EqualFold(endpointRef.Method, foundation.Target.Method) || endpointRef.Path != foundation.Target.Path {
		return fmt.Errorf("deferred command foundation target does not match its API-surface endpoint")
	}
	if b.Surface != nil {
		endpoint, err := deferredCommandSurfaceEndpoint(b.Surface, foundation.Target)
		if err != nil {
			return err
		}
		if endpoint.Excluded != nil {
			return fmt.Errorf("deferred command target %s %s is excluded, not a blocked operation", strings.ToUpper(foundation.Target.Method), foundation.Target.Path)
		}
		if endpoint.Operation == nil || endpoint.Operation.Status != "blocked" || !endpoint.Operation.BlockedByDefault {
			return fmt.Errorf("deferred command target %s %s must be a blocked api_surface operation", strings.ToUpper(foundation.Target.Method), foundation.Target.Path)
		}
		if !deferredCommandIntentMatchesOperation(cmd.Intent, endpoint.Operation.Model) {
			return fmt.Errorf("deferred command intent %q does not match blocked target model %q", cmd.Intent, endpoint.Operation.Model)
		}
	} else if !admitted {
		return fmt.Errorf("deferred command target %s %s is not in api_surface or the admitted source ledger", strings.ToUpper(foundation.Target.Method), foundation.Target.Path)
	}
	if admitted && !deferredCommandTargetsEqual(admittedTarget, foundation.Target) {
		return fmt.Errorf("deferred command foundation target does not match admitted source identity %q", foundation.Target.SourceID)
	}
	return deferredCommandEvidenceMatchesBundle(b, cmd)
}

func deferredCommandAdmittedTarget(b Bundle, target connectors.CommandFoundationTarget) (connectors.CommandFoundationTarget, bool, error) {
	if b.declarationTargets != nil {
		if strings.TrimSpace(target.SourceID) == "" {
			return connectors.CommandFoundationTarget{}, false, fmt.Errorf("deferred command foundation target requires its admitted source identity")
		}
		admitted, ok := b.declarationTargets.target(target.SourceID)
		if !ok {
			return connectors.CommandFoundationTarget{}, false, fmt.Errorf("deferred command foundation target references unknown admitted source identity %q", target.SourceID)
		}
		return admitted, true, nil
	}
	if deferredCommandTargetHasIdentity(target) {
		if strings.TrimSpace(target.SourceID) == "" || !validCommandBinding(target.Binding.Kind, target.Binding.ID) {
			return connectors.CommandFoundationTarget{}, false, fmt.Errorf("deferred command foundation target has an incomplete source identity or binding")
		}
		switch target.DestructiveKind {
		case "none", "delete", "destructive":
		default:
			return connectors.CommandFoundationTarget{}, false, fmt.Errorf("deferred command foundation target has invalid destructive semantic %q", target.DestructiveKind)
		}
	}
	return connectors.CommandFoundationTarget{}, false, nil
}

func deferredCommandTargetHasIdentity(target connectors.CommandFoundationTarget) bool {
	return target.SourceID != "" || target.ProviderOperationID != "" || target.Binding.Kind != "" || target.Binding.ID != "" || target.DestructiveKind != ""
}

func deferredCommandTargetsEqual(left, right connectors.CommandFoundationTarget) bool {
	return left.SourceID == right.SourceID && left.ProviderOperationID == right.ProviderOperationID &&
		left.Binding == right.Binding && left.DestructiveKind == right.DestructiveKind &&
		strings.EqualFold(left.Method, right.Method) && left.Path == right.Path
}

func deferredCommandSurfaceEndpoint(surface *APISurface, target connectors.CommandFoundationTarget) (SurfaceEndpoint, error) {
	if surface == nil {
		return SurfaceEndpoint{}, fmt.Errorf("deferred command target %s %s is not in api_surface", strings.ToUpper(target.Method), target.Path)
	}
	var match SurfaceEndpoint
	matches := 0
	for _, endpoint := range surface.Endpoints {
		if strings.EqualFold(endpoint.Method, target.Method) && endpoint.Path == target.Path {
			match = endpoint
			matches++
		}
	}
	if matches == 0 {
		return SurfaceEndpoint{}, fmt.Errorf("deferred command target %s %s is not in api_surface", strings.ToUpper(target.Method), target.Path)
	}
	if matches != 1 {
		return SurfaceEndpoint{}, fmt.Errorf("deferred command target %s %s is duplicated in api_surface", strings.ToUpper(target.Method), target.Path)
	}
	return match, nil
}

func deferredCommandIntentMatchesOperation(intent, model string) bool {
	switch intent {
	case "etl", "direct_read":
		return model == "direct_read"
	case "binary_download":
		return model == "binary_read"
	case "reverse_etl", "direct_write", "binary_upload":
		switch model {
		case "sensitive_reverse_etl", "admin_reverse_etl", "destructive_action":
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func deferredCommandEvidenceMatchesBundle(b Bundle, cmd connectors.CommandSurfaceCommand) error {
	if cmd.Foundation == nil || !validCommandBinding(cmd.Foundation.Target.Binding.Kind, cmd.Foundation.Target.Binding.ID) {
		return deferredCommandEvidenceMatchesBundleLegacy(b, cmd)
	}
	return deferredCommandIdentityEvidenceMatchesBundle(b, cmd)
}

func deferredCommandIdentityEvidenceMatchesBundle(b Bundle, cmd connectors.CommandSurfaceCommand) error {
	foundation := cmd.Foundation
	target := foundation.Target
	actualBinding, bindings, err := deferredCommandCommandBinding(cmd)
	if err != nil {
		return err
	}
	if bindings == 1 && actualBinding != target.Binding {
		return fmt.Errorf("deferred command runtime binding does not match admitted binding %s/%s", target.Binding.Kind, target.Binding.ID)
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
			return fmt.Errorf("typed_write_action foundation is stale: the admitted write action already exists")
		}
	case connectors.FoundationComponentTypedRecordSchema:
		if target.Binding.Kind != connectors.CommandBindingWrite || !resolved.exists || resolved.action == nil || bindings != 1 {
			return fmt.Errorf("typed_record_schema foundation requires its exact admitted write action")
		}
		if len(resolved.action.RecordSchema) != 0 {
			return fmt.Errorf("typed_record_schema foundation is stale: the target action has a record schema")
		}
	case connectors.FoundationComponentTypedRequestBody:
		if target.Binding.Kind != connectors.CommandBindingOperation || !resolved.exists || resolved.operation == nil || bindings != 1 || resolved.operation.REST == nil {
			return fmt.Errorf("typed_request_body foundation requires its exact admitted REST operation")
		}
		if len(resolved.operation.REST.BodySchema) != 0 {
			return fmt.Errorf("typed_request_body foundation is stale: the target operation has a request body schema")
		}
	case connectors.FoundationComponentTypedResponseDescriptor:
		if target.Binding.Kind != connectors.CommandBindingOperation || !resolved.exists || resolved.operation == nil || bindings != 1 {
			return fmt.Errorf("typed_response_descriptor foundation requires its exact admitted operation")
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
				return fmt.Errorf("binary_transfer_binding foundation does not reference its exact admitted operation")
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
				return fmt.Errorf("binary_transfer_binding foundation does not reference its exact admitted write action")
			}
			if resolved.action.BinaryUpload != nil {
				return fmt.Errorf("binary_transfer_binding foundation is stale: the target has a binary upload binding")
			}
		default:
			return fmt.Errorf("binary_transfer_binding foundation does not apply to %q", cmd.Intent)
		}
	case connectors.FoundationComponentSourceImporter, connectors.FoundationComponentRuntimeExecutor:
		if bindings != 0 {
			return fmt.Errorf("%s foundation must not reference an undeclared runtime binding", foundation.Component)
		}
		if resolved.exists {
			return fmt.Errorf("%s foundation is stale: the admitted runtime binding already exists", foundation.Component)
		}
	case connectors.FoundationComponentIdempotencyContract:
		if target.Binding.Kind != connectors.CommandBindingWrite || !resolved.exists || resolved.action == nil || bindings != 1 {
			return fmt.Errorf("idempotency_contract foundation requires its exact admitted write action")
		}
		if strings.TrimSpace(resolved.action.IdempotencyKeyHeader) != "" {
			return fmt.Errorf("idempotency_contract foundation is stale: the target action declares an idempotency header")
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
		if !strings.EqualFold(method, target.Method) || canonicalCommandBindingPath(stream.Path) != target.Path {
			return deferredResolvedBinding{}, fmt.Errorf("admitted stream binding resolves to a different provider target")
		}
		return deferredResolvedBinding{stream: &stream, exists: true}, nil
	case connectors.CommandBindingWrite:
		action, ok := commandBindingWrite(b, target.Binding.ID)
		if !ok {
			return deferredResolvedBinding{}, nil
		}
		if !strings.EqualFold(action.Method, target.Method) || canonicalCommandBindingPath(action.Path) != target.Path {
			return deferredResolvedBinding{}, fmt.Errorf("admitted write binding resolves to a different provider target")
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
			return deferredResolvedBinding{}, fmt.Errorf("admitted operation binding resolves to a different provider target")
		}
		return deferredResolvedBinding{operation: &operation, exists: true}, nil
	default:
		return deferredResolvedBinding{}, fmt.Errorf("unknown admitted binding kind %q", target.Binding.Kind)
	}
}

func deferredCommandEndpointClaimedByAnotherBinding(b Bundle, target connectors.CommandFoundationTarget) bool {
	for _, stream := range b.Streams {
		method := stream.Method
		if method == "" {
			method = "GET"
		}
		if strings.EqualFold(method, target.Method) && canonicalCommandBindingPath(stream.Path) == target.Path &&
			(target.Binding.Kind != connectors.CommandBindingStream || target.Binding.ID != stream.Name) {
			return true
		}
	}
	for _, action := range b.Writes {
		if strings.EqualFold(action.Method, target.Method) && canonicalCommandBindingPath(action.Path) == target.Path &&
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

func deferredCommandEvidenceMatchesBundleLegacy(b Bundle, cmd connectors.CommandSurfaceCommand) error {
	foundation := cmd.Foundation
	target := foundation.Target
	action, hasAction := deferredCommandWriteAction(b, target)
	operation, hasOperation := deferredCommandOperation(b, target)
	stream := deferredCommandStream(b, target)
	bindings := 0
	for _, binding := range []string{cmd.Stream, cmd.Write, cmd.Operation} {
		if strings.TrimSpace(binding) != "" {
			bindings++
		}
	}
	if bindings > 1 {
		return fmt.Errorf("deferred command references more than one runtime binding")
	}

	switch foundation.Component {
	case connectors.FoundationComponentTypedWriteAction:
		if cmd.Intent != "reverse_etl" && cmd.Intent != "binary_upload" {
			return fmt.Errorf("typed_write_action foundation does not apply to %q", cmd.Intent)
		}
		if bindings != 0 || hasAction {
			return fmt.Errorf("typed_write_action foundation is stale: a declared write action already maps to the target")
		}
	case connectors.FoundationComponentTypedRecordSchema:
		if !hasAction || strings.TrimSpace(cmd.Write) == "" || action.Name != cmd.Write {
			return fmt.Errorf("typed_record_schema foundation requires its exact declared write action")
		}
		if len(action.RecordSchema) != 0 {
			return fmt.Errorf("typed_record_schema foundation is stale: the target action has a record schema")
		}
	case connectors.FoundationComponentTypedRequestBody:
		if !hasOperation || strings.TrimSpace(cmd.Operation) == "" || operation.ID != cmd.Operation || operation.REST == nil {
			return fmt.Errorf("typed_request_body foundation requires its exact declared REST operation")
		}
		if len(operation.REST.BodySchema) != 0 {
			return fmt.Errorf("typed_request_body foundation is stale: the target operation has a request body schema")
		}
	case connectors.FoundationComponentTypedResponseDescriptor:
		if !hasOperation || strings.TrimSpace(cmd.Operation) == "" || operation.ID != cmd.Operation {
			return fmt.Errorf("typed_response_descriptor foundation requires its exact declared operation")
		}
		if (operation.REST != nil && operation.REST.Response != nil) || (operation.Binary != nil && operation.Binary.Response != nil) {
			return fmt.Errorf("typed_response_descriptor foundation is stale: the target operation has a response descriptor")
		}
	case connectors.FoundationComponentBinaryTransferBinding:
		switch cmd.Intent {
		case "binary_download":
			if !hasOperation {
				if bindings != 0 {
					return fmt.Errorf("binary_transfer_binding foundation references an unknown runtime binding")
				}
				return nil
			}
			if cmd.Operation != operation.ID || operation.Binary == nil {
				if cmd.Operation != operation.ID {
					return fmt.Errorf("binary_transfer_binding foundation does not reference its exact declared operation")
				}
				return nil
			}
			return fmt.Errorf("binary_transfer_binding foundation is stale: the target has a binary operation binding")
		case "binary_upload":
			if !hasAction {
				if bindings != 0 {
					return fmt.Errorf("binary_transfer_binding foundation references an unknown runtime binding")
				}
				return nil
			}
			if cmd.Write != action.Name || action.BinaryUpload == nil {
				if cmd.Write != action.Name {
					return fmt.Errorf("binary_transfer_binding foundation does not reference its exact declared write action")
				}
				return nil
			}
			return fmt.Errorf("binary_transfer_binding foundation is stale: the target has a binary upload binding")
		default:
			return fmt.Errorf("binary_transfer_binding foundation does not apply to %q", cmd.Intent)
		}
	case connectors.FoundationComponentSourceImporter, connectors.FoundationComponentRuntimeExecutor:
		if bindings != 0 {
			return fmt.Errorf("%s foundation must not reference an undeclared runtime binding", foundation.Component)
		}
		if hasAction || hasOperation || stream {
			return fmt.Errorf("%s foundation is stale: the target already has a declared runtime binding", foundation.Component)
		}
	case connectors.FoundationComponentIdempotencyContract:
		if !hasAction || strings.TrimSpace(cmd.Write) == "" || action.Name != cmd.Write {
			return fmt.Errorf("idempotency_contract foundation requires its exact declared write action")
		}
		if strings.TrimSpace(action.IdempotencyKeyHeader) != "" {
			return fmt.Errorf("idempotency_contract foundation is stale: the target action declares an idempotency header")
		}
	default:
		return fmt.Errorf("unknown deferred foundation component %q", foundation.Component)
	}
	return nil
}

func deferredCommandWriteAction(b Bundle, target connectors.CommandFoundationTarget) (WriteAction, bool) {
	for _, action := range b.Writes {
		if strings.EqualFold(action.Method, target.Method) && deferredCommandActionPath(action.Path) == target.Path {
			return action, true
		}
	}
	return WriteAction{}, false
}

func deferredCommandOperation(b Bundle, target connectors.CommandFoundationTarget) (OperationSpec, bool) {
	for _, operation := range b.Operations {
		if operation.REST != nil && strings.EqualFold(operation.REST.Method, target.Method) && operation.REST.Path == target.Path {
			return operation, true
		}
		if operation.Binary != nil && strings.EqualFold(operation.Binary.Method, target.Method) && operation.Binary.Path == target.Path {
			return operation, true
		}
		if operation.GraphQL != nil && strings.EqualFold("POST", target.Method) && operation.GraphQL.Path == target.Path {
			return operation, true
		}
	}
	return OperationSpec{}, false
}

func deferredCommandStream(b Bundle, target connectors.CommandFoundationTarget) bool {
	for _, stream := range b.Streams {
		method := stream.Method
		if method == "" {
			method = "GET"
		}
		if strings.EqualFold(method, target.Method) && stream.Path == target.Path {
			return true
		}
	}
	return false
}

func deferredCommandActionPath(path string) string {
	return deferredCommandActionTemplateRE.ReplaceAllString(path, "{$1}")
}
