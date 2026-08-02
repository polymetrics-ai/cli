package amazonsqs

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
)

const sqsBatchLimit = 10

type sqsIntFieldRule struct {
	min int
	max int
}

var sqsIntFieldRules = map[string]sqsIntFieldRule{
	"delay_seconds":                     {min: 0, max: 900},
	"max_number_of_messages_per_second": {min: 1, max: 500},
	"visibility_timeout":                {min: 0, max: 43200},
}

type writeActionDef struct {
	name     string
	method   string
	path     string
	kind     string
	required []string
	allowed  []string
	redact   []string
	risk     string
	confirm  string
	batch    bool
	queue    bool
	service  bool
	execute  func(url.Values, connectors.Record, int) error
}

var sqsWriteActions = map[string]writeActionDef{
	"add_permission":                  {name: "add_permission", method: "POST", path: "SQS.AddPermission", kind: "custom", required: []string{"label", "aws_account_ids", "actions"}, allowed: []string{"label", "aws_account_ids", "actions"}, risk: "adds an SQS queue resource policy permission statement for listed AWS account ids", queue: true, execute: buildAddPermissionForm},
	"cancel_message_move_task":        {name: "cancel_message_move_task", method: "POST", path: "SQS.CancelMessageMoveTask", kind: "custom", required: []string{"task_handle"}, allowed: []string{"task_handle"}, redact: []string{"task_handle"}, risk: "cancels an in-flight dead-letter-queue message move task", confirm: "destructive", service: true, execute: buildCancelMessageMoveTaskForm},
	"change_message_visibility":       {name: "change_message_visibility", method: "POST", path: "SQS.ChangeMessageVisibility", kind: "update", required: []string{"receipt_handle", "visibility_timeout"}, allowed: []string{"receipt_handle", "visibility_timeout"}, redact: []string{"receipt_handle"}, risk: "changes the visibility timeout for one in-flight message", queue: true, execute: buildChangeMessageVisibilityForm},
	"change_message_visibility_batch": {name: "change_message_visibility_batch", method: "POST", path: "SQS.ChangeMessageVisibilityBatch", kind: "update", required: []string{"receipt_handle"}, allowed: []string{"id", "receipt_handle", "visibility_timeout"}, redact: []string{"receipt_handle"}, risk: "changes visibility timeout for up to 10 in-flight messages per SQS batch request", batch: true, queue: true, execute: buildChangeMessageVisibilityBatchEntry},
	"create_queue":                    {name: "create_queue", method: "POST", path: "SQS.CreateQueue", kind: "create", required: []string{"queue_name"}, allowed: []string{"queue_name", "attributes", "tags"}, risk: "creates an SQS queue; SQS returns an existing queue URL when name and attributes match", service: true, execute: buildCreateQueueForm},
	"delete_message":                  {name: "delete_message", method: "POST", path: "SQS.DeleteMessage", kind: "delete", required: []string{"receipt_handle"}, allowed: []string{"receipt_handle"}, redact: []string{"receipt_handle"}, risk: "deletes one received message by receipt handle", confirm: "destructive", queue: true, execute: buildDeleteMessageForm},
	"delete_message_batch":            {name: "delete_message_batch", method: "POST", path: "SQS.DeleteMessageBatch", kind: "delete", required: []string{"receipt_handle"}, allowed: []string{"id", "receipt_handle"}, redact: []string{"receipt_handle"}, risk: "deletes up to 10 received messages per SQS batch request", confirm: "destructive", batch: true, queue: true, execute: buildDeleteMessageBatchEntry},
	"delete_queue":                    {name: "delete_queue", method: "POST", path: "SQS.DeleteQueue", kind: "delete", allowed: []string{}, risk: "deletes the configured SQS queue", confirm: "destructive", queue: true, execute: buildNoopForm},
	"purge_queue":                     {name: "purge_queue", method: "POST", path: "SQS.PurgeQueue", kind: "delete", allowed: []string{}, risk: "purges all available messages from the configured queue", confirm: "destructive", queue: true, execute: buildNoopForm},
	"remove_permission":               {name: "remove_permission", method: "POST", path: "SQS.RemovePermission", kind: "delete", required: []string{"label"}, allowed: []string{"label"}, risk: "removes an SQS queue resource policy permission statement", confirm: "destructive", queue: true, execute: buildRemovePermissionForm},
	"send_message":                    {name: "send_message", method: "POST", path: "SQS.SendMessage", kind: "create", required: []string{"message_body"}, allowed: []string{"message_body", "delay_seconds", "message_attributes", "message_system_attributes", "message_deduplication_id", "message_group_id"}, redact: []string{"message_body", "message_attributes", "message_system_attributes"}, risk: "sends one message to the configured queue; FIFO queues may use message_deduplication_id for provider-supported idempotency", queue: true, execute: buildSendMessageForm},
	"send_message_batch":              {name: "send_message_batch", method: "POST", path: "SQS.SendMessageBatch", kind: "create", required: []string{"message_body"}, allowed: []string{"id", "message_body", "delay_seconds", "message_attributes", "message_system_attributes", "message_deduplication_id", "message_group_id"}, redact: []string{"message_body", "message_attributes", "message_system_attributes"}, risk: "sends up to 10 messages per SQS batch request; FIFO queues may use message_deduplication_id for provider-supported idempotency", batch: true, queue: true, execute: buildSendMessageBatchEntry},
	"set_queue_attributes":            {name: "set_queue_attributes", method: "POST", path: "SQS.SetQueueAttributes", kind: "update", required: []string{"attribute_name", "attribute_value"}, allowed: []string{"attribute_name", "attribute_value", "attributes"}, redact: []string{"attribute_value", "attributes"}, risk: "sets typed SQS queue attributes such as policy, redrive, encryption, retention, and visibility settings", queue: true, execute: buildSetQueueAttributesForm},
	"start_message_move_task":         {name: "start_message_move_task", method: "POST", path: "SQS.StartMessageMoveTask", kind: "custom", required: []string{"source_arn"}, allowed: []string{"source_arn", "destination_arn", "max_number_of_messages_per_second"}, risk: "starts an SQS dead-letter queue redrive message move task", service: true, execute: buildStartMessageMoveTaskForm},
	"tag_queue":                       {name: "tag_queue", method: "POST", path: "SQS.TagQueue", kind: "update", required: []string{"tag_key", "tag_value"}, allowed: []string{"tag_key", "tag_value", "tags"}, risk: "adds or updates tags on the configured SQS queue", queue: true, execute: buildTagQueueForm},
	"untag_queue":                     {name: "untag_queue", method: "POST", path: "SQS.UntagQueue", kind: "delete", required: []string{"tag_keys"}, allowed: []string{"tag_keys"}, risk: "removes tags from the configured SQS queue", confirm: "destructive", queue: true, execute: buildUntagQueueForm},
}

