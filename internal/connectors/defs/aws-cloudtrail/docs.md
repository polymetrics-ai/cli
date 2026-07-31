# Overview

AWS CloudTrail connector parity was audited from the official AWS CloudTrail API Reference Actions page. The scope-corrected bundle still enumerates all 60 official CloudTrail API actions exactly once, but only the 19 ETL/read stream actions are implemented and runtime-reachable in this connector-local slice. The 10 provider query/direct-read actions and 31 write/admin actions are recorded as blocked/planned in `api_surface.json` because they require shared promoted-native command-surface, manifest, validation, dry-run, and operation-direct-read forwarding that was reverted from this branch. The official CloudTrail event-record contents page documents event record version 1.11 and 31 top-level event fields; those fields are schema payload fields for LookupEvents, not CDC/changefeed operations.

Implemented readable streams: `describe_trails`, `get_channel`, `get_dashboard`, `get_event_configuration`, `get_event_data_store`, `get_event_selectors`, `get_import`, `get_insight_selectors`, `get_resource_policy`, `get_trail`, `get_trail_status`, `list_channels`, `list_dashboards`, `list_event_data_stores`, `list_import_failures`, `list_imports`, `list_public_keys`, `list_tags`, `list_trails`.

Blocked/planned provider query operations: `CancelQuery`, `DescribeQuery`, `GenerateQuery`, `GetQueryResults`, `ListInsightsData`, `ListInsightsMetricData`, `ListQueries`, `LookupEvents`, `SearchSampleQueries`, `StartQuery`.

Blocked/planned write/admin operations: `AddTags`, `CreateChannel`, `CreateDashboard`, `CreateEventDataStore`, `CreateTrail`, `DeleteChannel`, `DeleteDashboard`, `DeleteEventDataStore`, `DeleteResourcePolicy`, `DeleteTrail`, `DeregisterOrganizationDelegatedAdmin`, `DisableFederation`, `EnableFederation`, `PutEventConfiguration`, `PutEventSelectors`, `PutInsightSelectors`, `PutResourcePolicy`, `RegisterOrganizationDelegatedAdmin`, `RemoveTags`, `RestoreEventDataStore`, `StartDashboardRefresh`, `StartEventDataStoreIngestion`, `StartImport`, `StartLogging`, `StopEventDataStoreIngestion`, `StopImport`, `StopLogging`, `UpdateChannel`, `UpdateDashboard`, `UpdateEventDataStore`, `UpdateTrail`.

## Auth setup

Use `pm credentials add <name> --connector aws-cloudtrail` and provide secrets only from environment variables or stdin:

- `aws_key_id` (required secret)
- `aws_secret_key` (required secret)
- `aws_region_name` (required config, for example `us-east-1`)

Optional config fields are `base_url` for local fixture endpoints, `page_size`, `max_pages`, `start_date`, and `mode=fixture` for credential-free tests. Do not place AWS secret values in chat, command history, docs, fixtures, or issue comments.

## Streams notes

Every implemented stream uses a fixed AWS CloudTrail JSON-RPC action with SigV4 authentication and no raw action/path/header/body escape hatch. Paginated actions pass bounded `MaxResults` and follow `NextToken` until it is absent or `max_pages` is reached. The connector keeps `query=false` at the metadata layer; CloudTrail provider query operations are blocked/planned until shared runtime forwarding makes operation-direct reads visible safely.

## Write actions & risks

No CloudTrail reverse-ETL write actions are exposed by the scope-corrected runtime surface. The audited write/admin API actions remain blocked/planned in `api_surface.json`; they must not be documented as executable until shared promoted-native forwarding exposes bundle manifests, command surfaces, write validation, and dry-run previews. Any future write slice must preserve the standard plan -> preview -> explicit approval -> execute flow, destructive confirmation metadata, fixed CloudTrail `X-Amz-Target` mapping, and no raw AWS action or request-body escape hatch.

## Known limits

- This work is fixture-only and local-test verified; it does not certify live AWS provider behavior.
- CloudTrail event record fields are parsed as payload/schema fields where CloudTrail returns them, but they are not counted as CDC because AWS does not document a CloudTrail changefeed/subscription API in the audited sources.
- Provider query direct reads are blocked/planned. They do not expose a generic CloudTrail Lake SQL runner.
- Write/admin actions are blocked/planned. No CloudTrail write action is listed as executable in generated catalog/docs/help for this corrective head.
