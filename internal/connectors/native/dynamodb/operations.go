package dynamodb

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors"
)

const (
	dynamoTargetPrefix  = "DynamoDB_20120810."
	streamsTargetPrefix = "DynamoDBStreams_20120810."
)

type readOperation struct {
	Stream      string
	Operation   string
	Target      string
	Description string
	Service     string
}

type writeOperation struct {
	Action    string
	Operation string
	Target    string
}

var readOperations = []readOperation{
	{Stream: "describe_backup", Operation: "DescribeBackup", Target: dynamoTargetPrefix + "DescribeBackup", Description: "DynamoDB backup metadata."},
	{Stream: "describe_continuous_backups", Operation: "DescribeContinuousBackups", Target: dynamoTargetPrefix + "DescribeContinuousBackups", Description: "DynamoDB point-in-time recovery metadata."},
	{Stream: "describe_contributor_insights", Operation: "DescribeContributorInsights", Target: dynamoTargetPrefix + "DescribeContributorInsights", Description: "DynamoDB Contributor Insights status."},
	{Stream: "describe_endpoints", Operation: "DescribeEndpoints", Target: dynamoTargetPrefix + "DescribeEndpoints", Description: "DynamoDB endpoint discovery metadata."},
	{Stream: "describe_export", Operation: "DescribeExport", Target: dynamoTargetPrefix + "DescribeExport", Description: "DynamoDB export metadata."},
	{Stream: "describe_global_table", Operation: "DescribeGlobalTable", Target: dynamoTargetPrefix + "DescribeGlobalTable", Description: "DynamoDB global table metadata."},
	{Stream: "describe_global_table_settings", Operation: "DescribeGlobalTableSettings", Target: dynamoTargetPrefix + "DescribeGlobalTableSettings", Description: "DynamoDB global table settings."},
	{Stream: "describe_import", Operation: "DescribeImport", Target: dynamoTargetPrefix + "DescribeImport", Description: "DynamoDB import metadata."},
	{Stream: "describe_kinesis_streaming_destination", Operation: "DescribeKinesisStreamingDestination", Target: dynamoTargetPrefix + "DescribeKinesisStreamingDestination", Description: "DynamoDB Kinesis streaming destination metadata."},
	{Stream: "describe_limits", Operation: "DescribeLimits", Target: dynamoTargetPrefix + "DescribeLimits", Description: "DynamoDB account limits."},
	{Stream: "describe_table", Operation: "DescribeTable", Target: dynamoTargetPrefix + "DescribeTable", Description: "DynamoDB table metadata."},
	{Stream: "describe_table_replica_auto_scaling", Operation: "DescribeTableReplicaAutoScaling", Target: dynamoTargetPrefix + "DescribeTableReplicaAutoScaling", Description: "DynamoDB table replica auto-scaling metadata."},
	{Stream: "describe_time_to_live", Operation: "DescribeTimeToLive", Target: dynamoTargetPrefix + "DescribeTimeToLive", Description: "DynamoDB TTL metadata."},
	{Stream: "get_resource_policy", Operation: "GetResourcePolicy", Target: dynamoTargetPrefix + "GetResourcePolicy", Description: "DynamoDB resource policy metadata."},
	{Stream: "list_backups", Operation: "ListBackups", Target: dynamoTargetPrefix + "ListBackups", Description: "DynamoDB backups."},
	{Stream: "list_contributor_insights", Operation: "ListContributorInsights", Target: dynamoTargetPrefix + "ListContributorInsights", Description: "DynamoDB Contributor Insights summaries."},
	{Stream: "list_exports", Operation: "ListExports", Target: dynamoTargetPrefix + "ListExports", Description: "DynamoDB exports."},
	{Stream: "list_global_tables", Operation: "ListGlobalTables", Target: dynamoTargetPrefix + "ListGlobalTables", Description: "DynamoDB global tables."},
	{Stream: "list_imports", Operation: "ListImports", Target: dynamoTargetPrefix + "ListImports", Description: "DynamoDB imports."},
	{Stream: "list_tables", Operation: "ListTables", Target: dynamoTargetPrefix + "ListTables", Description: "DynamoDB table names."},
	{Stream: "list_tags_of_resource", Operation: "ListTagsOfResource", Target: dynamoTargetPrefix + "ListTagsOfResource", Description: "DynamoDB resource tags."},
	{Stream: "items", Operation: "Scan", Target: dynamoTargetPrefix + "Scan", Description: "DynamoDB table items via bounded Scan."},
	{Stream: "query_items", Operation: "Query", Target: dynamoTargetPrefix + "Query", Description: "DynamoDB table items via bounded typed Query KeyConditions."},
	{Stream: "streams_describe_stream", Operation: "DescribeStream", Target: streamsTargetPrefix + "DescribeStream", Description: "DynamoDB Streams stream metadata.", Service: "streams"},
	{Stream: "streams_get_records", Operation: "GetRecords", Target: streamsTargetPrefix + "GetRecords", Description: "DynamoDB Streams records.", Service: "streams"},
	{Stream: "streams_get_shard_iterator", Operation: "GetShardIterator", Target: streamsTargetPrefix + "GetShardIterator", Description: "DynamoDB Streams shard iterator.", Service: "streams"},
	{Stream: "streams_list_streams", Operation: "ListStreams", Target: streamsTargetPrefix + "ListStreams", Description: "DynamoDB Streams stream list.", Service: "streams"},
}