func (c Connector) Manifest() connectors.Manifest {
	base := c.Base.Definition()
	cat, _ := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	fields := []connectors.ConfigField{}
	secrets := []connectors.SecretField{}
	for _, name := range []string{"queue_url", "region", "endpoint_url", "max_batch_size", "max_wait_time", "visibility_timeout", "attributes_to_return", "system_attributes_to_return", "max_polls", "mode"} {
		fields = append(fields, connectors.ConfigField{Name: name})
	}
	for _, name := range []string{"access_key", "secret_key", "session_token"} {
		secrets = append(secrets, connectors.SecretField{Name: name})
	}
	actions := make([]connectors.WriteActionSpec, 0, len(sqsWriteActions))
	for _, name := range sortedWriteActionNames() {
		def := sqsWriteActions[name]
		actions = append(actions, connectors.WriteActionSpec{
			Name:           def.name,
			RequiredFields: append([]string(nil), def.required...),
			OptionalFields: optionalFields(def),
			Method:         def.method,
			Path:           def.path,
			RedactFields:   append([]string(nil), def.redact...),
			Risk:           def.risk,
			Confirm:        def.confirm,
		})
	}
	return connectors.Manifest{
		Metadata:     c.Metadata(),
		ConfigFields: fields,
		SecretFields: secrets,
		Streams:      cat.Streams,
		WriteActions: actions,
		SyncModes:    []string{"full_refresh_append", "full_refresh_overwrite", "full_refresh_overwrite_deduped"},
		Risk: connectors.RiskSpec{
			Read:     base.Risk.Read,
			Write:    base.Risk.Write,
			Approval: base.Risk.Approval,
		},
	}
}

func optionalFields(def writeActionDef) []string {
	required := map[string]bool{}
	for _, field := range def.required {
		required[field] = true
	}
	var out []string
	for _, field := range def.allowed {
		if !required[field] {
			out = append(out, field)
		}
	}
	return out
}

