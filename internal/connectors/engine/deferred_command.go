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
	if len(cmd.APISurface) != 1 {
		return fmt.Errorf("deferred command must reference exactly one API-surface endpoint")
	}
	endpointRef := cmd.APISurface[0]
	if !strings.EqualFold(endpointRef.Method, foundation.Target.Method) || endpointRef.Path != foundation.Target.Path {
		return fmt.Errorf("deferred command foundation target does not match its API-surface endpoint")
	}
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
	return deferredCommandEvidenceMatchesBundle(b, cmd)
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
