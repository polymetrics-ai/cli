---
name: pm-aws-cloudtrail
description: AWS CloudTrail connector knowledge and safe action guide.
---

# pm-aws-cloudtrail

## Purpose

Reads and safely operates AWS CloudTrail trails, event data stores, channels, dashboards, and Lake queries through fixed AWS JSON-RPC streams, typed direct-read commands, and typed reverse-ETL write/admin actions. Only StartQuery, CreateDashboard, and UpdateDashboard stay blocked: each requires an unrestricted CloudTrail Lake SQL QueryStatement, which this project disables for every connector by policy. Breaking change: the earlier LookupEvents-backed management_events, read_only_events, write_only_events, and console_logins streams are removed and have no replacement here; use the events lookup direct-read command for CloudTrail event lookups instead. See the connector docs migration note.

## Icon

- id: aws-cloudtrail
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

## Reverse ETL Actions

- add_tags:
  - endpoint: POST CloudTrail_20131101.AddTags
  - required fields: ResourceId, TagsList
  - risk: adds tags to a CloudTrail trail, event data store, channel, or dashboard
- cancel_query:
  - endpoint: POST CloudTrail_20131101.CancelQuery
  - required fields: EventDataStore, QueryId
  - risk: cancels a running CloudTrail Lake query
- create_channel:
  - endpoint: POST CloudTrail_20131101.CreateChannel
  - required fields: Destinations, Name, Source
  - risk: creates a CloudTrail organization/service event channel
- create_event_data_store:
  - endpoint: POST CloudTrail_20131101.CreateEventDataStore
  - required fields: Name
  - risk: creates a CloudTrail Lake event data store
- create_trail:
  - endpoint: POST CloudTrail_20131101.CreateTrail
  - required fields: Name, S3BucketName
  - risk: creates a CloudTrail trail that delivers log files to an S3 bucket
- delete_channel:
  - endpoint: POST CloudTrail_20131101.DeleteChannel
  - required fields: Channel
  - risk: deletes a CloudTrail event channel
- delete_dashboard:
  - endpoint: POST CloudTrail_20131101.DeleteDashboard
  - required fields: DashboardId
  - risk: deletes a CloudTrail Lake dashboard
- delete_event_data_store:
  - endpoint: POST CloudTrail_20131101.DeleteEventDataStore
  - required fields: EventDataStore
  - risk: deletes a CloudTrail Lake event data store
- delete_resource_policy:
  - endpoint: POST CloudTrail_20131101.DeleteResourcePolicy
  - required fields: ResourceArn
  - risk: deletes the resource-based policy attached to a CloudTrail channel or event data store
- delete_trail:
  - endpoint: POST CloudTrail_20131101.DeleteTrail
  - required fields: Name
  - risk: deletes a CloudTrail trail
- deregister_organization_delegated_admin:
  - endpoint: POST CloudTrail_20131101.DeregisterOrganizationDelegatedAdmin
  - required fields: DelegatedAdminAccountId
  - risk: removes an AWS account's CloudTrail delegated administrator status for the organization
- disable_federation:
  - endpoint: POST CloudTrail_20131101.DisableFederation
  - required fields: EventDataStore
  - risk: disables Lake query federation for a CloudTrail event data store
- enable_federation:
  - endpoint: POST CloudTrail_20131101.EnableFederation
  - required fields: EventDataStore, FederationRoleArn
  - risk: enables Lake query federation for a CloudTrail event data store
- put_event_configuration:
  - endpoint: POST CloudTrail_20131101.PutEventConfiguration
  - risk: sets the maximum event size and context key selectors for a trail or event data store
- put_event_selectors:
  - endpoint: POST CloudTrail_20131101.PutEventSelectors
  - required fields: TrailName
  - risk: sets basic or advanced event selectors on a CloudTrail trail
- put_insight_selectors:
  - endpoint: POST CloudTrail_20131101.PutInsightSelectors
  - required fields: InsightSelectors
  - risk: enables CloudTrail Insights event types on a trail or event data store
