# Overview

AWS CloudTrail connector parity was audited from the official AWS CloudTrail API Reference Actions page. The bundle enumerates all 60 official CloudTrail API actions exactly once, and 19 ETL/read stream actions are implemented and runtime-reachable in the connector-local surface. Streams whose AWS read action requires an identifier use connector-local discovery/fan-out from the fixed list/describe streams rather than caller-supplied raw action/path/header/body input. The 10 provider query/direct-read actions and 31 write/admin actions are recorded as blocked/planned in `api_surface.json` because they require typed operation/write metadata plus shared promoted-native command-surface, write-validation, dry-run, and operation-direct-read forwarding outside this connector-local implementation. The official CloudTrail event-record contents page documents event record version 1.11 and 31 top-level event fields; those fields are schema payload fields for LookupEvents, not CDC/changefeed operations.

Implemented readable streams: `describe_trails`, `get_channel`, `get_dashboard`, `get_event_configuration`, `get_event_data_store`, `get_event_selectors`, `get_import`, `get_insight_selectors`, `get_resource_policy`, `get_trail`, `get_trail_status`, `list_channels`, `list_dashboards`, `list_event_data_stores`, `list_import_failures`, `list_imports`, `list_public_keys`, `list_tags`, `list_trails`.

Blocked/planned provider query operations: `CancelQuery`, `DescribeQuery`, `GenerateQuery`, `GetQueryResults`, `ListInsightsData`, `ListInsightsMetricData`, `ListQueries`, `LookupEvents`, `SearchSampleQueries`, `StartQuery`.

Blocked/planned write/admin operations: `AddTags`, `CreateChannel`, `CreateDashboard`, `CreateEventDataStore`, `CreateTrail`, `DeleteChannel`, `DeleteDashboard`, `DeleteEventDataStore`, `DeleteResourcePolicy`, `DeleteTrail`, `DeregisterOrganizationDelegatedAdmin`, `DisableFederation`, `EnableFederation`, `PutEventConfiguration`, `PutEventSelectors`, `PutInsightSelectors`, `PutResourcePolicy`, `RegisterOrganizationDelegatedAdmin`, `RemoveTags`, `RestoreEventDataStore`, `StartDashboardRefresh`, `StartEventDataStoreIngestion`, `StartImport`, `StartLogging`, `StopEventDataStoreIngestion`, `StopImport`, `StopLogging`, `UpdateChannel`, `UpdateDashboard`, `UpdateEventDataStore`, `UpdateTrail`.

## Breaking change and migration

The earlier CloudTrail connector exposed four `LookupEvents`-backed event streams: `management_events`, `read_only_events`, `write_only_events`, and `console_logins`. None of them is an executable read stream in this native implementation, and their schemas are removed. Existing ETL connections that name one of them now fail with `aws-cloudtrail stream <name> not found`, and the incremental `EventTime` cursor those streams provided is gone; all 19 current streams are full-refresh.

`LookupEvents` is now classified in the blocked/planned provider query/direct-read lane in `api_surface.json`, together with `ListInsightsData` and `ListInsightsMetricData`, so CloudTrail event and Insights record reads are not available from this connector at all until that lane is enabled.

There is no drop-in replacement for event-record ETL. The typed CloudTrail operations this connector does implement read configuration and resource metadata, not event records:

- Trail inventory and delivery state: `describe_trails`, `list_trails`, `get_trail`, `get_trail_status`.
- What a trail or event data store is configured to capture: `get_event_selectors`, `get_insight_selectors`, `get_event_configuration`.
- Lake and delivery resources: `list_event_data_stores`, `get_event_data_store`, `list_channels`, `get_channel`, `list_imports`, `get_import`, `list_import_failures`.

Until the query/direct-read lane lands, read CloudTrail event records through an AWS-native path (CloudTrail Lake queries, an S3 log-delivery bucket, or CloudWatch Logs) rather than through this connector.

## Auth setup