func sortedWriteActionNames() []string {
	names := make([]string, 0, len(sqsWriteActions))
	for name := range sqsWriteActions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (c Connector) ValidateWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	def, _, err := lookupSQSWriteAction(req.Action)
	if err != nil {
		return err
	}
	return validateSQSRecords(def, normalizedWriteRecords(records))
}

func (c Connector) DryRunWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WritePreview, error) {
	if err := ctx.Err(); err != nil {
		return connectors.WritePreview{}, err
	}
	def, action, err := lookupSQSWriteAction(req.Action)
	if err != nil {
		return connectors.WritePreview{}, err
	}
	normalized := normalizedWriteRecords(records)
	if err := validateSQSRecords(def, normalized); err != nil {
		return connectors.WritePreview{}, err
	}
	staged := len(normalized)
	warnings := []string{"amazon-sqs writes require reverse ETL plan -> preview -> explicit approval -> execute"}
	if def.confirm == "destructive" {
		warnings = append(warnings, "destructive confirmation required")
	}
	return connectors.WritePreview{RecordsStaged: staged, Action: action, Warnings: warnings}, nil
}

func (c Connector) writeSQS(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	if err := ctx.Err(); err != nil {
		return connectors.WriteResult{RecordsFailed: len(records)}, err
	}
	def, _, err := lookupSQSWriteAction(req.Action)
	if err != nil {
		return connectors.WriteResult{RecordsFailed: len(records)}, err
	}
	normalized := normalizedWriteRecords(records)
	if err := validateSQSRecords(def, normalized); err != nil {
		return connectors.WriteResult{RecordsFailed: len(records)}, err
	}
	written := 0
	failed := 0
	if def.batch {
		for start := 0; start < len(normalized); start += sqsBatchLimit {
			end := start + sqsBatchLimit
			if end > len(normalized) {
				end = len(normalized)
			}
			chunk := normalized[start:end]
			form := baseActionForm(actionName(def.path))
			addConfiguredQueueURL(form, req.Config)
			for i, rec := range chunk {
				if err := def.execute(form, rec, i+1); err != nil {
					return connectors.WriteResult{RecordsWritten: written, RecordsFailed: failed + len(chunk) - i}, err
				}
			}
			resp, err := c.doService(ctx, req.Config, form, 16<<20)
			if err != nil {
				return connectors.WriteResult{RecordsWritten: written, RecordsFailed: failed + len(chunk)}, err
			}
			successes, failures, err := parseBatchCounts(resp.body)
			if successes == 0 && failures == 0 && err == nil {
				successes = len(chunk)
			}
			written += successes
			failed += failures
			if err != nil {
				if unknown := len(chunk) - successes - failures; unknown > 0 {
					failed += unknown
				}
				return connectors.WriteResult{RecordsWritten: written, RecordsFailed: failed}, err
			}
		}
		return connectors.WriteResult{RecordsWritten: written, RecordsFailed: failed}, nil
	}
	for i, rec := range normalized {
		form := baseActionForm(actionName(def.path))
		if def.queue {
			addConfiguredQueueURL(form, req.Config)
		}
		if err := def.execute(form, rec, i+1); err != nil {
			return connectors.WriteResult{RecordsWritten: written, RecordsFailed: len(normalized) - written}, err
		}
		_, err := c.doService(ctx, req.Config, form, 16<<20)
		if err != nil {
			return connectors.WriteResult{RecordsWritten: written, RecordsFailed: len(normalized) - written}, err
		}
		written++
	}
	return connectors.WriteResult{RecordsWritten: written}, nil
}

func lookupSQSWriteAction(raw string) (writeActionDef, string, error) {
	action := strings.TrimSpace(raw)
	def, ok := sqsWriteActions[action]
	if !ok {
		return writeActionDef{}, "", fmt.Errorf("amazon-sqs write action %q not found", raw)
	}
	return def, action, nil
}

func normalizedWriteRecords(records []connectors.Record) []connectors.Record {
	out := make([]connectors.Record, len(records))
	for i, rec := range records {
		if rec == nil {
			out[i] = connectors.Record{}
		} else {
			out[i] = rec
		}
	}
	return out
}

