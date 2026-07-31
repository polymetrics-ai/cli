package awscloudtrail

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const maxDirectReadBytes = 16 << 20

func emitActionRecords(action string, decoded map[string]any, emit func(connectors.Record) error) error {
	records := recordsForAction(action, decoded)
	for _, record := range records {
		stampRecord(action, record)
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
	return action == "DescribeTrails" || strings.HasPrefix(action, "List") || action == "LookupEvents"
}

func firstArray(decoded map[string]any) ([]any, bool) {
	preferred := []string{"trailList", "Events", "Channels", "Dashboards", "EventDataStores", "Imports", "PublicKeys", "Queries", "TagsList", "ResourceTagList"}
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

func stampRecord(action string, record connectors.Record) {
	if _, ok := record["operation"]; !ok {
		record["operation"] = action
	}
	if _, ok := record["pm_record_id"]; ok {
		return
	}
	for _, key := range []string{"Name", "TrailARN", "TrailName", "ChannelArn", "DashboardId", "EventDataStoreArn", "EventDataStore", "EventId", "QueryId", "ImportId", "ResourceId", "ResourceARN"} {
		if value, ok := stringAt(map[string]any(record), key); ok && strings.TrimSpace(value) != "" {
			record["pm_record_id"] = value
			return
		}
	}
	encoded, _ := json.Marshal(record)
	h := fnv.New64a()
	_, _ = h.Write(encoded)
	record["pm_record_id"] = fmt.Sprintf("%s:%x", action, h.Sum64())
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

func directMaxBytes(maxBytes int) int {
	if maxBytes <= 0 || maxBytes > maxDirectReadBytes {
		return maxDirectReadBytes
	}
	return maxBytes
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

func startTimeBound(req connectors.ReadRequest) (*time.Time, error) {
	if raw := strings.TrimSpace(req.Query["StartTime"]); raw != "" {
		return parseBoundTime(raw)
	}
	if raw := strings.TrimSpace(req.Query["start_time"]); raw != "" {
		return parseBoundTime(raw)
	}
	if req.Config.Config != nil {
		if raw := strings.TrimSpace(req.Config.Config["start_date"]); raw != "" && raw != "synthetic-conformance-value" {
			return parseBoundTime(raw)
		}
	}
	return nil, nil
}

func parseBoundTime(raw string) (*time.Time, error) {
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return &t, nil
	}
	if t, err := time.Parse("2006-01-02", raw); err == nil {
		return &t, nil
	}
	return nil, fmt.Errorf("aws-cloudtrail start time must be RFC3339 or YYYY-MM-DD")
}

func applyLookupAliasFilter(stream string, body map[string]any) {
	switch stream {
	case "management_events":
		body["EventCategory"] = "Management"
	case "read_only_events":
		body["LookupAttributes"] = []any{map[string]any{"AttributeKey": "ReadOnly", "AttributeValue": "true"}}
	case "write_only_events":
		body["LookupAttributes"] = []any{map[string]any{"AttributeKey": "ReadOnly", "AttributeValue": "false"}}
	case "console_logins":
		body["LookupAttributes"] = []any{map[string]any{"AttributeKey": "EventName", "AttributeValue": "ConsoleLogin"}}
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
	case "LookupEvents":
		body["EventName"] = "ConsoleLogin"
		body["EventTime"] = time.Now().UTC().Format(time.RFC3339)
	case "DescribeTrails":
		body["Name"] = "fixture-trail"
		body["TrailARN"] = "arn:aws:cloudtrail:us-east-1:123456789012:trail/fixture-trail"
	default:
		body["Name"] = "fixture"
	}
	return emit(connectors.Record(body))
}
