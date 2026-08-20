package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/textproto"
	"net/url"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/safety"
)

type PreparedRequest struct {
	Method      string `json:"method"`
	URL         string `json:"url"`
	Target      string `json:"target,omitempty"`
	Query       string `json:"query,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	BodyFormat  string `json:"body_format,omitempty"`
	Body        string `json:"body,omitempty"`
	// Headers and HeaderValues are preview-bound through the digest projection
	// below, but never serialized with a prepared request. Runtime-owned auth
	// and declaration-bound request headers may contain a credential value.
	Headers      map[string]string   `json:"-"`
	HeaderValues map[string][]string `json:"-"`
}

type PreparedWrite struct {
	Target              DestructiveTarget
	CredentialRevision  string
	ConfigurationDigest string
	ApprovalScope       string
	Batchable           bool
	RecordsStaged       int
	Action              string
	Warnings            []string
	Definition          any
	HookIdentity        string
	Requests            []PreparedRequest
}

func PreviewPreparedWrite(prepared PreparedWrite) (connectors.WritePreview, error) {
	if strings.TrimSpace(prepared.Target.Connector) == "" || strings.TrimSpace(prepared.Target.Operation) == "" {
		return connectors.WritePreview{}, errors.New("engine: prepared write target is incomplete")
	}
	if err := validatePreparedRequests(prepared); err != nil {
		return connectors.WritePreview{}, err
	}
	if prepared.Target.RequiresApproval() && strings.TrimSpace(prepared.CredentialRevision) == "" {
		return connectors.WritePreview{}, fmt.Errorf("engine: destructive operation %q requires a credential revision", prepared.Target.Operation)
	}
	if prepared.Target.RequiresApproval() && strings.TrimSpace(prepared.ConfigurationDigest) == "" {
		return connectors.WritePreview{}, fmt.Errorf("engine: destructive operation %q requires a configuration digest", prepared.Target.Operation)
	}
	scope := strings.TrimSpace(prepared.ApprovalScope)
	if scope == "" {
		scope = connectors.WriteApprovalScopeProject
	}
	if prepared.Target.RequiresApproval() && scope != connectors.WriteApprovalScopeProject && scope != connectors.WriteApprovalScopeFixture {
		return connectors.WritePreview{}, fmt.Errorf("engine: destructive operation %q has an invalid approval scope", prepared.Target.Operation)
	}
	if prepared.Target.RequiresApproval() && scope == connectors.WriteApprovalScopeFixture {
		if err := validateFixturePreparedRequests(prepared.Requests, prepared.RecordsStaged); err != nil {
			return connectors.WritePreview{}, err
		}
	}
	targetDigest, err := digestPreparedTargets(prepared)
	if err != nil {
		return connectors.WritePreview{}, err
	}
	approvalTarget := connectors.WriteApprovalTarget{
		Connector:           prepared.Target.Connector,
		Operation:           prepared.Target.Operation,
		Method:              methodOrDefault(prepared.Target.Method),
		MutationClass:       prepared.Target.MutationClass,
		TargetDigest:        targetDigest,
		CredentialRevision:  prepared.CredentialRevision,
		ConfigurationDigest: prepared.ConfigurationDigest,
		Batchable:           prepared.Batchable,
		Scope:               scope,
	}
	// The preview publishes this target, so it must describe the gate the write
	// will actually meet. Keying the declaration on the same RequiresApproval
	// predicate GateDestructiveExecution uses keeps a safe write from previewing
	// a typed confirmation nothing will ask for; a destructive one still carries
	// it, which is what IssueWriteGrant demands before it mints any evidence.
	if prepared.Target.RequiresApproval() {
		approvalTarget.Confirmation = connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}
	}
	requests, err := projectPreparedRequests(prepared.Requests)
	if err != nil {
		return connectors.WritePreview{}, err
	}
	payload := struct {
		Version          int                               `json:"version"`
		ApprovalTarget   connectors.WriteApprovalTarget    `json:"approval_target"`
		RecordsStaged    int                               `json:"records_staged"`
		Action           string                            `json:"action"`
		Definition       any                               `json:"definition"`
		HookIdentity     string                            `json:"hook_identity,omitempty"`
		PreparedRequests []preparedRequestDigestProjection `json:"prepared_requests"`
	}{
		Version:          1,
		ApprovalTarget:   approvalTarget,
		RecordsStaged:    prepared.RecordsStaged,
		Action:           prepared.Action,
		Definition:       prepared.Definition,
		HookIdentity:     prepared.HookIdentity,
		PreparedRequests: requests,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return connectors.WritePreview{}, fmt.Errorf("engine: encode prepared write: %w", err)
	}
	digest := sha256.Sum256(raw)
	return connectors.WritePreview{
		RecordsStaged:  prepared.RecordsStaged,
		Action:         prepared.Action,
		Warnings:       append([]string(nil), prepared.Warnings...),
		Digest:         hex.EncodeToString(digest[:]),
		ApprovalTarget: approvalTarget,
	}, nil
}

type preparedRequestDigestProjection struct {
	Method        string   `json:"method"`
	URL           string   `json:"url"`
	Target        string   `json:"target,omitempty"`
	Query         string   `json:"query,omitempty"`
	ContentType   string   `json:"content_type,omitempty"`
	BodyFormat    string   `json:"body_format,omitempty"`
	Body          string   `json:"body,omitempty"`
	HeaderNames   []string `json:"header_names,omitempty"`
	HeadersSHA256 string   `json:"headers_sha256,omitempty"`
}

func projectPreparedRequests(requests []PreparedRequest) ([]preparedRequestDigestProjection, error) {
	projected := make([]preparedRequestDigestProjection, len(requests))
	for index, request := range requests {
		headerNames, headersSHA256, err := preparedRequestHeaderDigest(request.Headers, request.HeaderValues)
		if err != nil {
			return nil, fmt.Errorf("engine: prepared request %d headers: %w", index, err)
		}
		projected[index] = preparedRequestDigestProjection{
			Method:        request.Method,
			URL:           request.URL,
			Target:        request.Target,
			Query:         request.Query,
			ContentType:   request.ContentType,
			BodyFormat:    request.BodyFormat,
			Body:          request.Body,
			HeaderNames:   headerNames,
			HeadersSHA256: headersSHA256,
		}
	}
	return projected, nil
}

func preparedRequestHeaderDigest(headers map[string]string, headerValues map[string][]string) ([]string, string, error) {
	canonical, err := canonicalPreparedRequestHeaderValues(headers, headerValues)
	if err != nil {
		return nil, "", err
	}
	if len(canonical) == 0 {
		return nil, "", nil
	}
	names := make([]string, 0, len(canonical))
	for name := range canonical {
		names = append(names, name)
	}
	sort.Strings(names)
	payload := make([]struct {
		Name   string   `json:"name"`
		Values []string `json:"values"`
	}, len(names))
	for index, name := range names {
		payload[index].Name = name
		payload[index].Values = canonical[name]
	}
	raw, _ := json.Marshal(payload)
	digest := sha256.Sum256(raw)
	return names, hex.EncodeToString(digest[:]), nil
}

// canonicalPreparedRequestHeaders preserves the single-value helper used by
// ordinary prepared writes. Repeatable operation headers use the richer
// private digest projection below.
func canonicalPreparedRequestHeaders(headers map[string]string) (map[string]string, error) {
	values, err := canonicalPreparedRequestHeaderValues(headers, nil)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}
	canonical := make(map[string]string, len(values))
	for name, headerValues := range values {
		if len(headerValues) != 1 {
			return nil, fmt.Errorf("header %q has multiple values", name)
		}
		canonical[name] = headerValues[0]
	}
	return canonical, nil
}

func canonicalPreparedRequestHeaderValues(headers map[string]string, headerValues map[string][]string) (map[string][]string, error) {
	if len(headers) == 0 && len(headerValues) == 0 {
		return nil, nil
	}
	names := make([]string, 0, len(headers)+len(headerValues))
	for name := range headers {
		names = append(names, name)
	}
	for name := range headerValues {
		names = append(names, name)
	}
	sort.Strings(names)
	canonical := make(map[string][]string, len(names))
	for _, name := range names {
		canonicalName, err := canonicalPreparedRequestHeaderName(name)
		if err != nil {
			return nil, err
		}
		if _, exists := canonical[canonicalName]; exists {
			return nil, fmt.Errorf("duplicate header name %q", canonicalName)
		}
		values, repeated := headerValues[name]
		if !repeated {
			values = []string{headers[name]}
		}
		if len(values) == 0 {
			return nil, fmt.Errorf("header %q has no values", canonicalName)
		}
		for _, value := range values {
			if strings.ContainsAny(value, "\r\n") {
				return nil, fmt.Errorf("header %q contains CR/LF", canonicalName)
			}
			if err := safety.RejectDangerousChars(value, "header "+canonicalName); err != nil {
				return nil, err
			}
		}
		canonical[canonicalName] = append([]string(nil), values...)
	}
	return canonical, nil
}

func canonicalPreparedRequestHeaderName(name string) (string, error) {
	if name == "" || name != strings.TrimSpace(name) {
		return "", fmt.Errorf("invalid header name")
	}
	for index := 0; index < len(name); index++ {
		if !isPreparedRequestHeaderToken(name[index]) {
			return "", fmt.Errorf("invalid header name %q", name)
		}
	}
	canonical := textproto.CanonicalMIMEHeaderKey(name)
	if canonical == "" {
		return "", fmt.Errorf("invalid header name %q", name)
	}
	return canonical, nil
}

func isPreparedRequestHeaderToken(value byte) bool {
	return value >= 'a' && value <= 'z' ||
		value >= 'A' && value <= 'Z' ||
		value >= '0' && value <= '9' ||
		strings.ContainsRune("!#$%&'*+-.^_`|~", rune(value))
}

