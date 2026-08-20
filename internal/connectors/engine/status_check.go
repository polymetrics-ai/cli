package engine

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
)

const (
	defaultOperationStatusMaxBytes = 1024
	defaultOperationStatusTimeout  = 15 * time.Second
)

// OperationStatusCheck executes exactly one declared HEAD operation and
// returns status metadata only. This intentionally does not share the JSON
// direct-read executor: status operations cannot decode or print a body.
//
// pmcert:executes rest_status
func OperationStatusCheck(ctx context.Context, b Bundle, req connectors.OperationStatusCheckRequest, h Hooks) (connectors.OperationStatusCheckResult, error) {
	if err := ctx.Err(); err != nil {
		return connectors.OperationStatusCheckResult{}, err
	}
	op, err := operationStatusCheckSpec(b, req.Operation)
	if err != nil {
		return connectors.OperationStatusCheckResult{}, err
	}
	cfg := materializeConfigDefaults(b, req.Config)
	path, err := resolveSurfaceEndpointPath(op.REST.Path, cfg, req.PathParams)
	if err != nil {
		return connectors.OperationStatusCheckResult{}, err
	}
	query, err := directReadQuery(req.Query)
	if err != nil {
		return connectors.OperationStatusCheckResult{}, err
	}
	requestCtx, cancel := context.WithTimeout(ctx, defaultOperationStatusTimeout)
	defer cancel()
	rt, err := newRuntime(requestCtx, b, cfg, h)
	if err != nil {
		return connectors.OperationStatusCheckResult{}, err
	}
	requester, err := rt.requesterFor(http.MethodHead, op.REST.Path)
	if err != nil {
		return connectors.OperationStatusCheckResult{}, err
	}
	cap := op.REST.MaxBytes
	if cap == 0 {
		cap = defaultOperationStatusMaxBytes
	}
	response, err := requester.DoStatusCheck(requestCtx, normalizeDirectReadPathForBaseURL(path, directReadBaseURL(b, cfg)), query, cap)
	if err != nil {
		return connectors.OperationStatusCheckResult{}, fmt.Errorf("operation status %s %s: %w", http.MethodHead, op.REST.Path, err)
	}
	if len(response.Body) > cap {
		return connectors.OperationStatusCheckResult{}, fmt.Errorf("operation status response exceeded metadata cap")
	}
	return connectors.OperationStatusCheckResult{Connector: b.Name, Operation: op.ID, Method: http.MethodHead, Path: path, Status: response.Status, BodyBytes: len(response.Body)}, nil
}

func PreflightOperationStatusCheck(b Bundle, operation, method, path, outputPolicy string) error {
	op, err := operationStatusCheckSpec(b, operation)
	if err != nil {
		return err
	}
	if !strings.EqualFold(method, http.MethodHead) || !strings.EqualFold(op.REST.Method, http.MethodHead) {
		return fmt.Errorf("operation status check requires HEAD")
	}
	if path != op.REST.Path || outputPolicy != "status" {
		return fmt.Errorf("operation status check command does not match declared status contract")
	}
	return nil
}

func operationStatusCheckSpec(b Bundle, operation string) (OperationSpec, error) {
	op, err := findOperation(b, operation)
	if err != nil {
		return OperationSpec{}, err
	}
	if op.Kind != "rest_status" || op.REST == nil || strings.ToUpper(strings.TrimSpace(op.REST.Method)) != http.MethodHead || op.OutputPolicy != "status" {
		return OperationSpec{}, fmt.Errorf("operation %q is not a declared HEAD status operation", operation)
	}
	if op.REST.MaxBytes < 0 || op.REST.MaxBytes > defaultOperationStatusMaxBytes {
		return OperationSpec{}, fmt.Errorf("operation %q status response cap must be between 0 and %d bytes", operation, defaultOperationStatusMaxBytes)
	}
	if len(op.REST.Body) != 0 || len(op.REST.BodySchema) != 0 || strings.TrimSpace(op.REST.ContentType) != "" {
		return OperationSpec{}, fmt.Errorf("operation %q status check must not declare a request body", operation)
	}
	if err := requireOperationSurfaceEndpoint(b, http.MethodHead, op.REST.Path); err != nil {
		return OperationSpec{}, err
	}
	return op, nil
}
