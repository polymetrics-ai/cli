package dynamodb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

// scanRequest is the DynamoDB_20120810.Scan JSON-RPC request body, ported
// from legacy dynamodb.go and kept for the focused Scan regression tests.
type scanRequest struct {
	TableName         string                    `json:"TableName"`
	Limit             int                       `json:"Limit,omitempty"`
	ExclusiveStartKey map[string]attributeValue `json:"ExclusiveStartKey,omitempty"`
}

// scanResponse is the DynamoDB_20120810.Scan JSON-RPC response body.
type scanResponse struct {
	Items            []map[string]attributeValue `json:"Items"`
	LastEvaluatedKey map[string]attributeValue   `json:"LastEvaluatedKey"`
}

const defaultJSONResponseMaxBytes = 64 << 20

// InitialState satisfies connectors.StatefulReader. Native DynamoDB reads are
// bounded full reads; state is used only as an optional in-call cursor seed for
// callers that carry one, so the initial cursor is empty.
func (c Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return connsdk.WithCursor(map[string]string{"stream": stream}, ""), nil
}

// Read performs one of the reviewed DynamoDB/DynamoDB Streams read operations.
// Every operation is selected by a closed stream name from streams.json; there
// is no raw X-Amz-Target/body passthrough. Requests are bounded by page_size and
// max_pages and stop on the provider's documented pagination token fields.
func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = itemsStreamName
	}
	op, ok := readOperationByStream[stream]
	if !ok {
		return fmt.Errorf("dynamodb stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, op.Operation, emit)
	}

	conn, err := resolveConfig(req.Config)
	if err != nil {
		return err
	}
	if op.Service == "streams" {
		if endpoint, err := resolveEndpoint(req.Config, true); err == nil {
			conn.endpoint = endpoint
		} else {
			return err
		}
	}
	pageSize, err := intConfig(req.Config, "page_size", defaultReadPageSize)
	if err != nil {
		return err
	}
	maxPages, err := intConfig(req.Config, "max_pages", defaultMaxPages)
	if err != nil {
		return err
	}

	var token any
	for page := 0; page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		body, err := readRequestBody(op, req, pageSize, token)
		if err != nil {
			return err
		}
		var raw map[string]any
		if err := c.doJSON(ctx, conn, op.Target, body, &raw); err != nil {
			return err
		}
		for _, rec := range recordsForOperation(op, raw) {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(rec); err != nil {
				return err
			}
		}
		next := nextTokenForOperation(op, raw)
		if next == nil || fmt.Sprint(next) == "" {
			return nil
		}
		token = next
	}
	return nil
}

