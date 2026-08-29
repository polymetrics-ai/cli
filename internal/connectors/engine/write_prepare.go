package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"polymetrics.ai/internal/connectors"
)

type declarativeWriteDefinition struct {
	Connector           string            `json:"connector"`
	Metadata            Metadata          `json:"metadata"`
	HTTP                HTTPBase          `json:"http"`
	Action              WriteAction       `json:"action"`
	Config              map[string]string `json:"config,omitempty"`
	RawSpecDigest       string            `json:"raw_spec_digest,omitempty"`
	RawOperationsDigest string            `json:"raw_operations_digest,omitempty"`
}

type canonicalMultipartFile struct {
	FieldName         string                             `json:"field_name"`
	SourcePathDigest  string                             `json:"source_path_digest"`
	ContentSHA256     string                             `json:"content_sha256,omitempty"`
	ContentType       string                             `json:"content_type,omitempty"`
	MediaPolicy       connectors.BinaryUploadMediaPolicy `json:"media_policy,omitempty"`
	AllowedMediaTypes []string                           `json:"allowed_media_types,omitempty"`
	MaxBytes          int64                              `json:"max_bytes,omitempty"`
}

func prepareDeclarativeWrite(ctx context.Context, b Bundle, req connectors.WriteRequest, records []connectors.Record, h Hooks) (PreparedWrite, error) {
	if err := ValidateWrite(ctx, b, req, records); err != nil {
		return PreparedWrite{}, err
	}
	action, err := findWriteAction(b, req.Action)
	if err != nil {
		return PreparedWrite{}, err
	}
	target := DestructiveTargetForWrite(b.Name, action)
	hookIdentity := strings.TrimSpace(action.Hook)
	if h != nil {
		hookIdentity += ":" + fmt.Sprintf("%T", h) + ":" + h.ConnectorName()
	}
	cfg := materializeConfigDefaults(b, sealRuntimeConfig(req.Config))
	mapped, err := applyWriteRecordHook(h, action, records)
	if err != nil {
		return PreparedWrite{}, &Error{Connector: b.Name, Action: action.Name, Page: -1, RecordIndex: -1, Err: err}
	}
	warnings := []string{fmt.Sprintf("%s executes a live mutation only after approval; dry run performs no external call", action.Name)}
	requests, executionPlan, planned, err := prepareHookWritePlan(b, action, mapped, cfg, target.RequiresApproval(), h)
	if err != nil {
		return PreparedWrite{}, err
	}
	if !planned {
		if legacyWriteHookClaimsAction(h, action) {
			return PreparedWrite{}, fmt.Errorf("engine: write action %q uses a legacy hook without an exact prepared-request plan", action.Name)
		}
		requests, executionPlan, err = prepareOrdinaryWritePlan(b, action, mapped, cfg, target.RequiresApproval())
		if err != nil {
			return PreparedWrite{}, err
		}
	}
	if len(requests) > 0 {
		redactionValues := writeActionRedactionValues(action, mapped[0])
		for _, secret := range cfg.Secrets {
			redactionValues = append(redactionValues, secret)
		}
		warnings = append(warnings, fmt.Sprintf("resolved request: %s %s",
			requests[0].Method, publicWritePreviewURL(requests[0].URL, redactionValues)))
	}
	return PreparedWrite{
		Target:              target,
		CredentialRevision:  cfg.CredentialRevision,
		ConfigurationDigest: cfg.ConfigurationDigest,
		ApprovalScope:       cfg.WriteApprovalScope,
		Batchable:           action.IsBatchable(),
		RecordsStaged:       len(records),
		Action:              action.Name,
		Warnings:            warnings,
		Definition: declarativeWriteDefinition{
			Connector:           b.Name,
			Metadata:            b.Metadata,
			HTTP:                b.HTTP,
			Action:              action,
			Config:              cfg.Config,
			RawSpecDigest:       digestBytes(b.RawSpec),
			RawOperationsDigest: digestBytes(b.RawOperations),
		},
		HookIdentity:     hookIdentity,
		Requests:         requests,
		executionRecords: cloneWriteExecutionRecords(mapped),
		executionPlan:    executionPlan,
		executionConfig:  sealRuntimeConfig(cfg),
	}, nil
}

