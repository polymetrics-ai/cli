---
name: pm-dynamodb
description: DynamoDB connector knowledge and safe action guide.
---

# pm-dynamodb

## Purpose

Reads, changes, and writes Amazon DynamoDB tables through the AWS JSON API with connector-local SigV4 signing, typed operation schemas, bounded reads, DynamoDB Streams changefeed support, and reverse-ETL write gates.

## Icon

- asset: icons/dynamodb.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: database

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- backup_arn
- backup_name
- base_url
- endpoint
- export_arn
- global_table_name
- import_arn
- index_name
- iterator_type
- kinesis_stream_arn
- max_pages
- mode
- page_size
- query_key_name
- query_key_type
- query_key_value
- region
- region_name
- resource_arn
- sequence_number
- shard_id
- stream_arn
- streams_endpoint
- table
- table_arn
- table_name
- access_key_id (secret)
- secret_access_key (secret)

## ETL Streams

- describe_backup:
  - primary key: id
  - fields: BackupDescription(), id(), operation()
- describe_continuous_backups:
  - primary key: id
  - fields: ContinuousBackupsDescription(), id(), operation()
- describe_contributor_insights:
  - primary key: id
  - fields: ContributorInsightsRuleList(), ContributorInsightsStatus(), FailureException(), IndexName(), LastUpdateDateTime(), TableName(), id(), operation()
- describe_endpoints:
  - primary key: id
  - fields: Endpoints(), id(), operation()
- describe_export:
  - primary key: id
  - fields: ExportDescription(), id(), operation()
- describe_global_table:
  - primary key: id
  - fields: GlobalTableDescription(), id(), operation()
- describe_global_table_settings:
  - primary key: id
  - fields: GlobalTableSettings(), id(), operation()
- describe_import:
  - primary key: id
  - fields: ImportTableDescription(), id(), operation()
- describe_kinesis_streaming_destination:
  - primary key: id
  - fields: KinesisDataStreamDestinations(), TableName(), id(), operation()
- describe_limits:
  - primary key: id
  - fields: AccountMaxReadCapacityUnits(), AccountMaxWriteCapacityUnits(), TableMaxReadCapacityUnits(), TableMaxWriteCapacityUnits(), id(), operation()
- describe_table:
  - primary key: id
  - fields: Table(), id(), operation()
- describe_table_replica_auto_scaling:
  - primary key: id
  - fields: TableAutoScalingDescription(), id(), operation()
- describe_time_to_live:
  - primary key: id
  - fields: TimeToLiveDescription(), id(), operation()
- get_resource_policy:
  - primary key: id
  - fields: Policy(), id(), operation()
- list_backups:
  - primary key: id
  - fields: BackupSummaries(), LastEvaluatedBackupArn(), id(), operation()
- list_contributor_insights:
  - primary key: id
  - fields: ContributorInsightsSummaries(), NextToken(), id(), operation()
- list_exports:
  - primary key: id
  - fields: ExportSummaries(), NextToken(), id(), operation()
- list_global_tables:
  - primary key: id
  - fields: GlobalTables(), LastEvaluatedGlobalTableName(), id(), operation()
- list_imports:
  - primary key: id
  - fields: ImportSummaryList(), NextToken(), id(), operation()
- list_tables:
  - primary key: id
  - fields: LastEvaluatedTableName(), TableNames(), id(), operation()
- list_tags_of_resource:
  - primary key: id
  - fields: NextToken(), Tags(), id(), operation()
- items:
  - primary key: pk
  - fields: pk()
- query_items:
  - primary key: pk
  - fields: pk()
- streams_describe_stream:
  - primary key: id
  - fields: StreamDescription(), id(), operation()
- streams_get_records:
  - primary key: id
  - fields: awsRegion(), dynamodb(), eventID(), eventName(), eventSource(), eventVersion(), id(), operation(), userIdentity()
- streams_get_shard_iterator:
  - primary key: id
  - fields: ShardIterator(), id(), operation()
