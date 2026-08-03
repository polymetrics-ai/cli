# Overview

AWS CloudTrail connector parity was audited from the official AWS CloudTrail API Reference Actions page. The bundle enumerates all 60 official CloudTrail API actions exactly once. 19 ETL/read stream actions, 9 typed direct-read commands, and 29 typed reverse-ETL write/admin actions are implemented and runtime-reachable through the connector-local native surface. Streams and direct reads whose AWS action requires an identifier use connector-local discovery/fan-out from the fixed list/describe streams rather than caller-supplied raw action/path/header/body input. The official CloudTrail event-record contents page documents event record version 1.11 and 31 top-level event fields; those fields are schema payload fields for the `events lookup` direct-read command, not CDC/changefeed operations.

Implemented readable streams: `describe_trails`, `get_channel`, `get_dashboard`, `get_event_configuration`, `get_event_data_store`, `get_event_selectors`, `get_import`, `get_insight_selectors`, `get_resource_policy`, `get_trail`, `get_trail_status`, `list_channels`, `list_dashboards`, `list_event_data_stores`, `list_import_failures`, `list_imports`, `list_public_keys`, `list_tags`, `list_trails`.

Implemented direct-read operations: `cancel_query`, `describe_query`, `generate_query`, `get_query_results`, `list_insights_data`, `list_insights_metric_data`, `list_queries`, `lookup_events`, `search_sample_queries`.

Implemented write/admin actions: `add_tags`, `create_channel`, `create_event_data_store`, `create_trail`, `delete_channel`, `delete_dashboard`, `delete_event_data_store`, `delete_resource_policy`, `delete_trail`, `deregister_organization_delegated_admin`, `disable_federation`, `enable_federation`, `put_event_configuration`, `put_event_selectors`, `put_insight_selectors`, `put_resource_policy`, `register_organization_delegated_admin`, `remove_tags`, `restore_event_data_store`, `start_dashboard_refresh`, `start_event_data_store_ingestion`, `start_import`, `start_logging`, `stop_event_data_store_ingestion`, `stop_import`, `stop_logging`, `update_channel`, `update_event_data_store`, `update_trail`.

Blocked operations (3 of 60): `StartQuery`, `CreateDashboard`, `UpdateDashboard`. Each is blocked for a genuine, specific reason, not a missing-infrastructure placeholder: their request body requires an unrestricted CloudTrail Lake SQL `QueryStatement` (`StartQuery` directly; `CreateDashboard`/`UpdateDashboard` inside each `Widgets[].QueryStatement`). This project disables generic/unrestricted query-text execution for every connector — `capabilities.query` is fixed `false` repo-wide, and no connector ever exposes a raw SQL passthrough tool (see `AGENTS.md`). This is a standing project policy, not a per-connector gap, so these three stay blocked regardless of future shared-runtime work. `api_surface.json` records each with `operation.model: "disallowed"` and a `source_url` citation.

Note on the other 9 CloudTrail Lake/query-family actions (`CancelQuery`, `DescribeQuery`, `GenerateQuery`, `GetQueryResults`, `ListQueries`, `SearchSampleQueries`, plus the Insights actions and `LookupEvents`): none of their request fields accept raw SQL text — they take typed identifiers (`QueryId`, `EventDataStore`, bounded enums, a natural-language `Prompt` for `GenerateQuery`) — so they are genuinely safe to implement and are executable direct-read commands.

## Breaking change and migration

The earlier CloudTrail connector exposed four `LookupEvents`-backed event streams: `management_events`, `read_only_events`, `write_only_events`, and `console_logins`. None of them is an executable read stream in this native implementation, and their schemas are removed. Existing ETL connections that name one of them now fail with `aws-cloudtrail stream <name> not found`, and the incremental `EventTime` cursor those streams provided is gone; all 19 current streams are full-refresh.

`LookupEvents` is now available as the `events lookup` direct-read command (`pm aws-cloudtrail events lookup`), using typed `LookupAttributes`/time-range/event-category filters — this is the replacement path for CloudTrail event-record lookups, though it is a bounded direct read rather than an incremental ETL stream (no cursor/pagination-state persistence across runs).

The typed CloudTrail operations this connector implements read configuration, resource metadata, and (via direct reads) event/Insights records:

- Trail inventory and delivery state: `describe_trails`, `list_trails`, `get_trail`, `get_trail_status`.
- What a trail or event data store is configured to capture: `get_event_selectors`, `get_insight_selectors`, `get_event_configuration`.
- Lake and delivery resources: `list_event_data_stores`, `get_event_data_store`, `list_channels`, `get_channel`, `list_imports`, `get_import`, `list_import_failures`.
- Event and Insights records: `events lookup`, `insights data list`, `insights metric-data list`.
- CloudTrail Lake query lifecycle (typed, no raw SQL execution): `query list`, `query describe`, `query cancel`, `query results`, `query generate`, `query sample-search`.

## Auth setup

