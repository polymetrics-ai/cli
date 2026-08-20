package engine

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"mime"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/transportpolicy"
	"polymetrics.ai/internal/safety"
)

const (
	DefaultOperationDirectWriteMaxBytes = 1 << 20
	MaxOperationDirectWriteBytes        = 16 << 20

	defaultOperationDirectWriteMaxBytes = DefaultOperationDirectWriteMaxBytes
	maxOperationDirectWriteBytes        = MaxOperationDirectWriteBytes
	defaultOperationDirectWriteTimeout  = 30 * time.Second

	directWritePolicyNone                     = "none"
	directWritePolicyJSON                     = "json"
	directWritePolicyJSONRedacted             = "json_redacted"
	directWritePolicyWriteResultRedacted      = "write_result_redacted"
	directWritePolicyGongBoundedInputRedacted = "gong_bounded_input_redacted"
	directWritePolicySecretStored             = "secret_stored"
)

type preparedOperationDirectWrite struct {
	op                   OperationSpec
	cfg                  connectors.RuntimeConfig
	baseURL              string
	headers              map[string]string
	operationHeaders     http.Header
	runtimeAuth          []AuthSpec
	rateLimitAuth        []AuthSpec
	identity             string
	method               string
	path                 string
	requestPath          string
	query                url.Values
	body                 map[string]any
	form                 url.Values
	format               string
	contentType          string
	policy               string
	maxBytes             int
	redactionValues      []string
	sensitiveHTTPBinding bool
	prepared             PreparedWrite
}

// sealRuntimeConfig copies every mutable request input a prepared write can
// consult after approval. Runtime interfaces remain the caller's installed
// services, but config/secrets and approved file digests are values bound into
// the preview rather than live aliases.
func sealRuntimeConfig(cfg connectors.RuntimeConfig) connectors.RuntimeConfig {
	cloneMap := func(values map[string]string) map[string]string {
		if len(values) == 0 {
			return nil
		}
		out := make(map[string]string, len(values))
		for key, value := range values {
			out[key] = value
		}
		return out
	}
	cfg.Config = cloneMap(cfg.Config)
	cfg.Secrets = cloneMap(cfg.Secrets)
	cfg.ApprovedPayloadSHA256 = cloneMap(cfg.ApprovedPayloadSHA256)
	return cfg
}

type operationDirectWriteError struct {
	operation string
	message   string
	cause     error
}

func (e *operationDirectWriteError) Error() string {
	return fmt.Sprintf("operation direct write %q: %s", e.operation, e.message)
}

func (e *operationDirectWriteError) Unwrap() error { return e.cause }

// PreviewOperationDirectWrite prepares a declared REST or fixed-document
// GraphQL mutation without constructing a runtime or issuing any network
// request. Its digest binds the exact typed request that OperationDirectWrite
// may later dispatch.
func PreviewOperationDirectWrite(ctx context.Context, b Bundle, req connectors.OperationDirectWriteRequest, h Hooks) (connectors.WritePreview, error) {
	prepared, err := prepareOperationDirectWrite(ctx, b, req, h)
	if err != nil {
		return connectors.WritePreview{}, err
	}
	return PreviewPreparedWrite(prepared.prepared)
}

// OperationDirectWrite dispatches exactly one typed, declared REST or
// fixed-document GraphQL mutation after the preview-bound shared write gate
// authorizes it. It never retries: operation declarations carry no idempotency
// proof in this executor, so both transient retries and the requester's
// auth-refresh retry are off.
//
// pmcert:executes rest_write,graphql_mutation
func OperationDirectWrite(ctx context.Context, b Bundle, req connectors.OperationDirectWriteRequest, h Hooks) (connectors.OperationDirectWriteResult, error) {
	prepared, err := prepareOperationDirectWrite(ctx, b, req, h)
	if err != nil {
		return connectors.OperationDirectWriteResult{}, err
	}

	var result connectors.OperationDirectWriteResult
	err = ExecutePreparedWrite(ctx, prepared.prepared, req.Approval, req.PreviewDigest, func(gated context.Context) error {
		// A redirect can replay a POST/PUT/PATCH/DELETE below Requester's retry
		// loop. Reuse the shared prepared-write transport policy to refuse it:
		// every rest_write is exactly the target that preview bound, regardless
		// of whether the mutation also needs destructive confirmation evidence.
		gated = transportpolicy.MarkDestructive(gated)
		requestCtx, cancel := context.WithTimeout(gated, defaultOperationDirectWriteTimeout)
		defer cancel()

		runtimeBundle := b
		runtimeBundle.HTTP.UserAgent = ""
		runtimeBundle.HTTP.Auth = append([]AuthSpec(nil), prepared.runtimeAuth...)
		rt, err := newRuntimeWithResolvedHTTPBindings(requestCtx, runtimeBundle, prepared.cfg, h, prepared.baseURL, prepared.headers, "", prepared.runtimeAuth, prepared.rateLimitAuth)
		if err != nil {
			return err
		}
		resolvedRequester, err := rt.requesterFor(prepared.method, prepared.path)
		if err != nil {
			return err
		}
		resolvedRequester, err = requesterWithOperationHeaders(resolvedRequester, prepared.op, prepared.operationHeaders)
		if err != nil {
			return err
		}
		requester := *resolvedRequester
		requester.DisableRetries = true
		if len(prepared.prepared.Requests) != 1 {
			return fmt.Errorf("operation %q prepared request count %d, want exactly one", prepared.op.ID, len(prepared.prepared.Requests))
		}
		sealedRequest := prepared.prepared.Requests[0]
		sealedQuery, queryErr := url.ParseQuery(sealedRequest.Query)
		if queryErr != nil || sealedQuery.Encode() != sealedRequest.Query {
			return fmt.Errorf("operation %q prepared query is not canonical", prepared.op.ID)
		}

		// Crossing this point means the sealed request is about to be submitted.
		// Record its declaration-owned identity before transport I/O so timeout,
		// cancellation, DNS, TLS, and connection failures retain an attempt
		// receipt without fabricating response fields.
		result = connectors.OperationDirectWriteResult{
			Connector:              b.Name,
			Operation:              prepared.op.ID,
			Method:                 prepared.method,
			Path:                   prepared.path,
			OutputSecretFields:     operationDirectWriteOutputSecretFields(prepared.op),
			RequestSensitiveValues: append([]string(nil), prepared.redactionValues...),
		}

		var response *connsdk.Response
		switch prepared.format {
		case "form":
			response, err = requester.DoPreparedFormLimited(requestCtx, sealedRequest.Method, prepared.requestPath, sealedQuery, []byte(sealedRequest.Body), prepared.maxBytes)
		case "json", "none", "graphql":
			contentType := sealedRequest.ContentType
			if contentType == "" {
				contentType = "application/json"
			}
			response, err = requester.DoPreparedJSONLimited(requestCtx, sealedRequest.Method, prepared.requestPath, sealedQuery, []byte(sealedRequest.Body), contentType, prepared.maxBytes)
		case "multipart":
			root, rootErr := openMultipartRoot(prepared.cfg.ProjectDir)
			if rootErr != nil {
				return fmt.Errorf("operation %q open multipart project root: %w", prepared.op.ID, rootErr)
			}
			defer func() { _ = root.Close() }()
			// buildMultipartPayload is the established writes.json transport
			// builder. A tiny synthetic action carries only the already-validated
			// shared part declaration, keeping root confinement, regular-file,
			// aggregate-cap, and approved-digest behavior exactly in one place.
			action := WriteAction{Name: prepared.op.ID, Multipart: prepared.op.REST.Multipart}
			multipart, buildErr := buildMultipartPayload(action, connectors.Record(prepared.body), 0, prepared.cfg, root)
			if buildErr != nil {
				return buildErr
			}
			response, err = requester.DoMultipartLimited(requestCtx, prepared.method, prepared.requestPath, prepared.query, multipart, prepared.maxBytes)
		default:
			return fmt.Errorf("operation %q has unsupported prepared body format %q", prepared.op.ID, prepared.format)
		}
		var responseBody operationDirectWriteResponseBodyResult
		var responseBodyErr error
		if response != nil {
			responseBody, responseBodyErr = operationDirectWriteResponseBody(prepared.policy, response.Body, prepared.maxBytes, response.Header)
			result = operationDirectWriteResultFromResponse(b.Name, prepared, response, responseBody)
		}
		if err != nil {
			if operationDirectWriteHTTPErrorBodyExceedsLimit(err, prepared.maxBytes) {
				return &operationDirectWriteError{
					operation: prepared.op.ID,
					message:   fmt.Sprintf("provider response body exceeds declared limit %d", prepared.maxBytes),
					cause:     operationDirectWriteOversizeHTTPErrorCause(err, prepared.identity),
				}
			}
			class, hint := applyErrorMap(b.HTTP.ErrorMap, err)
			message := "provider returned an HTTP error"
			if prepared.op.Kind == "graphql_mutation" {
				message = "provider returned a GraphQL HTTP error"
			}
			if !prepared.sensitiveHTTPBinding {
				message = operationDirectWriteErrorText(
					err,
					prepared.identity,
					operationDirectWritePrintsProviderHTTPBody(prepared),
					operationDirectWriteDiagnosticValues(prepared.redactionValues, prepared.cfg.Secrets, prepared.cfg.Config),
				)
			}
			if hint != "" {
				message += ": " + hint
			}
			if class != "" {
				message = class + ": " + message
			}
			return &operationDirectWriteError{operation: prepared.op.ID, message: message, cause: operationDirectWriteErrorCause(err, prepared.identity)}
		}
		if response == nil {
			return &operationDirectWriteError{operation: prepared.op.ID, message: "provider returned no response"}
		}
		if responseBodyErr != nil {
			return operationDirectWritePostResponseError(prepared.op.ID, responseBodyErr, response, prepared.identity)
		}
		responseHeaders, headerErr := operationResponseHeaders(b, prepared.op, response.Header, prepared.cfg.Secrets)
		if headerErr != nil {
			return headerErr
		}
		if len(responseHeaders) > 0 && result.Headers == nil {
			result.Headers = make(map[string]connectors.OperationResponseHeader, len(responseHeaders))
		}
		for name, header := range responseHeaders {
			result.Headers[name] = header
		}
		// A fixed GraphQL document is a JSON protocol when a provider omits
		// Content-Type, but an explicitly non-JSON provider response remains a
		// byte-exact result rather than a fabricated GraphQL envelope.
		contentTypePresent, responseJSON := writeProviderResponseContentType(response.Header)
		if prepared.op.Kind == "graphql_mutation" && (!contentTypePresent || responseJSON) {
			envelope, envelopeErr := decodeDirectReadBody(response.Body, prepared.maxBytes)
			if envelopeErr != nil {
				return &operationDirectWriteError{operation: prepared.op.ID, message: "GraphQL response is invalid", cause: errors.Join(envelopeErr, operationDirectWriteProviderResponseCause(response, prepared.identity))}
			}
			data, metadata, parseErr := graphQLOperationResponseWithRuntimeErrorPolicy(response.Body, prepared.maxBytes, operationRetainsSecretRuntimeContent(prepared.op))
			if parseErr != nil {
				return &operationDirectWriteError{operation: prepared.op.ID, message: "GraphQL response is invalid", cause: errors.Join(parseErr, operationDirectWriteProviderResponseCause(response, prepared.identity))}
			}
			result.Body = envelope
			result.GraphQL = metadata
			observeGraphQLRateLimit(requestCtx, &requester, response, data)
			if len(metadata.Errors) != 0 {
				return &operationDirectWriteError{operation: prepared.op.ID, message: "graphql errors: provider returned application errors", cause: operationDirectWriteProviderResponseCause(response, prepared.identity)}
			}
			if data == nil {
				return &operationDirectWriteError{operation: prepared.op.ID, message: "GraphQL response has no data", cause: operationDirectWriteProviderResponseCause(response, prepared.identity)}
			}
			return nil
		}
		if prepared.policy == directWritePolicySecretStored {
			if err := storeOperationResponseSecret(gated, prepared.op, prepared.cfg, response.Body); err != nil {
				// Persistence of the declaration-owned response secret can fail
				// after the provider has completed the fixed mutation.  Preserve
				// that physical response as a cause (and in result above) so the
				// durable App run can report the exact receipt without fabricating
				// an in-memory terminal envelope.
				return operationDirectWritePostResponseError(prepared.op.ID, err, response, prepared.identity)
			}
		}
		return nil
	})
	if err != nil {
		return result, err
	}
	return result, nil
}

func operationDirectWriteOutputSecretFields(op OperationSpec) []string {
	if op.SensitivePolicy == nil || strings.TrimSpace(op.SensitivePolicy.ResponseSecretField) == "" {
		return nil
	}
	return []string{op.SensitivePolicy.ResponseSecretField}
}