- put_resource_policy:
  - endpoint: POST CloudTrail_20131101.PutResourcePolicy
  - required fields: ResourceArn, ResourcePolicy
  - risk: attaches a resource-based IAM policy document to a CloudTrail channel or event data store
- register_organization_delegated_admin:
  - endpoint: POST CloudTrail_20131101.RegisterOrganizationDelegatedAdmin
  - required fields: MemberAccountId
  - risk: grants an AWS account CloudTrail delegated administrator status for the organization
- remove_tags:
  - endpoint: POST CloudTrail_20131101.RemoveTags
  - required fields: ResourceId, TagsList
  - risk: removes tags from a CloudTrail trail, event data store, channel, or dashboard
- restore_event_data_store:
  - endpoint: POST CloudTrail_20131101.RestoreEventDataStore
  - required fields: EventDataStore
  - risk: restores a soft-deleted CloudTrail Lake event data store
- start_dashboard_refresh:
  - endpoint: POST CloudTrail_20131101.StartDashboardRefresh
  - required fields: DashboardId
  - risk: starts an on-demand refresh of a CloudTrail Lake dashboard
- start_event_data_store_ingestion:
  - endpoint: POST CloudTrail_20131101.StartEventDataStoreIngestion
  - required fields: EventDataStore
  - risk: resumes event ingestion into a stopped CloudTrail Lake event data store
- start_import:
  - endpoint: POST CloudTrail_20131101.StartImport
  - risk: starts or resumes a CloudTrail Lake import job from an S3 source
- start_logging:
  - endpoint: POST CloudTrail_20131101.StartLogging
  - required fields: Name
  - risk: starts CloudTrail log delivery for a trail
- stop_event_data_store_ingestion:
  - endpoint: POST CloudTrail_20131101.StopEventDataStoreIngestion
  - required fields: EventDataStore
  - risk: stops event ingestion into a CloudTrail Lake event data store
- stop_import:
  - endpoint: POST CloudTrail_20131101.StopImport
  - required fields: ImportId
  - risk: stops an in-progress CloudTrail Lake import job
- stop_logging:
  - endpoint: POST CloudTrail_20131101.StopLogging
  - required fields: Name
  - risk: stops CloudTrail log delivery for a trail
- update_channel:
  - endpoint: POST CloudTrail_20131101.UpdateChannel
  - required fields: Channel
  - risk: updates a CloudTrail event channel's name or destinations
- update_event_data_store:
  - endpoint: POST CloudTrail_20131101.UpdateEventDataStore
  - required fields: EventDataStore
  - risk: updates a CloudTrail Lake event data store's configuration
- update_trail:
  - endpoint: POST CloudTrail_20131101.UpdateTrail
  - required fields: Name
  - risk: updates a CloudTrail trail's configuration

## Security

