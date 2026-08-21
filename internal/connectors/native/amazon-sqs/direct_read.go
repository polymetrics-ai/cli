package amazonsqs

import (
	"context"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

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
	maxBytes, err := c.OperationDirectReadMaxBytes(req.Operation, req.MaxBytes)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	form, err := directReadForm(req.Operation, req.Config, req.Body)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	if err := applySQSDirectReadPageRequest(req, form); err != nil {
		return connectors.DirectReadResult{}, err
	}
	result := connectors.DirectReadResult{Connector: c.Name(), Operation: req.Operation, Method: "POST", Path: "SQS." + directReadAction(req.Operation), OutputSecretFields: append([]string(nil), req.RedactFields...)}
	resp, err := c.doService(ctx, req.Config, form, maxBytes)
	result.Receipt = sqsDirectReadReceipt(resp)
	if result.Receipt != nil {
		result.Status = result.Receipt.Status
	}
	if err != nil {
		return result, err
	}
	body, err := decodeDirectReadOperation(req.Operation, resp.body)
	if err != nil {
		return result, err
	}
	result.Receipt.Body = body
	page := sqsDirectReadPage(req.Operation, body)
	redacted := redactDirectBody(body, req.RedactFields)
	body, ok := redacted.(map[string]any)
	if !ok {
		return result, fmt.Errorf("redact sqs direct read %s: unexpected root type %T", req.Operation, redacted)
	}
	result.Body = body
	result.Page = page
	return result, nil
}

func sqsDirectReadReceipt(response sqsHTTPResponse) *connectors.ProviderResponseReceipt {
	if response.status == 0 {
		return nil
	}
	receipt := &connectors.ProviderResponseReceipt{
		ResponseReceived: true,
		Status:           response.status,
		Headers:          make(map[string]connectors.OperationResponseHeader, len(response.headers)),
		BodyPresent:      len(response.body) != 0,
		BodyBytes:        int64(len(response.body)),
	}
	for name, values := range response.headers {
		receipt.Headers[name] = connectors.OperationResponseHeader{Values: append([]string(nil), values...)}
	}
	if len(response.body) != 0 {
		if utf8.Valid(response.body) {
			receipt.BodyRaw, receipt.BodyRawEncoding = string(response.body), "text"
		} else {
			receipt.BodyRaw, receipt.BodyRawEncoding = base64.StdEncoding.EncodeToString(response.body), "base64"
		}
	}
	return receipt
}

// sqsDirectReadPaging declares how each SQS read operation pages, so the page
// context this connector reports is derived from the AWS contract rather than
// guessed per call.
//
// The SQS Query API pages by an opaque NextToken, which is the "cursor" family:
// there is no addressable page number, so --page is refused rather than
// answered with page one. Operations absent from this table return a single
// object (GetQueueUrl, GetQueueAttributes, ListQueueTags) or a collection AWS
// never pages (ListMessageMoveTasks).
type sqsDirectReadPaging struct {
	// collection is the decoded body key holding the paged rows.
	collection string
	// tokenField is the decoded body key holding the next-page token; empty
	// means the operation has no next page to hand back.
	tokenField string
	// formField is the request form field the token is replayed through.
	formField string
}

var sqsDirectReadPagingSpecs = map[string]sqsDirectReadPaging{
	"list_queues":                    {collection: "queue_urls", tokenField: "next_token", formField: "NextToken"},
	"list_dead_letter_source_queues": {collection: "queue_urls", tokenField: "next_token", formField: "NextToken"},
	"list_message_move_tasks":        {collection: "results"},
}

// applySQSDirectReadPageRequest puts the caller's navigation input on the wire,
// or refuses it. Accepting --page 3 and returning page one is the wrongness
// this contract exists to remove, so an input this operation cannot honour is
// an error, never a silently discarded field.
func applySQSDirectReadPageRequest(req connectors.OperationDirectReadRequest, form url.Values) error {
	spec, paged := sqsDirectReadPagingSpecs[req.Operation]
	cursored := paged && spec.formField != ""
	switch {
	case req.Page > 1 && cursored:
		return fmt.Errorf("amazon-sqs direct read %s pages by an opaque NextToken and has no addressable page number; pass the previous page's next_cursor to --page-cursor instead", req.Operation)
	case req.Page > 1:
		return fmt.Errorf("amazon-sqs direct read %s returns a single page and cannot address page %d", req.Operation, req.Page)
	case req.PageCursor != "" && !cursored:
		return fmt.Errorf("amazon-sqs direct read %s returns a single page and cannot address a page by cursor", req.Operation)
	case req.PageCursor == "":
		return nil
	}
	if err := connectors.ValidateDirectReadPageCursor(req.PageCursor); err != nil {
		return err
	}
	// The operation also declares its own next_token flag. Two navigation
	// inputs select two different pages, so the pairing is refused rather than
	// resolved in favour of one.
	if form.Get(spec.formField) != "" {
		return fmt.Errorf("amazon-sqs direct read %s received --page-cursor and --next-token; they select different pages, so pass one of them", req.Operation)
	}
	form.Set(spec.formField, req.PageCursor)
	return nil
}

