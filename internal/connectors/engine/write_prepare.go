package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
	FieldName         string   `json:"field_name"`
	SourcePathDigest  string   `json:"source_path_digest"`
	ContentSHA256     string   `json:"content_sha256,omitempty"`
	ContentType       string   `json:"content_type,omitempty"`
	AllowedMediaTypes []string `json:"allowed_media_types,omitempty"`
	MaxBytes          int64    `json:"max_bytes,omitempty"`
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
	if target.RequiresApproval() && strings.TrimSpace(action.Hook) != "" {
		if _, ok := h.(WriteHook); ok {
			classifier, classified := h.(WriteHookClassifier)
			if !classified || classifier.HandlesWriteAction(action) {
				return PreparedWrite{}, fmt.Errorf("engine: destructive write action %q uses a hook without an exact prepared-request preview", action.Name)
			}
		}
	}

	cfg := materializeConfigDefaults(b, req.Config)
	warnings := []string{fmt.Sprintf("%s executes a live mutation only after approval; dry run performs no external call", action.Name)}
	if len(records) > 0 {
		method, path, err := resolveWriteRequestLine(b, action, records[0], cfg)
		if err != nil {
			return PreparedWrite{}, err
		}
		warnings = append(warnings, fmt.Sprintf("resolved request: %s %s", method, path))
	}
	requests := make([]PreparedRequest, 0, len(records))
	for index, record := range records {
		prepared, err := prepareDeclarativeRequest(b, action, record, index, cfg, target.RequiresApproval())
		if err != nil {
			return PreparedWrite{}, &Error{Connector: b.Name, Action: action.Name, Page: -1, RecordIndex: index, Err: redactWriteActionError(err, action, record)}
		}
		requests = append(requests, prepared)
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
		HookIdentity: hookIdentity,
		Requests:     requests,
	}, nil
}

func prepareDeclarativeRequest(b Bundle, action WriteAction, record connectors.Record, recordIndex int, cfg connectors.RuntimeConfig, requirePayloadApproval bool) (PreparedRequest, error) {
	vars := Vars{Config: cfg.Config, Secrets: cfg.Secrets, Record: map[string]any(record)}
	baseURL, err := Interpolate(b.HTTP.URL, vars)
	if err != nil {
		return PreparedRequest{}, fmt.Errorf("engine: resolve write base url: %w", err)
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
	if action.Multipart == nil {
		return nil, fmt.Errorf("engine: write action %q: multipart spec is required", action.Name)
	}
	fields := map[string]string{}
	files := []canonicalMultipartFile{}
	for _, part := range action.Multipart.Parts {
		value, err := resolveRecordPathValue(map[string]any(record), strings.Split(part.Field, "."))
		if err != nil || value == nil {
			if part.Required {
				return nil, fmt.Errorf("engine: write action %q: multipart part %q is required", action.Name, part.Name)
			}
			continue
		}
		if part.Type == "field" {
			fields[part.Name] = stringifyAny(value)
			continue
		}
		if part.Type != "file" {
			return nil, fmt.Errorf("engine: write action %q: multipart part %q has unsupported type %q", action.Name, part.Name, part.Type)
		}
		path, ok := value.(string)
		if !ok || strings.TrimSpace(path) == "" {
			return nil, fmt.Errorf("engine: write action %q: multipart file part %q requires a file path string", action.Name, part.Name)
		}
		approved := cfg.ApprovedPayloadSHA256[connectors.PayloadApprovalKey(recordIndex, part.Field)]
		if requirePayloadApproval && strings.TrimSpace(approved) == "" {
			return nil, fmt.Errorf("engine: write action %q: multipart file part %q is missing its approved payload digest", action.Name, part.Name)
		}
		files = append(files, canonicalMultipartFile{
			FieldName:         part.Name,
			SourcePathDigest:  digestBytes([]byte(filepath.Clean(path))),
			ContentSHA256:     approved,
			ContentType:       part.ContentType,
			AllowedMediaTypes: append([]string(nil), part.AllowedMediaTypes...),
			MaxBytes:          part.MaxBytes,
		})
	}
	return struct {
		MaxBytes int64                    `json:"max_bytes,omitempty"`
		Fields   map[string]string        `json:"fields,omitempty"`
		Files    []canonicalMultipartFile `json:"files,omitempty"`
	}{MaxBytes: action.Multipart.MaxBytes, Fields: fields, Files: files}, nil
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

func digestBytes(value []byte) string {
	if len(value) == 0 {
		return ""
	}
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}
