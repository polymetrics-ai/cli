package amazonsqs

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

const sqsBatchLimit = 10

type writeActionDef struct {
	name        string
	method      string
	path        string
	kind        string
	required    []string
	requiredAny [][]string
	allowed     []string
	redact      []string
	risk        string
	confirm     string
	batch       bool
	queue       bool
	service     bool
	execute     func(url.Values, connectors.Record, int) error
}

var sqsWriteActions = map[string]writeActionDef{
	"add_permission":                  {name: "add_permission", method: "POST", path: "SQS.AddPermission", kind: "custom", required: []string{"label", "aws_account_ids", "actions"}, allowed: []string{"label", "aws_account_ids", "actions"}, risk: "adds an SQS queue resource policy permission statement for listed AWS account ids", queue: true, execute: buildAddPermissionForm},
	"cancel_message_move_task":        {name: "cancel_message_move_task", method: "POST", path: "SQS.CancelMessageMoveTask", kind: "custom", required: []string{"task_handle"}, allowed: []string{"task_handle"}, redact: []string{"task_handle"}, risk: "cancels an in-flight dead-letter-queue message move task", confirm: "destructive", service: true, execute: buildCancelMessageMoveTaskForm},
	"change_message_visibility":       {name: "change_message_visibility", method: "POST", path: "SQS.ChangeMessageVisibility", kind: "update", required: []string{"receipt_handle", "visibility_timeout"}, allowed: []string{"receipt_handle", "visibility_timeout"}, redact: []string{"receipt_handle"}, risk: "changes the visibility timeout for one in-flight message", queue: true, execute: buildChangeMessageVisibilityForm},
	"change_message_visibility_batch": {name: "change_message_visibility_batch", method: "POST", path: "SQS.ChangeMessageVisibilityBatch", kind: "update", required: []string{"receipt_handle", "visibility_timeout"}, allowed: []string{"id", "receipt_handle", "visibility_timeout"}, redact: []string{"receipt_handle"}, risk: "changes visibility timeout for up to 10 in-flight messages per SQS batch request", batch: true, queue: true, execute: buildChangeMessageVisibilityBatchEntry},
	"create_queue":                    {name: "create_queue", method: "POST", path: "SQS.CreateQueue", kind: "create", required: []string{"queue_name"}, allowed: []string{"queue_name", "attributes", "tags"}, redact: []string{"attributes"}, risk: "creates an SQS queue; SQS returns an existing queue URL when name and attributes match", service: true, execute: buildCreateQueueForm},
	"delete_message":                  {name: "delete_message", method: "POST", path: "SQS.DeleteMessage", kind: "delete", required: []string{"receipt_handle"}, allowed: []string{"receipt_handle"}, redact: []string{"receipt_handle"}, risk: "deletes one received message by receipt handle", confirm: "destructive", queue: true, execute: buildDeleteMessageForm},
	"delete_message_batch":            {name: "delete_message_batch", method: "POST", path: "SQS.DeleteMessageBatch", kind: "delete", required: []string{"receipt_handle"}, allowed: []string{"id", "receipt_handle"}, redact: []string{"receipt_handle"}, risk: "deletes up to 10 received messages per SQS batch request", confirm: "destructive", batch: true, queue: true, execute: buildDeleteMessageBatchEntry},
	"delete_queue":                    {name: "delete_queue", method: "POST", path: "SQS.DeleteQueue", kind: "delete", allowed: []string{}, risk: "deletes the configured SQS queue", confirm: "destructive", queue: true, execute: buildNoopForm},
	"purge_queue":                     {name: "purge_queue", method: "POST", path: "SQS.PurgeQueue", kind: "delete", allowed: []string{}, risk: "purges all available messages from the configured queue", confirm: "destructive", queue: true, execute: buildNoopForm},
	"remove_permission":               {name: "remove_permission", method: "POST", path: "SQS.RemovePermission", kind: "delete", required: []string{"label"}, allowed: []string{"label"}, risk: "removes an SQS queue resource policy permission statement", confirm: "destructive", queue: true, execute: buildRemovePermissionForm},
	"send_message":                    {name: "send_message", method: "POST", path: "SQS.SendMessage", kind: "create", required: []string{"message_body"}, allowed: []string{"message_body", "delay_seconds", "message_attributes", "message_system_attributes", "message_deduplication_id", "message_group_id"}, redact: []string{"message_body", "message_attributes", "message_system_attributes"}, risk: "sends one message to the configured queue; FIFO queues may use message_deduplication_id for provider-supported idempotency", queue: true, execute: buildSendMessageForm},
	"send_message_batch":              {name: "send_message_batch", method: "POST", path: "SQS.SendMessageBatch", kind: "create", required: []string{"message_body"}, allowed: []string{"id", "message_body", "delay_seconds", "message_attributes", "message_system_attributes", "message_deduplication_id", "message_group_id"}, redact: []string{"message_body", "message_attributes", "message_system_attributes"}, risk: "sends up to 10 messages per SQS batch request; FIFO queues may use message_deduplication_id for provider-supported idempotency", batch: true, queue: true, execute: buildSendMessageBatchEntry},
	"set_queue_attributes":            {name: "set_queue_attributes", method: "POST", path: "SQS.SetQueueAttributes", kind: "update", requiredAny: [][]string{{"attributes"}, {"attribute_name", "attribute_value"}}, allowed: []string{"attribute_name", "attribute_value", "attributes"}, redact: []string{"attribute_value", "attributes"}, risk: "sets typed SQS queue attributes such as policy, redrive, encryption, retention, and visibility settings", queue: true, execute: buildSetQueueAttributesForm},
	"start_message_move_task":         {name: "start_message_move_task", method: "POST", path: "SQS.StartMessageMoveTask", kind: "custom", required: []string{"source_arn"}, allowed: []string{"source_arn", "destination_arn", "max_number_of_messages_per_second"}, risk: "starts an SQS dead-letter queue redrive message move task", service: true, execute: buildStartMessageMoveTaskForm},
	"tag_queue":                       {name: "tag_queue", method: "POST", path: "SQS.TagQueue", kind: "update", requiredAny: [][]string{{"tags"}, {"tag_key", "tag_value"}}, allowed: []string{"tag_key", "tag_value", "tags"}, risk: "adds or updates tags on the configured SQS queue", queue: true, execute: buildTagQueueForm},
	"untag_queue":                     {name: "untag_queue", method: "POST", path: "SQS.UntagQueue", kind: "delete", required: []string{"tag_keys"}, allowed: []string{"tag_keys"}, risk: "removes tags from the configured SQS queue", confirm: "destructive", queue: true, execute: buildUntagQueueForm},
}