var readOperationByStream = func() map[string]readOperation {
	out := make(map[string]readOperation, len(readOperations))
	for _, op := range readOperations {
		out[op.Stream] = op
	}
	return out
}()

var writeOperations = []writeOperation{
	{Action: "batch_write_item", Operation: "BatchWriteItem", Target: dynamoTargetPrefix + "BatchWriteItem"},
	{Action: "create_backup", Operation: "CreateBackup", Target: dynamoTargetPrefix + "CreateBackup"},
	{Action: "create_global_table", Operation: "CreateGlobalTable", Target: dynamoTargetPrefix + "CreateGlobalTable"},
	{Action: "create_table", Operation: "CreateTable", Target: dynamoTargetPrefix + "CreateTable"},
	{Action: "delete_backup", Operation: "DeleteBackup", Target: dynamoTargetPrefix + "DeleteBackup"},
	{Action: "delete_item", Operation: "DeleteItem", Target: dynamoTargetPrefix + "DeleteItem"},
	{Action: "delete_resource_policy", Operation: "DeleteResourcePolicy", Target: dynamoTargetPrefix + "DeleteResourcePolicy"},
	{Action: "delete_table", Operation: "DeleteTable", Target: dynamoTargetPrefix + "DeleteTable"},
	{Action: "disable_kinesis_streaming_destination", Operation: "DisableKinesisStreamingDestination", Target: dynamoTargetPrefix + "DisableKinesisStreamingDestination"},
	{Action: "enable_kinesis_streaming_destination", Operation: "EnableKinesisStreamingDestination", Target: dynamoTargetPrefix + "EnableKinesisStreamingDestination"},
	{Action: "put_item", Operation: "PutItem", Target: dynamoTargetPrefix + "PutItem"},
	{Action: "put_resource_policy", Operation: "PutResourcePolicy", Target: dynamoTargetPrefix + "PutResourcePolicy"},
	{Action: "restore_table_from_backup", Operation: "RestoreTableFromBackup", Target: dynamoTargetPrefix + "RestoreTableFromBackup"},
	{Action: "restore_table_to_point_in_time", Operation: "RestoreTableToPointInTime", Target: dynamoTargetPrefix + "RestoreTableToPointInTime"},
	{Action: "tag_resource", Operation: "TagResource", Target: dynamoTargetPrefix + "TagResource"},
	{Action: "transact_write_items", Operation: "TransactWriteItems", Target: dynamoTargetPrefix + "TransactWriteItems"},
	{Action: "untag_resource", Operation: "UntagResource", Target: dynamoTargetPrefix + "UntagResource"},
	{Action: "update_continuous_backups", Operation: "UpdateContinuousBackups", Target: dynamoTargetPrefix + "UpdateContinuousBackups"},
	{Action: "update_contributor_insights", Operation: "UpdateContributorInsights", Target: dynamoTargetPrefix + "UpdateContributorInsights"},
	{Action: "update_global_table", Operation: "UpdateGlobalTable", Target: dynamoTargetPrefix + "UpdateGlobalTable"},
	{Action: "update_global_table_settings", Operation: "UpdateGlobalTableSettings", Target: dynamoTargetPrefix + "UpdateGlobalTableSettings"},
	{Action: "update_item", Operation: "UpdateItem", Target: dynamoTargetPrefix + "UpdateItem"},
	{Action: "update_kinesis_streaming_destination", Operation: "UpdateKinesisStreamingDestination", Target: dynamoTargetPrefix + "UpdateKinesisStreamingDestination"},
	{Action: "update_table", Operation: "UpdateTable", Target: dynamoTargetPrefix + "UpdateTable"},
	{Action: "update_table_replica_auto_scaling", Operation: "UpdateTableReplicaAutoScaling", Target: dynamoTargetPrefix + "UpdateTableReplicaAutoScaling"},
	{Action: "update_time_to_live", Operation: "UpdateTimeToLive", Target: dynamoTargetPrefix + "UpdateTimeToLive"},
}

