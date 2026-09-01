package engine

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	defaultOperationStatusMaxBytes = 1024
	defaultOperationStatusTimeout  = 15 * time.Second
)

// pmcert:executes rest_status
//
// OperationStatusCheck executes exactly one declared HEAD operation and
// returns final status metadata only, including a non-2xx response after normal
// retries. This intentionally does not share the JSON direct-read executor:
// status operations cannot decode or print a body.
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
	if _, err := requireOperationSuccessStatusPolicy(op); err != nil {
		return connectors.OperationStatusCheckResult{}, err
	}
	if _, err := operationRedirectPolicy(op); err != nil {
		return connectors.OperationStatusCheckResult{}, err
	}
	cfg := materializeConfigDefaults(b, req.Config)
	effectivePathParams, err := materializeOperationDirectReadPathParams(op, cfg, req.PathParams)
	if err != nil {
		return connectors.OperationStatusCheckResult{}, err
	}
	path, err := resolveSurfaceEndpointPath(op.REST.Path, cfg, effectivePathParams)
	if err != nil {
		return connectors.OperationStatusCheckResult{}, err
	}
	queryMap, err := operationDirectReadQuery(op, req.Query, nil)
	if err != nil {
		return connectors.OperationStatusCheckResult{}, err
	}
	if err := requireOperationQueryGroups(op, queryMap); err != nil {
		return connectors.OperationStatusCheckResult{}, err
	}
	query, err := directReadQuery(queryMap)
	if err != nil {
		return connectors.OperationStatusCheckResult{}, err
	}
	headers, err := operationRequestHeaders(b, op, req.Headers, req.HeaderValues)
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
	requester, err = requesterWithOperationHeaders(requester, op, headers)
	if err != nil {
		return connectors.OperationStatusCheckResult{}, err
	}
	cap := op.REST.MaxBytes
	response, err := requester.DoStatusCheck(requestCtx, normalizeDirectReadPathForBaseURL(path, directReadBaseURL(b, cfg)), query, cap)
	result := connectors.OperationStatusCheckResult{Connector: b.Name, Operation: op.ID, Method: http.MethodHead, Path: path}
	if response != nil {
		result.Receipt = statusCheckReceipt(b, response.Status, response.Header, response.Body)
		result.Status = response.Status
		result.BodyBytes = len(response.Body)
	}
	if result.Receipt == nil && err != nil {
		result.Receipt = statusCheckReceiptFromHTTPError(b, err)
		if result.Receipt != nil {
			result.Status = result.Receipt.Status
			result.BodyBytes = int(result.Receipt.BodyBytes)
		}
	}
	if err != nil {
		return result, fmt.Errorf("operation status %s %s: %w", http.MethodHead, op.REST.Path, err)
	}
	if len(response.Body) > cap {
		return result, fmt.Errorf("operation status response exceeded metadata cap")
	}
	responseHeaders, err := operationResponseHeaders(b, op, response.Header, cfg.Secrets)
	if err != nil {
		return result, err
	}
	result.Headers = responseHeaders
	return result, nil
}

// statusCheckReceipt intentionally records raw bounded HEAD bytes without
// decoding them. A status operation has no response-body authority even when
// an intermediary incorrectly supplies bytes on HEAD.
func statusCheckReceipt(b Bundle, status int, headers http.Header, raw []byte) *connectors.ProviderResponseReceipt {
	receipt := &connectors.ProviderResponseReceipt{
		ResponseReceived: true,
		Status:           status,
		Headers:          completeProviderResponseHeaders(b, headers),
		BodyPresent:      len(raw) != 0,
		BodyBytes:        int64(len(raw)),
	}
	if len(raw) != 0 {
		receipt.BodyRaw, receipt.BodyRawEncoding = writeProviderRawBody(raw)
	}
	return receipt
}

func statusCheckReceiptFromHTTPError(b Bundle, err error) *connectors.ProviderResponseReceipt {
	var httpErr *connsdk.HTTPError
	if !errors.As(err, &httpErr) || httpErr == nil {
		return nil
	}
	raw := append([]byte(nil), httpErr.RawBody...)
	if raw == nil && httpErr.Body != "" {
		raw = []byte(httpErr.Body)
	}
	return statusCheckReceipt(b, httpErr.Status, httpErr.Header, raw)
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
	if op.REST.MaxBytes <= 0 || op.REST.MaxBytes > defaultOperationStatusMaxBytes {
		return OperationSpec{}, fmt.Errorf("operation %q status response cap must be between 1 and %d bytes", operation, defaultOperationStatusMaxBytes)
	}
	if len(op.REST.Body) != 0 || len(op.REST.BodySchema) != 0 || strings.TrimSpace(op.REST.ContentType) != "" {
		return OperationSpec{}, fmt.Errorf("operation %q status check must not declare a request body", operation)
	}
	if _, err := requireOperationSuccessStatusPolicy(op); err != nil {
		return OperationSpec{}, err
	}
	if _, err := operationRedirectPolicy(op); err != nil {
		return OperationSpec{}, err
	}
	return op, nil
}