- streams_list_streams:
  - primary key: id
  - fields: LastEvaluatedStreamArn(), Streams(), id(), operation()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- batch_write_item:
  - endpoint: POST /
  - required fields: table_name, operation
  - risk: SigV4-signed DynamoDB BatchWriteItem mutation built from one typed put/delete request per record; no raw RequestItems body passthrough is accepted.
- create_backup:
  - endpoint: POST /
  - required fields: table_name, backup_name
  - risk: SigV4-signed DynamoDB CreateBackup mutation/admin operation with typed top-level schema.
- create_global_table:
  - endpoint: POST /
  - required fields: global_table_name, replication_group
  - risk: SigV4-signed DynamoDB CreateGlobalTable mutation/admin operation with typed top-level schema.
- create_table:
  - endpoint: POST /
  - required fields: table_name, attribute_definitions, key_schema, billing_mode
  - risk: SigV4-signed DynamoDB CreateTable mutation/admin operation with typed top-level schema.
- delete_backup:
  - endpoint: POST /
  - required fields: backup_arn
  - risk: SigV4-signed DynamoDB DeleteBackup mutation/admin operation with typed top-level schema.
- delete_item:
  - endpoint: POST /
  - required fields: table_name, key
  - risk: SigV4-signed DynamoDB DeleteItem mutation/admin operation with typed top-level schema.
- delete_resource_policy:
  - endpoint: POST /
  - required fields: resource_arn
  - risk: SigV4-signed DynamoDB DeleteResourcePolicy mutation/admin operation with typed top-level schema.
- delete_table:
  - endpoint: POST /
  - required fields: table_name
  - risk: SigV4-signed DynamoDB DeleteTable mutation/admin operation with typed top-level schema.
- disable_kinesis_streaming_destination:
  - endpoint: POST /
  - required fields: table_name, stream_arn
  - risk: SigV4-signed DynamoDB DisableKinesisStreamingDestination mutation/admin operation with typed top-level schema.
- enable_kinesis_streaming_destination:
  - endpoint: POST /
  - required fields: table_name, stream_arn
  - risk: SigV4-signed DynamoDB EnableKinesisStreamingDestination mutation/admin operation with typed top-level schema.
- put_item:
  - endpoint: POST /
  - required fields: table_name, item
  - risk: SigV4-signed DynamoDB PutItem mutation/admin operation with typed top-level schema.
- put_resource_policy:
  - endpoint: POST /
  - required fields: resource_arn, policy
  - risk: SigV4-signed DynamoDB PutResourcePolicy mutation/admin operation with typed top-level schema.
- restore_table_from_backup:
  - endpoint: POST /
  - required fields: target_table_name, backup_arn
  - risk: SigV4-signed DynamoDB RestoreTableFromBackup mutation/admin operation with typed top-level schema.
- restore_table_to_point_in_time:
  - endpoint: POST /
  - required fields: source_table_name, target_table_name
  - risk: SigV4-signed DynamoDB RestoreTableToPointInTime mutation/admin operation with typed top-level schema.
- tag_resource:
  - endpoint: POST /
  - required fields: resource_arn, tags
  - risk: SigV4-signed DynamoDB TagResource mutation/admin operation with typed top-level schema.
- transact_write_items:
  - endpoint: POST /
  - required fields: table_name, operation
  - risk: SigV4-signed DynamoDB TransactWriteItems mutation built from one typed put/delete transaction item per record; no raw TransactItems body passthrough is accepted.
- untag_resource:
  - endpoint: POST /
  - required fields: resource_arn, tag_keys
  - risk: SigV4-signed DynamoDB UntagResource mutation/admin operation with typed top-level schema.
- update_continuous_backups:
  - endpoint: POST /
  - required fields: table_name, point_in_time_recovery_specification
  - risk: SigV4-signed DynamoDB UpdateContinuousBackups mutation/admin operation with typed top-level schema.
