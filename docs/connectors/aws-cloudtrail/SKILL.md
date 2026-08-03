---
name: pm-aws-cloudtrail
description: AWS CloudTrail connector knowledge and safe action guide.
---

# pm-aws-cloudtrail

## Purpose

Reads AWS CloudTrail configuration and resource metadata through fixed AWS JSON-RPC streams. Provider query/direct-read and write/admin actions remain planned until shared promoted-native forwarding exposes them safely at runtime. Breaking change: the earlier LookupEvents-backed management_events, read_only_events, write_only_events, and console_logins streams are removed and have no replacement here; CloudTrail event and Insights record reads are blocked/planned. See the connector docs migration note.

## Icon

- asset: icons/aws-cloudtrail.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- aws_region_name
- base_url
- max_pages
- mode
- page_size
- aws_key_id (secret)
- aws_secret_key (secret)

## ETL Streams

- describe_trails:
  - primary key: pm_record_id
  - fields: CloudWatchLogsLogGroupArn(), CloudWatchLogsRoleArn(), HasCustomEventSelectors(), HasInsightSelectors(), HomeRegion(), IncludeGlobalServiceEvents(), IsMultiRegionTrail(), IsOrganizationTrail(), KmsKeyId(), LogFileValidationEnabled(), Name(), S3BucketName(), S3KeyPrefix(), SnsTopicARN(), SnsTopicName(), TrailARN(), operation(), pm_record_id()
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
  - fields: ChannelArn(), Destinations(), Name(), Source(), SourceConfig(), operation(), pm_record_id()
- list_dashboards:
  - primary key: pm_record_id
  - fields: DashboardArn(), DashboardId(), Name(), RefreshSchedule(), Status(), Type(), Widgets(), operation(), pm_record_id()
- list_event_data_stores:
  - primary key: pm_record_id
  - fields: AdvancedEventSelectors(), BillingMode(), CreatedTimestamp(), EventDataStoreArn(), KmsKeyId(), MultiRegionEnabled(), Name(), OrganizationEnabled(), RetentionPeriod(), Status(), TerminationProtectionEnabled(), UpdatedTimestamp(), operation(), pm_record_id()
- list_import_failures:
  - primary key: pm_record_id
  - fields: ErrorMessage(), ErrorType(), LastUpdatedTime(), Location(), Status(), operation(), pm_record_id()
- list_imports:
  - primary key: pm_record_id
  - fields: CreatedTimestamp(), Destinations(), EndEventTime(), ImportId(), ImportStatus(), StartEventTime(), UpdatedTimestamp(), operation(), pm_record_id()
- list_public_keys:
  - primary key: pm_record_id
  - fields: Fingerprint(), ValidityEndTime(), ValidityStartTime(), Value(), operation(), pm_record_id()
- list_tags:
  - primary key: pm_record_id
  - fields: ResourceId(), TagsList(), operation(), pm_record_id()
- list_trails:
  - primary key: pm_record_id
  - fields: HomeRegion(), Name(), TrailARN(), operation(), pm_record_id()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: bounded AWS CloudTrail JSON-RPC reads using fixed action names, SigV4 authentication, and connector-local resource discovery for parameterized streams
- write risk: blocked/planned: CloudTrail write/admin actions require typed write metadata plus shared promoted-native command-surface, validation, and dry-run forwarding before they can be safely exposed.
- approval: No CloudTrail writes are exposed in the current connector surface; future writes must preserve plan -> preview -> approval -> execute.
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

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