func validateSQSRecords(def writeActionDef, records []connectors.Record) error {
	allowed := map[string]bool{}
	for _, field := range def.allowed {
		allowed[field] = true
	}
	for i, rec := range records {
		for _, field := range def.required {
			if isEmptyRequiredRecordValue(field, rec[field]) {
				return fmt.Errorf("amazon-sqs action %s record %d requires field %q", def.name, i, field)
			}
		}
		for field, value := range rec {
			if !allowed[field] {
				return fmt.Errorf("amazon-sqs action %s record %d has unsupported field %q", def.name, i, field)
			}
			if err := validateSQSFieldValue(field, value); err != nil {
				return fmt.Errorf("amazon-sqs action %s record %d field %q: %w", def.name, i, field, err)
			}
		}
	}
	return nil
}

func validateSQSFieldValue(field string, value any) error {
	if field == "message_body" {
		if _, ok := value.(string); !ok {
			return fmt.Errorf("must be a string")
		}
		return nil
	}
	rule, ok := sqsIntFieldRules[field]
	if !ok || isEmptyRecordValue(value) {
		return nil
	}
	_, err := parseSQSInt(value, rule.min, rule.max)
	return err
}

func isEmptyRequiredRecordValue(field string, v any) bool {
	if field == "message_body" {
		return isEmptyPayloadValue(v)
	}
	return isEmptyRecordValue(v)
}

func isEmptyPayloadValue(v any) bool {
	switch typed := v.(type) {
	case nil:
		return true
	case string:
		return typed == ""
	default:
		return fmt.Sprint(v) == ""
	}
}

func isEmptyRecordValue(v any) bool {
	switch typed := v.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(typed) == ""
	case []string:
		return len(typed) == 0
	case []any:
		return len(typed) == 0
	case map[string]string:
		return len(typed) == 0
	case map[string]any:
		return len(typed) == 0
	case connectors.Record:
		return len(typed) == 0
	default:
		return false
	}
}

func baseActionForm(action string) url.Values {
	return url.Values{"Action": {action}, "Version": {apiVersion}}
}

func actionName(path string) string {
	_, action, ok := strings.Cut(path, ".")
	if !ok {
		return path
	}
	return action
}

func buildNoopForm(url.Values, connectors.Record, int) error { return nil }

func buildAddPermissionForm(form url.Values, rec connectors.Record, _ int) error {
	form.Set("Label", stringField(rec, "label"))
	addStringList(form, "AWSAccountId", stringSliceField(rec, "aws_account_ids"))
	addStringList(form, "ActionName", stringSliceField(rec, "actions"))
	return nil
}

func buildCancelMessageMoveTaskForm(form url.Values, rec connectors.Record, _ int) error {
	form.Set("TaskHandle", stringField(rec, "task_handle"))
	return nil
}

func buildChangeMessageVisibilityForm(form url.Values, rec connectors.Record, _ int) error {
	form.Set("ReceiptHandle", stringField(rec, "receipt_handle"))
	visibilityTimeout, err := intField(rec, "visibility_timeout", 0, 43200)
	if err != nil {
		return err
	}
	form.Set("VisibilityTimeout", strconv.Itoa(visibilityTimeout))
	return nil
}

func buildChangeMessageVisibilityBatchEntry(form url.Values, rec connectors.Record, index int) error {
	prefix := fmt.Sprintf("ChangeMessageVisibilityBatchRequestEntry.%d.", index)
	form.Set(prefix+"Id", entryID(rec, index))
	form.Set(prefix+"ReceiptHandle", stringField(rec, "receipt_handle"))
	if !isEmptyRecordValue(rec["visibility_timeout"]) {
		visibilityTimeout, err := intField(rec, "visibility_timeout", 0, 43200)
		if err != nil {
			return err
		}
		form.Set(prefix+"VisibilityTimeout", strconv.Itoa(visibilityTimeout))
	}
	return nil
}

func buildCreateQueueForm(form url.Values, rec connectors.Record, _ int) error {
	form.Set("QueueName", stringField(rec, "queue_name"))
	addStringMap(form, "Attribute", stringMapField(rec, "attributes"))
	addTagMap(form, "Tag", stringMapField(rec, "tags"))
	return nil
}

func buildDeleteMessageForm(form url.Values, rec connectors.Record, _ int) error {
	form.Set("ReceiptHandle", stringField(rec, "receipt_handle"))
	return nil
}

