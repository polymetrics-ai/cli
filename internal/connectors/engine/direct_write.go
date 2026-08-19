package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/transportpolicy"
	"polymetrics.ai/internal/safety"
)

const (
	defaultOperationDirectWriteMaxBytes = 1 << 20
	maxOperationDirectWriteBytes        = 16 << 20
	defaultOperationDirectWriteTimeout  = 30 * time.Second

	directWritePolicyNone                     = "none"
	directWritePolicyJSON                     = "json"
	directWritePolicyJSONRedacted             = "json_redacted"
	directWritePolicyWriteResultRedacted      = "write_result_redacted"
	directWritePolicyGongBoundedInputRedacted = "gong_bounded_input_redacted"
	directWritePolicySecretStored             = "secret_stored"
)

type preparedOperationDirectWrite struct {
	op              OperationSpec
	cfg             connectors.RuntimeConfig
	method          string
	path            string
	requestPath     string
	query           url.Values
	body            map[string]any
	form            url.Values
	format          string
	contentType     string
	policy          string
	maxBytes        int
	redactionValues []string
	prepared        PreparedWrite
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

		rt, err := newRuntime(requestCtx, b, prepared.cfg, h)
		if err != nil {
			return err
		}
		resolvedRequester, err := rt.requesterFor(prepared.method, prepared.path)
		if err != nil {
			return err
		}
		requester := *resolvedRequester
		requester.DisableRetries = true

		var response *connsdk.Response
		switch prepared.format {
		case "form":
			response, err = requester.DoFormLimited(requestCtx, prepared.method, prepared.requestPath, prepared.query, prepared.form, prepared.maxBytes)
		case "json", "none", "graphql":
			contentType := prepared.contentType
			if contentType == "" {
				contentType = "application/json"
			}
			response, err = requester.DoJSONLimited(requestCtx, prepared.method, prepared.requestPath, prepared.query, prepared.body, contentType, prepared.maxBytes)
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
		if err != nil {
			class, hint := applyErrorMap(b.HTTP.ErrorMap, err)
			message := operationDirectWriteErrorText(err, prepared.op.Kind == "graphql_mutation" && !operationRetainsSecretRuntimeContent(prepared.op), prepared.redactionValues)
			if hint != "" {
				message += ": " + hint
			}
			if class != "" {
				message = class + ": " + message
			}
			return &operationDirectWriteError{operation: prepared.op.ID, message: message, cause: err}
		}
		if len(response.Body) > prepared.maxBytes {
			return fmt.Errorf("operation direct write response too large: %d bytes exceeds limit %d", len(response.Body), prepared.maxBytes)
		}
		if prepared.op.Kind == "graphql_mutation" {
			data, metadata, parseErr := graphQLOperationResponseWithRuntimeErrorPolicy(response.Body, prepared.maxBytes, operationRetainsSecretRuntimeContent(prepared.op))
			if parseErr != nil {
				return &operationDirectWriteError{operation: prepared.op.ID, message: "GraphQL response: " + parseErr.Error(), cause: parseErr}
			}
			observeGraphQLRateLimit(requestCtx, &requester, response, data)
			if len(metadata.Errors) != 0 {
				return &operationDirectWriteError{operation: prepared.op.ID, message: "graphql errors: " + redactOperationDirectWriteErrorText(graphQLErrorSummary(metadata), !operationRetainsSecretRuntimeContent(prepared.op), prepared.redactionValues)}
			}
			if data == nil {
				return &operationDirectWriteError{operation: prepared.op.ID, message: "GraphQL response has no data"}
			}
			rawData, marshalErr := json.Marshal(data)
			if marshalErr != nil {
				return fmt.Errorf("operation %q encode GraphQL response data: %w", prepared.op.ID, marshalErr)
			}
			body, bodyErr := operationDirectWriteResponseBody(prepared.policy, rawData, prepared.maxBytes)
			if bodyErr != nil {
				return bodyErr
			}
			result = connectors.OperationDirectWriteResult{
				Connector: b.Name,
				Operation: prepared.op.ID,
				Method:    prepared.method,
				Path:      prepared.path,
				Status:    response.Status,
				Body:      body,
				GraphQL:   metadata,
			}
			return nil
		}
		if prepared.policy == directWritePolicySecretStored {
			if err := storeOperationResponseSecret(gated, prepared.op, prepared.cfg, response.Body); err != nil {
				return err
			}
		}
		body, err := operationDirectWriteResponseBody(prepared.policy, response.Body, prepared.maxBytes)
		if err != nil {
			return err
		}
		result = connectors.OperationDirectWriteResult{
			Connector: b.Name,
			Operation: prepared.op.ID,
			Method:    prepared.method,
			Path:      prepared.path,
			Status:    response.Status,
			Body:      body,
		}
		return nil
	})
	if err != nil {
		return connectors.OperationDirectWriteResult{}, err
	}
	return result, nil
}