// operationDirectWriteErrorText keeps generated diagnostics secret-safe. The
// corresponding typed result and wrapped provider cause retain the complete
// provider receipt for the durable reverse-ETL contract; no raw provider body
// is copied into a printable diagnostic.
func operationDirectWriteErrorText(err error, identity string, includeProviderBody bool, values []string) string {
	var output string
	var httpErr *connsdk.HTTPError
	if errors.Is(err, transportpolicy.ErrRedirectRefused) {
		output = transportpolicy.ErrRedirectRefused.Error()
	} else if errors.As(err, &httpErr) {
		output = fmt.Sprintf("provider returned HTTP status %d", httpErr.Status)
		if includeProviderBody && strings.TrimSpace(httpErr.Body) != "" {
			output += ": " + httpErr.Body
		}
	} else if operationDirectWriteErrorMayExposeURL(err) {
		output = fmt.Sprintf("request failed for %s", identity)
	} else {
		output = err.Error()
	}
	return redactOperationDirectWriteErrorText(output, false, values)
}

// operationDirectWriteDiagnosticValues extends declaration-owned body
// redaction with runtime values that a provider can echo in a diagnostic.
// The complete provider receipt stays available on the typed result/cause;
// only this printable error surface masks exact known values.
func operationDirectWriteDiagnosticValues(values []string, secrets, config map[string]string) []string {
	result := append([]string(nil), values...)
	for _, value := range secrets {
		if strings.TrimSpace(value) != "" {
			result = append(result, value)
		}
	}
	for _, value := range config {
		if strings.TrimSpace(value) != "" {
			result = append(result, value)
		}
	}
	return result
}

// operationDirectWritePrintsProviderHTTPBody is the intentionally narrow
// diagnostic policy. A plain declared REST JSON result may retain an
// actionable (with exact known values redacted) provider message. Multipart,
// GraphQL, secret, and explicitly redacted output contracts retain their
// complete receipt only through the typed result/cause, never synthetic text.
func operationDirectWritePrintsProviderHTTPBody(prepared preparedOperationDirectWrite) bool {
	return prepared.op.Kind == "rest_write" && prepared.format != "multipart" && prepared.policy == directWritePolicyJSON && !prepared.sensitiveHTTPBinding
}

func operationDirectWriteErrorCause(err error, identity string) error {
	var httpErr *connsdk.HTTPError
	if errors.As(err, &httpErr) {
		return &connsdk.HTTPError{Status: httpErr.Status, URL: identity, Header: httpErr.Header.Clone(), Body: httpErr.Body}
	}
	if operationDirectWriteErrorMayExposeURL(err) {
		return nil
	}
	return err
}

func operationDirectWriteHTTPErrorBodyExceedsLimit(err error, limit int) bool {
	var httpErr *connsdk.HTTPError
	return limit >= 0 && errors.As(err, &httpErr) && len(httpErr.Body) > limit
}

func operationDirectWriteOversizeHTTPErrorCause(err error, identity string) error {
	var httpErr *connsdk.HTTPError
	if !errors.As(err, &httpErr) {
		return nil
	}
	return &connsdk.HTTPError{Status: httpErr.Status, URL: identity, Header: httpErr.Header.Clone()}
}

func operationDirectWriteProviderResponseCause(response *connsdk.Response, identity string) error {
	if response == nil {
		return nil
	}
	return &connsdk.HTTPError{Status: response.Status, URL: identity, Header: response.Header.Clone(), Body: string(response.Body)}
}

func operationDirectWritePostResponseError(operation string, cause error, response *connsdk.Response, identity string) error {
	return &operationDirectWriteError{
		operation: operation,
		message:   cause.Error(),
		cause:     errors.Join(cause, operationDirectWriteProviderResponseCause(response, identity)),
	}
}

// preparedOperationDirectWriteHeaders joins runtime-owned resolved headers and
// exact operation request headers for the private, digest-bound prepared
// request. Caller headers have already been checked against the operation
// declaration; a collision is still refused because neither source may
// silently replace the other.
func preparedOperationDirectWriteHeaders(runtime map[string]string, operation http.Header) (map[string]string, map[string][]string, error) {
	combined := cloneResolvedHeaders(runtime)
	if combined == nil && len(operation) > 0 {
		combined = make(map[string]string, len(operation))
	}
	for name, value := range operationSingleHeaders(operation) {
		for existing := range combined {
			if strings.EqualFold(existing, name) {
				return nil, nil, fmt.Errorf("operation request header %q collides with runtime-owned header %q", name, existing)
			}
		}
		combined[name] = value
	}
	repeated := operationRepeatedHeaders(operation)
	for name := range repeated {
		for existing := range combined {
			if strings.EqualFold(existing, name) {
				return nil, nil, fmt.Errorf("operation request header %q collides with runtime-owned header %q", name, existing)
			}
		}
	}
	return combined, repeated, nil
}

func operationDirectWriteErrorMayExposeURL(err error) bool {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return true
	}
	message := err.Error()
	return strings.Contains(message, "http://") || strings.Contains(message, "https://")
}

func redactOperationDirectWriteErrorText(text string, redact bool, values []string) string {
	if !redact && len(values) == 0 {
		return text
	}
	return safety.RedactErrorText(redactWriteLiterals(text, values))
}

// operationRetainsSecretRuntimeContent reports whether an operation handles secret-bearing runtime content.
func operationRetainsSecretRuntimeContent(op OperationSpec) bool {
	return op.SecretSensitive || strings.EqualFold(op.MutationClass, "secret")
}

// OperationDirectWriteMetadata returns the closed plan-safe summary for one
// declared rest_write operation. It validates the operation's executable
// shape, but deliberately does not resolve config, build auth, or touch the
// network.
func OperationDirectWriteMetadata(b Bundle, operation string) (connectors.OperationDirectWriteMetadata, error) {
	op, _, err := operationDirectWriteSpec(b, operation)
	if err != nil {
		return connectors.OperationDirectWriteMetadata{}, err
	}
	if op.Kind == "rest_write" {
		if _, _, err := operationDirectWriteContentType(op, nil); err != nil {
			return connectors.OperationDirectWriteMetadata{}, err
		}
	}
	if err := validateOperationDirectWriteOutputPolicy(op.OutputPolicy); err != nil {
		return connectors.OperationDirectWriteMetadata{}, err
	}
	target := DestructiveTargetForOperation(b.Name, op)
	confirmation := ""
	if target.RequiresApproval() {
		confirmation = string(connectors.ConfirmationKindDestructive)
	}
	return connectors.OperationDirectWriteMetadata{
		Operation:             op.ID,
		MutationClass:         op.MutationClass,
		Risk:                  op.Risk,
		Approval:              op.Approval,
		ConfirmationChallenge: confirmation,
		OutputPolicy:          op.OutputPolicy,
		Batchable:             op.IsBatchable(),
		StructuredBody:        OperationDirectWriteHasStructuredRESTBody(op),
		PayloadFileFields:     operationDirectWritePayloadFileFields(op),
		PayloadFileMaxBytes:   operationDirectWritePayloadFileMaxBytes(op),
		RedactFields:          operationDirectWriteRedactFields(op),
	}, nil
}

func ApprovedMultipartPayloadSHA256ForOperation(ctx context.Context, b Bundle, req connectors.OperationDirectWriteRequest, _ Hooks) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	op, _, err := operationDirectWriteSpec(b, req.Operation)
	if err != nil {
		return nil, err
	}
	if op.Kind != "rest_write" || op.REST == nil || op.REST.Multipart == nil {
		return nil, nil
	}
	cfg := materializeConfigDefaults(b, sealRuntimeConfig(req.Config))
	body, err := operationWriteBody(op, req.Body)
	if err != nil {
		return nil, err
	}
	root, err := openMultipartRoot(cfg.ProjectDir)
	if err != nil {
		return nil, fmt.Errorf("operation %q open multipart project root: %w", op.ID, err)
	}
	defer func() { _ = root.Close() }()
	action := WriteAction{Name: op.ID, Multipart: op.REST.Multipart}
	form, err := buildMultipartPayload(action, connectors.Record(body), 0, cfg, root)
	if err != nil {
		return nil, err
	}
	partFields := make(map[string]string, len(op.REST.Multipart.Parts))
	for _, part := range op.REST.Multipart.Parts {
		if part.Type == "file" {
			partFields[part.Name] = part.Field
		}
	}
	approved := make(map[string]string, len(form.Files))
	var total int64
	for _, file := range form.Files {
		field, ok := partFields[file.FieldName]
		if !ok {
			return nil, fmt.Errorf("operation %q multipart file part %q is not declared", op.ID, file.FieldName)
		}
		digest, size, err := digestMultipartPayloadForApproval(file)
		if err != nil {
			return nil, fmt.Errorf("operation %q multipart file part %q: %w", op.ID, file.FieldName, err)
		}
		total += size
		if form.MaxBytes > 0 && total > form.MaxBytes {
			return nil, fmt.Errorf("operation %q multipart payload too large: %d bytes exceeds limit %d", op.ID, total, form.MaxBytes)
		}
		approved[connectors.PayloadApprovalKey(0, field)] = digest
	}
	if len(approved) == 0 {
		return nil, nil
	}
	return approved, nil
}

// PreflightOperationDirectWrite proves an implemented command's exact binding
// can reach the closed REST or fixed-document GraphQL write executor. It is
// deliberately no-network and shares operationDirectWriteSpec with execution,
// so an api_surface row cannot point a command at a different endpoint than
// the preview-bound request the runtime will actually dispatch.
func PreflightOperationDirectWrite(b Bundle, operation, method, endpointPath, outputPolicy string, queryFields ...string) error {
	op, declaredMethod, err := operationDirectWriteSpec(b, operation)
	if err != nil {
		return err
	}
	method = strings.ToUpper(strings.TrimSpace(method))
	if !strings.EqualFold(method, declaredMethod) {
		return fmt.Errorf("operation direct write method %s does not match declared operation method %s", method, declaredMethod)
	}
	var declaredPath string
	if op.Kind == "graphql_mutation" {
		declaredPath = op.GraphQL.Path
	} else {
		declaredPath = op.REST.Path
	}
	if endpointPath != declaredPath {
		return fmt.Errorf("operation direct write path %q does not match declared operation path %q", endpointPath, declaredPath)
	}
	if outputPolicy != op.OutputPolicy {
		return fmt.Errorf("operation direct write output_policy %q does not match declared operation output_policy %q", outputPolicy, op.OutputPolicy)
	}
	if err := validateOperationDirectWriteOutputPolicy(outputPolicy); err != nil {
		return err
	}
	authQueryParameters, err := OperationDirectWriteAuthOwnedQueryParameters(op, b.HTTP.Auth)
	if err != nil {
		return err
	}
	return validateOperationDirectWriteQueryFieldsWithAuth(op, queryFields, authQueryParameters)
}

func PreflightOperationDirectWriteBindings(b Bundle, operation string, pathFields, bodyFields []string) error {
	op, _, err := operationDirectWriteSpec(b, operation)
	if err != nil {
		return err
	}
	return ValidateOperationDirectWriteMappings(op, pathFields, bodyFields)
}

func ValidateOperationDirectWriteMappings(op OperationSpec, pathFields, bodyFields []string) error {
	switch op.Kind {
	case "rest_write":
		if op.REST == nil {
			return fmt.Errorf("operation %q has no rest declaration", op.ID)
		}
		if err := validateOperationDirectWritePathFields(op, pathFields); err != nil {
			return err
		}
		return validateOperationDirectWriteBodyFields(op, bodyFields)
	case "graphql_mutation":
		return validateOperationDirectWriteGraphQLVariableFields(op, pathFields, bodyFields)
	default:
		if len(pathFields) == 0 && len(bodyFields) == 0 {
			return nil
		}
		return fmt.Errorf("operation %q does not permit caller path or body fields", op.ID)
	}
}

func validateOperationDirectWriteGraphQLVariableFields(op OperationSpec, pathFields, bodyFields []string) error {
	if len(pathFields) != 0 {
		return fmt.Errorf("operation %q fixed GraphQL mutation does not accept path fields", op.ID)
	}
	if op.GraphQL == nil {
		return fmt.Errorf("operation %q has no GraphQL declaration", op.ID)
	}
	_, root, err := graphQLOperationVariablesSchema(op)
	if err != nil {
		return err
	}
	properties, ok := root["properties"].(map[string]any)
	if !ok {
		return fmt.Errorf("operation %q graphql.variables_schema must declare properties", op.ID)
	}
	seen := make(map[string]struct{}, len(bodyFields))
	for _, field := range bodyFields {
		if !graphQLNamePattern.MatchString(field) {
			return fmt.Errorf("operation %q GraphQL variable field %q must be a top-level GraphQL variable", op.ID, field)
		}
		if _, duplicate := seen[field]; duplicate {
			return fmt.Errorf("operation %q maps more than one command flag to GraphQL variable %q", op.ID, field)
		}
		seen[field] = struct{}{}
		if _, declared := properties[field]; !declared {
			return fmt.Errorf("operation %q GraphQL variable %q is not declared", op.ID, field)
		}
	}
	return nil
}