// sqsDirectReadPage reports where the returned page sits, from the decoded body
// AWS actually sent. It is read BEFORE redaction so a token the redactor would
// mask is still reported as the cursor the caller needs.
func sqsDirectReadPage(operation string, body map[string]any) connectors.DirectReadPage {
	spec, paged := sqsDirectReadPagingSpecs[operation]
	if !paged {
		// Not a collection: the single object is the whole answer.
		return connectors.DirectReadPage{Strategy: "none", Complete: true}
	}
	page := connectors.DirectReadPage{Strategy: "cursor"}
	switch rows := body[spec.collection].(type) {
	case []any:
		page.Records = len(rows)
	case []string:
		page.Records = len(rows)
	}
	if spec.formField == "" {
		// AWS pages this operation not at all, so one request is everything it
		// can offer — and that is not proof the collection ended.
		page.Strategy = "none"
		page.Reason = connectors.DirectReadPageReasonNoPagination
		return page
	}
	token, _ := body[spec.tokenField].(string)
	if strings.TrimSpace(token) == "" {
		page.Complete = true
		return page
	}
	page.HasMore = true
	page.NextCursor = token
	page.Reason = connectors.DirectReadPageReasonMorePages
	return page
}

// PreflightOperationDirectRead proves a command can reach one of the closed
// SQS Query API read operations without resolving config, signing a request,
// or issuing network I/O. The executor itself uses the same operation table,
// fixed POST method, action path, 16 MiB response ceiling, and redacted JSON
// output contract.
func (c Connector) PreflightOperationDirectRead(operation, method, path string, maxBytes int, outputPolicy string) error {
	if err := c.Base.PreflightOperationDirectRead(operation, method, path, maxBytes, outputPolicy); err != nil {
		return err
	}
	if _, ok := sqsDirectReadOperations[operation]; !ok {
		return fmt.Errorf("amazon-sqs direct read operation %q not found", operation)
	}
	if !strings.EqualFold(strings.TrimSpace(method), "POST") {
		return fmt.Errorf("amazon-sqs direct read method %q must be POST", method)
	}
	wantPath := "SQS." + directReadAction(operation)
	if path != wantPath {
		return fmt.Errorf("amazon-sqs direct read path %q does not match operation path %q", path, wantPath)
	}
	if maxBytes <= 0 {
		return fmt.Errorf("amazon-sqs direct read requires positive max_bytes")
	}
	if outputPolicy != "json_redacted" {
		return fmt.Errorf("amazon-sqs direct read output policy %q is not supported", outputPolicy)
	}
	return nil
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

type directReadFieldDef struct {
	name      string
	fieldType sqsRecordFieldType
	min       int
	max       int
}

type directReadOperationDef struct {
	required []string
	fields   []directReadFieldDef
}

var sqsDirectReadOperations = map[string]directReadOperationDef{
	"get_queue_attributes": {fields: []directReadFieldDef{{name: "attribute_names", fieldType: sqsFieldStringList}}},
	"get_queue_url": {required: []string{"queue_name"}, fields: []directReadFieldDef{
		{name: "queue_name", fieldType: sqsFieldString},
		{name: "queue_owner_aws_account_id", fieldType: sqsFieldString},
	}},
	"list_dead_letter_source_queues": {fields: []directReadFieldDef{
		{name: "next_token", fieldType: sqsFieldString},
		{name: "max_results", fieldType: sqsFieldInteger, min: 1, max: 1000},
	}},
	"list_message_move_tasks": {required: []string{"source_arn"}, fields: []directReadFieldDef{
		{name: "source_arn", fieldType: sqsFieldString},
		{name: "max_results", fieldType: sqsFieldInteger, min: 1, max: 10},
	}},
	"list_queues": {fields: []directReadFieldDef{
		{name: "queue_name_prefix", fieldType: sqsFieldString},
		{name: "next_token", fieldType: sqsFieldString},
		{name: "max_results", fieldType: sqsFieldInteger, min: 1, max: 1000},
	}},
	"list_queue_tags": {},
}

func directReadForm(operation string, cfg connectors.RuntimeConfig, body map[string]any) (url.Values, error) {
	if err := validateSQSDirectReadBody(operation, body); err != nil {
		return nil, err
	}
	form := baseActionForm(directReadAction(operation))
	switch operation {
	case "get_queue_attributes":
		addConfiguredQueueURL(form, cfg)
		attrs := exactStringValues(body["attribute_names"])
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
		if maxResults, ok := bodyInt(body, "max_results"); ok {
			form.Set("MaxResults", strconv.Itoa(maxResults))
		}
	case "list_message_move_tasks":
		sourceArn := bodyString(body, "source_arn")
		if sourceArn == "" {
			return nil, fmt.Errorf("amazon-sqs direct read %s requires source_arn", operation)
		}
		form.Set("SourceArn", sourceArn)
		if maxResults, ok := bodyInt(body, "max_results"); ok {
			form.Set("MaxResults", strconv.Itoa(maxResults))
		}
	case "list_queues":
		if prefix := bodyString(body, "queue_name_prefix"); prefix != "" {
			form.Set("QueueNamePrefix", prefix)
		}
		if token := bodyString(body, "next_token"); token != "" {
			form.Set("NextToken", token)
		}
		if maxResults, ok := bodyInt(body, "max_results"); ok {
			form.Set("MaxResults", strconv.Itoa(maxResults))
		}
	case "list_queue_tags":
		addConfiguredQueueURL(form, cfg)
	default:
		return nil, fmt.Errorf("amazon-sqs direct read operation %q not found", operation)
	}
	return form, nil
}

func validateSQSDirectReadBody(operation string, body map[string]any) error {
	def, ok := sqsDirectReadOperations[operation]
	if !ok {
		return fmt.Errorf("amazon-sqs direct read operation %q not found", operation)
	}
	for _, field := range def.required {
		if isEmptyRecordValue(body[field]) {
			return fmt.Errorf("amazon-sqs direct read %s requires %s", operation, field)
		}
	}
	allowed := map[string]directReadFieldDef{}
	for _, field := range def.fields {
		allowed[field.name] = field
	}
	keys := make([]string, 0, len(body))
	for field := range body {
		keys = append(keys, field)
	}
	sort.Strings(keys)
	for _, field := range keys {
		spec, ok := allowed[field]
		if !ok {
			return fmt.Errorf("amazon-sqs direct read %s body has unsupported field %q", operation, field)
		}
		if err := validateSQSDirectReadFieldValue(spec, body[field]); err != nil {
			return fmt.Errorf("amazon-sqs direct read %s field %q %w", operation, field, err)
		}
	}
	return nil
}

func validateSQSDirectReadFieldValue(spec directReadFieldDef, value any) error {
	switch spec.fieldType {
	case sqsFieldString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("must be string")
		}
	case sqsFieldInteger:
		n, ok := sqsIntegerValue(value)
		if !ok {
			return fmt.Errorf("must be integer")
		}
		if spec.max > 0 && (n < spec.min || n > spec.max) {
			return fmt.Errorf("must be between %d and %d", spec.min, spec.max)
		}
	case sqsFieldStringList:
		return validateStringListValue(value)
	default:
		return fmt.Errorf("has no declared schema type")
	}
	return nil
}

