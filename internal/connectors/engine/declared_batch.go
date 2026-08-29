package engine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

type declaredBatchInputAction struct {
	name   string
	record connectors.Record
}

func buildDeclaredBatchPayload(b Bundle, action WriteAction, record connectors.Record, cfg connectors.RuntimeConfig) (map[string]any, int, error) {
	spec := action.DeclaredBatch
	if spec == nil {
		return nil, 0, fmt.Errorf("engine: write action %q: declared_batch spec is required", action.Name)
	}
	inputs, err := parseDeclaredBatchInput(action, record)
	if err != nil {
		return nil, 0, err
	}
	outerBase, err := resolveWriteActionRoute(b, cfg, action)
	if err != nil {
		return nil, 0, fmt.Errorf("engine: write action %q: resolve batch route: %w", action.Name, err)
	}
	providerActions := make([]any, 0, len(inputs))
	for index, input := range inputs {
		inner, err := findWriteAction(b, input.name)
		if err != nil {
			return nil, 0, fmt.Errorf("engine: write action %q: batch action %d: %w", action.Name, index, err)
		}
		if err := validateWriteActionRecord(inner, input.record, nil); err != nil {
			return nil, 0, fmt.Errorf("engine: write action %q: batch action %d (%q): %w", action.Name, index, input.name, err)
		}
		prepared, err := prepareDeclarativeRequest(b, inner, input.record, index, cfg, false)
		if err != nil {
			return nil, 0, fmt.Errorf("engine: write action %q: batch action %d (%q): %w", action.Name, index, input.name, err)
		}
		if prepared.Query != "" {
			return nil, 0, fmt.Errorf("engine: write action %q: batch action %d (%q) resolves query parameters, which the declared batch subset does not admit", action.Name, index, input.name)
		}
		relativePath, err := declaredBatchRelativePath(outerBase, prepared.URL)
		if err != nil {
			return nil, 0, fmt.Errorf("engine: write action %q: batch action %d (%q): %w", action.Name, index, input.name, err)
		}
		providerAction := map[string]any{
			spec.ProviderMethodField: strings.ToLower(prepared.Method),
			spec.ProviderPathField:   relativePath,
		}
		data, present, err := declaredBatchPreparedData(inner, prepared, spec.InnerBodyField)
		if err != nil {
			return nil, 0, fmt.Errorf("engine: write action %q: batch action %d (%q): %w", action.Name, index, input.name, err)
		}
		if present {
			providerAction[spec.ProviderDataField] = data
		}
		providerActions = append(providerActions, providerAction)
	}
	return map[string]any{
		spec.ProviderEnvelopeField: map[string]any{
			spec.ProviderActionsField: providerActions,
		},
	}, len(inputs), nil
}

func parseDeclaredBatchInput(action WriteAction, record connectors.Record) ([]declaredBatchInputAction, error) {
	spec := action.DeclaredBatch
	raw, ok := record["actions"]
	if !ok {
		return nil, fmt.Errorf("engine: write action %q: declared batch requires record.actions", action.Name)
	}
	actions, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("engine: write action %q: record.actions must be an array", action.Name)
	}
	if len(actions) < 1 || len(actions) > spec.MaxActions {
		return nil, fmt.Errorf("engine: write action %q: record.actions must contain between 1 and %d actions", action.Name, spec.MaxActions)
	}
	allowed := make(map[string]struct{}, len(spec.AllowedActions))
	for _, name := range spec.AllowedActions {
		allowed[name] = struct{}{}
	}
	result := make([]declaredBatchInputAction, 0, len(actions))
	for index, rawAction := range actions {
		item, ok := rawAction.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("engine: write action %q: batch action %d must be an object", action.Name, index)
		}
		if len(item) != 2 {
			return nil, fmt.Errorf("engine: write action %q: batch action %d may contain only action and record", action.Name, index)
		}
		name, ok := item["action"].(string)
		name = strings.TrimSpace(name)
		if !ok || name == "" {
			return nil, fmt.Errorf("engine: write action %q: batch action %d requires a named action", action.Name, index)
		}
		if _, ok := allowed[name]; !ok {
			return nil, fmt.Errorf("engine: write action %q: batch action %d selects undeclared action %q", action.Name, index, name)
		}
		rawRecord, ok := item["record"]
		if !ok {
			return nil, fmt.Errorf("engine: write action %q: batch action %d requires record", action.Name, index)
		}
		var innerRecord connectors.Record
		switch value := rawRecord.(type) {
		case map[string]any:
			innerRecord = connectors.Record(value)
		case connectors.Record:
			innerRecord = value
		default:
			return nil, fmt.Errorf("engine: write action %q: batch action %d record must be an object", action.Name, index)
		}
		result = append(result, declaredBatchInputAction{name: name, record: innerRecord})
	}
	return result, nil
}

