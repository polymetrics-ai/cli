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

	switch op.Operation {
	case "DescribeBackup":
		return body, requireReadString(body, cfg, query, "BackupArn", "backup_arn", "describe_backup requires backup_arn")
	case "DescribeContinuousBackups", "DescribeKinesisStreamingDestination", "DescribeTable", "DescribeTableReplicaAutoScaling", "DescribeTimeToLive":
		return body, requireReadTableName(body, cfg, query, strings.ToLower(op.Stream)+" requires table_name")
	case "DescribeContributorInsights":
		if err := requireReadTableName(body, cfg, query, "describe_contributor_insights requires table_name"); err != nil {
			return nil, err
		}
		putReadString(body, cfg, query, "IndexName", "index_name")
		return body, nil
	case "DescribeEndpoints", "DescribeLimits":
		return body, nil
	case "DescribeExport":
		return body, requireReadString(body, cfg, query, "ExportArn", "export_arn", "describe_export requires export_arn")
	case "DescribeGlobalTable", "DescribeGlobalTableSettings":
		return body, requireReadString(body, cfg, query, "GlobalTableName", "global_table_name", strings.ToLower(op.Stream)+" requires global_table_name")
	case "DescribeImport":
		return body, requireReadString(body, cfg, query, "ImportArn", "import_arn", "describe_import requires import_arn")
	case "GetResourcePolicy", "ListTagsOfResource":
		if err := requireReadString(body, cfg, query, "ResourceArn", "resource_arn", strings.ToLower(op.Stream)+" requires resource_arn"); err != nil {
			return nil, err
		}
		if op.Operation == "ListTagsOfResource" {
			body["MaxResults"] = pageSize
			if token != nil {
				body["NextToken"] = token
			}
		}
		return body, nil
	case "Scan":
		if err := requireReadTableName(body, cfg, query, "connector requires config table_name"); err != nil {
			return nil, err
		}
		body["Limit"] = pageSize
		if token != nil {
			body["ExclusiveStartKey"] = token
		}
		return body, nil
	case "Query":
		if err := requireReadTableName(body, cfg, query, "query_items requires config table_name"); err != nil {
			return nil, err
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
		putReadString(body, cfg, query, "IndexName", "index_name")
		body["Limit"] = pageSize
		body["KeyConditions"] = map[string]any{keyName: map[string]any{"ComparisonOperator": "EQ", "AttributeValueList": []any{av}}}
		if token != nil {
			body["ExclusiveStartKey"] = token
		}
		return body, nil
	case "ListBackups":
		putReadTableName(body, cfg, query)
		body["Limit"] = pageSize
		if token != nil {
			body["ExclusiveStartBackupArn"] = token
		}
		return body, nil
	case "ListContributorInsights":
		putReadTableName(body, cfg, query)
		body["MaxResults"] = pageSize
		if token != nil {
			body["NextToken"] = token
		}
		return body, nil
	case "ListExports":
		putReadString(body, cfg, query, "TableArn", "table_arn")
		body["MaxResults"] = pageSize
		if token != nil {
			body["NextToken"] = token
		}
		return body, nil
	case "ListGlobalTables":
		putReadString(body, cfg, query, "RegionName", "region_name")
		body["Limit"] = pageSize
		if token != nil {
			body["ExclusiveStartGlobalTableName"] = token
		}
		return body, nil
	case "ListImports":
		putReadString(body, cfg, query, "TableArn", "table_arn")
		body["PageSize"] = pageSize
		if token != nil {
			body["NextToken"] = token
		}
		return body, nil
	case "ListTables":
		body["Limit"] = pageSize
		if token != nil {
			body["ExclusiveStartTableName"] = token
		}
		return body, nil
	case "DescribeStream":
		if err := requireReadString(body, cfg, query, "StreamArn", "stream_arn", "streams_describe_stream requires stream_arn"); err != nil {
			return nil, err
		}
		body["Limit"] = pageSize
		if token != nil {
			body["ExclusiveStartShardId"] = token
		}
		return body, nil
	case "GetRecords":
		if err := requireReadString(body, cfg, query, "ShardIterator", "shard_iterator", "streams_get_records requires shard_iterator"); err != nil {
			return nil, err
		}
		body["Limit"] = pageSize
		return body, nil
	case "GetShardIterator":
		if err := requireReadString(body, cfg, query, "StreamArn", "stream_arn", "streams_get_shard_iterator requires stream_arn"); err != nil {
			return nil, err
		}
		if err := requireReadString(body, cfg, query, "ShardId", "shard_id", "streams_get_shard_iterator requires shard_id"); err != nil {
			return nil, err
		}
		iteratorType := firstConfigOrQuery(cfg, query, "iterator_type")
		if iteratorType == "" {
			iteratorType = "TRIM_HORIZON"
		}
		body["ShardIteratorType"] = iteratorType
		putReadString(body, cfg, query, "SequenceNumber", "sequence_number")
		return body, nil
	case "ListStreams":
		putReadTableName(body, cfg, query)
		body["Limit"] = pageSize
		if token != nil {
			body["ExclusiveStartStreamArn"] = token
		}
		return body, nil
	default:
		return nil, fmt.Errorf("dynamodb stream operation %q not supported", op.Operation)
	}
}

