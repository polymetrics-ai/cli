package dynamodb

import (
	"context"
	"fmt"

	"polymetrics.ai/internal/connectors"
)

// attributeValue is one DynamoDB AttributeValue envelope — a single-key map
// naming the wire type (S/N/B/BOOL/NULL/M/L/SS/NS/BS) and carrying its raw value.
type attributeValue map[string]any

// Catalog returns the reviewed static operation streams from defs/dynamodb.
// DynamoDB item shapes are table-specific, so the item/query streams expose a
// generic pk field while metadata/changefeed streams expose official response
// member names plus connector record metadata. No live discovery call is made.
func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	streams := make([]connectors.Stream, 0, len(readOperations))
	for _, op := range readOperations {
		fields, pk := catalogFieldsForReadOperation(op)
		streams = append(streams, connectors.Stream{Name: op.Stream, Description: op.Description, Fields: fields, PrimaryKey: pk})
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams}, nil
}

func catalogFieldsForReadOperation(op readOperation) ([]connectors.Field, []string) {
	if op.Stream == itemsStreamName || op.Stream == "query_items" {
		return []connectors.Field{{Name: "pk", Type: "string"}}, []string{"pk"}
	}
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "operation", Type: "string"}}
	for _, name := range operationRecordFieldNames(op.Operation) {
		fields = append(fields, connectors.Field{Name: name, Type: dynamoRecordFieldType(name)})
	}
	return fields, []string{"id"}
}

func operationRecordFieldNames(operation string) []string {
	switch operation {
	case "DescribeBackup":
		return []string{"BackupDescription"}
	case "DescribeContinuousBackups":
		return []string{"ContinuousBackupsDescription"}
	case "DescribeContributorInsights":
		return []string{"TableName", "IndexName", "ContributorInsightsStatus", "ContributorInsightsRuleList", "LastUpdateDateTime", "FailureException"}
	case "DescribeEndpoints":
		return []string{"Endpoints"}
	case "DescribeExport":
		return []string{"ExportDescription"}
	case "DescribeGlobalTable":
		return []string{"GlobalTableDescription"}
	case "DescribeGlobalTableSettings":
		return []string{"GlobalTableSettings"}
	case "DescribeImport":
		return []string{"ImportTableDescription"}
	case "DescribeKinesisStreamingDestination":
		return []string{"TableName", "KinesisDataStreamDestinations"}
	case "DescribeLimits":
		return []string{"AccountMaxReadCapacityUnits", "AccountMaxWriteCapacityUnits", "TableMaxReadCapacityUnits", "TableMaxWriteCapacityUnits"}
	case "DescribeTable":
		return []string{"Table"}
	case "DescribeTableReplicaAutoScaling":
		return []string{"TableAutoScalingDescription"}
	case "DescribeTimeToLive":
		return []string{"TimeToLiveDescription"}
	case "GetResourcePolicy":
		return []string{"Policy"}
	case "ListBackups":
		return []string{"BackupSummaries", "LastEvaluatedBackupArn"}
	case "ListContributorInsights":
		return []string{"ContributorInsightsSummaries", "NextToken"}
	case "ListExports":
		return []string{"ExportSummaries", "NextToken"}
	case "ListGlobalTables":
		return []string{"GlobalTables", "LastEvaluatedGlobalTableName"}
	case "ListImports":
		return []string{"ImportSummaryList", "NextToken"}
	case "ListTables":
		return []string{"TableNames", "LastEvaluatedTableName"}
	case "ListTagsOfResource":
		return []string{"Tags", "NextToken"}
	case "DescribeStream":
		return []string{"StreamDescription"}
	case "GetRecords":
		return []string{"eventID", "eventName", "eventVersion", "eventSource", "awsRegion", "dynamodb", "userIdentity"}
	case "GetShardIterator":
		return []string{"ShardIterator"}
	case "ListStreams":
		return []string{"Streams", "LastEvaluatedStreamArn"}
	default:
		return nil
	}
}

func dynamoRecordFieldType(name string) string {
	switch name {
	case "id", "operation", "TableName", "IndexName", "ContributorInsightsStatus", "Policy", "LastEvaluatedBackupArn", "NextToken", "LastEvaluatedGlobalTableName", "LastEvaluatedTableName", "LastEvaluatedStreamArn", "eventID", "eventName", "eventVersion", "eventSource", "awsRegion", "ShardIterator":
		return "string"
	case "AccountMaxReadCapacityUnits", "AccountMaxWriteCapacityUnits", "TableMaxReadCapacityUnits", "TableMaxWriteCapacityUnits":
		return "number"
	default:
		return "json"
	}
}

// flattenItem converts one DynamoDB item (attribute name -> AttributeValue
// envelope) into a plain connectors.Record.
func flattenItem(item map[string]attributeValue) connectors.Record {
	out := connectors.Record{}
	for name, value := range item {
		out[name] = attribute(value)
	}
	return out
}

// attribute recursively unwraps a DynamoDB AttributeValue envelope into a
// plain Go value. Unknown/future DynamoDB value kinds are passed through raw
// rather than dropped.
func attribute(v attributeValue) any {
	for kind, raw := range v {
		switch kind {
		case "S", "N", "B":
			return fmt.Sprintf("%v", raw)
		case "SS", "NS", "BS":
			return raw
		case "BOOL":
			b, _ := raw.(bool)
			return b
		case "NULL":
			return nil
		case "M":
			m, ok := raw.(map[string]any)
			if !ok {
				return raw
			}
			out := connectors.Record{}
			for k, nested := range m {
				if av, ok := nested.(map[string]any); ok {
					out[k] = attribute(attributeValue(av))
				}
			}
			return out
		case "L":
			list, ok := raw.([]any)
			if !ok {
				return raw
			}
			out := make([]any, 0, len(list))
			for _, elem := range list {
				if av, ok := elem.(map[string]any); ok {
					out = append(out, attribute(attributeValue(av)))
				}
			}
			return out
		default:
			return raw
		}
	}
	return nil
}