func bodyString(body map[string]any, key string) string {
	if body == nil || body[key] == nil {
		return ""
	}
	value, ok := body[key].(string)
	if !ok {
		return ""
	}
	return value
}

func bodyInt(body map[string]any, key string) (int, bool) {
	if body == nil || body[key] == nil {
		return 0, false
	}
	return sqsIntegerValue(body[key])
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
	return redactDirectValue("", value, explicit, 0)
}

func redactDirectValue(key string, value any, explicit map[string]bool, depth int) any {
	if key != "" && shouldRedactDirectField(key, explicit, depth == 1) {
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
			out[k] = redactDirectValue(k, typed[k], explicit, depth+1)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = redactDirectValue("", item, explicit, depth+1)
		}
		return out
	default:
		return value
	}
}

func shouldRedactDirectField(key string, explicit map[string]bool, topLevel bool) bool {
	normalized := normalizeField(key)
	if explicit[normalized] {
		return true
	}
	if normalized == "next_token" && topLevel {
		return false
	}
	for _, marker := range []string{"receipt_handle", "task_handle", "policy", "secret", "token", "password", "api_key", "apikey", "access_key", "accesskey", "credential"} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func normalizeField(key string) string {
	return strings.ToLower(strings.NewReplacer("-", "_", ".", "_", " ", "_").Replace(key))
}