func publicWritePreviewURL(raw string, redactionValues []string) string {
	parsed, err := url.Parse(raw)
	if err == nil && parsed.User != nil {
		username := "redacted"
		if _, passwordPresent := parsed.User.Password(); passwordPresent {
			parsed.User = url.UserPassword(username, "redacted")
		} else {
			parsed.User = url.User(username)
		}
		raw = parsed.String()
	}
	return redactWriteLiterals(raw, redactionValues)
}

func cloneWriteExecutionRecords(records []connectors.Record) []connectors.Record {
	cloned := make([]connectors.Record, len(records))
	for index, record := range records {
		cloned[index] = connectors.Record(copyRecordMap(map[string]any(record)))
	}
	return cloned
}

const maxPreparedHookRequestsPerRecord = 8

// legacyWriteHookClaimsAction identifies an execution-time-only hook. A hook
// with no classifier is conservative by design: it could handle any action,
// so the engine must refuse it until it implements PreparedWriteHook.
func legacyWriteHookClaimsAction(h Hooks, action WriteAction) bool {
	if h == nil {
		return false
	}
	if _, ok := h.(PreparedWriteHook); ok {
		return false
	}
	if _, ok := h.(WriteHook); !ok {
		return false
	}
	classifier, classified := h.(WriteHookClassifier)
	return !classified || classifier.HandlesWriteAction(action)
}

func prepareOrdinaryWritePlan(b Bundle, action WriteAction, records []connectors.Record, cfg connectors.RuntimeConfig, requirePayloadApproval bool) ([]PreparedRequest, []preparedWriteExecutionRecord, error) {
	requests := make([]PreparedRequest, 0, len(records))
	plan := make([]preparedWriteExecutionRecord, len(records))
	for index, record := range records {
		sealed := connectors.Record(copyRecordMap(map[string]any(record)))
		prepared, err := prepareDeclarativeRequest(b, action, sealed, index, cfg, requirePayloadApproval)
		if err != nil {
			return nil, nil, &Error{Connector: b.Name, Action: action.Name, Page: -1, RecordIndex: index, Err: redactWriteActionError(err, action, sealed)}
		}
		requests = append(requests, prepared)
		plan[index].steps = []preparedWriteExecutionStep{{action: action, record: sealed}}
	}
	return requests, plan, nil
}