func (c Connector) Manifest() connectors.Manifest {
	base := c.Definition()
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
		actions = append(actions, connectors.WriteActionSpec{Name: def.name, RequiredFields: append([]string(nil), def.required...), RequiredAnyFields: documentedRequiredAnyFields(def), OptionalFields: optionalFields(def), Method: def.method, Path: def.path, RedactFields: append([]string(nil), def.redact...), Risk: def.risk, Confirm: def.confirm})
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

// documentedRequiredAnyFields projects a write action's either/or requirement
// onto the manifest.
//
// The manifest previously carried def.required alone, which silently dropped
// requiredAny — so a generated manual listed set_queue_attributes's
// attribute_name/attribute_value/attributes as merely optional while
// validateSQSRequiredFields rejects a write that supplies none of them. The
// documented contract understated the enforced one, which is the same class of
// defect as a read that reports a completeness it does not have.
//
// It projects the GROUPS, not a rendered sentence: WriteActionSpec.RequiredFields
// is a list of field names that `pm connectors inspect --json` publishes, and
// the renderer composes the prose from the groups instead.
func documentedRequiredAnyFields(def writeActionDef) [][]string {
	if len(def.requiredAny) == 0 {
		return nil
	}
	out := make([][]string, 0, len(def.requiredAny))
	for _, group := range def.requiredAny {
		out = append(out, append([]string(nil), group...))
	}
	return out
}

func optionalFields(def writeActionDef) []string {
	required := map[string]bool{}
	for _, field := range def.required {
		required[field] = true
	}
	// A field named by any requiredAny group is part of the required contract,
	// not an optional extra; listing it as optional contradicts the required
	// line rendered from the same definition.
	for _, group := range def.requiredAny {
		for _, field := range group {
			required[field] = true
		}
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
	prepared, err := c.prepareSQSWrite(ctx, req, records)
	if err != nil {
		return connectors.WritePreview{}, err
	}
	return engine.PreviewPreparedWrite(prepared.shared)
}

func (c Connector) writeSQS(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	prepared, err := c.prepareSQSWrite(ctx, req, records)
	if err != nil {
		return connectors.WriteResult{RecordsFailed: len(records)}, err
	}
	preview, err := engine.PreviewPreparedWrite(prepared.shared)
	if err != nil {
		return connectors.WriteResult{RecordsFailed: len(records)}, err
	}
	var result connectors.WriteResult
	err = engine.ExecutePreparedWrite(ctx, prepared.shared, req.Approval, preview.Digest, func(executeCtx context.Context) error {
		var executeErr error
		result, executeErr = c.executePreparedSQSWrite(executeCtx, prepared)
		return executeErr
	})
	if err != nil && result.RecordsWritten == 0 && result.RecordsFailed == 0 {
		result.RecordsFailed = len(prepared.normalized)
	}
	return result, err
}

type preparedSQSRequest struct {
	endpoint string
	form     url.Values
	entryIDs []string
	count    int
}

type preparedSQSWrite struct {
	shared     engine.PreparedWrite
	definition writeActionDef
	connection sqsConfig
	normalized []connectors.Record
	requests   []preparedSQSRequest
}

type sqsWriteDefinition struct {
	Name        string     `json:"name"`
	Method      string     `json:"method"`
	Path        string     `json:"path"`
	Kind        string     `json:"kind"`
	Required    []string   `json:"required,omitempty"`
	RequiredAny [][]string `json:"required_any,omitempty"`
	Allowed     []string   `json:"allowed,omitempty"`
	Redact      []string   `json:"redact,omitempty"`
	Risk        string     `json:"risk,omitempty"`
	Confirm     string     `json:"confirm,omitempty"`
	Batch       bool       `json:"batch"`
	Queue       bool       `json:"queue"`
	Service     bool       `json:"service"`
}

func (c Connector) prepareSQSWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (preparedSQSWrite, error) {
	if err := ctx.Err(); err != nil {
		return preparedSQSWrite{}, err
	}
	def, action, err := lookupSQSWriteAction(req.Action)
	if err != nil {
		return preparedSQSWrite{}, err
	}
	normalized := normalizedWriteRecords(records)
	if err := validateSQSRecords(def, normalized); err != nil {
		return preparedSQSWrite{}, err
	}
	connection, err := resolveConnConfig(req.Config)
	if err != nil {
		return preparedSQSWrite{}, err
	}
	requests := make([]preparedSQSRequest, 0, len(normalized))
	if def.batch {
		for start := 0; start < len(normalized); start += sqsBatchLimit {
			end := min(start+sqsBatchLimit, len(normalized))
			chunk := normalized[start:end]
			form := baseActionForm(actionName(def.path))
			form.Set("QueueUrl", connection.queueURL)
			entryIDs := make([]string, 0, len(chunk))
			for index, record := range chunk {
				entryIDs = append(entryIDs, entryID(record, index+1))
				if err := def.execute(form, record, index+1); err != nil {
					return preparedSQSWrite{}, err
				}
			}
			requests = append(requests, preparedSQSRequest{endpoint: connection.endpointURL, form: form, entryIDs: entryIDs, count: len(chunk)})
		}
	} else {
		for index, record := range normalized {
			form := baseActionForm(actionName(def.path))
			if def.queue {
				form.Set("QueueUrl", connection.queueURL)
			}
			if err := def.execute(form, record, index+1); err != nil {
				return preparedSQSWrite{}, err
			}
			requests = append(requests, preparedSQSRequest{endpoint: connection.endpointURL, form: form, count: 1})
		}
	}
	canonical := make([]engine.PreparedRequest, 0, len(requests))
	for _, request := range requests {
		target := request.endpoint
		if queueURL := request.form.Get("QueueUrl"); queueURL != "" {
			target = queueURL
		}
		canonical = append(canonical, engine.PreparedRequest{
			Method:      http.MethodPost,
			URL:         request.endpoint,
			Target:      target,
			ContentType: "application/x-www-form-urlencoded",
			BodyFormat:  "form",
			Body:        request.form.Encode(),
			Headers: map[string]string{
				"Accept":     "text/xml",
				"User-Agent": userAgent,
				"Signing":    "aws-sigv4/sqs/" + connection.region,
			},
		})
	}
	warnings := []string{"amazon-sqs writes require reverse ETL plan -> preview -> explicit approval -> execute"}
	if def.confirm == "destructive" {
		warnings = append(warnings, "destructive confirmation required")
	}
	confirmation := connectors.ConfirmationKind(strings.TrimSpace(def.confirm))
	return preparedSQSWrite{
		shared: engine.PreparedWrite{
			Target: engine.DestructiveTarget{
				Connector:     c.Name(),
				Operation:     action,
				Method:        def.method,
				MutationClass: def.kind,
				Confirmation:  confirmation,
			},
			CredentialRevision:  req.Config.CredentialRevision,
			ConfigurationDigest: req.Config.ConfigurationDigest,
			ApprovalScope:       req.Config.WriteApprovalScope,
			Batchable:           true,
			RecordsStaged:       len(normalized),
			Action:              action,
			Warnings:            warnings,
			Definition: sqsWriteDefinition{
				Name: def.name, Method: def.method, Path: def.path, Kind: def.kind,
				Required: append([]string(nil), def.required...), RequiredAny: append([][]string(nil), def.requiredAny...),
				Allowed: append([]string(nil), def.allowed...), Redact: append([]string(nil), def.redact...),
				Risk: def.risk, Confirm: def.confirm, Batch: def.batch, Queue: def.queue, Service: def.service,
			},
			HookIdentity: "amazon-sqs-query-api-v1",
			Requests:     canonical,
		},
		definition: def,
		connection: connection,
		normalized: normalized,
		requests:   requests,
	}, nil
}

func (c Connector) executePreparedSQSWrite(ctx context.Context, prepared preparedSQSWrite) (connectors.WriteResult, error) {
	written := 0
	if prepared.definition.batch {
		for _, request := range prepared.requests {
			resp, err := c.doEndpoint(ctx, prepared.connection, request.endpoint, cloneValues(request.form), 16<<20)
			if err != nil {
				return connectors.WriteResult{RecordsWritten: written, RecordsFailed: len(prepared.normalized) - written}, err
			}
			batchIDs, err := parseBatchResultIDs(resp.body)
			if err != nil {
				return connectors.WriteResult{RecordsWritten: written, RecordsFailed: len(prepared.normalized) - written}, fmt.Errorf("amazon-sqs action %s batch response parse failed: %w", prepared.definition.name, err)
			}
			successes, failures, err := verifyBatchResultIDs(batchIDs, request.entryIDs)
			if err != nil {
				return connectors.WriteResult{RecordsWritten: written, RecordsFailed: len(prepared.normalized) - written}, fmt.Errorf("amazon-sqs action %s batch response %w", prepared.definition.name, err)
			}
			written += successes
			if failures > 0 {
				return connectors.WriteResult{RecordsWritten: written, RecordsFailed: len(prepared.normalized) - written}, fmt.Errorf("amazon-sqs action %s batch response reported %d failed entries", prepared.definition.name, failures)
			}
		}
		return connectors.WriteResult{RecordsWritten: written}, nil
	}
	for _, request := range prepared.requests {
		_, err := c.doEndpoint(ctx, prepared.connection, request.endpoint, cloneValues(request.form), 16<<20)
		if err != nil {
			return connectors.WriteResult{RecordsWritten: written, RecordsFailed: len(prepared.normalized) - written}, err
		}
		written += request.count
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

type sqsRecordFieldType string

const (
	sqsFieldString              sqsRecordFieldType = "string"
	sqsFieldInteger             sqsRecordFieldType = "integer"
	sqsFieldStringList          sqsRecordFieldType = "string_list"
	sqsFieldStringMap           sqsRecordFieldType = "string_map"
	sqsFieldMessageAttributeMap sqsRecordFieldType = "message_attribute_map"
)

type sqsIntegerRange struct {
	min int
	max int
}

var sqsRecordFieldTypes = map[string]sqsRecordFieldType{
	"actions":                           sqsFieldStringList,
	"attribute_name":                    sqsFieldString,
	"attribute_value":                   sqsFieldString,
	"attributes":                        sqsFieldStringMap,
	"aws_account_ids":                   sqsFieldStringList,
	"delay_seconds":                     sqsFieldInteger,
	"destination_arn":                   sqsFieldString,
	"id":                                sqsFieldString,
	"label":                             sqsFieldString,
	"max_number_of_messages_per_second": sqsFieldInteger,
	"message_attributes":                sqsFieldMessageAttributeMap,
	"message_body":                      sqsFieldString,
	"message_deduplication_id":          sqsFieldString,
	"message_group_id":                  sqsFieldString,
	"message_system_attributes":         sqsFieldMessageAttributeMap,
	"queue_name":                        sqsFieldString,
	"receipt_handle":                    sqsFieldString,
	"source_arn":                        sqsFieldString,
	"tag_key":                           sqsFieldString,
	"tag_keys":                          sqsFieldStringList,
	"tag_value":                         sqsFieldString,
	"tags":                              sqsFieldStringMap,
	"task_handle":                       sqsFieldString,
	"visibility_timeout":                sqsFieldInteger,
}

var sqsIntegerFieldRanges = map[string]sqsIntegerRange{
	"delay_seconds":                     {min: 0, max: 900},
	"max_number_of_messages_per_second": {min: 1, max: 500},
	"visibility_timeout":                {min: 0, max: 43200},
}

var sqsMessageAttributeFields = map[string]bool{
	"binary_list_values": true,
	"binary_value":       true,
	"data_type":          true,
	"string_list_values": true,
	"string_value":       true,
}

var maxNativeInt = int64(^uint(0) >> 1)
var minNativeInt = -maxNativeInt - 1

func validateSQSRecords(def writeActionDef, records []connectors.Record) error {
	allowed := map[string]bool{}
	for _, field := range def.allowed {
		allowed[field] = true
	}
	for i, rec := range records {
		if err := validateSQSRequiredFields(def, rec); err != nil {
			return fmt.Errorf("amazon-sqs action %s record %d %w", def.name, i, err)
		}
		for field, value := range rec {
			if !allowed[field] {
				return fmt.Errorf("amazon-sqs action %s record %d has unsupported field %q", def.name, i, field)
			}
			if err := validateSQSFieldValue(field, value); err != nil {
				return fmt.Errorf("amazon-sqs action %s record %d field %q %w", def.name, i, field, err)
			}
		}
	}
	return nil
}

func validateSQSRequiredFields(def writeActionDef, rec connectors.Record) error {
	for _, field := range def.required {
		if isEmptyRequiredRecordValue(field, rec[field]) {
			return fmt.Errorf("requires field %q", field)
		}
	}
	if len(def.requiredAny) == 0 {
		return nil
	}
	anyComplete := false
	for _, group := range def.requiredAny {
		present := 0
		missing := 0
		for _, field := range group {
			if isEmptyRequiredRecordValue(field, rec[field]) {
				missing++
			} else {
				present++
			}
		}
		switch {
		case missing == 0:
			anyComplete = true
		case present > 0:
			return fmt.Errorf("requires fields %s together", formatRequiredGroup(group))
		}
	}
	if anyComplete {
		return nil
	}
	return fmt.Errorf("requires one of %s", formatRequiredAny(def.requiredAny))
}

func isEmptyRequiredRecordValue(field string, value any) bool {
	if field == "message_body" {
		text, ok := value.(string)
		if ok {
			return text == ""
		}
		return value == nil
	}
	return isEmptyRecordValue(value)
}

func formatRequiredAny(groups [][]string) string {
	parts := make([]string, 0, len(groups))
	for _, group := range groups {
		parts = append(parts, formatRequiredGroup(group))
	}
	return strings.Join(parts, " or ")
}

func formatRequiredGroup(group []string) string {
	quoted := make([]string, 0, len(group))
	for _, field := range group {
		quoted = append(quoted, strconv.Quote(field))
	}
	return strings.Join(quoted, " + ")
}

func validateSQSFieldValue(field string, value any) error {
	fieldType, ok := sqsRecordFieldTypes[field]
	if !ok {
		return errors.New("has no declared schema type")
	}
	switch fieldType {
	case sqsFieldString:
		if _, ok := value.(string); !ok {
			return errors.New("must be string")
		}
	case sqsFieldInteger:
		n, ok := sqsIntegerValue(value)
		if !ok {
			return errors.New("must be integer")
		}
		if r, ok := sqsIntegerFieldRanges[field]; ok && (n < r.min || n > r.max) {
			return fmt.Errorf("must be between %d and %d", r.min, r.max)
		}
	case sqsFieldStringList:
		if err := validateStringListValue(value); err != nil {
			return err
		}
	case sqsFieldStringMap:
		if err := validateStringMapValue(value); err != nil {
			return err
		}
	case sqsFieldMessageAttributeMap:
		if err := validateMessageAttributeMapValue(value, field == "message_system_attributes"); err != nil {
			return err
		}
	}
	return nil
}

func validateStringListValue(value any) error {
	switch typed := value.(type) {
	case []string:
		return nil
	case []any:
		for _, item := range typed {
			if _, ok := item.(string); !ok {
				return errors.New("must be array of strings")
			}
		}
		return nil
	default:
		return errors.New("must be array of strings")
	}
}

func validateStringMapValue(value any) error {
	items, ok := objectItems(value)
	if !ok {
		return errors.New("must be object with string values")
	}
	for _, item := range items {
		if _, ok := item.(string); !ok {
			return errors.New("must be object with string values")
		}
	}
	return nil
}

func validateMessageAttributeMapValue(value any, system bool) error {
	items, ok := objectItems(value)
	if !ok {
		return errors.New("must be object")
	}
	maxAttributes := 10
	if system {
		maxAttributes = 1
	}
	if len(items) > maxAttributes {
		return fmt.Errorf("must contain at most %d attributes", maxAttributes)
	}
	for name, raw := range items {
		if err := validateMessageAttributeName(name, system); err != nil {
			return err
		}
		if value, ok := raw.(string); ok {
			if value == "" {
				return fmt.Errorf("attribute %q value must not be empty", name)
			}
			if system {
				if err := validateSQSXRayTraceHeader(value); err != nil {
					return fmt.Errorf("system attribute %q %w", name, err)
				}
			}
			continue
		}
		attr, ok := objectItems(raw)
		if !ok {
			return fmt.Errorf("attribute %q must be string or object", name)
		}
		for field, item := range attr {
			if !sqsMessageAttributeFields[field] {
				return fmt.Errorf("attribute %q has unsupported field %q", name, field)
			}
			switch field {
			case "data_type", "string_value", "binary_value":
				if _, ok := item.(string); !ok {
					return fmt.Errorf("attribute %q field %q must be string", name, field)
				}
			case "string_list_values", "binary_list_values":
				return fmt.Errorf("attribute %q field %q is reserved and unsupported", name, field)
			}
		}
		dataType, _ := attr["data_type"].(string)
		if dataType == "" {
			return fmt.Errorf("attribute %q data_type must not be empty", name)
		}
		baseType, err := validateMessageAttributeDataType(dataType)
		if err != nil {
			return fmt.Errorf("attribute %q %w", name, err)
		}
		if system && dataType != "String" {
			return fmt.Errorf("system attribute %q must use data_type String", name)
		}
		stringValue, hasString := attr["string_value"].(string)
		binaryValue, hasBinary := attr["binary_value"].(string)
		switch baseType {
		case "String":
			if !hasString || stringValue == "" {
				return fmt.Errorf("attribute %q data_type %s requires non-empty string_value", name, dataType)
			}
			if hasBinary {
				return fmt.Errorf("attribute %q data_type %s does not allow binary_value", name, dataType)
			}
			if system {
				if err := validateSQSXRayTraceHeader(stringValue); err != nil {
					return fmt.Errorf("system attribute %q %w", name, err)
				}
			}
		case "Number":
			if !hasString || stringValue == "" {
				return fmt.Errorf("attribute %q data_type %s requires non-empty string_value", name, dataType)
			}
			if hasBinary {
				return fmt.Errorf("attribute %q data_type %s does not allow binary_value", name, dataType)
			}
			if err := validateSQSNumberAttributeValue(stringValue); err != nil {
				return fmt.Errorf("attribute %q %w", name, err)
			}
		case "Binary":
			if !hasBinary || binaryValue == "" {
				return fmt.Errorf("attribute %q data_type %s requires non-empty binary_value", name, dataType)
			}
			if hasString {
				return fmt.Errorf("attribute %q data_type %s does not allow string_value", name, dataType)
			}
			if err := validateSQSBinaryAttributeValue(binaryValue); err != nil {
				return fmt.Errorf("attribute %q %w", name, err)
			}
		}
	}
	return nil
}

func validateMessageAttributeName(name string, system bool) error {
	if system {
		if name != "AWSTraceHeader" {
			return fmt.Errorf("message system attributes only supports AWSTraceHeader, got %q", name)
		}
		return nil
	}
	invalid := name == "" || len(name) > 256 || name[0] == '.' || name[len(name)-1] == '.' || strings.Contains(name, "..")
	lower := strings.ToLower(name)
	invalid = invalid || strings.HasPrefix(lower, "aws.") || strings.HasPrefix(lower, "amazon.")
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.' {
			continue
		}
		invalid = true
		break
	}
	if invalid {
		return fmt.Errorf("message attribute has invalid name %q", name)
	}
	return nil
}

func validateMessageAttributeDataType(dataType string) (string, error) {
	if len(dataType) > 256 {
		return "", errors.New("data_type must be at most 256 characters")
	}
	baseType, suffix, custom := strings.Cut(dataType, ".")
	if baseType != "String" && baseType != "Number" && baseType != "Binary" {
		return "", fmt.Errorf("has unsupported data_type %q", dataType)
	}
	if custom && suffix == "" {
		return "", fmt.Errorf("has invalid data_type %q", dataType)
	}
	return baseType, nil
}

func objectItems(value any) (map[string]any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		return typed, true
	case connectors.Record:
		return map[string]any(typed), true
	case map[string]string:
		out := make(map[string]any, len(typed))
		for key, value := range typed {
			out[key] = value
		}
		return out, true
	default:
		return nil, false
	}
}

func sqsIntegerValue(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int8:
		return int(typed), true
	case int16:
		return int(typed), true
	case int32:
		return int(typed), true
	case int64:
		return int64ToInt(typed)
	case uint:
		return uint64ToInt(uint64(typed))
	case uint8:
		return int(typed), true
	case uint16:
		return int(typed), true
	case uint32:
		return uint64ToInt(uint64(typed))
	case uint64:
		return uint64ToInt(typed)
	case float32:
		return floatToInt(float64(typed))
	case float64:
		return floatToInt(typed)
	case json.Number:
		return jsonNumberToInt(typed)
	case jsonNumber:
		return jsonNumberToInt(json.Number(typed))
	default:
		return 0, false
	}
}