func buildDeleteMessageBatchEntry(form url.Values, rec connectors.Record, index int) error {
	prefix := fmt.Sprintf("DeleteMessageBatchRequestEntry.%d.", index)
	form.Set(prefix+"Id", entryID(rec, index))
	form.Set(prefix+"ReceiptHandle", stringField(rec, "receipt_handle"))
	return nil
}

func buildRemovePermissionForm(form url.Values, rec connectors.Record, _ int) error {
	form.Set("Label", stringField(rec, "label"))
	return nil
}

func buildSendMessageForm(form url.Values, rec connectors.Record, _ int) error {
	body, err := messageBodyField(rec)
	if err != nil {
		return err
	}
	form.Set("MessageBody", body)
	if err := addOptionalInt(form, "DelaySeconds", rec, "delay_seconds", 0, 900); err != nil {
		return err
	}
	addOptionalString(form, "MessageDeduplicationId", rec, "message_deduplication_id")
	addOptionalString(form, "MessageGroupId", rec, "message_group_id")
	addMessageAttributeMap(form, "MessageAttribute", messageAttributeMapField(rec, "message_attributes"))
	addMessageAttributeMap(form, "MessageSystemAttribute", messageAttributeMapField(rec, "message_system_attributes"))
	return nil
}

func buildSendMessageBatchEntry(form url.Values, rec connectors.Record, index int) error {
	prefix := fmt.Sprintf("SendMessageBatchRequestEntry.%d.", index)
	form.Set(prefix+"Id", entryID(rec, index))
	body, err := messageBodyField(rec)
	if err != nil {
		return err
	}
	form.Set(prefix+"MessageBody", body)
	if err := addOptionalInt(form, prefix+"DelaySeconds", rec, "delay_seconds", 0, 900); err != nil {
		return err
	}
	addOptionalString(form, prefix+"MessageDeduplicationId", rec, "message_deduplication_id")
	addOptionalString(form, prefix+"MessageGroupId", rec, "message_group_id")
	addMessageAttributeMap(form, prefix+"MessageAttribute", messageAttributeMapField(rec, "message_attributes"))
	addMessageAttributeMap(form, prefix+"MessageSystemAttribute", messageAttributeMapField(rec, "message_system_attributes"))
	return nil
}

func buildSetQueueAttributesForm(form url.Values, rec connectors.Record, _ int) error {
	attrs := stringMapField(rec, "attributes")
	if len(attrs) == 0 {
		attrs = map[string]string{stringField(rec, "attribute_name"): stringField(rec, "attribute_value")}
	}
	addStringMap(form, "Attribute", attrs)
	return nil
}

func buildStartMessageMoveTaskForm(form url.Values, rec connectors.Record, _ int) error {
	form.Set("SourceArn", stringField(rec, "source_arn"))
	addOptionalString(form, "DestinationArn", rec, "destination_arn")
	if err := addOptionalInt(form, "MaxNumberOfMessagesPerSecond", rec, "max_number_of_messages_per_second", 1, 500); err != nil {
		return err
	}
	return nil
}

func buildTagQueueForm(form url.Values, rec connectors.Record, _ int) error {
	tags := stringMapField(rec, "tags")
	if len(tags) == 0 {
		tags = map[string]string{stringField(rec, "tag_key"): stringField(rec, "tag_value")}
	}
	addTagMap(form, "Tag", tags)
	return nil
}

func buildUntagQueueForm(form url.Values, rec connectors.Record, _ int) error {
	addStringList(form, "TagKey", stringSliceField(rec, "tag_keys"))
	return nil
}

func addOptionalString(form url.Values, name string, rec connectors.Record, field string) {
	if value := stringField(rec, field); value != "" {
		form.Set(name, value)
	}
}

func addOptionalInt(form url.Values, name string, rec connectors.Record, field string, minValue, maxValue int) error {
	if isEmptyRecordValue(rec[field]) {
		return nil
	}
	value, err := intField(rec, field, minValue, maxValue)
	if err != nil {
		return err
	}
	form.Set(name, strconv.Itoa(value))
	return nil
}

func entryID(rec connectors.Record, index int) string {
	if id := stringField(rec, "id"); id != "" {
		return id
	}
	return fmt.Sprintf("entry_%d", index)
}

func stringField(rec connectors.Record, key string) string {
	return strings.TrimSpace(rawStringField(rec, key))
}