- update_contributor_insights:
  - endpoint: POST /
  - required fields: table_name, contributor_insights_action
  - risk: SigV4-signed DynamoDB UpdateContributorInsights mutation/admin operation with typed top-level schema.
- update_global_table:
  - endpoint: POST /
  - required fields: global_table_name, replica_updates
  - risk: SigV4-signed DynamoDB UpdateGlobalTable mutation/admin operation with typed top-level schema.
- update_global_table_settings:
  - endpoint: POST /
  - required fields: global_table_name
  - risk: SigV4-signed DynamoDB UpdateGlobalTableSettings mutation/admin operation with typed top-level schema.
- update_item:
  - endpoint: POST /
  - required fields: table_name, key, attribute_updates
  - risk: SigV4-signed DynamoDB UpdateItem mutation/admin operation with typed top-level schema.
- update_kinesis_streaming_destination:
  - endpoint: POST /
  - required fields: table_name, stream_arn, update_kinesis_streaming_configuration
  - risk: SigV4-signed DynamoDB UpdateKinesisStreamingDestination mutation/admin operation with typed top-level schema.
- update_table:
  - endpoint: POST /
  - required fields: table_name
  - risk: SigV4-signed DynamoDB UpdateTable mutation/admin operation with typed top-level schema.
- update_table_replica_auto_scaling:
  - endpoint: POST /
  - required fields: table_name
  - risk: SigV4-signed DynamoDB UpdateTableReplicaAutoScaling mutation/admin operation with typed top-level schema.
- update_time_to_live:
  - endpoint: POST /
  - required fields: table_name, time_to_live_specification
  - risk: SigV4-signed DynamoDB UpdateTimeToLive mutation/admin operation with typed top-level schema.

## Security

- read risk: bounded SigV4-signed DynamoDB and DynamoDB Streams JSON-RPC reads against configured table, resource, or stream identifiers
- write risk: typed SigV4-signed DynamoDB mutations/admin actions declared in writes.json; destructive actions require confirmation
- approval: reverse ETL writes require plan, preview, approval token, and destructive confirmation where declared
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Typed, bounded DynamoDB provider commands
- Usage: pm connectors run dynamodb <command> [flags]
- Source CLI: AWS CLI dynamodb (AWS CLI DynamoDB command reference)
- Global flags:
  - --max-bytes (integer): Maximum response bytes for direct reads; capped by the CLI runtime
- Direct reads
  - get-item - Get one DynamoDB item by typed partition and optional sort key attributes [intent=direct_read availability=implemented]; risk: Reads one item by closed typed scalar key attributes; key values are redacted from errors/output surfaces.; flags: --table-name, --partition-key-name, --partition-key-type, --partition-key-value, --sort-key-name, --sort-key-type, --sort-key-value, --consistent-read
  - batch-get-item - Read multiple DynamoDB items by typed partition and optional sort key lists [intent=direct_read availability=implemented]; risk: Reads bounded keyed items by closed typed scalar key attributes; key values are redacted from errors/output surfaces.; flags: --table-name, --partition-key-name, --partition-key-type, --partition-key-values, --sort-key-name, --sort-key-type, --sort-key-values, --consistent-read
  - transact-get-items - Transactionally read multiple DynamoDB items by typed partition and optional sort key lists [intent=direct_read availability=implemented]; risk: Transactionally reads bounded keyed items by closed typed scalar key attributes; key values are redacted from errors/output surfaces.; flags: --table-name, --partition-key-name, --partition-key-type, --partition-key-values, --sort-key-name, --sort-key-type, --sort-key-values
- Help topics:
  - dynamodb-safety - DynamoDB connector commands are typed and do not expose raw PartiQL, expressions, or generic HTTP bodies.

## Commands

### Inspect as a manual

```bash
pm connectors inspect dynamodb
```

### Inspect as structured JSON

```bash
pm connectors inspect dynamodb --json
```

## Agent Rules

- Run pm connectors inspect dynamodb before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
