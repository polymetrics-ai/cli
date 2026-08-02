package amazonsqs

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
)

// OperationDirectRead executes the fixed typed SQS read operations that are not
// the streaming ReceiveMessage path. It is intentionally closed over the
// operation names declared in cli_surface.json; raw AWS Action values, raw
// paths, and raw request bodies are not accepted.
func (c Connector) OperationDirectRead(ctx context.Context, req connectors.OperationDirectReadRequest) (connectors.DirectReadResult, error) {
	if err := ctx.Err(); err != nil {
		return connectors.DirectReadResult{}, err
	}
	form, err := directReadForm(req.Operation, req.Config, req.Body)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	maxBytes := req.MaxBytes
	if maxBytes <= 0 || maxBytes > 16<<20 {
		maxBytes = 16 << 20
	}
	resp, err := c.doService(ctx, req.Config, form, maxBytes)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	body, err := decodeDirectReadOperation(req.Operation, resp.body)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	redacted := redactDirectBody(body, req.RedactFields)
	body, ok := redacted.(map[string]any)
	if !ok {
		return connectors.DirectReadResult{}, fmt.Errorf("redact sqs direct read %s: unexpected root type %T", req.Operation, redacted)
	}
	return connectors.DirectReadResult{Connector: c.Name(), Method: "POST", Path: "SQS." + directReadAction(req.Operation), Status: resp.status, Body: body}, nil
}

func directReadAction(operation string) string {
	switch operation {
	case "get_queue_attributes":
		return "GetQueueAttributes"
	case "get_queue_url":
		return "GetQueueUrl"
	case "list_dead_letter_source_queues":
		return "ListDeadLetterSourceQueues"
	case "list_message_move_tasks":
		return "ListMessageMoveTasks"
	case "list_queues":
		return "ListQueues"
	case "list_queue_tags":
		return "ListQueueTags"
	default:
		return operation
	}
}

func directReadForm(operation string, cfg connectors.RuntimeConfig, body map[string]any) (url.Values, error) {
	form := baseActionForm(directReadAction(operation))
	switch operation {
	case "get_queue_attributes":
		addConfiguredQueueURL(form, cfg)
		attrs := valueToStrings(body["attribute_names"])
		if len(attrs) == 0 {
			attrs = []string{"All"}
		}
		addStringList(form, "AttributeName", attrs)
	case "get_queue_url":
		queueName := bodyString(body, "queue_name")
		if queueName == "" {
			return nil, fmt.Errorf("amazon-sqs direct read %s requires queue_name", operation)
		}
		form.Set("QueueName", queueName)
		if owner := bodyString(body, "queue_owner_aws_account_id"); owner != "" {
			form.Set("QueueOwnerAWSAccountId", owner)
		}
	case "list_dead_letter_source_queues":
		addConfiguredQueueURL(form, cfg)
		if token := bodyString(body, "next_token"); token != "" {
			form.Set("NextToken", token)
		}
		if maxResults, ok := bodyInt(body, "max_results", 0, 1, 1000); ok {
			form.Set("MaxResults", strconv.Itoa(maxResults))
		}
	case "list_message_move_tasks":
		sourceArn := bodyString(body, "source_arn")
		if sourceArn == "" {
			return nil, fmt.Errorf("amazon-sqs direct read %s requires source_arn", operation)
		}
		form.Set("SourceArn", sourceArn)
		if maxResults, ok := bodyInt(body, "max_results", 0, 1, 10); ok {
			form.Set("MaxResults", strconv.Itoa(maxResults))
		}
	case "list_queues":
		if prefix := bodyString(body, "queue_name_prefix"); prefix != "" {
			form.Set("QueueNamePrefix", prefix)
		}
		if token := bodyString(body, "next_token"); token != "" {
			form.Set("NextToken", token)
		}
		if maxResults, ok := bodyInt(body, "max_results", 0, 1, 1000); ok {
			form.Set("MaxResults", strconv.Itoa(maxResults))
		}
	case "list_queue_tags":
		addConfiguredQueueURL(form, cfg)
	default:
		return nil, fmt.Errorf("amazon-sqs direct read operation %q not found", operation)
	}
	return form, nil
}

func bodyString(body map[string]any, key string) string {
	if body == nil || body[key] == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(body[key]))
}

func bodyInt(body map[string]any, key string, def, min, max int) (int, bool) {
	if body == nil || body[key] == nil {
		return def, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(fmt.Sprint(body[key])))
	if err != nil {
		n = def
	}
	if n < min {
		n = min
	}
	if max > min && n > max {
		n = max
	}
	return n, true
}

type getQueueAttributesXML struct {
	Attributes []nameValue `xml:"GetQueueAttributesResult>Attribute"`
}

type getQueueURLXML struct {
	QueueURL string `xml:"GetQueueUrlResult>QueueUrl"`
}

type listQueuesXML struct {
	QueueURLs []string `xml:"ListQueuesResult>QueueUrl"`
	NextToken string   `xml:"ListQueuesResult>NextToken"`
}

type listDeadLetterSourceQueuesXML struct {
	QueueURLs []string `xml:"ListDeadLetterSourceQueuesResult>QueueUrl"`
	NextToken string   `xml:"ListDeadLetterSourceQueuesResult>NextToken"`
}

type listQueueTagsXML struct {
	Tags []tagValue `xml:"ListQueueTagsResult>Tag"`
}

type tagValue struct {
	Key   string `xml:"Key"`
	Value string `xml:"Value"`
}

type listMessageMoveTasksXML struct {
	Results []messageMoveTaskXML `xml:"ListMessageMoveTasksResult>Result"`
}