// prepareHookWritePlan validates the hook's compact declarative selection and
// turns it into the same PreparedRequest type ordinary writes use. There is no
// hook-provided method, route, header, or body channel: each step resolves a
// named, unhooked WriteAction from this bundle before the preview is minted.
func prepareHookWritePlan(b Bundle, root WriteAction, records []connectors.Record, cfg connectors.RuntimeConfig, requirePayloadApproval bool, h Hooks) ([]PreparedRequest, []preparedWriteExecutionRecord, bool, error) {
	planner, ok := h.(PreparedWriteHook)
	if !ok {
		return nil, nil, false, nil
	}
	planInput := cloneWriteExecutionRecords(records)
	declared, handled, err := planner.PrepareWrite(root, planInput)
	if err != nil {
		return nil, nil, false, &Error{Connector: b.Name, Action: root.Name, Page: -1, RecordIndex: -1, Err: err}
	}
	if !handled {
		if len(declared.Records) != 0 {
			return nil, nil, false, fmt.Errorf("engine: write hook for %q returned a plan without handling the action", root.Name)
		}
		return nil, nil, false, nil
	}
	if len(declared.Records) != len(records) {
		return nil, nil, false, fmt.Errorf("engine: prepared write hook plan has %d records, want %d", len(declared.Records), len(records))
	}

	requests := make([]PreparedRequest, 0, len(records))
	execution := make([]preparedWriteExecutionRecord, len(records))
	for recordIndex, recordPlan := range declared.Records {
		if len(recordPlan.Steps) > maxPreparedHookRequestsPerRecord {
			return nil, nil, false, fmt.Errorf("engine: prepared write hook record %d has %d requests, exceeds limit %d", recordIndex, len(recordPlan.Steps), maxPreparedHookRequestsPerRecord)
		}
		execution[recordIndex].steps = make([]preparedWriteExecutionStep, 0, len(recordPlan.Steps))
		for stepIndex, step := range recordPlan.Steps {
			stepAction, err := findWriteAction(b, strings.TrimSpace(step.Action))
			if err != nil {
				return nil, nil, false, &Error{Connector: b.Name, Action: root.Name, Page: -1, RecordIndex: recordIndex, Err: fmt.Errorf("hook step %d: %w", stepIndex, err)}
			}
			if strings.TrimSpace(stepAction.Hook) != "" && stepAction.Name != root.Name && !step.ResolvedDeclarative {
				return nil, nil, false, &Error{Connector: b.Name, Action: root.Name, Page: -1, RecordIndex: recordIndex, Err: fmt.Errorf("hook step %d selects nested hook action %q", stepIndex, stepAction.Name)}
			}
			sealed := connectors.Record(copyRecordMap(map[string]any(step.Record)))
			binding, err := validatePreparedWriteResponseBinding(step.ResponseBinding, stepAction, stepIndex)
			if err != nil {
				return nil, nil, false, &Error{Connector: b.Name, Action: root.Name, Page: -1, RecordIndex: recordIndex, Err: fmt.Errorf("hook step %d: %w", stepIndex, err)}
			}
			if err := validateWriteActionRecord(stepAction, sealed, nil); err != nil {
				return nil, nil, false, &Error{Connector: b.Name, Action: stepAction.Name, Page: -1, RecordIndex: recordIndex, Err: redactWriteActionError(err, stepAction, sealed)}
			}
			prepared, err := prepareDeclarativeRequest(b, stepAction, sealed, recordIndex, cfg, requirePayloadApproval)
			if err != nil {
				return nil, nil, false, &Error{Connector: b.Name, Action: stepAction.Name, Page: -1, RecordIndex: recordIndex, Err: redactWriteActionError(err, stepAction, sealed)}
			}
			prepared.Action = stepAction.Name
			prepared.ResponseBinding = binding
			if binding != nil {
				prepared.Target = preparedResponseBindingTarget(stepAction, binding)
				prepared.URL = preparedResponseBindingURL(stepAction, binding)
			}
			requests = append(requests, prepared)
			execution[recordIndex].steps = append(execution[recordIndex].steps, preparedWriteExecutionStep{
				action:          stepAction,
				record:          sealed,
				responseBinding: binding,
			})
		}
	}
	return requests, execution, true, nil
}

func validatePreparedWriteResponseBinding(binding *PreparedWriteResponseBinding, action WriteAction, stepIndex int) (*PreparedWriteResponseBinding, error) {
	if binding == nil {
		return nil, nil
	}
	if binding.SourceStep < 0 || binding.SourceStep >= stepIndex {
		return nil, fmt.Errorf("response binding source step %d must name an earlier request", binding.SourceStep)
	}
	if !isPreparedWriteBindingField(binding.Field) || !isPreparedWriteBindingField(binding.TargetField) {
		return nil, errors.New("response binding fields must be simple declared field names")
	}
	for _, field := range action.PathFields {
		if field == binding.TargetField {
			return clonePreparedWriteResponseBinding(binding), nil
		}
	}
	return nil, fmt.Errorf("response binding target %q is not a declared path field of action %q", binding.TargetField, action.Name)
}

func isPreparedWriteBindingField(field string) bool {
	if field == "" || len(field) > 128 {
		return false
	}
	for index := 0; index < len(field); index++ {
		value := field[index]
		if (value < 'a' || value > 'z') &&
			(value < 'A' || value > 'Z') &&
			(value < '0' || value > '9') &&
			value != '_' {
			return false
		}
	}
	return true
}

func preparedResponseBindingTarget(action WriteAction, binding *PreparedWriteResponseBinding) string {
	return fmt.Sprintf("%s: response[%d].%s -> path.%s", action.Name, binding.SourceStep, binding.Field, binding.TargetField)
}

func preparedResponseBindingURL(action WriteAction, binding *PreparedWriteResponseBinding) string {
	// This intentionally is not a sendable provider URL. It is an explicit
	// digest-bound representation of the one deferred path value; execution
	// resolves it only from the previous bounded provider response.
	return fmt.Sprintf("deferred://provider-response/%s/%d/%s/%s", action.Name, binding.SourceStep, binding.Field, binding.TargetField)
}