func int64ToInt(n int64) (int, bool) {
	if n < minNativeInt || n > maxNativeInt {
		return 0, false
	}
	return int(n), true
}

func uint64ToInt(n uint64) (int, bool) {
	if n > uint64(maxNativeInt) {
		return 0, false
	}
	return int(n), true
}

func floatToInt(n float64) (int, bool) {
	if math.IsNaN(n) || math.IsInf(n, 0) || math.Trunc(n) != n || n < float64(minNativeInt) || n > float64(maxNativeInt) {
		return 0, false
	}
	return int(n), true
}

func jsonNumberToInt(n json.Number) (int, bool) {
	if parsed, err := n.Int64(); err == nil {
		return int64ToInt(parsed)
	}
	parsed, err := n.Float64()
	if err != nil {
		return 0, false
	}
	return floatToInt(parsed)
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
	form.Set("VisibilityTimeout", strconv.Itoa(intField(rec, "visibility_timeout", 0, 0, 43200)))
	return nil
}

func buildChangeMessageVisibilityBatchEntry(form url.Values, rec connectors.Record, index int) error {
	prefix := fmt.Sprintf("ChangeMessageVisibilityBatchRequestEntry.%d.", index)
	form.Set(prefix+"Id", entryID(rec, index))
	form.Set(prefix+"ReceiptHandle", stringField(rec, "receipt_handle"))
	form.Set(prefix+"VisibilityTimeout", strconv.Itoa(intField(rec, "visibility_timeout", 0, 0, 43200)))
	return nil
}

