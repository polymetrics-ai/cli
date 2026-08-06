package awscloudtrail

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

func emitActionRecords(action string, decoded map[string]any, requestBody map[string]any, emit func(connectors.Record) error) error {
	records := recordsForAction(action, decoded)
	for _, record := range records {
		stampRecord(action, record, requestBody)
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

func recordsForAction(action string, decoded map[string]any) []connectors.Record {
	if isCollectionAction(action) {
		if arr, ok := firstArray(decoded); ok {
			out := make([]connectors.Record, 0, len(arr))
			for _, elem := range arr {
				out = append(out, recordFromValue(elem))
			}
			return out
		}
	}
	return []connectors.Record{recordFromValue(decoded)}
}

func isCollectionAction(action string) bool {
	return action == "DescribeTrails" || strings.HasPrefix(action, "List")
}

func firstArray(decoded map[string]any) ([]any, bool) {
	preferred := []string{"trailList", "Trails", "Events", "Channels", "Dashboards", "EventDataStores", "Failures", "Imports", "PublicKeys", "PublicKeyList", "Queries", "TagsList", "ResourceTagList"}
	for _, key := range preferred {
		if arr, ok := decoded[key].([]any); ok {
			return arr, true
		}
	}
	for _, value := range decoded {
		if arr, ok := value.([]any); ok {
			return arr, true
		}
	}
	return nil, false
}

func recordFromValue(value any) connectors.Record {
	if obj, ok := value.(map[string]any); ok {
		return connectors.Record(obj)
	}
	return connectors.Record{"value": value}
}

func stampRecord(action string, record connectors.Record, requestBody map[string]any) {
	if _, ok := record["operation"]; !ok {
		record["operation"] = action
	}
	if _, ok := record["pm_record_id"]; ok {
		return
	}
	if value, ok := firstIdentityValue(map[string]any(record), recordIdentityKeys()); ok {
		record["pm_record_id"] = value
		return
	}
	if value, ok := requestScopedRecordID(action, record, requestBody); ok {
		record["pm_record_id"] = value
		return
	}
	record["pm_record_id"] = hashedRecordID(action, record, nil)
}

type identityPart struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

func recordIdentityKeys() []string {
	return []string{"Name", "TrailARN", "TrailArn", "TrailName", "ChannelArn", "Channel", "DashboardId", "DashboardArn", "EventDataStoreArn", "EventDataStore", "Arn", "EventId", "QueryId", "ImportId", "ResourceId", "ResourceARN", "ResourceArn"}
}

func requestIdentityKeys() []string {
	return []string{"Name", "TrailName", "Channel", "DashboardId", "EventDataStore", "ImportId", "ResourceArn", "ResourceId", "ResourceIdList", "QueryId"}
}

func requestScopedRecordKeys() []string {
	return []string{"Location"}
}

func firstIdentityValue(obj map[string]any, keys []string) (string, bool) {
	for _, key := range keys {
		if value, ok := identityString(obj, key); ok {
			return value, true
		}
	}
	return "", false
}

func requestScopedRecordID(action string, record connectors.Record, requestBody map[string]any) (string, bool) {
	parts := requestIdentityParts(requestBody)
	if len(parts) == 0 {
		return "", false
	}
	if !isCollectionAction(action) {
		return identityPartsString(parts), true
	}
	if value, ok := firstIdentityValue(map[string]any(record), requestScopedRecordKeys()); ok {
		parts = append(append([]identityPart(nil), parts...), identityPart{Field: "Location", Value: value})
		return identityPartsString(parts), true
	}
	return hashedRecordID(action, record, parts), true
}

func requestIdentityParts(requestBody map[string]any) []identityPart {
	if len(requestBody) == 0 {
		return nil
	}
	var parts []identityPart
	for _, key := range requestIdentityKeys() {
		if value, ok := identityString(requestBody, key); ok {
			parts = append(parts, identityPart{Field: key, Value: value})
		}
	}
	return parts
}

func identityPartsString(parts []identityPart) string {
	if len(parts) == 1 {
		return parts[0].Value
	}
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		values = append(values, part.Field+"="+part.Value)
	}
	return strings.Join(values, "|")
}

func identityString(obj map[string]any, key string) (string, bool) {
	value, ok := obj[key]
	if !ok || value == nil {
		return "", false
	}
	switch value.(type) {
	case string, json.Number:
		stringValue, ok := stringAt(obj, key)
		if !ok {
			return "", false
		}
		stringValue = strings.TrimSpace(stringValue)
		return stringValue, stringValue != ""
	default:
		encoded, err := json.Marshal(value)
		if err != nil {
			stringValue := strings.TrimSpace(fmt.Sprint(value))
			return stringValue, stringValue != ""
		}
		stringValue := strings.TrimSpace(string(encoded))
		return stringValue, stringValue != "" && stringValue != "null"
	}
}

func hashedRecordID(action string, record connectors.Record, requestParts []identityPart) string {
	payload := map[string]any{"record": map[string]any(record)}
	if len(requestParts) > 0 {
		payload["request"] = requestParts
	}
	encoded, _ := json.Marshal(payload)
	h := fnv.New64a()
	_, _ = h.Write(encoded)
	return fmt.Sprintf("%s:%x", action, h.Sum64())
}

func stringAt(obj map[string]any, key string) (string, bool) {
	value, ok := obj[key]
	if !ok || value == nil {
		return "", false
	}
	switch v := value.(type) {
	case string:
		return v, true
	case json.Number:
		return v.String(), true
	default:
		return fmt.Sprint(v), true
	}
}

func decodeJSON(data []byte, into any) error {
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.UseNumber()
	return dec.Decode(into)
}

func supportsField(action, name string) bool {
	for _, field := range cloudTrailActionFields[action] {
		if field.Name == name {
			return true
		}
	}
	return false
}

func directRedactFields(fields []string) []string {
	defaults := []string{"CloudTrailEvent", "AccessKeyId", "userIdentity", "requestParameters", "responseElements", "additionalEventData", "QueryStatement", "Prompt"}
	seen := make(map[string]bool, len(defaults)+len(fields))
	out := make([]string, 0, len(defaults)+len(fields))
	for _, field := range append(defaults, fields...) {
		if strings.TrimSpace(field) == "" || seen[field] {
			continue
		}
		seen[field] = true
		out = append(out, field)
	}
	return out
}

func redactFields(value any, fields []string) any {
	if len(fields) == 0 {
		return value
	}
	redact := make(map[string]bool, len(fields))
	for _, field := range fields {
		redact[field] = true
	}
	return redactValue(value, redact)
}

func redactValue(value any, redact map[string]bool) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, elem := range v {
			if redact[key] {
				out[key] = "[REDACTED]"
				continue
			}
			out[key] = redactValue(elem, redact)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, elem := range v {
			out[i] = redactValue(elem, redact)
		}
		return out
	default:
		return v
	}
}

func isMissingOK(writeAction string, err error) bool {
	if !cloudTrailDeleteActions[writeAction] {
		return false
	}
	var httpErr *connsdk.HTTPError
	if !errors.As(err, &httpErr) {
		return false
	}
	if httpErr.Status == 404 {
		return true
	}
	body := strings.ToLower(httpErr.Body)
	return strings.Contains(body, "notfound") || strings.Contains(body, "not found")
}

func (c Connector) readFixture(ctx context.Context, stream, action string, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	body := map[string]any{"pm_record_id": stream + "_fixture", "operation": action}
	switch action {
	case "DescribeTrails":
		body["Name"] = "fixture-trail"
		body["TrailARN"] = "arn:aws:cloudtrail:us-east-1:123456789012:trail/fixture-trail"
	default:
		body["Name"] = "fixture"
	}
	return emit(connectors.Record(body))
}
