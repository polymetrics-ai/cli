---
name: pm-aws-cloudtrail
description: AWS CloudTrail connector knowledge and safe action guide.
---

# pm-aws-cloudtrail

## Purpose

Reads AWS CloudTrail configuration and event metadata, runs bounded CloudTrail query/lookups, and executes typed approval-gated CloudTrail administration actions through fixed AWS JSON-RPC operations.

## Icon

- asset: icons/aws-cloudtrail.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- aws_region_name
- base_url
- max_pages
- mode
- page_size
- start_date
- aws_key_id (secret)
- aws_secret_key (secret)

## ETL Streams

- describe_trails:
  - primary key: pm_record_id
  - fields: operation(), pm_record_id(), trailList()
- get_channel:
  - primary key: pm_record_id
  - fields: ChannelArn(), Destinations(), IngestionStatus(), Name(), Source(), SourceConfig(), operation(), pm_record_id()
- get_dashboard:
  - primary key: pm_record_id
  - fields: CreatedTimestamp(), DashboardArn(), LastRefreshFailureReason(), LastRefreshId(), RefreshSchedule(), Status(), TerminationProtectionEnabled(), Type(), UpdatedTimestamp(), Widgets(), operation(), pm_record_id()
- get_event_configuration:
  - primary key: pm_record_id
  - fields: AggregationConfigurations(), ContextKeySelectors(), EventDataStoreArn(), MaxEventSize(), TrailARN(), operation(), pm_record_id()
- get_event_data_store:
  - primary key: pm_record_id
  - fields: AdvancedEventSelectors(), BillingMode(), CreatedTimestamp(), EventDataStoreArn(), FederationRoleArn(), FederationStatus(), KmsKeyId(), MultiRegionEnabled(), Name(), OrganizationEnabled(), PartitionKeys(), RetentionPeriod(), Status(), TerminationProtectionEnabled(), UpdatedTimestamp(), operation(), pm_record_id()
- get_event_selectors:
  - primary key: pm_record_id
  - fields: AdvancedEventSelectors(), EventSelectors(), TrailARN(), operation(), pm_record_id()
- get_import:
  - primary key: pm_record_id
  - fields: CreatedTimestamp(), Destinations(), EndEventTime(), ImportId(), ImportSource(), ImportStatistics(), ImportStatus(), StartEventTime(), UpdatedTimestamp(), operation(), pm_record_id()
- get_insight_selectors:
  - primary key: pm_record_id
  - fields: EventDataStoreArn(), InsightSelectors(), InsightsDestination(), TrailARN(), operation(), pm_record_id()
- get_resource_policy:
  - primary key: pm_record_id
  - fields: DelegatedAdminResourcePolicy(), ResourceArn(), ResourcePolicy(), operation(), pm_record_id()
- get_trail:
  - primary key: pm_record_id
  - fields: Trail(), operation(), pm_record_id()
- get_trail_status:
  - primary key: pm_record_id
  - fields: IsLogging(), LatestCloudWatchLogsDeliveryError(), LatestCloudWatchLogsDeliveryTime(), LatestDeliveryAttemptSucceeded(), LatestDeliveryAttemptTime(), LatestDeliveryError(), LatestDeliveryTime(), LatestDigestDeliveryError(), LatestDigestDeliveryTime(), LatestNotificationAttemptSucceeded(), LatestNotificationAttemptTime(), LatestNotificationError(), LatestNotificationTime(), StartLoggingTime(), StopLoggingTime(), TimeLoggingStarted(), TimeLoggingStopped(), operation(), pm_record_id()
- list_channels:
  - primary key: pm_record_id
  - fields: Channels(), NextToken(), operation(), pm_record_id()
- list_dashboards:
  - primary key: pm_record_id
  - fields: Dashboards(), NextToken(), operation(), pm_record_id()
- list_event_data_stores:
  - primary key: pm_record_id
  - fields: EventDataStores(), NextToken(), operation(), pm_record_id()
- list_import_failures:
  - primary key: pm_record_id
  - fields: Failures(), NextToken(), operation(), pm_record_id()
- list_imports:
  - primary key: pm_record_id
  - fields: Imports(), NextToken(), operation(), pm_record_id()
- list_public_keys:
  - primary key: pm_record_id
  - fields: NextToken(), PublicKeyList(), operation(), pm_record_id()
- list_tags:
  - primary key: pm_record_id
  - fields: NextToken(), ResourceTagList(), operation(), pm_record_id()