func putReadString(body map[string]any, cfg connectors.RuntimeConfig, query map[string]string, member, key string) bool {
	if value := firstConfigOrQuery(cfg, query, key); value != "" {
		body[member] = value
		return true
	}
	return false
}

func requireReadString(body map[string]any, cfg connectors.RuntimeConfig, query map[string]string, member, key, message string) error {
	if putReadString(body, cfg, query, member, key) {
		return nil
	}
	return fmt.Errorf("dynamodb %s", message)
}

func putReadTableName(body map[string]any, cfg connectors.RuntimeConfig, query map[string]string) bool {
	if putReadString(body, cfg, query, "TableName", "table_name") {
		return true
	}
	return putReadString(body, cfg, query, "TableName", "table")
}

func requireReadTableName(body map[string]any, cfg connectors.RuntimeConfig, query map[string]string, message string) error {
	if putReadTableName(body, cfg, query) {
		return nil
	}
	return fmt.Errorf("dynamodb %s", message)
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
	case "DescribeStream":
		return raw["LastEvaluatedShardId"]
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
	return []connectors.Record{pageRecordForOperation(op, raw)}
}

func pageRecordForOperation(op readOperation, raw map[string]any) connectors.Record {
	rec := make(connectors.Record, len(raw)+2)
	for key, value := range raw {
		rec[key] = value
	}
	rec["operation"] = op.Operation
	rec["id"] = recordID(op.Stream+"#response", raw)
	return rec
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
		rec := make(connectors.Record, len(m)+2)
		for key, value := range m {
			rec[key] = value
		}
		rec["id"] = recordID(fmt.Sprintf("stream-record-%d", i), m)
		rec["operation"] = "GetRecords"
		out = append(out, rec)
	}
	return out
}