// operationDirectWriteErrorText preserves the established complete REST write
// diagnostics, but a fixed GraphQL mutation must redact its HTTP error body:
// unlike a successful GraphQL response it has no errors[] envelope sanitizer
// and can otherwise echo a caller value in its raw provider body.
func operationDirectWriteErrorText(err error, redact bool, values []string) string {
	var output string
	var httpErr *connsdk.HTTPError
	if errors.As(err, &httpErr) {
		message := strings.TrimSpace(httpErr.Body)
		if message == "" {
			message = http.StatusText(httpErr.Status)
		}
		output = fmt.Sprintf("http %d for %s: %s", httpErr.Status, httpErr.URL, message)
	} else {
		output = err.Error()
	}
	return redactOperationDirectWriteErrorText(output, redact, values)
}

func redactOperationDirectWriteErrorText(text string, redact bool, values []string) string {
	if !redact && len(values) == 0 {
		return text
	}
	return safety.RedactErrorText(redactWriteLiterals(text, values))
}

// operationRetainsSecretRuntimeContent identifies the closed secret-operation
// path. Its diagnostic output is complete by policy; the declared secret store
// protects the returned credential at rest rather than deleting runtime fields.
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
		PayloadFileFields:     operationDirectWritePayloadFileFields(op),
		RedactFields:          operationDirectWriteRedactFields(op),
	}, nil
}

// PreflightOperationDirectWrite proves an implemented command's exact binding
// can reach the closed REST or fixed-document GraphQL write executor. It is
// deliberately no-network and shares operationDirectWriteSpec with execution,
// so an api_surface row cannot point a command at a different endpoint than
// the preview-bound request the runtime will actually dispatch.
func PreflightOperationDirectWrite(b Bundle, operation, method, endpointPath, outputPolicy string) error {
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
	return validateOperationDirectWriteOutputPolicy(outputPolicy)
}

func operationDirectWriteRedactFields(op OperationSpec) []string {
	if op.SensitivePolicy == nil {
		return nil
	}
	return append([]string(nil), op.SensitivePolicy.RedactFields...)
}

func operationDirectWriteRedactionValues(op OperationSpec, body map[string]any) []string {
	// A live secret operation retains complete runtime output for diagnosis;
	// secrecy is provided by the encrypted credential store, not deletion from
	// responses, errors, logs, previews, reports, or fixtures.
	if op.SecretSensitive || strings.EqualFold(op.MutationClass, "secret") {
		return nil
	}
	if op.SensitivePolicy == nil || len(op.SensitivePolicy.RedactFields) == 0 {
		return nil
	}
	fields := make([]string, 0, len(op.SensitivePolicy.RedactFields))
	for _, field := range op.SensitivePolicy.RedactFields {
		fields = append(fields, strings.TrimPrefix(strings.TrimSpace(field), "body."))
	}
	return writeActionRedactionValues(WriteAction{RedactFields: fields}, connectors.Record(body))
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
	cfg := materializeConfigDefaults(b, req.Config)
	resolvedPath, err := resolveSurfaceEndpointPath(op.REST.Path, cfg, req.PathParams)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	queryMap := make(map[string]string, len(op.REST.Query)+len(req.Query))
	for key, value := range op.REST.Query {
		queryMap[key] = value
	}
	for key, value := range req.Query {
		queryMap[key] = value
	}
	if err := requireOperationQueryGroups(op, queryMap); err != nil {
		return preparedOperationDirectWrite{}, err
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
	if policy == directWritePolicySecretStored && cfg.SecretStore == nil {
		return preparedOperationDirectWrite{}, fmt.Errorf("operation %q secret response requires a credential secret store", op.ID)
	}
	body, err := operationWriteBody(op, req.Body)
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
	baseURL, err := operationDirectWriteBaseURL(b, cfg)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	requestPath := normalizeDirectReadPathForBaseURL(resolvedPath, baseURL)
	targetURL, err := operationDirectWriteRequestURL(baseURL, requestPath, query)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	targetURLWithoutQuery, err := operationDirectWriteRequestURL(baseURL, requestPath, nil)
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
			Method:      method,
			URL:         targetURL,
			Target:      targetURLWithoutQuery,
			Query:       query.Encode(),
			ContentType: contentType,
			BodyFormat:  format,
			Body:        encodedBody,
		}},
	}
	return preparedOperationDirectWrite{
		op:              op,
		cfg:             cfg,
		method:          method,
		path:            resolvedPath,
		requestPath:     requestPath,
		query:           query,
		body:            body,
		form:            form,
		format:          format,
		contentType:     contentType,
		policy:          policy,
		maxBytes:        maxBytes,
		redactionValues: operationDirectWriteRedactionValues(op, body),
		prepared:        prepared,
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
	cfg := materializeConfigDefaults(b, req.Config)
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
	maxBytes := clampOperationDirectWriteMaxBytes(op.GraphQL.MaxBytes)
	payload, encodedBody, err := buildGraphQLOperationPayload(op, variables, maxBytes)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	baseURL, err := operationDirectWriteBaseURL(b, cfg)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	requestPath := normalizeDirectReadPathForBaseURL(op.GraphQL.Path, baseURL)
	targetURL, err := operationDirectWriteRequestURL(baseURL, requestPath, nil)
	if err != nil {
		return preparedOperationDirectWrite{}, err
	}
	target := DestructiveTargetForOperation(b.Name, op)
	prepared := PreparedWrite{
		Target:              target,
		CredentialRevision:  cfg.CredentialRevision,
		ConfigurationDigest: cfg.ConfigurationDigest,
		ApprovalScope:       cfg.WriteApprovalScope,
		Batchable:           op.IsBatchable(),
		RecordsStaged:       1,
		Action:              op.ID,
		Warnings:            []string{fmt.Sprintf("prepared graphql_mutation operation %q (%s)", op.ID, method)},
		Definition: map[string]any{
			"kind":              op.Kind,
			"operation":         op.ID,
			"content_type":      "application/json",
			"output_policy":     policy,
			"graphql_operation": op.GraphQL.OperationName,
		},
		Requests: []PreparedRequest{{
			Method:      method,
			URL:         targetURL,
			Target:      targetURL,
			ContentType: "application/json",
			BodyFormat:  "json",
			Body:        encodedBody,
		}},
	}
	return preparedOperationDirectWrite{
		op:              op,
		cfg:             cfg,
		method:          method,
		path:            op.GraphQL.Path,
		requestPath:     requestPath,
		body:            payload,
		format:          "graphql",
		contentType:     "application/json",
		policy:          policy,
		maxBytes:        maxBytes,
		redactionValues: operationDirectWriteRedactionValues(op, variables),
		prepared:        prepared,
	}, nil
}