func operationDirectWritePathParameterNames(op OperationSpec) (map[string]struct{}, error) {
	if op.REST == nil {
		return nil, fmt.Errorf("operation %q has no rest declaration", op.ID)
	}
	path := op.REST.Path
	remaining := surfacePathVarPattern.ReplaceAllString(path, "")
	if strings.ContainsAny(remaining, "{}") {
		return nil, fmt.Errorf("operation %q has malformed path template %q", op.ID, path)
	}
	names := make(map[string]struct{})
	for _, match := range surfacePathVarPattern.FindAllStringSubmatch(path, -1) {
		if len(match) != 2 {
			continue
		}
		name := match[1]
		if err := safety.ValidateIdentifier(name, "operation path parameter"); err != nil {
			return nil, fmt.Errorf("operation %q path parameter: %w", op.ID, err)
		}
		names[name] = struct{}{}
	}
	return names, nil
}

func validateOperationDirectWritePathFields(op OperationSpec, pathFields []string) error {
	declared, err := operationDirectWritePathParameterNames(op)
	if err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(pathFields))
	for _, field := range pathFields {
		if err := safety.ValidateIdentifier(field, "operation path field"); err != nil {
			return fmt.Errorf("operation %q path field: %w", op.ID, err)
		}
		if _, duplicate := seen[field]; duplicate {
			return fmt.Errorf("operation %q maps more than one command flag to path parameter %q", op.ID, field)
		}
		seen[field] = struct{}{}
		if _, ok := declared[field]; !ok {
			return fmt.Errorf("operation %q path parameter %q is not declared by rest.path", op.ID, field)
		}
	}
	return nil
}

func validateOperationDirectWritePathParams(op OperationSpec, pathParams map[string]string) error {
	fields := make([]string, 0, len(pathParams))
	for field := range pathParams {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	if err := validateOperationDirectWritePathFields(op, fields); err != nil {
		return err
	}
	parameters, err := operationParametersForLocation(op, "path")
	if err != nil {
		return err
	}
	for _, field := range fields {
		parameter, declared := parameters[field]
		if !declared {
			parameter = OperationParameter{Name: field, In: "path", Type: "string"}
		}
		if err := validateOperationParameterWireValue(op, parameter, "path", pathParams[field]); err != nil {
			return err
		}
	}
	return nil
}

func operationDirectWriteBodySchemaRoot(op OperationSpec) (map[string]any, error) {
	if op.REST == nil || len(op.REST.BodySchema) == 0 {
		return nil, fmt.Errorf("operation %q does not declare body_schema", op.ID)
	}
	if operationDirectWriteUsesJSONBody(op) && operationHasStructuredRESTBodyField(op) {
		compiled, err := compileStructuredRESTBodySchema(op)
		if err != nil {
			return nil, err
		}
		return compiled.root, nil
	}
	var root map[string]any
	if err := json.Unmarshal(op.REST.BodySchema, &root); err != nil {
		return nil, fmt.Errorf("operation %q body_schema is not an object: %w", op.ID, err)
	}
	if root == nil {
		return nil, fmt.Errorf("operation %q body_schema must be an object", op.ID)
	}
	return root, nil
}

func validateOperationDirectWriteBodyFields(op OperationSpec, bodyFields []string) error {
	root, err := operationDirectWriteBodySchemaRoot(op)
	if err != nil {
		if len(bodyFields) == 0 && op.REST != nil && len(op.REST.BodySchema) == 0 {
			return nil
		}
		return err
	}
	var staticBody map[string]any
	if OperationDirectWriteHasStructuredRESTBody(op) {
		compiled, err := compileStructuredRESTBodySchema(op)
		if err != nil {
			return err
		}
		staticBody, err = canonicalizeStructuredRESTBodyFragment(compiled, op, op.REST.Body, "rest.body")
		if err != nil {
			return err
		}
	}
	seen := make([]operationDirectWriteBodyPath, 0, len(bodyFields))
	for _, field := range bodyFields {
		resolved, err := resolveOperationDirectWriteBodySchemaPath(root, field)
		if err != nil {
			return fmt.Errorf("operation %q body field %q: %w", op.ID, field, err)
		}
		for _, previous := range seen {
			if operationDirectWriteBodyPathsOverlap(previous, resolved) {
				return fmt.Errorf("operation %q maps overlapping body fields %q and %q", op.ID, previous.raw, field)
			}
		}
		if err := validateOperationDirectWriteStaticBodyMapping(staticBody, resolved); err != nil {
			return fmt.Errorf("operation %q body field %q: %w", op.ID, field, err)
		}
		seen = append(seen, resolved)
	}
	return nil
}

func validateOperationDirectWriteStaticBodyMapping(staticBody map[string]any, path operationDirectWriteBodyPath) error {
	if len(staticBody) == 0 {
		return nil
	}
	var current any = staticBody
	for index, step := range path.steps {
		var next any
		var exists bool
		if step.array {
			values, ok := current.([]any)
			if !ok {
				return fmt.Errorf("does not match its fixed rest.body structure")
			}
			if step.index > len(values) {
				return fmt.Errorf("uses sparse array index %d after its fixed rest.body prefix", step.index)
			}
			for prefix := 0; prefix < step.index && prefix < len(values); prefix++ {
				if _, ok := operationDirectWriteStaticBodyScaffold(values[prefix]); !ok {
					return fmt.Errorf("cannot follow fixed scalar rest.body array item %d", prefix)
				}
			}
			if step.index >= len(values) {
				return nil
			}
			next = values[step.index]
			exists = true
		} else {
			object, ok := current.(map[string]any)
			if !ok {
				return fmt.Errorf("overlaps a fixed rest.body value")
			}
			next, exists = object[step.key]
		}
		if !exists {
			return nil
		}
		if index == len(path.steps)-1 {
			object, array := operationDirectWriteBodyNodeKinds(path.node)
			if object != array && (object || array) {
				return nil
			}
			return fmt.Errorf("overlaps a fixed rest.body value")
		}
		current = next
	}
	return nil
}

func operationDirectWriteBodySchemaPath(root map[string]any, path string) (map[string]any, error) {
	resolved, err := resolveOperationDirectWriteBodySchemaPath(root, path)
	if err != nil {
		return nil, err
	}
	return resolved.node, nil
}

type operationDirectWriteBodyPathStep struct {
	key   string
	index int
	array bool
}

type operationDirectWriteBodyPath struct {
	raw   string
	node  map[string]any
	steps []operationDirectWriteBodyPathStep
}

func resolveOperationDirectWriteBodySchemaPath(root map[string]any, path string) (operationDirectWriteBodyPath, error) {
	if path == "" {
		return operationDirectWriteBodyPath{}, fmt.Errorf("body field is required")
	}
	parts := strings.Split(path, ".")
	for _, part := range parts {
		if part == "" {
			return operationDirectWriteBodyPath{}, fmt.Errorf("body field %q has an empty path segment", path)
		}
	}
	var candidates []operationDirectWriteBodyPath
	var firstErr error
	var visit func(map[string]any, int, []operationDirectWriteBodyPathStep)
	visit = func(node map[string]any, position int, steps []operationDirectWriteBodyPathStep) {
		if position == len(parts) {
			candidates = append(candidates, operationDirectWriteBodyPath{raw: path, node: node, steps: append([]operationDirectWriteBodyPathStep(nil), steps...)})
			return
		}
		object, array := operationDirectWriteBodyNodeKinds(node)
		if object && array {
			if firstErr == nil {
				firstErr = fmt.Errorf("declared schema is ambiguous at %q", parts[position])
			}
			return
		}
		if array {
			index, err := operationDirectWriteBodyArrayIndex(parts[position])
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				return
			}
			child, err := operationDirectWriteBodyArrayItemSchema(node, index)
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				return
			}
			visit(child, position+1, append(steps, operationDirectWriteBodyPathStep{index: index, array: true}))
			return
		}
		if !object {
			if firstErr == nil {
				firstErr = fmt.Errorf("descends into scalar schema at %q", parts[position])
			}
			return
		}
		properties, ok := node["properties"].(map[string]any)
		if !ok {
			if firstErr == nil {
				firstErr = fmt.Errorf("does not declare nested body fields at %q", parts[position])
			}
			return
		}
		for _, name := range sortedMapKeys(properties) {
			nameParts := strings.Split(name, ".")
			if position+len(nameParts) > len(parts) {
				continue
			}
			matched := true
			for index, namePart := range nameParts {
				if parts[position+index] != namePart {
					matched = false
					break
				}
			}
			if !matched {
				continue
			}
			child, ok := properties[name].(map[string]any)
			if !ok {
				if firstErr == nil {
					firstErr = fmt.Errorf("declared schema for %q must be an object", name)
				}
				continue
			}
			visit(child, position+len(nameParts), append(steps, operationDirectWriteBodyPathStep{key: name}))
		}
	}
	visit(root, 0, nil)
	if len(candidates) == 1 {
		return candidates[0], nil
	}
	if len(candidates) > 1 {
		return operationDirectWriteBodyPath{}, fmt.Errorf("body field %q is ambiguous in the declared schema", path)
	}
	if firstErr != nil {
		return operationDirectWriteBodyPath{}, firstErr
	}
	return operationDirectWriteBodyPath{}, fmt.Errorf("additional property %q is not declared", path)
}

func operationDirectWriteBodyPathsOverlap(left, right operationDirectWriteBodyPath) bool {
	if len(left.steps) > len(right.steps) {
		left, right = right, left
	}
	for index, step := range left.steps {
		other := right.steps[index]
		if step.array != other.array || step.key != other.key || step.index != other.index {
			return false
		}
	}
	return true
}

func operationDirectWriteBodyPathLess(left, right operationDirectWriteBodyPath) bool {
	limit := len(left.steps)
	if len(right.steps) < limit {
		limit = len(right.steps)
	}
	for index := 0; index < limit; index++ {
		leftStep := left.steps[index]
		rightStep := right.steps[index]
		if leftStep.array != rightStep.array {
			return !leftStep.array
		}
		if leftStep.array {
			if leftStep.index != rightStep.index {
				return leftStep.index < rightStep.index
			}
			continue
		}
		if leftStep.key != rightStep.key {
			return leftStep.key < rightStep.key
		}
	}
	return len(left.steps) < len(right.steps)
}

func MaterializeOperationDirectWriteBodyMappings(b Bundle, operation string, mappings map[string]any) (map[string]any, error) {
	op, _, err := operationDirectWriteSpec(b, operation)
	if err != nil {
		return nil, err
	}
	if op.Kind == "graphql_mutation" {
		return cloneAnyMap(mappings), nil
	}
	if len(mappings) == 0 && op.REST != nil && len(op.REST.BodySchema) == 0 {
		return map[string]any{}, nil
	}
	root, err := operationDirectWriteBodySchemaRoot(op)
	if err != nil {
		return nil, err
	}
	body := make(map[string]any, len(mappings))
	if OperationDirectWriteHasStructuredRESTBody(op) {
		compiled, err := compileStructuredRESTBodySchema(op)
		if err != nil {
			return nil, err
		}
		staticBody, err := canonicalizeStructuredRESTBodyFragment(compiled, op, op.REST.Body, "rest.body")
		if err != nil {
			return nil, err
		}
		if len(staticBody) != 0 && len(mappings) != 0 {
			shape, ok := operationDirectWriteStaticBodyScaffold(staticBody)
			shapeMap, okMap := shape.(map[string]any)
			if !ok || !okMap {
				return nil, fmt.Errorf("operation %q rest.body shape must be an object", op.ID)
			}
			body = shapeMap
		}
	}
	resolved := make([]operationDirectWriteBodyPath, 0, len(mappings))
	for _, path := range sortedMapKeys(mappings) {
		candidate, err := resolveOperationDirectWriteBodySchemaPath(root, path)
		if err != nil {
			return nil, fmt.Errorf("operation %q body field %q: %w", op.ID, path, err)
		}
		for _, previous := range resolved {
			if operationDirectWriteBodyPathsOverlap(previous, candidate) {
				return nil, fmt.Errorf("operation %q maps overlapping body fields %q and %q", op.ID, previous.raw, path)
			}
		}
		resolved = append(resolved, candidate)
	}
	sort.SliceStable(resolved, func(left, right int) bool {
		return operationDirectWriteBodyPathLess(resolved[left], resolved[right])
	})
	for _, candidate := range resolved {
		if _, err := setOperationDirectWriteBodyPathValue(body, candidate.steps, mappings[candidate.raw], candidate.raw); err != nil {
			return nil, err
		}
	}
	return body, nil
}

func ResolveOperationDirectWriteBodyMappingValue(b Bundle, operation string, body map[string]any, path string) (any, bool, error) {
	op, _, err := operationDirectWriteSpec(b, operation)
	if err != nil {
		return nil, false, err
	}
	if op.Kind != "rest_write" || len(op.REST.BodySchema) == 0 {
		value, found := operationDirectWriteLiteralBodyValue(body, path)
		return value, found, nil
	}
	root, err := operationDirectWriteBodySchemaRoot(op)
	if err != nil {
		return nil, false, err
	}
	resolved, err := resolveOperationDirectWriteBodySchemaPath(root, path)
	if err != nil {
		return nil, false, fmt.Errorf("operation %q body field %q: %w", op.ID, path, err)
	}
	value, found := operationDirectWriteBodyPathValue(body, resolved.steps)
	return value, found, nil
}