func declaredBatchPreparedData(action WriteAction, prepared PreparedRequest, innerBodyField string) (any, bool, error) {
	switch prepared.BodyFormat {
	case "none":
		return nil, false, nil
	case "json":
	default:
		return nil, false, fmt.Errorf("resolved body_type %q is not supported inside a declared batch", bodyTypeOf(action))
	}
	decoder := json.NewDecoder(strings.NewReader(prepared.Body))
	decoder.UseNumber()
	var body map[string]any
	if err := decoder.Decode(&body); err != nil {
		return nil, false, fmt.Errorf("decode resolved JSON body: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, false, errors.New("resolved JSON body contains multiple values")
		}
		return nil, false, fmt.Errorf("decode resolved JSON body: %w", err)
	}
	if len(body) != 1 {
		return nil, false, fmt.Errorf("resolved JSON body must contain only the declared %q envelope", innerBodyField)
	}
	data, ok := body[innerBodyField]
	if !ok {
		return nil, false, fmt.Errorf("resolved JSON body is missing declared %q envelope", innerBodyField)
	}
	return data, true, nil
}

func declaredBatchRelativePath(baseURL, requestURL string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse batch base URL: %w", err)
	}
	request, err := url.Parse(requestURL)
	if err != nil {
		return "", fmt.Errorf("parse batch subrequest URL: %w", err)
	}
	if !strings.EqualFold(base.Scheme, request.Scheme) || !strings.EqualFold(base.Host, request.Host) || request.User != nil || request.RawQuery != "" || request.Fragment != "" {
		return "", errors.New("batch subrequest must stay on the declared provider origin without query or fragment")
	}
	basePath := strings.TrimRight(base.EscapedPath(), "/")
	requestPath := request.EscapedPath()
	if basePath != "" {
		if requestPath != basePath && !strings.HasPrefix(requestPath, basePath+"/") {
			return "", errors.New("batch subrequest path escapes the declared provider base path")
		}
		requestPath = strings.TrimPrefix(requestPath, basePath)
	}
	if requestPath == "" {
		requestPath = "/"
	}
	if !strings.HasPrefix(requestPath, "/") || strings.Contains(requestPath, "?") || strings.Contains(requestPath, "#") {
		return "", errors.New("batch subrequest did not resolve to a relative provider path")
	}
	return requestPath, nil
}

func validateDeclaredBatchResponse(action WriteAction, record connectors.Record, response *connsdk.Response) error {
	if action.DeclaredBatch == nil {
		return nil
	}
	if response == nil {
		return errors.New("declared batch provider response is absent")
	}
	if !writeProviderResponseDeclaresJSON(response.Header) {
		return errors.New("declared batch provider response must declare JSON")
	}
	decoder := json.NewDecoder(bytes.NewReader(response.Body))
	decoder.UseNumber()
	var body map[string]any
	if err := decoder.Decode(&body); err != nil {
		return fmt.Errorf("decode declared batch provider response: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return errors.New("declared batch provider response contains multiple JSON values")
		}
		return fmt.Errorf("decode declared batch provider response: %w", err)
	}
	rawResults, ok := body[action.DeclaredBatch.ResponseEnvelopeField]
	if !ok {
		return fmt.Errorf("declared batch provider response is missing response field %q", action.DeclaredBatch.ResponseEnvelopeField)
	}
	results, ok := rawResults.([]any)
	if !ok {
		return fmt.Errorf("declared batch provider response field %q must be an array", action.DeclaredBatch.ResponseEnvelopeField)
	}
	inputs, err := parseDeclaredBatchInput(action, record)
	if err != nil {
		return err
	}
	if len(results) != len(inputs) {
		return fmt.Errorf("declared batch provider returned %d subrequest results, want %d subrequest results", len(results), len(inputs))
	}
	for index, rawResult := range results {
		item, ok := rawResult.(map[string]any)
		if !ok {
			return fmt.Errorf("declared batch subrequest result %d must be an object", index)
		}
		rawStatus, ok := item[action.DeclaredBatch.ResponseStatusField]
		if !ok {
			return fmt.Errorf("declared batch subrequest result %d is missing %q", index, action.DeclaredBatch.ResponseStatusField)
		}
		status, err := declaredBatchStatusCode(rawStatus)
		if err != nil {
			return fmt.Errorf("declared batch subrequest result %d %q: %w", index, action.DeclaredBatch.ResponseStatusField, err)
		}
		if status < http.StatusOK || status >= http.StatusMultipleChoices {
			return fmt.Errorf("declared batch subrequest %d returned status %d", index, status)
		}
	}
	return nil
}

func declaredBatchStatusCode(raw any) (int, error) {
	value, ok := raw.(json.Number)
	if !ok {
		return 0, errors.New("must be an integer")
	}
	parsed, err := value.Int64()
	if err != nil || parsed < 100 || parsed > 599 {
		return 0, errors.New("must be an HTTP status integer")
	}
	return int(parsed), nil
}