func recordID(fallback string, m map[string]any) string {
	for _, key := range []string{"TableName", "TableArn", "BackupArn", "BackupName", "ExportArn", "ImportArn", "ResourceArn", "StreamArn", "ShardId", "SequenceNumber", "eventID", "EventID", "Arn", "Name", "Policy", "ShardIterator"} {
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
	if stream == itemsStreamName || stream == "query_items" {
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
	op := readOperation{Stream: stream, Operation: operation}
	for _, rec := range recordsForOperation(op, fixtureResponseForOperation(operation)) {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func fixtureResponseForOperation(operation string) map[string]any {
	tableArn := "arn:aws:dynamodb:us-east-1:123456789012:table/fixture_table"
	streamArn := tableArn + "/stream/2026-01-01T00:00:00.000"
	switch operation {
	case "DescribeBackup":
		return map[string]any{"BackupDescription": map[string]any{"BackupDetails": map[string]any{"BackupArn": tableArn + "/backup/fixture_backup", "BackupName": "fixture_backup"}}}
	case "DescribeContinuousBackups":
		return map[string]any{"ContinuousBackupsDescription": map[string]any{"ContinuousBackupsStatus": "ENABLED", "PointInTimeRecoveryDescription": map[string]any{"PointInTimeRecoveryStatus": "ENABLED"}}}
	case "DescribeContributorInsights":
		return map[string]any{"TableName": "fixture_table", "ContributorInsightsStatus": "ENABLED", "ContributorInsightsRuleList": []any{"fixture-rule"}}
	case "DescribeEndpoints":
		return map[string]any{"Endpoints": []any{map[string]any{"Address": "dynamodb.us-east-1.amazonaws.com", "CachePeriodInMinutes": float64(60)}}}
	case "DescribeExport":
		return map[string]any{"ExportDescription": map[string]any{"ExportArn": tableArn + "/export/fixture_export", "ExportStatus": "COMPLETED"}}
	case "DescribeGlobalTable":
		return map[string]any{"GlobalTableDescription": map[string]any{"GlobalTableName": "fixture_global", "GlobalTableStatus": "ACTIVE"}}
	case "DescribeGlobalTableSettings":
		return map[string]any{"GlobalTableSettings": map[string]any{"GlobalTableName": "fixture_global", "ReplicaSettings": []any{}}}
	case "DescribeImport":
		return map[string]any{"ImportTableDescription": map[string]any{"ImportArn": tableArn + "/import/fixture_import", "ImportStatus": "COMPLETED"}}
	case "DescribeKinesisStreamingDestination":
		return map[string]any{"TableName": "fixture_table", "KinesisDataStreamDestinations": []any{map[string]any{"StreamArn": "arn:aws:kinesis:us-east-1:123456789012:stream/fixture", "DestinationStatus": "ACTIVE"}}}
	case "DescribeLimits":
		return map[string]any{"AccountMaxReadCapacityUnits": float64(80000), "AccountMaxWriteCapacityUnits": float64(80000), "TableMaxReadCapacityUnits": float64(40000), "TableMaxWriteCapacityUnits": float64(40000)}
	case "DescribeTable":
		return map[string]any{"Table": map[string]any{"TableArn": tableArn, "TableName": "fixture_table", "TableStatus": "ACTIVE"}}
	case "DescribeTableReplicaAutoScaling":
		return map[string]any{"TableAutoScalingDescription": map[string]any{"TableName": "fixture_table", "TableStatus": "ACTIVE"}}
	case "DescribeTimeToLive":
		return map[string]any{"TimeToLiveDescription": map[string]any{"TimeToLiveStatus": "ENABLED", "AttributeName": "ttl"}}
	case "GetResourcePolicy":
		return map[string]any{"Policy": `{"Version":"2012-10-17","Statement":[]}`}
	case "ListBackups":
		return map[string]any{"BackupSummaries": []any{map[string]any{"BackupArn": tableArn + "/backup/fixture_backup", "BackupName": "fixture_backup", "TableName": "fixture_table"}}}
	case "ListContributorInsights":
		return map[string]any{"ContributorInsightsSummaries": []any{map[string]any{"TableName": "fixture_table", "ContributorInsightsStatus": "ENABLED"}}}
	case "ListExports":
		return map[string]any{"ExportSummaries": []any{map[string]any{"ExportArn": tableArn + "/export/fixture_export", "ExportStatus": "COMPLETED"}}}
	case "ListGlobalTables":
		return map[string]any{"GlobalTables": []any{map[string]any{"GlobalTableName": "fixture_global", "ReplicationGroup": []any{}}}}
	case "ListImports":
		return map[string]any{"ImportSummaryList": []any{map[string]any{"ImportArn": tableArn + "/import/fixture_import", "ImportStatus": "COMPLETED"}}}
	case "ListTables":
		return map[string]any{"TableNames": []any{"fixture_table"}}
	case "ListTagsOfResource":
		return map[string]any{"Tags": []any{map[string]any{"Key": "fixture", "Value": "true"}}}
	case "DescribeStream":
		return map[string]any{"StreamDescription": map[string]any{"StreamArn": streamArn, "StreamStatus": "ENABLED", "Shards": []any{map[string]any{"ShardId": "shard-000"}}}}
	case "GetRecords":
		return map[string]any{"Records": []any{map[string]any{"eventID": "fixture-event-1", "eventName": "INSERT", "eventVersion": "1.1", "eventSource": "aws:dynamodb", "awsRegion": "us-east-1", "dynamodb": map[string]any{"SequenceNumber": "100", "NewImage": map[string]any{"pk": map[string]any{"S": "fixture#1"}}}}}}
	case "GetShardIterator":
		return map[string]any{"ShardIterator": "fixture-shard-iterator"}
	case "ListStreams":
		return map[string]any{"Streams": []any{map[string]any{"StreamArn": streamArn, "TableName": "fixture_table"}}}
	default:
		return map[string]any{"Fixture": true}
	}
}