var writeOperationByAction = func() map[string]writeOperation {
	out := make(map[string]writeOperation, len(writeOperations))
	for _, op := range writeOperations {
		out[op.Action] = op
	}
	return out
}()

func streamNames() []string {
	out := make([]string, 0, len(readOperations))
	for _, op := range readOperations {
		out = append(out, op.Stream)
	}
	sort.Strings(out)
	return out
}

func writeActionNames() []string {
	out := make([]string, 0, len(writeOperations))
	for _, op := range writeOperations {
		out = append(out, op.Action)
	}
	sort.Strings(out)
	return out
}

func awsMemberName(field string) string {
	special := map[string]string{
		"backup_arn": "BackupArn", "export_arn": "ExportArn", "global_table_name": "GlobalTableName", "import_arn": "ImportArn",
		"resource_arn": "ResourceArn", "stream_arn": "StreamArn", "table_arn": "TableArn", "table_name": "TableName",
		"target_table_name": "TargetTableName", "source_table_name": "SourceTableName", "key_schema": "KeySchema",
		"sse_specification": "SSESpecification", "sse_specification_override": "SSESpecificationOverride",
		"time_to_live_specification": "TimeToLiveSpecification", "point_in_time_recovery_specification": "PointInTimeRecoverySpecification",
		"global_secondary_indexes": "GlobalSecondaryIndexes", "global_secondary_index_override": "GlobalSecondaryIndexOverride",
		"global_secondary_index_updates": "GlobalSecondaryIndexUpdates", "global_secondary_index_settings_update": "GlobalTableGlobalSecondaryIndexSettingsUpdate",
		"local_secondary_indexes": "LocalSecondaryIndexes", "local_secondary_index_override": "LocalSecondaryIndexOverride",
		"provisioned_throughput": "ProvisionedThroughput", "provisioned_throughput_override": "ProvisionedThroughputOverride",
		"provisioned_write_capacity_auto_scaling_update":                       "ProvisionedWriteCapacityAutoScalingUpdate",
		"global_table_provisioned_write_capacity_units":                        "GlobalTableProvisionedWriteCapacityUnits",
		"global_table_provisioned_write_capacity_auto_scaling_settings_update": "GlobalTableProvisionedWriteCapacityAutoScalingSettingsUpdate",
		"global_table_billing_mode":                                            "GlobalTableBillingMode", "replica_settings_update": "ReplicaSettingsUpdate",
		"replica_updates": "ReplicaUpdates", "replication_group": "ReplicationGroup", "return_consumed_capacity": "ReturnConsumedCapacity",
		"return_item_collection_metrics": "ReturnItemCollectionMetrics", "return_values": "ReturnValues", "client_request_token": "ClientRequestToken",
		"kinesis_stream_arn": "StreamArn", "enable_kinesis_streaming_configuration": "EnableKinesisStreamingConfiguration",
		"update_kinesis_streaming_configuration": "UpdateKinesisStreamingConfiguration", "deletion_protection_enabled": "DeletionProtectionEnabled",
	}
	if v, ok := special[field]; ok {
		return v
	}
	parts := strings.Split(field, "_")
	for i, part := range parts {
		switch strings.ToLower(part) {
		case "arn":
			parts[i] = "Arn"
		case "sse":
			parts[i] = "SSE"
		default:
			if part == "" {
				continue
			}
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

func cloneBodyValue(v any) any {
	b, err := json.Marshal(v)
	if err != nil {
		return v
	}
	var out any
	if err := json.Unmarshal(b, &out); err != nil {
		return v
	}
	return out
}

func bodyFromRecord(rec connectors.Record) map[string]any {
	body := make(map[string]any, len(rec))
	for k, v := range rec {
		body[awsMemberName(k)] = cloneBodyValue(v)
	}
	return body
}

func buildWriteBody(action string, rec connectors.Record) (map[string]any, error) {
	switch action {
	case "batch_write_item":
		return batchWriteItemBody(rec)
	case "transact_write_items":
		return transactWriteItemsWriteBody(rec)
	default:
		return bodyFromRecord(rec), nil
	}
}

func batchWriteItemBody(rec connectors.Record) (map[string]any, error) {
	table := recordString(rec, "table_name")
	if table == "" {
		return nil, fmt.Errorf("dynamodb batch_write_item requires table_name")
	}
	request, err := batchWriteRequest(recordString(rec, "operation"), rec)
	if err != nil {
		return nil, err
	}
	body := map[string]any{"RequestItems": map[string]any{table: []any{request}}}
	copyOptionalWriteMembers(body, rec, "return_consumed_capacity", "return_item_collection_metrics")
	return body, nil
}

func batchWriteRequest(operation string, rec connectors.Record) (map[string]any, error) {
	switch strings.ToLower(strings.TrimSpace(operation)) {
	case "put":
		item, ok := recordObject(rec, "item")
		if !ok {
			return nil, fmt.Errorf("dynamodb batch_write_item put requires item")
		}
		return map[string]any{"PutRequest": map[string]any{"Item": item}}, nil
	case "delete":
		key, ok := recordObject(rec, "key")
		if !ok {
			return nil, fmt.Errorf("dynamodb batch_write_item delete requires key")
		}
		return map[string]any{"DeleteRequest": map[string]any{"Key": key}}, nil
	default:
		return nil, fmt.Errorf("dynamodb batch_write_item requires operation put or delete")
	}
}

func transactWriteItemsWriteBody(rec connectors.Record) (map[string]any, error) {
	table := recordString(rec, "table_name")
	if table == "" {
		return nil, fmt.Errorf("dynamodb transact_write_items requires table_name")
	}
	item, err := transactWriteItem(recordString(rec, "operation"), table, rec)
	if err != nil {
		return nil, err
	}
	body := map[string]any{"TransactItems": []any{item}}
	copyOptionalWriteMembers(body, rec, "client_request_token", "return_consumed_capacity", "return_item_collection_metrics")
	return body, nil
}

func transactWriteItem(operation, table string, rec connectors.Record) (map[string]any, error) {
	switch strings.ToLower(strings.TrimSpace(operation)) {
	case "put":
		item, ok := recordObject(rec, "item")
		if !ok {
			return nil, fmt.Errorf("dynamodb transact_write_items put requires item")
		}
		return map[string]any{"Put": map[string]any{"TableName": table, "Item": item}}, nil
	case "delete":
		key, ok := recordObject(rec, "key")
		if !ok {
			return nil, fmt.Errorf("dynamodb transact_write_items delete requires key")
		}
		return map[string]any{"Delete": map[string]any{"TableName": table, "Key": key}}, nil
	default:
		return nil, fmt.Errorf("dynamodb transact_write_items requires operation put or delete")
	}
}

func copyOptionalWriteMembers(body map[string]any, rec connectors.Record, fields ...string) {
	for _, field := range fields {
		value, ok := rec[field]
		if !ok || value == nil {
			continue
		}
		body[awsMemberName(field)] = cloneBodyValue(value)
	}
}

func recordString(rec connectors.Record, key string) string {
	if rec == nil {
		return ""
	}
	value, ok := rec[key]
	if !ok || value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func recordObject(rec connectors.Record, key string) (map[string]any, bool) {
	value, ok := rec[key]
	if !ok || value == nil {
		return nil, false
	}
	switch typed := value.(type) {
	case map[string]any:
		if len(typed) == 0 {
			return nil, false
		}
		out, ok := cloneBodyValue(typed).(map[string]any)
		return out, ok
	case connectors.Record:
		if len(typed) == 0 {
			return nil, false
		}
		cloned := cloneBodyValue(map[string]any(typed))
		out, ok := cloned.(map[string]any)
		return out, ok
	default:
		return nil, false
	}
}

func scalarAttributeValue(kind, value string) (map[string]any, error) {
	kind = strings.ToUpper(strings.TrimSpace(kind))
	if kind == "" {
		kind = "S"
	}
	switch kind {
	case "S", "N", "B":
		return map[string]any{kind: value}, nil
	default:
		return nil, fmt.Errorf("unsupported DynamoDB scalar key type %q", kind)
	}
}