Use `pm credentials add <name> --connector aws-cloudtrail` and provide secrets only from environment variables or stdin:

- `aws_key_id` (required secret)
- `aws_secret_key` (required secret)
- `aws_region_name` (required config, for example `us-east-1`)

Optional config fields are `base_url` for local fixture endpoints, `page_size`, `max_pages`, and `mode=fixture` for credential-free tests. Do not place AWS secret values in chat, command history, docs, fixtures, or issue comments.

## Streams notes

Every implemented stream, direct read, and write uses a fixed AWS CloudTrail JSON-RPC action name with SigV4 authentication and no raw action/path/header/body escape hatch: `internal/connectors/native/aws-cloudtrail/api_contract.go` declares a closed, typed field schema per official action, and `body.go`'s `buildActionBody` rejects any field not in that schema. Paginated actions pass bounded `MaxResults` and follow `NextToken` until it is absent or `max_pages` is reached. Streams and direct reads for resource-detail actions derive required identifiers through connector-local discovery/fan-out from `DescribeTrails`, `ListChannels`, `ListDashboards`, `ListEventDataStores`, and `ListImports`. The connector keeps `query=false` at the metadata layer; CloudTrail Lake's generic SQL query capability (`StartQuery`) is never exposed, by design.

## Write actions & risks

29 CloudTrail reverse-ETL write/admin actions are exposed by the runtime surface, dispatched through the same signed JSON-RPC requester as reads via `internal/connectors/hooks/aws-cloudtrail`'s `ExecuteWrite` hook. Every write follows the standard plan -> preview -> explicit approval -> execute flow; destructive actions (`delete_channel`, `delete_dashboard`, `delete_event_data_store`, `delete_resource_policy`, `delete_trail`, `disable_federation`, `remove_tags`, `stop_event_data_store_ingestion`, `stop_import`, `stop_logging`) additionally require `confirm: destructive` and use provider-supported idempotency (missing-resource 404s on a delete are treated as already-written, not failed). No raw AWS action, path, header, or request-body escape hatch is exposed — `writes.json`'s `record_schema` for every action is `additionalProperties: false` with an explicit field list, and the same closed schema is enforced again at the native Go layer via `cloudTrailActionFields`.

`create_dashboard` and `update_dashboard` are not exposed: see the query-text policy note in Overview.

## Known limits

- This work is fixture-only and local-test verified; it does not certify live AWS provider behavior.
- CloudTrail event record fields are parsed as payload/schema fields where CloudTrail returns them (via `events lookup`), but they are not counted as CDC because AWS does not document a CloudTrail changefeed/subscription API in the audited sources.
- `StartQuery`, `CreateDashboard`, and `UpdateDashboard` are blocked by design; this connector never exposes a generic CloudTrail Lake SQL runner or a widget query builder.
- Resource-detail streams and direct reads depend on the corresponding list/describe action returning identifiers; when no resources are discovered they emit no records.
- Complex request fields (tag lists, event/insight selectors, channel destinations, import sources) are accepted as JSON-encoded values through both the CLI surface and `pm etl write`; each still validates against the action's closed field schema, so no field outside the official AWS request shape is ever accepted.

## Generated manual fidelity

`docs/connectors/aws-cloudtrail/MANUAL.md` and `SKILL.md` are written by `pm docs generate` from the same manifest that `pm connectors inspect aws-cloudtrail` renders, so the generated files and that live command (including `pm connectors inspect aws-cloudtrail --json`) are still intentionally sparse: no ETL STREAMS section, no SYNC MODES section, `No connector-specific config fields.` under CONFIGURATION, generic `connector-specific` read/write risk instead of the authored risk text in `metadata.json`, and — misleadingly — `No secret authentication is required for this connector.` under AUTHENTICATION. This connector does require the `aws_key_id` and `aws_secret_key` secrets plus the `aws_region_name` config field; the authoritative list is the Auth setup section above and `spec.json`.

That is not a gap in this connector. Those sections are built from `connectors.ManifestOf`, and the shared `definitionConnector` wrapper in `internal/connectors/native/nativeset/promoted.go` currently implements only `Definition()`, so `ManifestOf` falls back to a metadata-only manifest for every bundle-backed promoted native (~30 connectors today). The fix is foundation PR #3676, `fix(connectors): derive nativeset manifest from bundle definition`, which is deliberately owned outside this connector-local branch and is still open/unmerged as of this writing. Once #3676 merges and this branch rebases onto it, re-running `pm docs generate` fills in the full ETL STREAMS, CONFIGURATION, SYNC MODES, and authored SECURITY sections from this bundle with no change to the connector itself.

Until then, `pm connectors catalog --json`, `pm etl catalog --connector aws-cloudtrail`, `pm aws-cloudtrail --help` (via `cli_surface.json`), the generated connector catalog under `docs/connectors/catalog/`, and the website data read the bundle directly and already report the complete 19-stream/9-direct-read/29-write surface. Use those for CloudTrail stream, command, config, and risk facts instead of `pm connectors inspect`.