func WithholdOperationDirectWriteBodyFields(b Bundle, operation string, body map[string]any, fields []string) (map[string]any, []string, error) {
	_, root, out, err := operationDirectWriteCanonicalBodyPlanFragment(b, operation, body)
	if err != nil {
		return nil, nil, err
	}
	withheld := make([]string, 0, len(fields))
	for _, rawField := range fields {
		field := operationDirectWriteBodyRelativePath(rawField)
		if field == "" {
			continue
		}
		resolved, err := resolveOperationDirectWriteBodySchemaPath(root, field)
		if err != nil {
			return nil, nil, fmt.Errorf("operation %q sensitive body field %q: %w", operation, field, err)
		}
		_, removed, err := deleteOperationDirectWriteBodyPathValue(out, resolved.steps, field)
		if err != nil {
			return nil, nil, err
		}
		if removed {
			withheld = append(withheld, field)
		}
	}
	if len(withheld) == 0 {
		return out, nil, nil
	}
	return out, withheld, nil
}

func RedactOperationDirectWriteBodyFields(b Bundle, operation string, body map[string]any, fields []string) (map[string]any, error) {
	_, root, out, err := operationDirectWriteCanonicalBodyPlanFragment(b, operation, body)
	if err != nil {
		return nil, err
	}
	for _, rawField := range fields {
		field := operationDirectWriteBodyRelativePath(rawField)
		if field == "" {
			continue
		}
		resolved, err := resolveOperationDirectWriteBodySchemaPath(root, field)
		if err != nil {
			return nil, fmt.Errorf("operation %q sensitive body field %q: %w", operation, field, err)
		}
		if _, found := operationDirectWriteBodyPathValue(out, resolved.steps); !found {
			continue
		}
		updated, err := setOperationDirectWriteBodyPathValue(out, resolved.steps, "redacted", field)
		if err != nil {
			return nil, err
		}
		bodyMap, ok := updated.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("operation %q redacted body must be an object", operation)
		}
		out = bodyMap
	}
	return out, nil
}

func MergeOperationDirectWriteBodyFragments(b Bundle, operation string, base, overlay map[string]any) (map[string]any, error) {
	_, root, err := operationDirectWriteStructuredBodyPlanRoot(b, operation)
	if err != nil {
		return nil, err
	}
	merged, err := mergeOperationDirectWriteBodyFragmentValue(root, base, true, overlay, true, "body")
	if err != nil {
		return nil, err
	}
	body, ok := merged.(map[string]any)
	if !ok || body == nil {
		return nil, fmt.Errorf("operation %q merged body must be an object", operation)
	}
	return body, nil
}

func OperationDirectWriteBodyPathContains(b Bundle, operation, parent, child string) (bool, error) {
	_, root, err := operationDirectWriteStructuredBodyPlanRoot(b, operation)
	if err != nil {
		return false, err
	}
	parentPath, err := resolveOperationDirectWriteBodySchemaPath(root, operationDirectWriteBodyRelativePath(parent))
	if err != nil {
		return false, fmt.Errorf("operation %q body field %q: %w", operation, parent, err)
	}
	childPath, err := resolveOperationDirectWriteBodySchemaPath(root, operationDirectWriteBodyRelativePath(child))
	if err != nil {
		return false, fmt.Errorf("operation %q body field %q: %w", operation, child, err)
	}
	if len(parentPath.steps) > len(childPath.steps) {
		return false, nil
	}
	for index, step := range parentPath.steps {
		other := childPath.steps[index]
		if step.array != other.array || step.key != other.key || step.index != other.index {
			return false, nil
		}
	}
	return true, nil
}

func operationDirectWriteStructuredBodyPlanRoot(b Bundle, operation string) (OperationSpec, map[string]any, error) {
	op, _, err := operationDirectWriteSpec(b, operation)
	if err != nil {
		return OperationSpec{}, nil, err
	}
	if op.Kind != "rest_write" || !OperationDirectWriteHasStructuredRESTBody(op) {
		return OperationSpec{}, nil, fmt.Errorf("operation %q does not expose a structured REST body", operation)
	}
	root, err := operationDirectWriteBodySchemaRoot(op)
	if err != nil {
		return OperationSpec{}, nil, err
	}
	return op, root, nil
}

func operationDirectWriteCanonicalBodyPlanFragment(b Bundle, operation string, body map[string]any) (OperationSpec, map[string]any, map[string]any, error) {
	op, root, err := operationDirectWriteStructuredBodyPlanRoot(b, operation)
	if err != nil {
		return OperationSpec{}, nil, nil, err
	}
	compiled, err := compileStructuredRESTBodySchema(op)
	if err != nil {
		return OperationSpec{}, nil, nil, err
	}
	canonical, err := canonicalizeStructuredRESTBodyFragment(compiled, op, body, "body")
	if err != nil {
		return OperationSpec{}, nil, nil, err
	}
	return op, root, canonical, nil
}

func operationDirectWriteBodyRelativePath(path string) string {
	return strings.TrimPrefix(strings.TrimSpace(path), "body.")
}

func deleteOperationDirectWriteBodyPathValue(current any, steps []operationDirectWriteBodyPathStep, path string) (any, bool, error) {
	if len(steps) == 0 {
		return current, false, nil
	}
	step := steps[0]
	if step.array {
		items, ok := current.([]any)
		if !ok {
			if current == nil {
				return current, false, nil
			}
			return nil, false, fmt.Errorf("body field %q conflicts with existing non-array value", path)
		}
		if step.index < 0 || step.index >= len(items) || items[step.index] == nil {
			return items, false, nil
		}
		if len(steps) == 1 {
			items[step.index] = nil
			return items, true, nil
		}
		updated, removed, err := deleteOperationDirectWriteBodyPathValue(items[step.index], steps[1:], path)
		if err != nil {
			return nil, false, err
		}
		if removed {
			items[step.index] = updated
		}
		return items, removed, nil
	}
	object, ok := current.(map[string]any)
	if !ok {
		if current == nil {
			return current, false, nil
		}
		return nil, false, fmt.Errorf("body field %q conflicts with existing non-object value", path)
	}
	value, found := object[step.key]
	if !found || value == nil {
		return object, false, nil
	}
	if len(steps) == 1 {
		delete(object, step.key)
		return object, true, nil
	}
	updated, removed, err := deleteOperationDirectWriteBodyPathValue(value, steps[1:], path)
	if err != nil {
		return nil, false, err
	}
	if removed {
		object[step.key] = updated
	}
	return object, removed, nil
}

func mergeOperationDirectWriteBodyFragmentValue(node map[string]any, base any, hasBase bool, overlay any, hasOverlay bool, path string) (any, error) {
	if !hasOverlay {
		return cloneOperationDirectWriteBodyValue(base), nil
	}
	if !hasBase {
		return cloneOperationDirectWriteBodyValue(overlay), nil
	}
	object, array := operationDirectWriteBodyNodeKinds(node)
	if object && array {
		return nil, fmt.Errorf("%s has an ambiguous declared schema", path)
	}
	if object {
		baseObject, ok := operationDirectWriteBodyObject(base)
		if !ok {
			return nil, fmt.Errorf("%s conflicts with existing non-object value", path)
		}
		overlayObject, ok := operationDirectWriteBodyObject(overlay)
		if !ok {
			return nil, fmt.Errorf("%s conflicts with a non-object replacement", path)
		}
		properties, ok := node["properties"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s has no declared object properties", path)
		}
		merged := cloneOperationDirectWriteBodyMap(baseObject)
		for _, name := range sortedMapKeys(overlayObject) {
			child, ok := properties[name].(map[string]any)
			if !ok {
				return nil, fmt.Errorf("%s.%s is not declared", path, name)
			}
			baseValue, basePresent := merged[name]
			value, err := mergeOperationDirectWriteBodyFragmentValue(child, baseValue, basePresent, overlayObject[name], true, path+"."+name)
			if err != nil {
				return nil, err
			}
			merged[name] = value
		}
		return merged, nil
	}
	if array {
		baseItems, ok := arrayElements(base)
		if !ok {
			return nil, fmt.Errorf("%s conflicts with existing non-array value", path)
		}
		overlayItems, ok := arrayElements(overlay)
		if !ok {
			return nil, fmt.Errorf("%s conflicts with a non-array replacement", path)
		}
		length := len(baseItems)
		if len(overlayItems) > length {
			length = len(overlayItems)
		}
		merged := make([]any, length)
		for index := 0; index < length; index++ {
			basePresent := index < len(baseItems) && baseItems[index] != nil
			overlayPresent := index < len(overlayItems)
			if !overlayPresent {
				if basePresent {
					merged[index] = cloneOperationDirectWriteBodyValue(baseItems[index])
				}
				continue
			}
			item, err := operationDirectWriteBodyArrayItemSchema(node, index)
			if err != nil {
				return nil, err
			}
			var baseValue any
			if basePresent {
				baseValue = baseItems[index]
			}
			value, err := mergeOperationDirectWriteBodyFragmentValue(item, baseValue, basePresent, overlayItems[index], true, fmt.Sprintf("%s.%d", path, index))
			if err != nil {
				return nil, err
			}
			merged[index] = value
		}
		return merged, nil
	}
	return cloneOperationDirectWriteBodyValue(overlay), nil
}

func operationDirectWriteBodyObject(value any) (map[string]any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		return typed, true
	case connectors.Record:
		return map[string]any(typed), true
	default:
		return nil, false
	}
}

func cloneOperationDirectWriteBodyMap(value map[string]any) map[string]any {
	out := make(map[string]any, len(value))
	for key, item := range value {
		out[key] = cloneOperationDirectWriteBodyValue(item)
	}
	return out
}

func cloneOperationDirectWriteBodyValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneOperationDirectWriteBodyMap(typed)
	case connectors.Record:
		return cloneOperationDirectWriteBodyMap(map[string]any(typed))
	case []any:
		out := make([]any, len(typed))
		for index, item := range typed {
			out[index] = cloneOperationDirectWriteBodyValue(item)
		}
		return out
	default:
		return value
	}
}

func operationDirectWriteBodyPathValue(body map[string]any, steps []operationDirectWriteBodyPathStep) (any, bool) {
	var current any = body
	for _, step := range steps {
		if step.array {
			items, ok := current.([]any)
			if !ok || step.index < 0 || step.index >= len(items) || items[step.index] == nil {
				return nil, false
			}
			current = items[step.index]
			continue
		}
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		next, ok := object[step.key]
		if !ok || next == nil {
			return nil, false
		}
		current = next
	}
	return current, true
}

func operationDirectWriteLiteralBodyValue(body map[string]any, path string) (any, bool) {
	if body == nil || strings.TrimSpace(path) == "" {
		return nil, false
	}
	var current any = body
	for _, part := range strings.Split(path, ".") {
		switch value := current.(type) {
		case map[string]any:
			next, ok := value[part]
			if !ok || next == nil {
				return nil, false
			}
			current = next
		case []any:
			index, err := operationDirectWriteBodyArrayIndex(part)
			if err != nil || index >= len(value) || value[index] == nil {
				return nil, false
			}
			current = value[index]
		default:
			return nil, false
		}
	}
	return current, true
}

func setOperationDirectWriteBodyPathValue(current any, steps []operationDirectWriteBodyPathStep, value any, path string) (any, error) {
	if len(steps) == 0 {
		return value, nil
	}
	step := steps[0]
	if step.array {
		items, ok := current.([]any)
		if !ok {
			return nil, fmt.Errorf("body field %q conflicts with existing non-array value", path)
		}
		if step.index > len(items) {
			return nil, fmt.Errorf("body field %q uses sparse array index %d", path, step.index)
		}
		if step.index == len(items) {
			items = append(items, nil)
		}
		if len(steps) == 1 {
			items[step.index] = value
			return items, nil
		}
		child := items[step.index]
		if child == nil {
			child = operationDirectWriteBodyPathContainer(steps[1])
		}
		updated, err := setOperationDirectWriteBodyPathValue(child, steps[1:], value, path)
		if err != nil {
			return nil, err
		}
		items[step.index] = updated
		return items, nil
	}
	object, ok := current.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("body field %q conflicts with existing non-object value", path)
	}
	if len(steps) == 1 {
		object[step.key] = value
		return object, nil
	}
	child, ok := object[step.key]
	if !ok {
		child = operationDirectWriteBodyPathContainer(steps[1])
	}
	updated, err := setOperationDirectWriteBodyPathValue(child, steps[1:], value, path)
	if err != nil {
		return nil, err
	}
	object[step.key] = updated
	return object, nil
}

func operationDirectWriteBodyPathContainer(step operationDirectWriteBodyPathStep) any {
	if step.array {
		return []any{}
	}
	return map[string]any{}
}