func readRequestBody(op readOperation, req connectors.ReadRequest, pageSize int, token any) (map[string]any, error) {
	cfg := req.Config
	query := req.Query
	body := map[string]any{}
	putString := func(member, key string) {
		if value := firstConfigOrQuery(cfg, query, key); value != "" {
			body[member] = value
		}
	}
	putString("TableName", "table_name")
	if body["TableName"] == nil {
		putString("TableName", "table")
	}
	putString("TableArn", "table_arn")
	putString("ResourceArn", "resource_arn")
	putString("BackupArn", "backup_arn")
	putString("BackupName", "backup_name")
	putString("ExportArn", "export_arn")
	putString("ImportArn", "import_arn")
	putString("GlobalTableName", "global_table_name")
	putString("IndexName", "index_name")
	putString("RegionName", "region_name")
	putString("StreamArn", "stream_arn")
	putString("ShardId", "shard_id")

	if op.Operation == "DescribeLimits" || op.Operation == "DescribeEndpoints" {
		return body, nil
	}
	if op.Operation == "Scan" {
		table := tableName(cfg)
		if table == "" {
			return nil, fmt.Errorf("dynamodb connector requires config table_name")
		}
		body["TableName"] = table
		body["Limit"] = pageSize
		if token != nil {
			body["ExclusiveStartKey"] = token
		}
		return body, nil
	}
	if op.Operation == "Query" {
		table := tableName(cfg)
		if table == "" {
			return nil, fmt.Errorf("dynamodb query_items requires config table_name")
		}
		keyName := firstConfigOrQuery(cfg, query, "query_key_name")
		keyValue := firstConfigOrQuery(cfg, query, "query_key_value")
		if keyName == "" || keyValue == "" {
			return nil, fmt.Errorf("dynamodb query_items requires query_key_name and query_key_value")
		}
		av, err := scalarAttributeValue(firstConfigOrQuery(cfg, query, "query_key_type"), keyValue)
		if err != nil {
			return nil, err
		}
		body["TableName"] = table
		body["Limit"] = pageSize
		body["KeyConditions"] = map[string]any{keyName: map[string]any{"ComparisonOperator": "EQ", "AttributeValueList": []any{av}}}
		if token != nil {
			body["ExclusiveStartKey"] = token
		}
		return body, nil
	}
	if op.Operation == "GetRecords" {
		putString("ShardIterator", "shard_iterator")
		body["Limit"] = pageSize
		return body, nil
	}
	if op.Operation == "GetShardIterator" {
		if body["StreamArn"] == nil || body["ShardId"] == nil {
			return nil, fmt.Errorf("dynamodb streams_get_shard_iterator requires stream_arn and shard_id")
		}
		iteratorType := firstConfigOrQuery(cfg, query, "iterator_type")
		if iteratorType == "" {
			iteratorType = "TRIM_HORIZON"
		}
		body["ShardIteratorType"] = iteratorType
		putString("SequenceNumber", "sequence_number")
		return body, nil
	}
	if op.Operation == "TransactGetItems" {
		return transactGetItemsBody(cfg)
	}

	addLimitAndToken(op.Operation, body, pageSize, token)
	return body, nil
}

func firstConfigOrQuery(cfg connectors.RuntimeConfig, query map[string]string, key string) string {
	if query != nil {
		if value := strings.TrimSpace(query[key]); value != "" {
			return value
		}
	}
	if cfg.Config == nil {
		return ""
	}
	return strings.TrimSpace(cfg.Config[key])
}

func addLimitAndToken(operation string, body map[string]any, pageSize int, token any) {
	switch operation {
	case "ListBackups":
		body["Limit"] = pageSize
		if token != nil {
			body["ExclusiveStartBackupArn"] = token
		}
	case "ListTables":
		body["Limit"] = pageSize
		if token != nil {
			body["ExclusiveStartTableName"] = token
		}
	case "ListGlobalTables":
		body["Limit"] = pageSize
		if token != nil {
			body["ExclusiveStartGlobalTableName"] = token
		}
	case "ListImports":
		body["PageSize"] = pageSize
		if token != nil {
			body["NextToken"] = token
		}
	case "ListStreams":
		body["Limit"] = pageSize
		if token != nil {
			body["ExclusiveStartStreamArn"] = token
		}
	case "ListContributorInsights", "ListExports", "ListTagsOfResource":
		body["MaxResults"] = pageSize
		if token != nil {
			body["NextToken"] = token
		}
	}
}

func transactGetItemsBody(cfg connectors.RuntimeConfig) (map[string]any, error) {
	raw := strings.TrimSpace(cfg.Config["transact_get_items"])
	if raw == "" {
		return nil, fmt.Errorf("dynamodb transact_get_items requires config transact_get_items JSON fixture")
	}
	var items []any
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil, fmt.Errorf("decode transact_get_items: %w", err)
	}
	return map[string]any{"TransactItems": items}, nil
}