- list_trails:
  - primary key: pm_record_id
  - fields: NextToken(), Trails(), operation(), pm_record_id()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- add_tags:
  - endpoint: POST /
  - required fields: ResourceId, TagsList
  - risk: Executes AWS CloudTrail AddTags through a fixed typed JSON-RPC action; no raw action or path is accepted.
- create_channel:
  - endpoint: POST /
  - required fields: Destinations, Name, Source
  - risk: Executes AWS CloudTrail CreateChannel through a fixed typed JSON-RPC action; no raw action or path is accepted.
- create_dashboard:
  - endpoint: POST /
  - required fields: Name
  - risk: Executes AWS CloudTrail CreateDashboard through a fixed typed JSON-RPC action; no raw action or path is accepted.
- create_event_data_store:
  - endpoint: POST /
  - required fields: Name
  - risk: Executes AWS CloudTrail CreateEventDataStore through a fixed typed JSON-RPC action; no raw action or path is accepted.
- create_trail:
  - endpoint: POST /
  - required fields: Name, S3BucketName
  - risk: Executes AWS CloudTrail CreateTrail through a fixed typed JSON-RPC action; no raw action or path is accepted.
- delete_channel:
  - endpoint: POST /
  - required fields: Channel
  - risk: Executes AWS CloudTrail DeleteChannel through a fixed typed JSON-RPC action; no raw action or path is accepted.
- delete_dashboard:
  - endpoint: POST /
  - required fields: DashboardId
  - risk: Executes AWS CloudTrail DeleteDashboard through a fixed typed JSON-RPC action; no raw action or path is accepted.
- delete_event_data_store:
  - endpoint: POST /
  - required fields: EventDataStore
  - risk: Executes AWS CloudTrail DeleteEventDataStore through a fixed typed JSON-RPC action; no raw action or path is accepted.
- delete_resource_policy:
  - endpoint: POST /
  - required fields: ResourceArn
  - risk: Executes AWS CloudTrail DeleteResourcePolicy through a fixed typed JSON-RPC action; no raw action or path is accepted.
- delete_trail:
  - endpoint: POST /
  - required fields: Name
  - risk: Executes AWS CloudTrail DeleteTrail through a fixed typed JSON-RPC action; no raw action or path is accepted.
- deregister_organization_delegated_admin:
  - endpoint: POST /
  - required fields: DelegatedAdminAccountId
  - risk: Executes AWS CloudTrail DeregisterOrganizationDelegatedAdmin through a fixed typed JSON-RPC action; no raw action or path is accepted.
- disable_federation:
  - endpoint: POST /
  - required fields: EventDataStore
  - risk: Executes AWS CloudTrail DisableFederation through a fixed typed JSON-RPC action; no raw action or path is accepted.
- enable_federation:
  - endpoint: POST /
  - required fields: EventDataStore, FederationRoleArn
  - risk: Executes AWS CloudTrail EnableFederation through a fixed typed JSON-RPC action; no raw action or path is accepted.
- put_event_configuration:
  - endpoint: POST /
  - risk: Executes AWS CloudTrail PutEventConfiguration through a fixed typed JSON-RPC action; no raw action or path is accepted.
- put_event_selectors:
  - endpoint: POST /
  - required fields: TrailName
  - risk: Executes AWS CloudTrail PutEventSelectors through a fixed typed JSON-RPC action; no raw action or path is accepted.
- put_insight_selectors:
  - endpoint: POST /
  - required fields: InsightSelectors
  - risk: Executes AWS CloudTrail PutInsightSelectors through a fixed typed JSON-RPC action; no raw action or path is accepted.
- put_resource_policy:
  - endpoint: POST /
  - required fields: ResourceArn, ResourcePolicy
  - risk: Executes AWS CloudTrail PutResourcePolicy through a fixed typed JSON-RPC action; no raw action or path is accepted.
- register_organization_delegated_admin:
  - endpoint: POST /
  - required fields: MemberAccountId
  - risk: Executes AWS CloudTrail RegisterOrganizationDelegatedAdmin through a fixed typed JSON-RPC action; no raw action or path is accepted.
- remove_tags:
  - endpoint: POST /
  - required fields: ResourceId, TagsList
  - risk: Executes AWS CloudTrail RemoveTags through a fixed typed JSON-RPC action; no raw action or path is accepted.
- restore_event_data_store:
  - endpoint: POST /
  - required fields: EventDataStore
  - risk: Executes AWS CloudTrail RestoreEventDataStore through a fixed typed JSON-RPC action; no raw action or path is accepted.
- start_dashboard_refresh:
  - endpoint: POST /
  - required fields: DashboardId
  - risk: Executes AWS CloudTrail StartDashboardRefresh through a fixed typed JSON-RPC action; no raw action or path is accepted.
