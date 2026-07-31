---
name: pm-aws-cloudtrail
description: Use the Polymetrics AWS CloudTrail connector for implemented read-stream ETL and for understanding blocked/planned CloudTrail direct-read and write/admin parity.
---

# AWS CloudTrail connector skill

## Implemented surface

- Connector: `aws-cloudtrail`
- Implemented counts: 19 ETL/read streams, 0 direct-read commands, 0 reverse-ETL write actions.
- Blocked/planned counts: 10 provider query/direct-read actions and 31 write/admin actions.
- Reason blocked: the command surface, manifest write metadata, write validation, dry-run preview, and operation-direct-read exposure require shared promoted-native forwarding that is intentionally not part of this connector-local corrective head.

## Agent Rules

- Never request, print, summarize, or store AWS secret values.
- Add credentials from environment variables or stdin only, for example `--from-env aws_key_id=AWS_ACCESS_KEY_ID` and `--from-env aws_secret_key=AWS_SECRET_ACCESS_KEY`.
- Inspect metadata with `pm connectors inspect aws-cloudtrail --json`; this does not read credentials.
- Use `pm etl catalog`, `pm etl read`, or configured `pm etl run` for implemented read streams.
- Do not use or invent `pm aws-cloudtrail ...` commands in this corrective head; the dynamic connector command surface is blocked/planned.
- Do not attempt CloudTrail reverse-ETL writes. No CloudTrail write/admin action is executable until a future shared-runtime slice preserves plan -> preview -> approval -> execute with typed confirmation metadata.
- Do not expose raw AWS action names, raw paths, raw headers, raw request bodies, shell, SQL, file, or generic HTTP write tools.

## Implemented ETL streams

`describe_trails`, `get_channel`, `get_dashboard`, `get_event_configuration`, `get_event_data_store`, `get_event_selectors`, `get_import`, `get_insight_selectors`, `get_resource_policy`, `get_trail`, `get_trail_status`, `list_channels`, `list_dashboards`, `list_event_data_stores`, `list_import_failures`, `list_imports`, `list_public_keys`, `list_tags`, `list_trails`.

## Blocked/planned operations

Direct/provider query operations: `CancelQuery`, `DescribeQuery`, `GenerateQuery`, `GetQueryResults`, `ListInsightsData`, `ListInsightsMetricData`, `ListQueries`, `LookupEvents`, `SearchSampleQueries`, `StartQuery`.

Write/admin operations: `AddTags`, `CreateChannel`, `CreateDashboard`, `CreateEventDataStore`, `CreateTrail`, `DeleteChannel`, `DeleteDashboard`, `DeleteEventDataStore`, `DeleteResourcePolicy`, `DeleteTrail`, `DeregisterOrganizationDelegatedAdmin`, `DisableFederation`, `EnableFederation`, `PutEventConfiguration`, `PutEventSelectors`, `PutInsightSelectors`, `PutResourcePolicy`, `RegisterOrganizationDelegatedAdmin`, `RemoveTags`, `RestoreEventDataStore`, `StartDashboardRefresh`, `StartEventDataStoreIngestion`, `StartImport`, `StartLogging`, `StopEventDataStoreIngestion`, `StopImport`, `StopLogging`, `UpdateChannel`, `UpdateDashboard`, `UpdateEventDataStore`, `UpdateTrail`.