func operationDirectWriteBodyNodeKinds(node map[string]any) (object, array bool) {
	object = isObjectType(node)
	array = isArrayType(node)
	if !object {
		_, object = node["additionalProperties"]
	}
	if !array {
		_, array = node["prefixItems"]
	}
	return object, array
}

func operationDirectWriteBodyArrayIndex(value string) (int, error) {
	if value == "" {
		return 0, fmt.Errorf("array body field index is required")
	}
	if len(value) > 1 && strings.HasPrefix(value, "0") {
		return 0, fmt.Errorf("array body field index %q must not have leading zeroes", value)
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("array body field index %q must be numeric", value)
		}
	}
	index, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("array body field index %q is invalid", value)
	}
	return index, nil
}

func operationDirectWriteBodyArrayItemSchema(node map[string]any, index int) (map[string]any, error) {
	if rawMaxItems, ok := node["maxItems"]; ok {
		maxItems, ok := rawMaxItems.(float64)
		if !ok || math.Trunc(maxItems) != maxItems || maxItems < 0 {
			return nil, fmt.Errorf("array body field has invalid maxItems")
		}
		if maxItems > maxStructuredRESTBodyItems {
			return nil, fmt.Errorf("array body field maxItems %.0f exceeds structured body limit %d", maxItems, maxStructuredRESTBodyItems)
		}
		if index >= int(maxItems) {
			return nil, fmt.Errorf("array body field index %d exceeds declared maxItems %.0f", index, maxItems)
		}
	}
	if rawPrefixItems, ok := node["prefixItems"]; ok {
		prefixItems, ok := rawPrefixItems.([]any)
		if !ok {
			return nil, fmt.Errorf("prefixItems must be an array of schema objects")
		}
		if index < len(prefixItems) {
			item, ok := prefixItems[index].(map[string]any)
			if !ok {
				return nil, fmt.Errorf("prefixItems/%d must be a schema object", index)
			}
			return item, nil
		}
	}
	rawItems, ok := node["items"]
	if !ok {
		return nil, fmt.Errorf("array body field has no item schema")
	}
	items, ok := rawItems.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("array body field items must be a schema object")
	}
	return items, nil
}

func validateOperationDirectWriteBodyOverrides(op OperationSpec, overrides map[string]any) error {
	if len(overrides) == 0 {
		return nil
	}
	root, err := operationDirectWriteBodySchemaRoot(op)
	if err != nil {
		return err
	}
	names := make([]string, 0, len(overrides))
	for name := range overrides {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		node, err := operationDirectWriteBodySchemaPath(root, name)
		if err != nil {
			return fmt.Errorf("operation %q body field %q: %w", op.ID, name, err)
		}
		if err := validateOperationDirectWriteBodyOverrideValue(node, overrides[name], name); err != nil {
			return fmt.Errorf("operation %q body field %q: %w", op.ID, name, err)
		}
	}
	return nil
}

func validateOperationDirectWriteBodyOverrideValue(node map[string]any, value any, path string) error {
	if value == nil {
		return nil
	}
	objectNode, arrayNode := operationDirectWriteBodyNodeKinds(node)
	if object, ok := value.(map[string]any); ok {
		if !objectNode || arrayNode {
			return fmt.Errorf("does not permit a nested object")
		}
		names := make([]string, 0, len(object))
		for name := range object {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			next, err := operationDirectWriteBodySchemaPath(node, name)
			if err != nil {
				return err
			}
			if err := validateOperationDirectWriteBodyOverrideValue(next, object[name], path+"."+name); err != nil {
				return err
			}
		}
		return nil
	}
	if values, ok := arrayElements(value); ok {
		if !arrayNode || objectNode {
			return fmt.Errorf("does not permit a nested array")
		}
		for index, item := range values {
			next, err := operationDirectWriteBodyArrayItemSchema(node, index)
			if err != nil {
				return err
			}
			if err := validateOperationDirectWriteBodyOverrideValue(next, item, fmt.Sprintf("%s.%d", path, index)); err != nil {
				return err
			}
		}
	}
	return nil
}

func operationDirectWriteRedactFields(op OperationSpec) []string {
	if op.SensitivePolicy == nil {
		return nil
	}
	return append([]string(nil), op.SensitivePolicy.RedactFields...)
}

func operationDirectWriteRedactionValues(op OperationSpec, body map[string]any) ([]string, error) {
	// A live secret operation retains complete runtime output for diagnosis;
	// secrecy is provided by the encrypted credential store, not deletion from
	// responses, errors, logs, previews, reports, or fixtures.
	if operationRetainsSecretRuntimeContent(op) {
		return nil, nil
	}
	if op.SensitivePolicy == nil || len(op.SensitivePolicy.RedactFields) == 0 {
		return nil, nil
	}
	if op.Kind == "rest_write" && OperationDirectWriteHasStructuredRESTBody(op) {
		root, err := operationDirectWriteBodySchemaRoot(op)
		if err != nil {
			return nil, err
		}
		seen := make(map[string]bool)
		for _, rawField := range op.SensitivePolicy.RedactFields {
			field := operationDirectWriteBodyRelativePath(rawField)
			if field == "" {
				continue
			}
			resolved, err := resolveOperationDirectWriteBodySchemaPath(root, field)
			if err != nil {
				return nil, fmt.Errorf("operation %q sensitive body field %q: %w", op.ID, field, err)
			}
			value, found := operationDirectWriteBodyPathValue(body, resolved.steps)
			if found {
				collectWriteRedactionValues(value, seen)
			}
		}
		values := make([]string, 0, len(seen))
		for value := range seen {
			values = append(values, value)
		}
		sortWriteRedactionLiterals(values)
		return values, nil
	}
	fields := make([]string, 0, len(op.SensitivePolicy.RedactFields))
	for _, field := range op.SensitivePolicy.RedactFields {
		fields = append(fields, strings.TrimPrefix(strings.TrimSpace(field), "body."))
	}
	return writeActionRedactionValues(WriteAction{RedactFields: fields}, connectors.Record(body)), nil
}

// operationDirectWritePayloadFileFields keeps multipart file identity
// discovery declaration-owned. Returning a non-nil empty slice for a multipart
// operation distinguishes it from the legacy name-based fallback used by
// non-multipart direct writes.
func operationDirectWritePayloadFileFields(op OperationSpec) []string {
	if op.REST == nil || op.REST.Multipart == nil {
		return nil
	}
	seen := make(map[string]struct{})
	fields := make([]string, 0, len(op.REST.Multipart.Parts))
	for _, part := range op.REST.Multipart.Parts {
		if part.Type != "file" {
			continue
		}
		field := strings.TrimSpace(part.Field)
		if field == "" {
			continue
		}
		if _, duplicate := seen[field]; duplicate {
			continue
		}
		seen[field] = struct{}{}
		fields = append(fields, field)
	}
	sort.Strings(fields)
	return fields
}

func operationDirectWritePayloadFileMaxBytes(op OperationSpec) map[string]int64 {
	if op.REST == nil || op.REST.Multipart == nil {
		return nil
	}
	maxBytes := make(map[string]int64)
	for _, part := range op.REST.Multipart.Parts {
		if part.Type != "file" {
			continue
		}
		field := strings.TrimSpace(part.Field)
		if field == "" {
			continue
		}
		limit := int64(part.MaxBytes)
		if current, found := maxBytes[field]; !found || limit < current {
			maxBytes[field] = limit
		}
	}
	return maxBytes
}

func operationDirectWriteHasSensitiveHTTPBinding(cfg connectors.RuntimeConfig, httpBase HTTPBase, headers map[string]string) (bool, error) {
	spec, found, err := selectedOperationDirectWriteAuthSpec(cfg, httpBase.Auth)
	if err != nil {
		return false, err
	}
	if found && !strings.EqualFold(strings.TrimSpace(spec.Mode), "none") {
		return true, nil
	}
	return operationDirectWriteHeaderTemplatesReferenceRuntimeValues(headers)
}

func operationDirectWriteHeaderTemplatesReferenceRuntimeValues(headers map[string]string) (bool, error) {
	names := make([]string, 0, len(headers))
	for name := range headers {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		tokens, err := parseWriteQueryTemplate(headers[name])
		if err != nil {
			return false, err
		}
		for _, token := range tokens {
			if token.expression == "" {
				continue
			}
			if operationDirectWriteHeaderExpressionReferencesRuntimeValues(token.expression) {
				return true, nil
			}
		}
	}
	return false, nil
}

func operationDirectWriteHeaderExpressionReferencesRuntimeValues(expr string) bool {
	if _, coalesced, err := coalesceRecordPathsExpression(expr); coalesced || err != nil {
		return false
	}
	ref := strings.TrimSpace(strings.Split(expr, "|")[0])
	parts := strings.Split(ref, ".")
	return len(parts) == 2 && (parts[0] == "config" || parts[0] == "secrets")
}

func bindOperationDirectWriteHTTPMutations(cfg connectors.RuntimeConfig, httpBase HTTPBase, headers, query map[string]string) (map[string]string, map[string]string, []AuthSpec, []AuthSpec, error) {
	boundHeaders := cloneResolvedHeaders(headers)
	if boundHeaders == nil {
		boundHeaders = make(map[string]string)
	}
	boundQuery := make(map[string]string, len(query)+1)
	for name, value := range query {
		boundQuery[name] = value
	}
	if userAgent := httpBase.UserAgent; userAgent != "" {
		if err := bindOperationDirectWriteHeader(boundHeaders, "User-Agent", userAgent, "declared user agent"); err != nil {
			return nil, nil, nil, nil, err
		}
	}
	spec, found, err := selectedOperationDirectWriteAuthSpec(cfg, httpBase.Auth)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if !found {
		return boundHeaders, boundQuery, nil, nil, nil
	}
	rateLimitAuth := []AuthSpec{spec}
	switch spec.Mode {
	case "none":
		return boundHeaders, boundQuery, nil, rateLimitAuth, nil
	case "bearer":
		token, err := interpolateOperationDirectWriteAuthValue(spec.Token, cfg)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("bearer token: %w", err)
		}
		if err := bindOperationDirectWriteHeader(boundHeaders, "Authorization", "Bearer "+strings.TrimSpace(token), "declared bearer authentication"); err != nil {
			return nil, nil, nil, nil, err
		}
		return boundHeaders, boundQuery, nil, rateLimitAuth, nil
	case "basic":
		username, err := interpolateOperationDirectWriteAuthValue(spec.Username, cfg)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("basic username: %w", err)
		}
		password, err := interpolateOperationDirectWriteAuthValue(spec.Password, cfg)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("basic password: %w", err)
		}
		value := "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
		if err := bindOperationDirectWriteHeader(boundHeaders, "Authorization", value, "declared basic authentication"); err != nil {
			return nil, nil, nil, nil, err
		}
		return boundHeaders, boundQuery, nil, rateLimitAuth, nil
	case "api_key_header":
		value, err := interpolateOperationDirectWriteAuthValue(spec.Value, cfg)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("api_key_header value: %w", err)
		}
		if err := safety.RejectDangerousChars(spec.Prefix, "api_key_header prefix"); err != nil || strings.ContainsAny(spec.Prefix, "\r\n") {
			if err != nil {
				return nil, nil, nil, nil, err
			}
			return nil, nil, nil, nil, fmt.Errorf("api_key_header prefix contains CR/LF")
		}
		if _, err := canonicalPreparedRequestHeaderName(spec.Header); err != nil {
			return nil, nil, nil, nil, fmt.Errorf("declared API key authentication header: %w", err)
		}
		headerValue := spec.Prefix + strings.TrimSpace(value)
		if headerValue == "" {
			return boundHeaders, boundQuery, nil, rateLimitAuth, nil
		}
		if err := bindOperationDirectWriteHeader(boundHeaders, spec.Header, headerValue, "declared API key authentication"); err != nil {
			return nil, nil, nil, nil, err
		}
		return boundHeaders, boundQuery, nil, rateLimitAuth, nil
	case "api_key_query":
		name := spec.Param
		if err := safety.ValidateIdentifier(name, "auth query parameter"); err != nil {
			return nil, nil, nil, nil, err
		}
		value, err := interpolateOperationDirectWriteAuthValue(spec.Value, cfg)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("api_key_query value: %w", err)
		}
		if _, exists := boundQuery[name]; exists {
			return nil, nil, nil, nil, fmt.Errorf("declared API key query parameter %q collides with a prepared operation query value", name)
		}
		boundQuery[name] = strings.TrimSpace(value)
		return boundHeaders, boundQuery, nil, rateLimitAuth, nil
	case "oauth2_client_credentials", "oauth2_refresh_token", "custom":
		return boundHeaders, boundQuery, []AuthSpec{spec}, rateLimitAuth, nil
	default:
		return nil, nil, nil, nil, fmt.Errorf("unknown auth mode %q", spec.Mode)
	}
}