- start_event_data_store_ingestion:
  - endpoint: POST /
  - required fields: EventDataStore
  - risk: Executes AWS CloudTrail StartEventDataStoreIngestion through a fixed typed JSON-RPC action; no raw action or path is accepted.
- start_import:
  - endpoint: POST /
  - risk: Executes AWS CloudTrail StartImport through a fixed typed JSON-RPC action; no raw action or path is accepted.
- start_logging:
  - endpoint: POST /
  - required fields: Name
  - risk: Executes AWS CloudTrail StartLogging through a fixed typed JSON-RPC action; no raw action or path is accepted.
- stop_event_data_store_ingestion:
  - endpoint: POST /
  - required fields: EventDataStore
  - risk: Executes AWS CloudTrail StopEventDataStoreIngestion through a fixed typed JSON-RPC action; no raw action or path is accepted.
- stop_import:
  - endpoint: POST /
  - required fields: ImportId
  - risk: Executes AWS CloudTrail StopImport through a fixed typed JSON-RPC action; no raw action or path is accepted.
- stop_logging:
  - endpoint: POST /
  - required fields: Name
  - risk: Executes AWS CloudTrail StopLogging through a fixed typed JSON-RPC action; no raw action or path is accepted.
- update_channel:
  - endpoint: POST /
  - required fields: Channel
  - risk: Executes AWS CloudTrail UpdateChannel through a fixed typed JSON-RPC action; no raw action or path is accepted.
- update_dashboard:
  - endpoint: POST /
  - required fields: DashboardId
  - risk: Executes AWS CloudTrail UpdateDashboard through a fixed typed JSON-RPC action; no raw action or path is accepted.
- update_event_data_store:
  - endpoint: POST /
  - required fields: EventDataStore
  - risk: Executes AWS CloudTrail UpdateEventDataStore through a fixed typed JSON-RPC action; no raw action or path is accepted.
- update_trail:
  - endpoint: POST /
  - required fields: Name
  - risk: Executes AWS CloudTrail UpdateTrail through a fixed typed JSON-RPC action; no raw action or path is accepted.

## Security

