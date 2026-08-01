# Overview

AWS CloudTrail connector parity was audited from the official AWS CloudTrail API Reference Actions page. The scope-corrected bundle still enumerates all 60 official CloudTrail API actions exactly once, but only the 9 ETL/read stream actions that need no required per-call request fields are implemented and runtime-reachable in this connector-local slice. The 10 provider query/direct-read actions, 10 parameterized read actions, and 31 write/admin actions are recorded as blocked/planned in `api_surface.json` because they require shared promoted-native command-surface, manifest, validation, dry-run, operation-direct-read forwarding, or a typed request-parameter boundary that is outside this corrective head. The official CloudTrail event-record contents page documents event record version 1.11 and 31 top-level event fields; those fields are schema payload fields for LookupEvents, not CDC/changefeed operations.

Implemented readable streams: `describe_trails`, `get_event_configuration`, `get_insight_selectors`, `list_channels`, `list_dashboards`, `list_event_data_stores`, `list_imports`, `list_public_keys`, `list_trails`.

Blocked/planned parameterized read operations: `GetChannel`, `GetDashboard`, `GetEventDataStore`, `GetEventSelectors`, `GetImport`, `GetResourcePolicy`, `GetTrail`, `GetTrailStatus`, `ListImportFailures`, `ListTags`.

Blocked/planned provider query operations: `CancelQuery`, `DescribeQuery`, `GenerateQuery`, `GetQueryResults`, `ListInsightsData`, `ListInsightsMetricData`, `ListQueries`, `LookupEvents`, `SearchSampleQueries`, `StartQuery`.

Blocked/planned write/admin operations: `AddTags`, `CreateChannel`, `CreateDashboard`, `CreateEventDataStore`, `CreateTrail`, `DeleteChannel`, `DeleteDashboard`, `DeleteEventDataStore`, `DeleteResourcePolicy`, `DeleteTrail`, `DeregisterOrganizationDelegatedAdmin`, `DisableFederation`, `EnableFederation`, `PutEventConfiguration`, `PutEventSelectors`, `PutInsightSelectors`, `PutResourcePolicy`, `RegisterOrganizationDelegatedAdmin`, `RemoveTags`, `RestoreEventDataStore`, `StartDashboardRefresh`, `StartEventDataStoreIngestion`, `StartImport`, `StartLogging`, `StopEventDataStoreIngestion`, `StopImport`, `StopLogging`, `UpdateChannel`, `UpdateDashboard`, `UpdateEventDataStore`, `UpdateTrail`.

## Auth setup

Use `pm credentials add <name> --connector aws-cloudtrail` and provide secrets only from environment variables or stdin:

- `aws_key_id` (required secret)
- `aws_secret_key` (required secret)
- `aws_region_name` (required config, for example `us-east-1`)

Optional config fields are `base_url` for local fixture endpoints, `page_size`, `max_pages`, and `mode=fixture` for credential-free tests. Do not place AWS secret values in chat, command history, docs, fixtures, or issue comments.

## Streams notes

Every implemented stream uses a fixed AWS CloudTrail JSON-RPC action with SigV4 authentication and no raw action/path/header/body escape hatch. Paginated actions pass bounded `MaxResults` and follow `NextToken` until it is absent or `max_pages` is reached. Actions that require resource identifiers, query identifiers, import identifiers, tag resource lists, or other per-call request fields are blocked/planned until a safe typed boundary supplies those fields. The connector keeps `query=false` at the metadata layer; CloudTrail provider query operations are blocked/planned until shared runtime forwarding makes operation-direct reads visible safely.

## Write actions & risks

No CloudTrail reverse-ETL write actions are exposed by the scope-corrected runtime surface. The audited write/admin API actions remain blocked/planned in `api_surface.json`; they must not be documented as executable until shared promoted-native forwarding exposes bundle manifests, command surfaces, write validation, and dry-run previews. Any future write slice must preserve the standard plan -> preview -> explicit approval -> execute flow, destructive confirmation metadata, fixed CloudTrail `X-Amz-Target` mapping, and no raw AWS action or request-body escape hatch.

## Known limits

- This work is fixture-only and local-test verified; it does not certify live AWS provider behavior.
- CloudTrail event record fields are parsed as payload/schema fields where CloudTrail returns them, but they are not counted as CDC because AWS does not document a CloudTrail changefeed/subscription API in the audited sources.
- Provider query direct reads are blocked/planned. They do not expose a generic CloudTrail Lake SQL runner.
- Parameterized reads are blocked/planned until a typed command, fan-out, or config boundary can supply the required per-call request fields.
- Write/admin actions are blocked/planned. No CloudTrail write action is listed as executable in generated catalog/docs/help for this corrective head.