func buildCreateQueueForm(form url.Values, rec connectors.Record, _ int) error {
	form.Set("QueueName", stringField(rec, "queue_name"))
	addStringMap(form, "Attribute", stringMapField(rec, "attributes"))
	addTagMap(form, stringMapField(rec, "tags"))
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
	form.Set("MessageBody", exactStringField(rec, "message_body"))
	addOptionalInt(form, "DelaySeconds", rec, "delay_seconds", 0, 0, 900)
	addOptionalString(form, "MessageDeduplicationId", rec, "message_deduplication_id")
	addOptionalString(form, "MessageGroupId", rec, "message_group_id")
	addMessageAttributeMap(form, "MessageAttribute", messageAttributeMapField(rec, "message_attributes"))
	addMessageAttributeMap(form, "MessageSystemAttribute", messageAttributeMapField(rec, "message_system_attributes"))
	return nil
}

func buildSendMessageBatchEntry(form url.Values, rec connectors.Record, index int) error {
	prefix := fmt.Sprintf("SendMessageBatchRequestEntry.%d.", index)
	form.Set(prefix+"Id", entryID(rec, index))
	form.Set(prefix+"MessageBody", exactStringField(rec, "message_body"))
	addOptionalInt(form, prefix+"DelaySeconds", rec, "delay_seconds", 0, 0, 900)
	addOptionalString(form, prefix+"MessageDeduplicationId", rec, "message_deduplication_id")
	addOptionalString(form, prefix+"MessageGroupId", rec, "message_group_id")
	addMessageAttributeMap(form, prefix+"MessageAttribute", messageAttributeMapField(rec, "message_attributes"))
	addMessageAttributeMap(form, prefix+"MessageSystemAttribute", messageAttributeMapField(rec, "message_system_attributes"))
	return nil
}