- read risk: bounded AWS CloudTrail JSON-RPC reads using fixed action names, SigV4 authentication, and connector-local resource discovery for parameterized streams
- write risk: typed AWS CloudTrail JSON-RPC write/admin actions using fixed action names, closed per-action request schemas, SigV4 authentication, and provider-supported idempotency for destructive actions; StartQuery, CreateDashboard, and UpdateDashboard stay blocked because their request body requires an unrestricted CloudTrail Lake SQL QueryStatement, which this project disables for every connector by policy.
- approval: All CloudTrail writes require plan -> preview -> explicit approval -> execute; destructive writes also require confirm: destructive.
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Inspect and safely operate AWS CloudTrail trails, event data stores, channels, dashboards, and Lake queries through fixed typed commands.
- Usage: pm aws-cloudtrail <command> [options]
- Source CLI: AWS CLI cloudtrail (https://docs.aws.amazon.com/awscloudtrail/latest/APIReference/API_Operations.html)
- Read
- Typed write/admin actions
- Other Commands
  - trails describe - Read the describe_trails stream (AWS CloudTrail DescribeTrails). [intent=etl availability=implemented stream=describe_trails]
  - channel get - Read the get_channel stream (AWS CloudTrail GetChannel). [intent=etl availability=implemented stream=get_channel]
  - dashboard get - Read the get_dashboard stream (AWS CloudTrail GetDashboard). [intent=etl availability=implemented stream=get_dashboard]
  - event-configuration get - Read the get_event_configuration stream (AWS CloudTrail GetEventConfiguration). [intent=etl availability=implemented stream=get_event_configuration]
  - event-data-store get - Read the get_event_data_store stream (AWS CloudTrail GetEventDataStore). [intent=etl availability=implemented stream=get_event_data_store]
  - event-selectors get - Read the get_event_selectors stream (AWS CloudTrail GetEventSelectors). [intent=etl availability=implemented stream=get_event_selectors]
  - import get - Read the get_import stream (AWS CloudTrail GetImport). [intent=etl availability=implemented stream=get_import]
  - insight-selectors get - Read the get_insight_selectors stream (AWS CloudTrail GetInsightSelectors). [intent=etl availability=implemented stream=get_insight_selectors]
  - resource-policy get - Read the get_resource_policy stream (AWS CloudTrail GetResourcePolicy). [intent=etl availability=implemented stream=get_resource_policy]
  - trail get - Read the get_trail stream (AWS CloudTrail GetTrail). [intent=etl availability=implemented stream=get_trail]
  - trail status - Read the get_trail_status stream (AWS CloudTrail GetTrailStatus). [intent=etl availability=implemented stream=get_trail_status]
  - channel list - Read the list_channels stream (AWS CloudTrail ListChannels). [intent=etl availability=implemented stream=list_channels]
  - dashboard list - Read the list_dashboards stream (AWS CloudTrail ListDashboards). [intent=etl availability=implemented stream=list_dashboards]
  - event-data-store list - Read the list_event_data_stores stream (AWS CloudTrail ListEventDataStores). [intent=etl availability=implemented stream=list_event_data_stores]
  - import failures list - Read the list_import_failures stream (AWS CloudTrail ListImportFailures). [intent=etl availability=implemented stream=list_import_failures]
  - import list - Read the list_imports stream (AWS CloudTrail ListImports). [intent=etl availability=implemented stream=list_imports]
  - public-keys list - Read the list_public_keys stream (AWS CloudTrail ListPublicKeys). [intent=etl availability=implemented stream=list_public_keys]
  - tags list - Read the list_tags stream (AWS CloudTrail ListTags). [intent=etl availability=implemented stream=list_tags]
  - trail list - Read the list_trails stream (AWS CloudTrail ListTrails). [intent=etl availability=implemented stream=list_trails]
  - query cancel - Cancel a running CloudTrail Lake query by query id. [intent=reverse_etl availability=implemented write=cancel_query]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: cancels a running CloudTrail Lake query; flags: --event-data-store (required), --event-data-store-owner-account-id, --query-id (required)
  - query describe - Describe a CloudTrail Lake query's status and metadata by query id or alias. [intent=direct_read availability=implemented operation=describe_query]; flags: --event-data-store, --event-data-store-owner-account-id, --query-alias, --query-id, --refresh-id
  - query generate - Generate a candidate CloudTrail Lake SQL query from a natural-language prompt; does not execute the query. [intent=direct_read availability=implemented operation=generate_query]; flags: --event-data-stores (required), --prompt (required)
  - query results - Fetch bounded result rows for a previously completed CloudTrail Lake query by query id. [intent=direct_read availability=implemented operation=get_query_results]; flags: --event-data-store, --event-data-store-owner-account-id, --max-query-results, --next-token, --query-id (required)
  - insights data list - List CloudTrail Insights event records for a configured insight source. [intent=direct_read availability=implemented operation=list_insights_data]; flags: --data-type (required), --end-time, --insight-source (required), --max-results, --next-token, --start-time
  - insights metric-data list - List CloudTrail Insights metric data points for an event name/source/insight type. [intent=direct_read availability=implemented operation=list_insights_metric_data]; flags: --data-type, --end-time, --error-code, --event-name (required), --event-source (required), --insight-type (required), --max-results, --next-token, --period, --start-time, --trail-name
  - query list - List CloudTrail Lake queries run against an event data store. [intent=direct_read availability=implemented operation=list_queries]; flags: --end-time, --event-data-store (required), --max-results, --next-token, --query-status, --start-time
  - events lookup - Look up CloudTrail management/data events using typed attribute filters and a time range. [intent=direct_read availability=implemented operation=lookup_events]; flags: --end-time, --event-category, --lookup-attribute-key, --lookup-attribute-value, --max-results, --next-token, --start-time
  - query sample-search - Search AWS-provided sample CloudTrail Lake queries by keyword phrase. [intent=direct_read availability=implemented operation=search_sample_queries]; flags: --max-results, --next-token, --search-phrase (required)
  - tags add - Plan a typed AWS CloudTrail AddTags operation. [intent=reverse_etl availability=implemented write=add_tags]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: adds tags to a CloudTrail trail, event data store, channel, or dashboard; flags: --resource-id (required), --tag-key (required), --tag-value
  - channel create - Plan a typed AWS CloudTrail CreateChannel operation. [intent=reverse_etl availability=implemented write=create_channel]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: creates a CloudTrail organization/service event channel; flags: --destination-location (required), --destination-type (required), --name (required), --source (required)
  - event-data-store create - Plan a typed AWS CloudTrail CreateEventDataStore operation. [intent=reverse_etl availability=implemented write=create_event_data_store]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: creates a CloudTrail Lake event data store; flags: --billing-mode, --kms-key-id, --multi-region-enabled, --name (required), --advanced-event-selector-name, --advanced-event-field, --advanced-event-ends-with, --advanced-event-equals, --advanced-event-not-ends-with, --advanced-event-not-equals, --advanced-event-not-starts-with, --advanced-event-starts-with, --organization-enabled, --retention-period, --start-ingestion, --termination-protection-enabled
  - trail create - Plan a typed AWS CloudTrail CreateTrail operation. [intent=reverse_etl availability=implemented write=create_trail]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: creates a CloudTrail trail that delivers log files to an S3 bucket; flags: --cloud-watch-logs-log-group-arn, --cloud-watch-logs-role-arn, --enable-log-file-validation, --include-global-service-events, --is-multi-region-trail, --is-organization-trail, --kms-key-id, --name (required), --s3-bucket-name (required), --s3-key-prefix, --sns-topic-name
  - channel delete - Plan a typed AWS CloudTrail DeleteChannel operation. [intent=reverse_etl availability=implemented write=delete_channel]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; destructive actions also require confirm: destructive; risk: deletes a CloudTrail event channel; flags: --channel (required)
  - dashboard delete - Plan a typed AWS CloudTrail DeleteDashboard operation. [intent=reverse_etl availability=implemented write=delete_dashboard]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; destructive actions also require confirm: destructive; risk: deletes a CloudTrail Lake dashboard; flags: --dashboard-id (required)
  - event-data-store delete - Plan a typed AWS CloudTrail DeleteEventDataStore operation. [intent=reverse_etl availability=implemented write=delete_event_data_store]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; destructive actions also require confirm: destructive; risk: deletes a CloudTrail Lake event data store; flags: --event-data-store (required)
  - resource-policy delete - Plan a typed AWS CloudTrail DeleteResourcePolicy operation. [intent=reverse_etl availability=implemented write=delete_resource_policy]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; destructive actions also require confirm: destructive; risk: deletes the resource-based policy attached to a CloudTrail channel or event data store; flags: --resource-arn (required)
  - trail delete - Plan a typed AWS CloudTrail DeleteTrail operation. [intent=reverse_etl availability=implemented write=delete_trail]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; destructive actions also require confirm: destructive; risk: deletes a CloudTrail trail; flags: --name (required)
  - org delegated-admin deregister - Plan a typed AWS CloudTrail DeregisterOrganizationDelegatedAdmin operation. [intent=reverse_etl availability=implemented write=deregister_organization_delegated_admin]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: removes an AWS account's CloudTrail delegated administrator status for the organization; flags: --delegated-admin-account-id (required)
  - federation disable - Plan a typed AWS CloudTrail DisableFederation operation. [intent=reverse_etl availability=implemented write=disable_federation]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; destructive actions also require confirm: destructive; risk: disables Lake query federation for a CloudTrail event data store; flags: --event-data-store (required)
  - federation enable - Plan a typed AWS CloudTrail EnableFederation operation. [intent=reverse_etl availability=implemented write=enable_federation]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: enables Lake query federation for a CloudTrail event data store; flags: --event-data-store (required), --federation-role-arn (required)
  - event-configuration set - Plan a typed AWS CloudTrail PutEventConfiguration operation. [intent=reverse_etl availability=implemented write=put_event_configuration]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: sets the maximum event size and context key selectors for a trail or event data store; flags: --event-data-store, --max-event-size, --aggregation-event-category, --aggregation-template, --context-key-type, --context-key-equals, --trail-name
  - event-selectors set - Plan a typed AWS CloudTrail PutEventSelectors operation. [intent=reverse_etl availability=implemented write=put_event_selectors]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: sets basic or advanced event selectors on a CloudTrail trail; flags: --trail-name (required), --advanced-event-selector-name, --advanced-event-field, --advanced-event-ends-with, --advanced-event-equals, --advanced-event-not-ends-with, --advanced-event-not-equals, --advanced-event-not-starts-with, --advanced-event-starts-with, --event-selector-data-resource-type, --event-selector-data-resource-value, --event-selector-exclude-management-event-source, --event-selector-include-management-events, --event-selector-read-write-type
  - insight-selectors set - Plan a typed AWS CloudTrail PutInsightSelectors operation. [intent=reverse_etl availability=implemented write=put_insight_selectors]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: enables CloudTrail Insights event types on a trail or event data store; flags: --event-data-store, --insights-destination, --disable-insights, --insight-type, --insight-event-category, --trail-name
  - resource-policy set - Plan a typed AWS CloudTrail PutResourcePolicy operation. [intent=reverse_etl availability=implemented write=put_resource_policy]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: attaches a resource-based IAM policy document to a CloudTrail channel or event data store; flags: --resource-arn (required), --resource-policy (required)
  - org delegated-admin register - Plan a typed AWS CloudTrail RegisterOrganizationDelegatedAdmin operation. [intent=reverse_etl availability=implemented write=register_organization_delegated_admin]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: grants an AWS account CloudTrail delegated administrator status for the organization; flags: --member-account-id (required)
  - tags remove - Plan a typed AWS CloudTrail RemoveTags operation. [intent=reverse_etl availability=implemented write=remove_tags]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; destructive actions also require confirm: destructive; risk: removes tags from a CloudTrail trail, event data store, channel, or dashboard; flags: --resource-id (required), --tag-key (required), --tag-value
  - event-data-store restore - Plan a typed AWS CloudTrail RestoreEventDataStore operation. [intent=reverse_etl availability=implemented write=restore_event_data_store]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: restores a soft-deleted CloudTrail Lake event data store; flags: --event-data-store (required)
  - dashboard refresh - Plan a typed AWS CloudTrail StartDashboardRefresh operation. [intent=reverse_etl availability=implemented write=start_dashboard_refresh]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: starts an on-demand refresh of a CloudTrail Lake dashboard; flags: --dashboard-id (required), --query-start-time, --query-end-time, --query-period, --event-data-store-id
  - event-data-store ingestion start - Plan a typed AWS CloudTrail StartEventDataStoreIngestion operation. [intent=reverse_etl availability=implemented write=start_event_data_store_ingestion]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: resumes event ingestion into a stopped CloudTrail Lake event data store; flags: --event-data-store (required)
  - import start - Plan a typed AWS CloudTrail StartImport operation. [intent=reverse_etl availability=implemented write=start_import]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: starts or resumes a CloudTrail Lake import job from an S3 source; flags: --destinations, --end-event-time, --import-id, --import-source-s3-bucket-access-role-arn, --import-source-s3-bucket-region, --import-source-s3-location-uri, --start-event-time
  - trail logging start - Plan a typed AWS CloudTrail StartLogging operation. [intent=reverse_etl availability=implemented write=start_logging]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: starts CloudTrail log delivery for a trail; flags: --name (required)
  - event-data-store ingestion stop - Plan a typed AWS CloudTrail StopEventDataStoreIngestion operation. [intent=reverse_etl availability=implemented write=stop_event_data_store_ingestion]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; destructive actions also require confirm: destructive; risk: stops event ingestion into a CloudTrail Lake event data store; flags: --event-data-store (required)
  - import stop - Plan a typed AWS CloudTrail StopImport operation. [intent=reverse_etl availability=implemented write=stop_import]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; destructive actions also require confirm: destructive; risk: stops an in-progress CloudTrail Lake import job; flags: --import-id (required)
  - trail logging stop - Plan a typed AWS CloudTrail StopLogging operation. [intent=reverse_etl availability=implemented write=stop_logging]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; destructive actions also require confirm: destructive; risk: stops CloudTrail log delivery for a trail; flags: --name (required)
  - channel update - Plan a typed AWS CloudTrail UpdateChannel operation. [intent=reverse_etl availability=implemented write=update_channel]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: updates a CloudTrail event channel's name or destinations; flags: --channel (required), --destination-location, --destination-type, --name
  - event-data-store update - Plan a typed AWS CloudTrail UpdateEventDataStore operation. [intent=reverse_etl availability=implemented write=update_event_data_store]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: updates a CloudTrail Lake event data store's configuration; flags: --billing-mode, --event-data-store (required), --advanced-event-selector-name, --advanced-event-field, --advanced-event-ends-with, --advanced-event-equals, --advanced-event-not-ends-with, --advanced-event-not-equals, --advanced-event-not-starts-with, --advanced-event-starts-with, --kms-key-id, --multi-region-enabled, --name, --organization-enabled, --retention-period, --termination-protection-enabled
  - trail update - Plan a typed AWS CloudTrail UpdateTrail operation. [intent=reverse_etl availability=implemented write=update_trail]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: updates a CloudTrail trail's configuration; flags: --cloud-watch-logs-log-group-arn, --cloud-watch-logs-role-arn, --enable-log-file-validation, --include-global-service-events, --is-multi-region-trail, --is-organization-trail, --kms-key-id, --name (required), --s3-bucket-name, --s3-key-prefix, --sns-topic-name
  - query start - AWS CloudTrail StartQuery (not exposed). [intent=direct_read availability=unsafe_or_disallowed]; notes: StartQuery accepts an unrestricted CloudTrail Lake SQL QueryStatement as input. This project disables generic/unrestricted query-text execution for every connector (capabilities.query is always false) and never exposes a raw SQL passthrough; see AGENTS.md. StartQuery therefore stays unavailable regardless of shared-runtime state.
  - dashboard create - AWS CloudTrail CreateDashboard (not exposed). [intent=reverse_etl availability=unsafe_or_disallowed]; notes: CreateDashboard requires a Widgets array where each RequestWidget embeds an unrestricted CloudTrail Lake SQL QueryStatement. This project disables generic/unrestricted query-text execution for every connector (capabilities.query is always false) and never exposes a raw SQL passthrough; see AGENTS.md. CreateDashboard therefore stays unavailable regardless of shared-runtime state.
  - dashboard update - AWS CloudTrail UpdateDashboard (not exposed). [intent=reverse_etl availability=unsafe_or_disallowed]; notes: UpdateDashboard requires a Widgets array where each RequestWidget embeds an unrestricted CloudTrail Lake SQL QueryStatement. This project disables generic/unrestricted query-text execution for every connector (capabilities.query is always false) and never exposes a raw SQL passthrough; see AGENTS.md. UpdateDashboard therefore stays unavailable regardless of shared-runtime state.

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
