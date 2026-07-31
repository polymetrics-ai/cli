package dynamodb

import (
	"context"
	"fmt"
	"strings"

	"polymetrics.ai/internal/connectors"
)

// OperationDirectRead executes the three bounded/keyed direct read helpers
// declared in cli_surface.json. Each helper builds a closed DynamoDB JSON-RPC
// body from typed command fields; no raw PartiQL, expression string, endpoint,
// or body passthrough is accepted.
func (c Connector) OperationDirectRead(ctx context.Context, req connectors.OperationDirectReadRequest) (connectors.DirectReadResult, error) {
	if err := ctx.Err(); err != nil {
		return connectors.DirectReadResult{}, err
	}
	body, target, err := directReadBody(req)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	if fixtureMode(req.Config) {
		return connectors.DirectReadResult{Connector: "dynamodb", Method: "POST", Path: target, Status: 200, Body: map[string]any{"fixture": true, "target": target, "request": redactDirectRequest(body)}}, nil
	}
	conn, err := resolveConfig(req.Config)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	var out map[string]any
	if err := c.doJSONLimited(ctx, conn, target, body, &out, directReadMaxBytes(req.MaxBytes)); err != nil {
		return connectors.DirectReadResult{}, err
	}
	return connectors.DirectReadResult{Connector: c.Name(), Method: "POST", Path: target, Status: 200, Body: out}, nil
}

func directReadBody(req connectors.OperationDirectReadRequest) (map[string]any, string, error) {
	switch req.Operation {
	case "get_item":
		key, err := directReadKey(req.Body)
		if err != nil {
			return nil, "", err
		}
		return map[string]any{"TableName": bodyString(req.Body, "table_name"), "Key": key, "ConsistentRead": bodyBool(req.Body, "consistent_read")}, dynamoTargetPrefix + "GetItem", nil
	case "batch_get_item":
		keys, err := directReadKeys(req.Body)
		if err != nil {
			return nil, "", err
		}
		table := bodyString(req.Body, "table_name")
		return map[string]any{"RequestItems": map[string]any{table: map[string]any{"Keys": keys, "ConsistentRead": bodyBool(req.Body, "consistent_read")}}}, dynamoTargetPrefix + "BatchGetItem", nil
	case "transact_get_items":
		keys, err := directReadKeys(req.Body)
		if err != nil {
			return nil, "", err
		}
		table := bodyString(req.Body, "table_name")
		items := make([]any, 0, len(keys))
		for _, key := range keys {
			items = append(items, map[string]any{"Get": map[string]any{"TableName": table, "Key": key}})
		}
		return map[string]any{"TransactItems": items}, dynamoTargetPrefix + "TransactGetItems", nil
	default:
		return nil, "", fmt.Errorf("dynamodb operation direct read %q not found", req.Operation)
	}
}

func directReadKey(body map[string]any) (map[string]any, error) {
	table := bodyString(body, "table_name")
	if table == "" {
		return nil, fmt.Errorf("dynamodb direct read requires table_name")
	}
	name := bodyString(body, "key_name")
	value := bodyString(body, "key_value")
	if name == "" || value == "" {
		return nil, fmt.Errorf("dynamodb direct read requires key_name and key_value")
	}
	av, err := scalarAttributeValue(bodyString(body, "key_type"), value)
	if err != nil {
		return nil, err
	}
	return map[string]any{name: av}, nil
}

func directReadKeys(body map[string]any) ([]any, error) {
	table := bodyString(body, "table_name")
	if table == "" {
		return nil, fmt.Errorf("dynamodb direct read requires table_name")
	}
	name := bodyString(body, "key_name")
	if name == "" {
		return nil, fmt.Errorf("dynamodb direct read requires key_name")
	}
	var values []string
	switch typed := body["key_values"].(type) {
	case []string:
		values = typed
	case []any:
		for _, item := range typed {
			values = append(values, fmt.Sprint(item))
		}
	case string:
		for _, item := range strings.Split(typed, ",") {
			if item = strings.TrimSpace(item); item != "" {
				values = append(values, item)
			}
		}
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("dynamodb direct read requires key_values")
	}
	keys := make([]any, 0, len(values))
	for _, value := range values {
		av, err := scalarAttributeValue(bodyString(body, "key_type"), value)
		if err != nil {
			return nil, err
		}
		keys = append(keys, map[string]any{name: av})
	}
	return keys, nil
}

func bodyString(body map[string]any, key string) string {
	if body == nil {
		return ""
	}
	value, ok := body[key]
	if !ok || value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func directReadMaxBytes(requested int) int {
	const maxDynamoDBDirectReadBytes = 16 << 20
	if requested <= 0 || requested > maxDynamoDBDirectReadBytes {
		return maxDynamoDBDirectReadBytes
	}
	return requested
}

func bodyBool(body map[string]any, key string) bool {
	v, ok := body[key]
	if !ok {
		return false
	}
	switch typed := v.(type) {
	case bool:
		return typed
	case string:
		return strings.EqualFold(strings.TrimSpace(typed), "true")
	default:
		return false
	}
}

func redactDirectRequest(body map[string]any) map[string]any {
	out := make(map[string]any, len(body))
	for k, v := range body {
		if k == "Key" || k == "RequestItems" || k == "TransactItems" {
			out[k] = "***"
			continue
		}
		out[k] = v
	}
	return out
}