func nextTokenForOperation(op readOperation, raw map[string]any) any {
	switch op.Operation {
	case "Scan", "Query":
		return raw["LastEvaluatedKey"]
	case "ListBackups":
		return raw["LastEvaluatedBackupArn"]
	case "ListTables":
		return raw["LastEvaluatedTableName"]
	case "ListGlobalTables":
		return raw["LastEvaluatedGlobalTableName"]
	case "ListStreams":
		return raw["LastEvaluatedStreamArn"]
	case "ListContributorInsights", "ListExports", "ListImports", "ListTagsOfResource":
		return raw["NextToken"]
	case "GetRecords":
		return raw["NextShardIterator"]
	default:
		return nil
	}
}

func recordsForOperation(op readOperation, raw map[string]any) []connectors.Record {
	if op.Operation == "Scan" || op.Operation == "Query" {
		return itemRecords(raw["Items"])
	}
	if op.Operation == "GetRecords" {
		return streamRecordEvents(raw["Records"])
	}
	for _, key := range responseArrayKeys(op.Operation) {
		if records := arrayRecords(op, key, raw[key]); len(records) > 0 {
			return records
		}
	}
	for _, key := range responseObjectKeys(op.Operation) {
		if obj, ok := raw[key].(map[string]any); ok {
			rec := connectors.Record(obj)
			rec["operation"] = op.Operation
			rec["id"] = recordID(op.Stream, obj)
			return []connectors.Record{rec}
		}
	}
	return []connectors.Record{{"id": op.Stream + "#response", "operation": op.Operation, "response": raw}}
}

func itemRecords(value any) []connectors.Record {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	out := make([]connectors.Record, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		av := make(map[string]attributeValue, len(m))
		for k, raw := range m {
			if inner, ok := raw.(map[string]any); ok {
				av[k] = attributeValue(inner)
			}
		}
		rec := flattenItem(av)
		if _, ok := rec["pk"]; !ok {
			rec["pk"] = recordID("item", m)
		}
		out = append(out, rec)
	}
	return out
}

func streamRecordEvents(value any) []connectors.Record {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	out := make([]connectors.Record, 0, len(items))
	for i, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		rec := connectors.Record{"id": recordID(fmt.Sprintf("stream-record-%d", i), m), "operation": "GetRecords", "response": m}
		if name, _ := m["eventName"].(string); name != "" {
			rec["event_name"] = name
		}
		out = append(out, rec)
	}
	return out
}

func arrayRecords(op readOperation, key string, value any) []connectors.Record {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	out := make([]connectors.Record, 0, len(items))
	field := strings.TrimSuffix(strings.TrimSuffix(key, "s"), "Summarie")
	for i, item := range items {
		switch typed := item.(type) {
		case map[string]any:
			rec := connectors.Record(typed)
			rec["operation"] = op.Operation
			rec["id"] = recordID(fmt.Sprintf("%s-%d", op.Stream, i), typed)
			out = append(out, rec)
		case string:
			out = append(out, connectors.Record{"id": typed, "operation": op.Operation, strings.ToLower(field): typed})
		default:
			out = append(out, connectors.Record{"id": fmt.Sprintf("%s-%d", op.Stream, i), "operation": op.Operation, "response": typed})
		}
	}
	return out
}

func responseArrayKeys(operation string) []string {
	switch operation {
	case "DescribeEndpoints":
		return []string{"Endpoints"}
	case "ListBackups":
		return []string{"BackupSummaries"}
	case "ListContributorInsights":
		return []string{"ContributorInsightsSummaries"}
	case "ListExports":
		return []string{"ExportSummaries"}
	case "ListGlobalTables":
		return []string{"GlobalTables"}
	case "ListImports":
		return []string{"ImportSummaryList"}
	case "ListTables":
		return []string{"TableNames"}
	case "ListTagsOfResource":
		return []string{"Tags"}
	case "ListStreams":
		return []string{"Streams"}
	default:
		return nil
	}
}