func operationDirectWriteSpec(b Bundle, id string) (OperationSpec, string, error) {
	op, err := findOperation(b, id)
	if err != nil {
		return OperationSpec{}, "", err
	}
	switch op.Kind {
	case "rest_write":
		if op.REST == nil {
			return OperationSpec{}, "", fmt.Errorf("operation direct write rest_write operation has no REST declaration")
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
		return op, method, nil
	case "graphql_mutation":
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
	hasStructuredOverride := false
	for field, value := range overrides {
		if !isStructuredJSONBodyValue(value) {
			continue
		}
		hasStructuredOverride = true
		if err := ValidateOperationStructuredJSONBodyField(op, field); err != nil {
			return nil, err
		}
	}
	body := cloneAnyMap(op.REST.Body)
	for key, value := range overrides {
		body[key] = value
	}
	if len(op.REST.BodySchema) > 0 {
		var sch *Schema
		if hasStructuredOverride {
			compiled, err := compileStructuredRESTBodySchema(op)
			if err != nil {
				return nil, err
			}
			sch = compiled.schema
		} else {
			var err error
			sch, err = CompileSchema(op.REST.BodySchema)
			if err != nil {
				return nil, fmt.Errorf("operation %q: compile body_schema: %w", op.ID, err)
			}
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

func operationDirectWriteBaseURL(b Bundle, cfg connectors.RuntimeConfig) (string, error) {
	baseURL, err := Interpolate(b.HTTP.URL, requestVars(cfg, nil, ""))
	if err != nil {
		return "", fmt.Errorf("operation direct write resolve base URL: %w", err)
	}
	if strings.TrimSpace(baseURL) == "" {
		return "", fmt.Errorf("operation direct write base URL is required")
	}
	return baseURL, nil
}

func operationDirectWriteRequestURL(baseURL, requestPath string, query url.Values) (string, error) {
	parsed, err := url.Parse(joinURL(baseURL, requestPath))
	if err != nil {
		return "", fmt.Errorf("resolve operation direct write URL: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("operation direct write base URL is invalid")
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

func operationDirectWriteResponseBody(policy string, raw []byte, maxBytes int) (any, error) {
	if policy == directWritePolicyNone {
		return nil, nil
	}
	decoded, err := decodeDirectReadBody(raw, maxBytes)
	if err != nil {
		return nil, fmt.Errorf("operation direct write response is not JSON: %w", err)
	}
	switch policy {
	case directWritePolicyJSON,
		directWritePolicyJSONRedacted,
		directWritePolicyWriteResultRedacted,
		directWritePolicyGongBoundedInputRedacted,
		directWritePolicySecretStored:
		// The legacy policy names remain valid declaration values, but direct
		// writes retain their complete decoded response content.
		return decoded, nil
	default:
		return nil, fmt.Errorf("operation direct write output policy %q is not supported", policy)
	}
}

// validateOperationResponseSecretContract makes a response carrying a
// credential store-bound before runtime construction and I/O. The response is
// deliberately still returned intact: the runtime does not redact content.
func validateOperationResponseSecretContract(op OperationSpec) error {
	if !op.SecretSensitive && !strings.EqualFold(op.MutationClass, "secret") {
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