func selectedOperationDirectWriteAuthSpec(cfg connectors.RuntimeConfig, specs []AuthSpec) (AuthSpec, bool, error) {
	for _, spec := range specs {
		matched, err := authSpecMatches(spec, authVars(cfg))
		if err != nil {
			return AuthSpec{}, false, fmt.Errorf("select auth mode %q: %w", spec.Mode, err)
		}
		if matched {
			return spec, true, nil
		}
	}
	if len(specs) == 0 {
		return AuthSpec{}, false, nil
	}
	return AuthSpec{}, false, fmt.Errorf("select auth: no auth spec matched for auth_type %q", cfg.Config["auth_type"])
}

func interpolateOperationDirectWriteAuthValue(template string, cfg connectors.RuntimeConfig) (string, error) {
	if err := validateOperationDirectWriteAuthTemplate(template); err != nil {
		return "", err
	}
	return interpolateDeclaredHeader(template, authVars(cfg))
}

func validateOperationDirectWriteAuthTemplate(template string) error {
	tokens, err := parseWriteQueryTemplate(template)
	if err != nil {
		return err
	}
	for _, token := range tokens {
		if token.expression == "" {
			continue
		}
		if err := validateDeclaredHeaderExpression(token.expression); err != nil {
			return err
		}
		parts := strings.Split(token.expression, "|")
		ref := strings.TrimSpace(parts[0])
		segments := strings.Split(ref, ".")
		if len(segments) != 2 || (segments[0] != "config" && segments[0] != "secrets") {
			return fmt.Errorf("interpolate: direct-write authentication reference %q must use config or secrets", ref)
		}
	}
	return nil
}

func bindOperationDirectWriteHeader(headers map[string]string, name, value, source string) error {
	canonicalName, err := canonicalPreparedRequestHeaderName(name)
	if err != nil {
		return fmt.Errorf("%s header: %w", source, err)
	}
	if _, exists := headers[canonicalName]; exists {
		return fmt.Errorf("%s header %q collides with an already prepared header", source, canonicalName)
	}
	if strings.ContainsAny(value, "\r\n") {
		return fmt.Errorf("%s header %q contains CR/LF", source, canonicalName)
	}
	if err := safety.RejectDangerousChars(value, "header "+canonicalName); err != nil {
		return err
	}
	headers[canonicalName] = value
	return nil
}

func prepareOperationDirectWrite(ctx context.Context, b Bundle, req connectors.OperationDirectWriteRequest, _ Hooks) (preparedOperationDirectWrite, error) {
	if err := ctx.Err(); err != nil {
		return preparedOperationDirectWrite{}, err
	}
	op, method, err := operationDirectWriteSpec(b, req.Operation)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	if op.Kind == "graphql_mutation" {
		return prepareOperationGraphQLDirectWrite(b, op, method, req)
	}
	if err := validateOperationDirectWritePathParams(op, req.PathParams); err != nil {
		return preparedOperationDirectWrite{}, err
	}
	cfg := materializeConfigDefaults(b, sealRuntimeConfig(req.Config))
	identity := operationDirectWriteIdentity(b, op, method)
	headers, err := resolveDirectWriteHeaders(b.HTTP.Headers, cfg, b.Spec)
	if err != nil {
		return preparedOperationDirectWrite{}, fmt.Errorf("%s: resolve declared headers: %w", identity, err)
	}
	resolvedPath, err := resolveSurfaceEndpointPath(op.REST.Path, cfg, req.PathParams)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	authQueryParameters, err := OperationDirectWriteAuthOwnedQueryParameters(op, b.HTTP.Auth)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	queryMap, err := operationDirectWriteQuery(op, req.Query, authQueryParameters)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	if err := requireOperationQueryGroups(op, queryMap); err != nil {
		return preparedOperationDirectWrite{}, err
	}
	headers, queryMap, runtimeAuth, rateLimitAuth, err := bindOperationDirectWriteHTTPMutations(cfg, b.HTTP, headers, queryMap)
	if err != nil {
		return preparedOperationDirectWrite{}, fmt.Errorf("%s: bind declared HTTP mutations: %w", identity, err)
	}
	sensitiveHTTPBinding, err := operationDirectWriteHasSensitiveHTTPBinding(cfg, b.HTTP, b.HTTP.Headers)
	if err != nil {
		return preparedOperationDirectWrite{}, fmt.Errorf("%s: inspect declared HTTP bindings: %w", identity, err)
	}
	query, err := directReadQuery(queryMap)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	operationHeaders, err := operationRequestHeaders(b, op, req.Headers, req.HeaderValues)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	preparedHeaders, preparedHeaderValues, err := preparedOperationDirectWriteHeaders(headers, operationHeaders)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	policy := strings.TrimSpace(op.OutputPolicy)
	if requested := strings.TrimSpace(req.OutputPolicy); requested != "" {
		if requested != policy {
			return preparedOperationDirectWrite{}, fmt.Errorf("operation %q output_policy must match declared policy %q", op.ID, policy)
		}
		policy = requested
	}
	if err := validateOperationDirectWriteOutputPolicy(policy); err != nil {
		return preparedOperationDirectWrite{}, err
	}
	if policy == directWritePolicySecretStored && cfg.SecretStore == nil {
		return preparedOperationDirectWrite{}, fmt.Errorf("operation %q secret response requires a credential secret store", op.ID)
	}
	body, err := operationWriteBody(op, req.Body)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	body, err = applyOperationSensitiveTransform(op, body)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	contentType, format, err := operationDirectWriteContentType(op, body)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	maxBytes := clampOperationDirectWriteMaxBytes(op.REST.MaxBytes)
	var form url.Values
	var encodedBody string
	if format == "multipart" {
		canonical, canonicalErr := prepareCanonicalOperationMultipart(op, connectors.Record(body), 0, cfg, true)
		if canonicalErr != nil {
			return preparedOperationDirectWrite{}, canonicalErr
		}
		raw, marshalErr := json.Marshal(canonical)
		if marshalErr != nil {
			return preparedOperationDirectWrite{}, fmt.Errorf("operation %q: encode canonical multipart request: %w", op.ID, marshalErr)
		}
		encodedBody = string(raw)
	} else {
		form, encodedBody, err = operationDirectWritePreparedBody(op, body, format, maxBytes)
		if err != nil {
			return preparedOperationDirectWrite{}, err
		}
	}
	baseURL, err := operationDirectWriteBaseURL(b, cfg, op, identity)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	requestPath := normalizeDirectReadPathForBaseURL(resolvedPath, baseURL)
	targetURL, err := operationDirectWriteRequestURL(baseURL, requestPath, query, identity)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	targetURLWithoutQuery, err := operationDirectWriteRequestURL(baseURL, requestPath, nil, identity)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	target := DestructiveTargetForOperation(b.Name, op)
	definition := map[string]any{
		"kind":          op.Kind,
		"operation":     op.ID,
		"content_type":  contentType,
		"output_policy": policy,
	}
	if err := bindOperationSensitiveTransform(definition, op); err != nil {
		return preparedOperationDirectWrite{}, err
	}
	if singles := operationSingleHeaders(operationHeaders); len(singles) > 0 {
		definition["headers"] = singles
	}
	if repeated := operationRepeatedHeaders(operationHeaders); len(repeated) > 0 {
		definition["header_values"] = repeated
	}
	if format == "multipart" {
		// Bind the whole fixed upload declaration, including absent optional
		// parts and their limits, rather than only the values present in this
		// invocation's canonical body.
		definition["multipart"] = op.REST.Multipart
	}
	prepared := PreparedWrite{
		Target:              target,
		CredentialRevision:  cfg.CredentialRevision,
		ConfigurationDigest: cfg.ConfigurationDigest,
		ApprovalScope:       cfg.WriteApprovalScope,
		Batchable:           op.IsBatchable(),
		RecordsStaged:       1,
		Action:              op.ID,
		Warnings:            []string{fmt.Sprintf("prepared rest_write operation %q (%s)", op.ID, method)},
		Definition:          definition,
		Requests: []PreparedRequest{{
			Method:       method,
			URL:          targetURL,
			Target:       targetURLWithoutQuery,
			Query:        query.Encode(),
			ContentType:  contentType,
			BodyFormat:   format,
			Body:         encodedBody,
			Headers:      preparedHeaders,
			HeaderValues: preparedHeaderValues,
		}},
	}
	redactionValues, err := operationDirectWriteRedactionValues(op, body)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	return preparedOperationDirectWrite{
		op:                   op,
		cfg:                  cfg,
		baseURL:              baseURL,
		headers:              cloneResolvedHeaders(headers),
		operationHeaders:     cloneOperationHeaders(operationHeaders),
		runtimeAuth:          runtimeAuth,
		rateLimitAuth:        rateLimitAuth,
		identity:             identity,
		method:               method,
		path:                 resolvedPath,
		requestPath:          requestPath,
		query:                query,
		body:                 body,
		form:                 form,
		format:               format,
		contentType:          contentType,
		policy:               policy,
		maxBytes:             maxBytes,
		redactionValues:      redactionValues,
		sensitiveHTTPBinding: sensitiveHTTPBinding,
		prepared:             prepared,
	}, nil
}

// prepareOperationGraphQLDirectWrite binds one fixed GraphQL mutation to the
// same PreparedWrite used by REST operations. The caller can supply only the
// declaration's closed variables object: path, query, document, selection,
// headers, and endpoint are all declaration-owned facts.
func prepareOperationGraphQLDirectWrite(b Bundle, op OperationSpec, method string, req connectors.OperationDirectWriteRequest) (preparedOperationDirectWrite, error) {
	if len(req.PathParams) != 0 {
		return preparedOperationDirectWrite{}, fmt.Errorf("operation %q fixed GraphQL mutation does not accept path parameters", op.ID)
	}
	if len(req.Query) != 0 {
		return preparedOperationDirectWrite{}, fmt.Errorf("operation %q fixed GraphQL mutation does not accept query overrides", op.ID)
	}
	if len(req.Headers) != 0 || len(req.HeaderValues) != 0 {
		return preparedOperationDirectWrite{}, fmt.Errorf("operation %q fixed GraphQL mutation does not accept request header overrides", op.ID)
	}
	cfg := materializeConfigDefaults(b, sealRuntimeConfig(req.Config))
	identity := operationDirectWriteIdentity(b, op, method)
	headers, err := resolveDirectWriteHeaders(b.HTTP.Headers, cfg, b.Spec)
	if err != nil {
		return preparedOperationDirectWrite{}, fmt.Errorf("%s: resolve declared headers: %w", identity, err)
	}
	headers, queryMap, runtimeAuth, rateLimitAuth, err := bindOperationDirectWriteHTTPMutations(cfg, b.HTTP, headers, nil)
	if err != nil {
		return preparedOperationDirectWrite{}, fmt.Errorf("%s: bind declared HTTP mutations: %w", identity, err)
	}
	sensitiveHTTPBinding, err := operationDirectWriteHasSensitiveHTTPBinding(cfg, b.HTTP, b.HTTP.Headers)
	if err != nil {
		return preparedOperationDirectWrite{}, fmt.Errorf("%s: inspect declared HTTP bindings: %w", identity, err)
	}
	query, err := directReadQuery(queryMap)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	policy := strings.TrimSpace(op.OutputPolicy)
	if requested := strings.TrimSpace(req.OutputPolicy); requested != "" {
		if requested != policy {
			return preparedOperationDirectWrite{}, fmt.Errorf("operation %q output_policy must match declared policy %q", op.ID, policy)
		}
		policy = requested
	}
	if err := validateOperationDirectWriteOutputPolicy(policy); err != nil {
		return preparedOperationDirectWrite{}, err
	}
	variables, err := graphQLOperationVariables(op, req.Body, 0, "")
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	variables, err = applyOperationSensitiveTransform(op, variables)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	maxBytes := clampOperationDirectWriteMaxBytes(op.GraphQL.MaxBytes)
	payload, encodedBody, err := buildGraphQLOperationPayload(op, variables, maxBytes)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	baseURL, err := operationDirectWriteBaseURL(b, cfg, op, identity)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	requestPath := normalizeDirectReadPathForBaseURL(op.GraphQL.Path, baseURL)
	targetURL, err := operationDirectWriteRequestURL(baseURL, requestPath, query, identity)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	targetURLWithoutQuery, err := operationDirectWriteRequestURL(baseURL, requestPath, nil, identity)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	target := DestructiveTargetForOperation(b.Name, op)
	definition := map[string]any{
		"kind":              op.Kind,
		"operation":         op.ID,
		"content_type":      "application/json",
		"output_policy":     policy,
		"graphql_operation": op.GraphQL.OperationName,
	}
	if err := bindOperationSensitiveTransform(definition, op); err != nil {
		return preparedOperationDirectWrite{}, err
	}
	prepared := PreparedWrite{
		Target:              target,
		CredentialRevision:  cfg.CredentialRevision,
		ConfigurationDigest: cfg.ConfigurationDigest,
		ApprovalScope:       cfg.WriteApprovalScope,
		Batchable:           op.IsBatchable(),
		RecordsStaged:       1,
		Action:              op.ID,
		Warnings:            []string{fmt.Sprintf("prepared graphql_mutation operation %q (%s)", op.ID, method)},
		Definition:          definition,
		Requests: []PreparedRequest{{
			Method:      method,
			URL:         targetURL,
			Target:      targetURLWithoutQuery,
			Query:       query.Encode(),
			ContentType: "application/json",
			BodyFormat:  "json",
			Body:        encodedBody,
			Headers:     cloneResolvedHeaders(headers),
		}},
	}
	redactionValues, err := operationDirectWriteRedactionValues(op, variables)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	return preparedOperationDirectWrite{
		op:                   op,
		cfg:                  cfg,
		baseURL:              baseURL,
		headers:              cloneResolvedHeaders(headers),
		runtimeAuth:          runtimeAuth,
		rateLimitAuth:        rateLimitAuth,
		identity:             identity,
		method:               method,
		path:                 op.GraphQL.Path,
		requestPath:          requestPath,
		query:                query,
		body:                 payload,
		format:               "graphql",
		contentType:          "application/json",
		policy:               policy,
		maxBytes:             maxBytes,
		redactionValues:      redactionValues,
		sensitiveHTTPBinding: sensitiveHTTPBinding,
		prepared:             prepared,
	}, nil
}