func validateFixturePreparedRequests(requests []PreparedRequest, recordsStaged int) error {
	if len(requests) == 0 {
		if recordsStaged == 0 {
			return nil
		}
		return errors.New("engine: fixture write approval requires a prepared loopback request")
	}
	for index, request := range requests {
		parsed, err := url.Parse(request.URL)
		if err != nil {
			return fmt.Errorf("engine: fixture prepared request %d has an invalid URL", index)
		}
		host := strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
		ip := net.ParseIP(host)
		if host != "localhost" && (ip == nil || !ip.IsLoopback()) {
			return fmt.Errorf("engine: fixture prepared request %d must target loopback", index)
		}
	}
	return nil
}

func ExecutePreparedWrite(ctx context.Context, prepared PreparedWrite, evidence *connectors.WriteApprovalEvidence, previewDigest string, execute func(context.Context) error) error {
	preview, err := PreviewPreparedWrite(prepared)
	if err != nil {
		return err
	}
	if strings.TrimSpace(previewDigest) == "" || preview.Digest != previewDigest {
		return fmt.Errorf("engine: operation %q no longer matches its prepared preview", prepared.Target.Operation)
	}
	if prepared.RecordsStaged == 0 && len(prepared.Requests) == 0 {
		execute = func(context.Context) error { return nil }
	}
	return GateDestructiveExecution(ctx, prepared.Target, evidence, preview.Digest, preview.ApprovalTarget, execute)
}