- read risk: bounded AWS CloudTrail JSON-RPC reads using fixed action names and SigV4 authentication
- write risk: typed CloudTrail trail, channel, dashboard, event-data-store, import, logging, federation, selector, tagging, and resource-policy actions only
- approval: reverse ETL writes require plan, preview, explicit approval, and destructive confirmation where declared
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- AWS CloudTrail fixed-action reads, provider queries, and approval-gated administration actions.
- Usage: pm aws-cloudtrail <command> [flags] --credential <name> [--json]
- Source CLI: AWS CloudTrail API (https://docs.aws.amazon.com/awscloudtrail/latest/APIReference/API_Operations.html)
- ETL read streams
- Bounded query and lookup operations
- Other Commands
  - read describe-trails - Run the documented AWS CloudTrail DescribeTrails operation through a fixed signed JSON-RPC request. [intent=etl availability=implemented stream=describe_trails]; notes: Runs a fixed CloudTrail JSON-RPC read action; request fields are allow-listed command flags.; flags: --include-shadow-trails, --trail-name-list
  - read get-channel - Run the documented AWS CloudTrail GetChannel operation through a fixed signed JSON-RPC request. [intent=etl availability=implemented stream=get_channel]; notes: Runs a fixed CloudTrail JSON-RPC read action; request fields are allow-listed command flags.; flags: --channel
  - read get-dashboard - Run the documented AWS CloudTrail GetDashboard operation through a fixed signed JSON-RPC request. [intent=etl availability=implemented stream=get_dashboard]; notes: Runs a fixed CloudTrail JSON-RPC read action; request fields are allow-listed command flags.; flags: --dashboard-id
  - read get-event-configuration - Run the documented AWS CloudTrail GetEventConfiguration operation through a fixed signed JSON-RPC request. [intent=etl availability=implemented stream=get_event_configuration]; notes: Runs a fixed CloudTrail JSON-RPC read action; request fields are allow-listed command flags.; flags: --event-data-store, --trail-name
  - read get-event-data-store - Run the documented AWS CloudTrail GetEventDataStore operation through a fixed signed JSON-RPC request. [intent=etl availability=implemented stream=get_event_data_store]; notes: Runs a fixed CloudTrail JSON-RPC read action; request fields are allow-listed command flags.; flags: --event-data-store
  - read get-event-selectors - Run the documented AWS CloudTrail GetEventSelectors operation through a fixed signed JSON-RPC request. [intent=etl availability=implemented stream=get_event_selectors]; notes: Runs a fixed CloudTrail JSON-RPC read action; request fields are allow-listed command flags.; flags: --trail-name
  - read get-import - Run the documented AWS CloudTrail GetImport operation through a fixed signed JSON-RPC request. [intent=etl availability=implemented stream=get_import]; notes: Runs a fixed CloudTrail JSON-RPC read action; request fields are allow-listed command flags.; flags: --import-id
  - read get-insight-selectors - Run the documented AWS CloudTrail GetInsightSelectors operation through a fixed signed JSON-RPC request. [intent=etl availability=implemented stream=get_insight_selectors]; notes: Runs a fixed CloudTrail JSON-RPC read action; request fields are allow-listed command flags.; flags: --event-data-store, --trail-name
  - read get-resource-policy - Run the documented AWS CloudTrail GetResourcePolicy operation through a fixed signed JSON-RPC request. [intent=etl availability=implemented stream=get_resource_policy]; notes: Runs a fixed CloudTrail JSON-RPC read action; request fields are allow-listed command flags.; flags: --resource-arn
  - read get-trail - Run the documented AWS CloudTrail GetTrail operation through a fixed signed JSON-RPC request. [intent=etl availability=implemented stream=get_trail]; notes: Runs a fixed CloudTrail JSON-RPC read action; request fields are allow-listed command flags.; flags: --name
  - read get-trail-status - Run the documented AWS CloudTrail GetTrailStatus operation through a fixed signed JSON-RPC request. [intent=etl availability=implemented stream=get_trail_status]; notes: Runs a fixed CloudTrail JSON-RPC read action; request fields are allow-listed command flags.; flags: --name
  - read list-channels - Run the documented AWS CloudTrail ListChannels operation through a fixed signed JSON-RPC request. [intent=etl availability=implemented stream=list_channels]; notes: Runs a fixed CloudTrail JSON-RPC read action; request fields are allow-listed command flags.; flags: --max-results, --next-token
  - read list-dashboards - Run the documented AWS CloudTrail ListDashboards operation through a fixed signed JSON-RPC request. [intent=etl availability=implemented stream=list_dashboards]; notes: Runs a fixed CloudTrail JSON-RPC read action; request fields are allow-listed command flags.; flags: --max-results, --name-prefix, --next-token, --type
  - read list-event-data-stores - Run the documented AWS CloudTrail ListEventDataStores operation through a fixed signed JSON-RPC request. [intent=etl availability=implemented stream=list_event_data_stores]; notes: Runs a fixed CloudTrail JSON-RPC read action; request fields are allow-listed command flags.; flags: --max-results, --next-token
  - read list-import-failures - Run the documented AWS CloudTrail ListImportFailures operation through a fixed signed JSON-RPC request. [intent=etl availability=implemented stream=list_import_failures]; notes: Runs a fixed CloudTrail JSON-RPC read action; request fields are allow-listed command flags.; flags: --import-id, --max-results, --next-token
  - read list-imports - Run the documented AWS CloudTrail ListImports operation through a fixed signed JSON-RPC request. [intent=etl availability=implemented stream=list_imports]; notes: Runs a fixed CloudTrail JSON-RPC read action; request fields are allow-listed command flags.; flags: --destination, --import-status, --max-results, --next-token
  - read list-public-keys - Run the documented AWS CloudTrail ListPublicKeys operation through a fixed signed JSON-RPC request. [intent=etl availability=implemented stream=list_public_keys]; notes: Runs a fixed CloudTrail JSON-RPC read action; request fields are allow-listed command flags.; flags: --end-time, --next-token, --start-time
  - read list-tags - Run the documented AWS CloudTrail ListTags operation through a fixed signed JSON-RPC request. [intent=etl availability=implemented stream=list_tags]; notes: Runs a fixed CloudTrail JSON-RPC read action; request fields are allow-listed command flags.; flags: --next-token, --resource-id-list
  - read list-trails - Run the documented AWS CloudTrail ListTrails operation through a fixed signed JSON-RPC request. [intent=etl availability=implemented stream=list_trails]; notes: Runs a fixed CloudTrail JSON-RPC read action; request fields are allow-listed command flags.; flags: --next-token
  - query cancel - Run the documented AWS CloudTrail CancelQuery operation through a fixed signed JSON-RPC request. [intent=direct_read availability=implemented]; approval: none: typed bounded CloudTrail provider query/lookup operation; risk: high; notes: No raw query text escape hatch is exposed beyond the documented CloudTrail QueryStatement field on StartQuery.; flags: --event-data-store, --event-data-store-owner-account-id, --query-id
  - query describe - Run the documented AWS CloudTrail DescribeQuery operation through a fixed signed JSON-RPC request. [intent=direct_read availability=implemented]; approval: none: typed bounded CloudTrail provider query/lookup operation; risk: medium; notes: No raw query text escape hatch is exposed beyond the documented CloudTrail QueryStatement field on StartQuery.; flags: --event-data-store, --event-data-store-owner-account-id, --query-alias, --query-id, --refresh-id
  - query generate - Run the documented AWS CloudTrail GenerateQuery operation through a fixed signed JSON-RPC request. [intent=direct_read availability=implemented]; approval: none: typed bounded CloudTrail provider query/lookup operation; risk: high; notes: No raw query text escape hatch is exposed beyond the documented CloudTrail QueryStatement field on StartQuery.; flags: --event-data-stores, --prompt
  - query results - Run the documented AWS CloudTrail GetQueryResults operation through a fixed signed JSON-RPC request. [intent=direct_read availability=implemented]; approval: none: typed bounded CloudTrail provider query/lookup operation; risk: medium; notes: No raw query text escape hatch is exposed beyond the documented CloudTrail QueryStatement field on StartQuery.; flags: --event-data-store, --event-data-store-owner-account-id, --max-query-results, --next-token, --query-id
  - insights data - Run the documented AWS CloudTrail ListInsightsData operation through a fixed signed JSON-RPC request. [intent=direct_read availability=implemented]; approval: none: typed bounded CloudTrail provider query/lookup operation; risk: medium; notes: No raw query text escape hatch is exposed beyond the documented CloudTrail QueryStatement field on StartQuery.; flags: --data-type, --dimensions, --end-time, --insight-source, --max-results, --next-token, --start-time
  - insights metric-data - Run the documented AWS CloudTrail ListInsightsMetricData operation through a fixed signed JSON-RPC request. [intent=direct_read availability=implemented]; approval: none: typed bounded CloudTrail provider query/lookup operation; risk: medium; notes: No raw query text escape hatch is exposed beyond the documented CloudTrail QueryStatement field on StartQuery.; flags: --data-type, --end-time, --error-code, --event-name, --event-source, --insight-type, --max-results, --next-token, --period, --start-time, --trail-name
  - query list - Run the documented AWS CloudTrail ListQueries operation through a fixed signed JSON-RPC request. [intent=direct_read availability=implemented]; approval: none: typed bounded CloudTrail provider query/lookup operation; risk: medium; notes: No raw query text escape hatch is exposed beyond the documented CloudTrail QueryStatement field on StartQuery.; flags: --end-time, --event-data-store, --max-results, --next-token, --query-status, --start-time
  - events lookup - Run the documented AWS CloudTrail LookupEvents operation through a fixed signed JSON-RPC request. [intent=direct_read availability=implemented]; approval: none: typed bounded CloudTrail provider query/lookup operation; risk: medium; notes: No raw query text escape hatch is exposed beyond the documented CloudTrail QueryStatement field on StartQuery.; flags: --end-time, --event-category, --lookup-attributes, --max-results, --next-token, --start-time
  - sample-queries search - Run the documented AWS CloudTrail SearchSampleQueries operation through a fixed signed JSON-RPC request. [intent=direct_read availability=implemented]; approval: none: typed bounded CloudTrail provider query/lookup operation; risk: medium; notes: No raw query text escape hatch is exposed beyond the documented CloudTrail QueryStatement field on StartQuery.; flags: --max-results, --next-token, --search-phrase
  - query start - Run the documented AWS CloudTrail StartQuery operation through a fixed signed JSON-RPC request. [intent=direct_read availability=implemented]; approval: none: typed bounded CloudTrail provider query/lookup operation; risk: high; notes: No raw query text escape hatch is exposed beyond the documented CloudTrail QueryStatement field on StartQuery.; flags: --delivery-s3-uri, --event-data-store-owner-account-id, --query-alias, --query-parameters, --query-statement
- Help topics:
  - aws-cloudtrail - AWS CloudTrail connector command surface and safety contract.
  - aws-cloudtrail writes - Use pm reverse plan/preview/run with aws-cloudtrail write actions; direct provider shortcuts are intentionally not exposed.

## Commands

### Inspect as a manual

```bash
pm connectors inspect aws-cloudtrail
```

### Inspect as structured JSON

```bash
pm connectors inspect aws-cloudtrail --json
```

## Agent Rules

- Run pm connectors inspect aws-cloudtrail before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