func responseObjectKeys(operation string) []string {
	switch operation {
	case "DescribeBackup":
		return []string{"BackupDescription"}
	case "DescribeContinuousBackups":
		return []string{"ContinuousBackupsDescription"}
	case "DescribeContributorInsights":
		return []string{"ContributorInsightsRuleList"}
	case "DescribeExport":
		return []string{"ExportDescription"}
	case "DescribeGlobalTable":
		return []string{"GlobalTableDescription"}
	case "DescribeGlobalTableSettings":
		return []string{"GlobalTableSettings"}
	case "DescribeImport":
		return []string{"ImportTableDescription"}
	case "DescribeKinesisStreamingDestination":
		return []string{"KinesisDataStreamDestinations"}
	case "DescribeLimits":
		return []string{"AccountMaxReadCapacityUnits"}
	case "DescribeTable", "DescribeTableReplicaAutoScaling":
		return []string{"Table"}
	case "DescribeTimeToLive":
		return []string{"TimeToLiveDescription"}
	case "GetResourcePolicy":
		return []string{"Policy"}
	case "GetShardIterator":
		return []string{"ShardIterator"}
	case "DescribeStream":
		return []string{"StreamDescription"}
	default:
		return nil
	}
}

func recordID(fallback string, m map[string]any) string {
	for _, key := range []string{"TableName", "TableArn", "BackupArn", "BackupName", "ExportArn", "ImportArn", "ResourceArn", "StreamArn", "ShardId", "SequenceNumber", "Arn", "Name", "Policy", "ShardIterator"} {
		if value, ok := m[key]; ok && fmt.Sprint(value) != "" {
			return fmt.Sprint(value)
		}
	}
	return fallback
}

// scan issues one signed DynamoDB_20120810.Scan request and decodes its
// response, retained for the legacy parity regression test.
func (c Connector) scan(ctx context.Context, conn connConfig, body scanRequest) (scanResponse, error) {
	var out scanResponse
	if err := c.doJSON(ctx, conn, scanTarget, body, &out); err != nil {
		return scanResponse{}, err
	}
	return out, nil
}

func (c Connector) doJSON(ctx context.Context, conn connConfig, target string, body any, out any) error {
	return c.doJSONLimited(ctx, conn, target, body, out, defaultJSONResponseMaxBytes)
}

func (c Connector) doJSONLimited(ctx context.Context, conn connConfig, target string, body any, out any, maxBytes int) error {
	if maxBytes <= 0 {
		maxBytes = defaultJSONResponseMaxBytes
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encode dynamodb %s: %w", target, err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, conn.endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("build dynamodb %s: %w", target, err)
	}
	httpReq.Header.Set("Content-Type", amzJSONContentType)
	httpReq.Header.Set("X-Amz-Target", target)
	httpReq.Header.Set("User-Agent", requesterUserAgent)

	now := time.Now()
	if c.Now != nil {
		now = c.Now()
	}
	c.sign(httpReq, conn, payload, now)

	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("send dynamodb %s: %w", target, err)
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(io.LimitReader(resp.Body, int64(maxBytes)+1))
	if err != nil {
		return fmt.Errorf("read dynamodb %s response: %w", target, err)
	}
	if len(data) > maxBytes {
		return fmt.Errorf("dynamodb %s response exceeds max_bytes %d", target, maxBytes)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("dynamodb %s returned http %d", target, resp.StatusCode)
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode dynamodb %s: %w", target, err)
	}
	return nil
}

// readFixture emits deterministic records without network access so tests and
// conformance can exercise native/dynamodb credential-free.
func readFixture(ctx context.Context, stream, operation string, emit func(connectors.Record) error) error {
	if stream == itemsStreamName {
		for i := 1; i <= 2; i++ {
			if err := ctx.Err(); err != nil {
				return err
			}
			rec := connectors.Record{"pk": fmt.Sprintf("fixture#%d", i), "name": fmt.Sprintf("Fixture %d", i), "fixture": true}
			if err := emit(rec); err != nil {
				return err
			}
		}
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return emit(connectors.Record{"id": stream + "#fixture", "operation": operation, "response": map[string]any{"fixture": true}})
}