func prepareDeclarativeRequest(b Bundle, action WriteAction, record connectors.Record, recordIndex int, cfg connectors.RuntimeConfig, requirePayloadApproval bool) (PreparedRequest, error) {
	vars := Vars{Config: cfg.Config, Secrets: cfg.Secrets, Record: map[string]any(record)}
	baseURL, err := resolveWriteActionRoute(b, cfg, action)
	if err != nil {
		return PreparedRequest{}, fmt.Errorf("engine: resolve write route: %w", err)
	}
	path, err := InterpolatePath(action.Path, vars)
	if err != nil {
		return PreparedRequest{}, fmt.Errorf("engine: write action %q: resolve path: %w", action.Name, err)
	}
	query, err := buildWriteQuery(action, vars)
	if err != nil {
		return PreparedRequest{}, err
	}
	fullURL, err := canonicalWriteURL(baseURL, path, query)
	if err != nil {
		return PreparedRequest{}, err
	}
	headers, err := resolveHeaders(b.HTTP.Headers, cfg, b.Spec)
	if err != nil {
		return PreparedRequest{}, err
	}
	prepared := PreparedRequest{
		Method:  methodOrDefault(action.Method),
		URL:     fullURL,
		Target:  fullURL,
		Query:   query.Encode(),
		Headers: headers,
	}
	body, format, contentType, err := prepareCanonicalWriteBody(action, record, recordIndex, cfg, requirePayloadApproval)
	if err != nil {
		return PreparedRequest{}, err
	}
	prepared.Body = body
	prepared.BodyFormat = format
	prepared.ContentType = contentType
	return prepared, nil
}

func canonicalWriteURL(baseURL, path string, query url.Values) (string, error) {
	full := joinURL(baseURL, path)
	parsed, err := url.Parse(full)
	if err != nil {
		return "", fmt.Errorf("engine: parse prepared write url: %w", err)
	}
	if len(query) > 0 {
		merged := parsed.Query()
		for key, values := range query {
			merged.Del(key)
			for _, value := range values {
				merged.Add(key, value)
			}
		}
		parsed.RawQuery = merged.Encode()
	}
	return parsed.String(), nil
}

func prepareCanonicalWriteBody(action WriteAction, record connectors.Record, recordIndex int, cfg connectors.RuntimeConfig, requirePayloadApproval bool) (string, string, string, error) {
	vars := Vars{Config: cfg.Config, Secrets: cfg.Secrets, Record: map[string]any(record)}
	marshalJSON := func(payload any) (string, string, string, error) {
		if payload == nil {
			return "", "none", "", nil
		}
		raw, err := json.Marshal(payload)
		if err != nil {
			return "", "", "", fmt.Errorf("engine: encode prepared write body: %w", err)
		}
		return string(raw), "json", "application/json", nil
	}
	switch bodyTypeOf(action) {
	case "form":
		body := buildForm(record, action.PathFields).Encode()
		if body == "" {
			return "", "none", "", nil
		}
		return body, "form", "application/x-www-form-urlencoded", nil
	case "graphql":
		payload, err := buildGraphQLPayload(action.GraphQL, vars)
		if err != nil {
			return "", "", "", err
		}
		return marshalJSON(payload)
	case "none":
		body := buildBodyFieldsPayload(record, action.BodyFields)
		if len(body) == 0 {
			return "", "none", "", nil
		}
		return marshalJSON(body)
	case "json_array":
		payload, err := buildJSONArrayPayload(action, record)
		if err != nil {
			return "", "", "", err
		}
		return marshalJSON(payload)
	case "multipart":
		payload, err := prepareCanonicalMultipart(action, record, recordIndex, cfg, requirePayloadApproval)
		if err != nil {
			return "", "", "", err
		}
		raw, err := json.Marshal(payload)
		if err != nil {
			return "", "", "", err
		}
		return string(raw), "multipart-canonical-v1", "multipart/form-data", nil
	case "base64_upload":
		if action.Base64Upload != nil && action.Base64Upload.Source == "base64" {
			payload, err := buildBase64UploadPayload(action, record, recordIndex, cfg)
			if err != nil {
				return "", "", "", err
			}
			return marshalJSON(payload)
		}
		payload, err := prepareCanonicalBase64Upload(action, record, recordIndex, cfg, requirePayloadApproval)
		if err != nil {
			return "", "", "", err
		}
		raw, err := json.Marshal(payload)
		if err != nil {
			return "", "", "", err
		}
		return string(raw), "base64-upload-canonical-v1", "application/json", nil
	case "binary_upload":
		payload, err := prepareCanonicalBinaryUpload(action, record, recordIndex, cfg, requirePayloadApproval)
		if err != nil {
			return "", "", "", err
		}
		raw, err := json.Marshal(payload)
		if err != nil {
			return "", "", "", err
		}
		return string(raw), "binary-upload-canonical-v1", "application/octet-stream", nil
	default:
		var body map[string]any
		if len(action.BodyFields) > 0 {
			body = buildBodyFieldsPayload(record, action.BodyFields)
		} else {
			body = buildJSONBody(record, action.PathFields)
		}
		body, err := applyDynamicFields(action, record, body)
		if err != nil {
			return "", "", "", err
		}
		if len(body) == 0 {
			return "", "none", "", nil
		}
		return marshalJSON(body)
	}
}

