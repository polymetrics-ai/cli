package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

// ValidateWrite validates DynamoDB write records against defs/dynamodb/writes.json.
// The schemas are closed at the top level and model typed DynamoDB operation
// inputs; no raw HTTP body or PartiQL statement passthrough is accepted.
func (c Connector) ValidateWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) error {
	return engine.ValidateWrite(ctx, c.definitionBundle(), req, records)
}

// DryRunWrite validates and previews the signed DynamoDB target. It performs no
// network call; execution remains under the app's plan -> preview -> approval ->
// execute reverse-ETL gate.
func (c Connector) DryRunWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WritePreview, error) {
	if err := c.ValidateWrite(ctx, req, records); err != nil {
		return connectors.WritePreview{}, err
	}
	op, ok := writeOperationByAction[req.Action]
	if !ok {
		return connectors.WritePreview{}, fmt.Errorf("dynamodb write action %q not found", req.Action)
	}
	warnings := []string{fmt.Sprintf("%s executes %s only after reverse-ETL approval; dry run performs no external call", req.Action, op.Target)}
	if len(records) > 0 {
		body, err := buildWriteBody(req.Action, records[0])
		if err != nil {
			return connectors.WritePreview{}, err
		}
		previewBody := redactPreviewBody(body)
		encoded, _ := json.Marshal(previewBody)
		warnings = append(warnings, fmt.Sprintf("resolved request: POST %s X-Amz-Target=%s body=%s", "/", op.Target, encoded))
	}
	return connectors.WritePreview{RecordsStaged: len(records), Action: req.Action, Warnings: warnings}, nil
}

// Write executes a typed DynamoDB write action against the configured endpoint.
// It is fail-fast like the engine write path: the first validation, context, or
// provider error stops the batch and returns completed counts.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	if err := c.ValidateWrite(ctx, req, records); err != nil {
		return connectors.WriteResult{RecordsFailed: len(records)}, err
	}
	if fixtureMode(req.Config) {
		return connectors.WriteResult{RecordsWritten: len(records)}, nil
	}
	op, ok := writeOperationByAction[req.Action]
	if !ok {
		return connectors.WriteResult{RecordsFailed: len(records)}, fmt.Errorf("dynamodb write action %q not found", req.Action)
	}
	conn, err := resolveConfig(req.Config)
	if err != nil {
		return connectors.WriteResult{RecordsFailed: len(records)}, err
	}
	var result connectors.WriteResult
	for i, rec := range records {
		if err := ctx.Err(); err != nil {
			result.RecordsFailed = len(records) - i
			return result, err
		}
		body, err := buildWriteBody(req.Action, rec)
		if err != nil {
			result.RecordsFailed = len(records) - i
			return result, err
		}
		var response map[string]any
		if err := c.doJSON(ctx, conn, op.Target, body, &response); err != nil {
			result.RecordsFailed = len(records) - i
			return result, redactWriteError(err, rec)
		}
		result.RecordsWritten++
	}
	return result, nil
}

func redactPreviewBody(body map[string]any) map[string]any {
	out := make(map[string]any, len(body))
	for k, v := range body {
		if sensitiveWriteMember(k) {
			out[k] = "***"
			continue
		}
		out[k] = v
	}
	return out
}

func redactWriteError(err error, rec connectors.Record) error {
	if err == nil {
		return nil
	}
	text := err.Error()
	for _, value := range rec {
		redactValueLiterals(&text, value)
	}
	return fmt.Errorf("%s", text)
}

func redactValueLiterals(text *string, value any) {
	switch typed := value.(type) {
	case string:
		if strings.TrimSpace(typed) != "" {
			*text = strings.ReplaceAll(*text, typed, "***")
		}
	case []string:
		for _, item := range typed {
			redactValueLiterals(text, item)
		}
	case []any:
		for _, item := range typed {
			redactValueLiterals(text, item)
		}
	case map[string]any:
		for _, item := range typed {
			redactValueLiterals(text, item)
		}
	case connectors.Record:
		for _, item := range typed {
			redactValueLiterals(text, item)
		}
	}
}

func sensitiveWriteMember(name string) bool {
	n := strings.ToLower(name)
	for _, marker := range []string{"key", "item", "requestitems", "transactitems", "attributeupdates", "policy"} {
		if strings.Contains(n, strings.ToLower(marker)) {
			return true
		}
	}
	return false
}