func digestPreparedTargets(prepared PreparedWrite) (string, error) {
	targets := make([]struct {
		Method        string   `json:"method"`
		URL           string   `json:"url"`
		Target        string   `json:"target,omitempty"`
		Query         string   `json:"query,omitempty"`
		BodyFormat    string   `json:"body_format,omitempty"`
		Body          string   `json:"body,omitempty"`
		HeaderNames   []string `json:"header_names,omitempty"`
		HeadersSHA256 string   `json:"headers_sha256,omitempty"`
	}, len(prepared.Requests))
	for i, request := range prepared.Requests {
		headerNames, headersSHA256, err := preparedRequestHeaderDigest(request.Headers, request.HeaderValues)
		if err != nil {
			return "", fmt.Errorf("engine: prepared request %d headers: %w", i, err)
		}
		targets[i].Method = request.Method
		targets[i].URL = request.URL
		targets[i].Target = request.Target
		targets[i].Query = request.Query
		targets[i].BodyFormat = request.BodyFormat
		targets[i].Body = request.Body
		targets[i].HeaderNames = headerNames
		targets[i].HeadersSHA256 = headersSHA256
	}
	if len(targets) == 0 {
		targets = append(targets, struct {
			Method        string   `json:"method"`
			URL           string   `json:"url"`
			Target        string   `json:"target,omitempty"`
			Query         string   `json:"query,omitempty"`
			BodyFormat    string   `json:"body_format,omitempty"`
			Body          string   `json:"body,omitempty"`
			HeaderNames   []string `json:"header_names,omitempty"`
			HeadersSHA256 string   `json:"headers_sha256,omitempty"`
		}{Method: prepared.Target.Method, Target: prepared.Target.Connector + "/" + prepared.Target.Operation})
	}
	raw, err := json.Marshal(targets)
	if err != nil {
		return "", fmt.Errorf("engine: encode prepared write targets: %w", err)
	}
	digest := sha256.Sum256(raw)
	return hex.EncodeToString(digest[:]), nil
}

func validatePreparedRequests(prepared PreparedWrite) error {
	targetMethod := methodOrDefault(prepared.Target.Method)
	for index, request := range prepared.Requests {
		if strings.TrimSpace(request.Method) == "" || strings.TrimSpace(request.URL) == "" {
			return fmt.Errorf("engine: prepared request %d is incomplete", index)
		}
		if !strings.EqualFold(request.Method, targetMethod) {
			return fmt.Errorf("engine: prepared request %d method %q does not match target method %q", index, request.Method, targetMethod)
		}
		if _, err := canonicalPreparedRequestHeaderValues(request.Headers, request.HeaderValues); err != nil {
			return fmt.Errorf("engine: prepared request %d headers: %w", index, err)
		}
	}
	return nil
}
