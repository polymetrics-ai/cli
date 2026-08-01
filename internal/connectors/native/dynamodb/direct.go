package dynamodb

import (
	"context"
	"fmt"
	"strings"

	"polymetrics.ai/internal/connectors"
)

func (c Connector) OperationDirectRead(ctx context.Context, req connectors.OperationDirectReadRequest) (connectors.DirectReadResult, error) {
	if err := ctx.Err(); err != nil {
		return connectors.DirectReadResult{}, err
	}
	body, target, err := directReadBody(req)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	if fixtureMode(req.Config) {
		fixture := map[string]any{"fixture": true, "target": target, "request": redactDirectRequest(body)}
		redacted, err := applyDynamoDBDirectReadPolicies(req, fixture)
		if err != nil {
			return connectors.DirectReadResult{}, err
		}
		return connectors.DirectReadResult{Connector: "dynamodb", Method: "POST", Path: target, Status: 200, Body: redacted}, nil
	}
	conn, err := resolveConfig(req.Config)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	var out map[string]any
	if err := c.doJSONLimited(ctx, conn, target, body, &out, directReadMaxBytes(req.MaxBytes)); err != nil {
		return connectors.DirectReadResult{}, err
	}
	redacted, err := applyDynamoDBDirectReadPolicies(req, out)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	return connectors.DirectReadResult{Connector: c.Name(), Method: "POST", Path: target, Status: 200, Body: redacted}, nil
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
		if bodyBool(req.Body, "consistent_read") {
			return nil, "", fmt.Errorf("dynamodb transact_get_items does not support consistent_read")
		}
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
	if bodyString(body, "table_name") == "" {
		return nil, fmt.Errorf("dynamodb direct read requires table_name")
	}
	if usesCompositeKeyFields(body) {
		return directReadCompositeKey(body)
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
	if bodyString(body, "table_name") == "" {
		return nil, fmt.Errorf("dynamodb direct read requires table_name")
	}
	if usesCompositeKeyFields(body) {
		return directReadCompositeKeys(body)
	}
	name := bodyString(body, "key_name")
	if name == "" {
		return nil, fmt.Errorf("dynamodb direct read requires key_name")
	}
	values := bodyStringSlice(body, "key_values")
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

func usesCompositeKeyFields(body map[string]any) bool {
	for _, key := range []string{"partition_key_name", "partition_key_type", "partition_key_value", "partition_key_values", "sort_key_name", "sort_key_type", "sort_key_value", "sort_key_values"} {
		if bodyString(body, key) != "" || len(bodyStringSlice(body, key)) > 0 {
			return true
		}
	}
	return false
}

func directReadCompositeKey(body map[string]any) (map[string]any, error) {
	partitionName := bodyString(body, "partition_key_name")
	partitionValue := bodyString(body, "partition_key_value")
	if partitionName == "" || partitionValue == "" {
		return nil, fmt.Errorf("dynamodb direct read requires partition_key_name and partition_key_value")
	}
	key := map[string]any{}
	av, err := scalarAttributeValue(bodyString(body, "partition_key_type"), partitionValue)
	if err != nil {
		return nil, err
	}
	key[partitionName] = av
	sortName := bodyString(body, "sort_key_name")
	sortValue := bodyString(body, "sort_key_value")
	if sortName != "" || sortValue != "" || bodyString(body, "sort_key_type") != "" {
		if sortName == "" || sortValue == "" {
			return nil, fmt.Errorf("dynamodb direct read requires sort_key_name and sort_key_value together")
		}
		av, err := scalarAttributeValue(bodyString(body, "sort_key_type"), sortValue)
		if err != nil {
			return nil, err
		}
		key[sortName] = av
	}
	return key, nil
}

func directReadCompositeKeys(body map[string]any) ([]any, error) {
	partitionName := bodyString(body, "partition_key_name")
	partitionValues := bodyStringSlice(body, "partition_key_values")
	if len(partitionValues) == 0 {
		if value := bodyString(body, "partition_key_value"); value != "" {
			partitionValues = []string{value}
		}
	}
	if partitionName == "" || len(partitionValues) == 0 {
		return nil, fmt.Errorf("dynamodb direct read requires partition_key_name and partition_key_values")
	}
	sortName := bodyString(body, "sort_key_name")
	sortValues := bodyStringSlice(body, "sort_key_values")
	if len(sortValues) == 0 {
		if value := bodyString(body, "sort_key_value"); value != "" {
			sortValues = []string{value}
		}
	}
	useSortKey := sortName != "" || len(sortValues) > 0 || bodyString(body, "sort_key_type") != ""
	if useSortKey {
		if sortName == "" || len(sortValues) == 0 {
			return nil, fmt.Errorf("dynamodb direct read requires sort_key_name and sort_key_values together")
		}
		if len(sortValues) != len(partitionValues) {
			return nil, fmt.Errorf("dynamodb direct read requires equal partition_key_values and sort_key_values counts")
		}
	}
	keys := make([]any, 0, len(partitionValues))
	for i, partitionValue := range partitionValues {
		partitionAV, err := scalarAttributeValue(bodyString(body, "partition_key_type"), partitionValue)
		if err != nil {
			return nil, err
		}
		key := map[string]any{partitionName: partitionAV}
		if useSortKey {
			sortAV, err := scalarAttributeValue(bodyString(body, "sort_key_type"), sortValues[i])
			if err != nil {
				return nil, err
			}
			key[sortName] = sortAV
		}
		keys = append(keys, key)
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

func bodyStringSlice(body map[string]any, key string) []string {
	if body == nil {
		return nil
	}
	switch typed := body[key].(type) {
	case []string:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if item = strings.TrimSpace(item); item != "" {
				out = append(out, item)
			}
		}
		return out
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if value := strings.TrimSpace(fmt.Sprint(item)); value != "" {
				out = append(out, value)
			}
		}
		return out
	case string:
		out := []string{}
		for _, item := range strings.Split(typed, ",") {
			if item = strings.TrimSpace(item); item != "" {
				out = append(out, item)
			}
		}
		return out
	default:
		return nil
	}
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

func applyDynamoDBDirectReadPolicies(req connectors.OperationDirectReadRequest, body any) (any, error) {
	out := body
	switch req.OutputPolicy {
	case "":
	case "json_redacted", "clinical_json_redacted":
		out = redactDynamoDBJSONValue(out)
	default:
		return nil, fmt.Errorf("direct read output policy %q is not supported", req.OutputPolicy)
	}
	if len(req.RedactFields) > 0 {
		out = redactDynamoDBNamedJSONFields(out, req.RedactFields)
	}
	return out, nil
}

func redactDynamoDBJSONValue(value any) any {
	return redactDynamoDBJSONFieldsByPredicate(value, shouldRedactDynamoDBJSONField)
}

func redactDynamoDBNamedJSONFields(value any, fields []string) any {
	fieldSet := make(map[string]bool, len(fields))
	for _, field := range fields {
		fieldSet[normalizeDynamoDBJSONFieldName(field)] = true
	}
	return redactDynamoDBJSONFieldsByPredicate(value, func(name string) bool {
		return fieldSet[normalizeDynamoDBJSONFieldName(name)]
	})
}

func redactDynamoDBJSONFieldsByPredicate(value any, shouldRedact func(string) bool) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, item := range v {
			if item != nil && shouldRedact(key) {
				out[key+"_redacted"] = true
				continue
			}
			out[key] = redactDynamoDBJSONFieldsByPredicate(item, shouldRedact)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = redactDynamoDBJSONFieldsByPredicate(item, shouldRedact)
		}
		return out
	default:
		return value
	}
}

func shouldRedactDynamoDBJSONField(name string) bool {
	normalized := normalizeDynamoDBJSONFieldName(name)
	switch normalized {
	case "content", "body", "payload", "raw", "download_url", "download_media_url", "clone_url", "api_key", "apikey", "access_key", "private_key", "authorization", "credential", "credentials":
		return true
	}
	if strings.Contains(normalized, "download") && strings.Contains(normalized, "url") {
		return true
	}
	if strings.Contains(normalized, "clone") && strings.Contains(normalized, "url") {
		return true
	}
	for _, marker := range []string{"token", "secret", "password", "private_key", "api_key", "apikey", "access_key", "authorization", "credential"} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func normalizeDynamoDBJSONFieldName(name string) string {
	return strings.ToLower(strings.NewReplacer("-", "_", " ", "_", ".", "_").Replace(name))
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