Use `pm credentials add <name> --connector aws-cloudtrail` and provide secrets only from environment variables or stdin:

- `aws_key_id` (required secret)
- `aws_secret_key` (required secret)
- `aws_region_name` (required config, for example `us-east-1`)

Optional config fields are `base_url` for local fixture endpoints, `page_size`, `max_pages`, and `mode=fixture` for credential-free tests. Do not place AWS secret values in chat, command history, docs, fixtures, or issue comments.

## Streams notes

Every implemented stream uses a fixed AWS CloudTrail JSON-RPC action with SigV4 authentication and no raw action/path/header/body escape hatch. Paginated actions pass bounded `MaxResults` and follow `NextToken` until it is absent or `max_pages` is reached. Streams for resource-detail actions derive required identifiers through connector-local discovery/fan-out from `DescribeTrails`, `ListChannels`, `ListDashboards`, `ListEventDataStores`, and `ListImports`. The connector keeps `query=false` at the metadata layer; CloudTrail provider query operations are blocked/planned until shared runtime forwarding makes operation-direct reads visible safely.

## Write actions & risks

No CloudTrail reverse-ETL write actions are exposed by the current runtime surface. The audited write/admin API actions remain blocked/planned in `api_surface.json`; they must not be documented as executable until typed write definitions and shared promoted-native command surfaces, write validation, and dry-run previews are available. Any future write slice must preserve the standard plan -> preview -> explicit approval -> execute flow, destructive confirmation metadata, fixed CloudTrail `X-Amz-Target` mapping, and no raw AWS action or request-body escape hatch.

## Generated manual fidelity

`docs/connectors/aws-cloudtrail/MANUAL.md` and `SKILL.md` are written by `pm docs generate` from the same manifest that `pm connectors inspect aws-cloudtrail` renders, so the generated files and that live command (including `pm connectors inspect aws-cloudtrail --json`) are all intentionally sparse right now: no ETL STREAMS section, no SYNC MODES section, `No connector-specific config fields.` under CONFIGURATION, generic `connector-specific` read/write risk instead of the authored risk text in `metadata.json`, and — misleadingly — `No secret authentication is required for this connector.` under AUTHENTICATION. This connector does require the `aws_key_id` and `aws_secret_key` secrets plus the `aws_region_name` config field; the authoritative list is the Auth setup section above and `spec.json`.

That is not a gap in this connector. Those sections are built from `connectors.ManifestOf`, and the shared `definitionConnector` wrapper in `internal/connectors/native/nativeset/promoted.go` currently implements only `Definition()`, so `ManifestOf` falls back to a metadata-only manifest for every bundle-backed promoted native (~30 connectors today). The fix is foundation PR #3676, `fix(connectors): derive nativeset manifest from bundle definition`, which is deliberately owned outside this connector-local branch. Once #3676 merges and this branch rebases onto it, re-running `pm docs generate` fills in the full ETL STREAMS, CONFIGURATION, SYNC MODES, and authored SECURITY sections from this bundle with no change to the connector itself.

Until then, `pm connectors catalog --json`, `pm etl catalog --connector aws-cloudtrail`, the generated connector catalog under `docs/connectors/catalog/`, and the website data read the bundle directly and already report the complete 19-stream, 0-write surface. Use those for CloudTrail stream, config, and risk facts instead of `pm connectors inspect`.

## Known limits

- This work is fixture-only and local-test verified; it does not certify live AWS provider behavior.
- CloudTrail event record fields are parsed as payload/schema fields where CloudTrail returns them, but they are not counted as CDC because AWS does not document a CloudTrail changefeed/subscription API in the audited sources.
- Provider query direct reads are blocked/planned. They do not expose a generic CloudTrail Lake SQL runner.
- Resource-detail streams depend on the corresponding list/describe action returning identifiers; when no resources are discovered they emit no records.
- Write/admin actions are blocked/planned. No CloudTrail write action is listed as executable in generated catalog/docs/help for the current connector surface.