func operationDirectWriteSpec(b Bundle, id string) (OperationSpec, string, error) {
	op, err := findOperation(b, id)
	if err != nil {
		return OperationSpec{}, "", err
	}
	if err := validateOperationDirectWriteDeclaredHeaders(b.HTTP.Headers); err != nil {
		return OperationSpec{}, "", err
	}
	if err := validateOperationDirectWriteBaseURLTemplate(b.HTTP.URL); err != nil {
		return OperationSpec{}, "", fmt.Errorf("declared base URL: %w", err)
	}
	if _, _, _, err := operationSensitiveTransform(op); err != nil {
		return OperationSpec{}, "", err
	}
	switch op.Kind {
	case "rest_write":
		if op.REST == nil {
			return OperationSpec{}, "", fmt.Errorf("operation direct write rest_write operation has no REST declaration")
		}
		if err := validateOperationRouteForOperation(b, op.Route, op.ID, op.REST.Path, op.SourceURL); err != nil {
			return OperationSpec{}, "", err
		}
		method := strings.ToUpper(strings.TrimSpace(op.REST.Method))
		if !isOperationDirectWriteMethod(method) {
			return OperationSpec{}, "", fmt.Errorf("operation direct write requires POST, PUT, PATCH, or DELETE, got %s", method)
		}
		if err := validateOperationResponseSecretContract(op); err != nil {
			return OperationSpec{}, "", err
		}
		if isAbsoluteHTTPURL(op.REST.Path) {
			return OperationSpec{}, "", fmt.Errorf("operation direct write endpoint must be connector-relative, got absolute URL")
		}
		if err := requireOperationDirectWriteEndpoint(b, method, op.REST.Path, ""); err != nil {
			return OperationSpec{}, "", err
		}
		if _, err := operationDirectWriteQueryParameters(op); err != nil {
			return OperationSpec{}, "", err
		}
		if _, err := OperationDirectWriteAuthOwnedQueryParameters(op, b.HTTP.Auth); err != nil {
			return OperationSpec{}, "", err
		}
		if _, _, err := operationDirectWriteContentType(op, map[string]any{"declared": true}); err != nil {
			return OperationSpec{}, "", err
		}
		if OperationDirectWriteHasStructuredRESTBody(op) {
			if _, err := compileStructuredRESTBodySchema(op); err != nil {
				return OperationSpec{}, "", err
			}
		}
		return op, method, nil
	case "graphql_mutation":
		if op.GraphQL == nil {
			return OperationSpec{}, "", fmt.Errorf("operation direct write graphql_mutation operation has no GraphQL declaration")
		}
		if err := validateOperationRouteForOperation(b, op.Route, op.ID, op.GraphQL.Path, op.SourceURL); err != nil {
			return OperationSpec{}, "", err
		}
		if err := validateGraphQLOperationDirectContract(op, "mutation"); err != nil {
			return OperationSpec{}, "", err
		}
		if err := requireOperationDirectWriteEndpoint(b, http.MethodPost, op.GraphQL.Path, op.ID); err != nil {
			return OperationSpec{}, "", err
		}
		return op, http.MethodPost, nil
	default:
		return OperationSpec{}, "", fmt.Errorf("operation direct write requires rest_write or graphql_mutation operation, got %q", op.Kind)
	}
}

func validateOperationDirectWriteDeclaredHeaders(headers map[string]string) error {
	if _, err := canonicalPreparedRequestHeaders(headers); err != nil {
		return fmt.Errorf("declared headers: %w", err)
	}
	names := make([]string, 0, len(headers))
	for name := range headers {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := validateDeclaredHeaderTemplate(headers[name]); err != nil {
			return fmt.Errorf("declared header %q: %w", name, err)
		}
	}
	return nil
}

func isOperationDirectWriteMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func requireOperationDirectWriteEndpoint(b Bundle, method, endpointPath, operation string) error {
	surface := b.Surface
	if surface == nil {
		// defs.FS omits api_surface.json. The runtime fallback is the endpoint
		// projection derived from its own shipped rest_write declarations.
		surface = b.directWriteSurface
	}
	if surface == nil {
		return fmt.Errorf("api_surface is required for direct-write endpoint %s %s", method, endpointPath)
	}
	for _, endpoint := range surface.Endpoints {
		if strings.EqualFold(endpoint.Method, method) && endpoint.Path == endpointPath {
			if operation != "" {
				for _, target := range endpoint.CoveredBy.OperationTargets() {
					if target == operation {
						return nil
					}
				}
				return fmt.Errorf("api_surface endpoint %s %s does not cover GraphQL operation %q", method, endpointPath, operation)
			}
			if endpoint.Operation == nil {
				return fmt.Errorf("api_surface endpoint %s %s is not declared as an operation", method, endpointPath)
			}
			return nil
		}
	}
	return fmt.Errorf("api_surface endpoint %s %s not found", method, endpointPath)
}

func operationWriteBody(op OperationSpec, overrides map[string]any) (map[string]any, error) {
	if op.REST == nil {
		return nil, nil
	}
	structured := operationDirectWriteUsesJSONBody(op) && operationHasStructuredRESTBodyField(op)
	if structured {
		body, err := materializeStructuredRESTBody(op, op.REST.Body, overrides)
		if err != nil {
			return nil, err
		}
		if len(body) == 0 {
			return nil, nil
		}
		return body, nil
	}
	canonicalOverrides := cloneAnyMap(overrides)
	if len(canonicalOverrides) > 0 {
		if len(op.REST.BodySchema) == 0 {
			return nil, fmt.Errorf("operation %q does not permit caller body fields without body_schema", op.ID)
		}
		if operationDirectWriteUsesJSONBody(op) {
			normalized, err := normalizeStructuredRESTBodyValue(canonicalOverrides, "body")
			if err != nil {
				return nil, fmt.Errorf("operation %q: %w", op.ID, err)
			}
			var ok bool
			canonicalOverrides, ok = normalized.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("operation %q: body must be an object", op.ID)
			}
		}
		if !structured {
			if err := validateOperationDirectWriteBodyOverrides(op, canonicalOverrides); err != nil {
				return nil, err
			}
		}
	}
	body := cloneAnyMap(op.REST.Body)
	for key, value := range canonicalOverrides {
		body[key] = value
	}
	if len(op.REST.BodySchema) > 0 {
		sch, err := CompileSchema(op.REST.BodySchema)
		if err != nil {
			return nil, fmt.Errorf("operation %q: compile body_schema: %w", op.ID, err)
		}
		if err := sch.Validate(body); err != nil {
			return nil, fmt.Errorf("operation %q: body_schema: %w", op.ID, err)
		}
	}
	if len(body) == 0 {
		return nil, nil
	}
	return body, nil
}

func operationDirectWriteUsesJSONBody(op OperationSpec) bool {
	if op.REST == nil || op.REST.Multipart != nil {
		return false
	}
	_, _, err := operationStructuredJSONContentType(op)
	return err == nil
}

func operationDirectWriteQueryParameters(op OperationSpec) (map[string]OperationParameter, error) {
	if op.Kind != "rest_write" || op.REST == nil {
		return nil, nil
	}
	return operationParametersForLocation(op, "query")
}

func validateOperationDirectWriteQueryFieldsWithAuth(op OperationSpec, queryFields []string, authQueryParameters map[string]struct{}) error {
	if op.Kind != "rest_write" || op.REST == nil {
		if len(queryFields) == 0 {
			return nil
		}
		return fmt.Errorf("operation %q does not permit caller query fields", op.ID)
	}
	parameters, err := operationDirectWriteQueryParameters(op)
	if err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(queryFields))
	for _, rawName := range queryFields {
		name := rawName
		if err := safety.ValidateIdentifier(name, "operation query parameter"); err != nil {
			return fmt.Errorf("operation %q query field: %w", op.ID, err)
		}
		if _, duplicate := seen[name]; duplicate {
			return fmt.Errorf("operation %q maps more than one command flag to query parameter %q", op.ID, name)
		}
		seen[name] = struct{}{}
		if _, authOwned := authQueryParameters[name]; authOwned {
			return fmt.Errorf("operation %q query parameter %q is owned by declared API key authentication and cannot be caller-bound", op.ID, name)
		}
		if _, fixed := op.REST.Query[name]; fixed {
			return fmt.Errorf("operation %q query parameter %q is fixed by rest.query and cannot be caller-bound", op.ID, name)
		}
		if _, declared := parameters[name]; !declared {
			return fmt.Errorf("operation %q query parameter %q is not source-declared in rest.parameters", op.ID, name)
		}
	}
	parameterNames := make([]string, 0, len(parameters))
	for name := range parameters {
		parameterNames = append(parameterNames, name)
	}
	sort.Strings(parameterNames)
	for _, name := range parameterNames {
		parameter := parameters[name]
		if !parameter.Required {
			continue
		}
		if _, fixed := op.REST.Query[name]; fixed {
			continue
		}
		if _, authOwned := authQueryParameters[name]; authOwned {
			continue
		}
		if _, bound := seen[name]; !bound {
			return fmt.Errorf("operation %q requires query parameter %q", op.ID, name)
		}
	}
	return nil
}

// OperationDirectWriteAuthOwnedQueryParameters identifies query values that
// only the declaration-selected API-key auth may provide. A required source
// parameter is rejected when a selectable auth branch can omit it: accepting
// that declaration would make a safety-marked operation unreachable at
// runtime, or tempt callers to supply an auth-owned raw query value.
func OperationDirectWriteAuthOwnedQueryParameters(op OperationSpec, specs []AuthSpec) (map[string]struct{}, error) {
	parameters := make(map[string]struct{})
	selectable, noMatchPossible := operationDirectWriteSelectableAuthSpecs(specs)
	for _, spec := range selectable {
		if strings.EqualFold(strings.TrimSpace(spec.Mode), "api_key_query") {
			if name := strings.TrimSpace(spec.Param); name != "" {
				parameters[name] = struct{}{}
			}
		}
	}
	if op.Kind != "rest_write" || op.REST == nil || len(parameters) == 0 {
		return parameters, nil
	}
	declared, err := operationDirectWriteQueryParameters(op)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(parameters))
	for name := range parameters {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		parameter, found := declared[name]
		if !found || !parameter.Required {
			continue
		}
		if noMatchPossible {
			return nil, fmt.Errorf("operation %q required query parameter %q is conditionally supplied by declared API key authentication; every selectable auth rule must supply it", op.ID, name)
		}
		for _, spec := range selectable {
			if strings.EqualFold(strings.TrimSpace(spec.Mode), "api_key_query") && strings.TrimSpace(spec.Param) == name {
				continue
			}
			return nil, fmt.Errorf("operation %q required query parameter %q is conditionally supplied by declared API key authentication; every selectable auth rule must supply it", op.ID, name)
		}
	}
	return parameters, nil
}

func operationDirectWriteSelectableAuthSpecs(specs []AuthSpec) ([]AuthSpec, bool) {
	selectable := make([]AuthSpec, 0, len(specs))
	for _, spec := range specs {
		selectable = append(selectable, spec)
		if strings.TrimSpace(spec.When) == "" {
			return selectable, false
		}
	}
	return selectable, len(selectable) != 0
}