func prepareCanonicalMultipart(action WriteAction, record connectors.Record, recordIndex int, cfg connectors.RuntimeConfig, requirePayloadApproval bool) (any, error) {
	return prepareCanonicalMultipartSpec(fmt.Sprintf("write action %q", action.Name), action.Multipart, record, recordIndex, cfg, requirePayloadApproval)
}

// prepareCanonicalOperationMultipart gives a declared rest_write the exact
// same preview representation as writes.json multipart actions. The caller
// passes requirePayloadApproval=true: operation-level multipart is only
// executable when every source is already bound to the plan's approved digest,
// even when its mutation class does not independently require confirmation.
func prepareCanonicalOperationMultipart(op OperationSpec, record connectors.Record, recordIndex int, cfg connectors.RuntimeConfig, requirePayloadApproval bool) (any, error) {
	if op.REST == nil {
		return nil, fmt.Errorf("engine: operation %q: rest multipart spec is required", op.ID)
	}
	return prepareCanonicalMultipartSpec(fmt.Sprintf("operation %q", op.ID), op.REST.Multipart, record, recordIndex, cfg, requirePayloadApproval)
}

// prepareCanonicalMultipartSpec is shared by reverse-ETL actions and declared
// rest_write operations. It intentionally carries identities and approved
// hashes, not file bytes or caller-selected wire shape, so the preview digest
// stays stable, bounded, and safe to carry through approval.
func prepareCanonicalMultipartSpec(subject string, multipart *MultipartSpec, record connectors.Record, recordIndex int, cfg connectors.RuntimeConfig, requirePayloadApproval bool) (any, error) {
	if multipart == nil {
		return nil, fmt.Errorf("engine: %s: multipart spec is required", subject)
	}
	fields := map[string]string{}
	files := []canonicalMultipartFile{}
	for _, part := range multipart.Parts {
		value, err := resolveRecordPathValue(map[string]any(record), strings.Split(part.Field, "."))
		if err != nil || value == nil {
			if part.Required {
				return nil, fmt.Errorf("engine: %s: multipart part %q is required", subject, part.Name)
			}
			continue
		}
		if part.Type == "field" {
			fields[part.Name] = stringifyAny(value)
			continue
		}
		if part.Type != "file" {
			return nil, fmt.Errorf("engine: %s: multipart part %q has unsupported type %q", subject, part.Name, part.Type)
		}
		path, ok := value.(string)
		if !ok || strings.TrimSpace(path) == "" {
			return nil, fmt.Errorf("engine: %s: multipart file part %q requires a file path string", subject, part.Name)
		}
		approved := cfg.ApprovedPayloadSHA256[connectors.PayloadApprovalKey(recordIndex, part.Field)]
		if requirePayloadApproval && strings.TrimSpace(approved) == "" {
			return nil, fmt.Errorf("engine: %s: multipart file part %q is missing its approved payload digest", subject, part.Name)
		}
		files = append(files, canonicalMultipartFile{
			FieldName:         part.Name,
			SourcePathDigest:  digestBytes([]byte(filepath.Clean(path))),
			ContentSHA256:     approved,
			ContentType:       part.ContentType,
			MediaPolicy:       part.MediaPolicy,
			AllowedMediaTypes: append([]string(nil), part.AllowedMediaTypes...),
			MaxBytes:          part.MaxBytes,
		})
	}
	return struct {
		MaxBytes int64                    `json:"max_bytes,omitempty"`
		Fields   map[string]string        `json:"fields,omitempty"`
		Files    []canonicalMultipartFile `json:"files,omitempty"`
	}{MaxBytes: multipart.MaxBytes, Fields: fields, Files: files}, nil
}