func rawStringField(rec connectors.Record, key string) string {
	value, ok := rec[key]
	if !ok || value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func messageBodyField(rec connectors.Record) (string, error) {
	value, ok := rec["message_body"]
	if !ok || value == nil {
		return "", fmt.Errorf("field %q is required", "message_body")
	}
	body, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("field %q must be a string", "message_body")
	}
	if body == "" {
		return "", fmt.Errorf("field %q is required", "message_body")
	}
	return body, nil
}

func intField(rec connectors.Record, key string, minValue, maxValue int) (int, error) {
	value, ok := rec[key]
	if !ok || value == nil {
		return 0, fmt.Errorf("field %q is required", key)
	}
	return parseSQSInt(value, minValue, maxValue)
}

func parseSQSInt(value any, minValue, maxValue int) (int, error) {
	rangeErr := func() error {
		return fmt.Errorf("must be between %d and %d", minValue, maxValue)
	}
	integerErr := func() error {
		return fmt.Errorf("must be an integer")
	}
	switch typed := value.(type) {
	case int:
		if typed < minValue || typed > maxValue {
			return 0, rangeErr()
		}
		return typed, nil
	case int8:
		return parseSQSInt(int(typed), minValue, maxValue)
	case int16:
		return parseSQSInt(int(typed), minValue, maxValue)
	case int32:
		return parseSQSInt(int(typed), minValue, maxValue)
	case int64:
		if typed < int64(minValue) || typed > int64(maxValue) {
			return 0, rangeErr()
		}
		return int(typed), nil
	case uint:
		if typed > uint(maxValue) {
			return 0, rangeErr()
		}
		return int(typed), nil
	case uint8:
		return parseSQSInt(uint(typed), minValue, maxValue)
	case uint16:
		return parseSQSInt(uint(typed), minValue, maxValue)
	case uint32:
		return parseSQSInt(uint(typed), minValue, maxValue)
	case uint64:
		if typed > uint64(maxValue) {
			return 0, rangeErr()
		}
		return int(typed), nil
	case float32:
		return parseSQSFloat(float64(typed), minValue, maxValue, integerErr, rangeErr)
	case float64:
		return parseSQSFloat(typed, minValue, maxValue, integerErr, rangeErr)
	case json.Number:
		return parseSQSIntString(typed.String(), minValue, maxValue, integerErr, rangeErr)
	case string:
		return parseSQSIntString(typed, minValue, maxValue, integerErr, rangeErr)
	default:
		return parseSQSIntString(fmt.Sprint(value), minValue, maxValue, integerErr, rangeErr)
	}
}

func parseSQSFloat(value float64, minValue, maxValue int, integerErr func() error, rangeErr func() error) (int, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) || math.Trunc(value) != value {
		return 0, integerErr()
	}
	if value < float64(minValue) || value > float64(maxValue) {
		return 0, rangeErr()
	}
	return int(value), nil
}

func parseSQSIntString(raw string, minValue, maxValue int, integerErr func() error, rangeErr func() error) (int, error) {
	parsed, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return 0, integerErr()
	}
	if parsed < int64(minValue) || parsed > int64(maxValue) {
		return 0, rangeErr()
	}
	return int(parsed), nil
}

func stringSliceField(rec connectors.Record, key string) []string {
	value := rec[key]
	switch typed := value.(type) {
	case []string:
		return compactStrings(typed)
	case []any:
		out := make([]string, 0, len(typed))
		for _, v := range typed {
			out = append(out, fmt.Sprint(v))
		}
		return compactStrings(out)
	case string:
		return splitCSV(typed)
	default:
		if value == nil {
			return nil
		}
		return compactStrings([]string{fmt.Sprint(value)})
	}
}

func compactStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if s := strings.TrimSpace(value); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func stringMapField(rec connectors.Record, key string) map[string]string {
	value := rec[key]
	switch typed := value.(type) {
	case map[string]string:
		return cloneStringMap(typed)
	case map[string]any:
		out := map[string]string{}
		for k, v := range typed {
			out[k] = fmt.Sprint(v)
		}
		return out
	case connectors.Record:
		out := map[string]string{}
		for k, v := range typed {
			out[k] = fmt.Sprint(v)
		}
		return out
	default:
		return nil
	}
}

func cloneStringMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

type messageAttributeValue struct {
	DataType         string
	StringValue      string
	BinaryValue      string
	StringListValues []string
	BinaryListValues []string
}

