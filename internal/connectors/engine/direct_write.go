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
)

type preparedOperationDirectWrite struct {
	op          OperationSpec
	cfg         connectors.RuntimeConfig
	method      string
	path        string
	requestPath string
	query       url.Values
	body        map[string]any
	form        url.Values
	format      string
	policy      string
	maxBytes    int
	prepared    PreparedWrite
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

// PreviewOperationDirectWrite prepares a declared rest_write operation without
// constructing a runtime or issuing any network request. Its digest binds the
// exact typed request that OperationDirectWrite may later dispatch.
func PreviewOperationDirectWrite(ctx context.Context, b Bundle, req connectors.OperationDirectWriteRequest, h Hooks) (connectors.WritePreview, error) {
	prepared, err := prepareOperationDirectWrite(ctx, b, req, h)
	if err != nil {
		return connectors.WritePreview{}, err
	}
	return PreviewPreparedWrite(prepared.prepared)
}

// OperationDirectWrite dispatches exactly one typed, declared rest_write
// operation after the preview-bound shared write gate authorizes it. It never
// retries: rest_write definitions carry no idempotency proof in this executor,
// so both transient retries and the requester’s auth-refresh retry are off.
func OperationDirectWrite(ctx context.Context, b Bundle, req connectors.OperationDirectWriteRequest, h Hooks) (connectors.OperationDirectWriteResult, error) {
	prepared, err := prepareOperationDirectWrite(ctx, b, req, h)
	if err != nil {
		return connectors.OperationDirectWriteResult{}, err
	}
	redactFields := operationDirectWriteRedactFields(prepared.op, req.RedactFields)
	redactionValues := operationDirectWriteRedactionValues(prepared.body, redactFields)

	var result connectors.OperationDirectWriteResult
	err = ExecutePreparedWrite(ctx, prepared.prepared, req.Approval, req.PreviewDigest, func(gated context.Context) error {
		// A redirect can replay a POST/PUT/PATCH/DELETE below Requester's retry
		// loop. Reuse the shared prepared-write transport policy to refuse it:
		// every rest_write is exactly the target that preview bound, regardless
		// of whether the mutation also needs destructive confirmation evidence.
		gated = transportpolicy.MarkDestructive(gated)
		requestCtx, cancel := context.WithTimeout(gated, defaultOperationDirectWriteTimeout)
		defer cancel()

		requestBundle := operationDirectWriteTransportBundle(b, prepared.op)
		rt, err := newRuntime(requestCtx, requestBundle, prepared.cfg, h)
		if err != nil {
			return err
		}
		resolvedRequester, err := rt.requesterFor(prepared.method, prepared.op.REST.Path)
		if err != nil {
			return err
		}
		requester := *resolvedRequester
		requester.DisableRetries = true

		var response *connsdk.Response
		switch prepared.format {
		case "form":
			response, err = requester.DoFormLimited(requestCtx, prepared.method, prepared.requestPath, prepared.query, prepared.form, prepared.maxBytes)
		case "json":
			response, err = requester.DoLimited(requestCtx, prepared.method, prepared.requestPath, prepared.query, prepared.body, prepared.maxBytes)
		case "none":
			// prepared.body is a typed nil map for a declared no-body action.
			// Passing that map through an interface would serialize it as JSON
			// null, so pass an untyped nil and preserve the provider's actual
			// empty-body contract.
			response, err = requester.DoLimited(requestCtx, prepared.method, prepared.requestPath, prepared.query, nil, prepared.maxBytes)
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
			class, hint := applyErrorMap(requestBundle.HTTP.ErrorMap, err)
			message := operationDirectWriteErrorText(err, prepared.policy, redactFields, redactionValues)
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
		body, err := operationDirectWriteResponseBody(prepared.policy, response.Body, prepared.maxBytes, redactFields)
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

// operationDirectWriteErrorText retains the request URL and status while
// applying a declared redacted output policy to provider diagnostics. A
// secret-returning operation must never turn an echoed request or response
// into a terminal-visible secret; non-JSON diagnostics therefore remain
// intentionally opaque under json_redacted rather than guessing their shape.
func operationDirectWriteErrorText(err error, policy string, redactFields, redactionValues []string) string {
	var httpErr *connsdk.HTTPError
	if errors.As(err, &httpErr) {
		message := strings.TrimSpace(httpErr.Body)
		if message == "" {
			message = http.StatusText(httpErr.Status)
		}
		if policy != directWritePolicyJSONRedacted {
			return fmt.Sprintf("http %d for %s: %s", httpErr.Status, httpErr.URL, message)
		}
		message = redactOperationDirectWriteErrorBody(message, policy, redactFields)
		message = redactWriteLiterals(message, redactionValues)
		return safety.RedactErrorText(fmt.Sprintf("http %d for %s: %s", httpErr.Status, httpErr.URL, message))
	}
	if policy != directWritePolicyJSONRedacted {
		return err.Error()
	}
	return safety.RedactErrorText(redactWriteLiterals(err.Error(), redactionValues))
}

func redactOperationDirectWriteErrorBody(message, policy string, redactFields []string) string {
	if policy != directWritePolicyJSONRedacted {
		return message
	}
	decoded, err := decodeDirectReadBody([]byte(message), maxOperationDirectWriteBytes)
	if err != nil {
		return "provider response redacted"
	}
	decoded = redactJSONValue(decoded)
	decoded = redactNamedJSONFields(decoded, redactFields)
	encoded, err := json.Marshal(decoded)
	if err != nil {
		return "provider response redacted"
	}
	return string(encoded)
}

func operationDirectWriteRedactFields(op OperationSpec, requested []string) []string {
	seen := make(map[string]struct{}, len(requested))
	fields := make([]string, 0, len(requested))
	add := func(raw string) {
		field := strings.TrimPrefix(strings.TrimSpace(raw), "record.")
		if field == "" {
			return
		}
		if _, duplicate := seen[field]; duplicate {
			return
		}
		seen[field] = struct{}{}
		fields = append(fields, field)
	}
	if op.SensitivePolicy != nil {
		for _, field := range op.SensitivePolicy.RedactFields {
			add(field)
		}
	}
	for _, field := range requested {
		add(field)
	}
	sort.Strings(fields)
	return fields
}

func operationDirectWriteRedactionValues(body map[string]any, fields []string) []string {
	if len(body) == 0 || len(fields) == 0 {
		return nil
	}
	seen := make(map[string]bool)
	for _, field := range fields {
		value, err := resolveRecordPathValue(copyRecordMap(body), strings.Split(field, "."))
		if err != nil {
			continue
		}
		collectWriteRedactionValues(value, seen)
	}
	values := make([]string, 0, len(seen))
	for value := range seen {
		values = append(values, value)
	}
	sortWriteRedactionLiterals(values)
	return values
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
	if _, _, err := operationDirectWriteContentType(op, nil); err != nil {
		return connectors.OperationDirectWriteMetadata{}, err
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
	}, nil
}

// PreflightOperationDirectWrite proves a command's named operation can reach
// the typed direct-write lifecycle with its declared endpoint and output
// policy. It shares operationDirectWriteSpec with execution and is deliberately
// no-network, so command surface sweeps can safely call it for every bundle.
func PreflightOperationDirectWrite(b Bundle, operation, method, endpointPath, outputPolicy string) error {
	op, declaredMethod, err := operationDirectWriteSpec(b, operation)
	if err != nil {
		return err
	}
	if !strings.EqualFold(strings.TrimSpace(method), declaredMethod) {
		return fmt.Errorf("operation direct write method %s does not match declared operation method %s", strings.ToUpper(strings.TrimSpace(method)), declaredMethod)
	}
	if endpointPath != op.REST.Path {
		return fmt.Errorf("operation direct write path %q does not match declared operation path %q", endpointPath, op.REST.Path)
	}
	return validateOperationDirectWriteOutputPolicy(outputPolicy)
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
	baseURL, err := operationDirectWriteBaseURL(b, cfg, op)
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
		op:          op,
		cfg:         cfg,
		method:      method,
		path:        resolvedPath,
		requestPath: requestPath,
		query:       query,
		body:        body,
		form:        form,
		format:      format,
		policy:      policy,
		maxBytes:    maxBytes,
		prepared:    prepared,
	}, nil
}

func operationDirectWriteSpec(b Bundle, id string) (OperationSpec, string, error) {
	op, err := findOperation(b, id)
	if err != nil {
		return OperationSpec{}, "", err
	}
	if op.Kind != "rest_write" || op.REST == nil {
		return OperationSpec{}, "", fmt.Errorf("operation direct write requires rest_write operation, got %q", op.Kind)
	}
	method := strings.ToUpper(strings.TrimSpace(op.REST.Method))
	if !isOperationDirectWriteMethod(method) {
		return OperationSpec{}, "", fmt.Errorf("operation direct write requires POST, PUT, PATCH, or DELETE, got %s", method)
	}
	if isAbsoluteHTTPURL(op.REST.Path) {
		return OperationSpec{}, "", fmt.Errorf("operation direct write endpoint must be connector-relative, got absolute URL")
	}
	hasOperationBaseURL := strings.TrimSpace(op.REST.BaseURL) != ""
	hasOperationAuth := len(op.REST.Auth) > 0
	if hasOperationBaseURL != hasOperationAuth {
		return OperationSpec{}, "", fmt.Errorf("operation direct write requires rest.base_url and rest.auth to be declared together")
	}
	if err := requireOperationDirectWriteEndpoint(b, method, op.REST.Path); err != nil {
		return OperationSpec{}, "", err
	}
	return op, method, nil
}

func isOperationDirectWriteMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func requireOperationDirectWriteEndpoint(b Bundle, method, endpointPath string) error {
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
			if endpoint.Operation == nil && (endpoint.CoveredBy == nil ||
				(endpoint.CoveredBy.DirectWrite == "" && len(endpoint.CoveredBy.DirectWrites) == 0)) {
				return fmt.Errorf("api_surface endpoint %s %s is not declared as an operation or direct_write command", method, endpointPath)
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
	body := cloneAnyMap(op.REST.Body)
	for key, value := range overrides {
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
	case "application/json":
		if len(body) == 0 {
			return "", "none", nil
		}
		return "application/json", "json", nil
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

func operationDirectWriteTransportBundle(b Bundle, op OperationSpec) Bundle {
	if op.REST == nil || strings.TrimSpace(op.REST.BaseURL) == "" {
		return b
	}
	requestBundle := b
	requestBundle.HTTP.URL = op.REST.BaseURL
	requestBundle.HTTP.Auth = append([]AuthSpec(nil), op.REST.Auth...)
	// A customer-hosted operation must not inherit any ordinary API headers:
	// unlike the operation's paired auth declaration, global headers have no
	// origin-specific ownership and can carry secret-derived credentials.
	requestBundle.HTTP.Headers = nil
	return requestBundle
}

func operationDirectWriteBaseURL(b Bundle, cfg connectors.RuntimeConfig, op OperationSpec) (string, error) {
	template := b.HTTP.URL
	if op.REST != nil && strings.TrimSpace(op.REST.BaseURL) != "" {
		template = op.REST.BaseURL
	}
	baseURL, err := Interpolate(template, requestVars(cfg, nil, ""))
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
	case directWritePolicyNone, directWritePolicyJSON, directWritePolicyJSONRedacted, directWritePolicyWriteResultRedacted, directWritePolicyGongBoundedInputRedacted:
		return nil
	default:
		return fmt.Errorf("operation direct write output policy %q is not supported", policy)
	}
}

func operationDirectWriteResponseBody(policy string, raw []byte, maxBytes int, declaredRedactFields ...[]string) (any, error) {
	if policy == directWritePolicyNone {
		return nil, nil
	}
	decoded, err := decodeDirectReadBody(raw, maxBytes)
	if err != nil {
		return nil, fmt.Errorf("operation direct write response is not JSON: %w", err)
	}
	var redactFields []string
	if len(declaredRedactFields) > 0 {
		redactFields = declaredRedactFields[0]
	}
	switch policy {
	case directWritePolicyJSON:
		return decoded, nil
	case directWritePolicyJSONRedacted:
		return redactNamedJSONFields(redactJSONValue(decoded), redactFields), nil
	case
		directWritePolicyWriteResultRedacted,
		directWritePolicyGongBoundedInputRedacted:
		// These legacy policy names retain their established response behavior.
		// json_redacted is the explicit opt-in for a secret-returning response.
		return decoded, nil
	default:
		return nil, fmt.Errorf("operation direct write output policy %q is not supported", policy)
	}
}