func prepareCanonicalBase64Upload(action WriteAction, record connectors.Record, recordIndex int, cfg connectors.RuntimeConfig, requirePayloadApproval bool) (any, error) {
	if action.Base64Upload == nil {
		return nil, fmt.Errorf("engine: write action %q: base64_upload spec is required", action.Name)
	}
	spec := action.Base64Upload
	value, err := resolveRecordPathValue(map[string]any(record), []string{spec.SourceField})
	if err != nil {
		return nil, err
	}
	path, ok := value.(string)
	if !ok || strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("engine: write action %q: base64_upload source_field %q requires a non-empty string", action.Name, spec.SourceField)
	}
	approved := cfg.ApprovedPayloadSHA256[connectors.PayloadApprovalKey(recordIndex, spec.SourceField)]
	if requirePayloadApproval && strings.TrimSpace(approved) == "" {
		return nil, fmt.Errorf("engine: write action %q: base64_upload source_field %q is missing its approved payload digest", action.Name, spec.SourceField)
	}
	var body map[string]any
	if len(action.BodyFields) > 0 {
		body = buildBodyFieldsPayload(record, action.BodyFields)
	} else {
		body = buildJSONBody(record, action.PathFields)
	}
	delete(body, spec.SourceField)
	return struct {
		Body             map[string]any   `json:"body,omitempty"`
		ContentField     string           `json:"content_field"`
		ContentSHA256    string           `json:"content_sha256,omitempty"`
		SourcePathDigest string           `json:"source_path_digest"`
		Spec             Base64UploadSpec `json:"spec"`
	}{Body: body, ContentField: spec.ContentField, ContentSHA256: approved, SourcePathDigest: digestBytes([]byte(filepath.Clean(path))), Spec: *spec}, nil
}

func prepareCanonicalBinaryUpload(action WriteAction, record connectors.Record, recordIndex int, cfg connectors.RuntimeConfig, requirePayloadApproval bool) (any, error) {
	spec := action.BinaryUpload
	if spec == nil {
		return nil, fmt.Errorf("engine: write action %q: binary_upload spec is required", action.Name)
	}
	raw, err := resolveRecordPathValue(map[string]any(record), strings.Split(spec.SourceField, "."))
	if err != nil {
		return nil, fmt.Errorf("engine: write action %q: resolve binary_upload source_field %q: %w", action.Name, spec.SourceField, err)
	}
	path, ok := raw.(string)
	if !ok || strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("engine: write action %q: binary_upload source_field %q requires a non-empty file path", action.Name, spec.SourceField)
	}
	approved := cfg.ApprovedPayloadSHA256[connectors.PayloadApprovalKey(recordIndex, spec.SourceField)]
	if requirePayloadApproval && strings.TrimSpace(approved) == "" {
		return nil, fmt.Errorf("engine: write action %q: binary_upload source_field %q is missing its approved payload digest", action.Name, spec.SourceField)
	}
	return struct {
		SourcePathDigest string `json:"source_path_digest"`
		ContentSHA256    string `json:"content_sha256,omitempty"`
		MaxBytes         int64  `json:"max_bytes"`
	}{
		SourcePathDigest: digestBytes([]byte(filepath.Clean(path))),
		ContentSHA256:    approved,
		MaxBytes:         spec.MaxBytes,
	}, nil
}

func digestBytes(value []byte) string {
	if len(value) == 0 {
		return ""
	}
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}