func buildSetQueueAttributesForm(form url.Values, rec connectors.Record, _ int) error {
	attrs := stringMapField(rec, "attributes")
	if attrs == nil {
		attrs = map[string]string{}
	}
	if !isEmptyRecordValue(rec["attribute_name"]) && !isEmptyRecordValue(rec["attribute_value"]) {
		attrs[stringField(rec, "attribute_name")] = stringField(rec, "attribute_value")
	}
	addStringMap(form, "Attribute", attrs)
	return nil
}

func buildStartMessageMoveTaskForm(form url.Values, rec connectors.Record, _ int) error {
	form.Set("SourceArn", stringField(rec, "source_arn"))
	addOptionalString(form, "DestinationArn", rec, "destination_arn")
	addOptionalInt(form, "MaxNumberOfMessagesPerSecond", rec, "max_number_of_messages_per_second", 0, 1, 500)
	return nil
}

func buildTagQueueForm(form url.Values, rec connectors.Record, _ int) error {
	tags := stringMapField(rec, "tags")
	if tags == nil {
		tags = map[string]string{}
	}
	if !isEmptyRecordValue(rec["tag_key"]) && !isEmptyRecordValue(rec["tag_value"]) {
		tags[stringField(rec, "tag_key")] = stringField(rec, "tag_value")
	}
	addTagMap(form, tags)
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

func addOptionalInt(form url.Values, name string, rec connectors.Record, field string, def, min, max int) {
	if isEmptyRecordValue(rec[field]) {
		return
	}
	form.Set(name, strconv.Itoa(intField(rec, field, def, min, max)))
}

func entryID(rec connectors.Record, index int) string {
	if id := stringField(rec, "id"); id != "" {
		return id
	}
	return fmt.Sprintf("entry_%d", index)
}

func stringField(rec connectors.Record, key string) string {
	value, ok := rec[key]
	if !ok || value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func exactStringField(rec connectors.Record, key string) string {
	value, ok := rec[key].(string)
	if !ok {
		return ""
	}
	return value
}

func intField(rec connectors.Record, key string, def, min, max int) int {
	value, ok := rec[key]
	if !ok || value == nil {
		return def
	}
	n, ok := sqsIntegerValue(value)
	if !ok {
		return def
	}
	if n < min {
		return min
	}
	if max > min && n > max {
		return max
	}
	return n
}

type jsonNumber string

func stringSliceField(rec connectors.Record, key string) []string {
	value := rec[key]
	switch typed := value.(type) {
	case []string:
		return compactStrings(typed)
	case []any:
		out := make([]string, 0, len(typed))
		for _, v := range typed {
			if s, ok := v.(string); ok {
				out = append(out, s)
			}
		}
		return compactStrings(out)
	case string:
		return splitCSV(typed)
	default:
		return nil
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
			if s, ok := v.(string); ok {
				out[k] = s
			}
		}
		return out
	case connectors.Record:
		out := map[string]string{}
		for k, v := range typed {
			if s, ok := v.(string); ok {
				out[k] = s
			}
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
	DataType    string
	StringValue string
	BinaryValue string
}

func messageAttributeMapField(rec connectors.Record, key string) map[string]messageAttributeValue {
	items, ok := objectItems(rec[key])
	if !ok {
		return nil
	}
	out := map[string]messageAttributeValue{}
	for name, raw := range items {
		if s, ok := raw.(string); ok {
			out[name] = messageAttributeValue{DataType: "String", StringValue: s}
			continue
		}
		m, ok := objectItems(raw)
		if !ok {
			continue
		}
		out[name] = messageAttributeValue{
			DataType:    stringMapAnyField(m, "data_type"),
			StringValue: stringMapAnyField(m, "string_value"),
			BinaryValue: stringMapAnyField(m, "binary_value"),
		}
	}
	return out
}

func stringMapAnyField(values map[string]any, key string) string {
	value, ok := values[key]
	if !ok || value == nil {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return text
}

func exactStringValues(value any) []string {
	switch typed := value.(type) {
	case []string:
		return append([]string(nil), typed...)
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
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

func addTagMap(form url.Values, values map[string]string) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for i, key := range keys {
		form.Set(fmt.Sprintf("Tag.%d.Key", i+1), key)
		form.Set(fmt.Sprintf("Tag.%d.Value", i+1), values[key])
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
	}
}

type sqsBatchResultIDs struct {
	successful []string
	failed     []string
}

func parseBatchResultIDs(raw []byte) (sqsBatchResultIDs, error) {
	decoder := xml.NewDecoder(bytes.NewReader(raw))
	var result sqsBatchResultIDs
	for {
		tok, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return sqsBatchResultIDs{}, fmt.Errorf("parse xml: %w", err)
		}
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch {
		case start.Name.Local == "BatchResultErrorEntry":
			id, err := decodeBatchResultID(decoder, start)
			if err != nil {
				return sqsBatchResultIDs{}, err
			}
			result.failed = append(result.failed, id)
		case strings.HasSuffix(start.Name.Local, "BatchResultEntry"):
			id, err := decodeBatchResultID(decoder, start)
			if err != nil {
				return sqsBatchResultIDs{}, err
			}
			result.successful = append(result.successful, id)
		}
	}
	return result, nil
}

func decodeBatchResultID(decoder *xml.Decoder, start xml.StartElement) (string, error) {
	var entry struct {
		ID string `xml:"Id"`
	}
	if err := decoder.DecodeElement(&entry, &start); err != nil {
		return "", fmt.Errorf("decode batch result entry: %w", err)
	}
	id := strings.TrimSpace(entry.ID)
	if id == "" {
		return "", errors.New("batch result entry missing id")
	}
	return id, nil
}

func verifyBatchResultIDs(result sqsBatchResultIDs, expected []string) (int, int, error) {
	expectedIDs := make(map[string]struct{}, len(expected))
	for _, id := range expected {
		if id == "" {
			return 0, 0, errors.New("batch request entry id is empty")
		}
		if _, ok := expectedIDs[id]; ok {
			return 0, 0, fmt.Errorf("duplicate batch request id %q", id)
		}
		expectedIDs[id] = struct{}{}
	}
	seen := make(map[string]struct{}, len(expected))
	for _, id := range result.successful {
		if err := verifyBatchResponseID(id, expectedIDs, seen); err != nil {
			return 0, 0, err
		}
	}
	for _, id := range result.failed {
		if err := verifyBatchResponseID(id, expectedIDs, seen); err != nil {
			return 0, 0, err
		}
	}
	if len(seen) != len(expectedIDs) {
		return 0, 0, fmt.Errorf("accounted for %d of %d entries", len(seen), len(expectedIDs))
	}
	return len(result.successful), len(result.failed), nil
}

func verifyBatchResponseID(id string, expectedIDs, seen map[string]struct{}) error {
	if _, ok := expectedIDs[id]; !ok {
		return fmt.Errorf("unknown batch response id %q", id)
	}
	if _, ok := seen[id]; ok {
		return fmt.Errorf("duplicate batch response id %q", id)
	}
	seen[id] = struct{}{}
	return nil
}