func operationDirectWriteQuery(op OperationSpec, requested map[string]string, authQueryParameters map[string]struct{}) (map[string]string, error) {
	parameters, err := operationDirectWriteQueryParameters(op)
	if err != nil {
		return nil, err
	}
	requestedNames := make([]string, 0, len(requested))
	for name := range requested {
		requestedNames = append(requestedNames, name)
	}
	sort.Strings(requestedNames)
	if err := validateOperationDirectWriteQueryFieldsWithAuth(op, requestedNames, authQueryParameters); err != nil {
		return nil, err
	}
	query := make(map[string]string, len(op.REST.Query)+len(requested))
	fixedNames := make([]string, 0, len(op.REST.Query))
	for name := range op.REST.Query {
		fixedNames = append(fixedNames, name)
	}
	sort.Strings(fixedNames)
	for _, name := range fixedNames {
		value := op.REST.Query[name]
		if parameter, declared := parameters[name]; declared {
			if err := validateOperationDirectWriteQueryValue(op, parameter, value); err != nil {
				return nil, err
			}
		}
		query[name] = value
	}
	for _, name := range requestedNames {
		value := requested[name]
		if err := validateOperationDirectWriteQueryValue(op, parameters[name], value); err != nil {
			return nil, err
		}
		query[name] = value
	}
	parameterNames := make([]string, 0, len(parameters))
	for name := range parameters {
		parameterNames = append(parameterNames, name)
	}
	sort.Strings(parameterNames)
	for _, name := range parameterNames {
		parameter := parameters[name]
		if parameter.Required && strings.TrimSpace(query[name]) == "" {
			return nil, fmt.Errorf("operation %q requires query parameter %q", op.ID, name)
		}
	}
	return query, nil
}

func validateOperationDirectWriteQueryValue(op OperationSpec, parameter OperationParameter, value string) error {
	return validateOperationParameterWireValue(op, parameter, "query", value)
}

func operationDirectWriteContentType(op OperationSpec, body map[string]any) (contentType, format string, err error) {
	if op.REST == nil {
		return "", "", fmt.Errorf("operation %q has no rest declaration", op.ID)
	}
	declared := strings.TrimSpace(op.REST.ContentType)
	if op.REST.Multipart != nil {
		if op.REST.ContentType != "multipart/form-data" {
			return "", "", fmt.Errorf("operation %q rest.multipart requires literal content_type multipart/form-data", op.ID)
		}
		return "multipart/form-data", "multipart", nil
	}
	if declared == "" {
		if len(body) == 0 {
			return "", "none", nil
		}
		return "application/json", "json", nil
	}
	mediaType, _, parseErr := mime.ParseMediaType(declared)
	if parseErr != nil {
		return "", "", fmt.Errorf("operation %q has invalid rest content_type %q: %w", op.ID, declared, parseErr)
	}
	switch strings.ToLower(mediaType) {
	case "application/json", "application/scim+json":
		if len(body) == 0 {
			return "", "none", nil
		}
		return strings.ToLower(mediaType), "json", nil
	case "application/x-www-form-urlencoded":
		if len(body) == 0 {
			return "", "none", nil
		}
		return "application/x-www-form-urlencoded", "form", nil
	default:
		return "", "", fmt.Errorf("operation %q rest_write content_type %q is not supported by the typed executor", op.ID, declared)
	}
}

func operationDirectWritePreparedBody(op OperationSpec, body map[string]any, format string, maxBytes int) (url.Values, string, error) {
	switch format {
	case "none":
		return nil, "", nil
	case "json":
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, "", fmt.Errorf("operation %q: encode JSON request body: %w", op.ID, err)
		}
		if len(raw) > maxBytes {
			return nil, "", fmt.Errorf("operation %q request body too large: %d bytes exceeds limit %d", op.ID, len(raw), maxBytes)
		}
		return nil, string(raw), nil
	case "form":
		form, err := operationDirectWriteForm(body)
		if err != nil {
			return nil, "", fmt.Errorf("operation %q form body: %w", op.ID, err)
		}
		encoded := form.Encode()
		if len(encoded) > maxBytes {
			return nil, "", fmt.Errorf("operation %q request body too large: %d bytes exceeds limit %d", op.ID, len(encoded), maxBytes)
		}
		return form, encoded, nil
	default:
		return nil, "", fmt.Errorf("operation %q has unsupported body format %q", op.ID, format)
	}
}

func operationDirectWriteForm(body map[string]any) (url.Values, error) {
	values := make(map[string]string, len(body))
	for key, value := range body {
		if value == nil {
			continue
		}
		values[key] = stringifyAny(value)
	}
	return directReadQuery(values)
}

func clampOperationDirectWriteMaxBytes(declared int) int {
	if declared <= 0 {
		return defaultOperationDirectWriteMaxBytes
	}
	if declared > maxOperationDirectWriteBytes {
		return maxOperationDirectWriteBytes
	}
	return declared
}

func operationDirectWriteIdentity(b Bundle, op OperationSpec, method string) string {
	path := ""
	if op.Kind == "graphql_mutation" && op.GraphQL != nil {
		path = op.GraphQL.Path
	} else if op.REST != nil {
		path = op.REST.Path
	}
	return fmt.Sprintf("connector %q operation %q %s %s", b.Name, op.ID, strings.ToUpper(method), path)
}

func operationDirectWriteBaseURL(b Bundle, cfg connectors.RuntimeConfig, op OperationSpec, identity string) (string, error) {
	if strings.TrimSpace(op.Route) != "" {
		return resolveOperationRoute(b, cfg, op.Route, op.ID, operationRoutePath(op), op.SourceURL)
	}
	if err := validateOperationDirectWriteBaseURLTemplate(b.HTTP.URL); err != nil {
		return "", operationDirectWriteInterpolationError(identity, "base URL", b.HTTP.URL)
	}
	baseURL, err := interpolateDeclaredHeader(b.HTTP.URL, requestVars(cfg, nil, ""))
	if err != nil {
		return "", operationDirectWriteInterpolationError(identity, "base URL", b.HTTP.URL)
	}
	if strings.TrimSpace(baseURL) == "" {
		return "", fmt.Errorf("%s: declared base URL is required", identity)
	}
	return baseURL, nil
}

func validateOperationDirectWriteBaseURLTemplate(template string) error {
	tokens, err := parseWriteQueryTemplate(template)
	if err != nil {
		return err
	}
	for _, token := range tokens {
		if token.expression == "" {
			continue
		}
		if err := validateDeclaredHeaderExpression(token.expression); err != nil {
			return err
		}
		parts := strings.Split(token.expression, "|")
		ref := strings.TrimSpace(parts[0])
		segments := strings.Split(ref, ".")
		if len(segments) != 2 || (segments[0] != "config" && segments[0] != "secrets") {
			return fmt.Errorf("interpolate: direct-write base URL reference %q must use config or secrets", ref)
		}
	}
	return nil
}

func operationDirectWriteInterpolationError(identity, location, template string) error {
	if reference := operationDirectWriteDeclaredSecretReference(template); reference != "" {
		return fmt.Errorf("%s: resolve declared %s for %s", identity, location, reference)
	}
	return fmt.Errorf("%s: resolve declared %s", identity, location)
}

func operationDirectWriteDeclaredSecretReference(template string) string {
	for {
		start := strings.Index(template, "{{")
		if start < 0 {
			return ""
		}
		template = template[start+2:]
		end := strings.Index(template, "}}")
		if end < 0 {
			return ""
		}
		expression := strings.TrimSpace(template[:end])
		reference, _, _ := strings.Cut(expression, "|")
		reference = strings.TrimSpace(reference)
		parts := strings.Split(reference, ".")
		if len(parts) == 2 && parts[0] == "secrets" && safety.ValidateIdentifier(parts[1], "secret reference") == nil {
			return reference
		}
		template = template[end+2:]
	}
}

func operationDirectWriteRequestURL(baseURL, requestPath string, query url.Values, identity string) (string, error) {
	parsed, err := url.Parse(joinURL(baseURL, requestPath))
	if err != nil {
		return "", fmt.Errorf("%s: declared request URL is invalid", identity)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("%s: declared base URL is invalid", identity)
	}
	if len(query) > 0 {
		existing := parsed.Query()
		for key, values := range query {
			existing.Del(key)
			for _, value := range values {
				existing.Add(key, value)
			}
		}
		parsed.RawQuery = existing.Encode()
	}
	return parsed.String(), nil
}

func validateOperationDirectWriteOutputPolicy(policy string) error {
	switch policy {
	case directWritePolicyNone, directWritePolicyJSON, directWritePolicyJSONRedacted, directWritePolicyWriteResultRedacted, directWritePolicyGongBoundedInputRedacted, directWritePolicySecretStored:
		return nil
	default:
		return fmt.Errorf("operation direct write output policy %q is not supported", policy)
	}
}

type operationDirectWriteResponseBodyResult struct {
	body     any
	present  bool
	raw      string
	encoding string
	bytes    int
}

func operationDirectWriteResultFromResponse(connector string, prepared preparedOperationDirectWrite, response *connsdk.Response, body operationDirectWriteResponseBodyResult) connectors.OperationDirectWriteResult {
	return connectors.OperationDirectWriteResult{
		Connector:              connector,
		Operation:              prepared.op.ID,
		Method:                 prepared.method,
		Path:                   prepared.path,
		ResponseReceived:       true,
		Status:                 response.Status,
		Headers:                writeProviderHeaders(response.Header),
		BodyPresent:            body.present,
		BodyBytes:              body.bytes,
		BodyRaw:                body.raw,
		BodyRawEncoding:        body.encoding,
		Body:                   body.body,
		OutputSecretFields:     operationDirectWriteOutputSecretFields(prepared.op),
		RequestSensitiveValues: append([]string(nil), prepared.redactionValues...),
	}
}

func operationDirectWriteResponseBody(policy string, raw []byte, maxBytes int, headers map[string][]string) (operationDirectWriteResponseBodyResult, error) {
	if err := validateOperationDirectWriteOutputPolicy(policy); err != nil {
		return operationDirectWriteResponseBodyResult{}, err
	}
	declaresJSON := operationDirectWriteResponseDeclaresJSON(policy, headers)
	if policy == directWritePolicyNone && len(raw) == 0 && !declaresJSON {
		return operationDirectWriteResponseBodyResult{}, nil
	}
	result := operationDirectWriteResponseBodyResult{present: len(raw) > 0, bytes: len(raw)}
	if utf8.Valid(raw) {
		if result.present {
			result.raw = string(raw)
			result.encoding = "text"
		}
	} else {
		result.raw = base64.StdEncoding.EncodeToString(raw)
		result.encoding = "base64"
	}
	if len(raw) > maxBytes {
		return result, fmt.Errorf("operation direct write response too large: %d bytes exceeds limit %d", len(raw), maxBytes)
	}
	if !declaresJSON {
		return result, nil
	}
	if !result.present {
		return result, nil
	}
	decoded, err := decodeDirectReadBody(raw, maxBytes)
	if err != nil {
		return result, errors.New("operation direct write response is not JSON")
	}
	result.body = decoded
	return result, nil
}

func operationDirectWriteResponseDeclaresJSON(policy string, headers map[string][]string) bool {
	if present, declared := writeProviderResponseContentType(headers); present {
		return declared
	}
	switch policy {
	case directWritePolicyJSON, directWritePolicyJSONRedacted, directWritePolicyWriteResultRedacted, directWritePolicyGongBoundedInputRedacted, directWritePolicySecretStored:
		return true
	default:
		return false
	}
}

// validateOperationResponseSecretContract makes a response carrying a
// credential store-bound before runtime construction and I/O. The response is
// deliberately still returned intact: the runtime does not redact content.
func validateOperationResponseSecretContract(op OperationSpec) error {
	if !operationRetainsSecretRuntimeContent(op) {
		return nil
	}
	if op.SensitivePolicy == nil {
		return fmt.Errorf("operation %q secret response requires sensitive_policy", op.ID)
	}
	p := op.SensitivePolicy
	if strings.TrimSpace(p.ResponseSecretField) == "" || strings.TrimSpace(p.ResponseSecretStoreKey) == "" {
		return fmt.Errorf("operation %q secret response requires a declared secret-store destination", op.ID)
	}
	if op.OutputPolicy != directWritePolicySecretStored {
		return fmt.Errorf("operation %q secret response must use output_policy %q", op.ID, directWritePolicySecretStored)
	}
	if err := safety.ValidateIdentifier(p.ResponseSecretField, "response secret field"); err != nil {
		return err
	}
	return safety.ValidateIdentifier(p.ResponseSecretStoreKey, "response secret store key")
}

func storeOperationResponseSecret(ctx context.Context, op OperationSpec, cfg connectors.RuntimeConfig, raw []byte) error {
	if err := validateOperationResponseSecretContract(op); err != nil {
		return err
	}
	if cfg.SecretStore == nil {
		return fmt.Errorf("operation %q secret response requires a credential secret store", op.ID)
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(raw, &body); err != nil {
		return fmt.Errorf("operation %q secret response is not a JSON object", op.ID)
	}
	field := op.SensitivePolicy.ResponseSecretField
	encoded, ok := body[field]
	if !ok {
		return fmt.Errorf("operation %q secret response does not contain its declared credential field", op.ID)
	}
	var value string
	if err := json.Unmarshal(encoded, &value); err != nil || value == "" {
		return fmt.Errorf("operation %q secret response credential field is not a non-empty string", op.ID)
	}
	if err := cfg.SecretStore.PutSecret(ctx, op.SensitivePolicy.ResponseSecretStoreKey, value); err != nil {
		return fmt.Errorf("operation %q store returned credential: %w", op.ID, err)
	}
	return nil
}