func messageAttributeMapField(rec connectors.Record, key string) map[string]messageAttributeValue {
	value := rec[key]
	items, ok := value.(map[string]any)
	if !ok {
		if cr, ok := value.(connectors.Record); ok {
			items = map[string]any(cr)
		} else {
			return nil
		}
	}
	out := map[string]messageAttributeValue{}
	for name, raw := range items {
		m, ok := raw.(map[string]any)
		if !ok {
			if cr, ok := raw.(connectors.Record); ok {
				m = map[string]any(cr)
			} else {
				out[name] = messageAttributeValue{DataType: "String", StringValue: rawStringValue(raw)}
				continue
			}
		}
		out[name] = messageAttributeValue{
			DataType:         trimmedStringValue(m["data_type"]),
			StringValue:      trimmedStringValue(m["string_value"]),
			BinaryValue:      trimmedStringValue(m["binary_value"]),
			StringListValues: valueToStrings(m["string_list_values"]),
			BinaryListValues: valueToStrings(m["binary_list_values"]),
		}
	}
	return out
}

func rawStringValue(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func trimmedStringValue(value any) string {
	return strings.TrimSpace(rawStringValue(value))
}

func valueToStrings(value any) []string {
	switch typed := value.(type) {
	case []string:
		return compactStrings(typed)
	case []any:
		out := make([]string, 0, len(typed))
		for _, v := range typed {
			out = append(out, fmt.Sprint(v))
		}
		return compactStrings(out)
	case string:
		return splitCSV(typed)
	default:
		return nil
	}
}

func addStringList(form url.Values, prefix string, values []string) {
	for i, value := range values {
		form.Set(fmt.Sprintf("%s.%d", prefix, i+1), value)
	}
}

func addStringMap(form url.Values, prefix string, values map[string]string) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for i, key := range keys {
		form.Set(fmt.Sprintf("%s.%d.Name", prefix, i+1), key)
		form.Set(fmt.Sprintf("%s.%d.Value", prefix, i+1), values[key])
	}
}

func addTagMap(form url.Values, prefix string, values map[string]string) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for i, key := range keys {
		form.Set(fmt.Sprintf("%s.%d.Key", prefix, i+1), key)
		form.Set(fmt.Sprintf("%s.%d.Value", prefix, i+1), values[key])
	}
}

func addMessageAttributeMap(form url.Values, prefix string, values map[string]messageAttributeValue) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for i, key := range keys {
		attr := values[key]
		base := fmt.Sprintf("%s.%d.", prefix, i+1)
		form.Set(base+"Name", key)
		if attr.DataType == "" {
			attr.DataType = "String"
		}
		form.Set(base+"Value.DataType", attr.DataType)
		if attr.StringValue != "" {
			form.Set(base+"Value.StringValue", attr.StringValue)
		}
		if attr.BinaryValue != "" {
			form.Set(base+"Value.BinaryValue", attr.BinaryValue)
		}
		addStringList(form, base+"Value.StringListValue", attr.StringListValues)
		addStringList(form, base+"Value.BinaryListValue", attr.BinaryListValues)
	}
}

type sqsBatchFailure struct {
	Code string `xml:"Code"`
}

func parseBatchCounts(raw []byte) (int, int, error) {
	decoder := xml.NewDecoder(strings.NewReader(string(raw)))
	successes := 0
	failures := 0
	firstFailureCode := ""
	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return successes, failures, fmt.Errorf("parse amazon-sqs batch response: %w", err)
		}
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch start.Name.Local {
		case "SendMessageBatchResultEntry", "DeleteMessageBatchResultEntry", "ChangeMessageVisibilityBatchResultEntry", "Successful":
			successes++
		case "BatchResultErrorEntry", "Failed":
			failures++
			var failure sqsBatchFailure
			if err := decoder.DecodeElement(&failure, &start); err != nil {
				return successes, failures, fmt.Errorf("parse amazon-sqs batch failure entry: %w", err)
			}
			if firstFailureCode == "" {
				firstFailureCode = strings.TrimSpace(failure.Code)
			}
		}
	}
	if failures == 0 {
		return successes, failures, nil
	}
	if firstFailureCode != "" {
		return successes, failures, fmt.Errorf("amazon-sqs batch request failed for %d entries with first code %s", failures, firstFailureCode)
	}
	return successes, failures, fmt.Errorf("amazon-sqs batch request failed for %d entries", failures)
}