type messageMoveTaskXML struct {
	TaskHandle                        string `xml:"TaskHandle"`
	Status                            string `xml:"Status"`
	SourceArn                         string `xml:"SourceArn"`
	DestinationArn                    string `xml:"DestinationArn"`
	MaxNumberOfMessagesPerSecond      string `xml:"MaxNumberOfMessagesPerSecond"`
	ApproximateNumberOfMessagesMoved  string `xml:"ApproximateNumberOfMessagesMoved"`
	ApproximateNumberOfMessagesToMove string `xml:"ApproximateNumberOfMessagesToMove"`
	FailureReason                     string `xml:"FailureReason"`
	StartedTimestamp                  string `xml:"StartedTimestamp"`
}

func decodeDirectReadOperation(operation string, raw []byte) (map[string]any, error) {
	switch operation {
	case "get_queue_attributes":
		var out getQueueAttributesXML
		if err := xml.Unmarshal(raw, &out); err != nil {
			return nil, fmt.Errorf("decode sqs get_queue_attributes xml: %w", err)
		}
		return map[string]any{"attributes": nameValuesToMap(out.Attributes)}, nil
	case "get_queue_url":
		var out getQueueURLXML
		if err := xml.Unmarshal(raw, &out); err != nil {
			return nil, fmt.Errorf("decode sqs get_queue_url xml: %w", err)
		}
		return map[string]any{"queue_url": out.QueueURL}, nil
	case "list_dead_letter_source_queues":
		var out listDeadLetterSourceQueuesXML
		if err := xml.Unmarshal(raw, &out); err != nil {
			return nil, fmt.Errorf("decode sqs list_dead_letter_source_queues xml: %w", err)
		}
		return withOptionalToken(map[string]any{"queue_urls": out.QueueURLs}, out.NextToken), nil
	case "list_message_move_tasks":
		var out listMessageMoveTasksXML
		if err := xml.Unmarshal(raw, &out); err != nil {
			return nil, fmt.Errorf("decode sqs list_message_move_tasks xml: %w", err)
		}
		items := make([]any, 0, len(out.Results))
		for _, task := range out.Results {
			items = append(items, task.toRecord())
		}
		return map[string]any{"results": items}, nil
	case "list_queues":
		var out listQueuesXML
		if err := xml.Unmarshal(raw, &out); err != nil {
			return nil, fmt.Errorf("decode sqs list_queues xml: %w", err)
		}
		return withOptionalToken(map[string]any{"queue_urls": out.QueueURLs}, out.NextToken), nil
	case "list_queue_tags":
		var out listQueueTagsXML
		if err := xml.Unmarshal(raw, &out); err != nil {
			return nil, fmt.Errorf("decode sqs list_queue_tags xml: %w", err)
		}
		return map[string]any{"tags": tagValuesToMap(out.Tags)}, nil
	default:
		return nil, fmt.Errorf("amazon-sqs direct read operation %q not found", operation)
	}
}

func (t messageMoveTaskXML) toRecord() map[string]any {
	out := map[string]any{}
	setIfNotEmpty(out, "task_handle", t.TaskHandle)
	setIfNotEmpty(out, "status", t.Status)
	setIfNotEmpty(out, "source_arn", t.SourceArn)
	setIfNotEmpty(out, "destination_arn", t.DestinationArn)
	setIfNotEmpty(out, "max_number_of_messages_per_second", t.MaxNumberOfMessagesPerSecond)
	setIfNotEmpty(out, "approximate_number_of_messages_moved", t.ApproximateNumberOfMessagesMoved)
	setIfNotEmpty(out, "approximate_number_of_messages_to_move", t.ApproximateNumberOfMessagesToMove)
	setIfNotEmpty(out, "failure_reason", t.FailureReason)
	setIfNotEmpty(out, "started_timestamp", t.StartedTimestamp)
	return out
}

func nameValuesToMap(values []nameValue) map[string]any {
	out := map[string]any{}
	for _, item := range values {
		if item.Name != "" {
			out[item.Name] = item.Value
		}
	}
	return out
}

func tagValuesToMap(values []tagValue) map[string]any {
	out := map[string]any{}
	for _, item := range values {
		if item.Key != "" {
			out[item.Key] = item.Value
		}
	}
	return out
}

func withOptionalToken(out map[string]any, token string) map[string]any {
	if strings.TrimSpace(token) != "" {
		out["next_token"] = token
	}
	return out
}

func setIfNotEmpty(out map[string]any, key, value string) {
	if strings.TrimSpace(value) != "" {
		out[key] = value
	}
}

func redactDirectBody(value any, fields []string) any {
	explicit := map[string]bool{}
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		explicit[normalizeField(field)] = true
	}
	return redactDirectValue("", value, explicit)
}

func redactDirectValue(key string, value any, explicit map[string]bool) any {
	if shouldRedactDirectField(key, explicit) {
		return "***"
	}
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		keys := make([]string, 0, len(typed))
		for k := range typed {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			out[k] = redactDirectValue(k, typed[k], explicit)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = redactDirectValue(key, item, explicit)
		}
		return out
	default:
		return value
	}
}

func shouldRedactDirectField(key string, explicit map[string]bool) bool {
	normalized := normalizeField(key)
	if explicit[normalized] {
		return true
	}
	if normalized == "next_token" {
		return false
	}
	for _, marker := range []string{"receipt_handle", "task_handle", "policy", "secret", "token", "password"} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func normalizeField(key string) string {
	return strings.ToLower(strings.NewReplacer("-", "_", ".", "_", " ", "_").Replace(key))
}
